// Package bundle creates and validates the content-addressed review
// bundle the broker freezes before any agent runs (Council doc 04 §6).
// The bundle binds a candidate diff, specification, and deterministic
// evidence to a set of hashes; if any input changes, the bundle is
// invalidated and review must restart.
package bundle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Bundle is the frozen, hashed input set for a review run.
type Bundle struct {
	RunID           string            `json:"run_id"`
	Repository      string            `json:"repository"`
	ChangeID        string            `json:"change_id"`
	RiskTier        string            `json:"risk_tier"`
	BaseSHA         string            `json:"base_sha"`
	HeadSHA         string            `json:"head_sha"`
	SpecSHA256      string            `json:"spec_sha256"`
	BundleSHA256    string            `json:"bundle_sha256"`
	PromptTemplates map[string]string `json:"prompt_template_sha256"`
	Files           map[string]string `json:"files"` // relpath -> sha256
	CreatedAt       string            `json:"created_at"`
}

// FileSHA256 returns the hex SHA-256 of a file's contents.
func FileSHA256(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

// hashString returns the hex SHA-256 of a string.
func hashString(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// Freeze walks the bundle directory, hashes every file deterministically
// (sorted by relative POSIX path), and computes an aggregate bundle hash
// that changes if any input changes. createdAt is passed in (not read
// from the clock) so the operation is reproducible in tests.
func Freeze(dir, runID, repository, changeID, riskTier, baseSHA, headSHA, specSHA256, createdAt string, promptTemplates map[string]string) (*Bundle, error) {
	b := &Bundle{
		RunID:           runID,
		Repository:      repository,
		ChangeID:        changeID,
		RiskTier:        riskTier,
		BaseSHA:         baseSHA,
		HeadSHA:         headSHA,
		SpecSHA256:      specSHA256,
		PromptTemplates: promptTemplates,
		Files:           map[string]string{},
		CreatedAt:       createdAt,
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		// bundle.json is written after freeze; never hash it into itself.
		if rel == "bundle.json" {
			return nil
		}
		sum, err := FileSHA256(path)
		if err != nil {
			return err
		}
		b.Files[rel] = sum
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk bundle dir: %w", err)
	}

	b.BundleSHA256 = b.computeAggregate()
	return b, nil
}

// computeAggregate hashes the sorted (path, filehash) pairs plus the
// pinned identity fields, so the bundle hash is stable and total.
func (b *Bundle) computeAggregate() string {
	paths := make([]string, 0, len(b.Files))
	for p := range b.Files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	var sb strings.Builder
	fmt.Fprintf(&sb, "change=%s\nbase=%s\nhead=%s\nspec=%s\nrisk=%s\n",
		b.ChangeID, b.BaseSHA, b.HeadSHA, b.SpecSHA256, b.RiskTier)
	for _, p := range paths {
		fmt.Fprintf(&sb, "%s=%s\n", p, b.Files[p])
	}
	tmplKeys := make([]string, 0, len(b.PromptTemplates))
	for k := range b.PromptTemplates {
		tmplKeys = append(tmplKeys, k)
	}
	sort.Strings(tmplKeys)
	for _, k := range tmplKeys {
		fmt.Fprintf(&sb, "tmpl:%s=%s\n", k, b.PromptTemplates[k])
	}
	return hashString(sb.String())
}

// Write persists bundle.json into the bundle directory.
func (b *Bundle) Write(dir string) error {
	out, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "bundle.json"), out, 0o644)
}

// frozenInputPrefixes are the immutable-input trees an agent reads. A
// file appearing under one of these after freeze (not just a changed
// file) breaks the frozen-bundle guarantee and must be rejected.
var frozenInputPrefixes = []string{"subject/", "spec/", "evidence/"}

// Verify recomputes every file hash and the aggregate, and rejects any
// file added under a frozen input tree after freeze. This is the tamper
// check run before consuming a bundle (doc 04 §6, §9.3).
func (b *Bundle) Verify(dir string) error {
	// 1. Every recorded file must still match its frozen hash.
	for rel, want := range b.Files {
		got, err := FileSHA256(filepath.Join(dir, filepath.FromSlash(rel)))
		if err != nil {
			return fmt.Errorf("missing bundled file %s: %w", rel, err)
		}
		if got != want {
			return fmt.Errorf("bundle file %s changed: have %s, want %s", rel, got, want)
		}
	}
	// 2. No unexpected file may appear under a frozen input tree. Output
	//    trees (reviews/, decisions/, repairs/) and bundle.json are written
	//    after freeze and are allowed.
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if _, known := b.Files[rel]; known {
			return nil
		}
		for _, p := range frozenInputPrefixes {
			if strings.HasPrefix(rel, p) {
				return fmt.Errorf("unexpected file added under frozen input after freeze: %s", rel)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if agg := b.computeAggregate(); agg != b.BundleSHA256 {
		return fmt.Errorf("bundle aggregate mismatch: have %s, want %s", agg, b.BundleSHA256)
	}
	return nil
}

// Load reads a previously written bundle.json.
func Load(dir string) (*Bundle, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "bundle.json"))
	if err != nil {
		return nil, err
	}
	var b Bundle
	if err := json.Unmarshal(raw, &b); err != nil {
		return nil, err
	}
	return &b, nil
}
