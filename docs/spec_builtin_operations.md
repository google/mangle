# Built-in operations

## Built-in predicates

The value of two expressions `Left`, `Right` can be compared:

- equality `Left = Right`
- inequality `Left != Right`.
- less than `Left < Right` (numeric)
- less than or equal `Left <= Right`

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
