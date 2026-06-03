package indicators

import "testing"

// NOTE ON CONVENTION
//
// The expected values below assume IV Percentile is computed as:
//
//	100 * count(iv_i strictly < iv_now) / (number of bars in the window)
//
// i.e. strict less-than, with the current bar counted in the denominator. That
// is one reasonable convention. If your README justifies a different one
// (e.g. <=, or excluding the current bar from the denominator), update the
// expectations here to match and note it in your write-up — we read both the
// code and the reasoning.
func TestIVP_GoldenValues(t *testing.T) {
	// lookback 4. Windows (most recent 4 incl. current):
	//   i=3: [10,20,30,40] current 40 -> 3 below -> 75
	//   i=4: [20,30,40,25] current 25 -> 1 below -> 25
	//   i=5: [30,40,25, 5] current  5 -> 0 below -> 0
	p := NewIVPercentile(4)
	got := feed(p, snapsFromIV(10, 20, 30, 40, 25, 5))

	for i := 0; i < 3; i++ {
		if got[i].ready {
			t.Fatalf("bar %d: should not be ready before the window is full", i)
		}
	}
	want := map[int]float64{3: 75, 4: 25, 5: 0}
	for i, w := range want {
		if !got[i].ready {
			t.Fatalf("bar %d: expected ready", i)
		}
		if !approx(got[i].v, w) {
			t.Errorf("bar %d: IVP = %v, want %v", i, got[i].v, w)
		}
	}
}

// Convention-independent sanity: IVP is always within [0,100], and the window
// minimum has nothing strictly below it (0% regardless of tie/denominator rule).
func TestIVP_Bounds(t *testing.T) {
	p := NewIVPercentile(4)
	got := feed(p, snapsFromIV(40, 30, 20, 10)) // current (10) is the window min
	last := got[len(got)-1]
	if !last.ready {
		t.Fatal("expected ready")
	}
	if last.v < 0 || last.v > 100 {
		t.Fatalf("IVP out of range: %v", last.v)
	}
	if !approx(last.v, 0) {
		t.Errorf("IVP of the window minimum = %v, want 0", last.v)
	}
}
