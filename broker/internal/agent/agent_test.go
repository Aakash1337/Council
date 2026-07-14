package agent

import (
	"context"
	"strings"
	"testing"
)

// The broker must never inject provider API keys or GitHub tokens into a
// review subprocess: subscription auth lives inside the first-party
// client's own state, and the review env is a strict allowlist
// (doc 04 §4.2, FR-RUN-002).
func TestReviewEnvIsAllowlistOnly(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "sk-test-leak")
	t.Setenv("OPENAI_API_KEY", "sk-test-leak")
	t.Setenv("GITHUB_TOKEN", "ghp-test-leak")
	t.Setenv("CLAUDE_CODE_OAUTH_TOKEN", "tok-test-leak")
	t.Setenv("PATH", "/usr/bin") // allowlisted — must pass through

	env := envSlice(reviewEnv())
	joined := strings.Join(env, "\n")
	for _, forbidden := range []string{"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GITHUB_TOKEN", "CLAUDE_CODE_OAUTH_TOKEN"} {
		if strings.Contains(joined, forbidden) {
			t.Errorf("review env must not contain %s", forbidden)
		}
	}
	if !strings.Contains(joined, "PATH=") {
		t.Error("allowlisted PATH must pass through")
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
