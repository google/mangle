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

## Date construction and conversion

Mangle provides helpers for working with `/date` constants.  All dates are
expressed in ISO-8601 calendar form (YYYY-MM-DD) and represent a calendar day in
UTC without a time-of-day component.

- `fn:date:from_string(String)` parses an ISO-8601 date string and returns a
  `/date` value.  Invalid strings (wrong format or impossible dates) produce a
  runtime error.
- `fn:date:from_parts(Year, Month, Day)` constructs a date from numeric parts.
  Each argument must be a number; the function validates that the combination is
  a real calendar date.
- `fn:date:to_string(Date)` converts a `/date` constant back into its ISO-8601
  string form.

## Date arithmetic

Date values support simple arithmetic helpers:

- `fn:date:add_days(Date, Days)` returns the date that is `Days` days after the
  provided date.  `Days` must be a number and may be negative.
- `fn:date:sub_days(Date, Days)` subtracts `Days` from the provided date.  This
  is equivalent to `fn:date:add_days(Date, -Days)` but expressed explicitly for
  clarity.
- `fn:date:diff_days(Left, Right)` returns the number of whole days between two
  dates (`Left - Right`).  The result is a numeric constant.

## Decimal conversions

Decimals can be created and converted using helper functions:

- `fn:decimal:from_string(String)` parses a decimal string and returns a
  `/decimal` value.
- `fn:decimal:from_number(Number)` and `fn:decimal:from_float64(Float)` convert
  existing numeric types into decimals.
- `fn:decimal:to_string(Decimal)` formats a decimal as a string.  Values with
  non-terminating decimal expansions are rounded to 34 fractional digits.
- `fn:decimal:to_number(Decimal)` converts decimals that represent integers
  into `/number` values.  A runtime error is raised if the decimal has a
  fractional component or does not fit into 64 bits.
- `fn:decimal:to_float64(Decimal)` produces an IEEE-754 `float64` approximation
  of the decimal value.
