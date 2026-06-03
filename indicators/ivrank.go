package indicators

import "takehome-vol-indicators/market"

// IVRank is the IV Rank indicator: where the current ATM IV sits within its
// trailing high/low range over the last `lookback` bars, scaled to 0..100:
//
//	IV Rank = 100 * (iv_now - min) / (max - min)   over the trailing window
//
// TODO(candidate): implement this.
//
// Requirements:
//   - STREAMING: Update must do bounded work per bar (maintain a rolling window),
//     not rescan all history ever seen.
//   - WARM-UP: not ready until the window holds `lookback` bars; Value's bool is
//     false until then.
//   - FLAT WINDOW: decide what to return when max == min (no range) and document
//     your choice in the README. Don't return NaN/Inf or panic.
type IVRank struct {
	lookback int
	// TODO(candidate): add whatever state you need (e.g. a rolling window).
}

func NewIVRank(lookback int) *IVRank {
	return &IVRank{lookback: lookback}
}

func (r *IVRank) Name() string { return "IV_RANK" }

func (r *IVRank) Update(s market.Snapshot) {
	// TODO(candidate): record this bar's ATM IV (s.ATMIV) into your window.
}

func (r *IVRank) Value() (float64, bool) {
	// TODO(candidate): return (ivRank, true) once warm; (0, false) until then.
	return 0, false
}
