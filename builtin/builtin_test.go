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

package builtin

import (
	"fmt"
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/mangle/ast"
	"github.com/google/mangle/functional"
	"github.com/google/mangle/parse"
	"github.com/google/mangle/symbols"
	"github.com/google/mangle/unionfind"
)

var emptySubst = unionfind.New()

func name(n string) ast.Constant {
	c, err := ast.Name(n)
	if err != nil {
		panic(err)
	}
	return c
}

func evalExpr(e ast.BaseTerm) ast.Constant {
	c, err := functional.EvalExpr(e, nil)
	if err != nil {
		panic(err)
	}
	constant, ok := c.(ast.Constant)
	if !ok {
		panic(fmt.Errorf("not a constant %v", c))
	}
	return constant
}

func extend(u unionfind.UnionFind, left ast.BaseTerm, right ast.BaseTerm) unionfind.UnionFind {
	subst, err := unionfind.UnifyTermsExtend([]ast.BaseTerm{left}, []ast.BaseTerm{right}, u)
	if err != nil {
		panic(fmt.Errorf("test data is invalid: %v %v", left, right))
	}
	return subst
}

func TestAbs(t *testing.T) {
	if abs(-1) != 1 || abs(1) != 1 {
		t.Error("abs: unexpected result.")
	}
	if abs(math.MinInt64) != math.MaxInt64 {
		t.Errorf("abs: abs(MinInt64) != MaxInt64")
	}
}

func TestLessThan(t *testing.T) {
	tests := []struct {
		left  ast.BaseTerm
		right ast.BaseTerm
		want  bool
	}{
		{ast.Number(1), ast.Number(2), true},
		{ast.Number(1), ast.Number(1), false},
		{ast.Number(2), ast.Number(1), false},
	}
	for _, test := range tests {
		atom := ast.NewAtom(":lt", test.left, test.right)
		got, nsubst, err := Decide(atom, &emptySubst)
		if err != nil {
			t.Fatal(err)
		}
		if len(nsubst) != 1 || nsubst[0] != &emptySubst {
			t.Errorf("LessThan: expected same subst %v %v %v", atom, nsubst, &emptySubst)
		}
		if got != test.want {
			t.Errorf("LessThan: for atom %v got %v want %v.", atom, got, test.want)
		}
	}
}

func TestLessThanError(t *testing.T) {
	atom := ast.NewAtom(":lt", ast.String("hello"), ast.Number(2))
	if got, _, err := Decide(atom, &emptySubst); err == nil { // if no error
		t.Errorf("Decide(%v) = %v want error", atom, got)
	}

	invalid := ast.NewAtom(":lt", ast.Number(2), ast.Number(2), ast.Number(2))
	if got, _, err := Decide(invalid, &emptySubst); err == nil { // if no error
		t.Errorf("Decide(%v) = %v want error", invalid, got)
	}
}

func TestLessThanOrEqual(t *testing.T) {
	tests := []struct {
		left  ast.BaseTerm
		right ast.BaseTerm
		want  bool
	}{
		{ast.Number(1), ast.Number(2), true},
		{ast.Number(1), ast.Number(1), true},
		{ast.Number(2), ast.Number(1), false},
	}
	for _, test := range tests {
		atom := ast.NewAtom(":le", test.left, test.right)
		got, nsubst, err := Decide(atom, &emptySubst)
		if err != nil {
			t.Fatal(err)
		}
		if len(nsubst) != 1 || nsubst[0] != &emptySubst {
			t.Errorf("LessThanOrEqual: expected same subst %v %v %v", atom, nsubst, &emptySubst)
		}
		if got != test.want {
			t.Errorf("abs: for atom %v expected %v got %v.", atom, test.want, got)
		}
	}
}

func TestLessThanOrEqualError(t *testing.T) {
	atom := ast.NewAtom(":le", ast.String("hello"), ast.Number(2))
	if got, _, err := Decide(atom, &emptySubst); err == nil { // if no error
		t.Errorf("Decide(%v) = %v want error", atom, got)
	}

	invalid := ast.NewAtom(":le", ast.Number(2), ast.Number(2), ast.Number(2))
	if got, _, err := Decide(invalid, &emptySubst); err == nil { // if no error
		t.Errorf("Decide(%v) = %v want error", invalid, got)
	}
}

func TestGreaterThan(t *testing.T) {
	tests := []struct {
		left  ast.BaseTerm
		right ast.BaseTerm
		want  bool
	}{
		{ast.Number(2), ast.Number(1), true},
		{ast.Number(1), ast.Number(1), false},
		{ast.Number(1), ast.Number(2), false},
	}
	for _, test := range tests {
		atom := ast.NewAtom(":gt", test.left, test.right)
		got, nsubst, err := Decide(atom, &emptySubst)
		if err != nil {
			t.Fatal(err)
		}
		if len(nsubst) != 1 || nsubst[0] != &emptySubst {
			t.Errorf("GreaterThan: expected same subst %v %v %v", atom, nsubst, &emptySubst)
		}
		if got != test.want {
			t.Errorf("GreaterThan: for atom %v got %v want %v.", atom, got, test.want)
		}
	}
}

func TestGreaterThanError(t *testing.T) {
	atom := ast.NewAtom(":gt", ast.String("hello"), ast.Number(2))
	if got, _, err := Decide(atom, &emptySubst); err == nil { // if no error
		t.Errorf("Decide(%v) = %v want error", atom, got)
	}

	invalid := ast.NewAtom(":gt", ast.Number(2), ast.Number(2), ast.Number(2))
	if got, _, err := Decide(invalid, &emptySubst); err == nil { // if no error
		t.Errorf("Decide(%v) = %v want error", invalid, got)
	}
}

func TestGreaterThanOrEqual(t *testing.T) {
	tests := []struct {
		left  ast.BaseTerm
		right ast.BaseTerm
		want  bool
	}{
		{ast.Number(2), ast.Number(1), true},
		{ast.Number(1), ast.Number(1), true},
		{ast.Number(1), ast.Number(2), false},
	}
	for _, test := range tests {
		atom := ast.NewAtom(":ge", test.left, test.right)
		got, nsubst, err := Decide(atom, &emptySubst)
		if err != nil {
			t.Fatal(err)
		}
		if len(nsubst) != 1 || nsubst[0] != &emptySubst {
			t.Errorf("GreaterThanOrEqual: expected same subst %v %v %v", atom, nsubst, &emptySubst)
		}
		if got != test.want {
			t.Errorf("abs: for atom %v expected %v got %v.", atom, test.want, got)
		}
	}
}

func TestGreaterThanOrEqualError(t *testing.T) {
	atom := ast.NewAtom(":ge", ast.String("hello"), ast.Number(2))
	if got, _, err := Decide(atom, &emptySubst); err == nil { // if no error
		t.Errorf("Decide(%v) = %v want error", atom, got)
	}

	invalid := ast.NewAtom(":ge", ast.Number(2), ast.Number(2), ast.Number(2))
	if got, _, err := Decide(invalid, &emptySubst); err == nil { // if no error
		t.Errorf("Decide(%v) = %v want error", invalid, got)
	}
}

func TestWithinDistance(t *testing.T) {
	tests := []struct {
		left     ast.BaseTerm
		right    ast.BaseTerm
		distance ast.BaseTerm
		want     bool
	}{
		{ast.Number(10), ast.Number(11), ast.Number(2), true},
		{ast.Number(10), ast.Number(12), ast.Number(2), false},
		{ast.Number(10), ast.Number(9), ast.Number(2), true},
	}
	for _, test := range tests {
		atom := ast.NewAtom(":within_distance", test.left, test.right, test.distance)
		got, nsubst, err := Decide(atom, &emptySubst)
		if err != nil {
			t.Fatal(err)
		}
		if len(nsubst) != 1 || nsubst[0] != &emptySubst {
			t.Errorf("LessThan: expected same subst %v %v %v", atom, nsubst, &emptySubst)
		}
		if got != test.want {
			t.Errorf("abs: for atom %v expected %v got %v.", atom, test.want, got)
		}
	}
}

func TestWithinDistanceError(t *testing.T) {
	atom := ast.NewAtom(":within_distance", ast.String("hello"), ast.Number(2), ast.Number(2))
	if got, _, err := Decide(atom, &emptySubst); err == nil { // if no error
		t.Errorf("Decide(%v)=%v want error", atom, got)
	}

	invalid := ast.NewAtom(":within_distance", ast.Number(2), ast.Number(2))
	if got, _, err := Decide(invalid, &emptySubst); err == nil { // if no error
		t.Errorf("Decide(%v)=%v want error", invalid, got)
	}
}

func TestMatchPrefix(t *testing.T) {
	tests := []struct {
		scrutinee ast.Constant
		pattern   ast.Constant
		want      bool
	}{
		{name("/foo/bar"), name("/foo"), true},
		{name("/foo"), name("/foo"), false},
		{name("/foo/bar"), name("/bar"), false},
		{ast.String("foo/bar"), name("/foo"), false},
	}
	for _, test := range tests {
		atom := ast.NewAtom(":match_prefix", test.scrutinee, test.pattern)
		got, _, err := Decide(atom, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("TestMatchPrefix(%v): got %v want %v", atom, got, test.want)
		}
	}
}

func TestStartsWith(t *testing.T) {
	tests := []struct {
		scrutinee ast.Constant
		pattern   ast.Constant
		want      bool
	}{
		{ast.String("foo/bar"), ast.String("foo"), true},
		{ast.String("foo"), ast.String("foo"), true},
		{ast.String("foo"), ast.String("food"), false},
		{ast.String("foo/bar"), ast.String("bar"), false},
		{ast.String("foo"), ast.String(""), true},
	}
	for _, test := range tests {
		atom := ast.NewAtom(":string:starts_with", test.scrutinee, test.pattern)
		got, _, err := Decide(atom, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("TestStartsWith(%v): got %v want %v", atom, got, test.want)
		}
	}
}

func TestEndsWith(t *testing.T) {
	tests := []struct {
		scrutinee ast.Constant
		pattern   ast.Constant
		want      bool
	}{
		{ast.String("foo/bar"), ast.String("bar"), true},
		{ast.String("foo"), ast.String("foo"), true},
		{ast.String("foo"), ast.String("dfoo"), false},
		{ast.String("foo/bar"), ast.String("foo"), false},
		{ast.String("foo"), ast.String(""), true},
	}
	for _, test := range tests {
		atom := ast.NewAtom(":string:ends_with", test.scrutinee, test.pattern)
		got, _, err := Decide(atom, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("TestEndsWith(%v): got %v want %v", atom, got, test.want)
		}
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		scrutinee ast.Constant
		pattern   ast.Constant
		want      bool
	}{
		{ast.String("foo/bar"), ast.String("bar"), true},
		{ast.String("foo"), ast.String("foo"), true},
		{ast.String("foo"), ast.String("dfoo"), false},
		{ast.String("foo/bar"), ast.String("foo"), true},
		{ast.String("foo/bar"), ast.String("oo"), true},
		{ast.String("foo/bar"), ast.String("ob"), false},
		{ast.String("foo"), ast.String(""), true},
	}
	for _, test := range tests {
		atom := ast.NewAtom(":string:contains", test.scrutinee, test.pattern)
		got, _, err := Decide(atom, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("TestContains(%v): got %v want %v", atom, got, test.want)
		}
	}
}

func TestFilter(t *testing.T) {
	tests := []struct {
		cond ast.BaseTerm
		want bool
	}{
		{cond: ast.TrueConstant, want: true},
		{cond: ast.FalseConstant, want: false},
		{
			cond: ast.ApplyFn{symbols.ListContains, []ast.BaseTerm{ast.ListNil, ast.Number(23)}},
			want: false,
		},
		{
			cond: ast.ApplyFn{symbols.ListContains, []ast.BaseTerm{ast.List([]ast.Constant{ast.Number(23)}), ast.Number(23)}},
			want: true,
		},
	}
	for _, test := range tests {
		atom := ast.NewAtom(":filter", test.cond)
		got, _, err := Decide(atom, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("TestFilter(%v): got %v want %v", atom, got, test.want)
		}
	}
}

func TestMatchPair(t *testing.T) {
	makePair := func(left, right ast.Constant) ast.Constant {
		return ast.Pair(&left, &right)
	}
	tests := []struct {
		scrutinee ast.BaseTerm
		fstVar    ast.BaseTerm
		sndVar    ast.BaseTerm
		want      bool
		subst     unionfind.UnionFind
	}{
		{makePair(ast.Number(1), ast.Number(2)), ast.Variable{"_"}, ast.Variable{"_"}, true, unionfind.New()},
		{makePair(ast.Number(1), ast.Number(2)), ast.Variable{"X"}, ast.Variable{"Y"}, true,
			extend(extend(unionfind.New(), ast.Variable{"X"}, ast.Number(1)), ast.Variable{"Y"}, ast.Number(2))},
	}
	for _, test := range tests {
		atom := ast.NewAtom(":match_pair", test.scrutinee, test.fstVar, test.sndVar)
		got, _, err := Decide(atom, &test.subst)
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("TestMatchPair(%v): got %v want %v", atom, got, test.want)
		}
	}
}

func TestMatchCons(t *testing.T) {
	tests := []struct {
		scrutinee ast.BaseTerm
		fstVar    ast.BaseTerm
		sndVar    ast.BaseTerm
		want      bool
	}{
		{ast.List([]ast.Constant{ast.Number(1)}), ast.Variable{"_"}, ast.Variable{"_"}, true},
		{ast.List([]ast.Constant{ast.Number(1)}), ast.Variable{"X"}, ast.Variable{"Y"}, true},
		{ast.Number(1), ast.Variable{"_"}, ast.Variable{"_"}, false},
		{ast.ListNil, ast.Variable{"_"}, ast.Variable{"_"}, false},
	}
	for _, test := range tests {
		atom := ast.NewAtom(":match_cons", test.scrutinee, test.fstVar, test.sndVar)
		var subst unionfind.UnionFind
		got, _, err := Decide(atom, &subst)
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("TestMatchCons(%v) got %v want %v.", atom, got, test.want)
		}
	}
}

func TestMatchEntry(t *testing.T) {
	tests := []struct {
		scrutinee ast.BaseTerm
		keyPat    ast.BaseTerm
		valPat    ast.BaseTerm
		want      bool
		subst     unionfind.UnionFind
	}{
		{evalExpr(ast.ApplyFn{symbols.Map, []ast.BaseTerm{
			ast.Number(3), ast.String("three"),
			ast.Number(4), ast.String("four"),
		}}), ast.Number(3), ast.Variable{"_"}, true, unionfind.New()},
		{evalExpr(ast.ApplyFn{symbols.Map, []ast.BaseTerm{
			ast.Number(3), ast.String("three"),
			ast.Number(4), ast.String("four"),
		}}), ast.Number(3), ast.Variable{"X"}, true, extend(unionfind.New(), ast.Variable{"X"}, ast.String("three"))},
	}
	for _, test := range tests {
		atom := ast.NewAtom(":match_entry", test.scrutinee, test.keyPat, test.valPat)
		got, _, err := Decide(atom, &test.subst)
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("TestMatchEntry(%v): got %v want %v", atom, got, test.want)
		}
	}
}

func TestMatchField(t *testing.T) {
	tests := []struct {
		scrutinee ast.BaseTerm
		keyPat    ast.BaseTerm
		valPat    ast.BaseTerm
		want      bool
		subst     unionfind.UnionFind
	}{
		{evalExpr(ast.ApplyFn{symbols.Struct, []ast.BaseTerm{
			name("/foo"), ast.Number(3), name("/bar"), ast.String("three"),
		}}), name("/foo"), ast.Variable{"_"}, true, unionfind.New()},
		{evalExpr(ast.ApplyFn{symbols.Struct, []ast.BaseTerm{
			name("/foo"), ast.Number(3), name("/bar"), ast.String("three"),
		}}), name("/bar"), ast.Variable{"X"}, true, extend(unionfind.New(), ast.Variable{"X"}, ast.String("three"))},
	}
	for _, test := range tests {
		atom := ast.NewAtom(":match_field", test.scrutinee, test.keyPat, test.valPat)
		got, _, err := Decide(atom, &test.subst)
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("TestMatchField(%v) got %v want %v.", atom, got, test.want)
		}
	}
}

func TestMatchConsNegative(t *testing.T) {
	var subst unionfind.UnionFind
	atom := ast.NewAtom(":match_cons", ast.ListNil, ast.Variable{"X"}, ast.Variable{"Y"})
	got, nsubst, err := Decide(atom, &subst)
	if err != nil {
		t.Fatal(err)
	}
	if got != false || nsubst != nil {
		t.Errorf("TestMatchConsNegative(%v): got %v, %v want false, nil", atom, got, nsubst)
	}
}

func atom(s string) ast.Atom {
	a, err := parse.Atom(s)
	if err != nil {
		panic(fmt.Errorf("bad syntax in test case: %s got %w", s, err))
	}
	return a
}

func TestTypeCheck(t *testing.T) {
	fooDecl, err := ast.NewDecl(
		ast.NewAtom("foo", ast.Variable{"SomeNum"}, ast.Variable{"SomeStr"}, ast.Variable{"SomeName"}),
		nil,
		[]ast.BoundDecl{
			ast.NewBoundDecl(ast.NumberBound, symbols.NewStructType(name("/bar"), symbols.NewListType(ast.StringBound)), ast.NameBound)}, nil)
	if err != nil {
		t.Fatal(err)
	}
	decls := map[ast.PredicateSym]ast.Decl{
		ast.PredicateSym{"foo", 3}: fooDecl,
	}
	checker, err := NewTypeChecker(decls)
	if err != nil {
		t.Fatalf("bad test setup, cannot construct type checker: %v", err)
	}
	okFact := evAtom("foo(12, {/bar: ['aaa']}, /bar)")
	if err := checker.CheckTypeBounds(okFact); err != nil {
		t.Errorf("CheckTypeBounds(%v) failed %v", okFact, err)
	}
	badFact := evAtom("foo('aaa', {/bar: ['b12']}, /bar)")
	if err := checker.CheckTypeBounds(badFact); err == nil { // if NO error
		t.Errorf("CheckTypeBounds(%v) succeeded expected error", badFact)
	}
}

func TestCheckTypeExpressionStructured(t *testing.T) {
	tests := []struct {
		tpe  ast.BaseTerm
		good []ast.Constant
		bad  []ast.Constant
	}{
		{
			tpe: symbols.NewMapType(
				ast.NumberBound,
				ast.StringBound,
			),
			good: []ast.Constant{
				evalExpr(ast.ApplyFn{symbols.Map, []ast.BaseTerm{
					ast.Number(3), ast.String("three"),
					ast.Number(4), ast.String("four"),
				}}),
			},
			bad: []ast.Constant{
				evalExpr(ast.ApplyFn{symbols.Map, []ast.BaseTerm{
					ast.String("three"), ast.Number(3),
					ast.String("four"), ast.Number(4),
				}}),
			},
		},
		{
			tpe: symbols.NewStructType(
				name("/foo"),
				ast.NumberBound,
				name("/bar"),
				ast.StringBound,
				name("/baz"),
				symbols.NewListType(ast.NumberBound),
			),
			good: []ast.Constant{
				evalExpr(ast.ApplyFn{symbols.Struct, []ast.BaseTerm{
					name("/foo"), ast.Number(3),
					name("/bar"), ast.String("three"),
					name("/baz"), ast.ApplyFn{symbols.List, []ast.BaseTerm{
						ast.Number(23),
					}}}}),
			},
			bad: []ast.Constant{
				evalExpr(ast.ApplyFn{symbols.Struct, []ast.BaseTerm{
					name("/foo"), ast.Number(3),
				}}),
			},
		},
	}
	for _, test := range tests {
		h, err := symbols.NewBoundHandle(test.tpe)
		if err != nil {
			t.Errorf("NewMonoTypeHandle(%v) failed %v", test.tpe, err)
		}
		for _, c := range test.good {
			if !h.HasType(c) {
				t.Errorf("NewMonoTypeHandle(%v).HasType(%v)=false want true", test.tpe, c)
			}
		}
		for _, c := range test.bad {
			if h.HasType(c) {
				t.Errorf("NewMonoTypeHandle(%v).HasType(%v)=true want false", test.tpe, c)
			}
		}
	}
}

func evAtom(s string) ast.Atom {
	term, err := parse.Term(s)
	if err != nil {
		panic(err)
	}
	eval, err := functional.EvalAtom(term.(ast.Atom), nil)
	if err != nil {
		panic(err)
	}
	return eval
}

func TestTimeLessThan(t *testing.T) {
	t1 := ast.Time(1705314600000000000) // Earlier
	t2 := ast.Time(1705314601000000000) // 1 second later

	tests := []struct {
		left  ast.BaseTerm
		right ast.BaseTerm
		want  bool
	}{
		{t1, t2, true},  // earlier < later
		{t2, t1, false}, // later < earlier
		{t1, t1, false}, // same time
	}
	for _, test := range tests {
		atom := ast.NewAtom(":time:lt", test.left, test.right)
		got, nsubst, err := Decide(atom, &emptySubst)
		if err != nil {
			t.Fatal(err)
		}
		if len(nsubst) != 1 || nsubst[0] != &emptySubst {
			t.Errorf("TimeLt: expected same subst %v %v %v", atom, nsubst, &emptySubst)
		}
		if got != test.want {
			t.Errorf("TimeLt: for atom %v got %v want %v.", atom, got, test.want)
		}
	}
}

func TestTimeLessThanOrEqual(t *testing.T) {
	t1 := ast.Time(1705314600000000000)
	t2 := ast.Time(1705314601000000000)

	tests := []struct {
		left  ast.BaseTerm
		right ast.BaseTerm
		want  bool
	}{
		{t1, t2, true},  // earlier <= later
		{t2, t1, false}, // later <= earlier
		{t1, t1, true},  // same time
	}
	for _, test := range tests {
		atom := ast.NewAtom(":time:le", test.left, test.right)
		got, _, err := Decide(atom, &emptySubst)
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("TimeLe: for atom %v got %v want %v.", atom, got, test.want)
		}
	}
}

func TestTimeGreaterThan(t *testing.T) {
	t1 := ast.Time(1705314600000000000)
	t2 := ast.Time(1705314601000000000)

	tests := []struct {
		left  ast.BaseTerm
		right ast.BaseTerm
		want  bool
	}{
		{t2, t1, true},  // later > earlier
		{t1, t2, false}, // earlier > later
		{t1, t1, false}, // same time
	}
	for _, test := range tests {
		atom := ast.NewAtom(":time:gt", test.left, test.right)
		got, _, err := Decide(atom, &emptySubst)
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("TimeGt: for atom %v got %v want %v.", atom, got, test.want)
		}
	}
}

func TestTimeGreaterThanOrEqual(t *testing.T) {
	t1 := ast.Time(1705314600000000000)
	t2 := ast.Time(1705314601000000000)

	tests := []struct {
		left  ast.BaseTerm
		right ast.BaseTerm
		want  bool
	}{
		{t2, t1, true},  // later >= earlier
		{t1, t2, false}, // earlier >= later
		{t1, t1, true},  // same time
	}
	for _, test := range tests {
		atom := ast.NewAtom(":time:ge", test.left, test.right)
		got, _, err := Decide(atom, &emptySubst)
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("TimeGe: for atom %v got %v want %v.", atom, got, test.want)
		}
	}
}

func TestTimeComparisonError(t *testing.T) {
	// Should fail when comparing time with non-time
	atom := ast.NewAtom(":time:lt", ast.Time(1705314600000000000), ast.Number(1705314601000000000))
	if got, _, err := Decide(atom, &emptySubst); err == nil {
		t.Errorf("Decide(%v) = %v want error", atom, got)
	}
}

func TestDurationLessThan(t *testing.T) {
	d1 := ast.Duration(3600000000000) // 1 hour
	d2 := ast.Duration(7200000000000) // 2 hours

	tests := []struct {
		left  ast.BaseTerm
		right ast.BaseTerm
		want  bool
	}{
		{d1, d2, true},  // shorter < longer
		{d2, d1, false}, // longer < shorter
		{d1, d1, false}, // same duration
	}
	for _, test := range tests {
		atom := ast.NewAtom(":duration:lt", test.left, test.right)
		got, _, err := Decide(atom, &emptySubst)
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("DurationLt: for atom %v got %v want %v.", atom, got, test.want)
		}
	}
}

func TestDurationLessThanOrEqual(t *testing.T) {
	d1 := ast.Duration(3600000000000)
	d2 := ast.Duration(7200000000000)

	tests := []struct {
		left  ast.BaseTerm
		right ast.BaseTerm
		want  bool
	}{
		{d1, d2, true},
		{d2, d1, false},
		{d1, d1, true},
	}
	for _, test := range tests {
		atom := ast.NewAtom(":duration:le", test.left, test.right)
		got, _, err := Decide(atom, &emptySubst)
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("DurationLe: for atom %v got %v want %v.", atom, got, test.want)
		}
	}
}

func TestDurationGreaterThan(t *testing.T) {
	d1 := ast.Duration(3600000000000)
	d2 := ast.Duration(7200000000000)

	tests := []struct {
		left  ast.BaseTerm
		right ast.BaseTerm
		want  bool
	}{
		{d2, d1, true},
		{d1, d2, false},
		{d1, d1, false},
	}
	for _, test := range tests {
		atom := ast.NewAtom(":duration:gt", test.left, test.right)
		got, _, err := Decide(atom, &emptySubst)
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("DurationGt: for atom %v got %v want %v.", atom, got, test.want)
		}
	}
}

func TestDurationGreaterThanOrEqual(t *testing.T) {
	d1 := ast.Duration(3600000000000)
	d2 := ast.Duration(7200000000000)

	tests := []struct {
		left  ast.BaseTerm
		right ast.BaseTerm
		want  bool
	}{
		{d2, d1, true},
		{d1, d2, false},
		{d1, d1, true},
	}
	for _, test := range tests {
		atom := ast.NewAtom(":duration:ge", test.left, test.right)
		got, _, err := Decide(atom, &emptySubst)
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("DurationGe: for atom %v got %v want %v.", atom, got, test.want)
		}
	}
}

func TestDurationComparisonError(t *testing.T) {
	// Should fail when comparing duration with non-duration
	atom := ast.NewAtom(":duration:lt", ast.Duration(3600000000000), ast.Number(7200000000000))
	if got, _, err := Decide(atom, &emptySubst); err == nil {
		t.Errorf("Decide(%v) = %v want error", atom, got)
	}
}

func TestExpand(t *testing.T) {
	tests := []struct {
		atom      ast.Atom
		subst     unionfind.UnionFind
		wantBool  bool
		wantSubst []map[ast.Variable]ast.Constant
	}{
		{
			atom:     evAtom(":list:member(X, [1,2,3])"),
			subst:    unionfind.New(),
			wantBool: true,
			wantSubst: []map[ast.Variable]ast.Constant{
				{ast.Variable{"X"}: ast.Number(1)},
				{ast.Variable{"X"}: ast.Number(2)},
				{ast.Variable{"X"}: ast.Number(3)},
			},
		},
		{
			atom:      evAtom(":list:member(2, [1,2,3])"),
			subst:     unionfind.New(),
			wantBool:  true,
			wantSubst: nil,
		},
		{
			atom:      evAtom(":list:member(4, [1,2,3])"),
			subst:     unionfind.New(),
			wantBool:  false,
			wantSubst: nil,
		},
		{
			atom:     evAtom(":list:member(X, [1,2,3])"),
			subst:    extend(unionfind.New(), ast.Variable{"X"}, ast.Number(2)),
			wantBool: true,
			wantSubst: []map[ast.Variable]ast.Constant{
				{ast.Variable{"X"}: ast.Number(2)},
			},
		},
		{
			atom:     evAtom(":list:member(2, [1,2,3])"),
			subst:    extend(unionfind.New(), ast.Variable{"X"}, ast.String("hello")),
			wantBool: true,
			wantSubst: []map[ast.Variable]ast.Constant{
				{ast.Variable{"X"}: ast.String("hello")},
			},
		},
	}
	for _, test := range tests {
		gotBool, gotSubsts, err := Decide(test.atom, &test.subst)
		if err != nil {
			t.Fatalf("Decide(%v,%v) failed with %v", test.atom, test.subst, err)
		}
		if gotBool != test.wantBool {
			t.Fatalf("Decide(%v,%v)=bool %v want %v", test.atom, test.subst, gotBool, test.wantBool)
		}
		if test.wantSubst != nil {
			domain := func(substMap map[ast.Variable]ast.Constant) []ast.Variable {
				var vars []ast.Variable
				for v := range substMap {
					vars = append(vars, v)
				}
				return vars
			}
			substToMap := func(subst *unionfind.UnionFind, substDomain []ast.Variable) map[ast.Variable]ast.Constant {
				substMap := make(map[ast.Variable]ast.Constant)
				for _, v := range substDomain {
					substMap[v] = subst.Get(v).(ast.Constant)
				}
				return substMap
			}
			if len(test.wantSubst) != len(gotSubsts) {
				t.Errorf("Decide(%v,%v)=%v want %v", test.atom, test.subst, gotSubsts, test.wantSubst)
			}
			for _, gotSubst := range gotSubsts {
				var found bool
				for _, wantSubstMap := range test.wantSubst {
					gotSubstMap := substToMap(gotSubst, domain(wantSubstMap))
					if cmp.Equal(gotSubstMap, wantSubstMap, cmp.AllowUnexported(ast.Constant{})) {
						found = true
					}
				}
				if !found {
					t.Errorf("Decide(%v,%v)=...%s... want %v", test.atom, test.subst, gotSubst, test.wantSubst)
				}
			}
		}
	}
}
