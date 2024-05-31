# Live-variable analysis is a classic dataflow analysis.
#
# We only represent whether the statement at a program point definitely
# assigns (def) or uses (use) a variable.
#
# The Dragon Book specifies the analysis using these dataflow equations,
# where In[B] (Out[B]) are the live variables at incoming (outgoing) side
# of basic block B and use_B (def_B) are the variables used (definitely
# assigned) in B.
#
#   In[exit] = \emptyset
#
#   In[B]  = use_B \union Out[B] - def_B
#   Out[B] = \Union_{S is successor of B} In[S]
#
# Information flows "backward": a variable Var is live at exit of a
# point Point if it is used at P or if it is live later and the statement
# at P does not definitely assign to V.
#
# When expressed as datalog rules, these equations are rearranged
# to talk about relations instead of input/output sets.

Decl edge(Source, Target).
Decl def(Point, Var).
Decl use(Point, Var).

live(Point, Var) :-
  use(Point, Var).

live(Point, Var) :-
  edge(Point, Succ), live(Succ, Var), !def(Point, Var).

# What follows is a representation of a Rust program with a borrow-check
# error.
#
# 1: let mut x = 1;
# 2: let y = &x;
# 3: *x = 1;
# 4: print(y);
# 5: // exit

edge(1, 2).
edge(2, 3).
edge(3, 4).
edge(4, 5).

# `let mut x = 1;`
def(1, "x").

# `let y = &x;`
def(2, "y").
use(2, "x").

# `*x = 2;`
def(3, "x").

# `print(y);`
use(4, "y").


# When you load this into the interpreter, you can query at which program
# points which variables are live:
#
# mg >?live
# live(2,"x")
# live(3,"y")
# live(4,"y")
