# The ancestor predicate is an example of a relation that can be defined
# using a recursive rule.

# Rules with a head but without a body simply define facts.

parent(/Oedipus, /Polynices).
parent(/Polynices, /Thersander).
parent(/Polynices, /Timeas).
parent(/Polynices, /Adrastus).

# Rules with a body define how to produce new facts from existing ones.
# We use `⟸` here but you can also use `:-` to separate head and body.

# - Every parent of Y is an ancestor of Y.
ancestor(X, Y) ⟸ parent(X, Y).

# - Every parent of an ancestor of Z is also an ancestor of Z.
ancestor(X, Z) ⟸ parent(X, Y), ancestor(Y, Z).

# Try loading this file and querying Adrastus' ancestors.
# mg > ?ancestor(_, /Adrastus)
# ancestor(/Oedipus,/Adrastus)
# ancestor(/Polynices,/Adrastus)
# Found 2 entries for ancestor(_,/Adrastus).

# There are other ways to define this relation. Try changing the second clause
# to "Every ancestor of an  # ancestor of Z is also an ancestor of Z" and
# reloading the file.
