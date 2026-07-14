#!/usr/bin/env bash
# Credential/LAN red-team test (Council ACT-004, PAC-008/PAC-010, exit
# gate G3). Runs INSIDE the ci-pilot VM. Proves the runner cannot reach
# home infrastructure and holds no AI/deployment credentials.
#
# Exit 0 = all isolation assertions hold. Any reachable target or found
# credential is a FAIL.
set -uo pipefail

FAILURES=0
pass() { echo "PASS  $1"; }
fail() { echo "FAIL  $1"; FAILURES=$((FAILURES+1)); }

echo "== 1. Home LAN must be unreachable =="
# Common home gateway/NAS addresses. Reachability is a failure.
for target in 192.168.1.1 192.168.0.1 192.168.1.10 10.0.0.1; do
  if timeout 3 bash -c "echo > /dev/tcp/${target}/443" 2>/dev/null; then
    fail "reached LAN host ${target}:443 (should be denied)"
  else
    pass "LAN host ${target} unreachable"
  fi
done

echo "== 2. Cloud metadata endpoint must be unreachable =="
if timeout 3 curl -s --max-time 3 http://169.254.169.254/ >/dev/null 2>&1; then
  fail "cloud metadata endpoint reachable"
else
  pass "cloud metadata endpoint unreachable"
fi

echo "== 3. GitHub must be reachable (runner needs it) =="
if timeout 8 curl -sS --max-time 8 https://api.github.com/zen >/dev/null 2>&1; then
  pass "api.github.com reachable"
else
  fail "api.github.com unreachable (runner cannot function)"
fi

echo "== 4. No AI/deployment credentials on this host =="
for path in \
  "$HOME/.claude/.credentials.json" \
  "$HOME/.codex/auth.json" \
  "$HOME/.config/gh/hosts.yml" \
  "$HOME/.aws/credentials" \
  "$HOME/.ssh/id_rsa" \
  "$HOME/.ssh/id_ed25519"; do
  if [ -e "$path" ]; then
    fail "credential-like file present: $path"
  else
    pass "absent: $path"
  fi
done

echo "== 5. CLAUDE_CODE_OAUTH_TOKEN / provider env must be empty =="
for var in CLAUDE_CODE_OAUTH_TOKEN ANTHROPIC_API_KEY OPENAI_API_KEY CODEX_HOME; do
  if [ -n "${!var:-}" ]; then
    fail "provider env set: $var"
  else
    pass "unset: $var"
  fi
done

echo ""
if [ "$FAILURES" -eq 0 ]; then
  echo "RESULT: pass — runner isolation verified"
  exit 0
else
  echo "RESULT: FAIL ($FAILURES isolation violations)"
  exit 1
fi
