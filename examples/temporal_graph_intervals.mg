# Temporal Graph Reachability - Intervals
#
# Demonstrates reachability when edges are valid for time intervals.
# Reachability is valid only during the intersection of edge intervals.

Decl link(X, Y) temporal bound [/name, /name].
Decl reachable(X, Y) temporal bound [/name, /name].

# a -> b valid for Jan 1-10
link(/a, /b)@[2024-01-01, 2024-01-10].

# b -> c valid for Jan 5-15
link(/b, /c)@[2024-01-05, 2024-01-15].

# c -> d valid for Jan 12-20
link(/c, /d)@[2024-01-12, 2024-01-20].

# Reachability propagates the intersection of intervals
reachable(X, Y)@[S, E] :- link(X, Y)@[S, E].

# Recursive step: intersection of intervals [S, E] = [S1, E1] intersect [S2, E2]
# Valid if S <= E.

# Case 1: S1 >= S2, E1 <= E2 (S=S1, E=E1)
reachable(X, Z)@[S1, E1] :-
    reachable(X, Y)@[S1, E1], link(Y, Z)@[S2, E2],
    :time:ge(S1, S2), :time:le(E1, E2), :time:le(S1, E1).

# Case 2: S1 >= S2, E2 < E1 (S=S1, E=E2)
reachable(X, Z)@[S1, E2] :-
    reachable(X, Y)@[S1, E1], link(Y, Z)@[S2, E2],
    :time:ge(S1, S2), :time:lt(E2, E1), :time:le(S1, E2).

# Case 3: S2 > S1, E1 <= E2 (S=S2, E=E1)
reachable(X, Z)@[S2, E1] :-
    reachable(X, Y)@[S1, E1], link(Y, Z)@[S2, E2],
    :time:gt(S2, S1), :time:le(E1, E2), :time:le(S2, E1).

# Case 4: S2 > S1, E2 < E1 (S=S2, E=E2)
reachable(X, Z)@[S2, E2] :-
    reachable(X, Y)@[S1, E1], link(Y, Z)@[S2, E2],
    :time:gt(S2, S1), :time:lt(E2, E1), :time:le(S2, E2).

# Expected Output:
# link(/a, /b) @[2024-01-01T00:00:00Z, 2024-01-10T00:00:00Z]
# link(/b, /c) @[2024-01-05T00:00:00Z, 2024-01-15T00:00:00Z]
# link(/c, /d) @[2024-01-12T00:00:00Z, 2024-01-20T00:00:00Z]
#
# reachable(/a, /b) @[2024-01-01T00:00:00Z, 2024-01-10T00:00:00Z]
# reachable(/b, /c) @[2024-01-05T00:00:00Z, 2024-01-15T00:00:00Z]
# reachable(/c, /d) @[2024-01-12T00:00:00Z, 2024-01-20T00:00:00Z]
#
# reachable(/a, /c) @[2024-01-05T00:00:00Z, 2024-01-10T00:00:00Z]
# (Intersection of [Jan 1, Jan 10] and [Jan 5, Jan 15])
#
# reachable(/b, /d) @[2024-01-12T00:00:00Z, 2024-01-15T00:00:00Z]
# (Intersection of [Jan 5, Jan 15] and [Jan 12, Jan 20])
#
# Note: reachable(/a, /d) is NOT derived because [Jan 5, Jan 10] does not overlap with [Jan 12, Jan 20].
