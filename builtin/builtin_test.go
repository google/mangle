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

	"github.com/google/mangle/ast"
	"github.com/google/mangle/parse"
	"github.com/google/mangle/symbols"
	"github.com/google/mangle/unionfind"
)

func name(n string) ast.Constant {
	c, err := ast.Name(n)
	if err != nil {
		panic(err)
	}
	return c
}

func evalExpr(e ast.BaseTerm) ast.Constant {
	c, err := EvalExpr(e, nil)
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
			t.Errorf("abs: for atom %v expected %v got %v.", atom, test.want, got)
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
			t.Errorf("match_entry: for atom %v expected %v got %v.", atom, test.want, got)
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
			t.Errorf("match_field: for atom %v expected %v got %v.", atom, test.want, got)
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
		t.Errorf("TestMatchConsNegative: expected false, nil got %v %v", got, nsubst)
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

func TestReducerMinMaxSum(t *testing.T) {
	tests := []struct {
		rows    [][]ast.Constant
		wantMin ast.Constant
		wantMax ast.Constant
		wantSum ast.Constant
	}{
		{
			rows: [][]ast.Constant{
				{ast.Number(1)},
				{ast.Number(1)},
				{ast.Number(3)},
			},
			wantMin: ast.Number(1),
			wantMax: ast.Number(3),
			wantSum: ast.Number(5),
		},
		{
			rows:    nil,
			wantMin: ast.Number(math.MaxInt64),
			wantMax: ast.Number(math.MinInt64),
			wantSum: ast.Number(0),
		},
	}
	for _, test := range tests {
		var rows []ast.ConstSubstList
		for _, row := range test.rows {
			rows = append(rows, makeConstSubstList([]ast.Variable{ast.Variable{"X"}}, row))
		}
		gotMax, err := EvalReduceFn(ast.ApplyFn{symbols.Max, []ast.BaseTerm{ast.Variable{"X"}}}, rows)
		if err != nil {
			t.Fatalf("EvalReduceFn(Max,%v) failed with %v", rows, err)
		}
		if test.wantMax != gotMax {
			t.Errorf("EvalReduceFn(Max, %v)=%v want %v", rows, gotMax, test.wantMax)
		}
		gotMin, err := EvalReduceFn(ast.ApplyFn{symbols.Min, []ast.BaseTerm{ast.Variable{"X"}}}, rows)
		if err != nil {
			t.Fatalf("EvalReduceFn(Min,%v) failed with %v", rows, err)
		}
		if test.wantMin != gotMin {
			t.Errorf("EvalReduceFn(Min, %v)=%v want %v", rows, gotMin, test.wantMin)
		}
		gotSum, err := EvalReduceFn(ast.ApplyFn{symbols.Sum, []ast.BaseTerm{ast.Variable{"X"}}}, rows)
		if err != nil {
			t.Fatalf("EvalReduceFn(Sum, %v) failed with %v", rows, err)
		}
		if test.wantSum != gotSum {
			t.Errorf("EvalReduceFn(Sum, %v)=%v want %v", rows, gotSum, test.wantSum)
		}
	}
}

func TestReducerMinMaxSumNegative(t *testing.T) {
	tests := []struct {
		rows [][]ast.Constant
	}{
		{
			rows: [][]ast.Constant{
				{ast.Number(1)},
				{ast.Float64(1.0)},
				{ast.Number(3)},
			},
		},
	}
	for _, test := range tests {
		var rows []ast.ConstSubstList
		for _, row := range test.rows {
			rows = append(rows, makeConstSubstList([]ast.Variable{ast.Variable{"X"}}, row))
		}
		if got, err := EvalReduceFn(ast.ApplyFn{symbols.Max, []ast.BaseTerm{ast.Variable{"X"}}}, rows); err == nil {
			// if NO error
			t.Fatalf("EvalReduceFn(Max,%v) = %v want error", rows, got)
		}
		if got, err := EvalReduceFn(ast.ApplyFn{symbols.Min, []ast.BaseTerm{ast.Variable{"X"}}}, rows); err == nil {
			// if NO error
			t.Fatalf("EvalReduceFn(Min,%v) = %v want error", rows, got)
		}
		if got, err := EvalReduceFn(ast.ApplyFn{symbols.Sum, []ast.BaseTerm{ast.Variable{"X"}}}, rows); err == nil {
			// if NO error
			t.Fatalf("EvalReduceFn(Sum,%v) = %v want error", rows, got)
		}
	}
}

func TestReducerFloatMinMaxSum(t *testing.T) {
	tests := []struct {
		rows    [][]ast.Constant
		wantMin ast.Constant
		wantMax ast.Constant
		wantSum ast.Constant
	}{
		{
			rows: [][]ast.Constant{
				{ast.Float64(1.0)},
				{ast.Float64(1.1)},
				{ast.Float64(3.0)},
			},
			wantMin: ast.Float64(1.0),
			wantMax: ast.Float64(3.0),
			wantSum: ast.Float64(5.1),
		},
		{
			rows:    nil,
			wantMin: ast.Float64(math.MaxFloat64),
			wantMax: ast.Float64(-1 * math.MaxFloat64),
			wantSum: ast.Float64(0.0),
		},
	}
	for _, test := range tests {
		var rows []ast.ConstSubstList
		for _, row := range test.rows {
			rows = append(rows, makeConstSubstList([]ast.Variable{ast.Variable{"X"}}, row))
		}
		gotMax, err := EvalReduceFn(ast.ApplyFn{symbols.FloatMax, []ast.BaseTerm{ast.Variable{"X"}}}, rows)
		if err != nil {
			t.Fatalf("EvalReduceFn(FloatMax,%v) failed with %v", rows, err)
		}
		if test.wantMax != gotMax {
			t.Errorf("EvalReduceFn(FloatMax, %v)=%v want %v", rows, gotMax, test.wantMax)
		}
		gotMin, err := EvalReduceFn(ast.ApplyFn{symbols.FloatMin, []ast.BaseTerm{ast.Variable{"X"}}}, rows)
		if err != nil {
			t.Fatalf("EvalReduceFn(FloatMin,%v) failed with %v", rows, err)
		}
		if test.wantMin != gotMin {
			t.Errorf("EvalReduceFn(FloatMin, %v)=%v want %v", rows, gotMin, test.wantMin)
		}
		gotSum, err := EvalReduceFn(ast.ApplyFn{symbols.FloatSum, []ast.BaseTerm{ast.Variable{"X"}}}, rows)
		if err != nil {
			t.Fatalf("EvalReduceFn(FloatSum, %v) failed with %v", rows, err)
		}
		if test.wantSum != gotSum {
			t.Errorf("EvalReduceFn(FloatSum, %v)=%v want %v", rows, gotSum, test.wantSum)
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
			name: "min of number list",
			expr: ast.ApplyFn{symbols.Min, []ast.BaseTerm{
				ast.List([]ast.Constant{ast.Number(2), ast.Number(5), ast.Number(32)})}},
			want: ast.Number(2),
		},
		{
			name: "sum of number list",
			expr: ast.ApplyFn{symbols.Sum, []ast.BaseTerm{
				ast.List([]ast.Constant{ast.Number(2), ast.Number(5), ast.Number(32)})}},
			want: ast.Number(39),
		},
		{
			name: "floatmax of float64 list",
			expr: ast.ApplyFn{symbols.FloatMax, []ast.BaseTerm{
				ast.List([]ast.Constant{ast.Float64(2.0), ast.Float64(5.0), ast.Float64(32.0)})}},
			want: ast.Float64(32.0),
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
		{
			name: "construct a map",
			expr: ast.ApplyFn{symbols.Map, []ast.BaseTerm{ast.Number(1), ast.String("v"), ast.Number(2), ast.String("foo")}},
			want: *ast.Map(map[*ast.Constant]*ast.Constant{ptr(ast.Number(1)): ptr(ast.String("v")), ptr(ast.Number(2)): ptr(ast.String("foo"))}),
		},
		{
			name: "lookup map",
			expr: ast.ApplyFn{symbols.MapGet, []ast.BaseTerm{
				ast.ApplyFn{symbols.Map, []ast.BaseTerm{ast.Number(1), ast.String("v"), ast.Number(2), ast.String("foo")}},
				ast.Number(1)}},
			want: ast.String("v"),
		},
		{
			name: "lookup map 2",
			expr: ast.ApplyFn{symbols.MapGet, []ast.BaseTerm{
				ast.ApplyFn{symbols.Map, []ast.BaseTerm{ast.Number(1), ast.String("v"), ast.Number(2), ast.String("foo")}},
				ast.Number(2)}},
			want: ast.String("foo"),
		},
		{
			name: "construct a struct",
			expr: ast.ApplyFn{symbols.Struct, []ast.BaseTerm{name("/field1"), ast.String("value"), name("/field2"), ast.Number(32)}},
			want: *ast.Struct(map[*ast.Constant]*ast.Constant{ptr(name("/field1")): ptr(ast.String("value")), ptr(name("/field2")): ptr(ast.Number(32))}),
		},
		{
			name: "field access",
			expr: ast.ApplyFn{symbols.StructGet, []ast.BaseTerm{
				ast.ApplyFn{symbols.Struct, []ast.BaseTerm{name("/field1"), ast.String("value"), name("/field2"), ast.Number(32)}},
				name("/field1")}},
			want: ast.String("value"),
		},
		{
			name: "field access 2",
			expr: ast.ApplyFn{symbols.StructGet, []ast.BaseTerm{
				ast.ApplyFn{symbols.Struct, []ast.BaseTerm{name("/field1"), ast.String("value"), name("/field2"), ast.Number(32)}},
				name("/field2")}},
			want: ast.Number(32),
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
		{
			name: "lookup map not found",
			expr: ast.ApplyFn{symbols.MapGet, []ast.BaseTerm{
				ast.ApplyFn{symbols.Map, []ast.BaseTerm{ast.Number(1), ast.String("v"), ast.Number(2), ast.String("foo")}},
				ast.Number(3)}},
		},
		{
			name: "lookup struct not found",
			expr: ast.ApplyFn{symbols.StructGet, []ast.BaseTerm{
				ast.ApplyFn{symbols.Struct, []ast.BaseTerm{name("/field1"), ast.String("value"), name("/field2"), ast.Number(32)}},
				name("/field3")}},
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

func TestCheckTypeExpressionStructured(t *testing.T) {
	tests := []struct {
		tpe  ast.BaseTerm
		good []ast.Constant
		bad  []ast.Constant
	}{
		{
			tpe: ast.ApplyFn{symbols.MapType, []ast.BaseTerm{
				ast.NumberBound,
				ast.StringBound,
			}},
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
			tpe: ast.ApplyFn{symbols.StructType, []ast.BaseTerm{
				name("/foo"),
				ast.NumberBound,
				name("/bar"),
				ast.StringBound,
				name("/baz"),
				ast.ApplyFn{symbols.ListType, []ast.BaseTerm{ast.NumberBound}},
			}},
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
		h, err := symbols.NewTypeHandle(test.tpe)
		if err != nil {
			t.Errorf("NewTypeHandle(%v) failed %v", test.tpe, err)
		}
		for _, c := range test.good {
			if !h.HasType(c) {
				t.Errorf("NewTypeHandle(%v).HasType(%v)=false want true", test.tpe, c)
			}
		}
		for _, c := range test.bad {
			if h.HasType(c) {
				t.Errorf("NewTypeHandle(%v).HasType(%v)=true want false", test.tpe, c)
			}
		}
	}
}

func evAtom(s string) ast.Atom {
	term, err := parse.Term(s)
	if err != nil {
		panic(err)
	}
	eval, err := EvalAtom(term.(ast.Atom), nil)
	if err != nil {
		panic(err)
	}
	return eval
}

func TestRoundTrip(t *testing.T) {
	tests := []ast.Atom{
		evAtom("bar(/abc, 1, 'def')"),
		evAtom("bar([/abc],1,/def)"),
		evAtom("bar([/abc, /def], 1, /def)"),
		evAtom("baz([/abc : 1,  /def : 2], 1, /def)"),
		evAtom("baz({/abc : 1,  /def : 2}, 1, /def)"),
	}
	for _, test := range tests {
		atom, err := parse.Atom(test.String())
		if err != nil {
			t.Fatal(err)
		}
		atom, err = EvalAtom(atom, nil)
		if err != nil {
			t.Fatal(err)
		}

		if !atom.Equals(test) {
			t.Errorf("(%v).Equals(%v) = false expected true", atom, test)

		}
	}
}
