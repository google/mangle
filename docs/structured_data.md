# Structured Data

![docu badge spec explanation](docu_spec_explanation.svg)

The [datamodel](spec_datamodel.md) mentions expressions function symbol.
Expressions involving functions are *not* part of Datalog. Functions can
construct "new" constants and possibly "new" facts, which may lead to
non-termination. 

However, expressions are useful and for given fixed set of facts involving
constants (including structured data), any set of recursive rules that only
does analytic processing (access information, takes things apart) does not
pose a risk of non-termination. (TODO: This statement needs a proof).

## Pair and tuple constants

Given two expressions `First, Second`, a pair is written `fn:pair(First, Second)`. The
`fn:pair` is a function symbol. There are also fixed length tuples,
expressions `fn:tuple(value1, ..., valueN)` for `N` > 2 which correspond
to nested pairs.

## List constants

A list constant is a finite sequences of values. List constants are
constructed using expressions `fn:list(value1, ..., valueN)`, which
can be conveniently written `[value1, ..., valueN]`, or through applications
of `fn:cons(value, listValue)`. Two lists are equal if they have the same
elements.

## Map constants

A map constant is a mapping from constants of some key type to constants of some
value type. Maps are constructed
with an expression `fn:map(key1, value1, ... keyN, valueN)`, which can
be conveniently written `[ key1 : value1, ..., keyN: valueN ]`. Two maps are
equal (correspond to the same constant) if they have exactly
the same keys and values.

## Struct constants

A struct constant is a mapping from field names to constants. Structs are
constructed with a expression `fn:struct(key1, value1, ... keyN, valueN)`,
which can be conveniently written `{ key1 : value1, ..., keyN : valueN }`. 
Two structs are equal (correspond to the same constant) if they have exactly
the same fields, and the value for each field is equal.

As a combined example, here is a fact that involves a single structured
data constant, which happens to be a list composed of three structs. 

`triangle_2d([ { /x: 1,  /y:  2 },
               { /x: 5, /y: 10 },
               { /x: 12,  /y:  5 } ])`

