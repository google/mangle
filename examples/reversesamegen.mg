# This is an example from the Alice Book (Foundations of Databases), Ch. 13
# We have partial information on ancestor same generation relationships
# and want to get a complete picture.

up(/a, /e).
up(/a, /f).
up(/f, /m).
up(/g, /n).
up(/h, /n).
up(/i, /o).
up(/j, /o).

flat(/g, /f).
flat(/m, /n).
flat(/m, /o).
flat(/p, /m).

down(/l, /f).
down(/m, /f).
down(/g, /b).
down(/h, /c).
down(/i, /d).
down(/p, /k).

rsg(X, Y) :- flat(X, Y).
rsg(X, Y) :- up(X, X1), rsg(Y1, X1), down(Y1, Y).
