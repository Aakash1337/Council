package main

// Command-level tests exercising the CLI functions end-to-end in
// process (audit gap: cmd coverage was mostly CLI drills outside
// `go test`). These run the real freeze -> review(mock) -> decide
// pipeline against the committed schemas and fixtures.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoPath resolves a path relative to the broker module root.
func repoPath(t *testing.T, rel string) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join("..", "..", rel))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func freezeBundle(t *testing.T, dir string) {
	t.Helper()
	write(t, filepath.Join(dir, "spec", "specification.md"), "spec content")
	err := cmdFreeze([]string{"--dir", dir, "--run-id", "t-run", "--repo", "r",
		"--change", "CDNS-002", "--risk", "high",
		"--base", "6c9ad4f63d14223c7a449f238b863fcb5e9bb1aa",
		"--head", "9e087cb9e20302af71f672559f8c25af990242bb",
		"--spec-sha", "s", "--created", "2026-07-14T00:00:00Z"})
	if err != nil {
		t.Fatalf("freeze: %v", err)
	}
}

func mockReview(t *testing.T, dir string) {
	t.Helper()
	mock := repoPath(t, "testdata/claude-approve.json") + ":anthropic," +
		repoPath(t, "testdata/codex-highbug.json") + ":openai"
	err := cmdReview([]string{"--dir", dir,
		"--schema", repoPath(t, "schemas/agent-review.schema.json"),
		"--mock", mock})
	if err != nil {
		t.Fatalf("review: %v", err)
	}
}

// The full pipeline: freeze -> mock blind reviews -> decide. The codex
// fixture carries a reproduced high finding, so the decision blocks.
func TestPipelineFreezeReviewDecide(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "bundle")
	freezeBundle(t, dir)
	mockReview(t, dir)

	// Seals must exist outside the bundle.
	if _, err := os.Stat(filepath.Join(dir+"-seals", "seals.json")); err != nil {
		t.Fatalf("seals.json missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "reviews")); !os.IsNotExist(err) {
		t.Fatal("no review may be written inside the bundle dir (blind invariant)")
	}

	err := cmdDecide([]string{"--dir", dir, "--hard-gates-pass=true",
		"--evidence=true", "--human-approvals=true"})
	if err == nil || !strings.Contains(err.Error(), "blocked") {
		t.Fatalf("decision with a reproduced high finding must be blocked, got %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "decision.json"))
	if err != nil {
		t.Fatal(err)
	}
	var d struct {
		Conclusion  string   `json:"conclusion"`
		BlockingIDs []string `json:"blocking_ids"`
	}
	if err := json.Unmarshal(raw, &d); err != nil {
		t.Fatal(err)
	}
	if d.Conclusion != "blocked" || len(d.BlockingIDs) != 1 || d.BlockingIDs[0] != "COD-001" {
		t.Fatalf("decision.json = %+v", d)
	}
}

// Review refuses to run against a tampered bundle (freeze-then-edit).
func TestReviewRefusesTamperedBundle(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "bundle")
	freezeBundle(t, dir)
	write(t, filepath.Join(dir, "spec", "specification.md"), "edited after freeze")
	mock := repoPath(t, "testdata/claude-approve.json") + ":anthropic," +
		repoPath(t, "testdata/codex-highbug.json") + ":openai"
	err := cmdReview([]string{"--dir", dir,
		"--schema", repoPath(t, "schemas/agent-review.schema.json"),
		"--mock", mock})
	if err == nil {
		t.Fatal("review must refuse a bundle that fails verification")
	}
}

// Decide with a head-scoped waiver over the real pipeline yields
// pass_with_waiver (the head SHA comes from bundle.json).
func TestPipelineDecideWithWaiver(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "bundle")
	freezeBundle(t, dir)
	mockReview(t, dir)
	waivers := filepath.Join(t.TempDir(), "waivers.json")
	write(t, waivers, `[{"finding_id":"COD-001","head_sha":"9e087cb9e20302af71f672559f8c25af990242bb","approver":"Aakash1337","expiry":"2026-08-01T00:00:00Z"}]`)
	err := cmdDecide([]string{"--dir", dir, "--hard-gates-pass=true",
		"--evidence=true", "--human-approvals=true",
		"--waivers-file", waivers, "--now", "2026-07-14T12:00:00Z"})
	if err != nil {
		t.Fatalf("waived decision must succeed: %v", err)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "decision.json"))
	if !strings.Contains(string(raw), `"pass_with_waiver"`) {
		t.Fatalf("decision.json = %s", raw)
	}
}

// Post-seal tampering with a sealed report is detected by decide.
func TestDecideDetectsSealTamper(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "bundle")
	freezeBundle(t, dir)
	mockReview(t, dir)
	sealed := filepath.Join(dir+"-seals", "blind", "openai.json")
	raw, err := os.ReadFile(sealed)
	if err != nil {
		t.Fatal(err)
	}
	write(t, sealed, strings.Replace(string(raw), "changes_required", "approve", 1))
	err = cmdDecide([]string{"--dir", dir, "--hard-gates-pass=true",
		"--evidence=true", "--human-approvals=true"})
	if err == nil || !strings.Contains(err.Error(), "seal") {
		t.Fatalf("post-seal tamper must fail seal verification, got %v", err)
	}
}
