# Decimal design notes

Mangle represents `/decimal` values using Go's `math/big.Rat` type.  This choice
keeps arithmetic exact for terminating decimal fractions and avoids pulling in a
large external dependency.  Decimal constants store the canonical rational form
(e.g. `3/2`) while user-facing formatting uses up to 34 fractional digits with
trailing zeros trimmed.

Arithmetic on mixed numeric types follows promotion rules:

* Pure integer arguments still produce `/number` results.
* Any decimal argument promotes the operation to `/decimal` and performs exact
  rational arithmetic.  Functions such as `fn:plus`, reducers (`fn:max`,
  `fn:sum`), and predicates (`:lt`, `:within_distance`) all adopt this behaviour.
* Float helpers (`fn:float:*`, `fn:avg`) accept decimals by converting them to
  IEEE-754 values.

Conversions between numeric kinds are exposed explicitly via
`fn:decimal:from_*` and `fn:decimal:to_*` helpers.  Converting a decimal to a
`/number` requires the value to be an integer that fits into 64 bits; otherwise a
runtime error is raised.  When converting to `float64` the closest representable
value is returned.

Displaying decimals relies on `FormatDecimal`, which rounds non-terminating
fractions to 34 digits.  Because the canonical stored value remains exact,
repeated calculations do not accumulate rounding errors beyond the selected
output precision.
