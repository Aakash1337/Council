// Package arbiter applies the deterministic decision rules over sealed
// reviews, cross-review dispositions, and deterministic gate state
// (Council doc 04 §7.5). It is code and policy, not a third model:
// agreement is supporting evidence, never truth, and no agent vote can
// override a failed hard gate.
package arbiter

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Severity ordering for comparisons.
var sevRank = map[string]int{"critical": 4, "high": 3, "medium": 2, "low": 1, "info": 0}

// Finding is the subset of an agent finding the arbiter reasons over.
type Finding struct {
	ID            string   `json:"id"`
	Fingerprint   string   `json:"fingerprint"`
	Severity      string   `json:"severity"`
	Category      string   `json:"category"`
	AcceptanceID  []string `json:"acceptance_ids"`
	Claim         string   `json:"claim"`
	HasReproducer bool     `json:"has_reproducer"`
}

// Review is one sealed blind report.
type Review struct {
	ReviewID string    `json:"review_id"`
	Provider string    `json:"provider"`
	Verdict  string    `json:"verdict"`
	Findings []Finding `json:"findings"`
}

// CrossResponse is one disposition of the other reviewer's finding.
type CrossResponse struct {
	SourceProvider  string `json:"source_provider"`
	TargetFindingID string `json:"target_finding_id"`
	Disposition     string `json:"disposition"` // accept|reject|modify|needs_reproducer|needs_human
}

// Waiver is an authorized, scoped, expiring exception (doc 05 §13).
// Waiver *eligibility* is declared by upstream policy — only findings
// policy marks waivable may appear here; the arbiter enforces scope
// (exact finding ID) and expiry, and records the approver. Agents can
// never author these (doc 04 §2).
type Waiver struct {
	FindingID string `json:"finding_id"` // exact blocking finding ID
	HeadSHA   string `json:"head_sha"`   // candidate the waiver applies to — a waiver never travels across runs
	Approver  string `json:"approver"`   // named human approver
	Expiry    string `json:"expiry"`     // RFC3339; compared lexically-safe via time parse
}

// GateState is the deterministic CI outcome for the candidate.
type GateState struct {
	HardGatesPass    bool
	RequiredEvidence bool     // all required evidence present, fresh, hash-matched
	HumanApprovals   bool     // required human approvals present for the tier
	Waivers          []Waiver // policy-eligible, human-approved waivers
	Now              string   // RFC3339 evaluation time (injected for determinism)
	HeadSHA          string   // the reviewed candidate; waivers must name it exactly
}

// Decision is the normalized broker output (doc 05 §13 statuses).
type Decision struct {
	Conclusion     string   `json:"conclusion"`      // pass|pass_with_waiver|blocked|pending
	ReviewCoverage string   `json:"review_coverage"` // complete|pending_agent|agent_substituted|human_required
	BlockingIDs    []string `json:"blocking_ids"`
	WaivedIDs      []string `json:"waived_ids,omitempty"`
	Reasons        []string `json:"reasons"`
}

// key uniquely identifies a finding for dedup across reviewers.
func key(f Finding) string {
	if f.Fingerprint != "" {
		return f.Fingerprint
	}
	acc := append([]string(nil), f.AcceptanceID...)
	sort.Strings(acc)
	return strings.Join(acc, ",") + "|" + f.Category + "|" + f.Claim
}

// Decide applies deterministic precedence. Order matters: evidence and
// hard gates first, then human controls, then agent findings.
func Decide(gate GateState, reviews []Review, cross []CrossResponse, agentsAvailable bool) Decision {
	d := Decision{}

	// 1. Input integrity / required evidence (doc 04 §7.5 row 1).
	if !gate.RequiredEvidence {
		d.Conclusion = "blocked"
		d.Reasons = append(d.Reasons, "required evidence missing, stale, or hash-mismatched")
		return d
	}

	// 2. Hard deterministic gates (row 2) — no agent vote can override.
	if !gate.HardGatesPass {
		d.Conclusion = "blocked"
		d.Reasons = append(d.Reasons, "deterministic hard gate failed")
		return d
	}

	// Agent coverage classification (doc 04 §9.1). Deterministic pass is
	// not turned into a failure by agent unavailability; coverage changes.
	if !agentsAvailable || len(reviews) == 0 {
		d.Conclusion = "pending"
		d.ReviewCoverage = "pending_agent"
		d.Reasons = append(d.Reasons, "agent review unavailable; deterministic CI passed but required review incomplete")
		return d
	}
	d.ReviewCoverage = "complete"

	// Index cross-review dispositions by (targetFindingID).
	disp := map[string][]string{}
	for _, c := range cross {
		disp[c.TargetFindingID] = append(disp[c.TargetFindingID], c.Disposition)
	}

	// Collect blocking findings by precedence rules (rows 4–10).
	type agg struct {
		f            Finding
		reviewers    map[string]bool
		reproduced   bool
		otherAccepts bool
		otherRejects bool
	}
	byKey := map[string]*agg{}
	for _, r := range reviews {
		for _, f := range r.Findings {
			k := key(f)
			a := byKey[k]
			if a == nil {
				a = &agg{f: f, reviewers: map[string]bool{}}
				byKey[k] = a
			}
			a.reviewers[r.Provider] = true
			if f.HasReproducer {
				a.reproduced = true
			}
			// Keep the highest severity seen for this root cause.
			if sevRank[f.Severity] > sevRank[a.f.Severity] {
				a.f.Severity = f.Severity
			}
			for _, dp := range disp[f.ID] {
				switch dp {
				case "accept":
					a.otherAccepts = true
				case "reject":
					a.otherRejects = true
				case "needs_reproducer", "needs_human":
					// neither confirms nor clears
				}
			}
		}
	}

	// Deterministic ordering of keys for stable output.
	keys := make([]string, 0, len(byKey))
	for k := range byKey {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// humanIDs tracks, per finding ID, the rows that demand human
	// adjudication. A valid waiver on that exact finding IS the human
	// adjudication, so waiving clears both the block and the human
	// requirement for that finding (review finding on #9).
	humanIDs := map[string]bool{}
	for _, k := range keys {
		a := byKey[k]
		sev := a.f.Severity
		highish := sev == "critical" || sev == "high"
		bothFound := len(a.reviewers) >= 2

		switch {
		case highish && a.reproduced:
			// Row: critical/high with deterministic reproducer -> block.
			d.BlockingIDs = append(d.BlockingIDs, a.f.ID)
			d.Reasons = append(d.Reasons, fmt.Sprintf("%s: %s finding with reproducer", a.f.ID, sev))
		case highish && (bothFound || a.otherAccepts) && !a.otherRejects:
			// Confirmed by the other reviewer but not yet reproduced ->
			// block pending reproducer/human.
			d.BlockingIDs = append(d.BlockingIDs, a.f.ID)
			humanIDs[a.f.ID] = true
			d.Reasons = append(d.Reasons, fmt.Sprintf("%s: %s finding confirmed by both reviewers, needs reproducer/human", a.f.ID, sev))
		case highish && !a.otherRejects:
			// Lone unrefuted high/critical -> not dismissed; route to human.
			humanIDs[a.f.ID] = true
			d.Reasons = append(d.Reasons, fmt.Sprintf("%s: unrefuted %s finding, human adjudication", a.f.ID, sev))
		case sev == "medium" && len(a.f.AcceptanceID) > 0 && a.reproduced:
			// Medium with demonstrated acceptance violation -> block.
			d.BlockingIDs = append(d.BlockingIDs, a.f.ID)
			d.Reasons = append(d.Reasons, fmt.Sprintf("%s: medium finding violates acceptance criterion", a.f.ID))
		default:
			// low/info/advisory or refuted -> annotate only.
		}
	}

	// 5. Human controls (row 5) — required approvals must be present.
	if !gate.HumanApprovals {
		d.Conclusion = "pending"
		d.ReviewCoverage = "human_required"
		d.Reasons = append(d.Reasons, "required human approval absent")
		return d
	}

	// Waiver application (doc 05 §13): a finding is waived only when an
	// unexpired, human-approved waiver names its exact ID AND the exact
	// candidate head SHA — a waiver never travels across runs (review
	// finding on #9). Hard-gate and evidence failures returned above are
	// non-waivable by construction.
	if len(gate.Waivers) > 0 && (len(d.BlockingIDs) > 0 || len(humanIDs) > 0) {
		valid := map[string]Waiver{}
		for _, w := range gate.Waivers {
			switch {
			case w.Approver == "":
				d.Reasons = append(d.Reasons, fmt.Sprintf("waiver for %s has no approver — not applied", w.FindingID))
			case expired(w.Expiry, gate.Now):
				d.Reasons = append(d.Reasons, fmt.Sprintf("waiver for %s expired %s — not applied", w.FindingID, w.Expiry))
			case gate.HeadSHA == "" || w.HeadSHA != gate.HeadSHA:
				d.Reasons = append(d.Reasons, fmt.Sprintf("waiver for %s is scoped to head %q, candidate is %q — not applied", w.FindingID, w.HeadSHA, gate.HeadSHA))
			default:
				valid[w.FindingID] = w
			}
		}
		var stillBlocking []string
		for _, id := range d.BlockingIDs {
			if w, ok := valid[id]; ok {
				d.WaivedIDs = append(d.WaivedIDs, id)
				delete(humanIDs, id) // the waiver is the human adjudication
				d.Reasons = append(d.Reasons, fmt.Sprintf("%s waived by %s until %s (head %s)", id, w.Approver, w.Expiry, w.HeadSHA))
			} else {
				stillBlocking = append(stillBlocking, id)
			}
		}
		d.BlockingIDs = stillBlocking
		// A waiver may also adjudicate a non-blocking human-required
		// finding (lone unrefuted high).
		for id := range humanIDs {
			if w, ok := valid[id]; ok {
				d.WaivedIDs = append(d.WaivedIDs, id)
				delete(humanIDs, id)
				d.Reasons = append(d.Reasons, fmt.Sprintf("%s waived by %s until %s (head %s)", id, w.Approver, w.Expiry, w.HeadSHA))
			}
		}
	}

	sort.Strings(d.BlockingIDs)
	sort.Strings(d.WaivedIDs)
	switch {
	case len(d.BlockingIDs) > 0:
		d.Conclusion = "blocked"
	case len(humanIDs) > 0:
		d.Conclusion = "pending"
		d.ReviewCoverage = "human_required"
	case len(d.WaivedIDs) > 0:
		d.Conclusion = "pass_with_waiver"
	default:
		d.Conclusion = "pass"
	}
	return d
}

// expired reports whether the waiver expiry is at or before now. Both
// are RFC3339; unparseable values fail closed (treated as expired).
func expired(expiry, now string) bool {
	e, err1 := time.Parse(time.RFC3339, expiry)
	n, err2 := time.Parse(time.RFC3339, now)
	if err1 != nil || err2 != nil {
		return true
	}
	return !e.After(n)
}
