// Command gendata writes the synthetic dataset used by the take-home:
// data/ohlc.csv (underlying bars) and data/chain.csv (per-bar option chain).
//
// It is deterministic (fixed seed) so the dataset is reproducible. You do not
// need to run it — the CSVs are already committed — but it's here so you can see
// exactly how the data was constructed.
//
//	go run ./tools/gendata
package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"
)

const (
	bars     = 160
	startUSD = 60000.0
)

func main() {
	rng := rand.New(rand.NewSource(42))
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	ohlc := [][]string{{"time", "open", "high", "low", "close", "volume"}}
	chain := [][]string{{"time", "strike", "type", "mark_iv", "oi"}}

	spot := startUSD
	for i := 0; i < bars; i++ {
		t := start.Add(time.Duration(i) * time.Hour).Format(time.RFC3339)

		// Underlying: gentle random walk.
		open := spot
		drift := rng.NormFloat64() * 120
		closePx := open + drift
		high := math.Max(open, closePx) + math.Abs(rng.NormFloat64()*60)
		low := math.Min(open, closePx) - math.Abs(rng.NormFloat64()*60)
		vol := 100 + rng.Float64()*50
		spot = closePx

		ohlc = append(ohlc, []string{
			t, f(open), f(high), f(low), f(closePx), f(vol),
		})

		// ATM IV: two volatility regimes (a calm stretch and a stressed one)
		// via a sine wave, so IV Rank crosses the 50/30 thresholds repeatedly.
		atmIV := 0.60 + 0.28*math.Sin(2*math.Pi*float64(i)/45.0) + rng.NormFloat64()*0.01

		// Option chain: 5 strikes around spot rounded to the nearest 1000.
		atmStrike := math.Round(closePx/1000) * 1000
		for k := -2; k <= 2; k++ {
			strike := atmStrike + float64(k)*1000
			// Smile: IV rises with distance from spot.
			skew := 0.02 * math.Abs(strike-closePx) / 1000.0
			markIV := atmIV + skew
			for _, typ := range []string{"CALL", "PUT"} {
				// Puts carry more OI when vol is high (a crude fear proxy) — used
				// only by the PCR stretch goal.
				base := 200 + rng.Float64()*100
				oi := base
				if typ == "PUT" {
					oi = base * (1 + atmIV)
				}
				chain = append(chain, []string{
					t, f(strike), typ, f(markIV), f(oi),
				})
			}
		}
	}

	write("data/ohlc.csv", ohlc)
	write("data/chain.csv", chain)
	fmt.Printf("wrote data/ohlc.csv (%d bars) and data/chain.csv\n", bars)
}

func f(v float64) string { return strconv.FormatFloat(v, 'f', 4, 64) }

func write(path string, rows [][]string) {
	out, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer out.Close()
	w := csv.NewWriter(out)
	if err := w.WriteAll(rows); err != nil {
		panic(err)
	}
	w.Flush()
}
