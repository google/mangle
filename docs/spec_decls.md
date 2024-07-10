# Declarations

![docu badge spec reference](docu_spec_reference.svg)

## Module Declarations

There are two kinds of declarations in a Mangle program that relate to the
module system.

### Package declarations `Package <pkg>!`

A package declaration `Package <pkg>!` that the current source unit belongs to
package named `<pkg>`.

### Uses declarations `Use <pkg>!`

A uses declaration `Uses <pkg>!` that the current source unit can refer to names
from a package named `<pkg>`.

## Type Declarations

Predicate declarations relate to name resolution and the type system.

A predicate declaration starts with `Decl <predicate name>(<Arg1>, ... <ArgN>)`

It is optionally followed by declaration items: descriptor items and bound
declarations.

A declaration always ends with a dot `.`.

### Descriptor items

Descriptor items appear in a list `descr []`.

Each descriptor item has the syntax of an *atom* `<pred>(<arg1>...<argJ>)`.

What follows is a list of builtin descriptor items:

*   `doc(<string>, ... <string>)` A doc item consists of one or more strings
    that describe what the predicate `<pred>` does. It is possible and
    recommended to use a single multi-line strings instead of multiple string
    arguments.
*   `arg(<Arg>, <string>)` describes the purpose of the argument `<Arg>`. All
    arguments need to be described, even if the text is the empty string.

### Bounds declaration

Bounds declaration appears as a list `bounds [ <Bound1>, ... <BoundN> ]`. There
can be zero, one or multiple bounds declarations.

Each bound in a bound declaration follows the syntax of an *expression*: it is
either a name constant or a function `fn:<F>` applied to expression arguments.
Since bounds are type expressions, all names should refer to types and all
functions type-level functions. The only exception to this is the `fn:Singleton`
type-level function, which turns a single name constant into a singleton type.

For each bounds declaration, the number of bounds has to match exactly the
number of arguments.

### First-order type expressions

A type expression is first-order if it does not contain function types or
relation types. What follows is a description of type constants and functions
that can be used to build first-order type expressions.

The following type constants:

*   `/any`
*   `/number`
*   `/float64`
*   `/string`

The following type-level functions, where all arguments are first-order type
expressions:

*   `fn:List(T)` type of list
*   `fn:Pair(S, T)` type of pairs
*   `fn:Tuple(T1, ..., Tn)` with `n >= 3` type of tuple
*   `fn:Map(Key, Value)` type of maps from `Key` type to `Value` type.
*   `fn:Struct(...)` type of structs with labels and types
    *   labels are name constants that specify fields names
    *   pair of consecutive arguments `..., <label>, <type1>, ...` specifies
        that the struct type has the given field with given type
    *   an argument `..., fn:opt(<label>, <type1>), ...` specifies that the
        struct type has an optional field with given label and type.
*   `fn:Singleton(Name)` singleton type for name constant
*   `fn:Union(Type1, ... TypeN)` type that stands for a union of
    `Type1`...`TypeN`
