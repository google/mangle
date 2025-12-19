package ast

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Escape returns a Mangle source representation of a string.
// In !isBytes mode:
// - newlines are escaped as \n, tags as \t
// - other control characters [0x00..0x20] are byte-escaped,
// - Quote characters " and ' are character-escaped,
// - UTF8 sequences for code points > 0x80 are unicode-escaped.
// In isBytes mode:
// - all characters [0x00..0x20] and [0x80...0xFF] are byte-escaped.
// - quote characters " and ' are byte-escaped,
func Escape(str string, isBytes bool) (string, error) {
	buf := make([]byte, 0, len(str))
	for len(str) > 0 {
		c := str[0]
		if c < utf8.RuneSelf {
			switch c {
			case '\'':
				if isBytes {
					buf = append(buf, `\x27`...)
				} else {
					buf = append(buf, `\'`...)
				}
			case '"':
				if isBytes {
					buf = append(buf, `\x22`...)
				} else {
					buf = append(buf, `\"`...)
				}
			case '\n':
				if isBytes {
					buf = append(buf, `\x0a`...)
				} else {
					buf = append(buf, `\n`...)
				}
			case '\t':
				if isBytes {
					buf = append(buf, `\x09`...)
				} else {
					buf = append(buf, `\t`...)
				}
			case '\\':
				if isBytes {
					buf = append(buf, `\x5c`...)
				} else {
					buf = append(buf, `\\`...)
				}
			default:
				buf = append(buf, byte(c))
			}
			str = str[1:]
			continue
		}
		if isBytes {
			buf = append(buf, `\x`...)
			buf = append(buf, hexdigit(byte(c>>4)))
			buf = append(buf, hexdigit(byte(c&0xf)))
			str = str[1:]
			continue
		}
		r, _, rest, err := unescapeCharPrefix(str, false)
		if err != nil {
			return "", err
		}
		if r == utf8.RuneError {
			if len(str)-len(rest) == 1 {
				return "", fmt.Errorf("invalid UTF-8 encoding")
			}
		}
		str = rest
		buf = append(buf, `\u{`...)
		for j := 2; 0 <= j; j-- {
			b := r >> (j << 3) & 0xff
			buf = append(buf, hexdigit(byte(b>>4)))
			buf = append(buf, hexdigit(byte(b&0xf)))
		}
		buf = append(buf, '}')
	}
	return string(buf), nil
}

// Unescape returns the Mangle source representation of a string and returns a string.
//   - if !isBytes, replaces \x0d \x0a sequences and single \x0d with single \x0a (remove carriage
//     return).
//   - replaces '\” '\"' '\\' with corresponding character
//   - replaces '\n' '\t ' with whitespace character
//   - replaces '\xHH' with byte. Unless `isBytes` is true, checks that result is UTF8 compatible,
//     i.e. in strings only 0x00..0x7F are permitted, in bytestrings anything goes.
//   - replaces escapes \u{hhhhh?h?} with UTF8 byte sequences, checks that result is unicode code point.
func Unescape(s string, isBytes bool) (string, error) {
	if !isBytes {
		s = replaceNewlines.Replace(s)
	}
	if !strings.ContainsRune(s, '\\') {
		return s, nil
	}
	// Otherwise the string contains escape characters.
	buf := make([]byte, 0, 3*len(s)/2)
	for len(s) > 0 {
		c, encode, rest, err := unescapeCharPrefix(s, isBytes)
		if err != nil {
			return "", err
		}
		s = rest
		if c < utf8.RuneSelf || !encode {
			buf = append(buf, byte(c))
		} else {
			var runeTmp [utf8.UTFMax]byte
			n := utf8.EncodeRune(runeTmp[:], c)
			buf = append(buf, runeTmp[:n]...)
		}
	}
	return string(buf), nil
}

// unescapeCharPrefix takes a string input and returns the following info:
//
//	value - the escaped unicode rune at the front of the string, or a byte value if isBytes.
//	encode - the value is valid unicode >x80 and should be unicode-encoded
//	tail - the remainder of the input string.
//	err - error value, if the character could not be unescaped.
func unescapeCharPrefix(s string, isBytes bool) (value int32, encode bool, tail string, err error) {
	// 1. Character is not an escape sequence.
	switch c := s[0]; {
	case c >= utf8.RuneSelf:
		r, size := utf8.DecodeRuneInString(s)
		return r, true, s[size:], nil
	case c != '\\':
		return rune(s[0]), false, s[1:], nil
	}

	// 2. Last character is the start of an escape sequence.
	if len(s) <= 1 {
		err = fmt.Errorf("unable to unescape string, found '\\' as last character")
		return
	}

	c := s[1]
	s = s[2:]
	switch c {
	case '\n':
		value = '\n'
	case 'n':
		value = '\n'
	case 't':
		value = '\t'
	case '\\':
		value = '\\'
	case '\'':
		value = '\''
	case '"':
		value = '"'
	case '`':
		value = '`'

		// Byte escapes
	case 'x':
		if len(s) < 2 {
			err = fmt.Errorf("unable to unescape string")
			return
		}
		var v int32
		hi, ok := unhex(s[0])
		if !ok {
			err = fmt.Errorf("unable to unescape string")
			return
		}
		lo, ok := unhex(s[1])
		if !ok {
			err = fmt.Errorf("unable to unescape string")
			return
		}
		v = hi<<4 | lo
		if !isBytes && v >= utf8.RuneSelf {
			err = fmt.Errorf("unable to unescape string: byte escape not in [0x00..0x7F]: %x", v)
			return
		}
		value = v
		s = s[2:]
		// Unicode escape sequences
	case 'u':
		if s[0] != '{' {
			err = fmt.Errorf("unable to unescape string")
			return
		}
		j := 1
		var v rune
		for s[j] != '}' && j < len(s) && j <= 7 {
			x, ok := unhex(s[j])
			if !ok {
				err = fmt.Errorf("unable to unescape string")
				return
			}
			v = v<<4 | x
			j++
		}
		if s[j] != '}' {
			err = fmt.Errorf("unable to unescape string")
			return
		}
		s = s[j+1:]
		if v > utf8.MaxRune {
			err = fmt.Errorf("unable to unescape string")
			return
		}
		value = v
		encode = true
		// Unknown escape sequence.
	default:
		err = fmt.Errorf("unable to unescape string")
	}

	tail = s
	return
}

func hexdigit(nibble byte) byte {
	if nibble < 10 {
		return byte('0' + nibble)
	}
	return byte('a' - 10 + nibble)
}

func unhex(b byte) (rune, bool) {
	c := rune(b)
	switch {
	case '0' <= c && c <= '9':
		return c - '0', true
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10, true
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}

var replaceNewlines = strings.NewReplacer("\r\n", "\n", "\r", "\n")
