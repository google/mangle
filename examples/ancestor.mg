# This is an example from the Alice Book (Foundations of Databases), Ch. 13
# The ancestor predicate is an example of a non-linear rule.

anc(X, Y) :- par(X, Y).
anc(X, Y) :- anc(X, Z), anc(Z, Y).

par(1, 2).
par(2, 3).
