package factstore

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/google/mangle/ast"
)

func BenchmarkAdd(b *testing.B) {
	for _, store := range []FactStoreWithRemove{
		NewSimpleInMemoryStore(),
		NewIndexedInMemoryStore(),
		NewMultiIndexedInMemoryStore(),
		NewMultiIndexedArrayInMemoryStore(),
		NewConcurrentFactStore(NewSimpleInMemoryStore()),
		NewColumnarStore(),
	} {
		b.Run(fmt.Sprintf("%T", store), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				p := ast.PredicateSym{Symbol: fmt.Sprintf("p%d", rand.Intn(10)), Arity: 2}
				c1 := ast.String(fmt.Sprintf("c%d", rand.Intn(100)))
				c2 := ast.String(fmt.Sprintf("c%d", rand.Intn(100)))
				store.Add(ast.Atom{p, []ast.BaseTerm{c1, c2}})
			}
		})
	}
}

func BenchmarkGetFacts(b *testing.B) {
	for _, store := range []FactStoreWithRemove{
		NewSimpleInMemoryStore(),
		NewIndexedInMemoryStore(),
		NewMultiIndexedInMemoryStore(),
		NewMultiIndexedArrayInMemoryStore(),
		NewConcurrentFactStore(NewSimpleInMemoryStore()),
		NewColumnarStore(),
	} {
		for i := 0; i < 1000000; i++ {
			p := ast.PredicateSym{Symbol: fmt.Sprintf("p%d", i%10), Arity: 2}
			c1 := ast.String(fmt.Sprintf("c%d", i%100))
			c2 := ast.String(fmt.Sprintf("c%d", i%100))
			store.Add(ast.Atom{p, []ast.BaseTerm{c1, c2}})
		}

		b.Run(fmt.Sprintf("%T", store), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				p := ast.PredicateSym{Symbol: fmt.Sprintf("p%d", rand.Intn(10)), Arity: 2}
				c1 := ast.String(fmt.Sprintf("c%d", rand.Intn(100)))
				store.GetFacts(ast.Atom{p, []ast.BaseTerm{c1, ast.Variable{"_"}}}, func(a ast.Atom) error {
					return nil
				})
			}
		})
	}
}

func BenchmarkGetFacts_BigQuery(b *testing.B) {
	for _, store := range []FactStoreWithRemove{
		NewSimpleInMemoryStore(),
		NewIndexedInMemoryStore(),
		NewMultiIndexedInMemoryStore(),
		NewMultiIndexedArrayInMemoryStore(),
		NewConcurrentFactStore(NewSimpleInMemoryStore()),
		NewColumnarStore(),
	} {
		for i := 0; i < 1000000; i++ {
			p := ast.PredicateSym{Symbol: "p", Arity: 3}
			c1 := ast.String(fmt.Sprintf("c%d", i%2)) // 2 distinct values.
			c2 := ast.String(fmt.Sprintf("c%d", i))
			c3 := ast.String(fmt.Sprintf("c%d", i))
			store.Add(ast.Atom{p, []ast.BaseTerm{c1, c2, c3}})
		}

		b.Run(fmt.Sprintf("%T", store), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				p := ast.PredicateSym{Symbol: "p", Arity: 3}
				c1 := ast.String("c0")
				store.GetFacts(ast.Atom{p, []ast.BaseTerm{c1, ast.Variable{"_"}, ast.Variable{"_"}}}, func(a ast.Atom) error {
					return nil
				})
			}
		})
	}
}

func BenchmarkMerge(b *testing.B) {
	sourceStore := NewSimpleInMemoryStore()
	mg := NewColumnarStore()
	for i := 0; i < 10000; i++ {
		p := ast.PredicateSym{Symbol: fmt.Sprintf("p%d", i%20), Arity: 2}
		c1 := ast.String(fmt.Sprintf("c%d", i%200))
		c2 := ast.String(fmt.Sprintf("c%d", i%200))
		sourceStore.Add(ast.Atom{p, []ast.BaseTerm{c1, c2}})
	}

	for i := 0; i < 10000; i++ {
		p := ast.PredicateSym{Symbol: fmt.Sprintf("p%d", i%20), Arity: 2}
		c1 := ast.String(fmt.Sprintf("c%d", i%200))
		c2 := ast.String(fmt.Sprintf("c%d", i%200))
		mg.Add(ast.Atom{p, []ast.BaseTerm{c1, c2}})
	}

	benchmarks := []struct {
		name      string
		storeFunc func() FactStore
	}{
		{
			name:      "SimpleInMemoryStore",
			storeFunc: func() FactStore { return NewSimpleInMemoryStore() },
		},
		{
			name:      "ColumnarFactStore",
			storeFunc: func() FactStore { return NewColumnarStore() },
		},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			// This inner loop is what's timed.
			for i := 0; i < b.N; i++ {
				// We must exclude the creation of the destination store from the benchmark time.
				b.StopTimer()
				destStore := bm.storeFunc()
				b.StartTimer()
				destStore.Merge(mg)
			}
		})
	}
}