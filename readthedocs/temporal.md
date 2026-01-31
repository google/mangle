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
store := factstore.NewSimpleTemporalStore()
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
// Add a temporal fact
store.Add(atom ast.Atom, interval ast.Interval) bool

// Add a fact that's always true
store.AddEternal(atom ast.Atom) bool

// Query facts valid at a specific time
store.GetFactsAt(query ast.Atom, t time.Time, callback) error

// Query facts overlapping an interval
store.GetFactsDuring(query ast.Atom, interval ast.Interval, callback) error

// Merge adjacent/overlapping intervals
store.Coalesce(predicate ast.PredicateSym) error
```
