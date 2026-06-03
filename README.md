# Vol-Aware Indicators
## Requirements

- Go 1.22+ (standard library only)

## Run

```bash
go test ./...          # unit + golden + e2e tests
go test -race ./...    # must be clean if you add concurrency
go run ./cmd/backtest  # run the example strategy over the provided data
```

A clean checkout builds and passes the harness timing test, but the IV Rank /
IVP golden tests and the end-to-end trade test fail until you implement the
stubs — that failing state is your starting line.

## Layout

```
data/                  underlying OHLC bars + per-bar option-chain snapshots (CSV)
spec/                  the strategy DSL: StrategySpec, Rule, Predicate          [PROVIDED]
engine/
  harness.go           backtest loop — signal-at-close → fill-at-next-open      [PROVIDED]
  position.go          position state machine, P&L                             [PROVIDED]
  evaluator.go         predicate evaluation; one TODO for you                  [PARTIAL]
indicators/
  indicator.go         the Indicator interface + registry                      [PROVIDED]
  ivrank.go            IV_RANK — you implement                                 [STUB]
  ivp.go               IV_PERCENTILE — you implement                           [STUB]
strategies/
  short_vol.json       example strategy referencing IV_RANK                    [PROVIDED]
cmd/backtest/          CLI that runs a strategy and prints trades + P&L         [PROVIDED]
tools/gendata/         deterministic generator for the data/ CSVs              [PROVIDED]
```

Start by reading `engine/harness.go`, `engine/position.go`, and `spec/` to
understand the contracts, then fill the three `TODO(candidate)` spots:
`indicators/ivrank.go`, `indicators/ivp.go`, and `resolveOperand` in
`engine/evaluator.go`.
