# Datalog and Mangle data model

![docu badge spec reference](docu_spec_reference.svg)

This section describes the Mangle data model. Mangle includes datalog
as a language fragment. This means you can write datalog programs in Mangle,
but you can also write programs that can do more than datalog.

This is an important distinction: as expressed in the
[Rule of Least Power](http://www.w3.org/2001/tag/doc/leastPower), there
is a value in choosing the language with least computational power
(expressivity). Datalog programs are guaranteed to terminate and permit
different strategies for evaluation.

## Constants and Facts

The Datalog model is based on the terms of first-order logic with relation symbols.
Every "object" is referred by an identifier called a *constant symbol*, or
*constants* for short.

Among the user-defined constants are *name constants* or *names*. Names can have
multiple parts, with each part starting with a slash `/` followed by a
non-empty sequence of characters. Examples are `/friday`, or `/person/hilbert`.

Constants are *unique*: distinct constants always refer to distinct objects.
In other words, there is only one built-on notion of equality. This has
consequences for assigning meaning to constants, for example
`/person/hilbert` and `/mathematician/hilbert` will never be equal.

Objects can be related to each other. These relationship can be expressed
using a *predicate symbol* with a specified number of arguments (arity).

Predicates have lower-case names. For example, we can use a unary predicate
name `p`, a binary predicate name `loves`, and a three-place predicate
`phd_supervised_by` to define facts like:
- `p(/asdf)`
- `loves(/hilbert, /topic/mathematics)`
- `phd_supervised_by(/jacques_herbrand, /topic/math_logic, /ernest_vessiot)`.

An `atom` is a predicate symbol applied to the right number of arguments.
If all arguments are constant symbols, the atom is called a `fact`.
It can help consider facts as rows (tuples) of a database table.

- `phd_supervised_by(/jacques_herbrand, /topic/math_logic, /ernest_vessiot)`
- `phd_supervised_by(/julia_robinson,   /topic/math_logic, /alfred_tarski)`
- `phd_supervised_by(/raymond_smullyan, /topic/math_logic, /alonzo_church)`

|  person            | phdtopic          | supervisor       |
| ------------------ | ----------------- | ---------------- |
| `/jacques_herbrand`|`/topic/math_logic`| `/ernest_vessiot`|
| `/julia_robinson`|`/topic/math_logic`  | `/alfred_tarski` |
| `/raymond_smullyan`|`/topic/math_logic`| `/alonzo_church` |

## Numbers and Strings

Numbers and strings like `42` and strings `"hello"` are also considered
constant symbols and can be part of facts.

`question_answer("what is the meaning of life?", 42).`

## Dates

Calendar dates are represented as first-class constants using the `/date`
type bound.  A date constant renders as an ISO-8601 string prefixed by `@`, for
example `@2023-10-06`.  The representation tracks only the calendar day (no
time of day), and uses the UTC calendar to ensure arithmetic is stable across
time zones.

Dates are commonly constructed with the built-ins described in
[`spec_builtin_operations.md`](spec_builtin_operations.md):

- `fn:date:from_string("2023-10-06")` parses ISO-8601 strings.
- `fn:date:from_parts(2023, 10, 6)` creates a date from its numeric
  year/month/day components.
- `fn:date:to_string(@2023-10-06)` returns the ISO representation as a string.

The resulting constants can be stored in facts, compared for equality, and used
anywhere a `/date` value is permitted by a predicate declaration.

## Structured data: pairs, lists, maps, structs

Structured data is constructed with function symbols. Any program that
contains a function symbol is not part of the datalog fragment.

Function symbols are always prefixed with `fn:` to distinguish them
syntactically from predicates. An expression involving a function symbol is
*not* by itself a constant but it is *evaluated* to some constant.

It is this constant that represents the structured data. At the source level, there
is no way to express a structured data constant "directly."
(TODO: The evaluation is very different from applying rules.)

Mangle also adds structured data types, pairs, tuples, lists, maps, structs.
The syntax is:

* `fn:pair(<fst>, <snd>)` for pairs
* `fn:tuple(<arg1>, ..., <argN>)` for tuples, with `N >= 3`
* `[ <elem1>, ... <elemN> ]` for lists
* `[ <key>: <value>, ... <key>: <value> ]` for maps
* `{ <label>: <value>, ... <label>: <value> }` for structs

For specifying types of arguments, users need to enter a
predicate [declaration](spec_decls.md). This is where the type
expressions can be used to "upper bound" for the argument.
