# Changelog

## v0.7.0

Released: 2026-10-15

### Type checking fixes

- **Fixed a type checker regression from v0.6.0** (#103): valid list
  arguments to `fn:max`, `fn:min` and `fn:sum` (e.g.
  `X = fn:max([1, 2])`) were rejected with
  `argument has type fn:List(/number), expected /number`. The bounds
  checker now checks the *element type* of the list argument, honoring
  the declared builtin signatures for `fn:max`, `fn:min`, `fn:sum` and
  the `fn:duration:*` and `fn:time:*` variants, including nested
  applications such as `fn:sum([fn:max([1, 2]), 3])`.
- The bounds checker now distinguishes the two reducer modes explicitly,
  instead of guessing from inferred types:
  - as ordinary functions over list values (`fn:max([1, 2])`), and
  - as aggregations in `|> do fn:group_by(...)` transforms, where the
    arguments are variables ranging over the reduced columns.
  This makes the checker agree with the runtime in both directions: a
  list literal as a group_by reducer argument (previously an engine
  panic at evaluation time) and reducing over a list-typed column
  (previously silently produced the empty-group sentinel, e.g.
  `-9223372036854775808`) are now type errors. Float reducers in
  aggregation mode are now typed `/float64` instead of `/any`.
- **Type expression unification repaired** (`unionfind.UnifyTypeExpr`):
  the unification error for constant types was dropped, so e.g.
  `fn:List(/number)` silently unified with `fn:List(/duration)`;
  union type expressions could not unify with their alternatives, which
  made `fn:float:sum`, `fn:float:min` and `fn:float:max` unusable in
  premises; conflicting variable bindings went undetected; and nested
  arity mismatches panicked instead of returning an error. An actual
  type now unifies with a declared union if it unifies with one of its
  alternatives, and an actual union with a declared type if all of its
  alternatives do. The `/any` and `/bot` bounds act as wildcards.
- The `fn:float:sum` signature now expects `.List(/float64)`, matching
  the runtime, which rejects number elements (previously the declared
  signature admitted programs that fail at evaluation time).
- **`:filter` premises are now type-checked**: a type error in the
  function argument (e.g. `:filter(fn:list:contains([1, 2], 'a'))`) or
  a non-boolean argument (e.g. `:filter(fn:string:concat('a', 'b'))`)
  is now rejected instead of passing silently. Unknown types such as
  `/any` are still admitted.

### Correctness fixes

- **factstore**: keep facts that share a hash apart (#102). The
  in-memory stores kept one fact per hash, so a second fact with the
  same hash was silently dropped, and `Contains` reported it as present.
  A constant's hash did not include its type (`p(/abc)`,
  `p("/abc")` and `p(b"/abc")` shared one), and hash pairing overflowed
  64 bits on deeply nested pairs (lists of seven or more items, maps of
  eight or more entries). The stores now keep a slice of facts per hash
  and tell them apart by `Equals`; pairs and lists mix hashes with a
  splitmix64 finalizer.
- **provenance**: return alternative derivations in a stable order
  (#101). `EventsFor` returned derivations in recording order, which
  changes from run to run due to Go's randomized map iteration, so the
  same program produced different proof trees on different runs.
  Derivations are now sorted by rule and then by the facts each one
  rests on.

### Documentation

- Document string comparisons (`:string:starts_with`,
  `:string:ends_with`, `:string:contains`, `:string:matches`) and
  premise evaluation order.

## Earlier releases

This changelog starts with v0.7.0; for earlier releases, see the git
history and tag annotations (`git tag -l -n99`).
