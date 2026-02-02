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

package parse

import (
	"strings"
	"testing"
	"time"

	"github.com/google/mangle/ast"
)

func TestParseTemporalFact(t *testing.T) {
	tests := []struct {
		name       string
		str        string
		wantHead   ast.Atom
		wantTime   *ast.Interval
		wantErr    bool
	}{
		{
			name:     "simple fact without temporal annotation",
			str:      "foo(/bar).",
			wantHead: ast.NewAtom("foo", name("/bar")),
			wantTime: nil,
			wantErr:  false,
		},
		{
			name:     "fact with date interval",
			str:      "foo(/bar)@[2024-01-15, 2024-06-30].",
			wantHead: ast.NewAtom("foo", name("/bar")),
			wantTime: func() *ast.Interval {
				start := ast.NewTimestampBound(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC))
				end := ast.NewTimestampBound(time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC))
				i := ast.NewInterval(start, end)
				return &i
			}(),
			wantErr: false,
		},
		{
			name:     "fact with datetime interval",
			str:      "foo(/bar)@[2024-01-15T10:30:00, 2024-01-15T18:00:00].",
			wantHead: ast.NewAtom("foo", name("/bar")),
			wantTime: func() *ast.Interval {
				start := ast.NewTimestampBound(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC))
				end := ast.NewTimestampBound(time.Date(2024, 1, 15, 18, 0, 0, 0, time.UTC))
				i := ast.NewInterval(start, end)
				return &i
			}(),
			wantErr: false,
		},
		{
			name:     "fact with point interval (single timestamp)",
			str:      "login(/alice)@[2024-03-15].",
			wantHead: ast.NewAtom("login", name("/alice")),
			wantTime: func() *ast.Interval {
				t := ast.NewTimestampBound(time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC))
				i := ast.NewInterval(t, t)
				return &i
			}(),
			wantErr: false,
		},
		{
			name:     "fact with unbounded start (negative infinity)",
			str:      "admin(/bob)@[_, 2024-12-31].",
			wantHead: ast.NewAtom("admin", name("/bob")),
			wantTime: func() *ast.Interval {
				start := ast.PositiveInfinity() // Note: _ defaults to positive infinity, will need adjustment
				end := ast.NewTimestampBound(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC))
				i := ast.NewInterval(start, end)
				return &i
			}(),
			wantErr: false,
		},
		{
			name:     "fact with unbounded end (positive infinity)",
			str:      "employed(/alice)@[2020-01-01, _].",
			wantHead: ast.NewAtom("employed", name("/alice")),
			wantTime: func() *ast.Interval {
				start := ast.NewTimestampBound(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
				end := ast.PositiveInfinity()
				i := ast.NewInterval(start, end)
				return &i
			}(),
			wantErr: false,
		},
		{
			name:     "fact with variable bounds",
			str:      "active(X)@[T1, T2].",
			wantHead: ast.NewAtom("active", ast.Variable{Symbol: "X"}),
			wantTime: func() *ast.Interval {
				start := ast.NewVariableBound(ast.Variable{Symbol: "T1"})
				end := ast.NewVariableBound(ast.Variable{Symbol: "T2"})
				i := ast.NewInterval(start, end)
				return &i
			}(),
			wantErr: false,
		},
		{
			name:     "fact with 'now' as end bound",
			str:      "active(/alice)@[2024-01-01, now].",
			wantHead: ast.NewAtom("active", name("/alice")),
			wantTime: func() *ast.Interval {
				start := ast.NewTimestampBound(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
				end := ast.Now()
				i := ast.NewInterval(start, end)
				return &i
			}(),
			wantErr: false,
		},
		{
			name:     "fact with 'now' as start and end",
			str:      "event(/login)@[now].",
			wantHead: ast.NewAtom("event", name("/login")),
			wantTime: func() *ast.Interval {
				bound := ast.Now()
				i := ast.NewInterval(bound, bound)
				return &i
			}(),
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clause, err := Clause(test.str)
			if (err != nil) != test.wantErr {
				t.Fatalf("Clause(%q) error = %v, wantErr = %v", test.str, err, test.wantErr)
			}
			if err != nil {
				return
			}

			if !clause.Head.Equals(test.wantHead) {
				t.Errorf("Clause(%q).Head = %v, want %v", test.str, clause.Head, test.wantHead)
			}

			if test.wantTime == nil {
				if clause.HeadTime != nil {
					t.Errorf("Clause(%q).HeadTime = %v, want nil", test.str, clause.HeadTime)
				}
			} else {
				if clause.HeadTime == nil {
					t.Errorf("Clause(%q).HeadTime = nil, want %v", test.str, test.wantTime)
				} else if !clause.HeadTime.Equals(*test.wantTime) {
					t.Errorf("Clause(%q).HeadTime = %v, want %v", test.str, clause.HeadTime, test.wantTime)
				}
			}
		})
	}
}

func TestParseTemporalRule(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		wantHead ast.Atom
		wantTime *ast.Interval
		wantErr  bool
	}{
		{
			name:     "rule without temporal annotation",
			str:      "foo(X) :- bar(X).",
			wantHead: ast.NewAtom("foo", ast.Variable{Symbol: "X"}),
			wantTime: nil,
			wantErr:  false,
		},
		{
			name:     "rule with temporal annotation on head",
			str:      "active(X)@[T, T] :- login(X).",
			wantHead: ast.NewAtom("active", ast.Variable{Symbol: "X"}),
			wantTime: func() *ast.Interval {
				bound := ast.NewVariableBound(ast.Variable{Symbol: "T"})
				i := ast.NewInterval(bound, bound)
				return &i
			}(),
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clause, err := Clause(test.str)
			if (err != nil) != test.wantErr {
				t.Fatalf("Clause(%q) error = %v, wantErr = %v", test.str, err, test.wantErr)
			}
			if err != nil {
				return
			}

			if !clause.Head.Equals(test.wantHead) {
				t.Errorf("Clause(%q).Head = %v, want %v", test.str, clause.Head, test.wantHead)
			}

			if test.wantTime == nil {
				if clause.HeadTime != nil {
					t.Errorf("Clause(%q).HeadTime = %v, want nil", test.str, clause.HeadTime)
				}
			} else {
				if clause.HeadTime == nil {
					t.Errorf("Clause(%q).HeadTime = nil, want %v", test.str, test.wantTime)
				} else if !clause.HeadTime.Equals(*test.wantTime) {
					t.Errorf("Clause(%q).HeadTime = %v, want %v", test.str, clause.HeadTime, test.wantTime)
				}
			}
		})
	}
}

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:    "date only",
			input:   "2024-01-15",
			want:    time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "date and time without Z",
			input:   "2024-01-15T10:30:00",
			want:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "date and time with Z",
			input:   "2024-01-15T10:30:00Z",
			want:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "not-a-date",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseTimestamp(test.input)
			if (err != nil) != test.wantErr {
				t.Fatalf("parseTimestamp(%q) error = %v, wantErr = %v", test.input, err, test.wantErr)
			}
			if err == nil && !got.Equal(test.want) {
				t.Errorf("parseTimestamp(%q) = %v, want %v", test.input, got, test.want)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "days not supported",
			input:   "7d",
			wantErr: true, // Go's time.ParseDuration doesn't support days
		},
		{
			name:    "hours",
			input:   "24h",
			want:    24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "minutes",
			input:   "30m",
			want:    30 * time.Minute,
			wantErr: false,
		},
		{
			name:    "seconds",
			input:   "60s",
			want:    60 * time.Second,
			wantErr: false,
		},
		{
			name:    "milliseconds",
			input:   "500ms",
			want:    500 * time.Millisecond,
			wantErr: false,
		},
		{
			name:    "invalid suffix",
			input:   "10x",
			wantErr: true,
		},
		{
			name:    "too short",
			input:   "s",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseDuration(test.input)
			if (err != nil) != test.wantErr {
				t.Fatalf("parseDuration(%q) error = %v, wantErr = %v", test.input, err, test.wantErr)
			}
			if err == nil && got != test.want {
				t.Errorf("parseDuration(%q) = %v, want %v", test.input, got, test.want)
			}
		})
	}
}

func TestParseTemporalOperator(t *testing.T) {
	tests := []struct {
		name    string
		str     string
		wantOp  ast.TemporalOperatorType
		wantErr bool
	}{
		{
			name:    "diamond minus with durations",
			str:     "recently_active(X) :- <-[0s, 168h] active(X).",
			wantOp:  ast.DiamondMinus,
			wantErr: false,
		},
		{
			name:    "box minus with durations",
			str:     "stable(X) :- [-[0s, 8760h] employed(X).",
			wantOp:  ast.BoxMinus,
			wantErr: false,
		},
		{
			name:    "diamond plus with durations",
			str:     "upcoming(X) :- <+[0s, 168h] scheduled(X).",
			wantOp:  ast.DiamondPlus,
			wantErr: false,
		},
		{
			name:    "box plus with durations",
			str:     "guaranteed(X) :- [+[0s, 720h] available(X).",
			wantOp:  ast.BoxPlus,
			wantErr: false,
		},
		{
			name:    "negative duration rejected",
			str:     "bad(X) :- <-[-30m, 1h] foo(X).",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			unit, err := Unit(strings.NewReader(test.str))
			if (err != nil) != test.wantErr {
				t.Fatalf("Unit(%q) error = %v, wantErr = %v", test.str, err, test.wantErr)
			}
			if err != nil {
				return
			}

			if len(unit.Clauses) != 1 {
				t.Fatalf("Unit(%q) has %d clauses, want 1", test.str, len(unit.Clauses))
			}

			clause := unit.Clauses[0]
			if len(clause.Premises) != 1 {
				t.Fatalf("Clause has %d premises, want 1", len(clause.Premises))
			}

			tempLit, ok := clause.Premises[0].(ast.TemporalLiteral)
			if !ok {
				t.Fatalf("Premise is %T, want ast.TemporalLiteral", clause.Premises[0])
			}

			if tempLit.Operator == nil {
				t.Fatalf("TemporalLiteral.Operator is nil")
			}

			if tempLit.Operator.Type != test.wantOp {
				t.Errorf("Operator.Type = %v, want %v", tempLit.Operator.Type, test.wantOp)
			}
		})
	}
}

func TestTemporalFactBackwardCompatibility(t *testing.T) {
	// These are existing Mangle programs that should continue to work unchanged
	tests := []struct {
		name string
		str  string
	}{
		{
			name: "simple fact",
			str:  "foo(/bar).",
		},
		{
			name: "rule with body",
			str:  "foo(X) :- bar(X).",
		},
		{
			name: "recursive rule",
			str:  "reachable(X, Y) :- edge(X, Y). reachable(X, Z) :- edge(X, Y), reachable(Y, Z).",
		},
		{
			name: "rule with negation",
			str:  "active(X) :- registered(X), !deleted(X).",
		},
		{
			name: "rule with transform",
			str:  "count(N) :- item(X) |> do fn:group_by(), let N = fn:Count().",
		},
		{
			name: "rule with comparison",
			str:  "adult(X) :- person(X, Age), Age >= 18.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			unit, err := Unit(strings.NewReader(test.str))
			if err != nil {
				t.Fatalf("Unit(%q) failed: %v", test.str, err)
			}
			if len(unit.Clauses) == 0 {
				t.Errorf("Unit(%q) has no clauses", test.str)
			}
		})
	}
}

func TestParseTemporalDecl(t *testing.T) {
	tests := []struct {
		name         string
		str          string
		wantTemporal bool
		wantErr      bool
	}{
		{
			name:         "simple temporal declaration",
			str:          "Decl foo(X) temporal.",
			wantTemporal: true,
			wantErr:      false,
		},
		{
			name:         "temporal declaration with bounds",
			str:          "Decl event(X, Y) temporal bound [/name, /number].",
			wantTemporal: true,
			wantErr:      false,
		},
		{
			name:         "temporal declaration with descr",
			str:          `Decl activity(X) temporal descr [doc("A temporal predicate")].`,
			wantTemporal: true,
			wantErr:      false,
		},
		{
			name:         "non-temporal declaration",
			str:          "Decl bar(X, Y) bound [/string, /number].",
			wantTemporal: false,
			wantErr:      false,
		},
		{
			name:         "temporal declaration with multiple bounds",
			str:          "Decl status(X, Y, Z) temporal bound [/name, /string, /number].",
			wantTemporal: true,
			wantErr:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			unit, err := Unit(strings.NewReader(test.str))
			if (err != nil) != test.wantErr {
				t.Fatalf("Unit(%q) error = %v, wantErr = %v", test.str, err, test.wantErr)
			}
			if err != nil {
				return
			}

			// The parser adds a default Package decl when none specified, so we look for
			// the user's declaration (second one) or find the non-Package declaration.
			var decl ast.Decl
			for _, d := range unit.Decls {
				if d.DeclaredAtom.Predicate.Symbol != "Package" {
					decl = d
					break
				}
			}

			if decl.DeclaredAtom.Predicate.Symbol == "" {
				t.Fatalf("Unit(%q) has no user declarations", test.str)
			}

			if decl.IsTemporal() != test.wantTemporal {
				t.Errorf("Decl.IsTemporal() = %v, want %v", decl.IsTemporal(), test.wantTemporal)
			}
		})
	}
}
