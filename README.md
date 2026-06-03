# Take-Home: Vol-Aware Indicators for the Strategy Engine

**Role:** Backend / Quant Engineer — Algorithmic Trading Platform
**Time budget:** 1 day of focused effort. We mean it — scope is set so a strong submission fits comfortably. Don't gold-plate; depth and correctness beat breadth.

---

## Context

We build a high-performance algorithmic trading engine — backtesting, paper, and live execution across crypto futures and options on Delta Exchange. Strategies are expressed in a small DSL (the *strategy spec*): **legs** (instruments), **rules** (entry/exit conditions), and **predicates** that compare indicator values to thresholds or to each other.

Options traders care a lot about **implied volatility regime** — not the absolute IV number, but *where today's IV sits relative to its recent history*. Two indicators capture this:

- **IV Rank** — where current IV falls within its trailing high/low range.
- **IV Percentile (IVP)** — what fraction of the trailing window had IV below today's.

This exercise asks you to build those two indicators **incrementally** (streaming, one update per bar — not a full recompute every bar), wire them into the strategy DSL, and make a provided example strategy trade on them.

You do **not** need prior options knowledge. Everything you need is defined below. We're evaluating engineering judgment on money-moving software — correctness, testing, and clear reasoning — not whether you already know derivatives.

---

## What we give you

A slimmed, self-contained version of the engine (no DB, no actor framework, no exchange connection). Layout:

```
.
├── data/
│   ├── ohlc.csv          # underlying OHLC bars (timestamp, o,h,l,c,v)
│   └── chain.csv         # per-bar option-chain snapshot (ts, strike, type, mark_iv, oi)
├── spec/
│   └── strategy_spec.go  # the DSL: StrategySpec, Rule, Predicate, RELATION/CROSS_OVER  [PROVIDED]
├── engine/
│   ├── harness.go        # backtest loop — signal-at-close → fill-at-next-open   [PROVIDED]
│   ├── position.go       # position state machine (entry/exit, P&L)              [PROVIDED]
│   └── evaluator.go      # predicate evaluation; one TODO for you to fill        [PARTIAL]
├── indicators/
│   ├── indicator.go      # the Indicator interface + ATM-IV series feed          [PROVIDED]
│   ├── ivrank.go         # STUB — you implement
│   └── ivp.go            # STUB — you implement
├── strategies/
│   └── short_vol.json    # example strategy that references IV_RANK              [PROVIDED]
└── *_test.go             # golden tests — some are failing until you implement   [PROVIDED]
```

You should be able to read `harness.go`, `position.go`, and `spec/` to understand the existing contracts before writing anything. **Reading the provided code and matching its style is part of the exercise.**

---

## Definitions (read carefully — the subtleties matter)

Let the lookback window be `N` bars and `iv_t` the ATM implied volatility at the current bar `t`. The window is the most recent `N` bars *up to and including* `t`.

**IV Rank** — position within the trailing range, scaled 0–100:
```
ivrank_t = 100 * (iv_t - min(iv over window)) / (max(iv over window) - min(iv over window))
```
- If `max == min` (flat window), the result is undefined — decide and document your convention.

**IV Percentile (IVP)** — fraction of the window strictly below the current value, scaled 0–100:
```
ivp_t = 100 * count(iv_i < iv_t for i in window) / (count of bars in window)
```
- Decide and document: is the current bar included in the denominator? How do you treat ties (`iv_i == iv_t`)?

> IV Rank and IVP are **not** the same. A value can be high-rank but mid-percentile and vice-versa. Getting both definitions right — and the edge cases — is the core of the task.

**Warm-up:** treat the indicator as *ready* once it has observed `N` bars — a full window. Until then the `Indicator` interface's `ready` bool is `false` and the value must not produce a tradable signal. (The two genuine ambiguities we want you to decide and document are the *flat-window* case in IV Rank and the *tie/denominator* choice in IVP — not the warm-up boundary.)

---

## Bar-timing contract (this is where correctness lives)

The provided harness already enforces this — your job is to not break it:

- An indicator updates on each bar's **close**.
- An entry/exit **signal** is evaluated on the **close** of bar `N`.
- The resulting fill happens at the **open of bar `N+1`**.

A signal that peeks at the same bar's open it fills on, or that fills on the close it was computed from, is a lookahead bug. We will test for this.

---

## Your task (required)

1. **Implement `IV_RANK` and `IV_PERCENTILE`** in `indicators/ivrank.go` and `indicators/ivp.go`, satisfying the `Indicator` interface.
   - **Incremental:** each `Update(bar)` does work proportional to the update, not a full O(N) rescan of the whole history every bar. (A bounded rolling window is fine; rescanning all of history is not.)
   - Respect the warm-up / `ready` semantics.
   - Make the lookback `N` configurable.
2. **Expose them as DSL operands** so a `RELATION` predicate can reference `IV_RANK` / `IV_PERCENTILE` against a constant (e.g. `IV_RANK > 50`). Follow the existing operand-resolution pattern in `evaluator.go` — there's a marked `TODO`.
3. **Make `strategies/short_vol.json` run** through the harness against the provided data and emit the resulting trades + summary P&L. (e.g. *enter short when `IV_RANK > 50`, exit when `IV_RANK < 30`.*)
4. **Tests.** Unit-test the indicators (the property *incremental result == batch recompute over the same window* is a strong one), cover the edge cases you identified, and add at least one end-to-end test that the example strategy produces the expected trades. Make the provided golden tests pass.
5. **README.** A short write-up (see below).

## Stretch goals (optional — for candidates with time to spare)

Pick what interests you; not required for a strong score.

- **PCR (Put-Call Ratio):** `chain.csv` includes an `oi` (open-interest) column. Add a `PCR` indicator = total put OI / total call OI per bar, expose it in the DSL, and write a strategy that combines it with `IV_RANK`. Note any market-structure assumptions (zero-OI strikes, which strikes to include).
- **O(1) rolling extremes:** IV Rank's min/max over a sliding window can be maintained in amortized O(1) with a monotonic deque instead of scanning the window each bar. Implement it and show the win with a benchmark.

---

## Constraints & ground rules

- **Go.** Use the standard library; add a dependency only if you can justify it in the README.
- **Race-free.** If you introduce concurrency, `go test -race ./...` must be clean. (Concurrency is *not* required — don't add it for show.)
- **No lookahead.** Respect the bar-timing contract.
- **Determinism.** Same input → same output.
- **AI tools:** use whatever you'd use day-to-day. Be ready to explain and defend every line in a follow-up conversation — we'll go deep on your reasoning, so don't ship code you can't account for.

---

## What we evaluate

| Area | What we look for |
|------|------------------|
| **Correctness** | Right definitions, edge cases handled, no lookahead, warm-up respected |
| **Streaming design** | Genuinely incremental; bounded memory; no full-history rescans |
| **Testing** | Meaningful unit + e2e tests; property/edge-case coverage; reference validation mindset |
| **Code-reading** | Fits the existing contracts and style; minimal, surgical changes |
| **Communication** | README clearly states assumptions, trade-offs, and what you'd do with more time |

## README — please cover

- Your conventions for the ambiguous cases (flat window in IV Rank; tie handling and denominator in IVP).
- How you kept the indicators incremental, and the time/space complexity per `Update`.
- How you convinced yourself there's no lookahead.
- What you tested and why; what you'd test next.
- Trade-offs you made for the time budget, and what you'd do differently in production.

## Submission

- A git repo (or PR against the provided one) with clear, reviewable commits.
- `go test ./...` and `go test -race ./...` should pass from a clean checkout.
- We'll schedule a short call to walk through your design and reasoning.

Good luck — and tell us where you spent your time. We'd rather see two indicators done carefully than four done hastily.
