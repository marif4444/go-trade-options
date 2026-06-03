package indicators

import "takehome-vol-indicators/market"

// IVPercentile is the IV Percentile indicator: the fraction of bars in the
// trailing window whose ATM IV was below the current bar's, scaled to 0..100.
//
//	IVP = 100 * count(iv_i < iv_now) / (bars in window)
//
// TODO(candidate): implement this.
//
// IVP is NOT the same as IV Rank — make sure your two implementations actually
// differ. The definition has two genuine ambiguities; decide and document both
// in the README:
//   - ties: is the comparison strict (`<`) or inclusive (`<=`)?
//   - denominator: do you count the current bar in it, or only prior bars?
//
// Same streaming + warm-up requirements as IV Rank apply.
type IVPercentile struct {
	lookback int
	// TODO(candidate): add whatever state you need.
}

func NewIVPercentile(lookback int) *IVPercentile {
	return &IVPercentile{lookback: lookback}
}

func (p *IVPercentile) Name() string { return "IV_PERCENTILE" }

func (p *IVPercentile) Update(s market.Snapshot) {
	// TODO(candidate): record this bar's ATM IV (s.ATMIV) into your window.
}

func (p *IVPercentile) Value() (float64, bool) {
	// TODO(candidate): return (ivp, true) once warm; (0, false) until then.
	return 0, false
}
