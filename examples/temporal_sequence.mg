# Temporal Sequence Example: Event A followed by Event B within 10 minutes.
#
# This example demonstrates how to use temporal facts and interval arithmetic
# to detect a sequence of events with a time constraint.

Decl event_a(Name) temporal bound [/name].
Decl event_b(Name) temporal bound [/name].
Decl match(Name) bound [/name].

# Event A happens at 10:00:00 for user /u1
event_a(/u1)@[2024-01-01T10:00:00].

# Event B happens at 10:05:00 for user /u1 (within 10 mins of A)
event_b(/u1)@[2024-01-01T10:05:00].

# Event A happens at 10:00:00 for user /u2
event_a(/u2)@[2024-01-01T10:00:00].

# Event B happens at 10:15:00 for user /u2 (15 mins after A - too late)
event_b(/u2)@[2024-01-01T10:15:00].

# Match rule:
# 1. B happened at time Tb
# 2. A happened at time Ta
# 3. A happened before B (:time:lt)
# 4. Duration between A and B is <= 10 minutes
match(U) :-
  event_b(U)@[Tb],
  event_a(U)@[Ta],
  :time:lt(Ta, Tb),
  Diff = fn:time:sub(Tb, Ta),
  Limit = fn:duration:parse('10m'),
  :duration:le(Diff, Limit).
