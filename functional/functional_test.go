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

func TestEvalFloatPlus(t *testing.T) {
	tests := []struct {
		args []ast.BaseTerm
		want ast.Constant
	}{
		{[]ast.BaseTerm{}, ast.Float64(0)},
		{[]ast.BaseTerm{ast.Float64(1.5)}, ast.Float64(1.5)},
		{[]ast.BaseTerm{ast.Float64(1.5), ast.Float64(2.5)}, ast.Float64(4.0)},
		{[]ast.BaseTerm{ast.Number(2), ast.Float64(3.5)}, ast.Float64(5.5)},
		{[]ast.BaseTerm{ast.Float64(2.5), ast.Number(1)}, ast.Float64(3.5)},
	}
	for _, test := range tests {
		expr := ast.ApplyFn{symbols.FloatPlus, test.args}
		got, err := EvalExpr(expr, ast.ConstSubstMap{})
		if err != nil {
			t.Errorf("EvalExpr(%v) error: %v", expr, err)
			continue
		}
		if !got.Equals(test.want) {
			t.Errorf("EvalExpr(%v) = %v, want %v", expr, got, test.want)
		}
	}
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

func TestReducerCollectCountDistinct(t *testing.T) {
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
		expr = ast.ApplyFn{symbols.CountDistinct, makeVarBaseTerms(width)}
		got, err = EvalReduceFn(expr, rows)
		if err != nil {
			t.Fatalf("EvalReduceFn(%v,%v) count_distinct failed with %v", expr, rows, err)
		}
		iter, _ := test.want.ListSeq()
		var expected int64
		for range iter {
			expected++
		}
		if !got.Equals(ast.Number(expected)) {
			t.Errorf("EvalReduceFn(%v,%v)=%v count_distinct want %d", expr, rows, got, expected)
		}
	}
}

func TestReducerMinMaxSum(t *testing.T) {
	tests := []struct {
		rows    [][]ast.Constant
		wantMin ast.Constant
		wantMax ast.Constant
		wantSum ast.Constant
		wantAvg ast.Constant
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
			wantAvg: ast.Float64((float64(1) + float64(1) + float64(3)) / float64(3)),
		},
		// Note that EvalReducerFn is never called with empty list of rows.
		// The reducer functions may be called in user code though.
		{
			rows:    nil,
			wantMin: ast.Number(math.MaxInt64),
			wantMax: ast.Number(math.MinInt64),
			wantSum: ast.Number(0),
			wantAvg: ast.Float64(math.NaN()),
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
		gotAvg, err := EvalReduceFn(ast.ApplyFn{symbols.Avg, []ast.BaseTerm{ast.Variable{"X"}}}, rows)
		if err != nil {
			t.Fatalf("EvalReduceFn(Avg, %v) failed with %v", rows, err)
		}
		if test.wantAvg != gotAvg {
			t.Errorf("EvalReduceFn(Avg, %v)=%v want %v", rows, gotAvg, test.wantAvg)
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

func TestNameFuns(t *testing.T) {
	tests := []struct {
		name string
		expr ast.ApplyFn
		want ast.Constant
	}{
		{
			name: "name root single part",
			expr: ast.ApplyFn{symbols.NameRoot, []ast.BaseTerm{name("/a")}},
			want: name("/a"),
		},
		{
			name: "name root",
			expr: ast.ApplyFn{symbols.NameRoot, []ast.BaseTerm{name("/a/b/c")}},
			want: name("/a"),
		},
		{
			name: "name list single part",
			expr: ast.ApplyFn{symbols.NameList, []ast.BaseTerm{name("/a")}},
			want: ast.List([]ast.Constant{name("/a")}),
		},
		{
			name: "name list",
			expr: ast.ApplyFn{symbols.NameList, []ast.BaseTerm{name("/a/b/c")}},
			want: ast.List([]ast.Constant{name("/a"), name("/b"), name("/c")}),
		},
		{
			name: "name tip",
			expr: ast.ApplyFn{symbols.NameTip, []ast.BaseTerm{name("/a/b/c")}},
			want: name("/c"),
		},
		{
			name: "name tip single part",
			expr: ast.ApplyFn{symbols.NameTip, []ast.BaseTerm{name("/c")}},
			want: name("/c"),
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

func TestStringReplace(t *testing.T) {
	tests := []struct {
		provided string
		old      string
		new      string
		count    int64
		want     ast.Constant
	}{
		{"borked", "o", "0", 1, ast.String("b0rked")},
		{"aaa", "a", "b", 2, ast.String("bba")},
		{"aaa", "a", "b", -1, ast.String("bbb")},
		{"/a/b/c", "/", "", -1, ast.String("abc")},
	}
	for _, test := range tests {
		term := ast.ApplyFn{
			symbols.StringReplace,
			[]ast.BaseTerm{
				ast.String(test.provided),
				ast.String(test.old),
				ast.String(test.new),
				ast.Number(test.count),
			}}
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

func TestSqrt(t *testing.T) {
	tests := []struct {
		arg  ast.Constant
		want ast.Constant
	}{
		{ast.Number(0), ast.Float64(0)},
		{ast.Number(4), ast.Float64(2)},
		{ast.Number(9), ast.Float64(3)},
		{ast.Float64(2.25), ast.Float64(1.5)},
	}
	for _, tc := range tests {
		term := ast.ApplyFn{symbols.Sqrt, []ast.BaseTerm{tc.arg}}
		gotBase, err := EvalExpr(term, ast.ConstSubstMap{})
		if err != nil {
			t.Fatalf("EvalExpr(%v) error: %v", term, err)
		}
		got, ok := gotBase.(ast.Constant)
		if !ok {
			t.Fatalf("EvalExpr(%v) did not return Constant, got %T", term, gotBase)
		}
		if !got.Equals(tc.want) {
			t.Errorf("EvalExpr(%v) = %v want %v", term, got, tc.want)
		}
	}
}

func TestSqrtNegative(t *testing.T) {
	term := ast.ApplyFn{symbols.Sqrt, []ast.BaseTerm{ast.Number(-1)}}
	gotBase, err := EvalExpr(term, ast.ConstSubstMap{})
	if err != nil {
		t.Fatalf("EvalExpr(%v) error: %v", term, err)
	}
	got, ok := gotBase.(ast.Constant)
	if !ok {
		t.Fatalf("EvalExpr(%v) did not return Constant, got %T", term, gotBase)
	}
	f, err := got.Float64Value()
	if err != nil {
		t.Fatalf("got.Float64Value() error: %v", err)
	}
	if !math.IsNaN(f) {
		t.Errorf("EvalExpr(%v) = %v want NaN", term, f)
	}
}

func TestReducerCollectToMap(t *testing.T) {
	tests := []struct {
		name string
		rows [][]ast.Constant
	}{
		{
			name: "simple key-value pairs",
			rows: [][]ast.Constant{
				{name("/key1"), ast.Number(1)},
				{name("/key2"), ast.Number(2)},
				{name("/key3"), ast.Number(3)},
			},
		},
		{
			name: "duplicate keys (should use first occurrence)",
			rows: [][]ast.Constant{
				{name("/key1"), ast.Number(1)},
				{name("/key2"), ast.Number(2)},
				{name("/key1"), ast.Number(10)}, // duplicate key, should be ignored
			},
		},
		{
			name: "empty input",
			rows: [][]ast.Constant{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var rows []ast.ConstSubstList
			for _, row := range test.rows {
				if len(row) > 0 {
					rows = append(rows, makeConstSubstList(makeVarList(len(row)), row))
				}
			}

			// CollectToMap expects exactly 2 arguments (key variable, value variable)
			args := []ast.BaseTerm{ast.Variable{"X0"}, ast.Variable{"X1"}}
			expr := ast.ApplyFn{symbols.CollectToMap, args}
			got, err := EvalReduceFn(expr, rows)
			if err != nil {
				t.Fatalf("EvalReduceFn(%v,%v) failed with %v", expr, rows, err)
			}

			// Validate that we got a map
			if got.Type != ast.MapShape {
				t.Errorf("EvalReduceFn(%v,%v) returned %v (type %v), expected map", expr, rows, got, got.Type)
				return
			}

			// For empty input, expect empty map
			if len(test.rows) == 0 {
				if !got.IsMapNil() {
					t.Errorf("EvalReduceFn with empty input should return empty map, got %v", got)
				}
				return
			}

			// Check that all expected keys are present and duplicate keys are handled correctly
			expectedKeys := make(map[string]ast.Constant)
			for _, row := range test.rows {
				key := row[0]
				value := row[1]
				keyStr := key.String()
				if _, exists := expectedKeys[keyStr]; !exists {
					expectedKeys[keyStr] = value
				}
			}

			actualEntries := make(map[string]ast.Constant)
			_, _ = got.MapValues(func(key, val ast.Constant) error {
				actualEntries[key.String()] = val
				return nil
			}, func() error { return nil })

			// Check lengths match
			if len(actualEntries) != len(expectedKeys) {
				t.Errorf("Expected %d entries, got %d", len(expectedKeys), len(actualEntries))
			}

			// Check each expected key-value pair
			for keyStr, expectedVal := range expectedKeys {
				actualVal, exists := actualEntries[keyStr]
				if !exists {
					t.Errorf("Missing key %s", keyStr)
				} else if !actualVal.Equals(expectedVal) {
					t.Errorf("For key %s, expected %v, got %v", keyStr, expectedVal, actualVal)
				}
			}
		})
	}
}

func TestCollectToMapIntegration(t *testing.T) {
	// Test that our new CollectToMap works with the broader function evaluation system

	// Create some test data representing key-value pairs
	keyVals := []ast.Constant{
		name("/key1"), ast.Number(100),
		name("/key2"), ast.Number(200),
		name("/key3"), ast.Number(300),
	}

	// Create a map using the Map function
	mapExpr := ast.ApplyFn{symbols.Map, []ast.BaseTerm{
		keyVals[0], keyVals[1],
		keyVals[2], keyVals[3],
		keyVals[4], keyVals[5],
	}}

	result, err := EvalApplyFn(mapExpr, ast.ConstSubstMap{})
	if err != nil {
		t.Fatalf("Failed to create map: %v", err)
	}

	if result.Type != ast.MapShape {
		t.Errorf("Expected map, got %v", result.Type)
	}

	// Test that we can look up values in the created map
	lookup1 := ast.ApplyFn{symbols.MapGet, []ast.BaseTerm{result, keyVals[0]}}
	val1, err := EvalApplyFn(lookup1, ast.ConstSubstMap{})
	if err != nil {
		t.Fatalf("Failed to lookup key: %v", err)
	}

	if !val1.Equals(keyVals[1]) {
		t.Errorf("Expected %v, got %v", keyVals[1], val1)
	}
}

func TestReducerCollectToMapErrors(t *testing.T) {
tests := []struct {
name        string
args        []ast.BaseTerm
expectedErr string
}{
{
name:        "wrong number of arguments - too few",
args:        []ast.BaseTerm{ast.Variable{"X0"}},
expectedErr: "collect_to_map requires exactly 2 arguments (key, value), got 1",
},
{
name:        "wrong number of arguments - too many",
args:        []ast.BaseTerm{ast.Variable{"X0"}, ast.Variable{"X1"}, ast.Variable{"X2"}},
expectedErr: "collect_to_map requires exactly 2 arguments (key, value), got 3",
},
}

for _, test := range tests {
t.Run(test.name, func(t *testing.T) {
expr := ast.ApplyFn{symbols.CollectToMap, test.args}
var rows []ast.ConstSubstList // empty rows
_, err := EvalReduceFn(expr, rows)
if err == nil {
t.Errorf("Expected error, but got none")
} else if err.Error() != test.expectedErr {
t.Errorf("Expected error %q, got %q", test.expectedErr, err.Error())
}
})
}
}
