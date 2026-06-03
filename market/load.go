package market

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"
)

// LoadOHLC reads underlying bars from a CSV with header:
//
//	time,open,high,low,close,volume
//
// time is RFC3339 (e.g. 2025-01-01T00:00:00Z).
func LoadOHLC(path string) ([]Bar, error) {
	rows, err := readCSV(path)
	if err != nil {
		return nil, err
	}
	bars := make([]Bar, 0, len(rows))
	for i, r := range rows {
		if len(r) < 6 {
			return nil, fmt.Errorf("ohlc row %d: want 6 columns, got %d", i+2, len(r))
		}
		t, err := time.Parse(time.RFC3339, r[0])
		if err != nil {
			return nil, fmt.Errorf("ohlc row %d: bad time %q: %w", i+2, r[0], err)
		}
		bars = append(bars, Bar{
			Time:   t,
			Open:   mustFloat(r[1]),
			High:   mustFloat(r[2]),
			Low:    mustFloat(r[3]),
			Close:  mustFloat(r[4]),
			Volume: mustFloat(r[5]),
		})
	}
	return bars, nil
}

// LoadChain reads option-chain snapshots from a CSV with header:
//
//	time,strike,type,mark_iv,oi
//
// and groups entries by bar time.
func LoadChain(path string) (map[time.Time][]ChainEntry, error) {
	rows, err := readCSV(path)
	if err != nil {
		return nil, err
	}
	out := map[time.Time][]ChainEntry{}
	for i, r := range rows {
		if len(r) < 5 {
			return nil, fmt.Errorf("chain row %d: want 5 columns, got %d", i+2, len(r))
		}
		t, err := time.Parse(time.RFC3339, r[0])
		if err != nil {
			return nil, fmt.Errorf("chain row %d: bad time %q: %w", i+2, r[0], err)
		}
		out[t] = append(out[t], ChainEntry{
			Strike: mustFloat(r[1]),
			Type:   OptionType(r[2]),
			MarkIV: mustFloat(r[3]),
			OI:     mustFloat(r[4]),
		})
	}
	return out, nil
}

// BuildSnapshots joins the OHLC series with the option chain by bar time and
// derives ATM IV for each bar (the implied vol of the strike closest to spot).
func BuildSnapshots(ohlcPath, chainPath string) ([]Snapshot, error) {
	bars, err := LoadOHLC(ohlcPath)
	if err != nil {
		return nil, err
	}
	chain, err := LoadChain(chainPath)
	if err != nil {
		return nil, err
	}
	snaps := make([]Snapshot, 0, len(bars))
	for _, b := range bars {
		entries := chain[b.Time]
		snaps = append(snaps, Snapshot{
			Time:  b.Time,
			Bar:   b,
			ATMIV: atmIVFromChain(entries, b.Close),
			Chain: entries,
		})
	}
	return snaps, nil
}

// atmIVFromChain returns the average mark IV of the strike closest to spot.
func atmIVFromChain(entries []ChainEntry, spot float64) float64 {
	bestStrike, found := 0.0, false
	for _, e := range entries {
		if !found || math.Abs(e.Strike-spot) < math.Abs(bestStrike-spot) {
			bestStrike, found = e.Strike, true
		}
	}
	if !found {
		return 0
	}
	var sum float64
	var n int
	for _, e := range entries {
		if e.Strike == bestStrike {
			sum += e.MarkIV
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}

func readCSV(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) <= 1 {
		return nil, fmt.Errorf("%s: no data rows", path)
	}
	return rows[1:], nil // drop header
}

func mustFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
