// Command broker is the Council agent broker (Council doc 04 §11): a
// small deterministic state machine that freezes an immutable review
// bundle, launches blind first-party reviews, seals them, runs one
// cross-review round, and applies deterministic arbitration. It never
// reads or relays subscription credentials and never holds merge or
// release authority.
package main

import (
	"context"
	"encoding/json"
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
		os.Exit(1)
	}
}

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
               --human-approvals BOOL [--agents-available BOOL]
        Apply deterministic arbitration; write D/decision.json.

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

	sealDir := filepath.Join(*dir, "reviews", "blind")
	if err := os.MkdirAll(sealDir, 0o755); err != nil {
		return err
	}

	providers := []agent.Provider{agent.Anthropic, agent.OpenAI}
	for _, p := range providers {
		res, err := runner.Review(context.Background(), agent.Request{
			Provider: p, Prompt: prompt, BundleDir: *dir, TimeoutSecs: 1800, MaxTurns: 12,
		})
		if err != nil {
			return fmt.Errorf("%s review: %w", p, err)
		}
		// Untrusted until validated.
		if err := val.ValidateBytes(res.RawJSON); err != nil {
			return fmt.Errorf("%s review failed schema validation: %w", p, err)
		}
		// Seal: write the exact validated bytes and record their hash.
		out := filepath.Join(sealDir, string(p)+".json")
		if err := os.WriteFile(out, res.RawJSON, 0o644); err != nil {
			return err
		}
		sum, _ := bundle.FileSHA256(out)
		fmt.Printf("sealed %s review: sha256=%s\n", p, sum)
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
	crossDir := filepath.Join(*dir, "reviews", "cross")
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
	_ = fs.Parse(args)
	if *dir == "" {
		return fmt.Errorf("--dir required")
	}

	reviews, err := loadReviews(filepath.Join(*dir, "reviews", "blind"))
	if err != nil {
		return err
	}
	cross, _ := loadCross(filepath.Join(*dir, "reviews", "cross"))

	d := arbiter.Decide(
		arbiter.GateState{HardGatesPass: *hard, RequiredEvidence: *evid, HumanApprovals: *human},
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
	// Non-zero exit on a blocking decision so CI can gate on it.
	if d.Conclusion == "blocked" {
		os.Exit(3)
	}
	return nil
}

// --- helpers ---

type rawReview struct {
	ReviewID string `json:"review_id"`
	Provider string `json:"provider"`
	Verdict  string `json:"verdict"`
	Findings []struct {
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
	} `json:"findings"`
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
			r.Findings = append(r.Findings, arbiter.Finding{
				ID: f.ID, Fingerprint: f.Fingerprint, Severity: f.Severity,
				Category: f.Category, AcceptanceID: f.AcceptanceID, Claim: f.Claim,
				HasReproducer: f.Reproducer != nil,
			})
		}
		out = append(out, r)
	}
	return out, nil
}

type rawCross struct {
	Provider  string `json:"provider"`
	Responses []struct {
		TargetFindingID string `json:"target_finding_id"`
		Disposition     string `json:"disposition"`
	} `json:"responses"`
}

func loadCross(dir string) ([]arbiter.CrossResponse, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil
	}
	var out []arbiter.CrossResponse
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
			out = append(out, arbiter.CrossResponse{
				SourceProvider: rc.Provider, TargetFindingID: r.TargetFindingID, Disposition: r.Disposition,
			})
		}
	}
	return out, nil
}

func loadFixtures(spec string) (map[agent.Provider][]byte, error) {
	out := map[agent.Provider][]byte{}
	if spec == "" {
		return out, fmt.Errorf("mock mode needs --mock FILE:provider,...")
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
