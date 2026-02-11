package ast

import (
	"testing"
)

func TestIntervalString_Point(t *testing.T) {
	v := Variable{Symbol: "T"}
	bound := NewVariableBound(v)
	interval := NewInterval(bound, bound)

	got := interval.String()
	want := "@[T]"
	if got != want {
		t.Errorf("Interval{T, T}.String() = %q, want %q", got, want)
	}
}

func TestTemporalAtomString_Point(t *testing.T) {
	v := Variable{Symbol: "T"}
	bound := NewVariableBound(v)
	interval := NewInterval(bound, bound)
	atom := NewAtom("p", Variable{Symbol: "X"})
	ta := TemporalAtom{Atom: atom, Interval: &interval}

	got := ta.String()
	want := "p(X)@[T]"
	if got != want {
		t.Errorf("TemporalAtom{p(X), @[T]}.String() = %q, want %q", got, want)
	}
}
