package engine

import (
	"testing"

	"takehome-vol-indicators/market"
	"takehome-vol-indicators/spec"
)

// End-to-end: the example strategy, run over the provided dataset, should
// produce trades once the indicators and operand resolution are implemented —
// and do so deterministically.
func TestExampleStrategy_ProducesDeterministicTrades(t *testing.T) {
	snaps, err := market.BuildSnapshots("../data/ohlc.csv", "../data/chain.csv")
	if err != nil {
		t.Fatalf("load data: %v", err)
	}
	sp, err := spec.Load("../strategies/short_vol.json")
	if err != nil {
		t.Fatalf("load strategy: %v", err)
	}

	run := func() Result {
		ev, err := NewEvaluator(sp)
		if err != nil {
			t.Fatalf("build evaluator: %v", err)
		}
		return Run(snaps, sp, ev)
	}

	res := run()
	if len(res.Trades) == 0 {
		t.Fatalf("example strategy produced 0 trades — implement IV_RANK and resolveOperand")
	}

	// Reproducibility: same inputs => same trades.
	res2 := run()
	if len(res2.Trades) != len(res.Trades) {
		t.Fatalf("non-deterministic: %d vs %d trades", len(res.Trades), len(res2.Trades))
	}
	for i := range res.Trades {
		if res.Trades[i] != res2.Trades[i] {
			t.Fatalf("trade %d differs between runs", i)
		}
	}
}
