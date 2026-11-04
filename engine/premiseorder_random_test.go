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
	"flag"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"unicode"

	"codeberg.org/TauCeti/mangle-go/analysis"
	"codeberg.org/TauCeti/mangle-go/factstore"
	"codeberg.org/TauCeti/mangle-go/parse"
)

// randomPred is a predicate a generated rule can use.
type randomPred struct {
	name  string
	arity int
	layer int // 0 for base facts
}

// randomProgram writes a stratified program with random rules over small
// numbers: joins, recursion, negation, inequalities, comparisons, function
// applications, equalities, and let- and do-transforms. Premises come in a
// random order, so many rules are written in a poor order.
func randomProgram(r *rand.Rand) string {
	var sb strings.Builder
	base := []randomPred{{"e1", 1, 0}, {"e2", 2, 0}, {"e3", 2, 0}, {"e4", 1, 0}}
	for _, p := range base {
		for i := 0; i < 3+r.Intn(8); i++ {
			args := make([]string, p.arity)
			for j := range args {
				args[j] = fmt.Sprint(r.Intn(6))
			}
			fmt.Fprintf(&sb, "%s(%s).\n", p.name, strings.Join(args, ", "))
		}
	}
	preds := append([]randomPred{}, base...)
	for layer := 1; layer <= 3; layer++ {
		var layerPreds []randomPred
		for i := 0; i < 2; i++ {
			layerPreds = append(layerPreds, randomPred{fmt.Sprintf("p%d_%d", layer, i), 1 + r.Intn(2), layer})
		}
		available := append(append([]randomPred{}, preds...), layerPreds...)
		for _, head := range layerPreds {
			for k := 0; k < 1+r.Intn(2); k++ {
				sb.WriteString(randomRule(r, head, available))
			}
		}
		preds = append(preds, layerPreds...)
	}
	// An aggregation over the last layer, with a body of several premises.
	agg := randomPred{"agg", 1, 4}
	body := randomRule(r, randomPred{"unused", 1, 4}, preds)
	body = body[strings.Index(body, ":-")+2 : len(body)-2]
	fmt.Fprintf(&sb, "%s(N) :- %s |> do fn:group_by(), let N = fn:count().\n", agg.name, body)
	return sb.String()
}

var randomVars = []string{"X", "Y", "Z", "W"}

// isVariable reports whether a generated argument is a variable, as opposed
// to a number or the wildcard.
func isVariable(arg string) bool {
	return arg != "_" && !unicode.IsDigit(rune(arg[0]))
}

func randomArg(r *rand.Rand) string {
	switch r.Intn(10) {
	case 0, 1:
		return fmt.Sprint(r.Intn(6))
	case 2:
		return "_"
	}
	return randomVars[r.Intn(len(randomVars))]
}

// randomRule writes a rule for head. Positive premises come from preds of
// the head's layer or below; negated ones from a lower layer only.
func randomRule(r *rand.Rand, head randomPred, preds []randomPred) string {
	var positives, lower []randomPred
	for _, p := range preds {
		if p.layer <= head.layer {
			positives = append(positives, p)
		}
		if p.layer < head.layer {
			lower = append(lower, p)
		}
	}
	type term struct {
		text  string
		binds []string // variables the premise binds
		uses  []string // variables that must have a value before it
		fail  bool     // can fail: goes after the premises that bind its uses
	}
	var terms []term
	bound := map[string]bool{}
	// A recursive rule must not compute new values, or it would never stop.
	recursive := false
	for i := 0; i < 1+r.Intn(4); i++ {
		p := positives[r.Intn(len(positives))]
		recursive = recursive || p.layer == head.layer
		args := make([]string, p.arity)
		var vars []string
		for j := range args {
			args[j] = randomArg(r)
			if isVariable(args[j]) {
				vars = append(vars, args[j])
				bound[args[j]] = true
			}
		}
		terms = append(terms, term{fmt.Sprintf("%s(%s)", p.name, strings.Join(args, ", ")), vars, nil, false})
	}
	var boundVars []string
	for _, v := range randomVars {
		if bound[v] {
			boundVars = append(boundVars, v)
		}
	}
	pick := func() string {
		if len(boundVars) == 0 || r.Intn(4) == 0 {
			return fmt.Sprint(r.Intn(6))
		}
		return boundVars[r.Intn(len(boundVars))]
	}
	usesOf := func(args ...string) []string {
		var vars []string
		for _, a := range args {
			if bound[a] {
				vars = append(vars, a)
			}
		}
		return vars
	}
	if len(lower) > 0 && r.Intn(3) == 0 {
		p := lower[r.Intn(len(lower))]
		args := make([]string, p.arity)
		for j := range args {
			args[j] = pick()
		}
		terms = append(terms, term{fmt.Sprintf("!%s(%s)", p.name, strings.Join(args, ", ")), nil, usesOf(args...), false})
	}
	if r.Intn(3) == 0 {
		a, b := pick(), pick()
		terms = append(terms, term{fmt.Sprintf("%s != %s", a, b), nil, usesOf(a, b), false})
	}
	if r.Intn(4) == 0 && len(boundVars) > 0 {
		a := boundVars[r.Intn(len(boundVars))]
		terms = append(terms, term{fmt.Sprintf("%s < %d", a, 1+r.Intn(5)), nil, []string{a}, true})
	}
	headVars := append([]string{}, boundVars...)
	if !recursive && r.Intn(4) == 0 && len(boundVars) > 0 {
		a := boundVars[r.Intn(len(boundVars))]
		terms = append(terms, term{fmt.Sprintf("N = fn:plus(%s, 1)", a), []string{"N"}, []string{a}, true})
		headVars = append(headVars, "N")
	}
	if !recursive && r.Intn(4) == 0 && len(boundVars) > 0 {
		// Built-ins over a list: one binds every element in turn, one takes
		// the list apart. Only elements, which are numbers, reach the head.
		a, b := pick(), pick()
		terms = append(terms, term{fmt.Sprintf("Lst = fn:list(%s, %s)", a, b), []string{"Lst"}, usesOf(a, b), true})
		if r.Intn(2) == 0 {
			terms = append(terms, term{":list:member(Elt, Lst)", []string{"Elt"}, []string{"Lst"}, true})
			headVars = append(headVars, "Elt")
			if r.Intn(2) == 0 {
				terms = append(terms, term{fmt.Sprintf("Elt < %d", 1+r.Intn(5)), nil, []string{"Elt"}, true})
			}
		} else {
			terms = append(terms, term{":match_cons(Lst, Hd, Tl)", []string{"Hd", "Tl"}, []string{"Lst"}, true})
			headVars = append(headVars, "Hd")
		}
	}
	if r.Intn(5) == 0 && len(boundVars) > 0 {
		a := boundVars[r.Intn(len(boundVars))]
		terms = append(terms, term{fmt.Sprintf("M = %s", a), nil, nil, false})
		headVars = append(headVars, "M")
	}
	if r.Intn(5) == 0 && len(boundVars) > 0 {
		terms = append(terms, term{fmt.Sprintf("%s = %d", boundVars[r.Intn(len(boundVars))], r.Intn(6)), nil, nil, false})
	}
	// Shuffle the premises. A premise that can fail goes after the premises
	// that bind its variables, so that the rule is valid as written.
	var free, failing []term
	for _, t := range terms {
		if t.fail {
			failing = append(failing, t)
		} else {
			free = append(free, t)
		}
	}
	r.Shuffle(len(free), func(i, j int) { free[i], free[j] = free[j], free[i] })
	body := free
	for _, f := range failing {
		after := 0
		for i, t := range body {
			for _, v := range f.uses {
				for _, w := range t.binds {
					if v == w && i+1 > after {
						after = i + 1
					}
				}
			}
		}
		at := after + r.Intn(len(body)-after+1)
		body = append(body[:at], append([]term{f}, body[at:]...)...)
	}
	var parts []string
	for _, t := range body {
		parts = append(parts, t.text)
	}
	args := make([]string, head.arity)
	for j := range args {
		if len(headVars) == 0 {
			args[j] = fmt.Sprint(r.Intn(6))
		} else {
			args[j] = headVars[r.Intn(len(headVars))]
		}
	}
	rule := fmt.Sprintf("%s(%s) :- %s", head.name, strings.Join(args, ", "), strings.Join(parts, ", "))
	if !recursive && r.Intn(6) == 0 && head.layer < 4 {
		// A let-transform: the head gets a computed value.
		rule = fmt.Sprintf("%s(%s) :- %s |> let L = fn:plus(%s, 0)", head.name,
			strings.Join(append(args[:head.arity-1], "L"), ", "), strings.Join(parts, ", "), pick())
	}
	return rule + ".\n"
}

var randomPrograms = flag.Int("reorder_programs", 300,
	"how many random programs TestPremiseReorderingRandomPrograms evaluates")

// Every random program derives the same facts with and without reordering,
// on both kinds of store. Run more of them with -reorder_programs=20000.
func TestPremiseReorderingRandomPrograms(t *testing.T) {
	programs := int64(*randomPrograms)
	var accepted, failing, reordered int
	for seed := int64(1); seed <= programs; seed++ {
		r := rand.New(rand.NewSource(seed))
		source := randomProgram(r)
		unit, err := parse.Unit(strings.NewReader(source))
		if err != nil {
			t.Fatalf("seed %d: parse: %v\n%s", seed, err, source)
		}
		pi, err := analysis.AnalyzeOneUnit(unit, nil)
		if err != nil {
			continue // The generator can write an invalid rule; skip it.
		}
		for _, newStore := range []func() factstore.FactStore{
			func() factstore.FactStore { return factstore.NewSimpleInMemoryStore() },
			func() factstore.FactStore { return factstore.NewMultiIndexedArrayInMemoryStore() },
		} {
			written := newStore()
			if _, err := EvalProgramWithStats(pi, written); err != nil {
				// Nothing to compare with.
				failing++
				break
			}
			store := newStore()
			stats, err := EvalProgramWithStats(pi, store, WithPremiseReordering())
			if err != nil {
				t.Fatalf("seed %d: reordered: %v\n%s", seed, err, source)
			}
			if got, want := strings.Join(allFacts(store), " "), strings.Join(allFacts(written), " "); got != want {
				t.Fatalf("seed %d: facts differ\nreordered: %s\nwritten:   %s\nprogram:\n%s\norders: %v", seed, got, want, source, stats.PremiseOrders)
			}
			if _, simple := written.(factstore.SimpleInMemoryStore); simple {
				accepted++
				if len(stats.PremiseOrders) > 0 {
					reordered++
				}
			}
		}
	}
	t.Logf("%d programs: %d valid, %d of them with at least one rule reordered; %d failing as written",
		programs, accepted, reordered, failing)
	// Make sure the generator exercises what the test is about.
	if int64(accepted) < programs/2 || reordered < accepted/4 {
		t.Errorf("only %d of %d programs valid and %d reordered; the generator needs work", accepted, programs, reordered)
	}
}
