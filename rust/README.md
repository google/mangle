# Mangle (Rust)

[Mangle](https://mangle.readthedocs.io/en/latest/) is a language for
deductive database programming based on Datalog.

The mangle-* Rust packages are a work-in-progress implementation of Mangle in
Rust. The goal is to eventually support everything that the go implementation
provides.

In Datalog, data is represented as facts. A fact combines several values into
a logical statement, like a row in a table. Rules describes in a declarative
way how to obtain new facts from existing facts.

The standard example is the definition of reachable paths in a graph:
given a relation `edge`, we can define it like so:

```
reachable(X, Y) :- edge(X, Y).
reachable(X, Z) :- edge(X, Y), reachable(Y, Z).
```

Mangle adds facilities for aggregation, and also function calls.

Please see https://github.com/google/mangle for more information.
