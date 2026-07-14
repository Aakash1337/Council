package bundle

import (
	"os"
	"path/filepath"
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
