# Relational algebra

![docu badge spec explanation](docu_spec_explanation.svg)

When learning datalog, one may come across a statement like 
"datalog can be translated to SQL." Since SQL is supported by many
relational database management systems, but SQL is also a rich language
support aggregations, it is worthwhile to understand what is really
meant.

This section informally explains the connection between non-recursive datalog
programs and relational algebra. Relational algebra is foundation for
relational calculus and SQL that mathematical explanation for the operations
of selection(filtering), projection, joining. For details, please refer to
Ullman's ["Principles of..." book](bibliography.md).

## Relational algebra

[Relational algebra](https://en.wikipedia.org/wiki/Relational_algebra) is a
mathematical notation for expressing simple computations on relations. In this
section, we discuss relations as sets of tuples, though in database literature
it is more common to treat a table with named columns. Relational algebra
expressions are built up from relational variables and the following operations:

*   `union(A, B)` the relation that contains a tuple if it is in `A` or in `B`
    (or both)
*   `set_difference(A, B)` the relation that contains all tuples of `A` that are
    not in `B`
*   `cartesian_product(A, B)` the relation that contains all combinations `t_1`,
    `t_2` of tuples `t_1` of `A` and `t_2` of `B`
*   `project_I(A)` the relation that contains tuples of `A` with some components
    removed (`I` specifies what to retain)
*   `select_F(A)` the relation that contains tuples matching some condition `F`

These are subject to some restrictions (e.g. `union` and `set_difference` are
only applied to relations that have the same width). One can define more
operations from these basic ones, for example:

*   `intersection(A, B)` as `set_difference(A, set_difference(A, B))`
*   `join(A, B)` as `project_I(select_F(cartesian_product(A, B)))` where `F` is
    the the join condition and `I` retains the relevant columns.

## From datalog rules to relational algebra

### The relation of a rule

We first need to define a concept of *relation of a datalog rule*.
For a rule *r* with subgoals `S_1, ..., S_n` containing variables `X_1, ..., X_m`,
the *relation of r* is the set of *m*-tuples `c_1, ..., c_m` that match
the subgoals, meaning that the substitution mapping each `X_i` to `c_i` makes
all subgoals true.

### Translating the body of a single rule.

This section sketches a translation from positive datalog rules to relational
algebra expression.

**Input**: a single datalog rule with subgoals `S_1, ..., S_n`

**Output**: an expression that computes from relations `R_1, ..., R_n` a
relation `R(X_1, ..., X_m)` containing all tuples matching `S_1, ..., S_n`.

For each `S_i` that is a positive atom, we build an expression `Q_i` that
operates on the corresponding relational variable `R_i`. We set `Q_i` to the
relational algebra expression `project_V(select_F(R_i))` like this:

1.  Collect all variables from arguments of `S_i` into a tuple of variables *V*.
1.  Build up a formula *F*:
    *   whenever there are two positions `j, k` containing the same variable,
        add an equation `@j = k`.
    *   whenever a position `j` contains a constant `/c`, add an equation `@j =
        /c`

Some of the variables `X` may be bound not by atoms but by equations `X = e`
where `e` is a constant or a variable bound by one of the other subgoals. For
these, we define an expression `D_X`:

*   If `e` is a constant `/c` we set `D_X` to `{ /c }`
*   If `e` is a variable `Y` bound as j-th argument of `S_i`, we set `D_X` to
    `project_j(R_i)`.

Let `Q` be the join of all `Q_i` and `D_X`. The resulting expression is
`select_F(Q)` where the condition `F` is built from all subgoals `S_i` that are
equations/inequations/comparisons.

Finally, for every negated atoms `!q(X...)`, we apply the above
construction `q(X...)` to obtain a positive relational algebra expression `N`
and then apply `set_difference(Q, N)`.

### Rectifying rules and translating all rules

In order to apply the translation to Datalog program, rules have to be
*rectified*, meaning: if the head of a rule mentions constants, or repeats a
variable, this is rewritten by introducing a fresh variable and an equation.

For example the rule

```
knows(X, X) :- person(X).
```

would be replaced with

```
knows(X, Y) :- person(X), Y = X.
```

When all rules are rectified, the relational expression for an intensional
predicate `p` is obtained by all its rules and taking the `union`. The
translation proceeds inductively: one starts with predicates that refer
only to extensional predicates (call that "level 0"). Then, for each level *i*,
one proceeds with predicates that refer to predicates whose
expressions were already computed at previous levels *j < i*.

## From relational algebra expressions to datalog rules

We now sketch the other direction. We assume that selection conditions have been
simplified so that all boolean combinations are pushed down into comparison of
individuals (negation normal form followed by exchanging comparison operators.)

We can then show that every function expressible in relational algebra is
expressible as a nonrecursive positive datalog program by induction on structure
of expression:

*   base case: we have a single relational variable or set of tuples. Then we
    just use the existing relation or invent a predicate for a fixed set of
    tuples.

*   case `union`: by ind. hyp, have two pred. `q1` and `q2`. Define a predicate
    with two rules
    
```
p(X...) :- q1(X...).
p(X...) :- q2(X...).
```

*   case `set_difference`: by ind. hyp, we have two pred. `q1` and `q2`. Define a
    predicate with a rule
    
```
p(X...) := q1(X...), !q2(X...).
```

*   case `project_I`: by ind. hyp, we have pred `q`. Define a predicate with a rule
    as follows, where `_` is used in those places that are not in the
    specification of retained columns `I`.

```
p(X...) = q(X..., _...).
``` 

*   case `cartesian_product`: by ind. hyp, we have two pred. `q1` and `q2`. Define a
    predicate with a rule that uses different variables for the arguments of
    `q1` and `q2`.
    
```
p(X..., Y...) :- q1(X...), q2(Y...).
```

*   case `select_F`: by ind. hyp, we have pred `q`. Define a predicate with a rule
    that builds datalog goals `F*` for the conjunction of datalog
    equality/inequality/arithmetic built-in predicates corresponding to `F`

```
p(X...) = q(X...), F*
```

