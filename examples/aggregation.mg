# We gradually build up a database of people and topics
# they might be in an expert in (/knows) or enjoy doing (/likes).

# observed(Subject, Verb, Topic, Weight, Description).

# We add a weight to each edge positive or negative evidence.
# This is very crude, but may still be more insightful to
# expose and critique the reasoning than feeding an AI chatbot
# with textual descriptions.

observed(/john, /knows, /cooking, 1, "Has lots of books").
observed(/john, /likes, /cooking, -1, "Has not been reading them.").

observed(/ahmed, /knows, /cooking, 1, "Is cooking regularly.").
observed(/ahmed, /knows, /cooking, -1, "He does not try out any new things.").
observed(/ahmed, /likes, /cooking, 1, "He invites friends over to cook together.").

observed(/mia, /knows, /management, -1, "Does not have a lot of experience.").
observed(/mia, /knows, /management, -1, "She rarely presents her team's work.").
observed(/mia, /likes, /management, 1, "She enjoys helping her people grow.").

# Now, let's add all the weights.

aggregated(Subject, Verb, Topic, Sum)
  :- observed(Subject, Verb, Topic, Weight, _)
  |> do fn:group_by(Subject, Verb, Topic), let Sum = fn:sum(Weight).

filtered(Subject, Verb, Topic)
  :- aggregated(Subject, Verb, Topic, Sum), Sum >= 1.
