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

package symbols

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/mangle/ast"
)

func pair(fst ast.Constant, snd ast.Constant) ast.Constant {
	return ast.Pair(&fst, &snd)
}

func TestHasType(t *testing.T) {
	fooName, _ := ast.Name("/foo")
	fooBarName, _ := ast.Name("/foo/bar")
	tests := []struct {
		tpe  ast.BaseTerm
		good []ast.Constant
		bad  []ast.Constant
	}{
		{
			tpe:  ast.AnyBound,
			good: []ast.Constant{ast.String("foo"), ast.Number(23), ast.ListNil},
			bad:  nil,
		},
		{
			tpe:  ast.NameBound,
			good: []ast.Constant{ast.NameBound, fooName},
			bad:  []ast.Constant{ast.String("foo"), ast.Number(23), ast.ListNil},
		},
		{
			tpe:  fooName,
			good: []ast.Constant{fooBarName},
			bad:  []ast.Constant{fooName, ast.Number(23), ast.ListNil},
		},
		{
			tpe:  ast.NumberBound,
			good: []ast.Constant{ast.Number(23)},
			bad:  []ast.Constant{ast.String("foo"), ast.AnyBound, ast.ListNil},
		},
		{
			tpe:  ast.StringBound,
			good: []ast.Constant{ast.String("foo")},
			bad:  []ast.Constant{ast.Number(23), ast.AnyBound, ast.ListNil},
		},
		{
			tpe: NewListType(ast.NumberBound),
			good: []ast.Constant{
				ast.ListNil,
				ast.List([]ast.Constant{ast.Number(2), ast.Number(2)}),
			},
			bad: []ast.Constant{
				ast.Number(23),
				ast.AnyBound,
				ast.List([]ast.Constant{ast.String("foo")}),
				ast.List([]ast.Constant{ast.Number(2), ast.String("foo")}),
			},
		},
		{
			tpe:  NewPairType(ast.Float64Bound, ast.StringBound),
			good: []ast.Constant{pair(ast.Float64(2.2), ast.String("foo"))},
			bad:  []ast.Constant{ast.Number(2), pair(ast.String("foo"), ast.Number(2))},
		},
		{
			tpe:  NewTupleType(ast.NumberBound, ast.StringBound, NewListType(ast.NumberBound)),
			good: []ast.Constant{pair(ast.Number(2), pair(ast.String("foo"), ast.ListNil))},
			bad:  []ast.Constant{ast.Number(2), pair(ast.String("foo"), ast.Number(2))},
		},
		{
			tpe: NewUnionType(ast.NumberBound, ast.StringBound, NewListType(ast.NumberBound)),
			good: []ast.Constant{
				ast.Number(2),
				ast.String("foo"),
				ast.ListNil,
				ast.List([]ast.Constant{ast.Number(2), ast.Number(2)}),
			},
			bad: []ast.Constant{
				ast.AnyBound,
				ast.List([]ast.Constant{ast.String("foo")}),
			},
		},
		{
			tpe:  NewSingletonType(ast.TrueConstant),
			good: []ast.Constant{ast.TrueConstant},
			bad:  []ast.Constant{ast.FalseConstant},
		},
		{
			tpe:  NewUnionType(NewSingletonType(ast.TrueConstant), NewSingletonType(ast.FalseConstant)),
			good: []ast.Constant{ast.TrueConstant, ast.FalseConstant},
			bad:  []ast.Constant{ast.AnyBound},
		},
		// Structured values that need evaluation for readability are tested in builtin_test.go
	}
	for _, test := range tests {
		h, err := NewSetHandle(test.tpe)
		if err != nil {
			t.Errorf("NewTypeHandle(%v) failed %v", test.tpe, err)
		}
		for _, c := range test.good {
			if !h.HasType(c) {
				t.Errorf("NewTypeHandle(%v).HasType(%v)=false want true", test.tpe, c)
			}
		}
		for _, c := range test.bad {
			if h.HasType(c) {
				t.Errorf("NewTypeHandle(%v).HasType(%v)=true want false", test.tpe, c)
			}
		}
	}
}

func TestWellformedBound(t *testing.T) {
	tests := []struct {
		tpe ast.BaseTerm
	}{
		{
			tpe: NewFunType(
				NewPairType(ast.Variable{"X"}, ast.Variable{"Y"}),
				// <=
				ast.Variable{"X"}, ast.Variable{"Y"}),
		},
	}
	for _, test := range tests {
		_, err := NewBoundHandle(test.tpe)
		if err != nil {
			t.Errorf("NewTypeHandle(%v) failed %v", test.tpe, err)
		}
	}
}

func TestWellformedType(t *testing.T) {
	tests := []struct {
		tpe  ast.BaseTerm
		vars map[ast.Variable]ast.BaseTerm
	}{
		{
			tpe: NewFunType(
				NewPairType(ast.Variable{"X"}, ast.Variable{"Y"}),
				// <=
				ast.Variable{"X"}, ast.Variable{"Y"}),
			vars: map[ast.Variable]ast.BaseTerm{
				ast.Variable{"X"}: ast.NumberBound,
				ast.Variable{"Y"}: ast.NumberBound,
			},
		},
	}
	for _, test := range tests {
		_, err := NewTypeHandle(test.vars, test.tpe)
		if err != nil {
			t.Errorf("NewTypeHandle(%v) failed %v", test.tpe, err)
		}
	}
}

func TestWellformedTypeNegative(t *testing.T) {
	tests := []struct {
		tpe  ast.BaseTerm
		vars map[ast.Variable]ast.BaseTerm
	}{
		{
			tpe: NewFunType(
				NewPairType(ast.Variable{"X"}, ast.Variable{"Y"}),
				// <=
				ast.Variable{"X"}, ast.Variable{"Y"}),
			vars: nil,
		},

		{
			tpe: NewFunType(
				NewPairType(ast.Variable{"X"}, ast.Variable{"Y"}),
				// <=
				ast.Variable{"X"}, ast.Variable{"Y"}),
			vars: map[ast.Variable]ast.BaseTerm{
				// missing "X"
				ast.Variable{"Y"}: ast.NumberBound,
			},
		},
	}
	for _, test := range tests {
		_, err := NewTypeHandle(test.vars, test.tpe)
		if err != nil {
			t.Errorf("NewTypeHandle(%v) failed %v", test.tpe, err)
		}
	}
}

func TestSetExpressionNegative(t *testing.T) {
	tests := []ast.BaseTerm{
		ast.ApplyFn{MapType, []ast.BaseTerm{name("/foo")}},
		ast.ApplyFn{MapType, []ast.BaseTerm{ast.Number(2), name("/foo")}},
		ast.ApplyFn{StructType, []ast.BaseTerm{name("/foo")}},
		ast.ApplyFn{StructType, []ast.BaseTerm{ast.Number(2), name("/foo")}},
	}
	for _, test := range tests {
		if h, err := NewSetHandle(test); err == nil { // if NO error
			t.Errorf("NewTypeHandle(%v)=%v succeeded, expected error", h, test)
		}
	}
}

func TestRelTypeExprFromDecl(t *testing.T) {
	relTypeExpr, err := RelTypeExprFromDecl(ast.Decl{
		ast.NewAtom("foo", ast.Variable{"X"}, ast.Variable{"Y"}),
		nil,
		[]ast.BoundDecl{
			ast.NewBoundDecl(ast.StringBound, ast.NumberBound),
			ast.NewBoundDecl(ast.NumberBound, ast.StringBound),
		},
		nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := NewUnionType(
		NewRelType(ast.StringBound, ast.NumberBound),
		NewRelType(ast.NumberBound, ast.StringBound))
	if !relTypeExpr.Equals(want) {
		t.Errorf("%v.Equals(%v)=false want true", relTypeExpr, want)
	}
}

func TestRelTypeMethods(t *testing.T) {
	tests := []struct {
		alternatives []ast.BaseTerm
		want         ast.BaseTerm
	}{
		{
			alternatives: []ast.BaseTerm{
				NewRelType(ast.StringBound, ast.NumberBound)},
			want: NewRelType(ast.StringBound, ast.NumberBound),
		},
		{
			alternatives: []ast.BaseTerm{
				NewRelType(ast.StringBound, ast.NumberBound), NewRelType(ast.NumberBound, ast.NumberBound)},
			want: NewUnionType(
				NewRelType(ast.StringBound, ast.NumberBound), NewRelType(ast.NumberBound, ast.NumberBound)),
		},
	}
	for _, test := range tests {
		got := RelTypeFromAlternatives(test.alternatives)
		if !got.Equals(test.want) {
			t.Errorf("RelTypeFromAlternatives(%v)=%v want %v", test.alternatives, got, test.want)
		}

		alts := RelTypeAlternatives(got)
		if diff := cmp.Diff(test.alternatives, alts, cmp.AllowUnexported(ast.Constant{})); diff != "" {
			t.Errorf("RelTypeAlternatives(RelTypeFromAlternatives(%v))=%v want %v", test.alternatives, alts, test.alternatives)
		}
	}
}

func TestSetConforms(t *testing.T) {
	tests := []struct {
		left  ast.BaseTerm
		right ast.BaseTerm
		want  bool
	}{
		{ast.NameBound, ast.AnyBound, true},
		{ast.BotBound, ast.AnyBound, true},
		{ast.AnyBound, ast.BotBound, false},
		{ast.NameBound, ast.NameBound, true},
		{name("/foo"), name("/foo"), true},
		{name("/foo"), ast.NameBound, true},
		{name("/true"), NewUnionType(name("/true"), name("/false")), true},
		{ast.NameBound, name("/foo"), false},
		{NewUnionType(name("/true"), name("/false")), name("/true"), false},
		{
			NewListType(ast.BotBound),
			NewListType(ast.NumberBound),
			true,
		},
		{
			NewMapType(ast.AnyBound, ast.NumberBound),
			NewMapType(ast.StringBound, ast.NumberBound),
			true,
		},
		{
			NewMapType(ast.NumberBound, ast.AnyBound),
			NewMapType(ast.StringBound, ast.AnyBound),
			false,
		},
		{
			NewStructType(name("/foo"), ast.AnyBound, name("/bar"), ast.NumberBound),
			NewStructType(name("/foo"), ast.AnyBound),
			true,
		},
		{
			NewStructType(name("/foo"), ast.AnyBound, NewOpt(name("/bar"), ast.NumberBound)),
			NewStructType(name("/foo"), ast.AnyBound, name("/bar"), ast.NumberBound),
			false,
		},
		{
			NewStructType(name("/foo"), ast.AnyBound, name("/bar"), ast.NumberBound),
			NewStructType(name("/foo"), ast.AnyBound, NewOpt(name("/bar"), ast.NumberBound)),
			true,
		},
		{
			NewStructType(name("/foo"), ast.AnyBound, name("/bar"), ast.NumberBound),
			NewStructType(),
			true,
		},
		{
			NewStructType(),
			NewStructType(),
			true,
		},
		{
			NewStructType(),
			NewStructType(NewOpt(name("/bar"), ast.NumberBound)),
			true,
		},
		{
			NewRelType(ast.StringBound, ast.NumberBound),
			NewRelType(ast.StringBound, ast.NumberBound),
			true,
		},
		{
			NewRelType(ast.StringBound, ast.NumberBound),
			NewRelType(ast.NumberBound, ast.NumberBound),
			false,
		},
		{
			NewRelType(ast.StringBound, ast.NumberBound),
			NewRelType(ast.AnyBound, ast.NumberBound),
			true,
		},
		{
			NewRelType(ast.StringBound, NewListType(ast.BotBound)),
			NewRelType(ast.StringBound, NewListType(ast.StringBound)),
			true,
		},
		{
			NewRelType(ast.StringBound, ast.NumberBound),
			NewUnionType(NewRelType(ast.StringBound, ast.NumberBound), NewRelType(ast.NumberBound, ast.NumberBound)),
			true,
		},
		{
			NewUnionType(NewRelType(ast.StringBound, ast.NumberBound), NewRelType(ast.NumberBound, ast.NumberBound)),
			NewUnionType(NewRelType(ast.StringBound, ast.NumberBound), NewRelType(ast.NumberBound, ast.NumberBound)),
			true,
		},
		{
			NewUnionType(NewRelType(ast.StringBound, ast.NumberBound), NewRelType(ast.NumberBound, ast.NumberBound)),
			NewUnionType(NewRelType(ast.AnyBound, ast.StringBound), NewRelType(ast.NumberBound, ast.NumberBound)),
			false,
		},
	}
	for _, test := range tests {
		if got := SetConforms(nil, test.left, test.right); got != test.want {
			t.Errorf("MonoTypeConforms(%v, %v)=%v want %v", test.left, test.right, got, test.want)
		}
	}
}

func TestTypeConforms(t *testing.T) {
	tests := []struct {
		ctx   map[ast.Variable]ast.BaseTerm
		left  ast.BaseTerm
		right ast.BaseTerm
		want  bool
	}{
		{
			map[ast.Variable]ast.BaseTerm{ast.Variable{"X"}: NewListType(ast.NumberBound)},
			NewUnionType(NewRelType(ast.StringBound, ast.NumberBound), NewRelType(ast.NumberBound, ast.NumberBound)),
			NewUnionType(NewRelType(ast.AnyBound, ast.StringBound), NewRelType(ast.NumberBound, ast.NumberBound)),
			false,
		},
		{
			map[ast.Variable]ast.BaseTerm{ast.Variable{"X"}: NewListType(ast.NumberBound)},
			NewFunType(ast.NumberBound /* <= */, ast.Variable{"X"}),
			NewFunType(ast.NumberBound /* <= */, NewListType(ast.NumberBound)),
			true,
		},
		{
			map[ast.Variable]ast.BaseTerm{ast.Variable{"X"}: NewListType(ast.NumberBound)},
			NewFunType(ast.NumberBound /* <= */, NewListType(ast.NumberBound)),
			NewFunType(ast.NumberBound /* <= */, ast.Variable{"X"}),
			true,
		},
		{
			map[ast.Variable]ast.BaseTerm{ast.Variable{"X"}: NewListType(ast.NumberBound)},
			NewFunType(ast.Variable{"X"} /* <= */, ast.AnyBound),
			NewFunType(ast.AnyBound /* <= */, NewListType(ast.NumberBound)),
			true,
		},
	}
	for _, test := range tests {
		if got := TypeConforms(test.ctx, test.left, test.right); got != test.want {
			t.Errorf("TypeConforms(%v, %v, %v)=%v want %v", test.ctx, test.left, test.right, got, test.want)
		}
	}
}

func TestAccess(t *testing.T) {
	tpe := NewListType(ast.AnyBound)
	if !IsListTypeExpression(tpe) {
		t.Errorf("IsListTypeExpression(%v)=false want true", tpe)
	}
	arg, err := ListTypeArg(tpe)
	if err != nil {
		t.Fatal(err)
	}
	if arg != ast.AnyBound {
		t.Errorf("ListTypeArg(%v)=%v want %v", tpe, arg, ast.AnyBound)
	}
	tpe = NewMapType(ast.StringBound, ast.NumberBound)
	if !IsMapTypeExpression(tpe) {
		t.Errorf("IsListTypeExpression(%v)=false want true", tpe)
	}
	key, val, err := MapTypeArgs(tpe)
	if err != nil {
		t.Fatal(err)
	}
	if key != ast.StringBound {
		t.Errorf("MapTypeArgs(%v)=[%v],%v want %v", tpe, key, val, ast.StringBound)
	}
	if val != ast.NumberBound {
		t.Errorf("MapTypeArgs(%v)=%v,[%v] want %v", tpe, key, val, ast.NumberBound)
	}
	tpe = NewStructType(name("/foo"), ast.StringBound, NewOpt(name("/bar"), ast.NumberBound))
	if !IsStructTypeExpression(tpe) {
		t.Errorf("IsStructTypeExpression(%v)=false want true", tpe)
	}
	requiredArgs, err := StructTypeRequiredArgs(tpe)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(requiredArgs, []ast.BaseTerm{name("/foo"), ast.StringBound}, cmp.AllowUnexported(ast.Constant{})) {
		t.Errorf("StructTypeRequiredArgs(%v)=%v want [/foo, /string]", tpe, requiredArgs)
	}
	optArgs, err := StructTypeOptionaArgs(tpe)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(optArgs, []ast.BaseTerm{NewOpt(name("/bar"), ast.NumberBound)}, cmp.AllowUnexported(ast.Constant{})) {
		t.Errorf("StructTypeOptionaArgs(%v)=%v want fn:opt(/bar, /number)", tpe, optArgs)
	}
	fooTpe, err := StructTypeField(tpe, name("/foo"))
	if err != nil {
		t.Fatal(err)
	}
	if !fooTpe.Equals(ast.StringBound) {
		t.Errorf("StructTypeField(%v,%v)=%v want /string", tpe, name("foo"), fooTpe)
	}
	barTpe, err := StructTypeField(tpe, name("/bar"))
	if err != nil {
		t.Fatal(err)
	}
	if !barTpe.Equals(ast.NumberBound) {
		t.Errorf("StructTypeField(%v,%v)=%v want /number", tpe, name("bar"), barTpe)
	}
}

func TestUpperBound(t *testing.T) {
	tests := []struct {
		exprs []ast.BaseTerm
		want  ast.BaseTerm
	}{
		{
			exprs: []ast.BaseTerm{ast.NumberBound},
			want:  ast.NumberBound,
		},
		{
			exprs: []ast.BaseTerm{ast.NumberBound, ast.StringBound},
			want:  NewUnionType(ast.StringBound, ast.NumberBound),
		},
		{
			exprs: []ast.BaseTerm{ast.NumberBound, NewUnionType(ast.NumberBound, ast.StringBound)},
			want:  NewUnionType(ast.StringBound, ast.NumberBound),
		},
		{
			exprs: []ast.BaseTerm{ast.AnyBound, ast.StringBound},
			want:  ast.AnyBound,
		},
		{
			exprs: []ast.BaseTerm{NewListType(ast.AnyBound), NewListType(ast.StringBound)},
			want:  NewListType(ast.AnyBound),
		},
		{
			exprs: []ast.BaseTerm{
				NewPairType(ast.StringBound, ast.NumberBound),
				NewPairType(ast.AnyBound, ast.NumberBound)},
			want: NewPairType(ast.AnyBound, ast.NumberBound),
		},
		{
			exprs: []ast.BaseTerm{
				NewTupleType(ast.AnyBound, ast.NumberBound, ast.NameBound),
				NewTupleType(ast.StringBound, ast.NumberBound, name("/foo"))},
			want: NewTupleType(ast.AnyBound, ast.NumberBound, ast.NameBound),
		},
		{
			exprs: nil,
			want:  EmptyType,
		},
	}
	for _, test := range tests {
		got := UpperBound(nil, test.exprs)
		if !got.Equals(test.want) {
			t.Errorf("UpperBound(%v)=%v want %v", test.exprs, got, test.want)
		}
	}
}

func TestLowerBound(t *testing.T) {
	foo := name("/foo")
	fooBarBaz := name("/foo/bar/baz")
	fooBarFoo := name("/foo/bar/foo")

	tests := []struct {
		exprs []ast.BaseTerm
		want  ast.BaseTerm
	}{
		{
			exprs: nil,
			want:  ast.AnyBound,
		},
		{
			exprs: []ast.BaseTerm{ast.NumberBound},
			want:  ast.NumberBound,
		},
		{
			exprs: []ast.BaseTerm{ast.NumberBound, ast.StringBound},
			want:  NewUnionType(),
		},
		{
			exprs: []ast.BaseTerm{ast.StringBound, ast.NumberBound},
			want:  NewUnionType(),
		},
		{
			exprs: []ast.BaseTerm{ast.AnyBound, ast.StringBound},
			want:  ast.StringBound,
		},
		{
			exprs: []ast.BaseTerm{NewPairType(ast.StringBound, ast.StringBound), ast.StringBound},
			want:  EmptyType,
		},
		{
			exprs: []ast.BaseTerm{
				NewUnionType(ast.NumberBound, ast.StringBound),
				ast.StringBound,
			},
			want: ast.StringBound,
		},
		{
			exprs: []ast.BaseTerm{
				ast.StringBound,
				NewUnionType(ast.NumberBound, ast.StringBound),
			},
			want: ast.StringBound,
		},
		{
			exprs: []ast.BaseTerm{
				NewUnionType(ast.NumberBound, ast.StringBound),
				NewUnionType(ast.NumberBound, ast.StringBound),
			},
			want: NewUnionType(ast.NumberBound, ast.StringBound),
		},
		{
			exprs: []ast.BaseTerm{
				NewUnionType(fooBarBaz, fooBarFoo),
				foo,
			},
			want: NewUnionType(fooBarBaz, fooBarFoo),
		},
		{
			exprs: []ast.BaseTerm{
				NewUnionType(fooBarBaz, foo),
				fooBarBaz,
			},
			want: fooBarBaz,
		},
		{
			exprs: []ast.BaseTerm{NewListType(ast.AnyBound), NewListType(ast.StringBound)},
			want:  NewListType(ast.StringBound),
		},
		{
			exprs: []ast.BaseTerm{
				NewPairType(ast.StringBound, ast.NumberBound),
				NewPairType(ast.AnyBound, ast.NumberBound)},
			want: NewPairType(ast.StringBound, ast.NumberBound),
		},
		{
			exprs: []ast.BaseTerm{
				NewTupleType(ast.AnyBound, ast.NumberBound, ast.NameBound),
				NewTupleType(ast.StringBound, ast.NumberBound, name("/foo"))},
			want: NewTupleType(ast.StringBound, ast.NumberBound, name("/foo")),
		},
	}
	for _, test := range tests {
		got := LowerBound(nil, test.exprs)
		if got == nil && test.want == nil {
			continue
		}
		if !(SetConforms(nil, got, test.want) && SetConforms(nil, test.want, got)) {
			t.Errorf("LowerBound(%v)=%v want %v", test.exprs, got, test.want)
		}
	}
}
