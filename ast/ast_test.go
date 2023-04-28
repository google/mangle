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

package ast

import (
	"encoding/binary"
	"hash/fnv"
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var (
	fooName                  = name("/foo")
	fooNameSame              = name("/foo")
	barName                  = name("/bar")
	barString                = String("bar")
	bazString                = String("baz")
	num                      = int64(-123)
	numConstant              = Number(num)
	floatNum         float64 = 3.1415
	floatNumConstant         = Float64(floatNum)
	fooBarPair               = Pair(&fooName, &barString)
	barFooPair               = Pair(&barString, &fooName)
	fooFooPair               = Pair(&fooName, &fooName)
	fooBarList               = List([]Constant{fooName, barString})
	barFooList               = List([]Constant{barString, fooName})
	fooFooList               = List([]Constant{fooName, fooName})
	mapExample               = Map(map[*Constant]*Constant{
		&barString: &fooName,
		&bazString: &barName})
	mapExampleSame = Map(map[*Constant]*Constant{
		&bazString: &barName,
		&barString: &fooNameSame})
	mapExampleOther = Map(map[*Constant]*Constant{
		&bazString: &barName})
	structExample = Struct(map[*Constant]*Constant{
		&fooName: &fooName,
		&barName: &numConstant,
	})
	structExampleSame = Struct(map[*Constant]*Constant{
		&fooNameSame: &fooNameSame,
		&barName:     &numConstant,
	})
	structExampleOther = Struct(map[*Constant]*Constant{
		&barName: &numConstant,
	})
)

func name(str string) Constant {
	res, _ := Name(str)
	return res
}

func makeDecl(t *testing.T, atom Atom, descr []Atom, boundDecls []BoundDecl, incl *InclusionConstraint) Decl {
	t.Helper()
	decl, err := NewDecl(atom, descr, boundDecls, incl)
	if err != nil {
		t.Fatal(err)
	}
	return decl
}

func makeSyntheticDecl(t *testing.T, atom Atom) Decl {
	t.Helper()
	decl, err := NewSyntheticDecl(atom)
	if err != nil {
		t.Fatal(err)
	}
	return decl
}

func TestSelfEquals(t *testing.T) {
	tests := []Term{
		Variable{"X"},
		name("/foo"),
		String("foo"),
		Number(-123),
		floatNumConstant,
		fooBarPair,
		fooBarList,
		NewAtom("foo", Variable{"X"}),
		Eq{Variable{"X"}, Variable{"Y"}},
		Ineq{Variable{"X"}, Variable{"Y"}},
		ApplyFn{FunctionSym{"fn:list", -1}, []BaseTerm{Number(1), name("/foo")}},
		mapExample,
		structExample,
	}
	for _, testcase := range tests {
		if !testcase.Equals(testcase) {
			t.Errorf("(%v).Equals(%v) expected true got false", testcase, testcase)
		}
	}
}

func TestEqualsStructured(t *testing.T) {
	tests := []struct {
		left  *Constant
		right *Constant
		want  bool
	}{
		{&fooBarPair, &fooBarPair, true},
		{&fooBarPair, &barFooPair, false},
		{&fooBarList, &barFooList, false},
		{&fooBarList, &fooBarList, true},
		{mapExample, mapExampleSame, true},
		{mapExample, mapExampleOther, false},
		{mapExample, &MapNil, false},
		{&MapNil, mapExample, false},
		{&MapNil, &MapNil, true},
		{structExample, structExampleSame, true},
		{structExample, structExampleOther, false},
		{structExample, &StructNil, false},
		{&StructNil, structExample, false},
	}
	for _, testcase := range tests {
		got := testcase.left.Equals(testcase.right)
		if got != testcase.want {
			t.Errorf("(%v).Equals(%v) got %v want %v", testcase.left, testcase.right, got, testcase.want)
		}
		lh, rh := testcase.left.Hash(), testcase.right.Hash()
		if testcase.want {
			if lh != rh {
				t.Errorf("(%v).Hash() %d != (%v).Hash() %d but want same", testcase.left, lh, testcase.right, rh)
			}
		} else {
			if lh == rh {
				t.Errorf("(%v).Hash() %d == (%v).Hash() %d but want different", testcase.left, lh, testcase.right, rh)
			}
		}
	}
}

func TestSortIndexInto(t *testing.T) {
	keys := []*Constant{&fooName, &barString, &fooBarPair}
	index := make([]int, len(keys))
	SortIndexInto(keys, index)
	want := []int{2, 1, 0}
	if !cmp.Equal(index, want) {
		t.Errorf("SortIndexInto(%v) got %v want %v", keys, index, want)
	}
}

func TestEqualsNegative(t *testing.T) {
	tests := []struct {
		left  Term
		right Term
	}{
		{left: fooBarPair, right: fooBarList},
		{left: fooBarPair, right: barFooPair},
		{left: fooBarPair, right: fooFooPair},
		{left: fooBarList, right: fooBarPair},
		{left: fooBarList, right: barFooList},
		{left: fooBarList, right: fooFooList},
		{left: fooBarList, right: ListNil},
		{left: floatNumConstant, right: numConstant},
		{left: ListNil, right: fooBarList},
		{left: Variable{"X"}, right: Variable{"Y"}},
		{left: name("/foo"), right: name("/bar")},
		{left: name("/foo"), right: String("foo")},
		{left: NewAtom("foo", Variable{"X"}), right: NewAtom("foo", Variable{"Y"})},
		{left: Eq{Variable{"X"}, Variable{"Y"}}, right: Eq{Variable{"Y"}, Variable{"X"}}},
		{left: Ineq{Variable{"X"}, Variable{"Y"}}, right: Ineq{Variable{"Y"}, Variable{"X"}}},
	}
	for _, test := range tests {
		if test.left.Equals(test.right) {
			t.Errorf("(%v).Equals(%v) expected false got true", test.left, test.right)
		}
	}
}

func TestHash(t *testing.T) {
	tests := []struct {
		c    Constant
		want uint64
	}{
		{
			c:    fooName,
			want: hashBytes([]byte("/foo")),
		},
		{
			c:    barString,
			want: hashBytes([]byte(`bar`)),
		},
		{
			c:    Number(num),
			want: uint64(num),
		},
		{
			c:    floatNumConstant,
			want: math.Float64bits(floatNum),
		},
		{
			c:    fooBarPair,
			want: uint64(hashPair(&fooName, &barString, PairShape)),
		},
		{
			c:    List([]Constant{fooName}),
			want: uint64(hashPair(&fooName, &ListNil, ListShape)),
		},
		{
			c:    MapCons(&fooName, &barString, &MapNil),
			want: uint64(hashPair(&fooBarPair, &MapNil, MapShape)),
		},
		{
			c:    StructCons(&fooName, &barString, &StructNil),
			want: uint64(hashPair(&fooBarPair, &StructNil, StructShape)),
		},
	}
	for _, test := range tests {
		if got := test.c.Hash(); got != test.want {
			t.Errorf("(%v).Hash() expected %v got %v", test.c, test.want, got)
		}
	}
}

func TestAtomHash(t *testing.T) {
	a := NewAtom("bar", String("foo"))
	h := fnv.New64()
	h.Write([]byte(a.Predicate.String()))
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, a.Args[0].(Constant).Hash())
	h.Write(b)
	expected := h.Sum64()
	if a.Hash() != expected {
		t.Errorf("(%v).Hash() expected %v got %v", a, expected, a.Hash())
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name string
		term Term
		want string
	}{
		{
			"name constant",
			name("/foo"),
			"/foo",
		},
		{
			"number constant",
			Number(52),
			"52",
		},
		{
			"pair constant",
			fooBarPair,
			`</foo; "bar">`,
		},
		{
			"list constant",
			fooBarList,
			`[/foo, "bar"]`,
		},
		{
			"variable",
			Variable{"X"},
			"X",
		},
		{
			"atom",
			NewAtom("bar", Variable{"X"}),
			"bar(X)",
		},
		{
			"moreargs",
			NewAtom("bar", Variable{"X"}, Variable{"Y"}, name("/foo")),
			"bar(X,Y,/foo)",
		},
		{
			"negated atom",
			NewNegAtom("bar", Variable{"X"}),
			"!bar(X)",
		},
		{
			"equality",
			Eq{name("/bar"), Variable{"X"}},
			"/bar = X",
		},
		{
			"equality 2",
			Eq{Variable{"Y"}, Variable{"X"}},
			"Y = X",
		},
		{
			"inequality",
			Ineq{Variable{"Y"}, Variable{"X"}},
			"Y != X",
		},
	}
	for _, test := range tests {
		str := test.term.String()
		if str != test.want {
			t.Errorf("wanted %s got %s", test.want, str)
		}
	}
}

func TestName(t *testing.T) {
	tests := []struct {
		str  string
		want BaseTerm // nil means: expect error
	}{
		{
			"/bar",
			name("/bar"),
		},
		{
			"/bar/baz",
			name("/bar/baz"),
		},
		{
			"/",
			nil,
		},
		{
			"/bar/",
			nil,
		},
		{
			"//bar/",
			nil,
		},
	}
	for _, test := range tests {
		c, err := Name(test.str)
		if test.want != nil && err != nil {
			t.Errorf("could not construct constant with %s got %v", test.str, err)
		} else if test.want != nil && !c.Equals(test.want) {
			t.Errorf("wanted %s got %s", test.want, c)
		} else if test.want == nil && err == nil {
			t.Errorf("expected error but got %s", c)
		}
		str, err := c.NameValue()
		if err != nil {
			t.Fatal(err)
		}
		if test.want != nil && str != test.str {
			t.Errorf("want %q got %q", test.str, str)
		}
	}
}

func TestStringConstant(t *testing.T) {
	tests := []struct {
		c    Constant
		want string
	}{
		{
			c:    String("Lean down your ear upon the earth and listen."),
			want: "Lean down your ear upon the earth and listen.",
		},
		{
			c:    String(`hello world " \\ "`),
			want: `hello world " \\ "`,
		},
	}
	for _, test := range tests {
		str, err := test.c.StringValue()
		if err != nil {
			t.Fatal(err)
		}
		if str != test.want {
			t.Errorf("want %q got %q", test.want, str)
		}
	}
}

func TestNumberConstant(t *testing.T) {
	c := Number(-42)
	v, err := c.NumberValue()
	if err != nil {
		t.Fatal(err)
	}
	if v != -42 {
		t.Errorf("wanted %d got %d", -42, v)
	}
}

func TestPairConstant(t *testing.T) {
	num := Number(-200)
	str := String("bar")
	c := Pair(&num, &str)
	n, s, err := c.PairValue()
	if err != nil {
		t.Fatal(err)
	}
	if !n.Equals(num) || !s.Equals(str) {
		t.Errorf("wanted %v, %v got %v, %v", num, str, n, s)
	}
}

func TestListConstant(t *testing.T) {
	elems := []Constant{Number(-42), String("foo")}
	c := List(elems)
	var out []Constant
	err1, err2 := c.ListValues(func(elem Constant) error {
		out = append(out, elem)
		return nil
	}, func() error { return nil })
	if err1 != nil {
		t.Fatal(err1)
	}
	if err2 != nil {
		t.Fatal(err2)
	}
	compareFn := func(left, right Constant) bool {
		return left.Equals(right)
	}
	if !cmp.Equal(out, elems, cmp.Comparer(compareFn)) {
		t.Errorf("not equal %v and %v", out, elems)
	}
}

func TestBasicData(t *testing.T) {
	tests := []struct {
		name  string
		subst Subst
		want  []ConstSubstPair
	}{
		{
			name: "tuple",
			subst: ConstSubstList{
				{Variable{"X"}, name("/bar")},
				{Variable{"Y"}, name("/baz")},
			},
			want: []ConstSubstPair{
				{Variable{"X"}, name("/bar")},
				{Variable{"Y"}, name("/baz")},
			},
		},
		{
			name: "map",
			subst: ConstSubstMap{
				Variable{"X"}: name("/bar"),
				Variable{"Y"}: name("/baz"),
			},
			want: []ConstSubstPair{
				{Variable{"X"}, name("/bar")},
				{Variable{"Y"}, name("/baz")},
			},
		},
	}
	for _, test := range tests {
		for _, pair := range test.want {
			if !test.subst.Get(pair.v).Equals(pair.c) {
				t.Errorf("%s: subst did not yield %v but %v ", test.name, pair.c, test.subst.Get(pair.v))
			}
		}
	}
}

func TestAddVars(t *testing.T) {
	tests := []struct {
		term Term
		want []string
	}{
		{
			term: Variable{"X"},
			want: []string{"X"},
		},
		{
			term: Variable{"_"},
			want: []string{"_"},
		},
		{
			term: NewAtom("foo", Variable{"X"}),
			want: []string{"X"},
		},
		{
			term: Eq{Variable{"Y"}, ApplyFn{FunctionSym{"fn:foo", 1}, []BaseTerm{Variable{"X"}}}},
			want: []string{"X", "Y"},
		},
	}
	for _, test := range tests {
		vars := make(map[Variable]bool)
		AddVars(test.term, vars)
		if len(vars) != len(test.want) {
			t.Errorf("AddVars(%v, {})=%v want %v", test.term, vars, test.want)
		}
		for _, v := range test.want {
			if !vars[Variable{v}] {
				t.Errorf("AddVars(%v, {})=%v does not contain %v", test.term, vars, v)
			}
		}
	}
}

func TestReplaceWildcards(t *testing.T) {
	tests := []struct {
		clause      Clause
		wantNumVars int
	}{
		{
			clause: NewClause(
				NewAtom("foo", Variable{"X"}),
				[]Term{
					NewAtom("baz", Variable{"X"}, Variable{"_"}),
					NewAtom("bar", Variable{"_"}, Variable{"X"}),
				}),
			wantNumVars: 3,
		},
	}
	for _, test := range tests {
		got := test.clause.ReplaceWildcards()
		vars := make(map[Variable]bool)
		AddVarsFromClause(got, vars)
		if vars[Variable{"_"}] {
			t.Errorf("(%v).ReplaceWildcards()=%v contains a wildcard", test.clause, got)
		}
		if len(vars) != test.wantNumVars {
			t.Errorf("(%v).ReplaceWildcards()=%v want %v variables", test.clause, got, test.wantNumVars)
		}
	}
}

func TestSyntheticDecl(t *testing.T) {
	decl := makeSyntheticDecl(t, NewQuery(PredicateSym{"foo", 1}))

	if !decl.IsSynthetic() {
		t.Fatalf("(%v).IsSynthetic() = false want true", decl)
	}

	modes := decl.Modes()
	if len(modes) != 1 {
		t.Fatalf("(%v).Modes() = %v want one mode", decl, modes)
	}
	want := Mode{ArgModeInputOutput}
	if diff := cmp.Diff(want, modes[0]); diff != "" {
		t.Fatalf("(%v).Modes() got diff (-want, +got) %s", decl, diff)
	}
}

func TestDeclPackage(t *testing.T) {
	tests := []struct {
		predicate string
		want      string
	}{
		{predicate: "bar", want: ""},
		{predicate: "foo.bar", want: "foo"},
		{predicate: "foo.baz.bar", want: "foo.baz"},
	}
	for _, test := range tests {
		decl := makeSyntheticDecl(t, NewQuery(PredicateSym{test.predicate, 1}))

		if got := decl.PackageID(); got != test.want {
			t.Fatalf("PackageID(%v)=%v want %v", decl, got, test.want)
		}

		if !decl.IsSynthetic() {
			t.Fatalf("(%v).IsSynthetic() = false want true", decl)
		}
	}
}

func TestDeclPackageDecl(t *testing.T) {
	tests := []string{
		"foo", "foo.bar", "foo.baz.bar",
	}
	for _, test := range tests {
		decl := Decl{
			DeclaredAtom: NewAtom("Package"),
			Descr:        []Atom{NewAtom("name", String(test))},
		}

		if got := decl.PackageID(); got != test {
			t.Fatalf("PackageID(%v)=%v want %v", decl, got, test)
		}
	}
}

func TestDeclVisible(t *testing.T) {
	tests := []struct {
		desc string
		decl Decl
		want bool
	}{
		{
			desc: "Private atom, predicate is not visible.",
			decl: makeDecl(t, NewAtom("foo.bar"), []Atom{NewAtom("private")}, nil, nil),
			want: false,
		},
		{
			desc: "Public atom, predicate is visible.",
			decl: makeDecl(t, NewAtom("foo.bar"), []Atom{NewAtom("public")}, nil, nil),
			want: true,
		},
		{
			desc: "No atom, predicate is visible",
			decl: makeDecl(t, NewAtom("foo.bar"), nil, nil, nil),
			want: true,
		},
	}
	for _, test := range tests {
		if got := test.decl.Visible(); got != test.want {
			t.Fatalf("Visible(%v)=%v want %v", test.decl, got, test.want)
		}
	}
}

func TestModeCheck(t *testing.T) {
	mode := Mode{ArgModeInput, ArgModeOutput, ArgModeInputOutput}
	tests := []struct {
		goal      Atom
		boundVars map[Variable]bool
		want      bool
	}{
		{
			goal:      NewAtom("foo", Variable{"X"}, Number(2), Number(3)),
			boundVars: nil,
			want:      false,
		},
		{
			goal:      NewAtom("foo", Number(2), Variable{"X"}, Number(3)),
			boundVars: nil,
			want:      true,
		},
		{
			goal:      NewAtom("foo", Variable{"X"}, Number(2), Number(3)),
			boundVars: map[Variable]bool{Variable{"X"}: true},
			want:      false,
		},
		{
			goal:      NewAtom("foo", Variable{"X"}, Variable{"Y"}, Number(3)),
			boundVars: map[Variable]bool{Variable{"X"}: true},
			want:      true,
		},
		{
			goal:      NewAtom("foo", Variable{"X"}, Variable{"Y"}, Variable{"Z"}),
			boundVars: map[Variable]bool{Variable{"X"}: true},
			want:      true,
		},
	}
	for _, test := range tests {
		err := mode.Check(test.goal, test.boundVars)
		if (err == nil) != test.want {
			if test.want {
				t.Errorf("Check(%v)=%v want no errors", test.goal, err)
			} else {
				t.Errorf("Check(%v)=%v want error", test.goal, err)
			}
		}
	}
}
