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
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/mangle/ast"
	"github.com/google/mangle/parse"
)

// SimpleColumn is a file format to store facts.
type SimpleColumn struct {
}

// ErrCouldNotRead is an error.
var ErrCouldNotRead = errors.New("could not read file")

// WriteTo writes contents of a fact store to writer.
func (SimpleColumn) WriteTo(store ReadOnlyFactStore, w io.Writer) error {
	preds := store.ListPredicates()
	if _, err := fmt.Fprintf(w, "%d\n", len(preds)); err != nil {
		return err
	}
	for _, p := range preds {
		if _, err := fmt.Fprintf(w, "%s %d\n", p.Symbol, p.Arity); err != nil {
			return err
		}
	}
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
		if _, err := fmt.Fprintf(w, "%d\n", len(facts)); err != nil {
			return err
		}
		for i := 0; i < p.Arity; i++ {
			for _, f := range facts {
				if _, err := fmt.Fprint(w, percentEscape(f.Args[i].String())); err != nil {
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

// ReadInto reads contents in simplecolumn format into a fact store.
func (SimpleColumn) ReadInto(r io.Reader, store FactStore) error {
	scanner := bufio.NewScanner(r)
	if ok := scanner.Scan(); !ok {
		return ErrCouldNotRead
	}
	numPreds, err := strconv.Atoi(scanner.Text())
	if err != nil {
		return ErrCouldNotRead
	}
	preds := make([]ast.PredicateSym, numPreds)
	for i := 0; i < numPreds; i++ {
		if ok := scanner.Scan(); !ok {
			return ErrCouldNotRead
		}
		var name string
		var arity int
		if _, err = fmt.Sscanf(scanner.Text(), "%s %d", &name, &arity); err != nil {
			return ErrCouldNotRead
		}
		preds[i] = ast.PredicateSym{name, arity}
	}
	for _, p := range preds {
		if p.Arity == 0 {
			store.Add(ast.Atom{p, nil})
			continue
		}
		var numFacts int
		if ok := scanner.Scan(); !ok {
			return ErrCouldNotRead
		}
		if _, err = fmt.Sscanf(scanner.Text(), "%d", &numFacts); err != nil {
			return ErrCouldNotRead
		}
		args := make([][]ast.BaseTerm, numFacts)
		for i := 0; i < numFacts; i++ {
			args[i] = make([]ast.BaseTerm, p.Arity)
		}
		for j := 0; j < p.Arity; j++ {
			for i := 0; i < numFacts; i++ {
				if ok := scanner.Scan(); !ok {
					return ErrCouldNotRead
				}
				s, err := percentUnescape(scanner.Text())
				if err != nil {
					return ErrCouldNotRead
				}
				c, err := parse.BaseTerm(s)
				if err != nil {
					return ErrCouldNotRead
				}
				args[i][j] = c
			}
		}
		for i := 0; i < numFacts; i++ {
			store.Add(ast.Atom{p, args[i]})
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
