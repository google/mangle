package analysis

import (
	"strings"
	"testing"

	"codeberg.org/TauCeti/mangle-go/parse"
)

// Regression test for https://codeberg.org/TauCeti/mangle-go/issues/103:
// v0.6.0 type checker rejected valid list arguments to fn:max, fn:min
// and fn:sum, because the bounds checker compared the type of the list
// argument itself (fn:List(/number)) against the element type (/number).
func TestIssue103(t *testing.T) {
	tests := []struct {
		src     string
		wantErr bool
	}{
		// List arguments to number reducers must be accepted.
		{"result(X) :- X = fn:max([1, 2]).", false},
		{"result(X) :- X = fn:min([1, 2]).", false},
		{"result(X) :- X = fn:sum([1, 2]).", false},
		// Nested reducer applications.
		{"result(X) :- X = fn:sum([fn:max([1, 2]), 3]).", false},
		{"result(X) :- X = fn:max([fn:min([1, 2]), fn:sum([3, 4])]).", false},
		// Type-specific reducers over lists.
		{"result(X) :- X = fn:duration:max([fn:duration:parse('1h'), fn:duration:parse('2h')]).", false},
		{"result(X) :- X = fn:duration:min([fn:duration:parse('1h'), fn:duration:parse('2h')]).", false},
		{"result(X) :- X = fn:duration:sum([fn:duration:parse('1h'), fn:duration:parse('2h')]).", false},
		{"result(X) :- X = fn:time:min([fn:time:parse_rfc3339('2024-01-01T00:00:00Z'), fn:time:parse_rfc3339('2024-06-01T00:00:00Z')]).", false},
		{"result(X) :- X = fn:time:max([fn:time:parse_rfc3339('2024-01-01T00:00:00Z'), fn:time:parse_rfc3339('2024-06-01T00:00:00Z')]).", false},
		// Reducers applied to lists of the wrong element type are rejected.
		{"result(X) :- X = fn:sum([fn:duration:parse('1h'), fn:duration:parse('2h')]).", true},
		{"result(X) :- X = fn:max([fn:duration:parse('1h'), fn:duration:parse('2h')]).", true},
		{"result(X) :- X = fn:duration:sum([1, 2]).", true},
		{"result(X) :- X = fn:time:max([fn:duration:parse('1h'), fn:duration:parse('2h')]).", true},
	}
	for _, tt := range tests {
		unit, err := parse.Unit(strings.NewReader(tt.src))
		if err != nil {
			t.Fatalf("parse.Unit() error = %v", err)
		}
		_, err = AnalyzeAndCheckBounds([]parse.SourceUnit{unit}, nil, ErrorForBoundsMismatch)
		if tt.wantErr && err == nil {
			t.Errorf("%s: expected error, got nil", tt.src)
		}
		if !tt.wantErr && err != nil {
			t.Errorf("%s: unexpected error: %v", tt.src, err)
		}
	}
}
