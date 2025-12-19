package ast

import (
	"testing"
)

func TestRoundtrip(t *testing.T) {
	tests := []struct {
		name  string
		str   string
		bytes bool
		want  string
	}{
		{
			name:  "string constant",
			str:   "h'e\"l\nl\to",
			bytes: false,
			want:  `h\'e\"l\nl\to`,
		},
		{
			name:  "bytestring constant",
			str:   "h'e\"l\nl\to\\",
			bytes: true,
			want:  `h\x27e\x22l\x0al\x09o\x5c`,
		},
		{
			name:  "string constant unicode face with steam from nose",
			str:   "\U0001f624",
			bytes: true,
			want:  `\xf0\x9f\x98\xa4`,
		},
		{
			name:  "string constant unicode face with steam from nose",
			str:   "\U0001f624",
			bytes: false,
			want:  `\u{01f624}`,
		},
		{
			name:  "byte string non utf8 compatible",
			str:   string([]byte{0x80, 0x81}),
			bytes: true,
			want:  `\x80\x81`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Escape(test.str, test.bytes)
			if err != nil {
				t.Errorf("Escape(%v, %v) failed with %v", test.str, test.bytes, err)
			} else if got != test.want {
				t.Errorf("Escape(%q, %v) = %v want %v", test.str, test.bytes, got, test.want)
			}

			if unesc, err := Unescape(got, test.bytes); err != nil {
				t.Errorf("Unescape(Escape(%q), %v) failed with %v", test.str, test.bytes, err)
			} else if unesc != test.str {
				t.Errorf("Unescape(Escape(%q), %v) = %v want %v", test.str, test.bytes, unesc, test.str)
			}
		})
	}
}

func TestBad(t *testing.T) {
	tests := []struct {
		name  string
		str   string
		bytes bool
		want  string
	}{
		{
			name:  "invalid byte escape in string constant",
			str:   `hello \x80`,
			bytes: false,
		},
		{
			name:  "invalid unicode escape in string constant",
			str:   `bogus \u{2affff}`,
			bytes: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Unescape(test.str, test.bytes)
			if err == nil {
				t.Errorf("Unescape(%v, %v) = %v, was supposed to fail.", test.str, test.bytes, got)
			}
		})
	}
}

func TestEscapeInvalidUTF8(t *testing.T) {
	// 0xff is invalid in UTF-8
	invalidStr := string([]byte{0xff})
	got, err := Escape(invalidStr, false)
	if err == nil {
		t.Errorf("Escape(%q, false) expected error, got nil. Result: %q", invalidStr, got)
	}
}
