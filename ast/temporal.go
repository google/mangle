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

package ast

import (
	"fmt"
	"strings"
	"time"
)

// TemporalBoundType indicates the kind of temporal bound.
type TemporalBoundType int

const (
	// TimestampBound is a concrete point in time.
	TimestampBound TemporalBoundType = iota
	// VariableBound is a variable to be bound during evaluation.
	VariableBound
	// UnboundedBound represents positive or negative infinity.
	UnboundedBound
	// NowBound represents the current evaluation time.
	NowBound
)

// TemporalBound represents a point in time, which can be:
// - A concrete timestamp
// - A variable (to be bound during evaluation)
// - Unbounded (positive or negative infinity)
type TemporalBound struct {
	Type TemporalBoundType

	// For TimestampBound: the concrete time value.
	// Stored as Unix nanoseconds for precision.
	Timestamp int64

	// For VariableBound: the variable name.
	Variable Variable

	// For UnboundedBound: true if positive infinity, false if negative infinity.
	IsPositiveInf bool
}

// NewTimestampBound creates a bound from a time.Time value.
func NewTimestampBound(t time.Time) TemporalBound {
	return TemporalBound{
		Type:      TimestampBound,
		Timestamp: t.UnixNano(),
	}
}

// NewVariableBound creates a bound from a variable.
func NewVariableBound(v Variable) TemporalBound {
	return TemporalBound{
		Type:     VariableBound,
		Variable: v,
	}
}

// NegativeInfinity returns a bound representing negative infinity.
func NegativeInfinity() TemporalBound {
	return TemporalBound{
		Type:          UnboundedBound,
		IsPositiveInf: false,
	}
}

// PositiveInfinity returns a bound representing positive infinity.
func PositiveInfinity() TemporalBound {
	return TemporalBound{
		Type:          UnboundedBound,
		IsPositiveInf: true,
	}
}

// Now returns a bound representing the current evaluation time.
func Now() TemporalBound {
	return TemporalBound{
		Type: NowBound,
	}
}

// Time returns the time.Time value for a TimestampBound.
// Returns zero time for non-timestamp bounds.
func (tb TemporalBound) Time() time.Time {
	if tb.Type != TimestampBound {
		return time.Time{}
	}
	return time.Unix(0, tb.Timestamp)
}

// String returns a string representation of the temporal bound.
func (tb TemporalBound) String() string {
	switch tb.Type {
	case TimestampBound:
		t := time.Unix(0, tb.Timestamp).UTC()
		// Use ISO 8601 format
		return t.Format("2006-01-02T15:04:05Z")
	case VariableBound:
		return tb.Variable.String()
	case UnboundedBound:
		return "_"
	case NowBound:
		return "now"
	default:
		return "?"
	}
}

// Equals returns true if two temporal bounds are equal.
func (tb TemporalBound) Equals(other TemporalBound) bool {
	if tb.Type != other.Type {
		return false
	}
	switch tb.Type {
	case TimestampBound:
		return tb.Timestamp == other.Timestamp
	case VariableBound:
		return tb.Variable == other.Variable
	case UnboundedBound:
		return tb.IsPositiveInf == other.IsPositiveInf
	case NowBound:
		return true // All 'now' bounds are equal
	}
	return false
}

// Interval represents a time interval [Start, End].
// Both endpoints are inclusive.
type Interval struct {
	Start TemporalBound
	End   TemporalBound
}

// NewInterval creates an interval from two bounds.
func NewInterval(start, end TemporalBound) Interval {
	return Interval{Start: start, End: end}
}

// NewPointInterval creates an interval representing a single point in time.
func NewPointInterval(t time.Time) Interval {
	bound := NewTimestampBound(t)
	return Interval{Start: bound, End: bound}
}

// EternalInterval returns an interval representing all time (negative infinity to positive infinity).
// This is used for facts without temporal annotations.
func EternalInterval() Interval {
	return Interval{
		Start: NegativeInfinity(),
		End:   PositiveInfinity(),
	}
}

// IsEternal returns true if this interval represents all time.
func (i Interval) IsEternal() bool {
	return i.Start.Type == UnboundedBound && !i.Start.IsPositiveInf &&
		i.End.Type == UnboundedBound && i.End.IsPositiveInf
}

// IsPoint returns true if this interval represents a single point in time.
func (i Interval) IsPoint() bool {
	return i.Start.Type == TimestampBound &&
		i.End.Type == TimestampBound &&
		i.Start.Timestamp == i.End.Timestamp
}

// String returns a string representation of the interval.
func (i Interval) String() string {
	if i.IsEternal() {
		return "" // Eternal intervals have no annotation
	}
	if i.IsPoint() {
		return fmt.Sprintf("@[%s]", i.Start.String())
	}
	return fmt.Sprintf("@[%s, %s]", i.Start.String(), i.End.String())
}

// Equals returns true if two intervals are equal.
func (i Interval) Equals(other Interval) bool {
	return i.Start.Equals(other.Start) && i.End.Equals(other.End)
}

// Contains returns true if time t is within this interval.
// Only works for intervals with concrete timestamp bounds.
func (i Interval) Contains(t time.Time) bool {
	tNano := t.UnixNano()

	// Check start bound
	switch i.Start.Type {
	case TimestampBound:
		if tNano < i.Start.Timestamp {
			return false
		}
	case UnboundedBound:
		if i.Start.IsPositiveInf {
			return false // Start is +inf, nothing can be after it
		}
		// Start is -inf, all times are after it
	case VariableBound:
		return false // Can't evaluate with unbound variable
	}

	// Check end bound
	switch i.End.Type {
	case TimestampBound:
		if tNano > i.End.Timestamp {
			return false
		}
	case UnboundedBound:
		if !i.End.IsPositiveInf {
			return false // End is -inf, nothing can be before it
		}
		// End is +inf, all times are before it
	case VariableBound:
		return false // Can't evaluate with unbound variable
	}

	return true
}

// Overlaps returns true if this interval overlaps with another.
// Only works for intervals with concrete timestamp bounds.
func (i Interval) Overlaps(other Interval) bool {
	// Two intervals overlap if neither ends before the other starts

	// Check if i ends before other starts
	if i.End.Type == TimestampBound && other.Start.Type == TimestampBound {
		if i.End.Timestamp < other.Start.Timestamp {
			return false
		}
	}

	// Check if other ends before i starts
	if other.End.Type == TimestampBound && i.Start.Type == TimestampBound {
		if other.End.Timestamp < i.Start.Timestamp {
			return false
		}
	}

	return true
}

// TemporalOperatorType represents the type of temporal operator.
type TemporalOperatorType int

const (
	// DiamondMinus: true at some point in the past interval
	// Syntax: <-[duration]
	DiamondMinus TemporalOperatorType = iota

	// BoxMinus: true continuously throughout the past interval
	// Syntax: [-[duration]
	BoxMinus

	// DiamondPlus: true at some point in the future interval
	// Syntax: <+[duration]
	DiamondPlus

	// BoxPlus: true continuously throughout the future interval
	// Syntax: [+[duration]
	BoxPlus
)

// TemporalOperator represents a metric temporal operator applied to a literal.
type TemporalOperator struct {
	Type     TemporalOperatorType
	Interval Interval
}

// String returns the string representation of a temporal operator.
func (op TemporalOperator) String() string {
	var prefix string
	switch op.Type {
	case DiamondMinus:
		prefix = "<-"
	case BoxMinus:
		prefix = "[-"
	case DiamondPlus:
		prefix = "<+"
	case BoxPlus:
		prefix = "[+"
	}
	return fmt.Sprintf("%s[%s, %s]", prefix, op.Interval.Start.String(), op.Interval.End.String())
}

// TemporalAtom wraps an Atom with an optional temporal annotation.
// This is used in clause heads and premises to represent temporally-annotated facts.
type TemporalAtom struct {
	Atom     Atom
	Interval *Interval // nil means eternal (no temporal annotation)
}

func (ta TemporalAtom) isTerm() {}

// String returns a string representation of the temporal atom.
func (ta TemporalAtom) String() string {
	if ta.Interval == nil || ta.Interval.IsEternal() {
		return ta.Atom.String()
	}
	return ta.Atom.String() + ta.Interval.String()
}

// Equals returns true if two temporal atoms are equal.
func (ta TemporalAtom) Equals(u Term) bool {
	other, ok := u.(TemporalAtom)
	if !ok {
		return false
	}
	if !ta.Atom.Equals(other.Atom) {
		return false
	}
	// Both nil = equal
	if ta.Interval == nil && other.Interval == nil {
		return true
	}
	// One nil, one not = not equal
	if ta.Interval == nil || other.Interval == nil {
		return false
	}
	return ta.Interval.Equals(*other.Interval)
}

// ApplySubst applies a substitution to the temporal atom.
// Note: Interval bounds with VariableBound type are resolved at query time
// by the TemporalEvaluator, not during substitution application. This is
// intentional as interval variable resolution requires temporal context.
func (ta TemporalAtom) ApplySubst(s Subst) Term {
	newAtom := ta.Atom.ApplySubst(s).(Atom)
	return TemporalAtom{Atom: newAtom, Interval: ta.Interval}
}

// TemporalLiteral represents a literal (atom or negated atom) with an optional temporal operator.
// This is used in clause premises.
type TemporalLiteral struct {
	// The underlying literal (Atom or NegAtom)
	Literal Term

	// Optional temporal operator (nil if none)
	Operator *TemporalOperator

	// Optional interval binding for the literal's validity time
	IntervalVar *Variable
}

func (tl TemporalLiteral) isTerm() {}

// String returns a string representation of the temporal literal.
func (tl TemporalLiteral) String() string {
	var sb strings.Builder
	if tl.Operator != nil {
		sb.WriteString(tl.Operator.String())
		sb.WriteString(" ")
	}
	sb.WriteString(tl.Literal.String())
	if tl.IntervalVar != nil {
		sb.WriteString("@")
		sb.WriteString(tl.IntervalVar.String())
	}
	return sb.String()
}

// Equals returns true if two temporal literals are equal.
func (tl TemporalLiteral) Equals(u Term) bool {
	other, ok := u.(TemporalLiteral)
	if !ok {
		return false
	}
	if !tl.Literal.Equals(other.Literal) {
		return false
	}
	// Compare operators
	if (tl.Operator == nil) != (other.Operator == nil) {
		return false
	}
	if tl.Operator != nil && *tl.Operator != *other.Operator {
		return false
	}
	// Compare interval variables
	if (tl.IntervalVar == nil) != (other.IntervalVar == nil) {
		return false
	}
	if tl.IntervalVar != nil && *tl.IntervalVar != *other.IntervalVar {
		return false
	}
	return true
}

// ApplySubst applies a substitution to the temporal literal.
// Note: The IntervalVar is a binding variable that captures the fact's validity
// interval during evaluation. It is not substituted here because it is populated
// by the TemporalEvaluator when the literal is matched against temporal facts.
func (tl TemporalLiteral) ApplySubst(s Subst) Term {
	newLiteral := tl.Literal.ApplySubst(s)
	return TemporalLiteral{
		Literal:     newLiteral,
		Operator:    tl.Operator,
		IntervalVar: tl.IntervalVar,
	}
}
