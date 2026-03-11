# Type Expressions

This section describes type expressions in more detail and introduces
more ways to construct type expressions.

For basic types like `/number`, we have used the same syntax as for names,
whereas for constructed types, we have introduced a syntax that involves
a *type constructor* `fn:Pair(/number, /string)`.

There is also a convenient *dot syntax* for type constructors that uses
angle brackets: `.Pair</number, /string>`. Both forms are equivalent.

## Type variables

A type expression may contain *type variables*. These start with a
capital letter.

**Examples of types containing type variables:**
```
X
fn:Pair(Y, /string)
.Pair<Y, /string>
```

## Any type

There is a type expression `/any`. Any datum has the type `/any`.

## Singleton types

For every name that appears in a program, there is a unique
type `fn:Singleton(`{math}`name``)`.

**Examples**

```
fn:Singleton(/foo)
.Singleton</foo>
```

## Union types

If {math}`T_1,...,T_n` are type expressions, then
`fn:Union(`{math}`T_1,...,T_n``)` is a union type.
A value has a union type if it has at least one of {math}`T_1,...,T_n`.

There is an empty union type `fn:Union()`.

**Examples**

```
fn:Union()
fn:Union(/name, /string)
fn:Union(fn:Singleton(/foo), fn:Singleton(/bar))
.Union</name, /string>
```

## Constructed type expressions

The basic types and type constructors for pairs, tuples, lists, maps,
and structs are described in the [Constructed Types](constructedtypes.md) section.
Here we use the dot syntax for brevity.

| Type | Dot Syntax | Internal Form |
|------|-----------|---------------|
| Pair | `.Pair</number, /string>` | `fn:Pair(/number, /string)` |
| Tuple ({math}`n \geq 3`) | `.Tuple</number, /string, /name>` | `fn:Tuple(/number, /string, /name)` |
| List | `.List</number>` | `fn:List(/number)` |
| Map | `.Map</string, /number>` | `fn:Map(/string, /number)` |
| Struct | `.Struct</x : /number, /y : /string>` | `fn:Struct(/x, /number, /y, /string)` |
| Option | `.Option</string>` | `fn:Option(/string)` |

## Struct types

A struct type describes the fields and their types. Required fields
are given as label-type pairs, and optional fields use `opt`.

```
.Struct</name : /string, /age : /number>
.Struct</name : /string, opt /nickname : /string>
```

In the internal form, required fields alternate between label and
type, and optional fields are wrapped with `fn:opt`:

```
fn:Struct(/name, /string, /age, /number)
fn:Struct(/name, /string, fn:opt(/nickname, /string))
```

Struct subtyping is structural: a struct type with more fields conforms
to a struct type with fewer fields.

## Tagged union types

A *tagged union* (also called discriminated union or internally-tagged union) is
a type where a designated *tag field* inside a struct determines which
*variant* is active. Each variant has a tag value (a name constant) and
a struct type describing the variant's fields.

A tagged union value is an ordinary struct that has the tag field set to a
variant's tag value, plus that variant's fields.

**Syntax**

```
.TaggedUnion<tag_field, /variant1 : .Struct<...>, /variant2 : .Struct<...>>
```

The first argument is the tag field name (a name constant). Each subsequent
pair is a variant tag and a struct type.

**Example: API message type**

Consider a JSON API that uses the `type` field as a discriminator:

```json
{"type": "create", "name": "widget", "count": 5}
{"type": "delete", "id": 42}
{"type": "ping"}
```

This is described with a tagged union:

```
.TaggedUnion</type,
  /create : .Struct</name : /string, /count : /number>,
  /delete : .Struct</id : /number>,
  /ping   : .Struct<>
>
```

The variant `/ping` takes no arguments (`.Struct<>`). The tag field `/type`
must not appear in any variant struct; it is added automatically.

**Variants with optional and list fields**

Variant structs can use all the features of struct types, including
optional fields and nested types:

```
.TaggedUnion</kind,
  /user_login  : .Struct</user_id : /number, opt /ip_address : /string>,
  /bulk_import : .Struct</items : .List</string>>
>
```

**Values**

Tagged union values are ordinary structs. They are constructed with the
standard struct syntax and destructured with `:match_field`:

```
event({/kind: /user_login, /user_id: 101, /ip_address: "10.0.0.1"}).
event({/kind: /bulk_import, /items: ["a", "b", "c"]}).

login_user(U) :-
  event(E),
  :match_field(E, /kind, K), K = /user_login,
  :match_field(E, /user_id, U).
```

**Semantics**

A tagged union is semantically equivalent to a union of struct types where
each alternative struct contains the tag field as a singleton type:

```
# .TaggedUnion</kind, /a : .Struct</x : /number>, /b : .Struct<>>
# is equivalent to:
.Union<
  .Struct</kind : .Singleton</a>, /x : /number>,
  .Struct</kind : .Singleton</b>>
>
```

The tagged union form is preferred because it makes the discriminator
explicit and is more concise.

## Definition of type expressions

Type expressions are the expressions that we can build from type variables,
basic types, `/any` and type constructors.
