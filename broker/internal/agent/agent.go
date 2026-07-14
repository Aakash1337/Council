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
	"fmt"
	"os"
	"os/exec"
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
	return &Result{Provider: req.Provider, RawJSON: stdout.Bytes(), Client: client, DurationMs: dur, ExitCode: exit}, nil
}

// reviewEnv passes through only what a review session needs, and never
// injects provider API keys (subscription auth lives in the client).
func reviewEnv() []env {
	keep := []string{"PATH", "HOME", "USERPROFILE", "SystemRoot", "TEMP", "TMP", "CLAUDE_CONFIG_DIR", "CODEX_HOME"}
	var out []env
	for _, k := range keep {
		if v, ok := os.LookupEnv(k); ok {
			out = append(out, env{k, v})
		}
	}
	return out
}

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
