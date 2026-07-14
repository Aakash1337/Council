package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Aakash1337/Council/broker/internal/bundle"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// verifySeals detects a post-seal edit and fails closed; intact seals
// verify; absent seals.json (agent-unavailable run) is not an error.
func TestVerifySeals(t *testing.T) {
	base := t.TempDir()
	report := `{"review_id":"r1"}`
	write(t, filepath.Join(base, "blind", "openai.json"), report)
	sum, err := bundle.FileSHA256(filepath.Join(base, "blind", "openai.json"))
	if err != nil {
		t.Fatal(err)
	}
	seals, _ := json.Marshal(map[string]string{"openai.json": sum})
	write(t, filepath.Join(base, "seals.json"), string(seals))

	if err := verifySeals(base); err != nil {
		t.Fatalf("intact seals must verify: %v", err)
	}
	// Tamper.
	write(t, filepath.Join(base, "blind", "openai.json"), `{"review_id":"edited"}`)
	if err := verifySeals(base); err == nil {
		t.Fatal("post-seal edit must fail verification")
	}
	// Missing sealed file.
	os.Remove(filepath.Join(base, "blind", "openai.json"))
	if err := verifySeals(base); err == nil {
		t.Fatal("missing sealed report must fail verification")
	}
	// No seals.json at all -> nothing to verify (agent-unavailable path).
	if err := verifySeals(t.TempDir()); err != nil {
		t.Fatalf("absent seals.json must not error: %v", err)
	}
}

// loadCross returns both dispositions and new_findings (the review fix:
// a cross-review-discovered blocker must not be dropped).
func TestLoadCrossCarriesNewFindings(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "codex.json"), `{
		"provider":"openai",
		"responses":[{"target_finding_id":"CLA-001","disposition":"reject"}],
		"new_findings":[{"id":"XR-1","fingerprint":"xf","severity":"high","category":"security",
			"acceptance_ids":[],"claim":"c","reproducer":{"kind":"test","ref":"t"}}]
	}`)
	resp, newFindings := loadCross(dir)
	if len(resp) != 1 || resp[0].TargetFindingID != "CLA-001" || resp[0].Disposition != "reject" {
		t.Fatalf("responses = %+v", resp)
	}
	if len(newFindings) != 1 || newFindings[0].ID != "XR-1" || !newFindings[0].HasReproducer {
		t.Fatalf("new findings = %+v", newFindings)
	}
}

// loadReviews maps sealed reports into arbiter reviews, including the
// reproducer-presence bit the precedence rules depend on.
func TestLoadReviews(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "openai.json"), `{
		"review_id":"r-codex","provider":"openai","verdict":"changes_required",
		"findings":[{"id":"F1","fingerprint":"fp","severity":"high","category":"concurrency",
			"acceptance_ids":["CDNS-002/AC-006"],"claim":"race",
			"reproducer":{"kind":"test","ref":"test://x"}}]
	}`)
	reviews, err := loadReviews(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(reviews) != 1 || reviews[0].Provider != "openai" {
		t.Fatalf("reviews = %+v", reviews)
	}
	f := reviews[0].Findings[0]
	if f.ID != "F1" || !f.HasReproducer || f.AcceptanceID[0] != "CDNS-002/AC-006" {
		t.Fatalf("finding = %+v", f)
	}
}

// loadWaivers parses a waivers file and rejects malformed input.
func TestLoadWaivers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "waivers.json")
	write(t, path, `[{"finding_id":"COD-001","approver":"Aakash1337","expiry":"2026-08-01T00:00:00Z"}]`)
	ws, err := loadWaivers(path)
	if err != nil || len(ws) != 1 || ws[0].FindingID != "COD-001" {
		t.Fatalf("ws=%+v err=%v", ws, err)
	}
	write(t, path, `{not json`)
	if _, err := loadWaivers(path); err == nil {
		t.Fatal("malformed waivers file must error")
	}
	if ws, err := loadWaivers(""); err != nil || ws != nil {
		t.Fatalf("empty path means no waivers: ws=%v err=%v", ws, err)
	}
}
