# Mangle (Rust)

This is an implementation of the Mangle language in Rust.

Mangle is a language for deductive database programming based on Datalog.

In Datalog, it is possible to define relations using recursive programs that query other relations.

The obligatory example is the definition of reachable paths in a graph: given a relation `edge`, we
can define it like so:

```
reachable(X, Y) :- edge(X, Y).
reachable(X, Z) :- edge(X, Y), reachable(Y, Z).
```

Mangle adds facilities for aggregation, and also funcion calls.

Please see https://github.com/google/mangle for more information.

