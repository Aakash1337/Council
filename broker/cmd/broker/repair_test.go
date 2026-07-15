package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func gate(t *testing.T, ledger, change, head string, extra ...string) error {
	t.Helper()
	args := []string{"--ledger", ledger, "--change", change, "--from-head", head,
		"--started-at", "2026-07-14T00:00:00Z"}
	args = append(args, extra...)
	return cmdRepairGate(args)
}

// The full ADR-008 lifecycle: two cycles authorized, the third refused
// with needs_human (PAC-007).
func TestRepairGateTwoCycleCap(t *testing.T) {
	ledger := filepath.Join(t.TempDir(), "CDNS-002.repair.json")
	if err := gate(t, ledger, "CDNS-002", "head-1", "--reason", "COD-001"); err != nil {
		t.Fatalf("cycle 1 must be authorized: %v", err)
	}
	if err := gate(t, ledger, "CDNS-002", "head-2", "--reason", "COD-001"); err != nil {
		t.Fatalf("cycle 2 must be authorized: %v", err)
	}
	err := gate(t, ledger, "CDNS-002", "head-3")
	if err == nil || !strings.Contains(err.Error(), "needs_human") {
		t.Fatalf("cycle 3 must be refused with needs_human, got %v", err)
	}
}

// Repairing the same failed head twice is a no-progress loop and is
// refused even under the cycle cap (doc 08 §13).
func TestRepairGateNoProgressLoop(t *testing.T) {
	ledger := filepath.Join(t.TempDir(), "l.json")
	if err := gate(t, ledger, "CDNS-002", "head-1"); err != nil {
		t.Fatal(err)
	}
	err := gate(t, ledger, "CDNS-002", "head-1")
	if err == nil || !strings.Contains(err.Error(), "no-progress") {
		t.Fatalf("same-head repair must be refused as no-progress, got %v", err)
	}
}

// ADR-008 is a ceiling: --max above 2 is refused outright, and a
// corrupt ledger fails closed rather than resetting the count.
func TestRepairGateCeilingAndCorruptLedger(t *testing.T) {
	ledger := filepath.Join(t.TempDir(), "l.json")
	if err := gate(t, ledger, "C-1", "h1", "--max", "3"); err == nil {
		t.Fatal("--max 3 must be refused (ADR-008 ceiling)")
	}
	write(t, ledger, "{corrupt")
	if err := gate(t, ledger, "C-1", "h1"); err == nil || !strings.Contains(err.Error(), "fail closed") {
		t.Fatalf("corrupt ledger must fail closed, got %v", err)
	}
}

// One ledger belongs to one change; cross-change reuse is refused.
func TestRepairGateLedgerScopedToChange(t *testing.T) {
	ledger := filepath.Join(t.TempDir(), "l.json")
	if err := gate(t, ledger, "CDNS-002", "h1"); err != nil {
		t.Fatal(err)
	}
	if err := gate(t, ledger, "CDNS-999", "h2"); err == nil {
		t.Fatal("ledger reuse across changes must be refused")
	}
}

// --max 1 (the doc 04 §9.2 default of one pass) is honored.
func TestRepairGateSingleCycleDefaultPolicy(t *testing.T) {
	ledger := filepath.Join(t.TempDir(), "l.json")
	if err := gate(t, ledger, "C-1", "h1", "--max", "1"); err != nil {
		t.Fatal(err)
	}
	if err := gate(t, ledger, "C-1", "h2", "--max", "1"); err == nil {
		t.Fatal("second cycle must be refused under --max 1")
	}
}
