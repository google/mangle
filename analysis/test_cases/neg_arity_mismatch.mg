p(1,2,3).

# This should fail because we call p with 2 arguments but it has arity 3.
q(X,Y) :- p(X,Y).