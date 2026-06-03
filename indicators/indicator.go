// Package indicators contains the streaming indicators the strategy DSL can
// reference. You implement IV_RANK and IV_PERCENTILE (and, optionally, PCR).
package indicators

import (
	"fmt"

	"takehome-vol-indicators/market"
)

// Indicator is a streaming (incremental) indicator. Update is called once per
// bar, in order, with that bar's Snapshot. Value returns the latest value and
// whether the indicator is "ready" (has seen enough history to be meaningful).
//
// Contract:
//   - Update does work proportional to the update, not a full rescan of all
//     history ever seen. A bounded rolling window is expected.
//   - Until ready, Value's bool is false and the float should be ignored.
type Indicator interface {
	Update(s market.Snapshot)
	Value() (value float64, ready bool)
	Name() string
}

// New constructs an indicator by its DSL name. The engine calls this when it
// sees an indicator operand in a strategy spec, passing the spec's lookback.
//
// (STRETCH) Add a "PCR" case here once you implement indicators/pcr.go.
func New(name string, lookback int) (Indicator, error) {
	switch name {
	case "IV_RANK":
		return NewIVRank(lookback), nil
	case "IV_PERCENTILE":
		return NewIVPercentile(lookback), nil
	default:
		return nil, fmt.Errorf("unknown indicator %q", name)
	}
}
