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
