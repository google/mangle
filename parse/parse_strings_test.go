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

package parse

import (
	"testing"

	"github.com/google/mangle/ast"
)

func TestParseStringConstant(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want ast.Term
	}{
		{
			name: "string with newline",
			str: `"hello\

world"`,
			want: ast.String("hello\n\nworld"),
		},
		{
			name: "string constant, double quote",
			str:  `"hello"`,
			want: ast.String("hello"),
		},
		{
			name: "bytestring constant, double quote",
			str:  `b"hello"`,
			want: ast.Bytes([]byte("hello")),
		},
		{
			name: "string constant, single quote",
			str:  `'hello'`,
			want: ast.String("hello"),
		},
		{
			name: "string constant, single quote containing double quote",
			str:  `'he"llo'`,
			want: ast.String(`he"llo`),
		},
		{
			name: "string constant escaped",
			str:  `"he\\ll\"o"`,
			want: ast.String(`he\ll"o`),
		},
		{
			name: "string byte and unicode escape",
			str:  `"\x68\u{0065}ll\"o"`,
			want: ast.String(`hell"o`),
		},
		{
			name: "byte string byte escape",
			str:  `b"\x80\x81\x82\n"`,
			want: ast.Bytes([]byte{0x80, 0x81, 0x82, 0x0A}),
		},
		{
			name: "byte string byte escape",
			str:  `b"\x80\x81\x82😤"`,
			want: ast.Bytes([]byte{0x80, 0x81, 0x82, 0xf0, 0x9f, 0x98, 0xa4}),
		},
		{
			name: "string emoji",
			str:  `"\u{01f624}"`,
			want: ast.String("😤"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			term, err := Term(test.str)
			if err != nil {
				t.Errorf("Term(%v) failed with %v", test.str, err)
			} else if term == nil {
				t.Errorf("Term(%v) = nil", test.str)
			} else if !term.Equals(test.want) {
				t.Errorf("Term(%q) = %v (%T) want %v (%T) ", test.str, term, term, test.want, test.want)
			}

			c, ok := term.(ast.Constant)
			if !ok && c.Type != ast.StringType && c.Type != ast.BytesType {
				t.Fatalf("Term(%q) is not a constant?", test.str) // cannot happen
			}
			// Convert to a (possibly different, but equivalent) string representation.
			tmp := c.String()
			newterm, err := Term(tmp)
			if err != nil {
				t.Errorf("Term(%q.String()) failed with %v, tmp = %v", test.str, err, tmp)
			}
			if !term.Equals(newterm) {
				t.Errorf("Term(%q) = %v differs from Term(%q.String()) = %v, tmp = %v ", test.str, term, test.str, newterm, tmp)
			}
		})
	}
}
