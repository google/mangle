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

package functional

import (
	"testing"

	"github.com/google/mangle/ast"
	"github.com/google/mangle/symbols"
)

func TestIntervalStart(t *testing.T) {
	tests := []struct {
		name    string
		start   int64
		end     int64
		want    int64
		wantErr bool
	}{
		{
			name:  "simple interval",
			start: 1000,
			end:   2000,
			want:  1000,
		},
		{
			name:  "point interval",
			start: 1500,
			end:   1500,
			want:  1500,
		},
		{
			name:  "negative start",
			start: -1000,
			end:   1000,
			want:  -1000,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			startConst := ast.Number(tc.start)
			endConst := ast.Number(tc.end)
			interval := ast.Pair(&startConst, &endConst)

			result, err := EvalApplyFn(ast.ApplyFn{
				Function: symbols.IntervalStart,
				Args:     []ast.BaseTerm{interval},
			}, nil)

			if (err != nil) != tc.wantErr {
				t.Fatalf("EvalApplyFn() error = %v, wantErr %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			got, err := result.NumberValue()
			if err != nil {
				t.Fatalf("result.NumberValue() error = %v", err)
			}
			if got != tc.want {
				t.Errorf("fn:interval:start got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestIntervalEnd(t *testing.T) {
	tests := []struct {
		name    string
		start   int64
		end     int64
		want    int64
		wantErr bool
	}{
		{
			name:  "simple interval",
			start: 1000,
			end:   2000,
			want:  2000,
		},
		{
			name:  "point interval",
			start: 1500,
			end:   1500,
			want:  1500,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			startConst := ast.Number(tc.start)
			endConst := ast.Number(tc.end)
			interval := ast.Pair(&startConst, &endConst)

			result, err := EvalApplyFn(ast.ApplyFn{
				Function: symbols.IntervalEnd,
				Args:     []ast.BaseTerm{interval},
			}, nil)

			if (err != nil) != tc.wantErr {
				t.Fatalf("EvalApplyFn() error = %v, wantErr %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			got, err := result.NumberValue()
			if err != nil {
				t.Fatalf("result.NumberValue() error = %v", err)
			}
			if got != tc.want {
				t.Errorf("fn:interval:end got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestIntervalDuration(t *testing.T) {
	tests := []struct {
		name    string
		start   int64
		end     int64
		want    int64
		wantErr bool
	}{
		{
			name:  "simple interval",
			start: 1000,
			end:   2000,
			want:  1000,
		},
		{
			name:  "point interval",
			start: 1500,
			end:   1500,
			want:  0,
		},
		{
			name:  "large duration",
			start: 0,
			end:   1000000000, // 1 second in nanoseconds
			want:  1000000000,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			startConst := ast.Number(tc.start)
			endConst := ast.Number(tc.end)
			interval := ast.Pair(&startConst, &endConst)

			result, err := EvalApplyFn(ast.ApplyFn{
				Function: symbols.IntervalDuration,
				Args:     []ast.BaseTerm{interval},
			}, nil)

			if (err != nil) != tc.wantErr {
				t.Fatalf("EvalApplyFn() error = %v, wantErr %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			got, err := result.NumberValue()
			if err != nil {
				t.Fatalf("result.NumberValue() error = %v", err)
			}
			if got != tc.want {
				t.Errorf("fn:interval:duration got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestIntervalFunctionsErrors(t *testing.T) {
	// Test wrong argument count
	t.Run("wrong arg count for start", func(t *testing.T) {
		_, err := EvalApplyFn(ast.ApplyFn{
			Function: symbols.IntervalStart,
			Args:     []ast.BaseTerm{ast.Number(1), ast.Number(2)},
		}, nil)
		if err == nil {
			t.Error("expected error for wrong argument count")
		}
	})

	// Test wrong type
	t.Run("wrong type for interval", func(t *testing.T) {
		_, err := EvalApplyFn(ast.ApplyFn{
			Function: symbols.IntervalStart,
			Args:     []ast.BaseTerm{ast.Number(1)},
		}, nil)
		if err == nil {
			t.Error("expected error for non-pair argument")
		}
	})
}
