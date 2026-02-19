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

package factstore

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/functional"
	"codeberg.org/TauCeti/mangle-go/parse"
)

const (
	// Limit on the number of predicates that are stored together.
	maxNumPreds = 1 << 16
	// Limit on the number of facts for a given predicate.
	maxFactsPerPredicate = 1 << 32
	// Limit on the number of arguments a predicate can take.
	maxArity = 1 << 10
)

// SimpleColumn is a file format to store a knowledge base.
//
// The simplecolumm format is meant to provide a good enough, "batteries included"
// library for storing facts.  SimpleColumn enforces the following limits:
// - not more than 2^16 predicates
// - not more than 2^32 facts per predicate
// - predicate cannot have more than 2^10 arguments
//
// The SimpleColumnStore implements the ReadOnlyFactStore interface however it does
// not provide any form for indexing. Large datasets that are frequently accessed would
// benefit from a different store implementation or splitting up into multiple files.
//
// For a list of predicates p_1 ... p_n, the format is as follows:
// line 1.     <number of predicates>
// line 1 + i. <predicate #i name> <arity> <num facts>   // i \in {1, n}
//
// For each predicate p_i \in {1, n} that has arity > 0 and num facts > 0:
//
//		let m be number of p_i facts
//		For each argument  j \in {1, arity(p_i)} ("column"):
//	    let h be the number of preceding lines: 1 + n + /preceding predicates/ + m * (j-1)
//		  For each fact p_i(x_1...x_arity) with index k \in {1, m} :
//
// line h + k: <serialized argument x_j for fact k>
//
// Constants are base64 encoded so the file can be opened in text editor,
// sent over all sorts of networks, put into JSON etc.
// TODO: fully specify serialization format of constants for forward compatibility.
type SimpleColumn struct {
	// If true, write methods ensure deterministic output.
	// When reading from a factstore, the order of facts is not guaranteed
	// to be deterministic. This setting enables sorting of the facts
	// according to order determined by fact hashes.
	// If a fact store implementation returns facts in a deterministic order,
	// then it is not necessary to enable this.
	Deterministic bool
}

// SimpleColumnStore is a read-only fact store backed by a simple column file.
type SimpleColumnStore struct {
	input              func() (io.ReadCloser, error)
	predicates         []ast.PredicateSym
	predicateFactCount []int
}

var _ ReadOnlyFactStore = (*SimpleColumnStore)(nil)

// ListPredicates implements a ReadOnlyFactStore method.
func (s *SimpleColumnStore) ListPredicates() []ast.PredicateSym {
	return s.predicates
}

// Contains implements a ReadOnlyFactStore method.
func (s *SimpleColumnStore) Contains(fact ast.Atom) bool {
	var found bool
	s.GetFacts(fact, func(a ast.Atom) error {
		found = true
		return nil
	})
	return found
}

// EstimateFactCount implements a ReadOnlyFactStore method.
func (s *SimpleColumnStore) EstimateFactCount() int {
	var numFacts int
	for _, c := range s.predicateFactCount {
		numFacts += c
	}
	return numFacts
}

// GetFacts implements a ReadOnlyFactStore method.
func (s *SimpleColumnStore) GetFacts(query ast.Atom, cb func(ast.Atom) error) error {
	pred := query.Predicate
	numFacts := 0
	toSkip := 1 + len(s.predicates)
	for i, p := range s.predicates {
		if p == pred {
			if p.Arity == 0 { // Special case 0-arity predicates, we are done.
				if s.predicateFactCount[i] > 0 {
					cb(ast.Atom{pred, nil})
				}
				return nil
			}
			numFacts = s.predicateFactCount[i]
			if numFacts == 0 {
				return nil // Special case empty set of facts.
			}
			break
		}
		if p.Arity == 0 {
			continue
		}
		toSkip += s.predicateFactCount[i] * p.Arity
	}

	f, inputErr := s.input()
	if inputErr != nil {
		return ErrCouldNotRead
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for i := 0; i < toSkip; i++ {
		if ok := scanner.Scan(); !ok {
			return ErrCouldNotRead
		}
	}

	var sc SimpleColumn
	if err := sc.readPred(scanner, pred, numFacts, query.Args, func(args []ast.BaseTerm) error {
		if err := cb(ast.Atom{pred, args}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

var (
	// ErrCouldNotRead is returned when the input file could not be read.
	ErrCouldNotRead = errors.New("could not read file")

	// ErrNotSupported is returned when ReadPred is called for 0-arity predicate.
	ErrNotSupported = errors.New("not supported")

	// ErrWrongArgument is returned when an argument does not make sense.
	ErrWrongArgument = errors.New("wrong argument")

	// ErrTooManyPreds is returned when a store has too many predicates.
	// The limit is unreasonably large and clients should shard their storage.
	ErrTooManyPreds = errors.New("too many preds")

	// ErrTooManyFacts is returned when a store has too many facts for a predicate.
	// The limit is unreasonably large and clients should implement a custom FactStore.
	ErrTooManyFacts = errors.New("too many facts")

	// ErrUnsupportedArity is returned when a predicate has too many arguments.
	// The limit is unreasonably large and clients should organize data differently or come
	// up with a different storage.
	ErrUnsupportedArity = errors.New("unsupported arity")
)

// NewSimpleColumnStoreFromBytes returns a new fact store backed by data in simplecolumn format.
func NewSimpleColumnStoreFromBytes(data []byte) (*SimpleColumnStore, error) {
	return NewSimpleColumnStore(func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	})
}

// NewSimpleColumnStoreFromGzipBytes returns a new fact store backed by data that is
// a gzipped file in simplecolumn format.
func NewSimpleColumnStoreFromGzipBytes(data []byte) (*SimpleColumnStore, error) {
	return NewSimpleColumnStore(func() (io.ReadCloser, error) {
		reader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		return reader, nil
	})
}

// NewSimpleColumnStore returns a new fact store backed by a simple column file.
// The input closure is called immediately to parse the header.
func NewSimpleColumnStore(input func() (io.ReadCloser, error)) (*SimpleColumnStore, error) {
	f, err := input()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	preds, predFactCount, err := readHeader(scanner)
	if err != nil {
		return nil, err
	}
	return &SimpleColumnStore{input, preds, predFactCount}, nil
}

func (sc SimpleColumn) writeHeader(preds []ast.PredicateSym, predFactCount []int, w io.Writer) error {
	// line 1. <number of predicates>
	if _, err := fmt.Fprintf(w, "%d\n", len(preds)); err != nil {
		return err
	}
	// line 1 + i. <predicate #i name> <arity> <num facts>
	for i, p := range preds {
		if _, err := fmt.Fprintf(w, "%s %d %d\n", p.Symbol, p.Arity, predFactCount[i]); err != nil {
			return err
		}
	}
	return nil
}

// WriteTo writes contents of a fact store to the writer. It only
// calls the ListPredicates and GetFacts methods on the store.
// Flushing or closing the writer is is the caller's responsibility.
func (sc SimpleColumn) WriteTo(store ReadOnlyFactStore, w io.Writer) error {
	preds := store.ListPredicates()
	if len(preds) > maxNumPreds {
		return ErrTooManyPreds
	}
	if sc.Deterministic {
		sort.Slice(preds, func(i, j int) bool {
			a := preds[i]
			b := preds[j]
			return a.Arity < b.Arity || a.Arity == b.Arity && a.Symbol < b.Symbol
		})
	}
	predFactCount := make([]int, len(preds))
	for i, p := range preds {
		if p.Arity > maxArity {
			return fmt.Errorf("pred %v: %w", p, ErrUnsupportedArity)
		}
		var numFacts int
		if err := store.GetFacts(ast.NewQuery(p), func(a ast.Atom) error {
			numFacts++
			return nil
		}); err != nil {
			return err
		}
		if numFacts > maxFactsPerPredicate {
			return fmt.Errorf("pred %v: %w", p, ErrTooManyFacts)
		}
		predFactCount[i] = numFacts
	}
	if err := sc.writeHeader(preds, predFactCount, w); err != nil {
		return err
	}
	// for each predicate p with arity > 0:
	for _, p := range preds {
		if p.Arity == 0 {
			continue
		}
		var facts []ast.Atom
		if err := store.GetFacts(ast.NewQuery(p), func(a ast.Atom) error {
			facts = append(facts, a)
			return nil
		}); err != nil {
			return err
		}
		if sc.Deterministic {
			sort.Slice(facts, func(i, j int) bool {
				h1, h2 := facts[i].Hash(), facts[j].Hash()
				if h1 == h2 {
					return facts[i].String() < facts[j].String()
				}
				return h1 < h2
			})
		}
		for i := 0; i < p.Arity; i++ {
			for _, f := range facts {
				if len(f.Args) != p.Arity {
					return fmt.Errorf("malformed fact: %v predicate arity %d: %w", f, p.Arity, ErrWrongArgument)
				}
				// line h + k: <column: argument x_j for fact k>
				if _, err := fmt.Fprint(w, f.Args[i].String()); err != nil {
					return err
				}
				if _, err := fmt.Fprintln(w); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// ReadPred reads matching facts for a single predicate with arity > 0.
// len(filter) must match p.Arity.
func (SimpleColumn) readPred(scanner *bufio.Scanner, p ast.PredicateSym, numFacts int, filter []ast.BaseTerm, cb func(args []ast.BaseTerm) error) error {
	args := make([][]ast.BaseTerm, numFacts)
	numSkip := 0
	skip := make([]bool, numFacts)
	for i := 0; i < numFacts; i++ {
		args[i] = make([]ast.BaseTerm, p.Arity)
	}
	// TODO: It would be smarter to load and traverse those columns that
	// have a filter present.
	for j := 0; j < p.Arity; j++ {
		for i := 0; i < numFacts; i++ {
			if ok := scanner.Scan(); !ok {
				return fmt.Errorf("scanning pred %v column %d fact %d: %w", p, j, i, ErrCouldNotRead)
			}
			if skip[i] { // Fact does not match anyway.
				continue
			}
			text := scanner.Text()
			if text[0] == '/' {
				var err error
				text, err = percentUnescape(text)
				if err != nil {
					return fmt.Errorf("unescape failed pred %v column %d fact %d: %w", p, j, i, ErrCouldNotRead)
				}
			}
			e, err := parse.BaseTerm(text)
			if err != nil {
				return fmt.Errorf("parsing failed pred %v column %d fact %d: %w", p, j, i, err)
			}

			c, err := functional.EvalExpr(e, nil)

			if err != nil {
				return fmt.Errorf("evaluating failed pred %v column %d fact %d: %w", p, j, i, ErrCouldNotRead)
			}
			args[i][j] = c
			if filter == nil {
				continue
			}
			want, ok := filter[j].(ast.Constant)
			if !ok {
				continue
			}
			skip[i] = !want.Equals(c)
			numSkip++
		}
	}
	for i := 0; i < numFacts; i++ {
		if skip[i] {
			continue
		}
		if err := cb(args[i]); err != nil {
			return err
		}
	}
	return nil
}

func readHeader(scanner *bufio.Scanner) ([]ast.PredicateSym, []int, error) {
	if ok := scanner.Scan(); !ok {
		return nil, nil, ErrCouldNotRead
	}
	numPreds, err := strconv.Atoi(scanner.Text())
	if err != nil {
		return nil, nil, ErrCouldNotRead
	}
	if numPreds < 0 {
		return nil, nil, fmt.Errorf("invalid number of predicates %d: %w", numPreds, ErrWrongArgument)
	}
	if numPreds > maxNumPreds {
		return nil, nil, ErrTooManyPreds
	}
	preds := make([]ast.PredicateSym, numPreds)
	predNumFacts := make([]int, numPreds)
	for i := 0; i < numPreds; i++ {
		if ok := scanner.Scan(); !ok {
			return nil, nil, ErrCouldNotRead
		}
		var (
			name     string
			arity    int
			numFacts int
		)
		if _, err = fmt.Sscanf(scanner.Text(), "%s %d %d", &name, &arity, &numFacts); err != nil {
			return nil, nil, ErrCouldNotRead
		}
		if _, err := parse.PredicateName(name); err != nil {
			return nil, nil, fmt.Errorf("invalid name %q predicate %d: %w", name, i, ErrWrongArgument)
		}
		if arity < 0 || arity > maxArity {
			return nil, nil, fmt.Errorf("for predicate %v: %w", name, ErrUnsupportedArity)
		}
		if numFacts > maxFactsPerPredicate {
			return nil, nil, fmt.Errorf("for predicate %v: %w", name, ErrTooManyFacts)
		}
		preds[i] = ast.PredicateSym{name, arity}
		predNumFacts[i] = numFacts
	}
	return preds, predNumFacts, nil
}

// ReadInto reads contents in simplecolumn format into a fact store.
func (sc SimpleColumn) ReadInto(r io.Reader, store FactStore) error {
	scanner := bufio.NewScanner(r)

	preds, predNumFacts, err := readHeader(scanner)
	if err != nil {
		return err
	}
	for i, p := range preds {
		if p.Arity == 0 {
			store.Add(ast.Atom{p, nil})
			continue
		}
		numFacts := predNumFacts[i]
		if err = sc.readPred(scanner, p, numFacts, nil, func(args []ast.BaseTerm) error {
			store.Add(ast.Atom{p, args})
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

// percentEscape escapes a string following RFC 3986.
func percentEscape(s string) string {
	// Note that url.QueryEscape() insists on replacing " " by "+" instead of "%20".
	return strings.Replace(url.QueryEscape(s), "+", "%20", -1)
}

// percentUnescape unescapes a string encoded with percentEscape
func percentUnescape(s string) (string, error) {
	return url.QueryUnescape(s)
}
