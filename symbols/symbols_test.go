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

	"github.com/google/mangle/ast"
)

func pair(fst ast.Constant, snd ast.Constant) ast.Constant {
	return ast.Pair(&fst, &snd)
}

func TestCheckTypeExpression(t *testing.T) {
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
		// Structured values that need evaluation for readability are tested in builtin_test.go
	}
	for _, test := range tests {
		h, err := NewTypeHandle(test.tpe)
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

func TestCheckTypeExpressionNegative(t *testing.T) {
	tests := []ast.BaseTerm{
		ast.ApplyFn{MapType, []ast.BaseTerm{name("/foo")}},
		ast.ApplyFn{MapType, []ast.BaseTerm{ast.Number(2), name("/foo")}},
		ast.ApplyFn{StructType, []ast.BaseTerm{name("/foo")}},
		ast.ApplyFn{StructType, []ast.BaseTerm{ast.Number(2), name("/foo")}},
	}
	for _, test := range tests {
		if h, err := NewTypeHandle(test); err == nil { // if NO error
			t.Errorf("NewTypeHandle(%v)=%v succeeded, expected error", h, test)
		}
	}
}

func TestTypeConforms(t *testing.T) {
	tests := []struct {
		left  ast.BaseTerm
		right ast.BaseTerm
		want  bool
	}{
		{ast.NameBound, ast.AnyBound, true},
		{ast.NameBound, ast.NameBound, true},
		{name("/foo"), name("/foo"), true},
		{name("/foo"), ast.NameBound, true},
		{name("/true"), NewUnionType(name("/true"), name("/false")), true},
		{ast.NameBound, name("/foo"), false},
		{NewUnionType(name("/true"), name("/false")), name("/true"), false},
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
			NewStructType(name("/foo"), ast.AnyBound, name("/bar"), ast.NumberBound),
			NewStructType(name("/foo"), ast.StringBound),
			false,
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
		//{
		//	NewFunType(name("/animal"), name("/genus_species")),
		//	NewFunType(name("/animal/bird"), ast.NameBound),
		//	true,
		//},
	}
	for _, test := range tests {
		if got := TypeConforms(test.left, test.right); got != test.want {
			t.Errorf("TypeConforms(%v, %v)=%v want %v", test.left, test.right, got, test.want)
		}
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
		got := UpperBound(test.exprs)
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
		got := LowerBound(test.exprs)
		if got == nil && test.want == nil {
			continue
		}
		if !got.Equals(test.want) {
			t.Errorf("LowerBound(%v)=%v want %v", test.exprs, got, test.want)
		}
	}
}
