# This is an example how dataflow analysis would look in datalog

# The example is adapted from https://souffle-lang.github.io/examples
# Encoded code fragment:
#
#   v1 = h1();
#   v2 = h2();
#   v1 = v2;
#   v3 = h3();
#   v1.f = v3;
#   v4 = v1.f;

Decl assign(VarL, VarR)
  bound [/v, /v].

Decl new(Var, Obj)
  bound [/v, /name].

Decl load(VarL, VarR, Field)
  bound [/v, /v, /field].

Decl store(VarL, Field, VarR)
  bound [/v, /field, /v].

# Facts

assign(/v/1,/v/2).

new(/v/1, /h/1).
new(/v/2, /h/2).
new(/v/3, /h/3).

store(/v/1, /field/f, /v/3).
load(/v/4, /v/1, /field/f).

# Analysis
Decl alias(Var1, Var2)
  bound [/v, /v].

alias(X,X) :- assign(X,_).
alias(X,X) :- assign(_,X).
alias(X,Y) :- assign(X,Y).
alias(X,Y) :- load(X,A,F), alias(A,B), store(B,F,Y).


Decl pointsTo(Var, Obj)
  bound [/v, /name].

pointsTo(X,Y) :- new(X,Y).
pointsTo(X,Y) :- alias(X,Z), pointsTo(Z,Y).


# Our "sum type" is add a prefix /instr
# /instr/read
# /instr/write
# /instr/jump

# Facts
Decl read(InstrRead, Var)
  bound [/instr/read, /v].
Decl write(InstrWrite, Var)
  bound [/instr/write, /v].
Decl succ(Instr1, Instr2)
  bound [/instr, /instr].

read(/instr/read/1, /v/1).
read(/instr/read/2, /v/1).
read(/instr/read/3, /v/2).
write(/instr/write/1, /v/1).
write(/instr/write/2, /v/2).
write(/instr/write/3, /v/2).

succ(/instr/write/1, /instr/o1).
succ(/instr/o1, /instr/read/1).
succ(/instr/o1, /instr/read/2).
succ(/instr/read/2, /instr/read/3).
succ(/instr/read/3, /instr/write/2).

# Analysis

Decl flow(Instr1, Intr2)
  bound [/instr, /instr].

flow(X,Y) :- succ(X,Y).
flow(X,Z) :- flow(X,Y), flow(Y,Z).

Decl defUse(InstrWrite , InstrRead)
  bound [/instr/write, /instr/read].

defUse(W,R) :- write(W,X), flow(W,R), read(R,X).
