// Package agent abstracts the blind reviewers. A Runner produces a
// schema-valid review report for a frozen bundle. Two implementations:
// CLIRunner invokes the real first-party clients (claude -p / codex
// exec); MockRunner replays fixture reports for deterministic tests and
// offline broker development. The broker treats every Runner output as
// untrusted until schema-validated (Council doc 04 §2, §8).
package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Provider identifies the reviewing model family.
type Provider string

const (
	Anthropic Provider = "anthropic"
	OpenAI    Provider = "openai"
)

// Request is the immutable input to one blind review.
type Request struct {
	Provider    Provider
	Prompt      string // the assembled review prompt (bundle already frozen)
	BundleDir   string
	TimeoutSecs int
	MaxTurns    int
}

// Result is the raw (unvalidated) review output plus invocation metadata.
type Result struct {
	Provider   Provider
	RawJSON    []byte
	Client     string
	DurationMs int64
	ExitCode   int
}

// Runner produces one review report.
type Runner interface {
	Review(ctx context.Context, req Request) (*Result, error)
}

// MockRunner returns a fixture keyed by provider. Used for deterministic
// arbiter/flow tests and offline development so the broker never spends
// subscription capacity during CI of the broker itself.
type MockRunner struct {
	Fixtures map[Provider][]byte
}

func (m *MockRunner) Review(_ context.Context, req Request) (*Result, error) {
	raw, ok := m.Fixtures[req.Provider]
	if !ok {
		return nil, fmt.Errorf("no mock fixture for provider %s", req.Provider)
	}
	client := "claude-code"
	if req.Provider == OpenAI {
		client = "codex-cli"
	}
	return &Result{Provider: req.Provider, RawJSON: raw, Client: client, DurationMs: 0, ExitCode: 0}, nil
}

// CLIRunner invokes the real first-party clients. It never reads or
// relays credentials; it launches the client, which owns auth (doc 04
// §4.2). Prompt goes on stdin; schema-constrained JSON is expected on
// stdout.
type CLIRunner struct {
	// ClaudeArgs / CodexArgs allow the operator to pin flags (model,
	// output format, max turns) without changing broker code.
	ClaudeBin  string
	ClaudeArgs []string
	CodexBin   string
	CodexArgs  []string
}

func (c *CLIRunner) Review(ctx context.Context, req Request) (*Result, error) {
	var bin string
	var args []string
	client := "claude-code"
	switch req.Provider {
	case Anthropic:
		bin = orDefault(c.ClaudeBin, "claude")
		args = c.ClaudeArgs
	case OpenAI:
		bin = orDefault(c.CodexBin, "codex")
		args = c.CodexArgs
		client = "codex-cli"
	default:
		return nil, fmt.Errorf("unknown provider %s", req.Provider)
	}

	if req.TimeoutSecs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.TimeoutSecs)*time.Second)
		defer cancel()
	}

	start := time.Now()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdin = bytes.NewReader([]byte(req.Prompt))
	// Child commands get no readable credentials and a clean cwd inside
	// the bundle (doc 04 §5.1). The broker does not export tokens here.
	cmd.Dir = req.BundleDir
	cmd.Env = envSlice(reviewEnv())
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	dur := time.Since(start).Milliseconds()
	exit := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exit = ee.ExitCode()
		} else {
			return nil, fmt.Errorf("%s launch failed: %v: %s", bin, err, stderr.String())
		}
	}
	review, err := ExtractReviewJSON(stdout.Bytes())
	if err != nil {
		return nil, fmt.Errorf("%s produced no extractable review JSON (exit %d): %v: %s",
			bin, exit, err, truncate(stderr.String(), 200))
	}
	return &Result{Provider: req.Provider, RawJSON: review, Client: client, DurationMs: dur, ExitCode: exit}, nil
}

// ExtractReviewJSON pulls the review object out of a first-party client's
// output. This is transport normalization only (doc 04 §11) — it never
// rewrites substantive findings:
//   - `claude -p --output-format json` wraps the answer in an envelope
//     whose `result` string holds the review;
//   - both clients may fence the JSON in ```json ... ``` markdown;
//   - otherwise the output already is the review object.
//
// It returns the first balanced top-level JSON object found.
func ExtractReviewJSON(raw []byte) ([]byte, error) {
	s := string(raw)
	// 1. Claude envelope: {"type":"result",...,"result":"<review-or-fenced>"}.
	var env struct {
		Result  string `json:"result"`
		IsError bool   `json:"is_error"`
		Subtype string `json:"subtype"`
	}
	if err := json.Unmarshal(raw, &env); err == nil && env.Result != "" {
		if env.IsError {
			return nil, fmt.Errorf("client reported error (subtype %q)", env.Subtype)
		}
		s = env.Result
	}
	// 2. Strip a leading/trailing markdown code fence if present.
	s = stripFence(s)
	// 3. Take the first balanced JSON object.
	obj, err := firstJSONObject(s)
	if err != nil {
		return nil, err
	}
	// 4. Confirm it parses.
	var probe map[string]any
	if err := json.Unmarshal(obj, &probe); err != nil {
		return nil, fmt.Errorf("extracted text is not valid JSON: %w", err)
	}
	return obj, nil
}

func stripFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		if i := strings.IndexByte(s, '\n'); i >= 0 {
			s = s[i+1:]
		}
		if j := strings.LastIndex(s, "```"); j >= 0 {
			s = s[:j]
		}
	}
	return strings.TrimSpace(s)
}

// firstJSONObject returns the first balanced {...} run, ignoring braces
// inside strings. Fails closed if none is balanced.
func firstJSONObject(s string) ([]byte, error) {
	start := strings.IndexByte(s, '{')
	if start < 0 {
		return nil, fmt.Errorf("no JSON object found in output")
	}
	depth, inStr, esc := 0, false, false
	for i := start; i < len(s); i++ {
		c := s[i]
		switch {
		case esc:
			esc = false
		case c == '\\' && inStr:
			esc = true
		case c == '"':
			inStr = !inStr
		case inStr:
		case c == '{':
			depth++
		case c == '}':
			depth--
			if depth == 0 {
				return []byte(s[start : i+1]), nil
			}
		}
	}
	return nil, fmt.Errorf("unbalanced JSON object in output")
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// reviewEnv builds the environment for a review subprocess. The security
// boundary (doc 04 §4.2, §5.1; ADR-012):
//   - the subscription OAuth token (CLAUDE_CODE_OAUTH_TOKEN) IS forwarded
//     verbatim by name — it is how the first-party client authenticates
//     the subscription, and the whole subscription-funded design depends
//     on it. The broker never reads or logs its value; it forwards the
//     env entry, treating the client as the auth boundary.
//   - provider API keys (ANTHROPIC_API_KEY, OPENAI_API_KEY) and write
//     credentials (GITHUB_TOKEN) are NEVER forwarded: API keys are the
//     prohibited paid path (ADR-012) and a GitHub token is merge/write
//     authority a reviewer must never hold.
//
// Residual risk (R-001/R-004): a review session can read its own env, so
// the forwarded subscription token is in principle reachable by a
// successful prompt injection in the candidate. Compensating controls:
// review sessions are read-only and egress-restricted to provider
// endpoints (doc 04 §5.1); the token is inference-scoped (cannot merge,
// deploy, or establish control sessions). Tracked, not eliminated.
func reviewEnv() []env {
	keep := []string{
		"PATH", "HOME", "USERPROFILE", "SystemRoot", "TEMP", "TMP",
		"CLAUDE_CONFIG_DIR", "CODEX_HOME",
		"CLAUDE_CODE_OAUTH_TOKEN", // subscription auth — forwarded, never read/logged
	}
	var out []env
	for _, k := range keep {
		if v, ok := os.LookupEnv(k); ok {
			out = append(out, env{k, v})
		}
	}
	return out
}

// forbiddenReviewEnv lists credentials that must NEVER reach a review
// subprocess (asserted by tests): API keys (paid path) and write tokens.
var forbiddenReviewEnv = []string{"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GITHUB_TOKEN", "GH_TOKEN"}

type env struct{ k, v string }

func envSlice(in []env) []string {
	s := make([]string, 0, len(in))
	for _, e := range in {
		s = append(s, e.k+"="+e.v)
	}
	return s
}

func orDefault(v, d string) string {
	if v == "" {
		return d
	}
	return v
}
