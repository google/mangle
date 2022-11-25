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

// Package interpreter provides functions for an interactive interpreter.
package interpreter

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chzyer/readline"
	"github.com/google/mangle/analysis"
	"github.com/google/mangle/ast"
	"github.com/google/mangle/engine"
	"github.com/google/mangle/factstore"
	"github.com/google/mangle/parse"
)

const bufSize = 4096

type sourceFragment struct {
	units      []parse.SourceUnit
	program    *analysis.ProgramInfo
	checkpoint factstore.FactStore
}

// Interpreter is an interactive interpreter.
type Interpreter struct {
	out   io.Writer
	root  string
	store factstore.FactStore
	// List of source paths that were loaded, in the order they were loaded.
	src []string
	// Maps source path to source fragment.
	sourceFragments map[string]*sourceFragment
	// Collects all declarations from source fragments.
	knownPredicates map[ast.PredicateSym]ast.Decl
	// For rules added interactively. If this is non-empty,
	// then src contains an "interactive" path as last element.
	// and sourceFragments contains an "interactive" entry.
	buffer string
	// Predicates for which we display stats.
	stats []string
	// post processors to run after evaluation.
	postProcessors []func(store factstore.FactStore) error
}

// New returns a new interpreter.
func New(out io.Writer, root string, stats []string) *Interpreter {
	return &Interpreter{
		out:             out,
		root:            root,
		store:           factstore.NewSimpleInMemoryStore(),
		src:             nil,
		sourceFragments: make(map[string]*sourceFragment),
		knownPredicates: map[ast.PredicateSym]ast.Decl{},
		stats:           stats,
		postProcessors:  nil,
	}
}

// AddPostProcessor adds a post processing function that is called after evaluation.
func (i *Interpreter) AddPostProcessor(postProcessor func(store factstore.FactStore) error) {
	i.postProcessors = append(i.postProcessors, postProcessor)
}

const (
	normalPrompt    = "mr >"
	continuedPrompt = ".. >"
	interactivePath = "interactive-buffer"
	preloadPath     = "interactive-preload"
)

func nextLine() (string, error) {
	return nextLineWithPrompt(normalPrompt)
}

func nextLineWithPrompt(prompt string) (string, error) {
	rl, err := readline.New(prompt)
	if err != nil {
		return "", err
	}
	line, err := rl.Readline()
	if err != nil {
		return "", err
	}
	readline.AddHistory(line)
	return strings.TrimSpace(line), nil
}

func (i *Interpreter) showPredicate(p ast.PredicateSym) {
	const docStringMargin = 50
	decl := i.knownPredicates[p]
	atomText := decl.DeclaredAtom.String()
	if docLines := decl.Doc(); len(docLines) != 0 && docLines[0] != "" {
		prefixLen := 0
		if len(atomText) < docStringMargin {
			prefixLen = len(atomText)
		}
		spacer := strings.Repeat(" ", docStringMargin-prefixLen)
		if len(atomText) < docStringMargin {
			fmt.Printf("%s%s%s\n", atomText, spacer, docLines[0])
		} else {
			fmt.Printf("%s\n%s%s\n", atomText, spacer, docLines[0])
		}
	} else {
		fmt.Printf("%s\n", atomText)
	}
}

// Show shows information about predicates.
// If arg = "all", it lists all predicate.
// Otherwise, it shows information about the predicate named arg.
func (i *Interpreter) Show(arg string) error {
	if arg == "all" {
		var preds []ast.PredicateSym
		for sym := range i.knownPredicates {
			preds = append(preds, sym)
		}
		sort.Slice(preds, func(i, j int) bool {
			return preds[i].Symbol < preds[j].Symbol
		})
		for _, p := range preds {
			i.showPredicate(p)
		}

		return nil
	}
	var matches []string
	for sym := range i.knownPredicates {
		if sym.Symbol == arg {
			i.showPredicate(sym)
			return nil
		}
		if strings.HasPrefix(sym.Symbol, arg) {
			matches = append(matches, sym.Symbol)
		}
	}

	if len(matches) != 0 {
		return fmt.Errorf("predicate %s not found, did you mean %v", arg, matches)
	}
	return fmt.Errorf("predicate %s not found", arg)
}

// Load loads source file at path.
func (i *Interpreter) Load(path string) error {
	i.resetInteractiveDefs("")
	f, err := os.Open(filepath.Join(i.root, path))
	if err != nil {
		return err
	}
	defer f.Close()
	r := bufio.NewReaderSize(f, bufSize)
	unit, err := parse.Unit(r)
	if err != nil {
		return err
	}
	programInfo, err := analysis.AnalyzeOneUnit(unit, i.knownPredicates)
	if err != nil {
		return err
	}
	i.pushSourceFragment(path, []parse.SourceUnit{unit}, programInfo)

	fmt.Fprintf(i.out, "loaded %s.\n", path)
	return i.evalProgram(programInfo)
}

// ParseQuery parses a query string. It can either be a predicate name,
// or an actual atom with constants, variables and wildcards.
func (i *Interpreter) ParseQuery(query string) (ast.Atom, error) {
	var (
		atom ast.Atom
		err  error
	)
	if strings.Contains(query, "(") {
		atom, err = parse.Atom(query)
	} else {
		err = fmt.Errorf("predicate %s not found", query)
		for sym := range i.knownPredicates {
			if sym.Symbol == query {
				atom = ast.NewQuery(sym)
				err = nil
				break
			}
		}
	}
	if err != nil {
		return ast.Atom{}, err
	}
	return atom, nil
}

// Query queries the interpreter's state.
func (i *Interpreter) Query(query ast.Atom) ([]ast.Atom, error) {
	var results []ast.Atom
	i.store.GetFacts(query, func(a ast.Atom) error {
		results = append(results, a)
		return nil
	})
	return results, nil
}

// QueryInteractive parses query string, queries the interpreter's state, returns
// results formatted as strings.
func (i *Interpreter) QueryInteractive(queryString string) error {
	query, err := i.ParseQuery(queryString)
	if err != nil {
		return err
	}
	facts, err := i.Query(query)
	if err != nil {
		return err
	}
	var results []string
	for _, fact := range facts {
		results = append(results, fact.String())
	}
	sort.Strings(results)
	fmt.Fprintf(i.out, "%s\n", strings.Join(results, "\n"))
	if len(results) == 0 {
		fmt.Fprintf(i.out, "No entries for %s.\n", query)
	} else {
		fmt.Fprintf(i.out, "Found %d entries for %s.\n", len(results), query)
	}
	return nil
}

// Define adds rule definitions the interpreter's state.
func (i *Interpreter) Define(clauseText string) error {
	// TODO: A nice idea would be to work with the parsed form,
	// like supporting removal of a particular clause. This would
	// require retracting the associated facts from the store, though,
	// which is currently not supported.
	buffer := i.buffer + clauseText
	unit, err := parse.Unit(strings.NewReader(buffer))
	if err != nil {
		return fmt.Errorf("parsing failed: %v", err)
	}
	i.resetInteractiveDefs(buffer)
	programInfo, err := analysis.AnalyzeOneUnit(unit, i.knownPredicates)
	if err != nil {
		return fmt.Errorf("analysis failed: %v", err)
	}
	i.pushSourceFragment(interactivePath, []parse.SourceUnit{unit}, programInfo)
	// We run evaluation every time a line is added. Alternatively, we could
	// let the user control when to evaluate rules.
	err = i.evalProgram(programInfo)
	if err != nil {
		return fmt.Errorf("evaluation failed: %v", err)
	}
	var preds []ast.PredicateSym
	for sym := range programInfo.Decls {
		preds = append(preds, sym)
	}
	fmt.Fprintf(i.out, "defined %s.\n", preds)
	return nil
}

// Pop resets the interpreter's state to what it was before the last change.
// Definitions entered interactively are always considered the last change.
func (i *Interpreter) Pop() {
	if len(i.src) <= 0 {
		fmt.Fprintln(i.out, "nothing to pop.")
		return
	}
	if i.hasInteractiveDefs() {
		i.resetInteractiveDefs("")
		fmt.Fprintln(i.out, "popped interactive defs.")
		return
	}
	fmt.Fprintf(i.out, "popped %q.\n", i.src[len(i.src)-1])
	i.popSourceFragment()
}

// ShowHelp displays help text.
func (i *Interpreter) ShowHelp() {
	fmt.Fprintln(i.out, `
<decl>.            adds declaration to interactive buffer
<clause>.          adds clause to interactive buffer, evaluates.
?<predicate>       looks up predicate name and queries all facts
?<goal>            queries all facts that match goal
::load <path>      pops interactive buffer and loads source file at <path>
::help             display this help text
::pop              reset state to before interactive defs. or last load command
::show <predicate> shows information about predicate
::show all         shows information about all available predicates
<Ctrl-D>           quit`)
}

// Loop reads lines from stdin and performs the commands.
func (i *Interpreter) Loop() error {
	i.ShowHelp()
	for {
		line, err := nextLine()
		if err != nil {
			return err
		}
		switch {
		case line == "::help":
			i.ShowHelp()

		case strings.HasPrefix(line, "::load "):
			if err := i.Load(strings.TrimPrefix(line, "::load ")); err != nil {
				fmt.Fprintf(i.out, "load failed: %v\n", err)
			}

		case line == "::pop":
			i.Pop()

		case strings.HasPrefix(line, "::show "):
			if err := i.Show(strings.TrimPrefix(line, "::show ")); err != nil {
				fmt.Fprintf(i.out, "show failed: %v\n", err)
			}

		case strings.HasPrefix(line, "?"):
			if err := i.QueryInteractive(strings.TrimPrefix(line, "?")); err != nil {
				fmt.Fprintf(i.out, "error evaluating query: %v\n", err)
			}

		default:
			savedBuffer := i.buffer
			clauseText := line
			for !strings.HasSuffix(clauseText, ".") && !strings.HasSuffix(clauseText, "!") {
				nextLine, err := nextLineWithPrompt(continuedPrompt)
				if err != nil {
					return err
				}
				clauseText = clauseText + nextLine
			}

			if err := i.Define(clauseText); err != nil {
				fmt.Fprintf(i.out, "definition failed: %v\n", err)
				i.buffer = savedBuffer
			}
		}
	}
}

// Preload evaluates decls and clauses before any interactive evaluation takes place.
// This is used for customizing the interpreter.
// TODO: Add optional path parameter so the user can ::pop individual
// preloaded sources.
func (i *Interpreter) Preload(units []parse.SourceUnit, store factstore.FactStore, knownPredicates map[ast.PredicateSym]ast.Decl) error {
	i.store = store
	for sym, decl := range knownPredicates {
		i.knownPredicates[sym] = decl
	}
	programInfo, err := analysis.Analyze(units, i.knownPredicates)
	if err != nil {
		return err
	}
	i.pushSourceFragment(preloadPath, units, programInfo)
	return i.evalProgram(programInfo)
}

func (i *Interpreter) pushSourceFragment(path string, units []parse.SourceUnit, programInfo *analysis.ProgramInfo) {
	i.src = append(i.src, path)
	i.sourceFragments[path] = &sourceFragment{units, programInfo, i.store}
	for _, decl := range programInfo.Decls {
		i.knownPredicates[decl.DeclaredAtom.Predicate] = *decl
	}
	i.store = factstore.NewTeeingStore(i.store)
}

func (i *Interpreter) evalProgram(programInfo *analysis.ProgramInfo) error {
	stats, err := engine.EvalProgramWithStats(programInfo, i.store)
	if err != nil {
		return err
	}
	if len(i.stats) > 0 {
		showAllStats := len(i.stats) == 1 && i.stats[0] == "all"
		for _, name := range i.stats {
			for pred, stratum := range stats.PredToStratum {
				if pred.Symbol == name || showAllStats {
					fmt.Fprintf(i.out, "[%s %s (stratum %d)]\n", pred.Symbol, stats.Duration[stratum], stratum)
				}
			}
		}
	}
	for _, p := range i.postProcessors {
		if err := p(i.store); err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) popSourceFragment() *sourceFragment {
	l := len(i.src)
	if l <= 0 {
		return nil
	}
	path := i.src[l-1]
	f := i.sourceFragments[path]
	i.src = i.src[:l-1]
	delete(i.sourceFragments, path)
	for _, decl := range f.program.Decls {
		delete(i.knownPredicates, decl.DeclaredAtom.Predicate)
	}
	i.store = f.checkpoint
	return f
}

func (i *Interpreter) hasInteractiveDefs() bool {
	l := len(i.src)
	return l > 0 && i.src[l-1] == interactivePath
}

func (i *Interpreter) resetInteractiveDefs(buffer string) {
	if i.hasInteractiveDefs() {
		i.popSourceFragment()
	}
	i.buffer = buffer
}
