# Declarations

Declarations provide type information and metadata for predicates.
They enable static checking: the system verifies that all facts and
rules conform to the declared types before evaluation.

## Syntax

A declaration starts with `Decl`, followed by the predicate with
argument variables, optional descriptor and bound blocks, and ends with a dot:

```
Decl predicate_name(Arg1, Arg2, ..., ArgN)
  descr [...]
  bound [Type1, Type2, ..., TypeN]
  bound [...]
  .
```

All parts except `Decl`, the predicate, and the final dot are optional.

## Bound declarations

A bound declaration specifies the types of the predicate's arguments.
The number of types in a bound must match the number of arguments.

```
Decl volunteer(ID, Name, Skill)
  bound [/number, /string, /name].
```

This says every `volunteer` fact has a number in the first position,
a string in the second, and a name in the third.

### Multiple bounds

Multiple bound declarations express alternatives. A fact must conform
to at least one of the declared bounds.

```
Decl entry(Key, Value)
  bound [/string, /number]
  bound [/string, /string].
```

This means `entry` can have a string key with either a number or string value.

### Constructed types in bounds

Bounds can use any type expression, including constructed types.
The dot syntax (`.List<...>`, `.Struct<...>`, etc.) is often more
readable than the `fn:` prefix form.

```
Decl person(P)
  bound [.Struct</name : /string, /age : /number>].
```

A list of numbers:

```
Decl numbers(Ns)
  bound [.List</number>].
```

A map from strings to lists of numbers:

```
Decl index(M)
  bound [.Map</string, .List</number>>].
```

### Tagged union bounds

Tagged unions are particularly useful in declarations because they
describe internally-tagged message types directly.

```
Decl event(E)
  bound [
    .TaggedUnion</kind,
      /user_login  : .Struct</user_id : /number, opt /ip_address : /string>,
      /user_logout : .Struct</user_id : /number>,
      /bulk_import : .Struct</items : .List</string>>
    >
  ].
```

See [Type Expressions](typeexpressions.md) for full details on tagged unions.

### Type variables in bounds

Bounds can contain type variables (capital letters). A type variable
stands for an unknown but fixed type. When a type variable appears
in multiple argument positions, it constrains them to have the
same type.

```
Decl first_element(List, Elem)
  bound [.List<X>, X].
```

This says that the element type of the list must match the type of
the second argument.

## Descriptors

Descriptors provide metadata about a predicate. They appear in a
`descr [...]` block.

### Documentation

The `doc` descriptor provides a human-readable description:

```
Decl volunteer(ID, Name, Skill)
  descr [
    doc("Volunteers and their skills."),
    arg(ID, "unique identifier"),
    arg(Name, "full name"),
    arg(Skill, "area of expertise")
  ]
  bound [/number, /string, /name].
```

### Extensional predicates

The `extensional` descriptor means the predicate's facts come from
external data, not from rules in the program:

```
Decl sensor_reading(Timestamp, Value)
  descr [extensional()]
  bound [/number, /float64].
```

### Modes

The `mode` descriptor specifies whether arguments are input (`+`),
output (`-`), or both (`?`):

```
Decl lookup(Key, Value)
  descr [mode(+, -)]
  bound [/string, /number].
```

### Functional dependencies

The `fundep` descriptor declares that some arguments functionally
determine others:

```
Decl config(Key, Value)
  descr [fundep([Key], [Value])]
  bound [/string, /string].
```

This means that for a given key, there is exactly one value.

## Combining bounds and rules

Declarations work together with rules. The system checks that
every fact and every rule's inferred types conform to the declared bounds.

```
Decl color(C)
  bound [.Union<.Singleton</red>, .Singleton</green>, .Singleton</blue>>].

color(/red).
color(/green).
color(/blue).
```

For predicates defined by rules, the system infers the types from the
rule premises and checks them against the declaration:

```
Decl api_message(M)
  bound [
    .TaggedUnion</type,
      /create : .Struct</name : /string, /count : /number>,
      /delete : .Struct</id : /number>,
      /ping   : .Struct<>
    >
  ].

api_message({/type: /create, /name: "widget", /count: 5}).
api_message({/type: /delete, /id: 42}).
api_message({/type: /ping}).

Decl message_type(M, T)
  bound [/any, /name].

message_type(M, T) :- api_message(M), :match_field(M, /type, T).
```
