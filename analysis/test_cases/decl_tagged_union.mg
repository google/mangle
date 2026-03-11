# A tagged union with two variants: move and quit.
Decl event(E)
  bound[
    .TaggedUnion</kind,
      /move : .Struct</x : /number, /y : /number>,
      /quit : .Struct<>
    >
  ].

# A move event.
event({/kind: /move, /x: 10, /y: 20}).

# A quit event.
event({/kind: /quit}).

# Extracting the tag field from a tagged union.
Decl event_kind(E, K)
  bound[
    .TaggedUnion</kind,
      /move : .Struct</x : /number, /y : /number>,
      /quit : .Struct<>
    >,
    /name
  ].
event_kind(E, K) :- event(E), :match_field(E, /kind, K).
