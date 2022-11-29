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
	"testing"

	"github.com/google/mangle/ast"
	"github.com/google/mangle/parse"
	"github.com/google/mangle/symbols"
	"github.com/google/mangle/unionfind"
)

func extend(u unionfind.UnionFind, left ast.BaseTerm, right ast.BaseTerm) unionfind.UnionFind {
	subst, err := unionfind.UnifyTermsExtend([]ast.BaseTerm{left}, []ast.BaseTerm{right}, u)
	if err != nil {
		panic(fmt.Errorf("test data is invalid: %v %v", left, right))
	}
	return subst
}

func TestLessThan(t *testing.T) {
	tests := []struct {
		left  ast.BaseTerm
		right ast.BaseTerm
		want  bool
		subst unionfind.UnionFind
	}{
		{ast.Number(1), ast.Number(2), true, unionfind.New()},
		{ast.Number(1), ast.Number(1), false, unionfind.New()},
		{ast.Number(2), ast.Number(1), false, unionfind.New()},
	}
	for _, test := range tests {
		atom := ast.NewAtom(":lt", test.left, test.right)
		got, _, err := Decide(atom, &test.subst)
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("abs: for atom %v expected %v got %v.", atom, test.want, got)
		}
	}
}

var emptySubst = unionfind.New()

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
		subst unionfind.UnionFind
	}{
		{ast.Number(1), ast.Number(2), true, unionfind.New()},
		{ast.Number(1), ast.Number(1), true, unionfind.New()},
		{ast.Number(2), ast.Number(1), false, unionfind.New()},
	}
	for _, test := range tests {
		atom := ast.NewAtom(":le", test.left, test.right)
		got, _, err := Decide(atom, &test.subst)
		if err != nil {
			t.Fatal(err)
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

func TestWithinDistance(t *testing.T) {
	tests := []struct {
		left     ast.BaseTerm
		right    ast.BaseTerm
		distance ast.BaseTerm
		want     bool
		subst    unionfind.UnionFind
	}{
		{ast.Number(10), ast.Number(11), ast.Number(2), true, unionfind.New()},
		{ast.Number(10), ast.Number(12), ast.Number(2), false, unionfind.New()},
		{ast.Number(10), ast.Number(9), ast.Number(2), true, unionfind.New()},
	}
	for _, test := range tests {
		atom := ast.NewAtom(":within_distance", test.left, test.right, test.distance)
		got, _, err := Decide(atom, &test.subst)
		if err != nil {
			t.Fatal(err)
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

func TestAbs(t *testing.T) {
	if abs(-1) != 1 || abs(1) != 1 {
		t.Error("abs: unexpected result.")
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
			t.Errorf("abs: for atom %v expected %v got %v.", atom, test.want, got)
		}
	}
}

func TestMatchCons(t *testing.T) {
	tests := []struct {
		scrutinee ast.BaseTerm
		fstVar    ast.BaseTerm
		sndVar    ast.BaseTerm
		want      bool
		subst     unionfind.UnionFind
	}{
		{ast.List([]ast.Constant{ast.Number(1)}), ast.Variable{"_"}, ast.Variable{"_"}, true, unionfind.New()},
		{ast.List([]ast.Constant{ast.Number(1)}), ast.Variable{"X"}, ast.Variable{"Y"}, true,
			extend(extend(unionfind.New(), ast.Variable{"X"}, ast.Number(1)), ast.Variable{"Y"}, ast.ListNil)},
	}
	for _, test := range tests {
		atom := ast.NewAtom(":match_cons", test.scrutinee, test.fstVar, test.sndVar)
		got, _, err := Decide(atom, &test.subst)
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("abs: for atom %v expected %v got %v.", atom, test.want, got)
		}
	}
}

func makeVarList(n int) []ast.Variable {
	var vars []ast.Variable
	for i := 0; i < n; i++ {
		varName := fmt.Sprintf("X%d", i)
		vars = append(vars, ast.Variable{varName})
	}
	return vars
}

func makeVarBaseTerms(n int) []ast.BaseTerm {
	var vars []ast.BaseTerm
	for i := 0; i < n; i++ {
		varName := fmt.Sprintf("X%d", i)
		vars = append(vars, ast.Variable{varName})
	}
	return vars
}

func makeConstSubstList(vars []ast.Variable, columns []ast.Constant) ast.ConstSubstList {
	var subst ast.ConstSubstList
	for i, v := range vars {
		subst = subst.Extend(v, columns[i])
	}
	return subst
}

func TestReducerCollect(t *testing.T) {
	tests := []struct {
		rows [][]ast.Constant
		want ast.Constant
	}{
		{
			rows: [][]ast.Constant{
				{ast.Number(1)},
				{ast.Number(1)},
				{ast.Number(3)},
			},
			want: ast.List([]ast.Constant{
				ast.Number(1),
				ast.Number(1),
				ast.Number(3),
			}),
		},
		{
			rows: [][]ast.Constant{
				{ast.Number(1), ast.Number(2)},
				{ast.Number(1), ast.Number(2)},
				{ast.Number(3), ast.Number(4)},
			},
			want: ast.List([]ast.Constant{
				*pair(ast.Number(1), ast.Number(2)),
				*pair(ast.Number(1), ast.Number(2)),
				*pair(ast.Number(3), ast.Number(4)),
			}),
		},
		{
			rows: [][]ast.Constant{
				{ast.Number(1), ast.Number(2), ast.Number(7)},
				{ast.Number(3), ast.Number(4), ast.Number(7)},
			},
			want: ast.List([]ast.Constant{
				*pair(ast.Number(1), *pair(ast.Number(2), ast.Number(7))),
				*pair(ast.Number(3), *pair(ast.Number(4), ast.Number(7))),
			}),
		},
	}
	for _, test := range tests {
		var rows []ast.ConstSubstList
		width := len(test.rows[0])
		for _, row := range test.rows {
			rows = append(rows, makeConstSubstList(makeVarList(width), row))
		}
		expr := ast.ApplyFn{symbols.Collect, makeVarBaseTerms(width)}
		got, err := EvalReduceFn(expr, rows)
		if err != nil {
			t.Fatalf("EvalReduceFn(%v,%v) failed with %v", expr, rows, err)
		}
		if !got.Equals(test.want) {
			t.Errorf("EvalReduceFn(%v,%v)=%v want %v", expr, rows, got, test.want)
		}
	}
}

func TestReducerCollectDistinct(t *testing.T) {
	tests := []struct {
		rows [][]ast.Constant
		want ast.Constant
	}{
		{
			rows: [][]ast.Constant{
				{ast.Number(1)},
				{ast.Number(1)},
				{ast.Number(3)},
			},
			want: ast.List([]ast.Constant{
				ast.Number(1),
				ast.Number(3),
			}),
		},
		{
			rows: [][]ast.Constant{
				{ast.Number(1), ast.Number(2)},
				{ast.Number(1), ast.Number(2)},
				{ast.Number(3), ast.Number(4)},
			},
			want: ast.List([]ast.Constant{
				*pair(ast.Number(1), ast.Number(2)),
				*pair(ast.Number(3), ast.Number(4)),
			}),
		},
		{
			rows: [][]ast.Constant{
				{ast.Number(1), ast.Number(2), ast.Number(7)},
				{ast.Number(3), ast.Number(4), ast.Number(7)},
			},
			want: ast.List([]ast.Constant{
				*pair(ast.Number(1), *pair(ast.Number(2), ast.Number(7))),
				*pair(ast.Number(3), *pair(ast.Number(4), ast.Number(7))),
			}),
		},
	}
	for _, test := range tests {
		var rows []ast.ConstSubstList
		width := len(test.rows[0])
		for _, row := range test.rows {
			rows = append(rows, makeConstSubstList(makeVarList(width), row))
		}
		expr := ast.ApplyFn{symbols.CollectDistinct, makeVarBaseTerms(width)}
		got, err := EvalReduceFn(expr, rows)
		if err != nil {
			t.Fatalf("EvalReduceFn(%v,%v) failed with %v", expr, rows, err)
		}
		if !got.Equals(test.want) {
			t.Errorf("EvalReduceFn(%v,%v)=%v want %v", expr, rows, got, test.want)
		}
	}
}

func pair(a, b ast.Constant) *ast.Constant {
	p := ast.Pair(&a, &b)
	return &p
}

func ptr(c ast.Constant) *ast.Constant {
	return &c
}

func TestEvalApplyFn(t *testing.T) {
	tests := []struct {
		name string
		expr ast.ApplyFn
		want ast.Constant
	}{
		{
			name: "construct a pair",
			expr: ast.ApplyFn{symbols.Pair, []ast.BaseTerm{ast.String("hello"), ast.Number(2)}},
			want: *pair(ast.String("hello"), ast.Number(2)),
		},
		{
			name: "construct a tuple, case pair",
			expr: ast.ApplyFn{symbols.Tuple, []ast.BaseTerm{ast.String("hello"), ast.Number(2)}},
			want: *pair(ast.String("hello"), ast.Number(2)),
		},
		{
			name: "construct a tuple, case single-element tuple",
			expr: ast.ApplyFn{symbols.Tuple, []ast.BaseTerm{ast.String("hello")}},
			want: ast.String("hello"),
		},
		{
			name: "construct a tuple, case more than two elements",
			expr: ast.ApplyFn{symbols.Tuple, []ast.BaseTerm{ast.String("hello"), ast.Number(2), ast.Number(32)}},
			want: *pair(ast.String("hello"), *pair(ast.Number(2), ast.Number(32))),
		},
		{
			name: "construct a list",
			expr: ast.ApplyFn{symbols.List, []ast.BaseTerm{ast.String("hello"), ast.Number(2), ast.Number(32)}},
			want: ast.List([]ast.Constant{ast.String("hello"), ast.Number(2), ast.Number(32)}),
		},
		{
			name: "get element of list",
			expr: ast.ApplyFn{symbols.ListGet, []ast.BaseTerm{
				ast.List([]ast.Constant{ast.String("hello"), ast.Number(2), ast.Number(32)}),
				ast.Number(2)}},
			want: ast.Number(32),
		},
		{
			name: "append element to empty list",
			expr: ast.ApplyFn{symbols.Append, []ast.BaseTerm{ast.ListNil, ast.String("hello")}},
			want: ast.List([]ast.Constant{ast.String("hello")}),
		},
		{
			name: "append element to list",
			expr: ast.ApplyFn{symbols.Append, []ast.BaseTerm{
				ast.List([]ast.Constant{ast.Number(2), ast.Number(32)}), ast.String("hello")}},
			want: ast.List([]ast.Constant{ast.Number(2), ast.Number(32), ast.String("hello")}),
		},
		{
			name: "length of empty list",
			expr: ast.ApplyFn{symbols.Len, []ast.BaseTerm{ast.ListNil}},
			want: ast.Number(0),
		},
		{
			name: "length of list",
			expr: ast.ApplyFn{symbols.Len, []ast.BaseTerm{ast.ListCons(ptr(ast.String("hello")), &ast.ListNil)}},
			want: ast.Number(1),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := EvalApplyFn(test.expr, unionfind.New())
			if err != nil {
				t.Fatalf("EvalApplyFn(%v) failed with %v", test.expr, err)
			}
			if !got.Equals(test.want) {
				t.Errorf("EvalApplyFn(%v) = %v want %v", test.expr, got, test.want)
			}
		})
	}
}

func TestEvalApplyFnNegative(t *testing.T) {
	tests := []struct {
		name string
		expr ast.ApplyFn
	}{
		{
			name: "len of non-list",
			expr: ast.ApplyFn{symbols.Len, []ast.BaseTerm{ast.Number(23)}},
		},
		{
			name: "get of non-list, non-struct",
			expr: ast.ApplyFn{symbols.ListGet, []ast.BaseTerm{ast.Number(23), ast.Number(23)}},
		},
		{
			name: "append of non-list",
			expr: ast.ApplyFn{symbols.Append, []ast.BaseTerm{ast.Number(23), ast.Number(23)}},
		},
		{
			name: "out of bounds",
			expr: ast.ApplyFn{symbols.ListGet, []ast.BaseTerm{ast.ListNil, ast.Number(1)}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if res, err := EvalApplyFn(test.expr, unionfind.New()); err == nil { // if no error
				t.Errorf("EvalApplyFn(%v)=%v, want error", test.expr, res)
			}
		})
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
			ast.NewBoundDecl(ast.NumberBound, ast.StringBound, ast.NameBound)}, nil)
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
	okFact := atom("foo(12, 'aaa', /bar)")
	if err := checker.CheckTypeBounds(okFact); err != nil {
		t.Errorf("CheckTypeBounds(%v) failed %v", okFact, err)
	}
	badFact := atom("foo('aaa', 12, /bar)")
	if err := checker.CheckTypeBounds(badFact); err == nil { // if NO error
		t.Errorf("CheckTypeBounds(%v) succeeded expected error", badFact)
	}
}
