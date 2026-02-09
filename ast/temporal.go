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
	"sync"
	"time"
)

// defaultTimezone is the timezone used by Date, DateTime, and DateInterval helpers.
// Defaults to UTC. Use SetDefaultTimezone to change it.
var (
	defaultTimezone   = time.UTC
	defaultTimezoneMu sync.RWMutex
)

// SetTimezone sets the timezone used by Date, DateTime, DateTimeSec,
// DateInterval, and related helper functions. Defaults to UTC.
//
// Accepts timezone names (IANA format), common abbreviations, or special values:
//   - "UTC", "utc" - Coordinated Universal Time
//   - "Local", "local" - System local timezone
//   - "America/New_York", "America/Los_Angeles", etc. - IANA timezone names
//   - "EST", "PST", "CET", etc. - Common abbreviations (mapped to IANA names)
//
// Call this once at program startup to ensure consistent timezone handling.
//
// Example:
//
//	ast.SetTimezone("UTC")                    // Default
//	ast.SetTimezone("Local")                  // System timezone
//	ast.SetTimezone("America/New_York")       // IANA name
//	ast.SetTimezone("PST")                    // Abbreviation
func SetTimezone(tz string) error {
	loc, err := parseTimezone(tz)
	if err != nil {
		return err
	}
	defaultTimezoneMu.Lock()
	defaultTimezone = loc
	defaultTimezoneMu.Unlock()
	return nil
}

// MustSetTimezone is like SetTimezone but panics on error.
// Useful for program initialization where invalid timezone is a fatal error.
func MustSetTimezone(tz string) {
	if err := SetTimezone(tz); err != nil {
		panic(fmt.Sprintf("ast.MustSetTimezone(%q): %v", tz, err))
	}
}

// SetDefaultTimezone sets the timezone using a *time.Location directly.
// Prefer SetTimezone(string) for simpler usage.
func SetDefaultTimezone(loc *time.Location) {
	if loc == nil {
		loc = time.UTC
	}
	defaultTimezoneMu.Lock()
	defaultTimezone = loc
	defaultTimezoneMu.Unlock()
}

// GetDefaultTimezone returns the currently configured default timezone.
func GetDefaultTimezone() *time.Location {
	defaultTimezoneMu.RLock()
	defer defaultTimezoneMu.RUnlock()
	return defaultTimezone
}

// Common timezone abbreviation mappings
var timezoneAbbreviations = map[string]string{
	// US timezones
	"EST":  "America/New_York",
	"EDT":  "America/New_York",
	"CST":  "America/Chicago",
	"CDT":  "America/Chicago",
	"MST":  "America/Denver",
	"MDT":  "America/Denver",
	"PST":  "America/Los_Angeles",
	"PDT":  "America/Los_Angeles",
	"AKST": "America/Anchorage",
	"AKDT": "America/Anchorage",
	"HST":  "Pacific/Honolulu",
	// European timezones
	"GMT":  "Europe/London",
	"BST":  "Europe/London",
	"CET":  "Europe/Paris",
	"CEST": "Europe/Paris",
	"EET":  "Europe/Helsinki",
	"EEST": "Europe/Helsinki",
	// Asian timezones
	"JST":       "Asia/Tokyo",
	"KST":       "Asia/Seoul",
	"CST_CHINA": "Asia/Shanghai",
	"IST":       "Asia/Kolkata",
	// Australian timezones
	"AEST": "Australia/Sydney",
	"AEDT": "Australia/Sydney",
	"AWST": "Australia/Perth",
}

func parseTimezone(tz string) (*time.Location, error) {
	switch strings.ToLower(tz) {
	case "utc", "":
		return time.UTC, nil
	case "local":
		return time.Local, nil
	}

	// Check abbreviations first
	if ianaName, ok := timezoneAbbreviations[strings.ToUpper(tz)]; ok {
		return time.LoadLocation(ianaName)
	}

	// Try as IANA name directly
	return time.LoadLocation(tz)
}

// TemporalBoundType indicates the kind of temporal bound.
type TemporalBoundType int

const (
	// TimestampBound is a concrete point in time.
	TimestampBound TemporalBoundType = iota
	// VariableBound is a variable to be bound during evaluation.
	VariableBound
	// NegativeInfinityBound represents negative infinity.
	NegativeInfinityBound
	// PositiveInfinityBound represents positive infinity.
	PositiveInfinityBound
	// NowBound represents the current evaluation time.
	NowBound
	// DurationTemporalBound represents a relative duration.
	DurationTemporalBound
)

// TemporalBound represents a point in time, which can be:
// - A concrete timestamp
// - A variable (to be bound during evaluation)
// - Unbounded (positive or negative infinity)
// - A duration (relative to reference time)
type TemporalBound struct {
	Type TemporalBoundType

	// For TimestampBound: the concrete time value (Unix nanoseconds).
	// For DurationTemporalBound: the duration value (nanoseconds).
	Timestamp int64

	// For VariableBound: the variable name.
	Variable Variable
}

// NewTimestampBound creates a bound from a time.Time value.
func NewTimestampBound(t time.Time) TemporalBound {
	return TemporalBound{
		Type:      TimestampBound,
		Timestamp: t.UnixNano(),
	}
}

// NewDurationBound creates a bound from a time.Duration value.
func NewDurationBound(d time.Duration) TemporalBound {
	return TemporalBound{
		Type:      DurationTemporalBound,
		Timestamp: d.Nanoseconds(),
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
		Type: NegativeInfinityBound,
	}
}

// PositiveInfinity returns a bound representing positive infinity.
func PositiveInfinity() TemporalBound {
	return TemporalBound{
		Type: PositiveInfinityBound,
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

// Date creates a time.Time for the given year, month, and day at midnight.
// Uses the default timezone (UTC unless changed via SetDefaultTimezone).
func Date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, GetDefaultTimezone())
}

// DateTime creates a time.Time for the given date and time.
// Uses the default timezone (UTC unless changed via SetDefaultTimezone).
func DateTime(year int, month time.Month, day, hour, min int) time.Time {
	return time.Date(year, month, day, hour, min, 0, 0, GetDefaultTimezone())
}

// DateTimeSec creates a time.Time for the given date and time with seconds.
// Uses the default timezone (UTC unless changed via SetDefaultTimezone).
func DateTimeSec(year int, month time.Month, day, hour, min, sec int) time.Time {
	return time.Date(year, month, day, hour, min, sec, 0, GetDefaultTimezone())
}

// DateIn creates a time.Time for the given year, month, and day in a specific timezone.
// Accepts timezone names like "America/New_York", "PST", "UTC", "Local".
// Use this when you need a different timezone than the default for a specific value.
func DateIn(year int, month time.Month, day int, tz string) time.Time {
	loc, err := parseTimezone(tz)
	if err != nil {
		loc = GetDefaultTimezone()
	}
	return time.Date(year, month, day, 0, 0, 0, 0, loc)
}

// DateTimeIn creates a time.Time for the given date and time in a specific timezone.
// Accepts timezone names like "America/New_York", "PST", "UTC", "Local".
// Use this when you need a different timezone than the default for a specific value.
func DateTimeIn(year int, month time.Month, day, hour, min int, tz string) time.Time {
	loc, err := parseTimezone(tz)
	if err != nil {
		loc = GetDefaultTimezone()
	}
	return time.Date(year, month, day, hour, min, 0, 0, loc)
}

// TimeInterval creates an interval from two time.Time values.
// This is a convenience function for creating intervals without calling NewTimestampBound.
func TimeInterval(start, end time.Time) Interval {
	return Interval{
		Start: NewTimestampBound(start),
		End:   NewTimestampBound(end),
	}
}

// DateInterval creates an interval from two dates (year, month, day).
// This is the most concise way to create a date-based interval.
func DateInterval(startYear int, startMonth time.Month, startDay int, endYear int, endMonth time.Month, endDay int) Interval {
	return TimeInterval(
		Date(startYear, startMonth, startDay),
		Date(endYear, endMonth, endDay),
	)
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
	case NegativeInfinityBound, PositiveInfinityBound:
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
	case NegativeInfinityBound, PositiveInfinityBound:
		return true // Type equality is enough
	case NowBound:
		return true // All 'now' bounds are equal
	}
	return false
}

// Interval represents a time interval [Start, End].
// Both endpoints are inclusive. Start may be NegativeInfinity
// and End may be PositiveInfinity, which represents a half-open
// or open interval.
type Interval struct {
	Start TemporalBound
	End   TemporalBound
}

// NewInterval creates an interval from two bounds.
// Silently changes Start to NegativeInfinity
// and End to PositiveInfinity in case invalid bounds are passed.
func NewInterval(start, end TemporalBound) Interval {
	if start.Type == PositiveInfinityBound {
		start = NegativeInfinity()
	}
	if end.Type == NegativeInfinityBound {
		end = PositiveInfinity()
	}
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
	return i.Start.Type == NegativeInfinityBound &&
		i.End.Type == PositiveInfinityBound
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
	case NegativeInfinityBound:
		// Start is -inf, all times are after it
	case PositiveInfinityBound:
		return false // Start is +inf, nothing can be after it
	case VariableBound:
		return false // Can't evaluate with unbound variable
	}

	// Check end bound
	switch i.End.Type {
	case TimestampBound:
		if tNano > i.End.Timestamp {
			return false
		}
	case NegativeInfinityBound:
		return false // End is -inf, nothing can be before it
	case PositiveInfinityBound:
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
	// DiamondMinus operator: true at some point in the past interval
	// Syntax: <-[duration]
	DiamondMinus TemporalOperatorType = iota

	// BoxMinus operator: true continuously throughout the past interval
	// Syntax: [-[duration]
	BoxMinus

	// DiamondPlus operator: true at some point in the future interval
	// Syntax: <+[duration]
	DiamondPlus

	// BoxPlus operator: true continuously throughout the future interval
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

	// Optional interval annotation for the literal.
	// Can be a variable binding @[T] or a concrete interval @[t1, t2].
	Interval *Interval
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
	if tl.Interval != nil {
		sb.WriteString(tl.Interval.String())
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
	// Compare intervals
	if (tl.Interval == nil) != (other.Interval == nil) {
		return false
	}
	if tl.Interval != nil && !tl.Interval.Equals(*other.Interval) {
		return false
	}
	return true
}

// ApplySubst applies a substitution to the temporal literal.
func (tl TemporalLiteral) ApplySubst(s Subst) Term {
	newLiteral := tl.Literal.ApplySubst(s)

	var newInterval *Interval
	if tl.Interval != nil {
		// Apply substitution to interval bounds
		newStart := tl.Interval.Start
		if newStart.Type == VariableBound {
			if repl := s.Get(newStart.Variable); repl != nil {
				if c, ok := repl.(Constant); ok && c.Type == NumberType {
					newStart = TemporalBound{Type: TimestampBound, Timestamp: c.NumValue}
				} else if v, ok := repl.(Variable); ok {
					newStart = NewVariableBound(v)
				}
			}
		}

		newEnd := tl.Interval.End
		if newEnd.Type == VariableBound {
			if repl := s.Get(newEnd.Variable); repl != nil {
				if c, ok := repl.(Constant); ok && c.Type == NumberType {
					newEnd = TemporalBound{Type: TimestampBound, Timestamp: c.NumValue}
				} else if v, ok := repl.(Variable); ok {
					newEnd = NewVariableBound(v)
				}
			}
		}

		i := NewInterval(newStart, newEnd)
		newInterval = &i
	}

	return TemporalLiteral{
		Literal:  newLiteral,
		Operator: tl.Operator,
		Interval: newInterval,
	}
}
