// Copyright 2026 The Mangle Authors
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

package engine

import (
	"fmt"
	"strings"
	"testing"

	"codeberg.org/TauCeti/mangle-go/analysis"
	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/factstore"
	"codeberg.org/TauCeti/mangle-go/parse"
)

// A benchmark program: rules (with Decls for the base predicates), and the
// base facts, which are made up.
type reorderBenchCase struct {
	name  string
	rules string
	facts func() []ast.Atom
}

func num(n int) ast.BaseTerm { return ast.Number(int64(n)) }

func reorderBenchCases() []reorderBenchCase {
	const n = 300
	return []reorderBenchCase{{
		// Two premises that share no variable, then the premise that joins them.
		name: "cross-product",
		rules: `Decl a(X). Decl b(Y). Decl link(X, Y).
			r(X, Y) :- a(X), b(Y), link(X, Y).`,
		facts: func() []ast.Atom {
			var facts []ast.Atom
			for i := 0; i < n; i++ {
				facts = append(facts,
					ast.NewAtom("a", num(i)), ast.NewAtom("b", num(i)),
					ast.NewAtom("link", num(i), num((i*7)%n)))
			}
			return facts
		},
	}, {
		// A negation written last that removes nine in ten rows.
		name: "late-negation",
		rules: `Decl a(X). Decl b(X, Y). Decl c(Y, Z). Decl excluded(X).
			s(X, Z) :- a(X), b(X, Y), c(Y, Z), !excluded(X).`,
		facts: func() []ast.Atom {
			var facts []ast.Atom
			for i := 0; i < n; i++ {
				facts = append(facts, ast.NewAtom("a", num(i)))
				if i%10 != 0 {
					facts = append(facts, ast.NewAtom("excluded", num(i)))
				}
				for j := 0; j < 5; j++ {
					facts = append(facts, ast.NewAtom("b", num(i), num(i*5+j)))
					facts = append(facts, ast.NewAtom("c", num(i*5+j), num(j)))
				}
			}
			return facts
		},
	}, {
		// A negation over a big relation that removes nothing, written after
		// a premise that keeps few rows. Moved ahead, it would run on every
		// row instead, which is much slower in a store without indexes.
		name: "negation-removes-little",
		rules: `Decl a(X). Decl sel(X, Y). Decl bad(X).
			s(X, Y) :- a(X), sel(X, Y), !bad(X).`,
		facts: func() []ast.Atom {
			var facts []ast.Atom
			for i := 0; i < 1000; i++ {
				facts = append(facts, ast.NewAtom("a", num(i)))
			}
			for i := 0; i < 10; i++ {
				facts = append(facts, ast.NewAtom("sel", num(i*100), num(i)))
			}
			for i := 0; i < 20000; i++ {
				facts = append(facts, ast.NewAtom("bad", num(100000+i)))
			}
			return facts
		},
	}, {
		// Paths in a graph, with the recursive premise written second.
		name: "recursion",
		rules: `Decl edge(X, Y).
			path(X, Y) :- edge(X, Y).
			path(X, Z) :- edge(Y, Z), path(X, Y).`,
		facts: func() []ast.Atom {
			var facts []ast.Atom
			for i := 0; i < 60; i++ {
				facts = append(facts, ast.NewAtom("edge", num(i), num(i+1)))
				if i%3 == 0 {
					facts = append(facts, ast.NewAtom("edge", num(i), num(i+2)))
				}
			}
			return facts
		},
	}, {
		// A rule already written in a good order: reordering changes nothing.
		name: "well-ordered",
		rules: `Decl a(X). Decl b(X, Y). Decl c(Y, Z).
			t(X, Z) :- a(X), b(X, Y), c(Y, Z).`,
		facts: func() []ast.Atom {
			var facts []ast.Atom
			for i := 0; i < n; i++ {
				facts = append(facts, ast.NewAtom("a", num(i)))
				for j := 0; j < 5; j++ {
					facts = append(facts, ast.NewAtom("b", num(i), num(i*5+j)))
					facts = append(facts, ast.NewAtom("c", num(i*5+j), num(j)))
				}
			}
			return facts
		},
	}}
}

func benchProgram(b *testing.B, rules string) *analysis.ProgramInfo {
	b.Helper()
	unit, err := parse.Unit(strings.NewReader(rules))
	if err != nil {
		b.Fatal(err)
	}
	pi, err := analysis.AnalyzeOneUnit(unit, nil)
	if err != nil {
		b.Fatal(err)
	}
	return pi
}

func BenchmarkPremiseReordering(b *testing.B) {
	stores := []struct {
		name string
		new  func() factstore.FactStore
	}{
		{"simple-store", func() factstore.FactStore { return factstore.NewSimpleInMemoryStore() }},
		{"indexed-store", func() factstore.FactStore { return factstore.NewMultiIndexedArrayInMemoryStore() }},
	}
	for _, c := range reorderBenchCases() {
		pi := benchProgram(b, c.rules)
		facts := c.facts()
		for _, s := range stores {
			for _, reorder := range []bool{false, true} {
				order := "written"
				var opts []EvalOption
				if reorder {
					order, opts = "reordered", []EvalOption{WithPremiseReordering()}
				}
				b.Run(fmt.Sprintf("%s/%s/%s", c.name, s.name, order), func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						b.StopTimer()
						store := s.new()
						for _, f := range facts {
							store.Add(f)
						}
						b.StartTimer()
						if _, err := EvalProgramWithStats(pi, store, opts...); err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		}
	}
}
