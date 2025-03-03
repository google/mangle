# Here are a few examples. '#' is a line-comment.

# A rule (without body) for a 0-argument predicate.
# These are not very useful, except for tests.
foo().

# A rule for a 1-argument predicate. Note that it is possible to have numbers
# or strings, by default there are no constraints on predicate arguments.
bar(/some/name/constant).
bar(606).
bar("a").

# You can use single quote for strings
bar('aaa').
bar('they asked "why?"').

# Multi-line
bar(`
Soon may the Wellerman come
To bring us sugar and tea and rum
`).

# We can use unary predicates to define a set of values, like an enumeration.
# These user-defined constants have no special meaning in Mangle.
my_boolean(/true).
my_boolean(/false).

# Datalog's logical data model is close to the relational model.
# Here is a predicate definition that corresponds to a table.
fruit(/apple, /sweet, /red).
fruit(/apple, /sweet, /yellow).
fruit(/apple, /sour, /green).
fruit(/banana, /sweet, /yellow).

# Try querying this by loading this file into the interactive interpreter
# and typing this at the prompt: 
# ?fruit
# ?fruit(_, /sweet, _)

# Here is a way to "unit test" rules. At the prompt, we query ?test_bar().
# and check that it is there.
test_bar() :- bar('a'), !bar('c').

# If you want to check for absence, you can use negation.
# We want to ensure that there is no false() atom...
false() :- bar('b').

# ... we define a rule that checks that there is no false() atom.
test_foobar() :- !false().

# First queries
example_number(1).
example_number(5).
example_number(19).

# This is no longer plain Datalog.
example_foobar(X) :- example_number(X), fn:minus(20, X) < 5.
