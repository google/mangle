# The following is not a datalog program.
# It uses function symbols and does not terminate.
# We should still be able to type-check it, though.

Decl p(N)
  bound [/number].

p(0).
p(X) :- p(Y), X = fn:plus(Y, 1).
