# Temporal Reasoning

Facts don't always stay true forever. Employees change teams. Certifications
expire. Systems go up and down. Mangle's temporal reasoning extension lets you
track *when* things are true, not just *that* they're true.

## Temporal Facts

Add a time interval to any fact with `@[start, end]`:

```
# Alice was on engineering from Jan 2020 to June 2023
team_member(/alice, /engineering)@[2020-01-01, 2023-06-15].

# Bob joined engineering in 2019 and is still there
team_member(/bob, /engineering)@[2019-06-01, _].

# Something that happened at a specific moment
login(/alice)@[2024-03-15T10:30:00].

# Currently active (from 2024 until now)
active(/alice)@[2024-01-01, now].

# Something happening right now
logged_in(/bob)@[now].
```

Special bounds:
- `_` means unbounded (beginning of time or ongoing into the future)
- `now` means the current evaluation time

Facts without `@[...]` work exactly like before - they're eternal truths.

## Temporal Operators

Four operators let you reason about time:

### Past Operators

**Diamond-minus `<-`** - "was true at some point in the past"

```
# Was active sometime in the last 30 days
recently_active(X) :- <-[0d, 30d] active(X).
```

**Box-minus `[-`** - "was true continuously in the past"

```
# Was continuously active for the last 30 days
reliably_active(X) :- [-[0d, 30d] active(X).
```

### Future Operators

**Diamond-plus `<+`** - "will be true at some point in the future"

```
# Will be on maintenance sometime in the next 7 days
maintenance_pending(X) :- <+[0d, 7d] maintenance(X).
```

**Box-plus `[+`** - "will be true continuously in the future"

```
# Contract will be valid for the entire next 30 days
covered(X) :- [+[0d, 30d] contract(X).
```

The interval `[0d, 30d]` means "from now to 30 days ago/ahead". You can use:
- `d` for days
- `h` for hours
- ISO timestamps like `2024-01-15T10:30:00`

## Allen's Interval Relations

Built-in predicates for comparing intervals (based on Allen's interval algebra):

| Predicate | Meaning |
|-----------|---------|
| `:interval:before(T1, T2)` | T1 ends before T2 starts |
| `:interval:after(T1, T2)` | T1 starts after T2 ends |
| `:interval:meets(T1, T2)` | T1 ends exactly when T2 starts |
| `:interval:overlaps(T1, T2)` | T1 and T2 share some time |
| `:interval:during(T1, T2)` | T1 is contained within T2 |
| `:interval:contains(T1, T2)` | T1 contains T2 |
| `:interval:starts(T1, T2)` | T1 and T2 start together |
| `:interval:finishes(T1, T2)` | T1 and T2 end together |
| `:interval:equals(T1, T2)` | T1 and T2 are identical |

## Interval Functions

Extract components from bound interval variables:

| Function | Returns |
|----------|---------|
| `fn:interval:start(T)` | Start time in nanoseconds |
| `fn:interval:end(T)` | End time in nanoseconds |
| `fn:interval:duration(T)` | Duration in nanoseconds |

Example:
```
# Get the duration of an event
event_duration(X, D) :-
    event(X)@T,
    D = fn:interval:duration(T).
```

## Interval Variable Binding

You can bind the interval of a matching fact to a variable:

```
# Bind the validity interval to T
event_with_time(X, T) :- event(X)@T.
```

The bound variable is a pair of timestamps (start, end) in nanoseconds.

## Interval Coalescing

When facts have adjacent or overlapping intervals, they get merged:

```
# If you assert:
employed(/alice)@[2020-01-01, 2021-12-31].
employed(/alice)@[2022-01-01, 2023-12-31].

# Coalescing produces:
employed(/alice)@[2020-01-01, 2023-12-31].
```

This prevents interval explosion in recursive rules.

## Temporal Declarations

Mark a predicate as temporal using the `temporal` keyword in declarations:

```
# Declare a temporal predicate
Decl employee_status(person, status) temporal bound [/name, /string].

# Regular (eternal) predicate for comparison
Decl config_setting(key, value) bound [/string, /string].
```

The `temporal` keyword signals that facts for this predicate have validity
intervals. You can check if a declaration is temporal programmatically:

```go
if decl.IsTemporal() {
    // This predicate has temporal semantics
}
```

## Example: Checking Certification Compliance

Here's a realistic use case - make sure operators had certifications before
using equipment:

```
# Declare predicates
Decl certified(person, cert) temporal.
Decl operated(person, equipment, timestamp) temporal.

# Facts
certified(/alice, /forklift)@[2023-01-01, 2024-01-01].
operated(/alice, /forklift, _)@[2023-06-15].

# Rule: Find violations - operated without 30 days of certification
violation(Person, Equipment) :-
    operated(Person, Equipment, _)@[Time, Time],
    ![-[30d, 30d] certified(Person, Equipment).
```

## Programmatic Usage

To use temporal features from Go:

```go
import (
    "time"
    "github.com/google/mangle/engine"
    "github.com/google/mangle/factstore"
)

// Create a temporal store and add facts
store := factstore.NewTemporalStore()
store.Add(myAtom, interval)

// Evaluate with temporal support
stats, err := engine.EvalProgramWithStats(
    programInfo,
    regularStore,
    engine.WithTemporalStore(store),
    engine.WithEvaluationTime(time.Now()),
)
```

### Temporal Store API

The `TemporalFactStore` interface provides these methods:

```go
// Add a temporal fact (returns added, error)
store.Add(atom ast.Atom, interval ast.Interval) (bool, error)

// Add a fact that's always true
store.AddEternal(atom ast.Atom) (bool, error)

// Query facts valid at a specific time
store.GetFactsAt(query ast.Atom, t time.Time, callback) error

// Query facts overlapping an interval
store.GetFactsDuring(query ast.Atom, interval ast.Interval, callback) error

// Merge adjacent/overlapping intervals
store.Coalesce(predicate ast.PredicateSym) error
```

### Time and Interval Helpers

The `ast` package provides convenience functions to reduce verbosity when creating times and intervals:

```go
// Create dates without typing all the zeros
t := ast.Date(2024, 1, 15)                    // midnight in default timezone
t := ast.DateTime(2024, 1, 15, 10, 30)        // with hour and minute
t := ast.DateTimeSec(2024, 1, 15, 10, 30, 45) // with seconds

// Create intervals concisely
interval := ast.TimeInterval(startTime, endTime)
interval := ast.DateInterval(2023, 1, 1, 2024, 12, 31)  // most concise
```

**Before** (verbose):
```go
store.Add(atom, ast.NewInterval(
    ast.NewTimestampBound(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
    ast.NewTimestampBound(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)),
))
```

**After** (concise):
```go
store.Add(atom, ast.DateInterval(2023, 1, 1, 2024, 12, 31))
```

| Helper | Example | Purpose |
|--------|---------|---------|
| `ast.Date(y, m, d)` | `ast.Date(2024, 1, 15)` | Date at midnight |
| `ast.DateTime(y, m, d, h, min)` | `ast.DateTime(2024, 1, 15, 10, 30)` | Date with time |
| `ast.DateTimeSec(y, m, d, h, min, s)` | `ast.DateTimeSec(2024, 1, 15, 10, 30, 45)` | Date with seconds |
| `ast.DateIn(y, m, d, loc)` | `ast.DateIn(2024, 1, 15, nyc)` | Date in specific timezone |
| `ast.DateTimeIn(y, m, d, h, min, loc)` | `ast.DateTimeIn(2024, 1, 15, 10, 30, nyc)` | Date+time in specific timezone |
| `ast.TimeInterval(start, end)` | `ast.TimeInterval(t1, t2)` | Interval from time.Time values |
| `ast.DateInterval(...)` | `ast.DateInterval(2023, 1, 1, 2024, 12, 31)` | Interval from date components |

### Timezone Configuration

All date/time helpers use a configurable default timezone (UTC by default). Set it once at program startup:

```go
// Default: UTC (no configuration needed)

ast.SetTimezone("UTC")                   // Explicit UTC
ast.SetTimezone("Local")                 // System timezone
ast.SetTimezone("America/New_York")      // IANA timezone name
ast.SetTimezone("PST")                   // Common abbreviations work too

// For init(), use MustSetTimezone (panics on error)
func init() {
    ast.MustSetTimezone("America/New_York")
}
```

Supported abbreviations: `EST`, `CST`, `MST`, `PST`, `GMT`, `CET`, `JST`, `IST`, and more.

**Important:** Set the timezone once at startup before creating any temporal facts.

For one-off values in a different timezone without changing the default, use `DateIn` or `DateTimeIn`:

```go
// Default is UTC, but this one event is in NYC time
store.Add(event, ast.TimeInterval(
    ast.DateTimeIn(2024, 1, 15, 19, 0, "America/New_York"),  // 7pm NYC
    ast.DateTimeIn(2024, 1, 15, 22, 0, "EST"),               // 10pm EST
))

## Time Bridge Functions

When mixing temporal reasoning with regular data columns containing timestamps:

| Function | Purpose |
|----------|---------|
| `fn:time:from_nanos(N)` | Convert nanoseconds to temporal-compatible value |
| `fn:time:to_nanos(T)` | Convert temporal bound to nanoseconds |
| `fn:time:add(T, D)` | Add duration (nanos) to timestamp |

Example:
```
# Bridge between a column timestamp and temporal queries
valid_order(Order) :-
    order(Order, CreatedAtNanos),
    Timestamp = fn:time:from_nanos(CreatedAtNanos),
    fn:time:after(Timestamp, fn:time:add(fn:time:now, -2592000000000000)).  # 30 days in nanos
```

## Decidability and Termination

Temporal reasoning introduces potential non-termination. The implementation includes
safeguards, but understanding safe vs dangerous patterns helps avoid issues.

### Safe Patterns (Guaranteed to Terminate)

```
# SAFE: Past-only lookback, no temporal head
recent_login(User) :-
    <-[7d] login(User).

# SAFE: Temporal head but no recursion through temporal predicates
expires_soon(License) @[now, End] :-
    license(License) @[_, End],
    :interval:before(End, fn:time:add(now, 2592000000000000)).
```

### Dangerous Patterns (May Not Terminate)

```
# DANGEROUS: Recursive temporal derivation with expanding intervals
extended(X) @[T1, T2] :-
    base(X) @[T1, T0],
    extended(X) @[T0, T2].  # Self-reference extends interval

# DANGEROUS: Unbounded future generation
will_happen(X) @[now, _] :-
    trigger(X),
    [+[1d] will_happen(X).  # Infinite future facts
```

### Built-in Safeguards

1. **Interval Coalescing**: Adjacent/overlapping intervals are merged automatically
2. **Interval Limit**: Default 1000 intervals per atom, configurable via `WithMaxIntervalsPerAtom(n)`
3. **Fact Limits**: Use `engine.WithCreatedFactLimit(n)` to cap total derived facts

Configure the interval limit:
```go
// Custom limit
store := factstore.NewTemporalStore(factstore.WithMaxIntervalsPerAtom(5000))

// No limit (use with caution)
store := factstore.NewTemporalStore(factstore.WithMaxIntervalsPerAtom(-1))
```

### Complexity

Based on DatalogMTL research:
- Non-recursive temporal queries: AC⁰ data complexity
- Full DatalogMTL: PSPACE-complete data complexity
- Forward-propagating programs: May not terminate without coalescing
