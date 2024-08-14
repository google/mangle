# Data Types

## Names

Names are identifiers. They refer to some entity or object.

Name can also be structured, in the sense that they can
have several non-empty parts separated by `/`. They are similar to URLs
but without the protocol.

```
/a
/test12
/antigone

/crates.io/fnv
/home.cern/news/news/computing/30-years-free-and-open-web
```

Two distinct names always refer to distinct objects ("unique name assumption").
The objects they refer to are also always different from other constants,
e.g. numbers, strings.

## Integer Numbers

The `/number` type stands for 64-bit signed integers.

```
0
1
128
-10000
```

## Floating point Numbers

The `/float64` type stands for 64-bit floating point numbers.

```
3.141592
-10.5
```

## Strings

The `/string` type stands for strings (sequences of characters).

Strings can be written in single or double quotes.

```
"foo"
'foo'
"something 'quoted'"
'something "quoted"'
```

Strings can contain escapes:

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

Multi-line strings are supported using backticks

```
`
I write, erase, rewrite

Erase again, and then

A poppy blooms.
`
```

## Byte strings

The `/bytes` type stands for sequences of arbitrary bytes.

Byte strings can still be written as literals.

```
b"A \x80 byte carries special meaning in UTF8 encoded strings"
```
