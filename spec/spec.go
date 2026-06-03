// Package spec is the (trimmed) strategy DSL. A StrategySpec describes one
// instrument's entry and exit rules; rules are predicates over indicator values.
//
// This is provided. You won't change the types — but you will make the engine
// resolve the indicators they reference (see engine/evaluator.go).
package spec

import (
	"encoding/json"
	"fmt"
	"os"
)

// StrategySpec is a single-instrument strategy.
type StrategySpec struct {
	Name     string  `json:"name"`
	Lookback int     `json:"lookback"` // window length (bars) for window-based indicators
	Side     string  `json:"side"`     // "LONG" or "SHORT"
	Qty      float64 `json:"qty"`
	Entry    Rule    `json:"entry"`
	Exit     Rule    `json:"exit"`
}

// Rule is a set of predicates combined with ALL (default) or ANY logic.
type Rule struct {
	Logic      string      `json:"logic"` // "ALL" or "ANY" (empty => ALL)
	Predicates []Predicate `json:"predicates"`
}

// Predicate compares two operands. Supported types:
//   - "RELATION":   left <op> right, where op is one of > < >= <= ==
//   - "CROSS_OVER": true on the bar where left crosses from <= right to > right
type Predicate struct {
	Type  string  `json:"type"`
	Left  Operand `json:"left"`
	Op    string  `json:"op"` // only used by RELATION
	Right Operand `json:"right"`
}

// Operand is either a named indicator or a constant. Exactly one should be set.
type Operand struct {
	Indicator string   `json:"indicator,omitempty"` // e.g. "IV_RANK", "IV_PERCENTILE"
	Const     *float64 `json:"const,omitempty"`
}

// Load reads and parses a strategy spec from a JSON file.
func Load(path string) (StrategySpec, error) {
	var sp StrategySpec
	b, err := os.ReadFile(path)
	if err != nil {
		return sp, err
	}
	if err := json.Unmarshal(b, &sp); err != nil {
		return sp, fmt.Errorf("parse %s: %w", path, err)
	}
	if sp.Lookback <= 0 {
		return sp, fmt.Errorf("%s: lookback must be > 0", path)
	}
	return sp, nil
}
