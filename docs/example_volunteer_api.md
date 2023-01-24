# Mangle as an API

![docu badge spec explanation](docu_spec_explanation.svg)

This page is explains how Mangle can be useful by sketching a flexible and
powerful query API for a knowledge base. It continues with example data from
[Volunteer database](example_volunteer_db.md).

## An API for Queries

Suppose we want to make the volunteer database part of application that follows
three-tier architecture. How can it practically work to use Mangle for
briding business concepts and data representation?

We focus on the following use cases:

*   find all volunteers who are available at a particular time slot

*   find all volunteers who have a particular skill

If we know the structure of the knowledge base (schema), both queries happen to
be one-liners:

```
?volunteer_time_available(VolunteerID, /tuesday, /afternoon)
?volunteer_skill(VolunteerID, /skill/workshop_facilitation)
```

This suggests a very simple query interface that works for all such queries:

```
input:  an atom
output: all matching facts
```

## Supporting Joins and Unions

If we want to find all volunteers who are available on either Tuesday afternoon,
or Wednesday morning, we would have to make two datalog queries. We could adapt
our simple interface accordingly:

```
input:  list of atoms
output: facts that match at least one of the atoms
```

## Supporting more flexible queries

This is still not as flexible as we'd like: for example, we may want to
retrieve not only the `VolunteerID` but also the name.

Therefore, let's make our interface slightly more general by adding 
an option to specify an additional datalog program:

```
program: a Mangle datalog program (optional)
input:   an atom
output:  all matching facts
```

Now our queries can do everything that is possible with relational algebra
plus recursion.

```
program: """
good_time(/tuesday, /afternoon).
good_time(/wednesday, /morning).
matching_availability(VolunteerID, Name, Weekday, Timeslot) :-
    volunteer_time_available(VolunteerID, Weekday, Timeslot),
    good_time(Weekday, Timeslot),
    volunteer_name(VolunteerID, Name).
"""
input: "matching_availability(VolunteerID, Name, Weekday, Timeslot)"
```

At request time, we need to parse the program, and evaluate it on the already
existing knowledge base. Care must be taken that queries do not become too
expensive in terms of machine resources.

When the query is for something specific (mentions constants), then there are
some optimization opportunities.

## Information hiding and adaptability

Parnas 1972 coined the term "information hiding" that when dividing responsibilities
software one should do this along the lines of "difficult design decisions or design
decisions which are likely to change." The internal structure of our knowledge
base could be the result of such a difficult or likely-to-change design decision.

Suppose we want to use internally more fine-grained timeslots than `/morning`,
`/afternoon`. But the client team is busy and cannot be bother with
changing the queries until next quarter. How can we design an querying API
without having to change the queries at the same time? There are two ways to do this:

1. Maintain the possibility of querying `volunteer_time_available`
1. Make `matching_availability` part of the "API surface" that can evolve
in sync

Both are the same solution: introduce an indirection, in the form of a set
of "public" predicates (relations/views) that can be queried, and internal
ones that remain private/hidden. If queries that we intend to support are
captured with named definitions, we actually revert to the first API design:

```
input:  an atom
output: all matching facts
```

## Queries with structured data

Supplemental data (such as `good_time` above) can be passed as part of 
the query atom when we use structured data types. The rules providing
"public" views may make it undesirable or impossible to evaluate all
results ahead of time (all possible lengths).

```
# We can only evaluate this predicate in the context of a query that
# binds the GoodTimeSlots argument to a value.
matching_availability(VolunteerID, Name, GoodTimeSlots) :-
    volunteer_time_available(VolunteerID, Weekday, Timeslot),
    :list:member(fn:pair(Weekday, Timeslot), GoodTimeSlots),
    volunteer_name(VolunteerID, Name).
```

In Mangle, we can declare that such predicates are invoked in a particular
"mode", that is: some arguments are input arguments, others are output
arguments.

NOTE: not yet implemented.

```
# GoodTimeSlots must be provided at request time
Decl matching_availability(VolunteerID, Name, GoodTimeSlots)
  descr [mode("+", "+", "-"), public()].
```

## On the Wire

The above high-level API design does not depend on a string representation
of programs and queries. It is possible to serialize and deserialize,
but one may prefer to fix the API surface using any API specification mechanism
such as gRPC.

```
// gRPC service definition
message ByAvRequest { repeated Availability availability = 1; }
message ByAvReply { repeated Volunteer reply = 1; }
service VolunteerQuery {
  rpc GetByMatchingAvailability (ByAvRequest) returns (ByAvReply) {}
}
```

The implementation of `GetByMatchingAvailability` service translates
the wire format to a Mangle query (or program) and let evaluation take care
of the rest. The design decision concerning the API is not in the choice of
technology; it is the choice to tie all service methods of the API to
the definition of Mangle relations (predicates) that are implemented as
rules. The rules can evolve with the schema without clients noticing.
What is more, the rules can use the ubiquitous language that users
are familiar with.
