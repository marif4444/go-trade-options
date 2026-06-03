package indicators

import (
	"testing"
	"time"

	"takehome-vol-indicators/market"
)

// snapsFromIV builds bar snapshots carrying only an ATM IV value — enough to
// drive the vol indicators.
func snapsFromIV(ivs ...float64) []market.Snapshot {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	out := make([]market.Snapshot, len(ivs))
	for i, iv := range ivs {
		out[i] = market.Snapshot{Time: base.Add(time.Duration(i) * time.Hour), ATMIV: iv}
	}
	return out
}

// feed runs an indicator over a series and returns its (value, ready) at each bar.
func feed(ind Indicator, snaps []market.Snapshot) []struct {
	v     float64
	ready bool
} {
	out := make([]struct {
		v     float64
		ready bool
	}, len(snaps))
	for i, s := range snaps {
		ind.Update(s)
		out[i].v, out[i].ready = ind.Value()
	}
	return out
}

func TestIVRank_WarmUp(t *testing.T) {
	r := NewIVRank(4)
	got := feed(r, snapsFromIV(10, 20, 30))
	for i, g := range got {
		if g.ready {
			t.Fatalf("bar %d: indicator should not be ready before the window is full", i)
		}
	}
}

func TestIVRank_GoldenValues(t *testing.T) {
	// lookback 4. Windows (most recent 4 incl. current):
	//   i=3: [10,20,30,40] -> (40-10)/(40-10) = 100
	//   i=4: [20,30,40,25] -> (25-20)/(40-20) = 25
	//   i=5: [30,40,25, 5] -> ( 5- 5)/(40- 5) = 0
	r := NewIVRank(4)
	got := feed(r, snapsFromIV(10, 20, 30, 40, 25, 5))

	want := map[int]float64{3: 100, 4: 25, 5: 0}
	for i := 0; i < 3; i++ {
		if got[i].ready {
			t.Fatalf("bar %d: should not be ready yet", i)
		}
	}
	for i, w := range want {
		if !got[i].ready {
			t.Fatalf("bar %d: expected ready", i)
		}
		if !approx(got[i].v, w) {
			t.Errorf("bar %d: IV Rank = %v, want %v", i, got[i].v, w)
		}
	}
}

// A flat window has no range. We don't assert a specific value here (it's a
// convention you choose and document) — only that it doesn't panic or NaN.
func TestIVRank_FlatWindowDoesNotExplode(t *testing.T) {
	r := NewIVRank(3)
	got := feed(r, snapsFromIV(20, 20, 20))
	last := got[len(got)-1]
	if last.ready && isNaNOrInf(last.v) {
		t.Fatalf("flat window produced a non-finite value: %v", last.v)
	}
}

func approx(a, b float64) bool {
	const eps = 1e-9
	d := a - b
	return d < eps && d > -eps
}

func isNaNOrInf(f float64) bool {
	return f != f || f > 1e308 || f < -1e308
}
