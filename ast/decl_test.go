package ast

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewSyntheticDecl(t *testing.T) {
	for _, tt := range []struct {
		atom         Atom
		declaredAtom Atom
	}{
		{
			atom: Atom{
				Predicate: PredicateSym{Symbol: "foo", Arity: 4},
				Args: []BaseTerm{
					Variable{"A"},
					String("constant 1"),
					String("constant 2"),
					Variable{"B"},
				},
			},
			declaredAtom: NewAtom("foo", Variable{"A"}, Variable{"X0"}, Variable{"X1"}, Variable{"B"}),
		},
		{
			atom: Atom{
				Predicate: PredicateSym{Symbol: "foo", Arity: 3},
				Args: []BaseTerm{
					String("constant"),
					Variable{"X0"},
					String("constant"),
				},
			},
			declaredAtom: NewAtom("foo", Variable{"X1"}, Variable{"X0"}, Variable{"X2"}),
		},
		{
			atom:         NewQuery(PredicateSym{Symbol: "foo", Arity: 3}),
			declaredAtom: NewAtom("foo", Variable{"X0"}, Variable{"X1"}, Variable{"X2"}),
		},
	} {
		got, err := NewSyntheticDecl(tt.atom)
		if err != nil {
			t.Errorf("NewSyntheticDecl(%v) = %v, unexpected error", tt.atom, err)
		} else if !cmp.Equal(got.DeclaredAtom, tt.declaredAtom) {
			t.Errorf("NewSyntheticDecl(%v) = %+v, want %+v", tt.atom, got.DeclaredAtom, tt.declaredAtom)
		}
	}
}
