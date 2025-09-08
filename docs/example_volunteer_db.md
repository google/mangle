# Mangle as a Database

![docu badge spec explanation](docu_spec_explanation.svg)

Here is an example how datalog relates to relational databases. To do something
different than the tiring employee, department, salary example, let's look
at a volunteer database.

## The Extensional DB

We define predicates (tables) by listing all the facts that involve the
predicate, right in source. These predicates are called *extensional* (the
set of facts is the *extension* of the predicate). In a real system, these facts
would reside somewhere else, like a file or an actual relational DB.

We shall use constant symbols like `/v/{number}` as identifiers. A volunteer
has a *name*, some time-slots where they might be available for volunteering
work, some *skills* they have and some skills they are *interested* in
developing.

```
# We shall use constant symbols like /v/{number} as identifiers.
volunteer(/v/1).

# Volunteer has a name, some timeslots where they might be available for
# volunteering work, some interests and some skills.
volunteer_name(/v/1, "Aisha Salehi").
volunteer_time_available(/v/1, /monday, /afternoon).
volunteer_time_available(/v/1, /monday, /morning).
volunteer_interest(/v/1, /skill/frontline).
volunteer_skill(/v/1, /skill/admin).
volunteer_skill(/v/1, /skill/facilitate).
volunteer_skill(/v/1, /skill/teaching).

volunteer(/v/2).
volunteer_name(/v/2, "Xin Watson").
volunteer_time_available(/v/2, /monday, /afternoon).
volunteer_interest(/v/2, /skill/facilitate).
```

## Some Queries

When managing volunteers, we may want to match a volunteer who has
a skill with a volunteer who is interest in learning that skill.

```
teacher_learner_match(Teacher, Learner, Skill) :-
  volunteer_skill(Teacher, Skill),
  volunteer_interest(Learner, Skill).
```

That was easy, but how do they know if their time-slots also match? Let's
make a better query.

```
teacher_learner_match_session(Teacher, Learner, Skill, PreferredDay, Slot) :-
  teacher_learner_match(Teacher, Learner, Skill),
  volunteer_time_available(Teacher, PreferredDay, Slot),
  volunteer_time_available(Learner, PreferredDay, Slot).
```

These predicates are *intensional*: they come with rules that tell us how
we can derive more data from available data.

Try these queries:
```
?teacher_learner_match_session
?teacher_learner_match_session(Teacher, Learner, _, _)
```

## Vocabulary

Mangle has checks that ensure that predicate are used with the right number of
arguments. There are also mechanisms to constrain the set of constants that
are used. Before we go there, it would be interesting to collect the known
good values into unary predicates.

```
skill(/skill/admin).
skill(/skill/facilitate).
skill(/skill/frontline).
skill(/skill/speaking).
skill(/skill/teaching).
skill(/skill/recruiting).
skill(/skill/workshop_facilitation).
```

Using declarations, we can constrain what goes into an argument position.

```
Decl volunteer_interest(Volunteer, Skill)
  bound [ /v, /skill ].

Decl volunteer_skill(Volunteer, Skill)
  bound [ /v, /skill ].
```

Now, if you use something that is not a skill, you should get an error.

```
volunteer_interest(/v/1, /monday). # This does not look right.
```

## Alternatives: Tables and Value Tables

Above, we used one predicate per property. There are alternatives that work
well when the structure of the data (schema) is well-known. This section
discusses some of these alternatives.

In a typical relational database management system, data is organized in
"tables:" each rows has a fixed number of (named) columns. When we drop the
names, a table is the same thing as a relation.

A volunteer table could look like as follows. For some properties like
`time_available`, `interest`, `skill` we use structured types
(lists and pairs). One can avoid the use of structured types by adding more
tables; however structured data types are essential for modeling.

| id | name | time_available | interest | skill |
|----|------|----------------|----------|-------|
|`/v/1`|"Aisha Salehi"| `[ fn:pair(/monday, /morning), fn:pair(/monday, /afternoon) ]` | `[/skill/frontline]` | `[/skill/admin, /skill/facilitate, /skill/teaching]` |
|`/v/2`|"Xin Watson"  | `[ fn:pair(/monday, /afternoon) ]`               | `[/skill/facilitate]`         |       |

We can take the structured types further structs (records) and work with
so-called [value tables](https://github.com/google/zetasql/blob/master/docs/data-model.md#value-tables):
these are tables with a single column of a structured type. In Mangle, this
would look as follows:

```
volunteer_record({
   /id:             /v/1,
   /name:           "Aisha Salehi",
   /time_available: [ fn:pair(/monday, /morning), fn:pair(/monday, /afternoon) ],
   /interest:       [ /skill/frontline ],
   /skill:          [ /skill/admin, /skill/facilitate, /skill/teaching ]
}).

volunteer_record({
   /id:             /v/2,
   /name:           "Xin Watson",
   /time_available: [ fn:pair(/monday, /afternoon) ],
   /interest:       [ /skill/frontline ],
   /skill:          [ /skill/admin, /skill/facilitate, /skill/teaching ]
}).
```

In Mangle, it is easy to switch between these views. We can get back the unary
relation of ids and the per-property predicates by defining some rules:

```
volunteer(Id) :-
  volunteer_record(R), :match_field(R, /id, Id).

volunteer_name(Id, Name) :-
  volunteer_record(R), :match_field(R, /id, Id), :match_field(R, /name, Name).
```
