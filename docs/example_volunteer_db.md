# Mangle as a Database

Here is an example how datalog relates to relational databases. To do something
different than the tiring employee, department, salary example, let's look
at a volunteer database.

## The Extensional DB

We define predicates (tables) by listing all the facts that involve the
predicate, right in source. These predicates are called *extensional* (the
set of facts is the *extension* of the predicate). In a real system, these facts
would reside somewhere else, like a file or an actual relational DB.

We shall use constant symbols like `/v/{number}` as identifers. A volunteer
has a *name*, some time-slots where they might be available for volunteering
work, some *skills* they have and some skills they are *interested* in
developing.

```
# We shall use constant symbols like /v/{number} as identifers.
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
