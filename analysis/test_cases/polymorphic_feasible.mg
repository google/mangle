Decl o(Owner) bound[/any].

foo(X) :-
  o(O),
  #
  # {} | {} | {O: /any} |- O : /any
  #
  :match_pair(O, A, X),
  #
  # {?X1, ?X2} | {} | {O: Pair(?X1,?X2), A: X1, X: X2}
  #
  :string:contains(X, "foo")
  #
  # {?X1, ?X2} | {?X2 <: /string} | {... X : /string}
  #
.
