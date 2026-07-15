package bundle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFreezeIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "subject/diff.patch", "diff content")
	writeFile(t, dir, "spec/specification.md", "the spec")

	b1, err := Freeze(dir, "run-1", "repo", "CDNS-002", "high", "aaa", "bbb", "specsha", "2026-07-14T00:00:00Z", nil)
	if err != nil {
		t.Fatal(err)
	}
	b2, err := Freeze(dir, "run-1", "repo", "CDNS-002", "high", "aaa", "bbb", "specsha", "2026-07-14T00:00:00Z", nil)
	if err != nil {
		t.Fatal(err)
	}
	if b1.BundleSHA256 != b2.BundleSHA256 {
		t.Fatalf("freeze not deterministic: %s != %s", b1.BundleSHA256, b2.BundleSHA256)
	}
	if b1.BundleSHA256 == "" {
		t.Fatal("empty aggregate hash")
	}
}

func TestVerifyDetectsTamper(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "spec/specification.md", "the spec")
	b, err := Freeze(dir, "run-1", "repo", "CDNS-002", "high", "aaa", "bbb", "specsha", "t", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := b.Write(dir); err != nil {
		t.Fatal(err)
	}
	// Clean verify passes.
	if err := b.Verify(dir); err != nil {
		t.Fatalf("verify should pass on untouched bundle: %v", err)
	}
	// Tamper with a bundled file.
	writeFile(t, dir, "spec/specification.md", "the spec, altered")
	if err := b.Verify(dir); err == nil {
		t.Fatal("verify should fail after tampering with a bundled file")
	}
}

func TestVerifyRejectsAddedInputFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "spec/specification.md", "the spec")
	b, err := Freeze(dir, "r", "repo", "CDNS-002", "high", "a", "b", "s", "t", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := b.Verify(dir); err != nil {
		t.Fatalf("clean verify: %v", err)
	}
	// Add a NEW file under a frozen input tree after freeze.
	writeFile(t, dir, "spec/injected.md", "attacker content")
	if err := b.Verify(dir); err == nil {
		t.Fatal("verify must reject a file added under a frozen input tree")
	}
	// A file under an output tree is allowed.
	dir2 := t.TempDir()
	writeFile(t, dir2, "spec/specification.md", "the spec")
	b2, _ := Freeze(dir2, "r", "repo", "CDNS-002", "high", "a", "b", "s", "t", nil)
	writeFile(t, dir2, "reviews/blind/anthropic.json", "{}")
	if err := b2.Verify(dir2); err != nil {
		t.Fatalf("output-tree files must be allowed: %v", err)
	}
}

func TestBundleHashChangesWithHead(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "spec/specification.md", "the spec")
	b1, _ := Freeze(dir, "r", "repo", "CDNS-002", "high", "aaa", "bbb", "s", "t", nil)
	b2, _ := Freeze(dir, "r", "repo", "CDNS-002", "high", "aaa", "ccc", "s", "t", nil)
	if b1.BundleSHA256 == b2.BundleSHA256 {
		t.Fatal("bundle hash must change when head SHA changes")
	}
}

// Write -> Load must round-trip every field and the loaded bundle must
// verify against the same directory (audit gap: Load/Write untested).
func TestWriteLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "spec/specification.md", "the spec")
	writeFile(t, dir, "subject/diff.patch", "a diff")
	prompts := map[string]string{"review": "prompt-template-hash"}
	b, err := Freeze(dir, "run-rt", "repo", "CDNS-002", "high", "aaa", "bbb", "specsha", "2026-07-14T00:00:00Z", prompts)
	if err != nil {
		t.Fatal(err)
	}
	if err := b.Write(dir); err != nil {
		t.Fatal(err)
	}
	got, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got.RunID != b.RunID || got.HeadSHA != b.HeadSHA || got.BaseSHA != b.BaseSHA ||
		got.BundleSHA256 != b.BundleSHA256 || got.RiskTier != b.RiskTier ||
		got.ChangeID != b.ChangeID || len(got.Files) != len(b.Files) {
		t.Fatalf("round-trip mismatch:\n frozen %+v\n loaded %+v", b, got)
	}
	if got.PromptTemplates["review"] != "prompt-template-hash" {
		t.Fatalf("prompt templates lost in round trip: %+v", got.PromptTemplates)
	}
	if err := got.Verify(dir); err != nil {
		t.Fatalf("loaded bundle must verify: %v", err)
	}
}

// Load fails cleanly on a missing or corrupt bundle.json.
func TestLoadFailsClosed(t *testing.T) {
	if _, err := Load(t.TempDir()); err == nil {
		t.Fatal("Load on empty dir must error")
	}
	dir := t.TempDir()
	writeFile(t, dir, "bundle.json", "{corrupt")
	if _, err := Load(dir); err == nil {
		t.Fatal("Load on corrupt bundle.json must error")
	}
}

// A tampered bundle.json (edited head SHA) breaks the aggregate hash.
func TestLoadedTamperedManifestFailsVerify(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "spec/specification.md", "the spec")
	b, err := Freeze(dir, "r", "repo", "CDNS-002", "high", "aaa", "bbb", "s", "t", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := b.Write(dir); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "bundle.json"))
	if err != nil {
		t.Fatal(err)
	}
	tampered := []byte(strings.Replace(string(raw), `"bbb"`, `"evil"`, 1))
	if string(tampered) == string(raw) {
		t.Fatal("tamper needle not found")
	}
	if err := os.WriteFile(filepath.Join(dir, "bundle.json"), tampered, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := got.Verify(dir); err == nil {
		t.Fatal("verify must fail for a manifest whose fields were edited after freeze")
	}
}
