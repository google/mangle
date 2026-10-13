package analysis

import (
	"fmt"
	"testing"

	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/symbols"
)

// Tests for the two modes in which reducer functions are used, and the
// explicit distinction between them in the bounds checker:
//
//   - aggregation mode: reducer statements of a group_by transform,
//     whose arguments are variables ranging over the values of the
//     reduced columns, one per row of each group (typed by
//     typeOfReducerStmt), and
//   - list mode: ordinary per-row applications over list values, e.g.
//     fn:max([1, 2]) in a premise or nested inside another function
//     (typed via boundOfArg/listReducerBound).
//
// The runtime keeps the two modes strictly apart (functional.EvalReduceFn
// vs functional.EvalApplyFn); the tests here pin down that the type
// checker agrees with the runtime in both directions.
//
// All rules are of the shape
//
//	result(X, Y) :- source(X, Z) |> do fn:group_by(X), let Y = <reducer>(Z).
//
// with the group-by key X declared as /string and the reduced column Z
// declared as sourceBound.

// runBoundsCheck runs BoundsCheck on the given rule, with result(Y)
// declared as resultBound and source(X, Z) declared as (string, sourceBound).
func runBoundsCheck(t *testing.T, rule string, resultBound, sourceBound ast.BaseTerm) error {
	t.Helper()
	tc := newBoundsTestCase(t, []ast.Clause{
		clause(rule),
	}, []ast.Decl{
		makeSimpleDecl(atom("result(X, Y)"), ast.StringBound, resultBound),
		makeSimpleDecl(atom("source(X, Z)"), ast.StringBound, sourceBound),
	})
	bc, err := newBoundsAnalyzer(&tc.programInfo, symbols.NewNameTrie(), nil, tc.rulesMap)
	if err != nil {
		t.Fatal(err)
	}
	return bc.BoundsCheck()
}

func TestReducerModes_Aggregation(t *testing.T) {
	tests := []struct {
		name        string
		reducer     string
		sourceBound ast.BaseTerm // declared bound of the reduced column
		resultBound ast.BaseTerm // declared bound of the result column
		wantErr     bool
	}{
		// Aggregations over scalar columns work and produce the
		// element type.
		{"sum of numbers", "fn:sum", ast.NumberBound, ast.NumberBound, false},
		{"max of numbers", "fn:max", ast.NumberBound, ast.NumberBound, false},
		{"min of numbers", "fn:min", ast.NumberBound, ast.NumberBound, false},
		{"float:sum of floats", "fn:float:sum", ast.Float64Bound, ast.Float64Bound, false},
		{"float:max of floats", "fn:float:max", ast.Float64Bound, ast.Float64Bound, false},
		{"duration:sum of durations", "fn:duration:sum", ast.DurationBound, ast.DurationBound, false},
		{"time:max of times", "fn:time:max", ast.TimeBound, ast.TimeBound, false},
		// Type mismatches between the column and the reducer are rejected.
		{"sum over strings", "fn:sum", ast.StringBound, ast.NumberBound, true},
		{"float:max over numbers", "fn:float:max", ast.NumberBound, ast.Float64Bound, true},
		{"duration:sum over numbers", "fn:duration:sum", ast.NumberBound, ast.DurationBound, true},
		// The result type is checked (float reducers used to be typed
		// as /any, letting wrong declarations through).
		{"float:sum declared as string", "fn:float:sum", ast.Float64Bound, ast.StringBound, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := fmt.Sprintf("result(X, Y) :- source(X, Z) |> do fn:group_by(X), let Y = %s(Z).", tt.reducer)
			err := runBoundsCheck(t, rule, tt.resultBound, tt.sourceBound)
			if tt.wantErr && err == nil {
				t.Errorf("BoundsCheck passed, want error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("BoundsCheck failed: %v", err)
			}
		})
	}
}

func TestReducerModes_AggregationRequiresVariableArguments(t *testing.T) {
	// The runtime iterates the values of the argument variable across
	// the rows of the group (EvalReduceFn type-asserts the argument to
	// a variable and panics otherwise). A list literal argument in a
	// group_by reducer statement must therefore be rejected here,
	// instead of panicking at evaluation time.
	err := runBoundsCheck(t,
		"result(X, Y) :- source(X, Z) |> do fn:group_by(X), let Y = fn:max([1, 2]).",
		ast.NumberBound, ast.NumberBound)
	if err == nil {
		t.Error("BoundsCheck passed for fn:max over a list literal in a group_by, want error")
	}
}

func TestReducerModes_AggregationOverListColumn(t *testing.T) {
	// Reducing over a column of lists is also rejected: the runtime
	// iterates the (scalar) values of the column, and list values are
	// not among them (the reduction would silently fall back to the
	// empty-group sentinel).
	err := runBoundsCheck(t,
		"result(X, Y) :- source(X, Z) |> do fn:group_by(X), let Y = fn:max(Z).",
		ast.NumberBound, symbols.NewListType(ast.NumberBound))
	if err == nil {
		t.Error("BoundsCheck passed for fn:max over a list column, want error")
	}
}

func TestReducerModes_ListModeInsideGroupBy(t *testing.T) {
	// Non-reducer statements of a group_by transform are ordinary
	// per-row applications; a reducer nested inside one operates on a
	// list value (list mode), e.g. fn:max([1, 2]) = 2.
	err := runBoundsCheck(t,
		"result(X, Y) :- source(X, Z) |> do fn:group_by(X), let Y = fn:plus(fn:max([1, 2]), 3).",
		ast.NumberBound, ast.NumberBound)
	if err != nil {
		t.Errorf("BoundsCheck failed: %v", err)
	}
	// ... and must agree with the declared type.
	err = runBoundsCheck(t,
		"result(X, Y) :- source(X, Z) |> do fn:group_by(X), let Y = fn:plus(fn:max([1, 2]), 3).",
		ast.StringBound, ast.NumberBound)
	if err == nil {
		t.Error("BoundsCheck passed for fn:max(...) declared as /string, want error")
	}
}

func TestReducerModes_AggregationCollect(t *testing.T) {
	// fn:collect in a group_by collects the per-row values into a list.
	err := runBoundsCheck(t,
		"result(X, Y) :- source(X, Z) |> do fn:group_by(X), let Y = fn:collect(Z).",
		symbols.NewListType(ast.NumberBound), ast.NumberBound)
	if err != nil {
		t.Errorf("BoundsCheck failed: %v", err)
	}
	// The collected element type must match the declaration.
	err = runBoundsCheck(t,
		"result(X, Y) :- source(X, Z) |> do fn:group_by(X), let Y = fn:collect(Z).",
		symbols.NewListType(ast.StringBound), ast.NumberBound)
	if err == nil {
		t.Error("BoundsCheck passed for fn:collect of /number declared as .List(/string), want error")
	}
}
