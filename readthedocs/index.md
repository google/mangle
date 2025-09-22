# Mangle Datalog

Mangle Datalog is a logic programming language and deductive database system.

Data is represented as *facts*. A fact combines several values into a
logical statement, like a row in a table. *Rules* describes how to compute new
facts from existing facts.

* Mangle is declarative: users communicate to the system what they want, not a series of commands.
* values can be structured data types like lists, maps, records (structs).
* programs can be split into modules, benefit from static type checking and
type inference.
* it offers support for more complex queries that involve aggregation and function calls.

## A first example

Here is an example describing a family from a Greek tragedy, Antigone:

```cplint
parent(/oedipus, /antigone).
parent(/oedipus, /ismene).
parent(/oedipus, /eteocles).
parent(/oedipus, /polynices).

sibling(Person1, Person2) ⟸
   parent(P, Person1), parent(P, Person2), Person1 != Person2.
```

We state a few `parent` facts that describe the state of the world.
The `sibling` rule describes when two people are siblings.
We can now load this program and ask who Antigone's siblings are:

```cplint
mg >? sibling(/antigone, X)
sibling(/antigone,/eteocles)
sibling(/antigone,/ismene)
sibling(/antigone,/polynices)
Found 3 entries for sibling(/antigone,_).
```

Here `X` is a variable, `? sibling(/antigone, X)` is a query asking
for all possible values of `X`.

Documentation is under construction. In the meantime, please see the docs
in the [documentation](https://github.com/google/mangle/blob/main/docs/README.md)
and [examples directory](https://github.com/google/mangle/tree/main/examples)
for more information.

## Table of contents

```{toctree}
---
maxdepth: 2
---
Installing <installing.md>
Getting to know Datalog <datalog.md>
Aggregation <aggregation.md>
Basic Types <basictypes.md>
Constructed Types <constructedtypes.md>
Type Expressions <typeexpressions.md>
