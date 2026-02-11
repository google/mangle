package interpreter

import (
	"bytes"
	"testing"

	"github.com/google/mangle/ast"
)

func TestInterpreter_TemporalSequence_Pop(t *testing.T) {
	out := &bytes.Buffer{}
	i := New(out, "", nil)

	// Define temporal facts and rules
	prog := `
Decl event_a(Name) temporal bound [/name].
Decl event_b(Name) temporal bound [/name].
Decl match(Name) bound [/name].

event_a(/u1)@[2024-01-01T10:00:00].
event_b(/u1)@[2024-01-01T10:05:00].

event_a(/u2)@[2024-01-01T10:00:00].
event_b(/u2)@[2024-01-01T10:15:00].

match(U) :-
  event_b(U)@[Tb],
  event_a(U)@[Ta],
  :time:lt(Ta, Tb),
  Diff = fn:time:sub(Tb, Ta),
  Limit = fn:duration:parse('10m'),
  :duration:le(Diff, Limit).
`
	if err := i.Define(prog); err != nil {
		t.Fatalf("Define failed: %v", err)
	}

	// Query match(X)
	queryAtom, _ := i.ParseQuery("match(X)")
	res, err := i.Query(queryAtom)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Expect /u1 only
	foundU1 := false
	foundU2 := false
	for _, atom := range res {
		a, _ := atom.(ast.Atom)
		if a.Args[0].String() == "/u1" {
			foundU1 = true
		}
		if a.Args[0].String() == "/u2" {
			foundU2 = true
		}
	}

	if !foundU1 {
		t.Errorf("Expected match(/u1), not found. Results: %v", res)
	}
	if foundU2 {
		t.Errorf("Did not expect match(/u2), found it. Results: %v", res)
	}

	// Test Pop
	i.Pop()

	// Query again. Facts should be gone.
	queryAtom, _ = i.ParseQuery("match(X)")
	res, err = i.Query(queryAtom)
	if err != nil {
		t.Fatalf("Query failed after Pop: %v", err)
	}
	if len(res) != 0 {
		t.Errorf("Expected 0 results for match(X) after Pop, got %d: %v", len(res), res)
	}

	// Verify event_a facts are gone
	queryAtom, _ = i.ParseQuery("event_a(X)")
	res, err = i.Query(queryAtom)
	if err != nil {
		t.Fatalf("Query failed after Pop: %v", err)
	}
	if len(res) != 0 {
		t.Errorf("Expected 0 results for event_a(X) after Pop, got %d: %v", len(res), res)
	}

	// Verify predicate is unknown
	if err := i.Show("match"); err == nil {
		t.Error("Expected error from Show('match') after Pop, got nil")
	}
}
