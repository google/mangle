# Temporal Graph Reachability - Point in Time
#
# Demonstrates graph connectivity at distinct points in time.
# The graph structure changes:
# At T1 (2024-01-01): a -> b -> c
# At T2 (2024-01-02): a -> c -> d

Decl link(X, Y) temporal bound [/name, /name].
Decl reachable(X, Y) temporal bound [/name, /name].

# T1: 2024-01-01
link(/a, /b)@[2024-01-01].
link(/b, /c)@[2024-01-01].

# T2: 2024-01-02
link(/a, /c)@[2024-01-02].
link(/c, /d)@[2024-01-02].

# Recursive reachability rule propagates the timestamp
reachable(X, Y)@[T] :- link(X, Y)@[T].
reachable(X, Z)@[T] :- reachable(X, Y)@[T], link(Y, Z)@[T].

# Expected Output:
# reachable(/a, /b) @[2024-01-01T00:00:00Z]
# reachable(/b, /c) @[2024-01-01T00:00:00Z]
# reachable(/a, /c) @[2024-01-01T00:00:00Z]  <-- derived at T1
#
# reachable(/a, /c) @[2024-01-02T00:00:00Z]
# reachable(/c, /d) @[2024-01-02T00:00:00Z]
# reachable(/a, /d) @[2024-01-02T00:00:00Z]  <-- derived at T2
