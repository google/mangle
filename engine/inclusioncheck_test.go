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

package engine

import (
	"fmt"
	"testing"

	"github.com/google/mangle/ast"
	"github.com/google/mangle/factstore"
)

func TestInclusionCheck(t *testing.T) {
	store := factstore.NewSimpleInMemoryStore()

	for i := 0; i < 255; i++ {
		store.Add(atom(fmt.Sprintf("range(%d)", i)))
	}
	store.Add(atom("color(1,2,3)"))
	store.Add(atom("color(123,243,33)"))
	store.Add(atom("color(13,23,33)"))

	rangeDecl, err := ast.NewDecl(
		ast.NewAtom("range", ast.Variable{"N"}),
		nil,
		[]ast.BoundDecl{
			ast.NewBoundDecl(ast.NumberBound)}, nil)
	if err != nil {
		t.Fatal(err)
	}
	colorDecl, err := ast.NewDecl(
		ast.NewAtom("color", ast.Variable{"R"}, ast.Variable{"G"}, ast.Variable{"B"}),
		nil,
		[]ast.BoundDecl{
			ast.NewBoundDecl(ast.String("range"), ast.String("range"), ast.String("range"))}, nil)
	if err != nil {
		t.Fatal(err)
	}
	paletteDecl, err := ast.NewDecl(
		ast.NewAtom("palette", ast.Variable{"Name"}, ast.Variable{"R"}, ast.Variable{"G"}, ast.Variable{"B"}),
		nil,
		[]ast.BoundDecl{
			ast.NewBoundDecl(ast.StringBound, ast.NumberBound, ast.NumberBound, ast.NumberBound)},
		&ast.InclusionConstraint{[]ast.Atom{
			ast.NewAtom("color", ast.Variable{"R"}, ast.Variable{"G"}, ast.Variable{"B"})}, nil})
	if err != nil {
		t.Fatal(err)
	}
	decl := map[ast.PredicateSym]ast.Decl{
		ast.PredicateSym{"palette", 4}: paletteDecl,
		ast.PredicateSym{"color", 3}:   colorDecl,
		ast.PredicateSym{"range", 1}:   rangeDecl,
	}

	checker, err := NewInclusionChecker(decl)
	if err != nil {
		t.Fatal(err)
	}
	badColor := atom("color(999,99,99)")
	if err = checker.CheckFact(badColor, store); err == nil { // if NO error
		t.Errorf("CheckFact(%v) succeeded, expected error %v", badColor, store)
	}
	okFact := atom("palette('kind of blue', 13, 23, 33)")
	if err = checker.CheckFact(okFact, store); err != nil {
		t.Errorf("CheckFact(%v) failed %v", okFact, err)
	}
	badFact := atom("palette('kind of black', 0, 0, 0)")
	if err = checker.CheckFact(badFact, store); err == nil { // if NO error
		t.Errorf("CheckFact(%v) succeeded, expected error %v", badFact, store)
	}
}
