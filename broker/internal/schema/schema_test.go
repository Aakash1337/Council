package schema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func reviewSchema(t *testing.T) *Validator {
	t.Helper()
	v, err := New(filepath.Join("..", "..", "schemas", "agent-review.schema.json"))
	if err != nil {
		t.Fatalf("compile agent-review schema: %v", err)
	}
	return v
}

// fixture loads a committed single-line JSON review fixture.
func fixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// mutate replaces old with new and fails the test if old is absent, so a
// fixture reformat can't silently turn a negative test into a no-op.
func mutate(t *testing.T, doc, old, new string) []byte {
	t.Helper()
	if !strings.Contains(doc, old) {
		t.Fatalf("fixture does not contain %q — update the test", old)
	}
	return []byte(strings.Replace(doc, old, new, 1))
}

// The committed fixtures must validate — they anchor the whole mock lane.
func TestFixturesValidate(t *testing.T) {
	v := reviewSchema(t)
	for _, f := range []string{"claude-approve.json", "codex-highbug.json"} {
		if err := v.ValidateBytes([]byte(fixture(t, f))); err != nil {
			t.Errorf("%s should validate: %v", f, err)
		}
	}
}

// Controlled vocabularies are enforced: an out-of-enum verdict fails.
func TestInvalidVerdictRejected(t *testing.T) {
	v := reviewSchema(t)
	bad := mutate(t, fixture(t, "claude-approve.json"), `"verdict":"approve"`, `"verdict":"request_changes"`)
	if err := v.ValidateBytes(bad); err == nil {
		t.Fatal("out-of-enum verdict must be rejected")
	}
}

// The ECMA-262 lookahead patterns work through the regexp2 engine: a
// repo path containing backslashes must be rejected by the repoPath
// pattern (which uses lookaheads Go's RE2 cannot compile).
func TestBackslashPathRejectedByPattern(t *testing.T) {
	v := reviewSchema(t)
	bad := mutate(t, fixture(t, "codex-highbug.json"),
		`"path":"internal/resolver/cache.go"`, `"path":"internal\\\\resolver\\\\cache.go"`)
	if err := v.ValidateBytes(bad); err == nil {
		t.Fatal("backslash repo path must be rejected (regexp2 lookahead)")
	}
}

// Malformed JSON is an error, not a pass.
func TestMalformedJSONRejected(t *testing.T) {
	v := reviewSchema(t)
	if err := v.ValidateBytes([]byte("{not json")); err == nil {
		t.Fatal("malformed JSON must be rejected")
	}
}

// Unknown top-level fields are rejected (additionalProperties: false).
func TestUnknownTopLevelFieldRejected(t *testing.T) {
	v := reviewSchema(t)
	bad := mutate(t, fixture(t, "claude-approve.json"),
		`"schema_version":1`, `"schema_version":1,"sneaky_extra":true`)
	if err := v.ValidateBytes(bad); err == nil {
		t.Fatal("unknown top-level field must be rejected")
	}
}

// Wrong provider/client pairing violates the conditional schema.
func TestProviderClientPairingEnforced(t *testing.T) {
	v := reviewSchema(t)
	bad := mutate(t, fixture(t, "claude-approve.json"),
		`"client":"claude-code"`, `"client":"codex-cli"`)
	if err := v.ValidateBytes(bad); err == nil {
		t.Fatal("anthropic provider with codex-cli client must be rejected")
	}
}
