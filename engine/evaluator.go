package engine

import (
	"takehome-vol-indicators/indicators"
	"takehome-vol-indicators/market"
	"takehome-vol-indicators/spec"
)

// Evaluator evaluates a strategy spec's rules against streaming indicator
// values. It owns one indicator instance per distinct indicator referenced by
// the spec, advances them each bar, and answers Enter()/Exit().
type Evaluator struct {
	spec spec.StrategySpec
	inds map[string]indicators.Indicator

	// previous-bar values, kept so CROSS_OVER can detect a crossing.
	prev     map[string]float64
	havePrev map[string]bool
}

// NewEvaluator builds the indicator set the spec references and returns a
// ready-to-run evaluator. Indicators are constructed via indicators.New, so
// adding a new indicator name there is all it takes to make it usable in the DSL.
func NewEvaluator(sp spec.StrategySpec) (*Evaluator, error) {
	e := &Evaluator{
		spec:     sp,
		inds:     map[string]indicators.Indicator{},
		prev:     map[string]float64{},
		havePrev: map[string]bool{},
	}
	preds := append(append([]spec.Predicate{}, sp.Entry.Predicates...), sp.Exit.Predicates...)
	for _, p := range preds {
		for _, op := range []spec.Operand{p.Left, p.Right} {
			if op.Indicator == "" {
				continue
			}
			if _, ok := e.inds[op.Indicator]; ok {
				continue
			}
			ind, err := indicators.New(op.Indicator, sp.Lookback)
			if err != nil {
				return nil, err
			}
			e.inds[op.Indicator] = ind
		}
	}
	return e, nil
}

// Update snapshots the current indicator values as "previous" (for CROSS_OVER)
// and then advances every indicator with this bar.
func (e *Evaluator) Update(s market.Snapshot) {
	for name, ind := range e.inds {
		if v, ok := ind.Value(); ok {
			e.prev[name] = v
			e.havePrev[name] = true
		}
	}
	for _, ind := range e.inds {
		ind.Update(s)
	}
}

func (e *Evaluator) Enter() bool { return e.evalRule(e.spec.Entry) }
func (e *Evaluator) Exit() bool  { return e.evalRule(e.spec.Exit) }

func (e *Evaluator) evalRule(r spec.Rule) bool {
	if len(r.Predicates) == 0 {
		return false
	}
	if r.Logic == "ANY" {
		for _, p := range r.Predicates {
			if e.evalPredicate(p) {
				return true
			}
		}
		return false
	}
	// ALL (default)
	for _, p := range r.Predicates {
		if !e.evalPredicate(p) {
			return false
		}
	}
	return true
}

func (e *Evaluator) evalPredicate(p spec.Predicate) bool {
	switch p.Type {
	case "RELATION":
		l, lok := e.resolveOperand(p.Left)
		r, rok := e.resolveOperand(p.Right)
		if !lok || !rok {
			return false
		}
		return compare(l, p.Op, r)
	case "CROSS_OVER":
		lNow, lok := e.resolveOperand(p.Left)
		rNow, rok := e.resolveOperand(p.Right)
		lPrev, lpok := e.prevOperand(p.Left)
		rPrev, rpok := e.prevOperand(p.Right)
		if !lok || !rok || !lpok || !rpok {
			return false
		}
		return lPrev <= rPrev && lNow > rNow
	default:
		return false
	}
}

// resolveOperand returns the current numeric value of an operand and whether it
// is available this bar.
//
// TODO(candidate): implement this — it's the hook that "exposes" your indicators
// as DSL operands (ASSIGNMENT.md, task #2).
//
//   - A constant operand (Operand.Const != nil) resolves to that constant and is
//     always available.
//   - An indicator operand (Operand.Indicator != "") resolves to the current
//     value of the registered indicator in e.inds; it is unavailable until the
//     indicator reports ready.
//
// While this returns (0, false), no predicate can fire and the example strategy
// produces no trades.
func (e *Evaluator) resolveOperand(op spec.Operand) (float64, bool) {
	// TODO(candidate): replace this stub.
	return 0, false
}

// prevOperand returns the operand's value from the previous bar (CROSS_OVER).
func (e *Evaluator) prevOperand(op spec.Operand) (float64, bool) {
	if op.Const != nil {
		return *op.Const, true
	}
	v := e.prev[op.Indicator]
	return v, e.havePrev[op.Indicator]
}

func compare(l float64, op string, r float64) bool {
	switch op {
	case ">":
		return l > r
	case "<":
		return l < r
	case ">=":
		return l >= r
	case "<=":
		return l <= r
	case "==":
		return l == r
	default:
		return false
	}
}
