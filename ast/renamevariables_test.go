// Copyright 2026 The Mangle Authors
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

import "testing"

func TestRenameVariables(t *testing.T) {
	pred := PredicateSym{Symbol: "p", Arity: 2}
	fn := FunctionSym{Symbol: "fn:max", Arity: 1}
	// rename prefixes every variable with "R".
	rename := func(v Variable) Variable { return Variable{"R" + v.Symbol} }
	for _, test := range []struct {
		name string
		in   Term
		want Term
	}{
		{"variable", Variable{"X"}, Variable{"RX"}},
		{"constant", Constant{}, Constant{}},
		{"atom", Atom{pred, []BaseTerm{Variable{"X"}, Constant{}}},
			Atom{pred, []BaseTerm{Variable{"RX"}, Constant{}}}},
		{"neg atom", NegAtom{Atom{pred, []BaseTerm{Variable{"X"}, Variable{"Y"}}}},
			NegAtom{Atom{pred, []BaseTerm{Variable{"RX"}, Variable{"RY"}}}}},
		{"equality", Eq{Variable{"X"}, Constant{}}, Eq{Variable{"RX"}, Constant{}}},
		{"inequality", Ineq{Constant{}, Variable{"X"}}, Ineq{Constant{}, Variable{"RX"}}},
		{"apply fn", ApplyFn{fn, []BaseTerm{Variable{"X"}}},
			ApplyFn{fn, []BaseTerm{Variable{"RX"}}}},
		{"temporal atom", TemporalAtom{
			Atom:     Atom{pred, []BaseTerm{Variable{"X"}, Variable{"Y"}}},
			Interval: &Interval{Start: NewVariableBound(Variable{"S"}), End: NewVariableBound(Variable{"E"})},
		}, TemporalAtom{
			Atom:     Atom{pred, []BaseTerm{Variable{"RX"}, Variable{"RY"}}},
			Interval: &Interval{Start: NewVariableBound(Variable{"RS"}), End: NewVariableBound(Variable{"RE"})},
		}},
		{"temporal literal", TemporalLiteral{
			Literal:  NegAtom{Atom{pred, []BaseTerm{Variable{"X"}, Variable{"Y"}}}},
			Operator: &TemporalOperator{Type: DiamondMinus, Interval: Interval{Start: NewVariableBound(Variable{"S"}), End: NewVariableBound(Variable{"E"})}},
			Interval: &Interval{Start: NewVariableBound(Variable{"T"}), End: PositiveInfinity()},
		}, TemporalLiteral{
			Literal:  NegAtom{Atom{pred, []BaseTerm{Variable{"RX"}, Variable{"RY"}}}},
			Operator: &TemporalOperator{Type: DiamondMinus, Interval: Interval{Start: NewVariableBound(Variable{"RS"}), End: NewVariableBound(Variable{"RE"})}},
			Interval: &Interval{Start: NewVariableBound(Variable{"RT"}), End: PositiveInfinity()},
		}},
	} {
		got := RenameVariables(test.in, rename)
		if !got.Equals(test.want) {
			t.Errorf("%s: RenameVariables(%v) = %v, want %v", test.name, test.in, got, test.want)
		}
	}
}

// A term without renamed variables is returned unchanged.
func TestRenameVariablesKeepsTermWithoutRenamedVariables(t *testing.T) {
	pred := PredicateSym{Symbol: "p", Arity: 1}
	atom := Atom{pred, []BaseTerm{Variable{"X"}, Constant{}}}
	same := RenameVariables(atom, func(v Variable) Variable { return v })
	if !same.Equals(atom) {
		t.Errorf("RenameVariables(%v) = %v, want %v", atom, same, atom)
	}
	if base := RenameVariablesBase(Variable{"X"}, func(v Variable) Variable { return v }); base != (BaseTerm)(Variable{"X"}) {
		t.Errorf("RenameVariablesBase(X) = %v, want X", base)
	}
}

// ReplaceWildcards is renaming: each wildcard becomes a fresh variable,
// including one in a temporal interval bound.
func TestReplaceWildcardsTemporal(t *testing.T) {
	pred := PredicateSym{Symbol: "p", Arity: 1}
	used := map[Variable]bool{Variable{"X"}: true}
	ta := TemporalAtom{
		Atom:     Atom{pred, []BaseTerm{Variable{"X"}}},
		Interval: &Interval{Start: NewVariableBound(Variable{"_"}), End: PositiveInfinity()},
	}
	got := ReplaceWildcards(used, ta).(TemporalAtom)
	start := got.Interval.Start.Variable
	if start.Symbol == "_" || start.Symbol == "X" {
		t.Errorf("wildcard in interval bound not replaced by a fresh variable: %v", got)
	}
}
