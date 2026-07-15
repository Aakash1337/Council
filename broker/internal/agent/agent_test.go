package agent

import (
	"context"
	"strings"
	"testing"
)

// The review env forwards the subscription OAuth token (how the
// first-party client authenticates — doc 04 §4.2) but NEVER provider API
// keys (the prohibited paid path, ADR-012) or write tokens (merge
// authority a reviewer must not hold).
func TestReviewEnvBoundary(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "sk-test-leak")
	t.Setenv("OPENAI_API_KEY", "sk-test-leak")
	t.Setenv("GITHUB_TOKEN", "ghp-test-leak")
	t.Setenv("GH_TOKEN", "ghp-test-leak-2")
	t.Setenv("CLAUDE_CODE_OAUTH_TOKEN", "tok-subscription")
	t.Setenv("PATH", "/usr/bin")

	joined := strings.Join(envSlice(reviewEnv()), "\n")
	for _, forbidden := range forbiddenReviewEnv {
		if strings.Contains(joined, forbidden+"=") {
			t.Errorf("review env must NOT contain %s (paid/write credential)", forbidden)
		}
	}
	if !strings.Contains(joined, "CLAUDE_CODE_OAUTH_TOKEN=") {
		t.Error("subscription OAuth token must be forwarded — it is how the client authenticates")
	}
	if !strings.Contains(joined, "PATH=") {
		t.Error("allowlisted PATH must pass through")
	}
}

// ExtractReviewJSON handles the three real output shapes and fails
// closed on garbage.
func TestExtractReviewJSON(t *testing.T) {
	want := `{"schema_version":1,"verdict":"approve"}`

	// a) Claude envelope with a fenced inner review.
	env := `{"type":"result","subtype":"success","is_error":false,"result":"` + "```json\\n" + `{\"schema_version\":1,\"verdict\":\"approve\"}` + "\\n```" + `"}`
	got, err := ExtractReviewJSON([]byte(env))
	if err != nil || string(got) != want {
		t.Fatalf("envelope+fence: got %q err %v", got, err)
	}

	// b) Bare fenced JSON (codex style).
	got, err = ExtractReviewJSON([]byte("```json\n" + want + "\n```"))
	if err != nil || string(got) != want {
		t.Fatalf("bare fence: got %q err %v", got, err)
	}

	// c) Plain JSON object with trailing prose.
	got, err = ExtractReviewJSON([]byte(want + "\n\nHope that helps!"))
	if err != nil || string(got) != want {
		t.Fatalf("plain+prose: got %q err %v", got, err)
	}

	// d) A Claude envelope reporting an error must fail closed.
	if _, err := ExtractReviewJSON([]byte(`{"type":"result","is_error":true,"subtype":"error_auth","result":"nope"}`)); err == nil {
		t.Fatal("error envelope must fail closed")
	}

	// e) No JSON at all.
	if _, err := ExtractReviewJSON([]byte("total failure, no json here")); err == nil {
		t.Fatal("no-JSON output must error")
	}

	// f) Braces inside strings must not confuse the balancer.
	tricky := `{"claim":"the map { has a brace } inside","verdict":"x"}`
	got, err = ExtractReviewJSON([]byte(tricky))
	if err != nil || string(got) != tricky {
		t.Fatalf("string-brace: got %q err %v", got, err)
	}
}

// MockRunner returns the exact fixture bytes for the requested provider
// and errors (never fabricates) for an unknown one.
func TestMockRunner(t *testing.T) {
	m := &MockRunner{Fixtures: map[Provider][]byte{Anthropic: []byte(`{"x":1}`)}}
	res, err := m.Review(context.Background(), Request{Provider: Anthropic})
	if err != nil || string(res.RawJSON) != `{"x":1}` || res.Client != "claude-code" {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	if _, err := m.Review(context.Background(), Request{Provider: OpenAI}); err == nil {
		t.Fatal("missing fixture must be an error, not an empty review")
	}
}

// CLIRunner surfaces a launch failure as an error (never a silent empty
// result that could be mistaken for a review).
func TestCLIRunnerLaunchFailureIsError(t *testing.T) {
	c := &CLIRunner{ClaudeBin: "definitely-not-a-real-binary-xyz"}
	if _, err := c.Review(context.Background(), Request{Provider: Anthropic, TimeoutSecs: 5}); err == nil {
		t.Fatal("nonexistent client binary must produce an error")
	}
}
