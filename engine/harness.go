package engine

import (
	"time"

	"takehome-vol-indicators/market"
	"takehome-vol-indicators/spec"
)

// Decider is the per-bar decision interface the backtest loop depends on.
// *Evaluator implements it; tests may supply their own.
type Decider interface {
	// Update advances any internal state with this bar's snapshot (called at
	// the bar's close, before Enter/Exit are consulted).
	Update(s market.Snapshot)
	// Enter reports whether the entry rule fires on the current bar's close.
	Enter() bool
	// Exit reports whether the exit rule fires on the current bar's close.
	Exit() bool
}

// EquityPoint is the mark-to-market equity at one bar's close.
type EquityPoint struct {
	Time   time.Time
	Equity float64
}

// Result is the outcome of a backtest run.
type Result struct {
	Trades []Trade
	Equity []EquityPoint
	NetPnL float64
}

// Run executes a strategy over snapshots in order.
//
// Bar-timing contract (do not weaken this — it is what keeps backtests honest):
//
//	A signal is evaluated on bar N's CLOSE; the resulting fill happens at the
//	OPEN of bar N+1. Concretely, each iteration first fills any order scheduled
//	on the previous bar, then updates indicators on this bar's close, then
//	evaluates the rules to (maybe) schedule an order for the next bar.
func Run(snaps []market.Snapshot, sp spec.StrategySpec, d Decider) Result {
	side := Side(sp.Side)
	if side != Long && side != Short {
		side = Long
	}

	var pos Position
	var res Result
	pending := none
	realized := 0.0

	for _, s := range snaps {
		// 1) Fill the order scheduled on the previous bar, at THIS bar's open.
		switch pending {
		case enterPending:
			pos = Position{Open: true, Side: side, EntryTime: s.Time, EntryPrice: s.Bar.Open, Qty: sp.Qty}
		case exitPending:
			if pos.Open {
				res.Trades = append(res.Trades, closeAt(pos, s.Time, s.Bar.Open))
				realized += pos.PnL(s.Bar.Open)
				pos = Position{}
			}
		}
		pending = none

		// 2) Update indicators on this bar's close.
		d.Update(s)

		// 3) Evaluate rules on the close; schedule a fill for the NEXT bar's open.
		if !pos.Open {
			if d.Enter() {
				pending = enterPending
			}
		} else {
			if d.Exit() {
				pending = exitPending
			}
		}

		res.Equity = append(res.Equity, EquityPoint{Time: s.Time, Equity: realized + pos.PnL(s.Bar.Close)})
	}

	// Force-close any position still open at the final bar's close.
	if pos.Open && len(snaps) > 0 {
		last := snaps[len(snaps)-1]
		res.Trades = append(res.Trades, closeAt(pos, last.Time, last.Bar.Close))
		realized += pos.PnL(last.Bar.Close)
	}

	res.NetPnL = realized
	return res
}

type pendingAction int

const (
	none pendingAction = iota
	enterPending
	exitPending
)

func closeAt(pos Position, t time.Time, price float64) Trade {
	return Trade{
		EntryTime:  pos.EntryTime,
		ExitTime:   t,
		Side:       pos.Side,
		EntryPrice: pos.EntryPrice,
		ExitPrice:  price,
		Qty:        pos.Qty,
		PnL:        pos.PnL(price),
	}
}
