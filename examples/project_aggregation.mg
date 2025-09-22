# This file contains the data and queries from the aggregation.md documentation.

# --- Data from tables ---

# volunteer(ID, Name, Skill)
volunteer(/v/3, "Alyssa P. Hacker", /software_development).
volunteer(/v/3, "Alyssa P. Hacker", /organizer).
volunteer(/v/4, "Ivan Hassenovich", /software_development).
volunteer(/v/4, "Ivan Hassenovich", /organizer).
volunteer(/v/5, "Claudio Ferrari", /organizer).

# project(ProjectID, Name)
project(/p/10, "Ultimate Kubernetes Control Plane UI").
project(/p/11, "YAML Engineer Online Courseware").
project(/p/12, "Personal Dopamine Fasting Tracker").
project(/p/13, "LLM Fact Checker").

# project_assignment(ProjectID, VolunteerID, Role, Hours)
project_assignment(/p/10, /v/3, /organizer, 2).
project_assignment(/p/10, /v/3, /software_development, 2).
project_assignment(/p/10, /v/4, /software_development, 2).
project_assignment(/p/11, /v/4, /software_development, 20).
project_assignment(/p/12, /v/5, /organizer, 20).


# --- Queries (rules) from the documentation ---

# This rule calculates the number of developers and total hours for projects
# that have at least one developer.
project_dev_energy(ProjectID, NumDevelopers, TotalHours) :-
  project_assignment(ProjectID, _, /software_development, Hours)
  |> do fn:group_by(ProjectID),
     let NumDevelopers = fn:count(),
     let TotalHours = fn:sum(Hours).

# Helper rule to identify projects with developers.
project_with_developers(ProjectID) :-
  project_assignment(ProjectID, _, /software_development, _).

# Helper rule to identify projects without developers.
project_without_developers(ProjectID) :-
  project(ProjectID, _),
  !project_with_developers(ProjectID).

# This rule handles projects with zero developers.
project_dev_energy(ProjectID, 0, 0) :-
  project_without_developers(ProjectID).
