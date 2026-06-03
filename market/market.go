// Package market holds the core market-data types and CSV loaders.
//
// You normally won't need to change anything in this package — it's provided so
// you can see exactly what data flows into your indicators. Read it before you
// start so you understand the shape of a Snapshot.
package market

import "time"

// OptionType distinguishes calls from puts in the option chain.
type OptionType string

const (
	Call OptionType = "CALL"
	Put  OptionType = "PUT"
)

// Bar is one OHLCV bar of the underlying instrument.
type Bar struct {
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

// ChainEntry is one strike/right in the option-chain snapshot for a bar.
// MarkIV is the per-strike implied volatility (already supplied by the feed —
// you do not compute it). OI is open interest, used only by the PCR stretch goal.
type ChainEntry struct {
	Strike float64
	Type   OptionType
	MarkIV float64
	OI     float64
}

// Snapshot is everything known at one bar's close: the underlying bar, the ATM
// implied volatility for that bar, and the full option chain. Indicators receive
// one Snapshot per bar via Update.
type Snapshot struct {
	Time  time.Time
	Bar   Bar
	ATMIV float64
	Chain []ChainEntry
}
