# Built-in operations

## Built-in predicates

The value of two expressions `Left`, `Right` can be compared:

- equality `Left = Right`
- inequality `Left != Right`.
- less than `Left < Right` (numeric, int64)
- less than or equal `Left <= Right` (numeric, int64)
- `:float:lt`, `:float:le`, `:float:gt`, `:float:ge` (float64)
- `:time:lt`, `:time:le`, `:time:gt`, `:time:ge` (time)
- `:duration:lt`, `:duration:le`, `:duration:gt`, `:duration:ge` (duration)

The infix operators `<`, `<=`, `>`, `>=` operate on `/number` (int64) only.
For `/float64`, `/time`, and `/duration` values, use the corresponding
prefixed predicate form (e.g. `:float:lt(X, 2.0)`). Equality (`=`) and
inequality (`!=`) are type-polymorphic and work on any pair of values of the
same type.

A pair can be matched using pattern `:match_pair(Pair, First, Second)`.

A list can be matched using patterns 
`:match_cons(List, Head, Tail)` and `:match_nil(List)`. Here "cons" is
used to mean first element of a non-empty list (and) the rest (tail) of the
list.

A map can be matched using pattern `:match_entry(Map, Key, Value)`.

A struct can be matched using pattern `:match_field(Struct, FieldName, Value)`.

## Built-in accessor functions

Accessing the first member of a pair `fn:pair:fst(Pair)`. Accessing the second member of a pair `fn:pair:snd(Pair)`. 

The n-th member of a list can be accessed using `fn:list:get(ListValue, Index)`. 
