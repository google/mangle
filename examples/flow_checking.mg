# In the following, we develop a simple form of "flow checking", a form of
# static analysis that checks whether a program is free of a specified class
# of bugs.
# This intraprocedural analysis (only one function body) and does not deal with
# loops.

# We assume a program that is type-checked and has been in a CFG where
# all complex operations have been reduced and intermediate results named.
#
# A pointer has an associated heap type, which is either int or a record
# type with members.

heap_type(/i32).
heap_type(/Box_i32).

type(/i32ptr).
type(X) :- heap_type(X).

# The /Box_i32 type has a single member, .ptr, which is a pointer i32*
members(/Box_i32, ".ptr").

# Variables, including temporaries.
variable("b1").
variable("b2").
variable("tmp1").
variable("tmp2").

var_type("b1", /Box_i32).
var_type("b2", /Box_i32).
var_type("tmp1", /i32ptr).
var_type("tmp1", /i32).

#
# Whenever we have a variable with region of a type with members, then
# we also have a "projection" region.
#
variable_region_projection(Var, Region, Member, MemberRegion) :-
  variable_region(Var, Region),
  var_type(Var, Type),
  members(Type, Member)
  |> let MemberRegion = fn:string:concat(Region, Member).


# Program Point - instruction mapping.
#
# "move" var var: assigning variable to another, destructive move (/Box_i32)
#
#    before the move, y.ptr must be valid. 
#    open: x.ptr may not need to be valid?
#    x = y includes x.destroy, invalidating x.ptr.
#    open: support a move where x.ptr does not need to be valid?
#
#    after the move, x.ptr has been assigned y.ptr and is valid
#    y.ptr is invalid.
#
# "copy" var var: assigning variable to another, copy (/i32 and /i32ptr)
# "alloc" var type region: allocate and assign to var (/Box_i32)
# "store" var member rhsvar: instructions assign to fields (/Box_i32)
# "load" var var member: copy-assign after member access (/i32 or /i32ptr)
# "deref" dereferences a pointer (only /i32ptr)

edge(/p_entry, /p1).

#
# var b1 = Box_i32::Make()
#
pp_instr(/p1, { /var: "b1", /instr: "alloc", /type: "/Box_i32", /region: "^A1" }).
edge(/p1, /p1mid).
edge(/p1mid, /p2).

#
# var b2 = Box_i32::Make()
#
pp_instr(/p2, { /var: "b2", /instr: "alloc", /type: "/Box_i32", /region: "^A2" }).
edge(/p2, /p2mid).
edge(/p2mid, /p3).

#
# b1 = b2
#
pp_instr(/p3, { /var: "b1", /instr: "move", /rhs: "b2" }).
edge(/p3, /p3mid).
edge(/p3mid, /p4).

#
# tmp1 = b2.ptr
#
pp_instr(/p4, { /var: "tmp1", /instr: "load", /rhs: "b2", /member: ".ptr" }).
edge(/p4, /p4mid).
edge(/p4mid, /p5).

#
# tmp2 = *tmp1
#
pp_instr(/p5, { /var: "tmp2", /instr: "deref", /rhs: "tmp1" }).
edge(/p5, /p5mid).
edge(/p5mid, /p6).

#
# b1 = b2
#
pp_instr(/p6, { /var: "b1", /instr: "move", /rhs: "b2" }).
edge(/p6, /p6mid).
edge(/p6mid, /p_pexit).

#
# Rules
#

pp_var_region(/p_entry, Var, Region) :- variable_region(Var, Region).

pp_var_region(/p_entry, FakeVar, MemberRegion) :-
  variable_region_projection(Var, Region, Member, MemberRegion),
  FakeVar = fn:string:concat("fake-", MemberRegion).

pp_var_region(P, Var, Region) :-
  edge(/p_entry, P),
  pp_var_region(/p_entry, Var, Region).

# alloc:
#
# For an alloc instruction "x = alloc ...", use the region from alloc operation.
# This is based on the declaration. The region in the alloc instruction is
# fresh, ie. the name is different from regions of other variables.
#
pp_var_region(Pmid, Var, Region) :-
  pp_instr(P, Instr),
  edge(P, Pmid),
  :match_field(Instr, /var, Var),
  :match_field(Instr, /instr, Op), Op = "alloc",
  :match_field(Instr, /region, Region).


# move:
#
# For a move instruction "x = y":

# 1. mark the region of y as invalid.
#
invalid_region(Pmid, MovedVar, MovedRegion) :-
  pp_instr(P, Instr),
  edge(P, Pmid),
  :match_field(Instr, /var, Var),
  :match_field(Instr, /instr, Op), Op = "move",
  :match_field(Instr, /rhs, MovedVar),
  var_type(Var, /Box_i32),
  pp_var_region(P, MovedVar, MovedRegion).

# move: (ctd.)
#
# Propagate all mappings
#
pp_var_region(Pmid, Var, Region) :-
  pp_instr(P, Instr),
  edge(P, Pmid),
  :match_field(Instr, /instr, Op), Op = "move",
  pp_var_region(P, Var, Region).

# store:
#
# For a store instruction, add region alias.
#
region_alias_base(Pmid, Var, Region, OtherRegion) :-
  pp_instr(P, Instr),
  edge(P, Pmid),
  :match_field(Instr, /var, Var),
  :match_field(Instr, /instr, Op), Op = "store",
  :match_field(Instr, /rhs, OtherVar),
  pp_var_region(P, Var, Region),
  pp_var_region(P, OtherVar, OtherRegion).

# store: (ctd).
#
# For a store instruction, propagate all (!) mappings.
#
pp_var_region(Pmid, Var, Region) :-
  pp_instr(P, Instr),
  edge(P, Pmid),
  :match_field(Instr, /instr, Op), Op = "store",
  pp_var_region(P, Var, Region).

# load:
#
# For a load instruction, take the projection region.
#
region_alias_base(Pmid, Var, Region, RhsRegionProjection) :-
  pp_instr(P, Instr),
  edge(P, Pmid),
  :match_field(Instr, /var, Var),
  :match_field(Instr, /instr, Op), Op = "load",
  :match_field(Instr, /rhs, Rhs),
  :match_field(Instr, /member, Member),
  pp_var_region(P, Var, Region),
  pp_var_region(P, Rhs, RhsRegion),
  variable_region_projection(Rhs, RhsRegion, Member, RhsRegionProjection).

# load:
#
# For a load instruction, copy all (!) mappings - we added the alias above.
#
pp_var_region(Pmid, Var, Region) :-
  pp_instr(P, Instr),
  edge(P, Pmid),
  :match_field(Instr, /var, Var),
  pp_var_region(P, Var, Region),
  :match_field(Instr, /instr, Op), Op = "load".

# deref:
#
# For a deref instruction, propagate.
#
pp_var_region(Pmid, Var, Region) :-
  pp_instr(P, Instr),
  edge(P, Pmid),
  :match_field(Instr, /instr, Op), Op = "deref",
  pp_var_region(P, Var, Region).

#
# Copy over region information from previous mid to point.
#
pp_var_region(P, Var, Region) :-
  pp_var_region(PrevMid, Var, Region),
  edge(PrevMid, P).

#
# Copy over invalid region info from previous mid to point.
# Cases without "move" instruction.
#
invalid_region(P, Var, Region) :-
  invalid_region(PrevMid, Var, Region),
  edge(PrevMid, P),
  pp_instr(P, Instr),
  :match_field(Instr, /instr, Op), Op != "move".

#
# Copy over invalid region info from previous mid to point.
# For "move": copy over the region from the rhs of the move instruction.
#
invalid_region(P, Var, Region) :-
  invalid_region(PrevMid, Var, Region),
  edge(PrevMid, P),
  pp_instr(P, Instr),
  :match_field(Instr, /instr, Op), Op = "move",
  :match_field(Instr, /rhs, RhsVar), Var = RhsVar.

invalid_region(PMid, Var, Region) :-
  edge(P, PMid),
  invalid_region(P, Var, Region),
  pp_instr(P, _).

#
# When a region is invalid, then also its projections.
#
invalid_region(P, FakeVar, ProjectionRegion) :-
  invalid_region(P, Var, Region),
  pp_var_region(P, FakeVar, ProjectionRegion),
  variable_region_projection(Var, Region, Member, ProjectionRegion).

# Error condition: deref of something with invalid region.

error_condition(P, "invalid region accessed") :-
  pp_instr(P, Instr),
  :match_field(Instr, /var, Var),
  :match_field(Instr, /instr, Op), Op = "deref",
  :match_field(Instr, /rhs, RhsVar),
  pp_var_region(P, RhsVar, RhsRegion),
  region_alias(P, _, RhsRegion, InvalidRegion),
  invalid_region(P, _, InvalidRegion).


error_condition(P, "invalid move") :-
  pp_instr(P, Instr),
  :match_field(Instr, /var, Var),
  :match_field(Instr, /instr, Op), Op = "move",
  :match_field(Instr, /rhs, RhsVar),
  pp_var_region(P, RhsVar, RhsRegion),
  region_alias(P, _, RhsRegion, InvalidRegion),
  invalid_region(P, _, InvalidRegion).

# Convenient short-hands to refer to region at a program point.
variable_region("b1", "^b1").
variable_region("b2", "^b2").
variable_region("tmp1", "^tmp1").
variable_region("tmp2", "^tmp2").

#
# Propagate Region aliases, unless "store" or "load"
#
region_alias_base(Pmid, Var, Region, OtherRegion) :-
  edge(P, Pmid),
  pp_instr(P, Instr),
  :match_field(Instr, /instr, Op), Op != "load", Op != "store",
  region_alias_base(P, Var, Region, OtherRegion).

region_alias_base(P, Var, Region, OtherRegion) :-
  edge(PrevMid, P),
  region_alias_base(PrevMid, Var, Region, OtherRegion).

# Embed region_alias_base into region_alias.
#
region_alias(P, Var, Region, OtherRegion) :-
  region_alias_base(P, Var, Region, OtherRegion).

# Make region_alias reflexive.
#
region_alias(P, Var, Region, Region) :-
  pp_var_region(P, Var, Region).

# Make region_alias symmetric.
#
region_alias(P, Var, Region, OtherRegion) :-
  region_alias(P, Var, OtherRegion, Region).

# Make region_alias transitive.
#
region_alias(P, Var, Region, OtherRegion) :-
  region_alias(P, Var, Region, SomeRegion),
  region_alias(P, Var, SomeRegion, OtherRegion).

# Add shorthands (necessary?)
#
region_alias(P, Var, Region, RegionShorthand) :-
  pp_var_region(P, Var, Region),
  variable_region(Var, RegionShorthand).

# When you load this file into the interpreter, it correctly detects
# that an invalid region has been dereferenced.
#
# mg >?error_condition
# error_condition(/p5,"invalid region accessed")
# error_condition(/p6,"invalid move")
# Found 2 entries for error_condition(X0,X1).
