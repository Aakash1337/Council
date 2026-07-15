// Command broker is the Council agent broker (Council doc 04 §11): a
// small deterministic state machine that freezes an immutable review
// bundle, launches blind first-party reviews, seals them, runs one
// cross-review round, and applies deterministic arbitration. It never
// reads or relays subscription credentials and never holds merge or
// release authority.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Aakash1337/Council/broker/internal/agent"
	"github.com/Aakash1337/Council/broker/internal/arbiter"
	"github.com/Aakash1337/Council/broker/internal/bundle"
	"github.com/Aakash1337/Council/broker/internal/schema"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "freeze":
		err = cmdFreeze(os.Args[2:])
	case "review":
		err = cmdReview(os.Args[2:])
	case "cross-review":
		err = cmdCross(os.Args[2:])
	case "decide":
		err = cmdDecide(os.Args[2:])
	case "repair-gate":
		err = cmdRepairGate(os.Args[2:])
	case "help", "-h", "--help":
		usage()
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		if errors.Is(err, errBlocked) {
			os.Exit(3)
		}
		os.Exit(1)
	}
}

// errBlocked marks a blocking policy decision (CLI exit code 3).
var errBlocked = errors.New("blocked")

func usage() {
	fmt.Fprint(os.Stderr, `broker <command>

  freeze       --dir D --run-id ID --repo R --change C --risk T \
               --base SHA --head SHA --spec-sha S --created TS
        Freeze and hash the review bundle in D; writes bundle.json.

  review       --dir D --schema agent-review.schema.json [--mock FILE:PROV,...] \
               [--real] [--claude-args '...'] [--codex-args '...']
        Run blind reviews, validate against schema, seal (hash) each
        report into D/reviews/blind/. Mock mode replays fixtures.

  cross-review --dir D --schema agent-cross-review.schema.json
        (scaffold) validate cross-review reports in D/reviews/cross/.

  decide       --dir D --hard-gates-pass BOOL --evidence BOOL \
               --human-approvals BOOL [--agents-available BOOL] \
               [--waivers-file F --now TS]
        Apply deterministic arbitration; write D/decision.json.

  repair-gate  --ledger F --change C --from-head SHA [--reason R] \
               [--max N<=2] [--started-at TS]
        Authorize the next bounded repair cycle (ADR-008): records the
        cycle in the per-change ledger, refuses beyond the maximum or
        for an already-repaired head (no-progress loop).

The broker treats all agent output as untrusted until schema-validated,
seals blind reports before cross-review, and never lets an agent vote
override a failed hard gate.
`)
}

func cmdFreeze(args []string) error {
	fs := flag.NewFlagSet("freeze", flag.ExitOnError)
	dir := fs.String("dir", "", "bundle directory")
	runID := fs.String("run-id", "", "run id")
	repo := fs.String("repo", "", "repository")
	change := fs.String("change", "", "change id")
	risk := fs.String("risk", "medium", "risk tier")
	base := fs.String("base", "", "base sha")
	head := fs.String("head", "", "head sha")
	specSHA := fs.String("spec-sha", "", "spec sha256")
	created := fs.String("created", "", "created timestamp (RFC3339)")
	_ = fs.Parse(args)
	if *dir == "" || *runID == "" {
		return fmt.Errorf("--dir and --run-id required")
	}
	b, err := bundle.Freeze(*dir, *runID, *repo, *change, *risk, *base, *head, *specSHA, *created, nil)
	if err != nil {
		return err
	}
	if err := b.Write(*dir); err != nil {
		return err
	}
	fmt.Printf("frozen: bundle_sha256=%s (%d files)\n", b.BundleSHA256, len(b.Files))
	return nil
}

func cmdReview(args []string) error {
	fs := flag.NewFlagSet("review", flag.ExitOnError)
	dir := fs.String("dir", "", "bundle directory")
	schemaPath := fs.String("schema", "", "agent-review schema")
	mock := fs.String("mock", "", "comma list of FILE:provider fixtures for mock mode")
	real := fs.Bool("real", false, "use real first-party CLIs")
	claudeBin := fs.String("claude-bin", "", "path to the claude client (e.g. claude.cmd on Windows)")
	codexBin := fs.String("codex-bin", "", "path to the codex client (e.g. codex.cmd on Windows)")
	claudeArgs := fs.String("claude-args", "", "space-separated claude args")
	codexArgs := fs.String("codex-args", "", "space-separated codex args")
	promptFile := fs.String("prompt", "", "review prompt file")
	_ = fs.Parse(args)
	if *dir == "" || *schemaPath == "" {
		return fmt.Errorf("--dir and --schema required")
	}

	// Verify the frozen bundle before spending any review.
	b, err := bundle.Load(*dir)
	if err != nil {
		return fmt.Errorf("load bundle (run freeze first): %w", err)
	}
	if err := b.Verify(*dir); err != nil {
		return fmt.Errorf("bundle verification failed: %w", err)
	}

	val, err := schema.New(*schemaPath)
	if err != nil {
		return err
	}

	var runner agent.Runner
	if *real {
		runner = &agent.CLIRunner{
			ClaudeBin:  *claudeBin,
			CodexBin:   *codexBin,
			ClaudeArgs: splitArgs(*claudeArgs),
			CodexArgs:  splitArgs(*codexArgs),
		}
	} else {
		fixtures, err := loadFixtures(*mock)
		if err != nil {
			return err
		}
		runner = &agent.MockRunner{Fixtures: fixtures}
	}

	prompt := ""
	if *promptFile != "" {
		p, err := os.ReadFile(*promptFile)
		if err != nil {
			return err
		}
		prompt = string(p)
	}

	// Blind-review invariant (doc 04 §7.2–7.3): no reviewer may see
	// another's report. Each reviewer runs with the bundle dir as cwd, so
	// we must NOT write any sealed report there while another review is
	// still pending. Run every review first, buffering validated output
	// in memory; only after all reviews complete do we write and hash the
	// seals — into a seal dir OUTSIDE the bundle (a sibling), so review
	// subprocesses never have a sealed peer report on a readable path.
	type sealed struct {
		provider agent.Provider
		raw      []byte
	}
	var results []sealed

	providers := []agent.Provider{agent.Anthropic, agent.OpenAI}
	for _, p := range providers {
		// Each reviewer gets a prompt carrying ITS OWN identity and the
		// frozen bundle's real hashes — never another provider's identity
		// (the seeded-defect probe found a shared prompt made Claude report
		// Codex's identity). In mock mode the fixture prompt is used as-is.
		reviewerPrompt := prompt
		if *real {
			reviewerPrompt = buildReviewPrompt(prompt, p, b)
		}
		res, err := runner.Review(context.Background(), agent.Request{
			Provider: p, Prompt: reviewerPrompt, BundleDir: *dir, TimeoutSecs: 1800, MaxTurns: 12,
		})
		if err != nil {
			return fmt.Errorf("%s review: %w", p, err)
		}
		// Untrusted until validated. One bounded schema-repair attempt
		// (doc 04 §7.2): if the raw report fails validation, re-prompt the
		// SAME model once with the exact errors, asking it to fix ONLY the
		// JSON shape — not the substance. A second failure is a hard error,
		// never a pass.
		if verr := val.ValidateBytes(res.RawJSON); verr != nil {
			if !*real {
				return fmt.Errorf("%s review failed schema validation: %w", p, verr)
			}
			repairPrompt := fmt.Sprintf(
				"Your previous JSON review did not conform to the schema. Fix ONLY the JSON shape (field names, enums, types) — do not change any finding's substance. Errors:\n%s\n\nYour previous output was:\n%s\n\nOutput ONLY the corrected JSON review object.",
				verr.Error(), string(res.RawJSON))
			res, err = runner.Review(context.Background(), agent.Request{
				Provider: p, Prompt: repairPrompt, BundleDir: *dir, TimeoutSecs: 1800, MaxTurns: 12,
			})
			if err != nil {
				return fmt.Errorf("%s schema-repair attempt: %w", p, err)
			}
			if verr2 := val.ValidateBytes(res.RawJSON); verr2 != nil {
				return fmt.Errorf("%s review still invalid after one schema-repair attempt (infrastructure_error, not approval): %w", p, verr2)
			}
			fmt.Printf("%s review: one schema-repair attempt applied\n", p)
		}
		// Semantic validation beyond JSON Schema (doc 04 §9.3): the
		// reported provider/client MUST match the actual invocation. A
		// model that self-reports the wrong identity (or copies an
		// example) is rejected — the broker trusts the invocation record,
		// not the model's self-description.
		if err := checkReviewIdentity(res.RawJSON, p, res.Client); err != nil {
			return fmt.Errorf("%s review identity check failed: %w", p, err)
		}
		results = append(results, sealed{provider: p, raw: res.RawJSON})
	}

	// Seal after all reviews finished. Seals live in <dir>-seals/blind so
	// they are outside the bundle any reviewer used as cwd. A seals.json
	// manifest records the hashes so `decide` can detect post-seal edits.
	sealDir := filepath.Join(*dir+"-seals", "blind")
	if err := os.MkdirAll(sealDir, 0o755); err != nil {
		return err
	}
	seals := map[string]string{}
	for _, s := range results {
		out := filepath.Join(sealDir, string(s.provider)+".json")
		if err := os.WriteFile(out, s.raw, 0o644); err != nil {
			return err
		}
		sum, _ := bundle.FileSHA256(out)
		seals[string(s.provider)+".json"] = sum
		fmt.Printf("sealed %s review: sha256=%s\n", s.provider, sum)
	}
	sealsJSON, _ := json.MarshalIndent(seals, "", "  ")
	if err := os.WriteFile(filepath.Join(*dir+"-seals", "seals.json"), sealsJSON, 0o644); err != nil {
		return err
	}
	return nil
}

func cmdCross(args []string) error {
	fs := flag.NewFlagSet("cross-review", flag.ExitOnError)
	dir := fs.String("dir", "", "bundle directory")
	schemaPath := fs.String("schema", "", "agent-cross-review schema")
	_ = fs.Parse(args)
	if *dir == "" || *schemaPath == "" {
		return fmt.Errorf("--dir and --schema required")
	}
	val, err := schema.New(*schemaPath)
	if err != nil {
		return err
	}
	crossDir := filepath.Join(*dir+"-seals", "cross")
	entries, err := os.ReadDir(crossDir)
	if err != nil {
		return fmt.Errorf("no cross-review reports (dir %s): %w", crossDir, err)
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(crossDir, e.Name()))
		if err != nil {
			return err
		}
		if err := val.ValidateBytes(raw); err != nil {
			return fmt.Errorf("%s failed cross-review schema: %w", e.Name(), err)
		}
		fmt.Printf("validated cross-review %s\n", e.Name())
	}
	return nil
}

func cmdDecide(args []string) error {
	fs := flag.NewFlagSet("decide", flag.ExitOnError)
	dir := fs.String("dir", "", "bundle directory")
	hard := fs.Bool("hard-gates-pass", false, "deterministic hard gates passed")
	evid := fs.Bool("evidence", false, "required evidence present and hash-matched")
	human := fs.Bool("human-approvals", false, "required human approvals present")
	agentsAvail := fs.Bool("agents-available", true, "agent lane available")
	waiversFile := fs.String("waivers-file", "", "JSON array of policy-eligible waivers (finding_id, approver, expiry)")
	now := fs.String("now", "", "RFC3339 evaluation time for waiver expiry (required with --waivers-file)")
	_ = fs.Parse(args)
	if *dir == "" {
		return fmt.Errorf("--dir required")
	}
	waivers, err := loadWaivers(*waiversFile)
	if err != nil {
		return fmt.Errorf("waivers: %w", err)
	}
	if len(waivers) > 0 && *now == "" {
		return fmt.Errorf("--now is required when waivers are supplied (expiry is checked deterministically)")
	}

	// Re-verify sealed blind reports against the hashes recorded at seal
	// time (doc 04 §9.3): any post-seal edit that removes a blocking
	// finding must be detected, not silently trusted.
	sealBase := *dir + "-seals"
	if err := verifySeals(sealBase); err != nil {
		return fmt.Errorf("seal verification failed: %w", err)
	}

	reviews, err := loadReviews(filepath.Join(sealBase, "blind"))
	if err != nil {
		return err
	}
	cross, crossFindings := loadCross(filepath.Join(sealBase, "cross"))
	// Cross-review may surface new findings (schema `new_findings`); a
	// newly discovered blocker must reach the arbiter, so carry them as an
	// extra review lane rather than dropping them.
	if len(crossFindings) > 0 {
		reviews = append(reviews, arbiter.Review{
			ReviewID: "cross-new-findings", Provider: "cross", Verdict: "changes_required",
			Findings: crossFindings,
		})
	}

	// Waivers are scoped to the exact reviewed candidate: take the head
	// SHA from the frozen bundle. Without a bundle, no waiver can apply
	// (the arbiter fails closed on an empty gate head).
	headSHA := ""
	if b, err := bundle.Load(*dir); err == nil {
		headSHA = b.HeadSHA
	} else if len(waivers) > 0 {
		return fmt.Errorf("waivers supplied but bundle.json is unreadable (needed for head-SHA scoping): %w", err)
	}

	d := arbiter.Decide(
		arbiter.GateState{
			HardGatesPass: *hard, RequiredEvidence: *evid, HumanApprovals: *human,
			Waivers: waivers, Now: *now, HeadSHA: headSHA,
		},
		reviews, cross, *agentsAvail,
	)
	out, _ := json.MarshalIndent(d, "", "  ")
	if err := os.WriteFile(filepath.Join(*dir, "decision.json"), out, 0o644); err != nil {
		return err
	}
	fmt.Printf("decision: %s coverage=%s blocking=%v\n", d.Conclusion, d.ReviewCoverage, d.BlockingIDs)
	for _, r := range d.Reasons {
		fmt.Println("  reason:", r)
	}
	// A blocking decision is a distinguishable error so main can exit 3
	// for CI gating while tests can assert on it (no os.Exit here).
	if d.Conclusion == "blocked" {
		return fmt.Errorf("decision blocked (%v): %w", d.BlockingIDs, errBlocked)
	}
	return nil
}

// --- helpers ---

type rawReview struct {
	ReviewID string       `json:"review_id"`
	Provider string       `json:"provider"`
	Verdict  string       `json:"verdict"`
	Findings []rawFinding `json:"findings"`
}

func loadReviews(dir string) ([]arbiter.Review, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil // no reviews yet -> caller treats as agent-unavailable
	}
	var out []arbiter.Review
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		var rr rawReview
		if err := json.Unmarshal(raw, &rr); err != nil {
			return nil, err
		}
		r := arbiter.Review{ReviewID: rr.ReviewID, Provider: rr.Provider, Verdict: rr.Verdict}
		for _, f := range rr.Findings {
			r.Findings = append(r.Findings, f.toArbiter())
		}
		out = append(out, r)
	}
	return out, nil
}

type rawFinding struct {
	ID           string   `json:"id"`
	Fingerprint  string   `json:"fingerprint"`
	Severity     string   `json:"severity"`
	Category     string   `json:"category"`
	AcceptanceID []string `json:"acceptance_ids"`
	Claim        string   `json:"claim"`
	Reproducer   *struct {
		Kind string `json:"kind"`
		Ref  string `json:"ref"`
	} `json:"reproducer"`
}

func (f rawFinding) toArbiter() arbiter.Finding {
	return arbiter.Finding{
		ID: f.ID, Fingerprint: f.Fingerprint, Severity: f.Severity,
		Category: f.Category, AcceptanceID: f.AcceptanceID, Claim: f.Claim,
		HasReproducer: f.Reproducer != nil,
	}
}

type rawCross struct {
	Provider  string `json:"provider"`
	Responses []struct {
		TargetFindingID string `json:"target_finding_id"`
		Disposition     string `json:"disposition"`
	} `json:"responses"`
	NewFindings []rawFinding `json:"new_findings"`
}

// loadCross returns both the dispositions on the other reviewer's findings
// and any findings newly discovered during cross-review (schema
// `new_findings`), which must not be dropped before arbitration.
func loadCross(dir string) ([]arbiter.CrossResponse, []arbiter.Finding) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil
	}
	var resp []arbiter.CrossResponse
	var newFindings []arbiter.Finding
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		raw, _ := os.ReadFile(filepath.Join(dir, e.Name()))
		var rc rawCross
		if json.Unmarshal(raw, &rc) != nil {
			continue
		}
		for _, r := range rc.Responses {
			resp = append(resp, arbiter.CrossResponse{
				SourceProvider: rc.Provider, TargetFindingID: r.TargetFindingID, Disposition: r.Disposition,
			})
		}
		for _, f := range rc.NewFindings {
			newFindings = append(newFindings, f.toArbiter())
		}
	}
	return resp, newFindings
}

// verifySeals recomputes each sealed blind report's hash and compares it
// to the manifest written at seal time. A mismatch means a report was
// edited after sealing — fail closed.
func verifySeals(sealBase string) error {
	raw, err := os.ReadFile(filepath.Join(sealBase, "seals.json"))
	if err != nil {
		// No seals recorded yet (e.g. agent-unavailable run) — nothing to verify.
		return nil
	}
	var seals map[string]string
	if err := json.Unmarshal(raw, &seals); err != nil {
		return err
	}
	for name, want := range seals {
		got, err := bundle.FileSHA256(filepath.Join(sealBase, "blind", name))
		if err != nil {
			return fmt.Errorf("sealed report %s missing: %w", name, err)
		}
		if got != want {
			return fmt.Errorf("sealed report %s changed after sealing: have %s want %s", name, got, want)
		}
	}
	return nil
}

// loadWaivers reads a JSON array of policy-eligible waivers. An empty
// path means no waivers. Agents never author these files (doc 04 §2);
// they arrive from the human waiver process (doc 05 §13).
// checkReviewIdentity enforces doc 04 §9.3: the review's self-reported
// provider/client must match the actual invocation.
func checkReviewIdentity(raw []byte, wantProvider agent.Provider, wantClient string) error {
	var r struct {
		Provider string `json:"provider"`
		Client   string `json:"client"`
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return err
	}
	if r.Provider != string(wantProvider) {
		return fmt.Errorf("reported provider %q != invoked %q", r.Provider, wantProvider)
	}
	if r.Client != wantClient {
		return fmt.Errorf("reported client %q != invoked %q", r.Client, wantClient)
	}
	return nil
}

// buildReviewPrompt appends a per-reviewer identity block and the frozen
// bundle's real hashes to the base review instructions, so each provider
// returns a correctly-identified report bound to the actual candidate.
func buildReviewPrompt(base string, p agent.Provider, b *bundle.Bundle) string {
	client := "claude-code"
	if p == agent.OpenAI {
		client = "codex-cli"
	}
	reviewID := "review-claude-01"
	if p == agent.OpenAI {
		reviewID = "review-codex-01"
	}
	hex64 := func(s string) string { // valid lowercase 64-hex for the schema
		for len(s) < 64 {
			s += "0"
		}
		out := []byte(s[:64])
		for i, c := range out {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				out[i] = '0'
			}
		}
		return string(out)
	}
	return base + fmt.Sprintf(`

## Required identity for YOUR report (use these EXACT values; do not copy any example)
schema_version: 1
run_id: %q
review_id: %q
provider: %q
client: %q
model: "recorded-by-client"
role: "independent_reviewer"
phase: "blind"
base_sha: %q
head_sha: %q
spec_sha256: %q
bundle_sha256: %q
prompt_template_sha256: %q

Constraints: severity is exactly one of {critical,high,medium,low,info} lowercase; location is an object {path,start_line,end_line} or null; reproducer is an object {kind,ref} or null; include unreviewed_areas[] , evidence_gaps[] , and completed_at (RFC3339).

Output ONLY the JSON review object. No markdown fences, no prose.`,
		b.RunID, reviewID, string(p), client, b.BaseSHA, b.HeadSHA,
		hex64(b.SpecSHA256), hex64(b.BundleSHA256), hashHex(base))
}

// hashHex returns the lowercase hex SHA-256 of s.
func hashHex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func loadWaivers(path string) ([]arbiter.Waiver, error) {
	if path == "" {
		return nil, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var ws []arbiter.Waiver
	if err := json.Unmarshal(raw, &ws); err != nil {
		return nil, err
	}
	return ws, nil
}

func loadFixtures(spec string) (map[agent.Provider][]byte, error) {
	out := map[agent.Provider][]byte{}
	if spec == "" {
		return out, fmt.Errorf("mock mode needs one or more --mock FILE:provider pairs")
	}
	for _, pair := range splitComma(spec) {
		file, prov, ok := cut(pair, ":")
		if !ok {
			return nil, fmt.Errorf("bad --mock entry %q (want FILE:provider)", pair)
		}
		b, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		out[agent.Provider(prov)] = b
	}
	return out, nil
}

func splitArgs(s string) []string {
	if s == "" {
		return nil
	}
	return splitWS(s)
}
