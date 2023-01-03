# Documentation

This directory has some documentation in markdown format.

## Meta considerations and rationale

Documentation falls into four categories:

- learning-oriented tutorials
- goal-oriented how-to guides
- understanding-oriented discussions
- information-oriented reference material

Mangle, we want to distinguish between language and its implementation.
The language part is not only creating a specification that would permit
different implementations. While multiple implementations of Mangle would be
nice, the rationale for Mangle includes propagating an abstract way of thinking
about data processing which is independent of implementation and can be
realized in various implementation contexts.

While such documentation of the language is necessarily abstract and will
employ concepts and terminology from logic, we want to avoid becoming so
rigorous as to become inaccessible.

## Language

- [Rationale](rationale.md)
- [Data Model](spec_datamodel.md)
- [Built-in Operations](spec_builtin_operations.md)
- [Relational algebra](spec_explain_relational_algebra.md)

## Implementation

### Understanding

- [Example: Volunteer DB](example_volunteer_db.md)
- [Understanding Derivation of Facts](explanation_derived_facts.md)
- [Using the interactive interpreter](using_the_interpreter.md)

Some additional resources for datalog:

- [Bibliography](bibliography.md)