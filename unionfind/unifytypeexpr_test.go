package unionfind

import (
	"testing"

	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/symbols"
)

var (
	listNum    = symbols.NewListType(ast.NumberBound)
	listStr    = symbols.NewListType(ast.StringBound)
	listFloat  = symbols.NewListType(ast.Float64Bound)
	numOrFloat = symbols.NewUnionType(ast.NumberBound, ast.Float64Bound)
	numOrStr   = symbols.NewUnionType(ast.NumberBound, ast.StringBound)
	floatOrNum = symbols.NewUnionType(ast.Float64Bound, ast.NumberBound)
	x          = ast.Variable{"X"}
)

func unify(t *testing.T, xs, ys []ast.BaseTerm) (map[ast.Variable]ast.BaseTerm, error) {
	t.Helper()
	return UnifyTypeExpr(xs, ys)
}

func TestUnifyTypeExprMismatchingConstants(t *testing.T) {
	// Mismatching constants must not unify. The error used to be
	// dropped for constants inside type constructor applications,
	// letting e.g. fn:List(/number) unify with fn:List(/duration).
	for _, pair := range [][2]ast.BaseTerm{
		{ast.NumberBound, ast.DurationBound},
		{listNum, symbols.NewListType(ast.DurationBound)},
		{symbols.NewMapType(ast.NumberBound, ast.StringBound), symbols.NewMapType(ast.StringBound, ast.StringBound)},
		{symbols.NewTupleType(ast.NumberBound, ast.StringBound), symbols.NewTupleType(ast.NumberBound, ast.NumberBound)},
		{symbols.NewStructType(ast.NameBound, ast.NumberBound), symbols.NewStructType(ast.NameBound, ast.StringBound)},
	} {
		if _, err := unify(t, []ast.BaseTerm{pair[0]}, []ast.BaseTerm{pair[1]}); err == nil {
			t.Errorf("UnifyTypeExpr(%v, %v) succeeded, want error", pair[0], pair[1])
		}
	}
	// Matching constants do unify.
	if _, err := unify(t, []ast.BaseTerm{listNum}, []ast.BaseTerm{listNum}); err != nil {
		t.Errorf("UnifyTypeExpr(%v, %v) failed: %v", listNum, listNum, err)
	}
}

func TestUnifyTypeExprArityMismatch(t *testing.T) {
	// Type constructor applications with a different number of
	// arguments must fail with an error, not panic.
	pairs := [][2]ast.BaseTerm{
		{symbols.NewListType(ast.NumberBound), ast.ApplyFn{symbols.ListType, nil}},
		{symbols.NewTupleType(ast.NumberBound, ast.StringBound), symbols.NewTupleType(ast.NumberBound)},
		{symbols.NewStructType(ast.NameBound, ast.NumberBound), symbols.NewStructType(ast.NameBound, ast.NumberBound, ast.NameBound, ast.StringBound)},
	}
	for _, pair := range pairs {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("UnifyTypeExpr(%v, %v) panicked: %v", pair[0], pair[1], r)
				}
			}()
			if _, err := unify(t, []ast.BaseTerm{pair[0]}, []ast.BaseTerm{pair[1]}); err == nil {
				t.Errorf("UnifyTypeExpr(%v, %v) succeeded, want error", pair[0], pair[1])
			}
		}()
	}
}

func TestUnifyTypeExprVariables(t *testing.T) {
	// Type variables bind to the other side, inside constructor
	// applications...
	subst, err := unify(t, []ast.BaseTerm{symbols.NewListType(x)}, []ast.BaseTerm{listNum})
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if got, want := subst[x], ast.NumberBound; !got.Equals(want) {
		t.Errorf("subst[%v] = %v, want %v", x, got, want)
	}
	// ... and wholesale to type constructor applications, including
	// unions.
	subst, err = unify(t, []ast.BaseTerm{symbols.NewListType(numOrFloat)}, []ast.BaseTerm{symbols.NewListType(x)})
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if got, want := subst[x], numOrFloat; !got.Equals(want) {
		t.Errorf("subst[%v] = %v, want %v", x, got, want)
	}
	// Conflicting bindings fail.
	_, err = unify(t,
		[]ast.BaseTerm{symbols.NewListType(x), symbols.NewListType(x)},
		[]ast.BaseTerm{listNum, listStr})
	if err == nil {
		t.Error("conflicting bindings for X succeeded, want error")
	}
}

func TestUnifyTypeExprUnionMembership(t *testing.T) {
	// An actual type unifies with a declared union if it unifies with
	// one of its alternatives, in any position.
	for _, pair := range [][2]ast.BaseTerm{
		{ast.NumberBound, numOrFloat},
		{ast.Float64Bound, numOrFloat},
		{listNum, symbols.NewListType(numOrFloat)},
		{listFloat, symbols.NewListType(numOrFloat)},
	} {
		if _, err := unify(t, []ast.BaseTerm{pair[0]}, []ast.BaseTerm{pair[1]}); err != nil {
			t.Errorf("UnifyTypeExpr(%v, %v) failed: %v", pair[0], pair[1], err)
		}
	}
	// Non-members do not.
	for _, pair := range [][2]ast.BaseTerm{
		{ast.DurationBound, numOrFloat},
		{listStr, symbols.NewListType(numOrFloat)},
	} {
		if _, err := unify(t, []ast.BaseTerm{pair[0]}, []ast.BaseTerm{pair[1]}); err == nil {
			t.Errorf("UnifyTypeExpr(%v, %v) succeeded, want error", pair[0], pair[1])
		}
	}
}

func TestUnifyTypeExprUnionSubset(t *testing.T) {
	// An actual union unifies with a declared type only if all of its
	// alternatives do.
	for _, pair := range [][2]ast.BaseTerm{
		{numOrFloat, floatOrNum}, // order does not matter
		{numOrFloat, ast.AnyBound},
	} {
		if _, err := unify(t, []ast.BaseTerm{pair[0]}, []ast.BaseTerm{pair[1]}); err != nil {
			t.Errorf("UnifyTypeExpr(%v, %v) failed: %v", pair[0], pair[1], err)
		}
	}
	for _, pair := range [][2]ast.BaseTerm{
		{numOrFloat, ast.Float64Bound}, // /number is not a /float64
		{numOrFloat, numOrStr},         // /float64 is not a member
		{numOrFloat, symbols.NewListType(numOrFloat)},
	} {
		if _, err := unify(t, []ast.BaseTerm{pair[0]}, []ast.BaseTerm{pair[1]}); err == nil {
			t.Errorf("UnifyTypeExpr(%v, %v) succeeded, want error", pair[0], pair[1])
		}
	}
}

func TestUnifyTypeExprWildcardBounds(t *testing.T) {
	// The /any bound is the top of the type lattice and /bot the
	// bottom; neither conflicts with any other type, in any position.
	for _, pair := range [][2]ast.BaseTerm{
		{ast.AnyBound, ast.NumberBound},
		{ast.NumberBound, ast.AnyBound},
		{ast.BotBound, ast.NumberBound},
		{listNum, symbols.NewListType(ast.AnyBound)},
		{numOrFloat, ast.AnyBound},
		{symbols.NewListType(ast.BotBound), listNum}, // empty list
	} {
		if _, err := unify(t, []ast.BaseTerm{pair[0]}, []ast.BaseTerm{pair[1]}); err != nil {
			t.Errorf("UnifyTypeExpr(%v, %v) failed: %v", pair[0], pair[1], err)
		}
	}
	// Variables still bind to wildcard bounds.
	subst, err := unify(t, []ast.BaseTerm{symbols.NewListType(x)}, []ast.BaseTerm{symbols.NewListType(ast.AnyBound)})
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if got, want := subst[x], ast.AnyBound; !got.Equals(want) {
		t.Errorf("subst[%v] = %v, want %v", x, got, want)
	}
}

func TestUnifyTypeExprWildcardBindingThenConstraint(t *testing.T) {
	// A type variable that first unifies with a wildcard (here: the
	// /bot element type of the empty list) must remain compatible with
	// later, concrete constraints. This is the shape produced by e.g.
	// fn:list:contains([], 23) against the signature
	// fn:Bool() <= fn:List(X), X.
	subst, err := unify(t,
		[]ast.BaseTerm{symbols.NewListType(ast.BotBound), ast.NumberBound},
		[]ast.BaseTerm{symbols.NewListType(x), x})
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	// The variable may end up bound to the wildcard; either way the
	// unification must succeed without conflict.
	if got := subst[x]; !got.Equals(ast.BotBound) && !got.Equals(ast.NumberBound) && !got.Equals(x) {
		t.Errorf("subst[%v] = %v, want %v, %v or %v", x, got, ast.BotBound, ast.NumberBound, x)
	}
}

func TestUnifyTypeExprUnionBacktracking(t *testing.T) {
	// A failed alternative must not leave state behind: /float64 does
	// not unify with the first alternative (/number), and the second
	// unification must still see a clean state.
	subst, err := unify(t,
		[]ast.BaseTerm{ast.Float64Bound, symbols.NewListType(x)},
		[]ast.BaseTerm{numOrFloat, listNum})
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if got, want := len(subst), 1; got != want {
		t.Errorf("subst has %d entries (%v), want %d", got, subst, want)
	}
	if got, want := subst[x], ast.NumberBound; !got.Equals(want) {
		t.Errorf("subst[%v] = %v, want %v", x, got, want)
	}

	// If no alternative matches, the state must be restored as well.
	subst, err = unify(t,
		[]ast.BaseTerm{ast.DurationBound, symbols.NewListType(x)},
		[]ast.BaseTerm{numOrFloat, listNum})
	if err == nil {
		t.Fatal("unifying /duration with union(/number,/float64) succeeded, want error")
	}
	if subst != nil {
		t.Errorf("subst = %v on error, want nil", subst)
	}
}

func TestUnifyTypeExprEmptyUnion(t *testing.T) {
	empty := symbols.NewUnionType()
	if _, err := unify(t, []ast.BaseTerm{ast.NumberBound}, []ast.BaseTerm{empty}); err == nil {
		t.Error("unifying with an empty union succeeded, want error")
	}
}
