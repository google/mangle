// Copyright 2026 Google LLC
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

// Command mgwhy explains why a fact was derived by a Mangle program.
//
// Usage:
//
//	mgwhy -program PROG.mg [-facts STORE.sc[.gz|.zst]]
//	      [-mode simple|full] [-format tree|facts]
//	      [-max-proofs N] [-max-depth N] GOAL
//
// The program is loaded and evaluated against the fact store (if any); then
// the goal atom is resolved to one or more derivations and printed as either
// an indented proof tree (default) or a set of Mangle facts matching the
// provenance schema (proves, uses_rule, premise, edb_leaf, binding,
// rule_source).
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"codeberg.org/TauCeti/mangle-go/analysis"
	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/engine"
	"codeberg.org/TauCeti/mangle-go/factstore"
	"codeberg.org/TauCeti/mangle-go/parse"
	"codeberg.org/TauCeti/mangle-go/provenance"
)

func main() {
	var (
		programPath = flag.String("program", "", "Mangle program source file (required)")
		factsPath   = flag.String("facts", "", "simplecolumn factstore file (optional; auto-detects .gz / .zst)")
		mode        = flag.String("mode", "simple", "provenance mode: simple (post-hoc) or full (engine-recorded, handles aggregation)")
		format      = flag.String("format", "tree", "output format: tree or facts")
		maxProofs   = flag.Int("max-proofs", 1, "maximum number of alternative proofs to return")
		maxDepth    = flag.Int("max-depth", 64, "maximum proof depth")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: mgwhy -program PROG.mg [-facts STORE.sc] [-mode simple|full] [-format tree|facts] GOAL\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *programPath == "" || flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	if *format != "tree" && *format != "facts" {
		fmt.Fprintf(os.Stderr, "mgwhy: -format must be \"tree\" or \"facts\"\n")
		os.Exit(2)
	}
	if *mode != "simple" && *mode != "full" {
		fmt.Fprintf(os.Stderr, "mgwhy: -mode must be \"simple\" or \"full\"\n")
		os.Exit(2)
	}
	goalStr := flag.Arg(0)
	if err := run(*programPath, *factsPath, goalStr, *mode, *format, *maxProofs, *maxDepth); err != nil {
		fmt.Fprintf(os.Stderr, "mgwhy: %v\n", err)
		os.Exit(1)
	}
}

func run(programPath, factsPath, goalStr, mode, format string, maxProofs, maxDepth int) error {
	pf, err := os.Open(programPath)
	if err != nil {
		return fmt.Errorf("open program: %w", err)
	}
	defer pf.Close()
	unit, err := parse.Unit(pf)
	if err != nil {
		return fmt.Errorf("parse program: %w", err)
	}

	var store factstore.FactStore
	if factsPath != "" {
		loaded, err := loadFactStore(factsPath)
		if err != nil {
			return fmt.Errorf("load facts: %w", err)
		}
		s := factstore.NewSimpleInMemoryStore()
		if err := loaded.GetFacts(ast.NewAtom("_dummy"), func(a ast.Atom) error {
			s.Add(a)
			return nil
		}); err != nil {
			// Fallback: enumerate per predicate.
		}
		for _, p := range loaded.ListPredicates() {
			loaded.GetFacts(ast.NewQuery(p), func(a ast.Atom) error {
				s.Add(a)
				return nil
			})
		}
		store = &s
	} else {
		s := factstore.NewSimpleInMemoryStore()
		store = &s
	}

	knownPreds := make(map[ast.PredicateSym]ast.Decl)
	for _, p := range store.ListPredicates() {
		knownPreds[p] = ast.NewSyntheticDeclFromSym(p)
	}
	pi, err := analysis.AnalyzeOneUnit(unit, knownPreds)
	if err != nil {
		return fmt.Errorf("analyze: %w", err)
	}
	for _, f := range pi.InitialFacts {
		store.Add(f)
	}
	strata, predToStratum, err := analysis.Stratify(analysis.Program{
		EdbPredicates: pi.EdbPredicates,
		IdbPredicates: pi.IdbPredicates,
		Rules:         pi.Rules,
	})
	if err != nil {
		return fmt.Errorf("stratify: %w", err)
	}
	var recorder *provenance.MemoryRecorder
	var evalOpts []engine.EvalOption
	if mode == "full" {
		recorder = provenance.NewMemoryRecorder()
		evalOpts = append(evalOpts, engine.WithDerivationRecorder(recorder))
	}
	if _, err := engine.EvalStratifiedProgramWithStats(pi, strata, predToStratum, store, evalOpts...); err != nil {
		return fmt.Errorf("eval: %w", err)
	}

	goal, err := parse.Atom(goalStr)
	if err != nil {
		return fmt.Errorf("parse goal: %w", err)
	}

	opts := provenance.Options{MaxProofs: maxProofs, MaxDepth: maxDepth}
	var proofs []*provenance.ProofNode
	if mode == "full" {
		proofs, err = provenance.BuildFromRecording(recorder, store, goal, opts)
	} else {
		proofs, err = provenance.Explain(pi, store, goal, opts)
	}
	if err != nil {
		return err
	}

	switch format {
	case "tree":
		return provenance.Print(os.Stdout, proofs)
	case "facts":
		out := factstore.NewSimpleInMemoryStore()
		if err := provenance.EmitFacts(proofs, &out); err != nil {
			return err
		}
		return printFactsAsMangleSource(os.Stdout, &out)
	}
	return nil
}

func loadFactStore(path string) (factstore.ReadOnlyFactStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	switch {
	case strings.HasSuffix(path, ".gz"):
		return factstore.NewSimpleColumnStoreFromGzipBytes(data)
	case strings.HasSuffix(path, ".zst"), strings.HasSuffix(path, ".zstd"):
		return factstore.NewSimpleColumnStoreFromZstdBytes(data)
	default:
		return factstore.NewSimpleColumnStoreFromBytes(data)
	}
}

// printFactsAsMangleSource writes every fact in the store as "pred(args).\n",
// suitable for piping back into `mg` or loading with parse.Unit.
func printFactsAsMangleSource(w *os.File, store factstore.ReadOnlyFactStore) error {
	for _, p := range store.ListPredicates() {
		err := store.GetFacts(ast.NewQuery(p), func(a ast.Atom) error {
			_, err := fmt.Fprintf(w, "%s.\n", a.String())
			return err
		})
		if err != nil {
			return err
		}
	}
	return nil
}
