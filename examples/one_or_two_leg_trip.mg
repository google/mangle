# An example that shows how we can collect results.
# This is not how ticket pricing works for Swiss trains.

# A list of train connections between cities where we
# do not have to change trains, with prices in CHF.
direct_conn(/code/zl, /zurich, /lausanne, 60).
direct_conn(/code/zb, /zurich, /bern, 30).
direct_conn(/code/bl, /bern, /lausanne, 30).

# We want to retrieve all possibilities of reaching
# cities that would require not more than one change of trains.

# The first possibility is that there is a direct connection.
one_or_two_leg_trip(Codes, Start, Destination, Price) :-
  direct_conn(Code, Start, Destination, Price)
  |> let Codes = [Code].

# The second possibility is that we have to change trains somewhere.
one_or_two_leg_trip(Codes, Start, Destination, Price) :-
  direct_conn(FirstCode, Start, Connecting, FirstLegPrice),
  direct_conn(SecondCode, Connecting, Destination, SecondLegPrice)
  |> let Codes = [FirstCode, SecondCode],
     let Price = fn:plus(FirstLegPrice, SecondLegPrice).
