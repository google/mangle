package factstore

import (
	"bytes"
	"compress/gzip"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/mangle/ast"
)

const (
	wantNumFooFacts = 2
	wantNumBarFacts = 5
	wantNumQazFacts = 2
)

func testStore(t *testing.T) *SimpleInMemoryStore {
	t.Helper()
	m := NewSimpleInMemoryStore()
	facts := []ast.Atom{
		atom("baz()"),
		atom("foo(`\n/bar`)"),
		atom("foo(/zzz)"),
		atom("bar(/r,1,/z)"),
		atom("bar(/t,2,'😤')"),
		atom("bar(/g,3,/h)"),
		evalAtom("bar([/abc],4,/def)"),
		evalAtom("bar([/abc, /def], 5, /def)"),
		evalAtom("qaz([/abc : 123,  /def : 345], 10, b'\\x80\\x81')"),
		evalAtom("qaz({/abc : 456,  /def : 678}, 20, /zzz)"),
	}
	for _, f := range facts {
		m.Add(f)
	}
	if m.EstimateFactCount() != len(facts) {
		t.Fatalf("SimpleInMemoryStore.EstimateFactCount() =  %d want %d", m.EstimateFactCount(), len(facts))
	}
	return &m
}

func TestOutput(t *testing.T) {
	m := testStore(t)
	sc := SimpleColumn{true /* deterministic */}
	var buf bytes.Buffer
	if err := sc.WriteTo(m, &buf); err != nil {
		t.Fatal(err)
	}

	want := `4
baz 0 1
foo 1 2
bar 3 5
qaz 3 2
"\n/bar"
/zzz
/g
[/abc]
/t
/r
[/abc, /def]
3
4
2
1
5
/h
/def
"\u{01f624}"
/z
/def
[/def : 345, /abc : 123]
{/def : 678, /abc : 456}
10
20
b"\x80\x81"
/zzz
`
	if diff := cmp.Diff(want, string(buf.Bytes())); diff != "" {
		t.Errorf("WriteTo() unexpected difference -want +got %v", diff)
	}
}

func TestRoundTrip(t *testing.T) {
	m := testStore(t)
	sc := SimpleColumn{true /* deterministic */}
	var buf bytes.Buffer
	if err := sc.WriteTo(m, &buf); err != nil {
		t.Fatal(err)
	}

	n := NewSimpleInMemoryStore()
	if err := sc.ReadInto(bytes.NewReader(buf.Bytes()), n); err != nil {
		t.Fatal(err)
	}
	if n.EstimateFactCount() != m.EstimateFactCount() {
		t.Fatalf("fact count %d want %d", n.EstimateFactCount(), m.EstimateFactCount())
	}
	for _, p := range m.ListPredicates() {
		m.GetFacts(ast.NewQuery(p), func(fact ast.Atom) error {
			if !n.Contains(fact) {
				t.Errorf("missing fact: %s", fact.String())
			}
			return nil
		})
	}
	for _, p := range n.ListPredicates() {
		n.GetFacts(ast.NewQuery(p), func(fact ast.Atom) error {
			if !m.Contains(fact) {
				t.Errorf("extra fact: %s", fact.String())
			}
			return nil
		})
	}
}

func TestStore(t *testing.T) {
	m := testStore(t)
	sc := SimpleColumn{true /* deterministic */}
	var buf bytes.Buffer
	if err := sc.WriteTo(m, &buf); err != nil {
		t.Fatal(err)
	}

	s, err := NewSimpleColumnStore(func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(buf.Bytes())), nil
	})

	if err != nil {
		t.Fatal(err)
	}

	if len(s.predicateFactCount) != len(m.ListPredicates()) {
		t.Errorf("NewSimpleColumnStore: got %d predicates want %d", len(s.predicateFactCount), len(m.ListPredicates()))
	}

	var foundBaz bool
	s.GetFacts(atom("baz()"), func(a ast.Atom) error {
		if !a.Equals(atom("baz()")) {
			t.Errorf("GetFacts(baz()): got %v want baz()", a)
		}
		foundBaz = true
		return nil
	})
	if !foundBaz {
		t.Errorf("GetFacts(baz()): got nothing want baz()")
	}

	tests := []struct {
		query string
		want  int
	}{
		{"foo(X)", wantNumFooFacts},
		{"bar(X, Y, Z)", wantNumBarFacts},
		{"qaz(X, Y, Z)", wantNumQazFacts},
	}
	for _, test := range tests {
		var count int
		s.GetFacts(atom(test.query), func(a ast.Atom) error {
			if !m.Contains(a) {
				t.Errorf("GetFacts(%s): unexpected fact: %v", test.query, a)
			}
			count++
			return nil
		})
		if count != test.want {
			t.Errorf("GetFacts(%s): got %d want %d facts", test.query, count, test.want)
		}
	}
}

func TestFiltered(t *testing.T) {
	m := testStore(t)
	sc := SimpleColumn{true /* deterministic */}
	var buf bytes.Buffer
	if err := sc.WriteTo(m, &buf); err != nil {
		t.Fatal(err)
	}

	s, err := NewSimpleColumnStore(func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(buf.Bytes())), nil
	})

	if err != nil {
		t.Fatal(err)
	}

	if !s.Contains(evalAtom("bar([/abc, /def], 5, /def)")) {
		t.Errorf("Contains(bar([/abc, /def], 5, /def))=false want true")
	}

	if s.Contains(evalAtom("bar(/nope, /nope, /nope)")) {
		t.Errorf("Contains(bar(/nope, /nope, /nope)=true want false")
	}
}

func TestEmptyStore(t *testing.T) {
	emptyStore := NewSimpleInMemoryStore()
	var b bytes.Buffer
	sc := SimpleColumn{}
	if err := sc.WriteTo(emptyStore, &b); err != nil {
		t.Fatal(err)
	}
	store, err := NewSimpleColumnStoreFromBytes(b.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]ast.PredicateSym{}, store.ListPredicates()); diff != "" {
		t.Errorf("NewSimpleColumnStoreFromBytes: diff (-want +got) %v", diff)
	}
}

func sortBySymbol(a ast.PredicateSym, b ast.PredicateSym) bool {
	return a.Symbol < b.Symbol
}

func TestNewBytes(t *testing.T) {
	tmpStore := NewSimpleInMemoryStore()
	tmpStore.Add(atom("i(/persist)"))
	tmpStore.Add(atom("we(/persist)"))
	var b bytes.Buffer
	sc := SimpleColumn{}
	if err := sc.WriteTo(tmpStore, &b); err != nil {
		t.Fatal(err)
	}
	store, err := NewSimpleColumnStoreFromBytes(b.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(tmpStore.ListPredicates(), store.ListPredicates(),
		cmpopts.SortSlices(sortBySymbol)); diff != "" {
		t.Errorf("NewSimpleColumnStoreFromBytes: diff (-want +got) %v", diff)
	}
	if !store.Contains(atom("i(/persist)")) {
		t.Errorf("NewSimpleColumnStoreFromBytes: expected atom i(/persist)")
	}
	if !store.Contains(atom("we(/persist)")) {
		t.Errorf("NewSimpleColumnStoreFromBytes: expected atom i(/persist)")
	}
}

func TestGzip(t *testing.T) {
	tmpStore := NewSimpleInMemoryStore()
	tmpStore.Add(atom("i(/persist)"))
	tmpStore.Add(atom("we(/persist)"))
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	sc := SimpleColumn{}
	if err := sc.WriteTo(tmpStore, w); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	store, err := NewSimpleColumnStoreFromGzipBytes(b.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(tmpStore.ListPredicates(), store.ListPredicates(),
		cmpopts.SortSlices(sortBySymbol)); diff != "" {
		t.Errorf("NewSimpleColumnStoreFromGzipBytes: diff (-want +got) %v", diff)
	}
	if !store.Contains(atom("i(/persist)")) {
		t.Errorf("NewSimpleColumnStoreFromGzipBytes: expected atom i(/persist)")
	}
	if !store.Contains(atom("we(/persist)")) {
		t.Errorf("NewSimpleColumnStoreFromGzipBytes: expected atom i(/persist)")
	}
}
