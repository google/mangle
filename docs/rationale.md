# Rationale

There are three roots that led to the design of Mangle:

- SQL is taken as the "standard" for querying relational data, but SQL was
never designed as programming language and there are *severe* shortcomings
that prevent the use of SQL for maintainable advanced queries or libraries of compositional query building blocks.

- the reality of modern systems is that they need to deal with various
data sources using heterogeneous representations, with no central control.
We need to deal with multiple languages / type systems / schema / on-the-wire
representations that are on equal footing, but we do not want to give up the
possibility of properties guaranteed by static analysis.

- knowledge graph and semantic web technology based on triples would enable a uniform representation of heterogeneous data. However, structured data
becomes too cumbersome to handle in this paradigm and query languages do not seem to extend well to
more programming tasks. JSON-LD, the practical answer, is defined in terms
of the RDF, and semantics is ultimately still based on triples and "blank
nodes". Similar concerns apply to property graphs.

In our view, this explains the recent re-emergence of Datalog in both academia and industry.
There are multiple languages from the Datalog family in active use
today, each with extensions and implementations tailored to different
contexts. For some of these practical languages,
it is not obvious how they even relate to Datalog.

We need a practically useful and yet principled point in the space of
possible designs that overcomes these problems. At the same time, it should be self-consistent and
accessible so we can teach potential users what Datalog is about and what can be
done with fixpoint reasoning. Therefore we wanted a language that has
relatively pure datalog as a fragment.

As for the implementation, we also had practical use cases that required
structured data, aggregations and extensibility, as well as something that
could be easily embedded in golang applications. Mangle is the result and
we think it can serve as a recipe for similar efforts.