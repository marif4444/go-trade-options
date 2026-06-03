package engine

import (
	"testing"
	"time"

	"takehome-vol-indicators/market"
	"takehome-vol-indicators/spec"
)

// scriptedDecider fires Enter/Exit on specific (0-based) bar indices, so we can
// assert exactly when fills land — independent of any real indicator.
type scriptedDecider struct {
	cur     int
	enterAt int
	exitAt  int
}

func newScripted(enterAt, exitAt int) *scriptedDecider {
	return &scriptedDecider{cur: -1, enterAt: enterAt, exitAt: exitAt}
}

func (d *scriptedDecider) Update(market.Snapshot) { d.cur++ }
func (d *scriptedDecider) Enter() bool            { return d.cur == d.enterAt }
func (d *scriptedDecider) Exit() bool             { return d.cur == d.exitAt }

func barsWithOpens(opens ...float64) []market.Snapshot {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	out := make([]market.Snapshot, len(opens))
	for i, o := range opens {
		out[i] = market.Snapshot{
			Time: base.Add(time.Duration(i) * time.Hour),
			Bar:  market.Bar{Open: o, Close: o + 1},
		}
	}
	return out
}

// The signal-at-close / fill-at-next-open contract: a signal on bar N must fill
// at the OPEN of bar N+1, never on bar N itself.
func TestHarness_FillsAtNextBarOpen(t *testing.T) {
	snaps := barsWithOpens(100, 101, 102, 103, 104, 105, 106, 107)
	sp := spec.StrategySpec{Side: "LONG", Qty: 1}

	res := Run(snaps, sp, newScripted(2, 5)) // enter signal on bar 2, exit on bar 5

	if len(res.Trades) != 1 {
		t.Fatalf("want 1 trade, got %d", len(res.Trades))
	}
	tr := res.Trades[0]
	if tr.EntryTime != snaps[3].Time {
		t.Errorf("entry time = %s, want bar 3 (%s)", tr.EntryTime, snaps[3].Time)
	}
	if tr.EntryPrice != snaps[3].Bar.Open {
		t.Errorf("entry price = %v, want bar-3 open %v (signal was on bar 2)", tr.EntryPrice, snaps[3].Bar.Open)
	}
	if tr.ExitPrice != snaps[6].Bar.Open {
		t.Errorf("exit price = %v, want bar-6 open %v (signal was on bar 5)", tr.ExitPrice, snaps[6].Bar.Open)
	}
}
