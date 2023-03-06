package functional

import (
	"fmt"
	"math"
	"testing"

	"github.com/google/mangle/ast"
	"github.com/google/mangle/parse"
	"github.com/google/mangle/symbols"
)

func name(n string) ast.Constant {
	c, err := ast.Name(n)
	if err != nil {
		panic(err)
	}
	return c
}

func TestListContains(t *testing.T) {
	tests := []struct {
		listTerm   ast.BaseTerm
		memberTerm ast.BaseTerm
		want       ast.Constant
		subst      ast.Subst
	}{
		{ast.List([]ast.Constant{ast.Number(10), ast.Number(11), ast.Number(2)}), ast.Number(2), ast.TrueConstant, ast.ConstSubstMap{}},
		{ast.ListNil, ast.Number(2), ast.FalseConstant, ast.ConstSubstMap{}},
		{ast.List([]ast.Constant{ast.Number(10)}), ast.Number(2), ast.FalseConstant, ast.ConstSubstMap{}},
		{
			ast.List([]ast.Constant{ast.Number(10)}), ast.Variable{"X"}, ast.FalseConstant,
			ast.ConstSubstMap{ast.Variable{"X"}: ast.Number(2)},
		},
		{
			ast.List([]ast.Constant{ast.Number(2)}), ast.Variable{"X"}, ast.TrueConstant,
			ast.ConstSubstMap{ast.Variable{"X"}: ast.Number(2)},
		},
		{
			ast.Variable{"Y"},
			ast.Variable{"X"},
			ast.TrueConstant,
			ast.ConstSubstMap{
				ast.Variable{"X"}: ast.Number(2),
				ast.Variable{"Y"}: ast.List([]ast.Constant{ast.Number(2)}),
			},
		},
	}
	for _, test := range tests {
		term := ast.ApplyFn{symbols.ListContains, []ast.BaseTerm{test.listTerm, test.memberTerm}}
		got, err := EvalExpr(term, test.subst)
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("TestListContains(%v, %v)=%v want %v.", term, test.subst, got, test.want)
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
			got, err := EvalApplyFn(test.expr, ast.ConstSubstMap{})
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
			if res, err := EvalApplyFn(test.expr, ast.ConstSubstMap{}); err == nil { // if no error
				t.Errorf("EvalApplyFn(%v)=%v, want error", test.expr, res)
			}
		})
	}
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

func TestNumberToString(t *testing.T) {
	tests := []struct {
		input []ast.BaseTerm
		want  ast.Constant
	}{
		{[]ast.BaseTerm{ast.Number(123)}, ast.String("123")},
		{[]ast.BaseTerm{ast.Number(-42)}, ast.String("-42")},
	}

	for _, test := range tests {
		term := ast.ApplyFn{symbols.NumberToString, test.input}
		got, err := EvalExpr(term, ast.ConstSubstMap{})
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("EvalExpr(%v)=%v want %v.", term, got, test.want)
		}
	}
}

func TestNumberToStringFailure(t *testing.T) {
	tests := [][]ast.BaseTerm{
		[]ast.BaseTerm{ast.Float64(3.14)},
		[]ast.BaseTerm{ast.String("abc")},
	}
	for _, test := range tests {
		term := ast.ApplyFn{symbols.NumberToString, test}
		got, err := EvalExpr(term, ast.ConstSubstMap{})
		if err == nil {
			t.Errorf("EvalExpr(%v)=%v want error.", term, got)
		}
	}
}

func TestFloat64ToString(t *testing.T) {
	tests := []struct {
		input []ast.BaseTerm
		want  ast.Constant
	}{
		{[]ast.BaseTerm{ast.Float64(123)}, ast.String("123")},
		{[]ast.BaseTerm{ast.Float64(-3.14)}, ast.String("-3.14")},
		{[]ast.BaseTerm{ast.Float64(1000000)}, ast.String("1000000")},
		{[]ast.BaseTerm{ast.Float64(1000001)}, ast.String("1000001")},
		{[]ast.BaseTerm{ast.Float64(0.123456789)}, ast.String("0.123456789")},
		{[]ast.BaseTerm{ast.Float64(0.999999999)}, ast.String("0.999999999")},
	}

	for _, test := range tests {
		term := ast.ApplyFn{symbols.Float64ToString, test.input}
		got, err := EvalExpr(term, ast.ConstSubstMap{})
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("EvalExpr(%v)=%v want %v.", term, got, test.want)
		}
	}
}

func TestFloat64ToStringFailure(t *testing.T) {
	tests := [][]ast.BaseTerm{
		[]ast.BaseTerm{ast.Number(42)},
		[]ast.BaseTerm{ast.String("abc")},
	}
	for _, test := range tests {
		term := ast.ApplyFn{symbols.Float64ToString, test}
		got, err := EvalExpr(term, ast.ConstSubstMap{})
		if err == nil {
			t.Errorf("EvalExpr(%v)=%v want error.", term, got)
		}
	}
}

func TestNameToString(t *testing.T) {
	name, err := ast.Name("/named/constant")
	if err != nil {
		t.Fatal("failed to create name constant: ", err)
	}
	term := ast.ApplyFn{symbols.NameToString, []ast.BaseTerm{name}}
	got, err := EvalExpr(term, ast.ConstSubstMap{})
	if err != nil {
		t.Fatal(err)
	}
	want := ast.String("/named/constant")
	if got != want {
		t.Errorf("EvalExpr(%v)=%v want %v.", term, got, want)
	}
}

func TestNameToStringFailure(t *testing.T) {
	tests := [][]ast.BaseTerm{
		[]ast.BaseTerm{ast.Float64(3.14)},
		[]ast.BaseTerm{ast.String("abc")},
	}
	for _, test := range tests {
		term := ast.ApplyFn{symbols.NameToString, test}
		got, err := EvalExpr(term, ast.ConstSubstMap{})
		if err == nil {
			t.Errorf("EvalExpr(%v)=%v want error.", term, got)
		}
	}
}

func TestStringConcatenate(t *testing.T) {
	tests := []struct {
		input []ast.BaseTerm
		want  ast.Constant
	}{
		{[]ast.BaseTerm{}, ast.String("")},
		{[]ast.BaseTerm{ast.String("abc")}, ast.String("abc")},
		{[]ast.BaseTerm{ast.String("abc"), ast.String("123")}, ast.String("abc123")},
		{[]ast.BaseTerm{ast.Float64(3.14)}, ast.String("3.14")},
		{[]ast.BaseTerm{ast.Number(42)}, ast.String("42")},
	}
	for _, test := range tests {
		term := ast.ApplyFn{symbols.StringConcatenate, test.input}
		got, err := EvalExpr(term, ast.ConstSubstMap{})
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("EvalExpr(%v)=%v want %v.", term, got, test.want)
		}
	}
}

func TestStringConcatenateForNameConstant(t *testing.T) {
	name, err := ast.Name("/named/constant")
	if err != nil {
		t.Fatal("failed to create name constant: ", err)
	}

	term := ast.ApplyFn{symbols.StringConcatenate, []ast.BaseTerm{name}}
	got, err := EvalExpr(term, ast.ConstSubstMap{})
	if err != nil {
		t.Fatal(err)
	}
	want := ast.String("/named/constant")
	if got != want {
		t.Errorf("EvalExpr(%v)=%v want %v.", term, got, want)
	}
}

func TestStringConcatenateFailure(t *testing.T) {
	tests := [][]ast.BaseTerm{
		[]ast.BaseTerm{ast.ListNil},
		[]ast.BaseTerm{ast.String("abc"), ast.ListNil},
	}
	for _, test := range tests {
		term := ast.ApplyFn{symbols.StringConcatenate, test}
		got, err := EvalExpr(term, ast.ConstSubstMap{})
		if err == nil {
			t.Errorf("EvalExpr(%v)=%v want error.", term, got)
		}
	}
}
