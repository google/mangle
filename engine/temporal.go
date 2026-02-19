// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package engine

import (
	"fmt"
	"math"
	"time"

	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/factstore"
	"codeberg.org/TauCeti/mangle-go/functional"
	"codeberg.org/TauCeti/mangle-go/unionfind"
)

// TemporalEvaluator provides temporal reasoning capabilities for the engine.
type TemporalEvaluator struct {
	// The temporal fact store for time-indexed facts
	temporalStore factstore.TemporalFactStore
	// The current evaluation time (for relative time calculations)
	evaluationTime time.Time
}

// NewTemporalEvaluator creates a new temporal evaluator.
func NewTemporalEvaluator(store factstore.TemporalFactStore, evalTime time.Time) *TemporalEvaluator {
	return &TemporalEvaluator{
		temporalStore:  store,
		evaluationTime: evalTime,
	}
}

// EvalTemporalLiteral evaluates a temporal literal (an atom with temporal operator).
// Returns solutions (substitutions) that satisfy the temporal constraint.
func (te *TemporalEvaluator) EvalTemporalLiteral(
	tl ast.TemporalLiteral,
	subst unionfind.UnionFind,
) ([]unionfind.UnionFind, error) {
	// Extract the underlying atom
	atom, ok := tl.Literal.(ast.Atom)
	if !ok {
		return nil, fmt.Errorf("temporal literal must wrap an atom, got %T", tl.Literal)
	}

	// Evaluate the atom with current substitution
	evaledAtom, err := functional.EvalAtom(atom, subst)
	if err != nil {
		return nil, err
	}

	// If no temporal operator, this is just a regular lookup with optional interval binding/filtering
	if tl.Operator == nil {
		return te.evalTemporalAtomWithoutOperator(evaledAtom, tl.Interval, subst)
	}

	// Evaluate the temporal operator
	switch tl.Operator.Type {
	case ast.DiamondMinus:
		return te.evalDiamondMinus(evaledAtom, tl.Operator.Interval, tl.Interval, subst)
	case ast.BoxMinus:
		return te.evalBoxMinus(evaledAtom, tl.Operator.Interval, tl.Interval, subst)
	case ast.DiamondPlus:
		return te.evalDiamondPlus(evaledAtom, tl.Operator.Interval, tl.Interval, subst)
	case ast.BoxPlus:
		return te.evalBoxPlus(evaledAtom, tl.Operator.Interval, tl.Interval, subst)
	default:
		return nil, fmt.Errorf("unknown temporal operator type: %v", tl.Operator.Type)
	}
}

// evalTemporalAtomWithoutOperator looks up facts and optionally binds/filters by interval.
func (te *TemporalEvaluator) evalTemporalAtomWithoutOperator(
	atom ast.Atom,
	interval *ast.Interval,
	subst unionfind.UnionFind,
) ([]unionfind.UnionFind, error) {
	var solutions []unionfind.UnionFind

	// Resolve the query interval if provided
	var queryInterval ast.Interval
	var hasConcreteInterval bool
	var err error

	if interval != nil {
		// Try to resolve bounds to concrete timestamps
		resolved, err := ResolveHeadTime(interval, subst, te.evaluationTime)
		if err == nil && resolved != nil {
			queryInterval = *resolved
			hasConcreteInterval = true
		} else {
			// If we can't resolve (e.g. unbound variables), we'll filter after retrieval
			// But for now, let's assume we fetch all and filter/bind
		}
	}

	// Helper to process a fact
	processFact := func(tf factstore.TemporalFact) error {
		// Check if the fact is valid at current evaluation time (if no specific interval requested)
		// OR if it satisfies the requested interval
		if !hasConcreteInterval {
			// Only filter by evaluation time if no interval was requested at all
			if interval == nil && !tf.Interval.Contains(te.evaluationTime) {
				return nil
			}
		} else {
			// If we have a concrete query interval, what does it mean?
			// foo(X)@[t1, t2] usually means we are looking for facts that cover this interval
			// or are equal to it?
			// DatalogMTL semantics usually imply:
			// If interval is variable @[T], we bind T to the fact's interval.
			// If interval is concrete @[t1, t2], we filter facts that COVER [t1, t2].
			// Let's implement containment check: fact.Interval MUST CONTAIN queryInterval.
			if !te.intervalContains(tf.Interval, queryInterval) {
				return nil
			}
		}

		// Unify the fact with the query atom
		newSubst, err := unionfind.UnifyTermsExtend(atom.Args, tf.Atom.Args, subst)
		if err != nil {
			return nil // No match, continue
		}

		// Bind interval variables if present in the original query interval
		if interval != nil {
			newSubst = te.bindIntervalVariables(*interval, tf.Interval, newSubst)
		}

		solutions = append(solutions, newSubst)
		return nil
	}

	if hasConcreteInterval {
		// Optimization: Use GetFactsDuring to narrow down search
		err = te.temporalStore.GetFactsDuring(atom, queryInterval, processFact)
	} else {
		err = te.temporalStore.GetAllFacts(atom, processFact)
	}

	if err != nil {
		return nil, err
	}

	return solutions, nil
}

// evalDiamondMinus implements the diamond-minus operator (<-).
// It returns true if the atom was true at SOME point in the past interval.
// Syntax: <-[d1, d2] p(X) means "p(X) was true at some point between d1 and d2 ago"
func (te *TemporalEvaluator) evalDiamondMinus(
	atom ast.Atom,
	opInterval ast.Interval,
	interval *ast.Interval,
	subst unionfind.UnionFind,
) ([]unionfind.UnionFind, error) {
	// Calculate the query interval based on the operator's interval
	// The interval [d1, d2] means "from d2 ago to d1 ago"
	queryInterval, err := te.resolveOperatorInterval(opInterval)
	if err != nil {
		return nil, err
	}

	var solutions []unionfind.UnionFind

	// Query facts that overlap with the query interval
	err = te.temporalStore.GetFactsDuring(atom, queryInterval, func(tf factstore.TemporalFact) error {
		// Unify the fact with the query atom
		newSubst, err := unionfind.UnifyTermsExtend(atom.Args, tf.Atom.Args, subst)
		if err != nil {
			return nil // No match, continue
		}

		// Bind interval variables if present
		if interval != nil {
			newSubst = te.bindIntervalVariables(*interval, tf.Interval, newSubst)
		}

		solutions = append(solutions, newSubst)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return solutions, nil
}

// evalBoxMinus implements the box-minus operator ([-).
// It returns true if the atom was true CONTINUOUSLY throughout the past interval.
// Syntax: [-[d1, d2] p(X) means "p(X) was true for the entire period from d2 ago to d1 ago"
func (te *TemporalEvaluator) evalBoxMinus(
	atom ast.Atom,
	opInterval ast.Interval,
	interval *ast.Interval,
	subst unionfind.UnionFind,
) ([]unionfind.UnionFind, error) {
	// Calculate the query interval based on the operator's interval
	queryInterval, err := te.resolveOperatorInterval(opInterval)
	if err != nil {
		return nil, err
	}

	var solutions []unionfind.UnionFind

	// Query all facts matching the atom
	err = te.temporalStore.GetAllFacts(atom, func(tf factstore.TemporalFact) error {
		// Unify first to check if this fact matches
		newSubst, err := unionfind.UnifyTermsExtend(atom.Args, tf.Atom.Args, subst)
		if err != nil {
			return nil // No match, continue
		}

		// For box-minus, the fact's interval must CONTAIN the query interval
		// (the fact must be true throughout the entire query period)
		if te.intervalContains(tf.Interval, queryInterval) {
			// Bind interval variables if present
			if interval != nil {
				newSubst = te.bindIntervalVariables(*interval, tf.Interval, newSubst)
			}

			solutions = append(solutions, newSubst)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return solutions, nil
}

// resolveOperatorInterval converts an operator interval (with durations) to absolute timestamps.
// For past operators, durations are interpreted as offsets from the evaluation time.
func (te *TemporalEvaluator) resolveOperatorInterval(interval ast.Interval) (ast.Interval, error) {
	start, err := te.resolveBound(interval.Start, true) // true = past operator
	if err != nil {
		return ast.Interval{}, err
	}

	end, err := te.resolveBound(interval.End, true)
	if err != nil {
		return ast.Interval{}, err
	}

	// For past operators with duration bounds, the semantics are:
	// <-[0d, 7d] means "from 7 days ago to now"
	// So start should be the larger duration (further in past) and end the smaller
	return ast.NewInterval(end, start), nil // Note: swapped because past operators
}

// resolveBound converts a temporal bound to an absolute timestamp.
func (te *TemporalEvaluator) resolveBound(bound ast.TemporalBound, isPast bool) (ast.TemporalBound, error) {
	switch bound.Type {
	case ast.TimestampBound:
		// Already an absolute timestamp
		return bound, nil

	case ast.DurationTemporalBound:
		// Convert duration to absolute time relative to evaluation time
		duration := time.Duration(bound.Timestamp)
		if isPast {
			absoluteTime := te.evaluationTime.Add(-duration)
			return ast.NewTimestampBound(absoluteTime), nil
		}
		absoluteTime := te.evaluationTime.Add(duration)
		return ast.NewTimestampBound(absoluteTime), nil

	case ast.VariableBound:
		// Variables need to be resolved at runtime - return as-is for now
		return bound, nil

	case ast.NegativeInfinityBound, ast.PositiveInfinityBound:
		return bound, nil

	case ast.NowBound:
		// 'now' resolves to the current evaluation time
		return ast.NewTimestampBound(te.evaluationTime), nil

	default:
		return ast.TemporalBound{}, fmt.Errorf("unknown bound type: %v", bound.Type)
	}
}

// intervalContains checks if interval a fully contains interval b.
func (te *TemporalEvaluator) intervalContains(a, b ast.Interval) bool {
	// Handle unbounded intervals
	aStartOK := a.Start.Type == ast.NegativeInfinityBound
	bStartOK := b.Start.Type == ast.NegativeInfinityBound

	if !aStartOK && a.Start.Type == ast.TimestampBound {
		if !bStartOK && b.Start.Type == ast.TimestampBound {
			// a.Start must be <= b.Start
			if a.Start.Timestamp > b.Start.Timestamp {
				return false
			}
		} else if bStartOK {
			// b starts at -inf but a doesn't
			return false
		}
	}

	aEndOK := a.End.Type == ast.PositiveInfinityBound
	bEndOK := b.End.Type == ast.PositiveInfinityBound

	if !aEndOK && a.End.Type == ast.TimestampBound {
		if !bEndOK && b.End.Type == ast.TimestampBound {
			// a.End must be >= b.End
			if a.End.Timestamp < b.End.Timestamp {
				return false
			}
		} else if bEndOK {
			// b ends at +inf but a doesn't
			return false
		}
	}

	return true
}

// evalDiamondPlus implements the diamond-plus operator (<+).
// It returns true if the atom will be true at SOME point in the future interval.
// Syntax: <+[d1, d2] p(X) means "p(X) will be true at some point between d1 and d2 from now"
func (te *TemporalEvaluator) evalDiamondPlus(
	atom ast.Atom,
	opInterval ast.Interval,
	interval *ast.Interval,
	subst unionfind.UnionFind,
) ([]unionfind.UnionFind, error) {
	// Calculate the query interval based on the operator's interval
	queryInterval, err := te.resolveFutureOperatorInterval(opInterval)
	if err != nil {
		return nil, err
	}

	var solutions []unionfind.UnionFind

	// Query facts that overlap with the future query interval
	err = te.temporalStore.GetFactsDuring(atom, queryInterval, func(tf factstore.TemporalFact) error {
		// Unify the fact with the query atom
		newSubst, err := unionfind.UnifyTermsExtend(atom.Args, tf.Atom.Args, subst)
		if err != nil {
			return nil // No match, continue
		}

		// Bind interval variables if present
		if interval != nil {
			newSubst = te.bindIntervalVariables(*interval, tf.Interval, newSubst)
		}

		solutions = append(solutions, newSubst)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return solutions, nil
}

// evalBoxPlus implements the box-plus operator ([+).
// It returns true if the atom will be true CONTINUOUSLY throughout the future interval.
// Syntax: [+[d1, d2] p(X) means "p(X) will be true for the entire period from d1 to d2 from now"
func (te *TemporalEvaluator) evalBoxPlus(
	atom ast.Atom,
	opInterval ast.Interval,
	interval *ast.Interval,
	subst unionfind.UnionFind,
) ([]unionfind.UnionFind, error) {
	// Calculate the query interval based on the operator's interval
	queryInterval, err := te.resolveFutureOperatorInterval(opInterval)
	if err != nil {
		return nil, err
	}

	var solutions []unionfind.UnionFind

	// Query all facts matching the atom
	err = te.temporalStore.GetAllFacts(atom, func(tf factstore.TemporalFact) error {
		// Unify first to check if this fact matches
		newSubst, err := unionfind.UnifyTermsExtend(atom.Args, tf.Atom.Args, subst)
		if err != nil {
			return nil // No match, continue
		}

		// For box-plus, the fact's interval must CONTAIN the query interval
		// (the fact must be true throughout the entire future query period)
		if te.intervalContains(tf.Interval, queryInterval) {
			// Bind interval variables if present
			if interval != nil {
				newSubst = te.bindIntervalVariables(*interval, tf.Interval, newSubst)
			}

			solutions = append(solutions, newSubst)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return solutions, nil
}

// resolveFutureOperatorInterval converts a future operator interval to absolute timestamps.
func (te *TemporalEvaluator) resolveFutureOperatorInterval(interval ast.Interval) (ast.Interval, error) {
	start, err := te.resolveBound(interval.Start, false) // false = future operator
	if err != nil {
		return ast.Interval{}, err
	}

	end, err := te.resolveBound(interval.End, false)
	if err != nil {
		return ast.Interval{}, err
	}

	// For future operators, the interval is from start to end (not swapped)
	return ast.NewInterval(start, end), nil
}

// bindIntervalVariables binds interval variables in the query to the fact's interval.
// Supports binding @[T] (point/variable) or @[T1, T2] (range).
func (te *TemporalEvaluator) bindIntervalVariables(queryInterval ast.Interval, factInterval ast.Interval, subst unionfind.UnionFind) unionfind.UnionFind {
	newSubst := subst

	// Bind start variable if present
	if queryInterval.Start.Type == ast.VariableBound {
		// Create a constant for the start timestamp
		startNano := factstore.GetStartTime(factInterval)
		startConst := ast.Time(startNano)
		if s, err := unionfind.UnifyTermsExtend([]ast.BaseTerm{queryInterval.Start.Variable}, []ast.BaseTerm{startConst}, newSubst); err == nil {
			newSubst = s
		}
	}

	// Bind end variable if present
	if queryInterval.End.Type == ast.VariableBound {
		endNano := factstore.GetEndTime(factInterval)
		endConst := ast.Time(endNano)
		if s, err := unionfind.UnifyTermsExtend([]ast.BaseTerm{queryInterval.End.Variable}, []ast.BaseTerm{endConst}, newSubst); err == nil {
			newSubst = s
		}
	}

	return newSubst
}

// intervalToConstant converts an interval to a Mangle constant (pair of numbers).
func intervalToConstant(interval ast.Interval) ast.Constant {
	var startNano, endNano int64

	if interval.Start.Type == ast.TimestampBound {
		startNano = interval.Start.Timestamp
	} else if interval.Start.Type == ast.NegativeInfinityBound {
		startNano = math.MinInt64 // -inf
	}

	if interval.End.Type == ast.TimestampBound {
		endNano = interval.End.Timestamp
	} else if interval.End.Type == ast.PositiveInfinityBound {
		endNano = math.MaxInt64 // +inf
	}

	startConst := ast.Number(startNano)
	endConst := ast.Number(endNano)
	return ast.Pair(&startConst, &endConst)
}

// premiseTemporalLiteral evaluates a temporal literal premise.
// This is called from oneStepEvalPremise when encountering a TemporalLiteral.
func premiseTemporalLiteral(
	tl ast.TemporalLiteral,
	temporalStore factstore.TemporalFactStore,
	evalTime time.Time,
	subst unionfind.UnionFind,
) ([]unionfind.UnionFind, error) {
	te := NewTemporalEvaluator(temporalStore, evalTime)
	return te.EvalTemporalLiteral(tl, subst)
}

// ResolveHeadTime resolves variables and special bounds (like 'now') in a HeadTime interval
// using the given substitution. Returns the resolved interval or nil if HeadTime is nil.
func ResolveHeadTime(headTime *ast.Interval, subst unionfind.UnionFind, evalTime time.Time) (*ast.Interval, error) {
	if headTime == nil {
		return nil, nil
	}

	// Resolve start bound
	start, err := resolveBoundWithSubst(headTime.Start, subst, evalTime)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve start bound: %w", err)
	}

	// Resolve end bound
	end, err := resolveBoundWithSubst(headTime.End, subst, evalTime)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve end bound: %w", err)
	}

	resolved := ast.NewInterval(start, end)
	return &resolved, nil
}

// resolveBoundWithSubst resolves a temporal bound, substituting variables and handling 'now'.
func resolveBoundWithSubst(bound ast.TemporalBound, subst unionfind.UnionFind, evalTime time.Time) (ast.TemporalBound, error) {
	switch bound.Type {
	case ast.TimestampBound:
		return bound, nil

	case ast.VariableBound:
		// Look up the variable in the substitution
		term := subst.Get(bound.Variable)
		if term == nil {
			return ast.TemporalBound{}, fmt.Errorf("variable %v not bound in substitution", bound.Variable.Symbol)
		}

		// The term should be a constant (either a number representing nanoseconds or a pair for an interval)
		switch t := term.(type) {
		case ast.Constant:
			// If it's a number, interpret as nanoseconds since epoch
			if t.Type == ast.NumberType {
				nano, err := t.NumberValue()
				if err != nil {
					return ast.TemporalBound{}, fmt.Errorf("failed to get number value: %w", err)
				}
				return ast.TemporalBound{Type: ast.TimestampBound, Timestamp: nano}, nil
			}
			if t.Type == ast.TimeType {
				nano, err := t.TimeValue()
				if err != nil {
					return ast.TemporalBound{}, fmt.Errorf("failed to get time value: %w", err)
				}
				return ast.TemporalBound{Type: ast.TimestampBound, Timestamp: nano}, nil
			}
			return ast.TemporalBound{}, fmt.Errorf("expected number or time constant for temporal bound, got %v", t.Type)
		case ast.Variable:
			// Variable is still unbound
			return ast.TemporalBound{}, fmt.Errorf("variable %v not fully resolved", t.Symbol)
		default:
			return ast.TemporalBound{}, fmt.Errorf("unexpected term type for temporal bound: %T", term)
		}

	case ast.NegativeInfinityBound, ast.PositiveInfinityBound:
		return bound, nil

	case ast.NowBound:
		// 'now' resolves to the current evaluation time
		return ast.NewTimestampBound(evalTime), nil

	default:
		return ast.TemporalBound{}, fmt.Errorf("unknown bound type: %v", bound.Type)
	}
}

// DerivedTemporalFact represents a temporal fact derived from a rule.
type DerivedTemporalFact struct {
	Atom     ast.Atom
	Interval *ast.Interval
}

// EvalClauseWithTemporalHead evaluates a clause and produces temporal facts if the clause
// has a temporal annotation on its head. This is used for deriving new temporal facts.
func EvalClauseWithTemporalHead(
	clause ast.Clause,
	solutions []unionfind.UnionFind,
	evalTime time.Time,
) ([]DerivedTemporalFact, error) {
	var results []DerivedTemporalFact

	for _, sol := range solutions {
		// Evaluate the head atom
		head, err := functional.EvalAtom(clause.Head, sol)
		if err != nil {
			return nil, err
		}

		// Resolve the temporal annotation if present
		var interval *ast.Interval
		if clause.HeadTime != nil {
			interval, err = ResolveHeadTime(clause.HeadTime, sol, evalTime)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve HeadTime: %w", err)
			}
		}

		results = append(results, DerivedTemporalFact{
			Atom:     head,
			Interval: interval,
		})
	}

	return results, nil
}
