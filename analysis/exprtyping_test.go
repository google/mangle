// Copyright 2024 Google LLC
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

package analysis

// Tests documenting expression typing limitations.
//
// Each test is labeled with its category:
//   [FALSE NEGATIVE] bounds check passes but shouldn't (unsound)
//   [FALSE POSITIVE] bounds check fails but shouldn't (incomplete)
//   [WRONG TYPE]     inferred type is incorrect
//
// When a limitation is fixed, the corresponding test should be
// updated to assert the correct behavior.

import (
	"testing"

	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/symbols"
)

// --- Transform arithmetic: typeOfFn does not validate argument types ---
//
// typeOfFn returns a fixed result type (e.g. NumberBound for fn:plus)
// without checking whether arguments actually have compatible types.
// This lets type errors through when arithmetic is used in transforms.

func TestExprTyping_TransformPlusOnString(t *testing.T) {
	// [FALSE NEGATIVE] fn:plus applied to a string variable in a transform.
	// typeOfFn returns NumberBound without checking that X is actually a number.
	test := newBoundsTestCase(t, []ast.Clause{
		clause("result(Y) :- source(X) |> let Y = fn:plus(X, 1)."),
	}, []ast.Decl{
		makeSimpleDecl(atom("result(Y)"), ast.NumberBound),
		makeSimpleDecl(atom("source(X)"), ast.StringBound),
	})
	bc, err := newBoundsAnalyzer(&test.programInfo, symbols.NewNameTrie(), nil, test.rulesMap)
	if err != nil {
		t.Fatal(err)
	}
	err = bc.BoundsCheck()
	if err == nil {
		t.Error("BoundsCheck should reject fn:plus applied to /string argument, but it passes")
	}
}

func TestExprTyping_TransformMinusOnString(t *testing.T) {
	// [FALSE NEGATIVE] fn:minus applied to a string variable in a transform.
	// Same issue as fn:plus: typeOfFn doesn't validate argument types.
	test := newBoundsTestCase(t, []ast.Clause{
		clause("result(Y) :- source(X) |> let Y = fn:minus(X, 1)."),
	}, []ast.Decl{
		makeSimpleDecl(atom("result(Y)"), ast.NumberBound),
		makeSimpleDecl(atom("source(X)"), ast.StringBound),
	})
	bc, err := newBoundsAnalyzer(&test.programInfo, symbols.NewNameTrie(), nil, test.rulesMap)
	if err != nil {
		t.Fatal(err)
	}
	err = bc.BoundsCheck()
	if err == nil {
		t.Error("BoundsCheck should reject fn:minus applied to /string argument, but it passes")
	}
}

func TestExprTyping_TransformMultOnName(t *testing.T) {
	// [FALSE NEGATIVE] fn:mult applied to a name variable in a transform.
	test := newBoundsTestCase(t, []ast.Clause{
		clause("result(Y) :- source(X) |> let Y = fn:mult(X, 2)."),
	}, []ast.Decl{
		makeSimpleDecl(atom("result(Y)"), ast.NumberBound),
		makeSimpleDecl(atom("source(X)"), ast.NameBound),
	})
	bc, err := newBoundsAnalyzer(&test.programInfo, symbols.NewNameTrie(), nil, test.rulesMap)
	if err != nil {
		t.Fatal(err)
	}
	err = bc.BoundsCheck()
	if err == nil {
		t.Error("BoundsCheck should reject fn:mult applied to /name argument, but it passes")
	}
}

// --- Reducer argument types not checked in transforms ---

func TestExprTyping_TransformSumOnStrings(t *testing.T) {
	// [FALSE NEGATIVE] fn:sum applied to string values.
	// typeOfFn returns NumberBound for fn:sum regardless of argument types.
	// At runtime, fn:sum("hello") would fail.
	test := newBoundsTestCase(t, []ast.Clause{
		clause("result(X, Y) :- source(X, Z) |> do fn:group_by(X), let Y = fn:sum(Z)."),
	}, []ast.Decl{
		makeSimpleDecl(atom("result(X, Y)"), ast.StringBound, ast.NumberBound),
		makeSimpleDecl(atom("source(X, Z)"), ast.StringBound, ast.StringBound),
	})
	bc, err := newBoundsAnalyzer(&test.programInfo, symbols.NewNameTrie(), nil, test.rulesMap)
	if err != nil {
		t.Fatal(err)
	}
	err = bc.BoundsCheck()
	if err == nil {
		t.Error("BoundsCheck should reject fn:sum applied to /string values, but it passes")
	}
}

func TestExprTyping_TransformMaxOnStrings(t *testing.T) {
	// [FALSE NEGATIVE] fn:max applied to string values.
	// typeOfFn returns NumberBound for fn:max without checking argument types.
	test := newBoundsTestCase(t, []ast.Clause{
		clause("result(X, Y) :- source(X, Z) |> do fn:group_by(X), let Y = fn:max(Z)."),
	}, []ast.Decl{
		makeSimpleDecl(atom("result(X, Y)"), ast.StringBound, ast.NumberBound),
		makeSimpleDecl(atom("source(X, Z)"), ast.StringBound, ast.StringBound),
	})
	bc, err := newBoundsAnalyzer(&test.programInfo, symbols.NewNameTrie(), nil, test.rulesMap)
	if err != nil {
		t.Fatal(err)
	}
	err = bc.BoundsCheck()
	if err == nil {
		t.Error("BoundsCheck should reject fn:max applied to /string values, but it passes")
	}
}

// --- Float arithmetic: typeOfFn returns wrong type or AnyBound ---

func TestExprTyping_TransformFloatDivReturnsNumberBound(t *testing.T) {
	// [WRONG TYPE] fn:float:div returns Float64Bound in boundOfArg but
	// NumberBound in typeOfFn. When used in a transform with a /float64
	// declaration, this causes a false positive.
	test := newBoundsTestCase(t, []ast.Clause{
		clause("result(Y) :- source(X, Z) |> let Y = fn:float:div(X, Z)."),
	}, []ast.Decl{
		makeSimpleDecl(atom("result(Y)"), ast.Float64Bound),
		makeSimpleDecl(atom("source(X, Z)"), ast.Float64Bound, ast.Float64Bound),
	})
	bc, err := newBoundsAnalyzer(&test.programInfo, symbols.NewNameTrie(), nil, test.rulesMap)
	if err != nil {
		t.Fatal(err)
	}
	err = bc.BoundsCheck()
	if err != nil {
		t.Errorf("BoundsCheck should accept fn:float:div -> /float64, but got: %v", err)
	}
}

func TestExprTyping_TransformFloatPlusReturnsAnyBound(t *testing.T) {
	// [FALSE NEGATIVE] fn:float:plus has emptyType in builtin.Functions,
	// so typeOfFn falls through to checkFunApply which fails, returning AnyBound.
	// AnyBound conforms to any declared type, so type errors go undetected.
	//
	// Here we declare result as /string, but fn:float:plus should produce /float64.
	// Bounds checking passes because typeOfFn returns AnyBound.
	test := newBoundsTestCase(t, []ast.Clause{
		clause("result(Y) :- source(X, Z) |> let Y = fn:float:plus(X, Z)."),
	}, []ast.Decl{
		makeSimpleDecl(atom("result(Y)"), ast.StringBound), // intentionally wrong
		makeSimpleDecl(atom("source(X, Z)"), ast.Float64Bound, ast.Float64Bound),
	})
	bc, err := newBoundsAnalyzer(&test.programInfo, symbols.NewNameTrie(), nil, test.rulesMap)
	if err != nil {
		t.Fatal(err)
	}
	err = bc.BoundsCheck()
	if err == nil {
		t.Error("BoundsCheck should reject fn:float:plus result declared as /string, but it passes (typeOfFn returns AnyBound)")
	}
}

func TestExprTyping_TransformFloatMultReturnsAnyBound(t *testing.T) {
	// [FALSE NEGATIVE] Same issue as fn:float:plus: emptyType → AnyBound.
	test := newBoundsTestCase(t, []ast.Clause{
		clause("result(Y) :- source(X, Z) |> let Y = fn:float:mult(X, Z)."),
	}, []ast.Decl{
		makeSimpleDecl(atom("result(Y)"), ast.StringBound), // intentionally wrong
		makeSimpleDecl(atom("source(X, Z)"), ast.Float64Bound, ast.Float64Bound),
	})
	bc, err := newBoundsAnalyzer(&test.programInfo, symbols.NewNameTrie(), nil, test.rulesMap)
	if err != nil {
		t.Fatal(err)
	}
	err = bc.BoundsCheck()
	if err == nil {
		t.Error("BoundsCheck should reject fn:float:mult result declared as /string, but it passes (typeOfFn returns AnyBound)")
	}
}

// --- Float arithmetic in rule body: boundOfArg returns EmptyType ---

func TestExprTyping_FloatPlusInBody(t *testing.T) {
	// [FALSE POSITIVE] fn:float:plus in rule body (not transform).
	// boundOfArg calls checkFunApply with emptyType → returns EmptyType.
	// addOrRefine rejects EmptyType, so valid float arithmetic fails.
	test := newBoundsTestCase(t, []ast.Clause{
		clause("result(Y) :- source(X, Z), Y = fn:float:plus(X, Z)."),
	}, []ast.Decl{
		makeSimpleDecl(atom("result(Y)"), ast.Float64Bound),
		makeSimpleDecl(atom("source(X, Z)"), ast.Float64Bound, ast.Float64Bound),
	})
	bc, err := newBoundsAnalyzer(&test.programInfo, symbols.NewNameTrie(), nil, test.rulesMap)
	if err != nil {
		t.Fatal(err)
	}
	err = bc.BoundsCheck()
	if err != nil {
		t.Errorf("BoundsCheck should accept fn:float:plus on /float64 args, but got: %v", err)
	}
}

func TestExprTyping_FloatMultInBody(t *testing.T) {
	// [FALSE POSITIVE] fn:float:mult in rule body, same issue.
	test := newBoundsTestCase(t, []ast.Clause{
		clause("result(Y) :- source(X, Z), Y = fn:float:mult(X, Z)."),
	}, []ast.Decl{
		makeSimpleDecl(atom("result(Y)"), ast.Float64Bound),
		makeSimpleDecl(atom("source(X, Z)"), ast.Float64Bound, ast.Float64Bound),
	})
	bc, err := newBoundsAnalyzer(&test.programInfo, symbols.NewNameTrie(), nil, test.rulesMap)
	if err != nil {
		t.Fatal(err)
	}
	err = bc.BoundsCheck()
	if err != nil {
		t.Errorf("BoundsCheck should accept fn:float:mult on /float64 args, but got: %v", err)
	}
}

// --- Sqrt has emptyType ---

func TestExprTyping_SqrtInBody(t *testing.T) {
	// [FALSE POSITIVE] fn:sqrt has emptyType, so checkFunApply fails.
	// Valid usage of fn:sqrt on a number is rejected.
	test := newBoundsTestCase(t, []ast.Clause{
		clause("result(Y) :- source(X), Y = fn:sqrt(X)."),
	}, []ast.Decl{
		makeSimpleDecl(atom("result(Y)"), ast.Float64Bound),
		makeSimpleDecl(atom("source(X)"), ast.Float64Bound),
	})
	bc, err := newBoundsAnalyzer(&test.programInfo, symbols.NewNameTrie(), nil, test.rulesMap)
	if err != nil {
		t.Fatal(err)
	}
	err = bc.BoundsCheck()
	if err != nil {
		t.Errorf("BoundsCheck should accept fn:sqrt on /float64 arg, but got: %v", err)
	}
}

// --- boundOfArg vs typeOfFn inconsistency ---
//
// boundOfArg handles fn:plus etc. by checking argument types and returning
// EmptyType on mismatch. typeOfFn handles the same functions by returning
// a fixed result type without checking arguments. This means the same
// expression is typed differently depending on whether it appears in a
// rule body (boundOfArg) or a transform (typeOfFn).

func TestExprTyping_PlusOnStringInBody(t *testing.T) {
	// [REFERENCE] fn:plus on string in rule body IS rejected by boundOfArg.
	// This is correct behavior — contrast with TestExprTyping_TransformPlusOnString.
	test := newBoundsTestCase(t, []ast.Clause{
		clause("result(Y) :- source(X), Y = fn:plus(X, 1)."),
	}, []ast.Decl{
		makeSimpleDecl(atom("result(Y)"), ast.NumberBound),
		makeSimpleDecl(atom("source(X)"), ast.StringBound),
	})
	bc, err := newBoundsAnalyzer(&test.programInfo, symbols.NewNameTrie(), nil, test.rulesMap)
	if err != nil {
		t.Fatal(err)
	}
	err = bc.BoundsCheck()
	if err == nil {
		t.Error("BoundsCheck should reject fn:plus on /string arg in body")
	}
}

// --- Type context not threaded through inferState ---
//
// inferRelTypesFromPremise creates a fresh typeCtx for each premise,
// so type constraints from polymorphic predicates don't accumulate.
// This means type refinements from one premise don't flow to the next.

func TestExprTyping_TypeCtxNotThreaded(t *testing.T) {
	// [FALSE NEGATIVE + CRASH] The typeCtx is not threaded through inferState.
	// inferState.addOrRefine calls LowerBound(nil, ...) which can crash
	// when polymorphic type variables are involved.
	//
	// When a polymorphic predicate :match_pair(Pair, First, Second) refines
	// type variables, those refinements should constrain subsequent premises.
	// Here the pair is Pair(/string, /number), so X should be /string.
	// A subsequent X < 10 should fail because X is /string, not /number.
	//
	// Currently the nil typeCtx at infercontext.go:68 causes a panic in
	// symbols.TypeConforms when it encounters polymorphic type variables.
	test := newBoundsTestCase(t, []ast.Clause{
		clause("result(X) :- source(P), :match_pair(P, X, Y), X < 10."),
	}, []ast.Decl{
		makeSimpleDecl(atom("result(X)"), ast.NumberBound),
		makeSimpleDecl(atom("source(P)"), symbols.NewPairType(ast.StringBound, ast.NumberBound)),
	})
	bc, err := newBoundsAnalyzer(&test.programInfo, symbols.NewNameTrie(), nil, test.rulesMap)
	if err != nil {
		t.Fatal(err)
	}

	// With the nil typeCtx fix, BoundsCheck no longer panics.
	// The type mismatch may or may not be detected depending on
	// how :match_pair resolves types without full typeCtx threading.
	_ = bc.BoundsCheck()
}
