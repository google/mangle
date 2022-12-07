# Datalog and Mangle data model

![docu badge spec reference](docu_spec_reference.svg)

This section describes the Mangle data model. We start with an informal
discussion of the Datalog data model, and how Mangle extends it.

## Constants and Facts

The Datalog model is based on the terms of first-order logic with relational symbols. The fundamental objects of Datalog are *constant symbols* which identify some object in the domain of discourse.

In Mangle, such constants are called *name constants* and start with a slash `/`, like `/friday`, `/sweet`, `/wood`, `/asdf` or `/true`. Name constants can have multiple parts, like `/person/hilbert` though such a constant remains a single constant, no matter how many "parts" appear in the name.

We assume that constants are *unique* in the sense that distinct constants refer to distinct entities: objects designated by constants 
can only be equal in one way. This assumptions has consequences for assigning meaning of constants, for example if we have `/person/hilbert` and `/mathematician/hilbert` these will refer to distinct objects in the
Datalog and Mangle data model.

Objects can be related to each other. These relationship can be represented as terms, using a *relation symbol* (predicate) 
with a specified number of arguments (arity). A predicate names a
relation and a *fact* indicates that a tuple of arguments is present
in the corresponding relation.

In Mangle, predicates have lower-case names. For example, we can use a unary predicate name `p`, a binary predicate name `loves`, and a three-place predicate `person_thesis_supervisor` to define facts like:
- `p(/asdf)`
- `loves(/hilbert, /topic/mathematics)`
- `person_phdtopic_supervisor(/jacques_herbrand, /topic/math_logic, /ernest_vessiot)`.

This simple data model means we can use facts just as we would use the
rows of a database table.

- `person_phdtopic_supervisor(/jacques_herbrand, /topic/math_logic, /ernest_vessiot)`
- `person_phdtopic_supervisor(/julia_robinson,   /topic/math_logic, /alfred_tarski)`
- `person_phdtopic_supervisor(/raymond_smullyan, /topic/math_logic, /alonzo_church)`

|  person            | phdtopic          | supervisor       |
| ------------------ | ----------------- | ---------------- |
| `/jacques_herbrand`|`/topic/math_logic`| `/ernest_vessiot`|
| `/julia_robinson`|`/topic/math_logic`  | `/alfred_tarski` |
| `/raymond_smullyan`|`/topic/math_logic`| `/alonzo_church` |

In Datalog there are no function symbols. The only thing that
facts can describe are relationships between objects. This
is already enough for simple forms of relational database programming.

## Numbers and Strings

Mangle extends the simple model in a number of ways. Users can
write numeric constants `42` and strings `"hello"` and these constants
can be part of facts.

`question_answer("what is the meaning of life?", 42).`

These values of basic data types are immutable and come with special syntax (literal syntax). We can think of a literal as a special constant.
The use of basic data types without function symbols is limited, but we
defer a discussion of function symbols to later.

## Structured data: pairs, lists, maps, structs

Mangle also adds structured data types, pairs, tuples and lists. 
This involves expressions containing function symbols. Function symbols are
always prefixed with `fn:` to distinguish them syntactically from predicates.
An expression involving a function symbol is *not* by itself a constant but
it is evaluated to some constant that represents the structured data
(or "code"). At the source level, there is no way to express a structured
data constant "directly."

Expressions are not part of the Datalog fragment, they lead to
"new" constants and possibly "new" facts, which may lead to non-termination.
However, for given set of facts and constants (including structured data
constants), any set of recursive rules that only deals with constant
structured data analytically (by accessing information) does not pose a
risk of non-termination.

Given two expressions `First, Second`, a pair is written `fn:pair(First, Second)`. The
`fn:pair` is a function symbol. There are also fixed length tuples,
expressions `fn:tuple(value1, ..., valueN)` for `N` > 2 which correspond
to nested pairs.

A list constant is a finite sequences of values. List constants are
constructed using expressions `fn:list(value1, ..., valueN)`, which
can be conveniently written `[value1, ..., valueN]`, or through applications
of `fn:cons(value, listValue)`. Two lists are equal if they have the same
elements.


A map constant is a mapping from constants of some key type to constants of some
value type. Maps are constructed
with an expression `fn:map(key1, value1, ... keyN, valueN)`, which can
be conveniently written `[ key1 : value1, ..., keyN: valueN ]`. Two maps are
equal (correspond to the same constant) if they have exactly
the same keys and values.

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

