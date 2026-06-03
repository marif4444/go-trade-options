// Package engine runs a strategy spec over a series of snapshots: it evaluates
// predicates per bar and manages position state. Provided — read it to learn the
// contracts your indicators plug into.
package engine

import "time"

// Side is the direction of a position.
type Side string

const (
	Long  Side = "LONG"
	Short Side = "SHORT"
)

// Position is the currently-open position (if any).
type Position struct {
	Open       bool
	Side       Side
	EntryTime  time.Time
	EntryPrice float64
	Qty        float64
}

// PnL returns the mark-to-market profit of the open position at the given price.
func (p Position) PnL(price float64) float64 {
	if !p.Open {
		return 0
	}
	if p.Side == Short {
		return (p.EntryPrice - price) * p.Qty
	}
	return (price - p.EntryPrice) * p.Qty
}

// Trade is a completed round-trip.
type Trade struct {
	EntryTime  time.Time
	ExitTime   time.Time
	Side       Side
	EntryPrice float64
	ExitPrice  float64
	Qty        float64
	PnL        float64
}
