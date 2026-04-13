# Fact provenance

Given a Mangle program and a populated fact store, the `provenance` package
and the `mgwhy` command answer the question: *why does this derived fact
hold?* The answer is a proof tree (or DAG): the rule that fired, the
substitution that satisfied it, and recursively the proofs of its body
premises, bottoming out at stored (EDB) facts.

## Scope and limits

This is **simple provenance** — a post-hoc explainer that backward-chains
against the final fact store. It handles:

- positive Datalog (rules whose body consists of atoms),
- equality (`=`) and inequality (`!=`) premises,
- **stratified negation** via closed-world absence (see below),
- recursion (with cycle detection),
- multiple alternative derivations of the same fact.

It does **not** yet handle:

- aggregation / transforms (`|> do …`) — rules using them are skipped,
- temporal annotations on heads or premises.

For those, **full provenance** is deferred future work: it would
instrument the evaluator to emit derivation events during computation,
giving first-class access to the substitutions and grouped rows that
bulk-emit operators like aggregation consume.

### How negation is handled

Mangle uses stratified evaluation, so by the time a rule containing
`!p(a)` is evaluated, the stratum defining `p` has run to fixpoint. That
means *absence in the final fact store is sound evidence of
non-derivability*.

When the explainer encounters `!p(a)` in a rule body under the current
substitution, it grounds the negated atom and checks the store:

- If `p(a)` is **absent**, the negated premise is satisfied and the proof
  records a leaf of kind `KindAbsence` (predicate `absence_leaf` in the
  emitted schema) pointing at `p(a)`.
- If `p(a)` is **present**, the rule cannot fire under this substitution
  and the explainer backtracks.

This is the "bare absence" flavor of negative proof — it asserts
non-derivability without enumerating the failed rules that would have had
to derive it. A richer version that witnesses *why* each candidate rule
fails is possible but is not in simple provenance.

## Using the `mgwhy` command

```
mgwhy -program PROG.mg [-facts STORE.sc[.gz|.zst]]
      [-format tree|facts] [-max-proofs N] [-max-depth N] GOAL
```

The program is parsed and evaluated against the fact store (or an empty one
if `-facts` is omitted), then `GOAL` is explained. `GOAL` must be a ground
atom.

### Example

Program `path.mg`:

```
edge(1, 2).
edge(2, 3).
edge(3, 4).
path(X, Y) :- edge(X, Y).
path(X, Z) :- edge(X, Y), path(Y, Z).
```

Ask why `path(1, 3)` holds:

```
$ mgwhy -program path.mg 'path(1, 3)'
proof 1 of 1:
  path(1,3)  (/proof/763e3b51b4680148b139172de260f3ff)
    by rule /rule/b23b52ec…:  path(X,Z) :- edge(X,Y), path(Y,Z).
    with X=1, Y=2, Z=3
    premises:
      edge(1,2)  (/proof/24541127…)
        [EDB]
      path(2,3)  (/proof/86b33a99…)
        by rule /rule/e844ff77…:  path(X,Y) :- edge(X,Y).
        with X=2, Y=3
        premises:
          edge(2,3)  (/proof/e5f6834a…)
            [EDB]
```

Ask for up to three alternative proofs:

```
$ mgwhy -program path.mg -max-proofs 3 'path(1, 3)'
```

## Provenance as queryable Mangle facts

With `-format facts`, `mgwhy` emits the proof DAG as Mangle facts. You can
pipe them back into `mg` to query the proof itself in Datalog.

### Schema

| Predicate | Arity | Meaning |
|---|---|---|
| `proves(ProofID, Fact)` | 2 | `ProofID` concludes `Fact`. |
| `uses_rule(ProofID, RuleID)` | 2 | Derived proofs: which rule fired. |
| `premise(ProofID, Index, SubID)` | 3 | The `Index`-th premise was proved by `SubID`. |
| `edb_leaf(ProofID, Fact)` | 2 | Terminal leaf; `Fact` comes from the store. |
| `absence_leaf(ProofID, Fact)` | 2 | Terminal leaf; `Fact` is **not** in the store (closed-world proof of a negated premise). |
| `binding(ProofID, VarName, Value)` | 3 | One entry of the rule's substitution σ. |
| `rule_source(RuleID, ClauseText)` | 2 | Human-readable form of a rule. |

Note: `absence_leaf` nodes do **not** emit a `proves(...)` fact —
semantically they prove the *negation* of `Fact`, not `Fact` itself. Walk
them via `premise(Parent, Index, AbsenceID)`.

Facts are encoded uniformly as **lists**: `edge(1, 2)` becomes the
structured value `[/edge, 1, 2]`. The predicate symbol is prefixed with `/`
so it can be represented as a name constant.

Proof and rule identifiers are **128-bit content hashes**, encoded as hex:
`/proof/<32-hex-chars>` and `/rule/<32-hex-chars>`. Identical sub-proofs
share an identifier automatically, turning the forest into a DAG.

### Example

```
$ mgwhy -program path.mg -format facts 'path(1, 3)' > proof.mg
```

Then in `mg`:

```
$ mg
::load proof.mg

# Which EDB facts ultimately support path(1,3)?
supports(P, F) :- proves(P, _), edb_leaf(P, F).
supports(P, F) :- proves(P, _), premise(P, _, Sub), supports(Sub, F).

? supports(/proof/…path-1-3…, F)
# → F = [/edge, 1, 2]
# → F = [/edge, 2, 3]

# Which rules fired, directly or transitively?
rule_used(P, R) :- proves(P, _), uses_rule(P, R).
rule_used(P, R) :- proves(P, _), premise(P, _, Sub), rule_used(Sub, R).
```

## Using the Go API

```go
import (
    "codeberg.org/TauCeti/mangle-go/provenance"
)

proofs, err := provenance.Explain(programInfo, factstore, goal, provenance.Options{
    MaxProofs: 3,
    MaxDepth:  32,
})
if err == provenance.ErrNoProof {
    // Goal is not derivable (or only derivable via unsupported features).
}

// Pretty-print:
provenance.Print(os.Stdout, proofs)

// Or emit to a factstore:
out := factstore.NewSimpleInMemoryStore()
provenance.EmitFacts(proofs, &out)
```

### Options

- `MaxProofs` (default 1) — cap on alternative proofs per goal. Pass a
  large value if you want all derivations.
- `MaxDepth` (default 64) — recursion cutoff. Proofs deeper than this are
  returned with the `Partial` flag set.

### Proof identity

`ProofNode.ID` is `/proof/<128-bit-content-hash>` — derived from the rule,
the goal, and the (ordered) IDs of sub-proofs. For EDB leaves it's hashed
from the fact alone. This makes IDs **stable across runs** and makes shared
sub-proofs automatically share a single node, bounding the size of the
output.

Collision probability at 128 bits is negligible for any realistic proof
graph (birthday bound ≈ 2⁶⁴).

## When provenance is expensive

Backward chaining re-queries the fact store for each premise. For large
stores or high-fan-out rules, enumerating *all* proofs can blow up quickly.
Start with `MaxProofs=1` (the default) and only raise it when you need
alternatives.

Cycles are cut automatically: if the explainer is asked to prove a fact
while already trying to prove it further up the stack, that branch returns
no proofs (so the cycle does not contribute infinite expansions but the
original proof via a non-cyclic rule still succeeds).
