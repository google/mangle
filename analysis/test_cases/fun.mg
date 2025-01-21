Decl f(R)
  descr [
    mode("+")
  ]
  bound [.Struct</g : .List</string>>].

f(R) :-
  :match_field(R, /g, G),
  fn:list:len(G) > 0.
