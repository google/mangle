# Example demonstrating the new fn:collect_to_map reducer function
# This shows how to aggregate results into a map structure.

# Sample data: users and their preferred programming languages
user_language(/alice, /python).
user_language(/alice, /go).
user_language(/bob, /javascript).
user_language(/bob, /typescript).
user_language(/charlie, /rust).

# Sample data: languages and their popularity scores
language_popularity(/python, 95).
language_popularity(/go, 85).
language_popularity(/javascript, 90).
language_popularity(/typescript, 88).
language_popularity(/rust, 80).

# Collect all languages for each user into a list (existing functionality)
user_languages_list(User, Languages)
  :- user_language(User, Language)
  |> do fn:group_by(User), let Languages = fn:collect(Language).

# NEW: Collect language preferences as a map with popularity scores
# This creates a map from language to popularity score for each user
user_language_scores(User, LanguageScoreMap)
  :- user_language(User, Language),
     language_popularity(Language, Score)
  |> do fn:group_by(User), let LanguageScoreMap = fn:collect_to_map(Language, Score).

# Another example: Create a map from user to their highest-scored language
# First find the max score per user
user_max_score(User, MaxScore)
  :- user_language(User, Language),
     language_popularity(Language, Score)
  |> do fn:group_by(User), let MaxScore = fn:max(Score).

# Then create a user-to-top-language map
top_language_per_user(UserLanguageMap)
  :- user_language(User, Language),
     language_popularity(Language, Score),
     user_max_score(User, Score)  # Only take languages with max score
  |> do fn:group_by(), let UserLanguageMap = fn:collect_to_map(User, Language).