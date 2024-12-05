# Constructed Types

This section introduces *constructed types* which are data types
where each datum is constructed from a smaller data items.

## Pairs

A *pair datum* (pair) is a combination of two data items.

**Type**. A pair has a type constructed by `fn:Pair(`{math}`S``,` {math}`T``)`
where {math}`S` and {math}`T` are type expressions for the types of the
left and right components of the pair.

**Syntax**.
If {math}`e_1` and {math}`e_2` are data items, then
a pair of these is constructed by `fn:pair(`{math}`e_1, e_2``)`.

**Examples**.

```
fn:pair("web", 2.0)
fn:pair("hello", fn:pair("world", "!"))
```

## Tuples

A *tuple datum* (tuple) is a combination of {math}`n` data items where {math}`n >= 3`

**Type**. A tuple has type `fn:Tuple(`{math}`T_1``,...,` {math}`T_n``)`
where {math}`T_1`...{math}`T_n` are types and {math}`n >= 3`.

**Syntax**.
Let {math}`e_1` ... {math}`e_n` be {math}`n` data components where {math}`n >= 3`,
then a tuple of these components is written `fn:tuple(`{math}`e_1,...,e_n``)`.

**Examples**.

```
fn:tuple("hello", "world", "!")
fn:tuple(1, 2, "three", "4")
```

## Lists

A list datum (list) is a (possibly empty) sequence of data items.

**Type**. A list has type `fn:List(`{math}`T``)` where
{math}`T` is a types.

**Syntax**.

If {math}`e_1` and {math}`e_n` is syntax for elements,
the list of these elements is written `[`{math}`e_1,...,e_n``]`.

The empty list is written `[]`.

A trailing comma is permitted.

**Examples**

```
[]             # empty list
[,]            # also empty list
[0]            # list containing single element 0
[/a, /b, /c]
[/a, /b, /c,]
```

## Maps

A map datum (map) is a mapping from data items to data items.

**Type**. A map has type `fn:Map(`{math}`S``:` {math}`T``)` where
{math}`S` and {math}`T` are types.

**Syntax**.
A map is written `[`...,{math}`k_i``:` {math}`e_i`,...`]`
where
{math}`k_1` ... {math}`k_n` are the keys and 
{math}`e_1` ... {math}`e_n` are the values.

A trailing comma is permitted.

There is no syntax for empty maps, but an empty map can
be constructed with `fn:map()`. This syntax is described in a later section.

**Examples**

```
[/a: /foo, /b: /bar]
[0: "zero", 1: "one",]
[/a: 1,]
```

## Structs

A structure datum (struct) is a record with fields where each field has 
a designated type.

**Type**. A struct has type `fn:Struct({`{math}`name_i``:` {math}`T_i``})`.

**Syntax**.
A struct is written `{`...,{math}`name_i``:` {math}`e_i`,...`}`
where {math}`e_1` and {math}`e_n` are the values.

A trailing comma is permitted.

An empty struct is written `{}`.

**Examples**

```
{}
{,}
{/a: /foo, /b: [/bar, /baz]}
```
