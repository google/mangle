package factstore

import (
	"bytes"
	"testing"

	"github.com/google/mangle/ast"
)

func TestRoundTrip(t *testing.T) {
	sc := SimpleColumn{}
	m := NewSimpleInMemoryStore()
	facts := []ast.Atom{
		atom("baz()"),
		atom("foo(`\n/bar`)"),
		atom("foo(/zzz)"),
		atom("bar(/bar,1,/baz)"),
		atom("bar(/bar,0,/def)"),
		atom("bar(/abc,1,/def)"),
		evalAtom("bar([/abc],1,/def)"),
		evalAtom("bar([/abc, /def], 1, /def)"),
		evalAtom("baz([/abc : 1,  /def : 2], 1, /def)"),
		evalAtom("baz({/abc : 1,  /def : 2}, 1, /def)"),
	}
	for _, f := range facts {
		m.Add(f)
	}
	var buf bytes.Buffer
	if err := sc.WriteTo(m, &buf); err != nil {
		t.Fatal(err)
	}

	n := NewSimpleInMemoryStore()

	if err := sc.ReadInto(bytes.NewReader(buf.Bytes()), n); err != nil {
		t.Fatal(err)
	}
	if n.EstimateFactCount() != len(facts) {
		t.Fatalf("fact count %d want %d", n.EstimateFactCount(), len(facts))
	}
	for _, f := range facts {
		if !n.Contains(f) {
			t.Errorf("missing fact: %s", f.String())
		}
	}
}
