Decl shortest_path(Source, Target, Path)
  descr [
    fundep([Source, Target], [Path]),
    merge([Path], "shorter")
  ].

Decl shorter(P1, P2, ShorterPath)
  descr[
    mode('+', '+', '-'),
    deferred(),
  ].

edge(/a, /b).
edge(/b, /c).
edge(/c, /d).
edge(/a, /d).

# The all_paths relation contains "all paths."

all_paths(X, Y, [X, Y]) :-
  edge(X, Y).
all_paths(X, Z, NewPath) :-
	all_paths(X, Y, Path), edge(Y, Z)
  |> let NewPath = fn:list:append(Path, Z).

# The shortest path relation contains only the shortest paths.
# The definition is exactly the same as all_paths, but the declaration
# defines a merge predicate that will remove the longer paths.

shortest_path(X, Y, [X, Y]) :-
  edge(X, Y).
shortest_path(X, Z, NewPath) :-
	shortest_path(X, Y, Path), edge(Y, Z)
  |> let NewPath = fn:list:append(Path, Z).

shorter(P1, P2, ShorterPath) :-
  fn:list:len(P1) < fn:list:len(P2),
  ShorterPath = P1.

shorter(P1, P2, ShorterPath) :-
  fn:list:len(P2) <= fn:list:len(P1),
  ShorterPath = P2.

# Interpreter session:

# mg >?all_paths(/a,/d,_)
# all_paths(/a,/d,[/d, /a])
# all_paths(/a,/d,[/d, /c, /b, /a])
# Found 2 entries for all_paths(/a,/d,_).

# mg >?shortest_path(/a, /d, _)
# shortest_path(/a,/d,[/d, /a])
# Found 1 entries for shortest_path(/a,/d,_).
