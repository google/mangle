package interpreter

import (
	"fmt"
	"testing"

	"github.com/google/mangle/ast"
)

// TestDisplayString tests the DisplayString() method of ast.Constant.
// This method returns a string representation without escaping Unicode characters
func TestDisplayString(t *testing.T) {
	tests := []struct {
		name     string
		constant ast.Constant
		want     string
	}{
		{
			name:     "name constant",
			constant: mustName("/foo"),
			want:     "/foo",
		},
		{
			name:     "simple string constant",
			constant: ast.String("hello"),
			want:     `"hello"`,
		},
		{
			name:     "string with unicode characters",
			constant: ast.String("hello世界"),
			want:     `"hello世界"`,
		},
		{
			name:     "string with special characters",
			constant: ast.String("hello\nworld\ttab"),
			want:     "\"hello\nworld\ttab\"",
		},
		{
			name:     "string with quotes",
			constant: ast.String(`say "hello"`),
			want:     `"say "hello""`,
		},
		{
			name:     "bytes constant",
			constant: ast.Bytes([]byte("hello")),
			want:     `b"hello"`,
		},
		{
			name:     "bytes with unicode",
			constant: ast.Bytes([]byte("hello世界")),
			want:     `b"hello世界"`,
		},
		{
			name:     "number constant",
			constant: ast.Number(42),
			want:     "42",
		},
		{
			name:     "negative number constant",
			constant: ast.Number(-123),
			want:     "-123",
		},
		{
			name:     "float constant",
			constant: ast.Float64(3.14159),
			want:     "3.14159",
		},
		{
			name:     "negative float constant",
			constant: ast.Float64(-2.718),
			want:     "-2.718",
		},
		{
			name:     "pair constant",
			constant: makePair(mustName("/foo"), ast.String("bar")),
			want:     `fn:pair(/foo, "bar")`,
		},
		{
			name:     "empty list",
			constant: ast.ListNil,
			want:     "[]",
		},
		{
			name:     "simple list",
			constant: ast.List([]ast.Constant{mustName("/foo"), ast.String("bar")}),
			want:     `[/foo, "bar"]`,
		},
		{
			name:     "nested list with unicode",
			constant: ast.List([]ast.Constant{ast.String("hello世界"), ast.Number(42)}),
			want:     `["hello世界", 42]`,
		},
		{
			name:     "empty map",
			constant: ast.MapNil,
			want:     "fn:map()",
		},
		{
			name:     "simple map",
			constant: makeMap(map[string]ast.Constant{"foo": ast.String("bar")}),
			want:     `[/foo : "bar"]`,
		},
		{
			name:     "map with unicode values",
			constant: makeMap(map[string]ast.Constant{"greeting": ast.String("hello世界")}),
			want:     `[/greeting : "hello世界"]`,
		},
		{
			name:     "empty struct",
			constant: ast.StructNil,
			want:     "{}",
		},
		{
			name:     "simple struct",
			constant: makeStruct(map[string]ast.Constant{"name": ast.String("John")}),
			want:     `{/name : "John"}`,
		},
		{
			name:     "struct with unicode values",
			constant: makeStruct(map[string]ast.Constant{"message": ast.String("こんにちは")}),
			want:     `{/message : "こんにちは"}`,
		},
		{
			name: "complex nested structure",
			constant: ast.List([]ast.Constant{
				makeStruct(map[string]ast.Constant{
					"name":   ast.String("世界"),
					"value":  ast.Number(42),
					"active": mustName("/true"),
				}),
				ast.String("end"),
			}),
			want: `[{/name : "世界", /value : 42, /active : /true}, "end"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.constant.DisplayString()
			fmt.Println("DisplayString:", got)
			if got != tt.want {
				t.Errorf("DisplayString() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Helper functions

func mustName(symbol string) ast.Constant {
	c, err := ast.Name(symbol)
	if err != nil {
		panic(err)
	}
	return c
}

func makePair(fst, snd ast.Constant) ast.Constant {
	return ast.Pair(&fst, &snd)
}

func makeMap(kvMap map[string]ast.Constant) ast.Constant {
	constMap := make(map[*ast.Constant]*ast.Constant)
	for k, v := range kvMap {
		key := mustName("/" + k)
		val := v
		constMap[&key] = &val
	}
	return *ast.Map(constMap)
}

func makeStruct(kvMap map[string]ast.Constant) ast.Constant {
	constMap := make(map[*ast.Constant]*ast.Constant)
	for k, v := range kvMap {
		key := mustName("/" + k)
		val := v
		constMap[&key] = &val
	}
	return *ast.Struct(constMap)
}
