# Basic Types

This section describes basic Mangle data types, through syntax,
type expressions, examples.

We do this by defining words in a technical, precise sense. For readability,
we also introduce names from everyday language. So "name datum" is
shortened to "name", and similarly for other data types.

## Names

A *name datum* (name) refers to an entity or object in the domain of
discourse.

**Syntax.** A name consists of one or several name *parts*.

A name part always starts with a slash `/` followed by non-empty
sequence of characters:
 * letters `A`..`Z` | `a`..`z`
 * digits `0`..`9`
 * punctuation characters `.` | `-` | `_` | `~` | `%`

Through the use of [percent-encoding](https://en.wikipedia.org/wiki/Percent-encoding)
it is possible to encode many other characters.

Names play a role similar to uniform resource locators
([URL](https://en.wikipedia.org/wiki/URL)s) for the world wide web.

**Examples.**

```
/a
/test12
/antigone

/crates.io/fnv
/home.cern/news/news/computing/30-years-free-and-open-web
```

**Type**. A name has type `/name`.

**Unique name assumption.** Two distinct names are assumed to always refer to
distinct objects. Names are never considered equal to objects of other types.

This means, the only built-in notion of equality is syntactic.
This assumption has deep consequences for the meaning of Mangle programs.

## Numbers

An *integer number datum* (number) is a number between
-(2^63-1)-1 and 2^63-1. In other words, a 64-bit signed integer in
two's complement representation.

**Type**. A number has type `/number`.

**Examples**.

```
0
1
128
-10000
```

## Floating-point Numbers

A *floating-point number datum* (float) is a 64-bit floating point number.

**Type**. A float has type `/float64`.

**Examples**.

```
3.141592
-10.5
```

## Strings

A *string datum* (string) is a sequence of Unicode characters
in the UTF-8 encoding.

**Type**. A string has type `/string`.

Two strings are equal if they have the same sequence of bytes. Note that
this is more fine-grained than the Unicode standard concept of canonical
equivalence and compatibility.

Strings can be written in single or double quotes.

```
"foo"
'foo'
"something 'quoted'"
'something "quoted"'
```

The textual representation of a string in a program is a string literal.
String literals can contain *escape sequences*:

* `\'`: single quote character
* `\"`: double quote character
* `\n` and `\`/newline/: newline character
* `\t`: tab character
* `\\`: backslash character
* `\x`*hh*: unicode character with given code specified by 2 hexadecimal digits *hh*
* `\u{`*hhhhh?h?*`}`: unicode character with 4-6 hexadecimal digits *hhhhh?h?*

**Examples**.
```
"something \"quoted\" with escapes."
'A single quote \' surrounded by single quotes'
"A single quote \' surrounded by double quotes"
"A double quote \" surrounded by double quotes"
"A newline \n"
"A tab \t"
"Java class files start with \xca\xfe\xba\xbe"
"The \u{01f624} emoji was originally called 'Face with Look of Triumph'"
```

**Multi-line strings** are supported using backticks

```
`
I write, erase, rewrite

Erase again, and then

A poppy blooms.
`
```

## Byte strings

A *byte string datum* (byte string) stands for a sequence of arbitrary bytes.

**Type**. A byte string has type `/bytes`.

Byte strings can be written as a string literal with a `b` prefix.
Characters in a byte string literal are going to be UTF-8 encoded.

Two byte strings are equal when their byte sequences are equal. A
byte string is never equal to a string, even both data items would
have the same byte-sequences.

**Examples**.
```
b"A \x80 byte carries special meaning in UTF8 encoded strings"
b"\x80\x81\x82\n"
b"😤",
```

The next-to-last example represented as the byte sequence \x80\x81\x82\0a.
It is not possible to write such a sequence in valid UTF-8 encoding.

The last example is represented as the byte sequence \xf0\x9f\x98\xa4 which
is the UTF-8 encoding of \u{01f624}.

