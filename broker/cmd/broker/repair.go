package main

// repair-gate enforces the bounded-repair invariant (Council ADR-008,
// doc 04 §9.2, PAC-007): one repair pass by default, an absolute
// maximum of two, then mandatory human adjudication. The counter is a
// per-change ledger keyed by change ID — a repair produces a NEW head
// SHA and a NEW bundle, so the ledger deliberately lives outside any
// single bundle directory and survives across candidates of the same
// change.

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type repairLedger struct {
	ChangeID  string        `json:"change_id"`
	MaxCycles int           `json:"max_cycles"`
	Cycles    []repairCycle `json:"cycles"`
}

type repairCycle struct {
	Number    int    `json:"number"`
	FromHead  string `json:"from_head"` // candidate that failed and is being repaired
	Reason    string `json:"reason"`    // blocking finding IDs or free text
	StartedAt string `json:"started_at"`
}

// cmdRepairGate authorizes (or refuses) the next repair cycle.
// Exit codes: 0 = cycle authorized and recorded; process error = refused.
func cmdRepairGate(args []string) error {
	fs := flag.NewFlagSet("repair-gate", flag.ExitOnError)
	ledgerPath := fs.String("ledger", "", "repair ledger JSON path (per change)")
	change := fs.String("change", "", "change ID")
	fromHead := fs.String("from-head", "", "head SHA of the failed candidate being repaired")
	reason := fs.String("reason", "", "blocking finding IDs / reason")
	maxCycles := fs.Int("max", 2, "absolute repair-cycle maximum (ADR-008: never above 2)")
	startedAt := fs.String("started-at", "", "RFC3339 timestamp (injected for determinism)")
	_ = fs.Parse(args)
	if *ledgerPath == "" || *change == "" || *fromHead == "" {
		return fmt.Errorf("--ledger, --change, and --from-head are required")
	}
	// ADR-008 is a ceiling, not a knob: refuse configs that raise it.
	if *maxCycles > 2 {
		return fmt.Errorf("--max %d exceeds the ADR-008 absolute maximum of 2", *maxCycles)
	}
	if *maxCycles < 1 {
		return fmt.Errorf("--max must be at least 1")
	}

	ledger := repairLedger{ChangeID: *change, MaxCycles: *maxCycles}
	if raw, err := os.ReadFile(*ledgerPath); err == nil {
		if err := json.Unmarshal(raw, &ledger); err != nil {
			return fmt.Errorf("ledger is corrupt — refusing to guess repair count (fail closed): %w", err)
		}
		if ledger.ChangeID != *change {
			return fmt.Errorf("ledger belongs to change %q, not %q — one ledger per change", ledger.ChangeID, *change)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read ledger: %w", err)
	}

	// A repair against the SAME failed head twice is a no-progress loop
	// regardless of the count (doc 08 §13: patch hash unchanged).
	for _, c := range ledger.Cycles {
		if c.FromHead == *fromHead {
			return fmt.Errorf("a repair cycle for head %s was already authorized (cycle %d) — no-progress loop, human adjudication required", *fromHead, c.Number)
		}
	}

	if len(ledger.Cycles) >= ledger.MaxCycles {
		heads := make([]string, 0, len(ledger.Cycles))
		for _, c := range ledger.Cycles {
			heads = append(heads, c.FromHead)
		}
		sort.Strings(heads)
		return fmt.Errorf("repair-cycle limit reached (%d/%d, repaired heads: %v) — needs_human (PAC-007)", len(ledger.Cycles), ledger.MaxCycles, heads)
	}

	ledger.Cycles = append(ledger.Cycles, repairCycle{
		Number:    len(ledger.Cycles) + 1,
		FromHead:  *fromHead,
		Reason:    *reason,
		StartedAt: *startedAt,
	})
	if err := os.MkdirAll(filepath.Dir(*ledgerPath), 0o755); err != nil {
		return err
	}
	out, _ := json.MarshalIndent(ledger, "", "  ")
	if err := os.WriteFile(*ledgerPath, out, 0o644); err != nil {
		return err
	}
	fmt.Printf("repair cycle %d/%d authorized for %s (from head %s)\n",
		len(ledger.Cycles), ledger.MaxCycles, *change, *fromHead)
	return nil
}
