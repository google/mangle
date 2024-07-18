Decl f(R)
  descr [
    mode("+")
  ]
  bound [fn:Struct(/g, fn:List(/string))].

f(R) :-
  :match_field(R, /g, G),
  fn:list:len(G) > 0.
