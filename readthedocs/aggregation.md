# Aggregation

Datalog as a basic, yet powerful way to define and query data.
However, real-life use cases frequently need more than querying.
For example, instead of getting a large table of results, we
may want to group values, count the number of results, compute
sums, sort. In database community, this functionality is called
*aggregation*.

## Grouping

We reuse our volunteer database example but with an eye on
software development projects. Every project
has a person that act as an organizer, a number of software developers
that are assigned to it with a number of hours per week.

The volunteer table `volunteer(ID, Name, Skill)` has records for
people in our volunteer database. Only look at people with skills
`/organizer` and `/software_development` are assigned to projects.

| ID | Name | Skill |
|----|------|-------|
| /v/3 | Alyssa P. Hacker | /software_development |
| /v/3 | Alyssa P. Hacker | /organizer |
| /v/4 | Ivan Hassenovich | /software_development |
| /v/4 | Ivan Hassenovich | /organizer |
| /v/5 | Claudio Ferrari | /organizer |
| ... | ... | ... |

In this example, we use Mangle names like `/v/1` for IDs, which
helps us distinguish between different kinds of IDs.

The project table `project(ProjectID, Name)` is where projects are defined.

| ProjectID | Name |
|-----------|-------------|
| ... | ... |
| /p/10 | Ultimate Kubernetes Control Plane UI |
| /p/11 | YAML Engineer Online Courseware |
| /p/12 | Personal Dopamine Fasting Tracker |
| /p/13 | LLM Fact Checker |
| ... | ... |

The table `project_assignment(ProjectID, VolunteerID, Role, Hours)`
has information how many hours a person is contributing to which project.
It is volunteering, so people assign themselves. Yet, that also means,
hours may be low and sometimes a project is missing an organizer or
developer or simply dead.

| ProjectID | VolunteerID | Role | Hours |
|-----------|-------------|------|-------|
| ... | ... | ... | ... |
| /p/10 | /v/3 | /organizer | 2 |
| /p/10 | /v/3 | /software_development | 2 |
| /p/10 | /v/4 | /software_development | 2 |
| /p/11 | /v/4 | /software_development | 20 |
| /p/12 | /v/5 | /organizer | 20

Suppose we want to list *how many* software developers are assigned
to a project, and the total number of hours. In Mangle, we separate this
into two parts: first, we identify all records, then, we *group* them rows
and specify what aggregation we want.

```
project_dev_energy(ProjectID, NumDevelopers, TotalHours) ⟸
  project_assignment(ProjectID, VolunteerID, /software_development, Hours)
  |> do fn:group_by(ProjectID),
     let NumDevelopers = fn:count(),
     let TotalHours = fn:sum(TotalHours).
```

The query `project_assignment(ProjectID, VolunteerID, /software_development)`
gives us a result relation like this:

| ProjectID | VolunteerID | Hours |
|-----------|-------------|------|
| /p/10 | /v/3 | 2 |
| /p/10 | /v/4 | 2 |
| /p/11 | /v/4 | 20 |

The part after the `|>` is called a do-transformation: we take the whole
result relation as a whole and do something to it. In this case, we
group by `ProjectID`, which we can imagine as creating a separate relation
for each `ProjectID` value.

ProjectID /p/10

| VolunteerID | Hours |
|-------------|------|
| /v/3 | 2 |
| /v/4 | 2 |

ProjectID /p/11

| VolunteerID | Hours |
|-------------|------|
| /v/4 | 20 |

Finally, the other parts of the do-transform say what we should do with each
of the per-`ProjectID` tables. With `fn:count()` we count the number of rows,
and with `fn:sum(Hours)` we sum the values in the `Hours` column.

This then yields the final `project_dev_energy` table that contains the
aggregated values.

| ProjectID | NumDevelopers | TotalHours |
|-----------|-------------|------|
| /p/10 | 2 | 4 |
| /p/11 | 1 | 20 |

:::{admonition} Click to see SQL translation
:class: dropdown
Here is how the `project_work_energy` rule look like in SQL:
```sql
CREATE TABLE project_work_energy AS
SELECT ProjectID, COUNT(VolunteerID) as NumDevelopers, SUM(Hours) as TotalHours
FROM project_assignment
WHERE Role = '/software_developer'
GROUP BY ProjectID
```
:::

## Absence of values and Negation

If you followed the example carefully, you may notice that there is
no way that we can end up with 0 developers. A project that has no
developers assigned will simply not show up in the result relation.
Datalog is a mechanism that deals with positive information.

We can deal with absence, by making use of negation. First we
define a helper table that identifies projects without developers:

```
project_with_developers(ProjectID) ⟸
  project_assignment(ProjectID, _, /software_development, _).

project_without_developers(ProjectID) ⟸
  project_name(ProjectID, _).
  !project_with_developers(ProjectID)
```

Datalog negation is explained elswhere, for now, do note
that the `project_without_developers` query mentions
the `ProjectID` appears in a positive subquery as well as a negated subquery.
Every variable mentioned in the head of a rule has to be mentioned in at
least one subquery that is not negated.

With this, we can now add another rule to `project_dev_energy` that handles
the case of zero developers.

```
project_dev_energy(ProjectID, NumDevelopers, TotalHours) ⟸
  project_without_developers(ProjectID),
  NumDevelopers = 0,
  TotalHours = 0.
```

The final `project_dev_energy` table then looks like this:

| ProjectID | NumDevelopers | TotalHours |
|-----------|-------------|------|
| /p/10 | 2 | 4 |
| /p/11 | 1 | 20 |
| /p/12 | 0 | 0 |
| /p/13 | 0 | 0 |

:::{admonition} Click to see SQL translation for the zero case
:class: dropdown
Here is how to get a result that includes projects with zero developers in SQL.
We need to `LEFT JOIN` the project table with the assignments.

```sql
CREATE TABLE project_dev_energy AS
SELECT
  p.ProjectID,
  COUNT(pa.VolunteerID) AS NumDevelopers,
  COALESCE(SUM(pa.Hours), 0) AS TotalHours
FROM
  project AS p
LEFT JOIN
  project_assignment AS pa
ON
  p.ProjectID = pa.ProjectID AND pa.Role = '/software_development'
GROUP BY
  p.ProjectID
```
:::
