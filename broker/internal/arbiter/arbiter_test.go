package arbiter

import "testing"

func hardPass() GateState {
	return GateState{HardGatesPass: true, RequiredEvidence: true, HumanApprovals: true}
}

// A failed hard gate blocks even when both reviewers approve — the core
// invariant (doc 04 §7.5, PAC-005).
func TestHardGateBlocksDespiteApproval(t *testing.T) {
	g := GateState{HardGatesPass: false, RequiredEvidence: true, HumanApprovals: true}
	reviews := []Review{
		{Provider: "anthropic", Verdict: "approve"},
		{Provider: "openai", Verdict: "approve"},
	}
	d := Decide(g, reviews, nil, true)
	if d.Conclusion != "blocked" {
		t.Fatalf("expected blocked, got %s", d.Conclusion)
	}
}

// Missing evidence fails closed before anything else.
func TestMissingEvidenceFailsClosed(t *testing.T) {
	g := GateState{HardGatesPass: true, RequiredEvidence: false, HumanApprovals: true}
	d := Decide(g, []Review{{Provider: "anthropic", Verdict: "approve"}}, nil, true)
	if d.Conclusion != "blocked" {
		t.Fatalf("expected blocked on missing evidence, got %s", d.Conclusion)
	}
}

// Agent unavailability yields pending/pending_agent, not a failure, when
// deterministic CI passed (doc 04 §9.1).
func TestAgentUnavailablePending(t *testing.T) {
	d := Decide(hardPass(), nil, nil, false)
	if d.Conclusion != "pending" || d.ReviewCoverage != "pending_agent" {
		t.Fatalf("expected pending/pending_agent, got %s/%s", d.Conclusion, d.ReviewCoverage)
	}
}

// Both reviewers approve, no findings, gates green -> pass.
func TestCleanApprove(t *testing.T) {
	reviews := []Review{
		{Provider: "anthropic", Verdict: "approve"},
		{Provider: "openai", Verdict: "approve"},
	}
	d := Decide(hardPass(), reviews, nil, true)
	if d.Conclusion != "pass" {
		t.Fatalf("expected pass, got %s (%v)", d.Conclusion, d.Reasons)
	}
}

// A high finding with a reproducer blocks regardless of the other model.
func TestHighWithReproducerBlocks(t *testing.T) {
	reviews := []Review{
		{Provider: "openai", Verdict: "changes_required", Findings: []Finding{
			{ID: "COD-001", Fingerprint: "fp1", Severity: "high", Category: "concurrency",
				AcceptanceID: []string{"CDNS-002/AC-006"}, HasReproducer: true},
		}},
		{Provider: "anthropic", Verdict: "approve"},
	}
	d := Decide(hardPass(), reviews, nil, true)
	if d.Conclusion != "blocked" {
		t.Fatalf("expected blocked, got %s", d.Conclusion)
	}
	if len(d.BlockingIDs) != 1 || d.BlockingIDs[0] != "COD-001" {
		t.Fatalf("expected COD-001 blocking, got %v", d.BlockingIDs)
	}
}

// A lone unreproduced high finding is not dismissed because the other
// model missed it — it routes to a human (doc 04 §7.5 last paragraph).
func TestLoneHighRoutesToHuman(t *testing.T) {
	reviews := []Review{
		{Provider: "anthropic", Verdict: "changes_required", Findings: []Finding{
			{ID: "CLA-004", Fingerprint: "fp2", Severity: "high", Category: "security"},
		}},
		{Provider: "openai", Verdict: "approve"},
	}
	d := Decide(hardPass(), reviews, nil, true)
	if d.Conclusion != "pending" || d.ReviewCoverage != "human_required" {
		t.Fatalf("expected pending/human_required, got %s/%s", d.Conclusion, d.ReviewCoverage)
	}
}

// The other reviewer refuting a lone high finding clears it to annotate.
func TestRefutedHighClears(t *testing.T) {
	reviews := []Review{
		{Provider: "anthropic", Verdict: "changes_required", Findings: []Finding{
			{ID: "CLA-005", Fingerprint: "fp3", Severity: "high", Category: "perf"},
		}},
		{Provider: "openai", Verdict: "approve"},
	}
	cross := []CrossResponse{
		{SourceProvider: "openai", TargetFindingID: "CLA-005", Disposition: "reject"},
	}
	d := Decide(hardPass(), reviews, cross, true)
	if d.Conclusion != "pass" {
		t.Fatalf("expected pass after refutation, got %s (%v)", d.Conclusion, d.Reasons)
	}
}

// Both reviewers independently find the same high issue (no reproducer):
// confirmed -> block pending reproducer/human.
func TestBothFindSameHighBlocks(t *testing.T) {
	reviews := []Review{
		{Provider: "anthropic", Verdict: "changes_required", Findings: []Finding{
			{ID: "CLA-006", Fingerprint: "shared", Severity: "high", Category: "concurrency"},
		}},
		{Provider: "openai", Verdict: "changes_required", Findings: []Finding{
			{ID: "COD-006", Fingerprint: "shared", Severity: "high", Category: "concurrency"},
		}},
	}
	d := Decide(hardPass(), reviews, nil, true)
	if d.Conclusion != "blocked" {
		t.Fatalf("expected blocked on independent corroboration, got %s", d.Conclusion)
	}
}

// A blocker discovered only in cross-review (new_findings), when both
// blind reviews approved, must still block (regression for the dropped
// new_findings bug).
func TestCrossReviewNewFindingBlocks(t *testing.T) {
	blind := []Review{
		{Provider: "anthropic", Verdict: "approve"},
		{Provider: "openai", Verdict: "approve"},
	}
	// Simulate the decide-layer behavior: cross-review new findings are
	// carried as an extra review lane.
	crossLane := Review{Provider: "cross", Verdict: "changes_required", Findings: []Finding{
		{ID: "XR-001", Fingerprint: "xrfp", Severity: "high", Category: "security", HasReproducer: true},
	}}
	d := Decide(hardPass(), append(blind, crossLane), nil, true)
	if d.Conclusion != "blocked" {
		t.Fatalf("cross-review-discovered blocker must block, got %s", d.Conclusion)
	}
}

// Row 9: a medium finding with a reproduced acceptance violation blocks.
func TestMediumWithReproducedAcceptanceViolationBlocks(t *testing.T) {
	reviews := []Review{
		{Provider: "openai", Verdict: "changes_required", Findings: []Finding{
			{ID: "COD-010", Fingerprint: "m1", Severity: "medium", Category: "correctness",
				AcceptanceID: []string{"CDNS-001/AC-004"}, HasReproducer: true},
		}},
		{Provider: "anthropic", Verdict: "approve"},
	}
	d := Decide(hardPass(), reviews, nil, true)
	if d.Conclusion != "blocked" {
		t.Fatalf("expected blocked, got %s (%v)", d.Conclusion, d.Reasons)
	}
}

// A medium finding WITHOUT a reproducer does not block (annotate only).
func TestMediumWithoutReproducerAnnotates(t *testing.T) {
	reviews := []Review{
		{Provider: "openai", Verdict: "changes_required", Findings: []Finding{
			{ID: "COD-011", Fingerprint: "m2", Severity: "medium", Category: "style",
				AcceptanceID: []string{"CDNS-001/AC-001"}},
		}},
		{Provider: "anthropic", Verdict: "approve"},
	}
	d := Decide(hardPass(), reviews, nil, true)
	if d.Conclusion != "pass" {
		t.Fatalf("expected pass, got %s (%v)", d.Conclusion, d.Reasons)
	}
}

// Row 10: low/info findings annotate, never block.
func TestLowFindingAnnotatesOnly(t *testing.T) {
	reviews := []Review{
		{Provider: "anthropic", Verdict: "approve", Findings: []Finding{
			{ID: "CLA-010", Fingerprint: "l1", Severity: "low", Category: "style"},
		}},
		{Provider: "openai", Verdict: "approve"},
	}
	d := Decide(hardPass(), reviews, nil, true)
	if d.Conclusion != "pass" {
		t.Fatalf("expected pass, got %s", d.Conclusion)
	}
}

// Row 7: a high finding the other reviewer ACCEPTS (confirmed, no
// reproducer) blocks pending reproducer/human.
func TestHighAcceptedByOtherBlocks(t *testing.T) {
	reviews := []Review{
		{Provider: "anthropic", Verdict: "changes_required", Findings: []Finding{
			{ID: "CLA-011", Fingerprint: "h-acc", Severity: "high", Category: "security"},
		}},
		{Provider: "openai", Verdict: "approve"},
	}
	cross := []CrossResponse{{SourceProvider: "openai", TargetFindingID: "CLA-011", Disposition: "accept"}}
	d := Decide(hardPass(), reviews, cross, true)
	if d.Conclusion != "blocked" {
		t.Fatalf("expected blocked on accepted high, got %s (%v)", d.Conclusion, d.Reasons)
	}
}

// needs_reproducer neither confirms nor clears: a lone high finding with
// that disposition still routes to a human, exactly like unrefuted.
func TestNeedsReproducerRoutesToHuman(t *testing.T) {
	reviews := []Review{
		{Provider: "anthropic", Verdict: "changes_required", Findings: []Finding{
			{ID: "CLA-012", Fingerprint: "h-nr", Severity: "high", Category: "concurrency"},
		}},
		{Provider: "openai", Verdict: "approve"},
	}
	cross := []CrossResponse{{SourceProvider: "openai", TargetFindingID: "CLA-012", Disposition: "needs_reproducer"}}
	d := Decide(hardPass(), reviews, cross, true)
	if d.Conclusion != "pending" || d.ReviewCoverage != "human_required" {
		t.Fatalf("expected pending/human_required, got %s/%s", d.Conclusion, d.ReviewCoverage)
	}
}

// Row 4 (waivers, doc 05 §13): a valid unexpired waiver on the exact
// blocking finding yields pass_with_waiver and records the waived ID.
func TestValidWaiverYieldsPassWithWaiver(t *testing.T) {
	g := hardPass()
	g.Now = "2026-07-14T00:00:00Z"
	g.Waivers = []Waiver{{FindingID: "COD-001", Approver: "Aakash1337", Expiry: "2026-08-01T00:00:00Z"}}
	reviews := []Review{
		{Provider: "openai", Verdict: "changes_required", Findings: []Finding{
			{ID: "COD-001", Fingerprint: "w1", Severity: "high", Category: "security", HasReproducer: true},
		}},
		{Provider: "anthropic", Verdict: "approve"},
	}
	d := Decide(g, reviews, nil, true)
	if d.Conclusion != "pass_with_waiver" {
		t.Fatalf("expected pass_with_waiver, got %s (%v)", d.Conclusion, d.Reasons)
	}
	if len(d.WaivedIDs) != 1 || d.WaivedIDs[0] != "COD-001" || len(d.BlockingIDs) != 0 {
		t.Fatalf("waived=%v blocking=%v", d.WaivedIDs, d.BlockingIDs)
	}
}

// An EXPIRED waiver is not applied — the finding still blocks.
func TestExpiredWaiverStillBlocks(t *testing.T) {
	g := hardPass()
	g.Now = "2026-07-14T00:00:00Z"
	g.Waivers = []Waiver{{FindingID: "COD-001", Approver: "Aakash1337", Expiry: "2026-07-01T00:00:00Z"}}
	reviews := []Review{
		{Provider: "openai", Verdict: "changes_required", Findings: []Finding{
			{ID: "COD-001", Fingerprint: "w2", Severity: "high", Category: "security", HasReproducer: true},
		}},
	}
	d := Decide(g, reviews, nil, true)
	if d.Conclusion != "blocked" {
		t.Fatalf("expected blocked with expired waiver, got %s", d.Conclusion)
	}
}

// A waiver without a named approver is invalid; unparseable expiry fails
// closed. A waiver can never clear a hard-gate failure.
func TestWaiverEdgeCases(t *testing.T) {
	g := hardPass()
	g.Now = "2026-07-14T00:00:00Z"
	g.Waivers = []Waiver{
		{FindingID: "COD-001", Approver: "", Expiry: "2026-08-01T00:00:00Z"}, // unowned
		{FindingID: "COD-002", Approver: "Aakash1337", Expiry: "not-a-date"}, // unparseable
	}
	reviews := []Review{
		{Provider: "openai", Verdict: "changes_required", Findings: []Finding{
			{ID: "COD-001", Fingerprint: "w3", Severity: "high", Category: "security", HasReproducer: true},
			{ID: "COD-002", Fingerprint: "w4", Severity: "high", Category: "security", HasReproducer: true},
		}},
	}
	d := Decide(g, reviews, nil, true)
	if d.Conclusion != "blocked" || len(d.BlockingIDs) != 2 {
		t.Fatalf("invalid waivers must not apply: %s %v", d.Conclusion, d.BlockingIDs)
	}
	// Waivers never rescue a hard-gate failure.
	g2 := GateState{HardGatesPass: false, RequiredEvidence: true, HumanApprovals: true,
		Now:     "2026-07-14T00:00:00Z",
		Waivers: []Waiver{{FindingID: "anything", Approver: "a", Expiry: "2027-01-01T00:00:00Z"}}}
	d2 := Decide(g2, reviews, nil, true)
	if d2.Conclusion != "blocked" {
		t.Fatalf("waiver must not clear a hard gate: %s", d2.Conclusion)
	}
}

// Missing human approval yields pending/human_required even when clean.
func TestMissingHumanApprovalPending(t *testing.T) {
	g := GateState{HardGatesPass: true, RequiredEvidence: true, HumanApprovals: false}
	reviews := []Review{
		{Provider: "anthropic", Verdict: "approve"},
		{Provider: "openai", Verdict: "approve"},
	}
	d := Decide(g, reviews, nil, true)
	if d.Conclusion != "pending" || d.ReviewCoverage != "human_required" {
		t.Fatalf("expected pending/human_required, got %s/%s", d.Conclusion, d.ReviewCoverage)
	}
}
