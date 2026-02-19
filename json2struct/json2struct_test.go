// Copyright 2023 Google LLC
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

package json2struct

import (
	"testing"

	"codeberg.org/TauCeti/mangle-go/ast"
)

func name(n string) ast.Constant {
	c, err := ast.Name(n)
	if err != nil {
		panic(err)
	}
	return c
}

func ptr(c ast.Constant) *ast.Constant {
	return &c
}

func TestJSONtoStruct(t *testing.T) {
	tests := []struct {
		jsonBlob []byte
		want     *ast.Constant
	}{
		{
			jsonBlob: []byte(`{"foo": "bar", "fnum": 3.13, "b": [true, false], "c": { "d": "e" }}`),
			want: ast.Struct(map[*ast.Constant]*ast.Constant{
				ptr(name("/foo")):  ptr(ast.String("bar")),
				ptr(name("/fnum")): ptr(ast.Float64(3.13)),
				ptr(name("/b")):    ptr(ast.ListCons(&ast.TrueConstant, ptr(ast.ListCons(&ast.FalseConstant, &ast.ListNil)))),
				ptr(name("/c")): ast.Struct(map[*ast.Constant]*ast.Constant{
					ptr(name("/d")): ptr(ast.String("e")),
				}),
			}),
		},
	}
	for _, test := range tests {
		got, err := JSONtoStruct(test.jsonBlob)
		if err != nil {
			t.Errorf("JSONtoStruct(%v) failed: %v", test.jsonBlob, err)
			continue
		}
		if !got.Equals(test.want) {
			t.Errorf("JSONtoStruct(%v) = %v, want %v", test.jsonBlob, got, test.want)
		}
	}
}
