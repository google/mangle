# Tagged unions (internally-tagged discriminated unions).
#
# A tagged union is a type where a designated tag field inside a struct
# determines which variant is active. The value is an ordinary struct
# with the tag field plus the variant's fields.
#
# Syntax: .TaggedUnion<tag_field, /variant1 : .Struct<...>, /variant2 : .Struct<...>>

# --- API message type ---
#
# Models a JSON API message with internally-tagged variants.
# The /type field is the discriminator.

Decl api_message(M)
  bound[
    .TaggedUnion</type,
      /create : .Struct</name : /string, /count : /number>,
      /delete : .Struct</id : /number>,
      /ping   : .Struct<>
    >
  ].

api_message({/type: /create, /name: "widget", /count: 5}).
api_message({/type: /create, /name: "gadget", /count: 12}).
api_message({/type: /delete, /id: 42}).
api_message({/type: /ping}).

# Extracting the tag lets us dispatch on the message type.
Decl message_type(M, T)
  bound[
    .TaggedUnion</type,
      /create : .Struct</name : /string, /count : /number>,
      /delete : .Struct</id : /number>,
      /ping   : .Struct<>
    >,
    /name
  ].
message_type(M, T) :- api_message(M), :match_field(M, /type, T).

# --- Rich event type with optional and list fields ---
#
# Demonstrates opt (optional fields) and .List inside a tagged union.

Decl event(E)
  bound[
    .TaggedUnion</kind,
      /user_login  : .Struct</user_id : /number, opt /ip_address : /string>,
      /user_logout : .Struct</user_id : /number>,
      /bulk_import : .Struct</items : .List</string>, opt /dry_run : .Union<.Singleton</true>, .Singleton</false>>>
    >
  ].

event({/kind: /user_login, /user_id: 101, /ip_address: "10.0.0.1"}).
event({/kind: /user_login, /user_id: 102}).
event({/kind: /user_logout, /user_id: 101}).
event({/kind: /bulk_import, /items: ["a", "b", "c"], /dry_run: /true}).
event({/kind: /bulk_import, /items: ["x"]}).

# Extract login user IDs.
Decl login_user(U)
  bound [/number].
login_user(U) :-
  event(E),
  :match_field(E, /kind, K), K = /user_login,
  :match_field(E, /user_id, U).

# Count items in bulk imports.
Decl bulk_import_size(Items, N)
  bound[.List</string>, /number].
bulk_import_size(Items, N) :-
  event(E),
  :match_field(E, /kind, K), K = /bulk_import,
  :match_field(E, /items, Items)
  |> let N = fn:list:len(Items).
