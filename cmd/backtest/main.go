// Command backtest runs a strategy spec over the provided dataset and prints the
// resulting trades and net P&L. Once you've implemented the indicators and the
// operand-resolution hook, this should report a handful of trades.
//
//	go run ./cmd/backtest
//	go run ./cmd/backtest -strategy strategies/short_vol.json
package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"takehome-vol-indicators/engine"
	"takehome-vol-indicators/market"
	"takehome-vol-indicators/spec"
)

func main() {
	ohlc := flag.String("ohlc", "data/ohlc.csv", "path to OHLC CSV")
	chain := flag.String("chain", "data/chain.csv", "path to option-chain CSV")
	strat := flag.String("strategy", "strategies/short_vol.json", "path to strategy JSON")
	flag.Parse()

	snaps, err := market.BuildSnapshots(*ohlc, *chain)
	if err != nil {
		log.Fatalf("load data: %v", err)
	}
	sp, err := spec.Load(*strat)
	if err != nil {
		log.Fatalf("load strategy: %v", err)
	}
	ev, err := engine.NewEvaluator(sp)
	if err != nil {
		log.Fatalf("build evaluator: %v", err)
	}

	res := engine.Run(snaps, sp, ev)

	fmt.Printf("strategy=%s  bars=%d  trades=%d  net_pnl=%.2f\n",
		sp.Name, len(snaps), len(res.Trades), res.NetPnL)
	for i, t := range res.Trades {
		fmt.Printf("  #%d %-5s entry %s @ %9.2f  ->  exit %s @ %9.2f   pnl %+9.2f\n",
			i+1, t.Side,
			t.EntryTime.Format(time.RFC3339), t.EntryPrice,
			t.ExitTime.Format(time.RFC3339), t.ExitPrice, t.PnL)
	}
	if len(res.Trades) == 0 {
		fmt.Println("  (no trades — have you implemented the indicators and resolveOperand?)")
	}
}
