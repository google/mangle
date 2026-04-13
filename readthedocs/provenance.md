# Provenance: why is this fact true?

When a Datalog program derives a fact, it's often useful to ask *why*: which
rule fired, what values did its variables take, and which stored facts
ultimately supported the derivation? Mangle ships with a provenance tool,
`mgwhy`, and a matching Go package that answers exactly this question.

Given a program and the fact store it produced, `mgwhy` prints a *proof
tree*: the rule that fired, its substitution, and recursively the proofs
of each body premise, bottoming out in stored (EDB) facts.

## A first example

Consider a small graph and its transitive closure:

```
edge(1, 2).
edge(2, 3).
edge(3, 4).

path(X, Y) :- edge(X, Y).
path(X, Z) :- edge(X, Y), path(Y, Z).
```

Save this as `path.mg` and ask why `path(1, 3)` holds:

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

Each node has an identifier of the form `/proof/<hex>` — a 128-bit content
hash of the rule, the derived fact, and the IDs of its sub-proofs.
Identical sub-proofs share an identifier, so when a fact is reused across
several parent derivations, it appears once in the output. Shared
sub-proofs turn the proof tree into a DAG.

## Multiple derivations

A single fact can be derivable in several ways. By default `mgwhy` prints
one proof; ask for more with `-max-proofs`:

```
$ mgwhy -program path.mg -max-proofs 3 'path(1, 4)'
```

## Provenance as queryable data

Proof trees are themselves relational. With `-format facts`, `mgwhy`
emits the proof DAG as Mangle facts that you can load back into `mg` and
query in Datalog:

```
$ mgwhy -program path.mg -format facts 'path(1, 3)' > proof.mg
```

The schema is:

| Predicate | Arity | Meaning |
|---|---|---|
| `proves(ProofID, Fact)` | 2 | `ProofID` concludes `Fact`. |
| `uses_rule(ProofID, RuleID)` | 2 | Which rule fired (omitted for leaves). |
| `premise(ProofID, Index, SubID)` | 3 | The `Index`-th body atom was proved by `SubID`. |
| `edb_leaf(ProofID, Fact)` | 2 | Leaf: `Fact` comes from the store. |
| `absence_leaf(ProofID, Fact)` | 2 | Leaf: `Fact` is **not** in the store (see *Negation* below). |
| `binding(ProofID, VarName, Value)` | 3 | One entry of the rule's substitution σ. |
| `rule_source(RuleID, ClauseText)` | 2 | Human-readable form of a rule. |

Facts are encoded uniformly as **lists**: `edge(1, 2)` becomes the
structured value `[/edge, 1, 2]`. The predicate symbol is prefixed with
`/` so it can live in a name constant.

Once loaded into `mg`, recursive queries on the proof DAG are just
Datalog:

```
# All EDB facts that ultimately support a given derived fact.
supports(P, F) :- proves(P, _), edb_leaf(P, F).
supports(P, F) :- proves(P, _), premise(P, _, Sub), supports(Sub, F).

# Every rule that fired, directly or transitively.
rule_used(P, R) :- proves(P, _), uses_rule(P, R).
rule_used(P, R) :- proves(P, _), premise(P, _, Sub), rule_used(Sub, R).
```

Using a logic language to query its own proofs is pleasingly
self-referential, and is often the quickest way to ask things like "what
sources contributed to this conclusion?"

## Negation

Mangle supports stratified negation: a rule may use a negated premise
`!p(X)` whose predicate `p` is fully computed before the rule runs. The
explainer handles this by recording *evidence of absence*: when `!p(a)` is
checked, the grounded atom is looked up in the final store, and if absent,
the negated premise is satisfied and the proof records an `absence_leaf`
pointing at `p(a)`.

```
node(1).
node(2).
node(3).
reachable(1).
reachable(2).

unreachable(X) :- node(X), !reachable(X).
```

```
$ mgwhy -program graph.mg 'unreachable(3)'
proof 1 of 1:
  unreachable(3)  (/proof/d960840a…)
    by rule …:  unreachable(X) :- node(X), !reachable(X).
    with X=3
    premises:
      node(3)  (/proof/f824ff43…)
        [EDB]
      reachable(3)  (/proof/e1f375a6…)
        [absent: !reachable(3)]
```

This works because stratified evaluation closes the predicate `reachable`
before the rule for `unreachable` runs — so absence in the final store is
sound evidence that the fact is not derivable.

Note: `absence_leaf` nodes do **not** emit a `proves(...)` fact, since
semantically they prove the *negation* of the fact, not the fact itself.
Walk them via `premise(Parent, Index, AbsenceID)`.

## What is *not* explained

The current explainer is a post-hoc backward-chainer — it works from the
finished fact store, without instrumenting the engine. That approach is
cheap when you don't ask for provenance and free when you do, but it has
limits. Specifically, the following are **not** yet explained:

- **Aggregation and transforms** (`|> do …`). A fact produced by
  `fn:sum` or `fn:collect` depends on a whole group of input rows; the
  exact grouping is known to the evaluator but not easy to reconstruct
  after the fact.
- **Temporal annotations** on rule heads or premises.

Rules using these features are skipped, and affected proofs are flagged.
A future "full provenance" mode would instrument the evaluator to emit
derivation events during computation, which would capture these cleanly.

## Caps and performance

Backward chaining re-queries the fact store for each premise. For big
stores or high-fan-out rules, enumerating every possible proof can blow
up, so the explainer has two caps, both exposed on the command line:

- `-max-proofs N` — at most `N` alternative proofs per goal (default 1).
- `-max-depth N` — cut proof depth (default 64). Sub-proofs beyond this
  depth are returned marked `partial`.

Cycles are detected automatically: when the explainer is asked to prove a
fact while already trying to prove it further up the call stack, that
branch yields no proofs. Non-cyclic derivations of the same fact still
succeed.

## Using the Go API

```go
import "codeberg.org/TauCeti/mangle-go/provenance"

proofs, err := provenance.Explain(programInfo, factstore, goal, provenance.Options{
    MaxProofs: 3,
    MaxDepth:  32,
})
if err == provenance.ErrNoProof {
    // goal is not derivable (or only derivable via unsupported features)
}

// Pretty-print a tree:
provenance.Print(os.Stdout, proofs)

// Or write the proof DAG as Mangle facts into another store:
out := factstore.NewSimpleInMemoryStore()
provenance.EmitFacts(proofs, &out)
```
