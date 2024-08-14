# Mangle Datalog

Mangle Datalog is a logic programming language and deductive database system.

It is declarative:
users communicate to the system what they want, not a series of commands.

Data is represented using *facts* and derived with *rules*. A fact combines
several values into a statement, and a rule describes how to compute new
facts from existing facts.

Here is an example describing a family from a Greek tragedy, Antigone:

```cplint
parent(/oedipus, /antigone).
parent(/oedipus, /ismene).
parent(/oedipus, /eteocles).
parent(/oedipus, /polynices).

sibling(Person1, Person2) :-
   parent(P, Person1), parent(P, Person2), Person1 != Person2.
```

The `sibling` rule declaration describes what a sibling is.
We can now ask who Antigone's siblings are:

```cplint
mg >? sibling(/antigone, X)
sibling(/antigone,/eteocles)
sibling(/antigone,/ismene)
sibling(/antigone,/polynices)
Found 3 entries for sibling(/antigone,_).
```

Here `X` is a variable, `? sibling(/antigone, X)` is a query that is asking
what are possible values of `X`.

Mangle extends Datalog with facilities like aggregation, function calls,
static type-checking and type inference.

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
Data Types <datatypes.md>
