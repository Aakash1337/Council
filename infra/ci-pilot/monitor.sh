#!/usr/bin/env bash
# Lightweight runner-host health check (Council doc 06 §13, ops runbook
# §4 daily checklist). Runs INSIDE the ci-pilot VM. Prints one JSON line
# suitable for logging; exits non-zero if any threshold is breached.
set -uo pipefail

DISK_PCT=$(df --output=pcent / | tail -1 | tr -dc '0-9')
MEM_PCT=$(free | awk '/Mem:/ {printf "%d", $3/$2*100}')
LOAD=$(cut -d' ' -f1 /proc/loadavg)
RUNNER_UP=false
if systemctl is-active --quiet 'actions.runner.*' 2>/dev/null || pgrep -f Runner.Listener >/dev/null 2>&1; then
  RUNNER_UP=true
fi
GITHUB_OK=false
timeout 8 curl -sS --max-time 8 https://api.github.com/zen >/dev/null 2>&1 && GITHUB_OK=true

printf '{"disk_pct":%s,"mem_pct":%s,"load1":%s,"runner_up":%s,"github_reachable":%s}\n' \
  "$DISK_PCT" "$MEM_PCT" "$LOAD" "$RUNNER_UP" "$GITHUB_OK"

status=0
if [ "$DISK_PCT" -ge 85 ]; then echo "ALERT: disk ${DISK_PCT}% >= 85% (R-039)" >&2; status=1; fi
if [ "$RUNNER_UP" != true ]; then echo "ALERT: runner listener not active" >&2; status=1; fi
if [ "$GITHUB_OK" != true ]; then echo "ALERT: GitHub unreachable" >&2; status=1; fi
exit $status
