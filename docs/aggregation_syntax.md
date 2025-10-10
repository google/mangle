# Enhanced Aggregation Syntax

This document describes the enhanced aggregation syntax implemented in response to issue #52, which provides a more expressive and intuitive way to perform aggregations in Mangle.

## Overview

The traditional Mangle aggregation syntax uses `|> do fn:group_by()` followed by reduction functions. While functional, this syntax can be verbose for complex aggregations. The new `fn:aggregate_by()` function provides a more expressive alternative that enables more sophisticated aggregation patterns.

## New Aggregation Functions

### fn:aggregate_by()

The `fn:aggregate_by()` function is a drop-in replacement for `fn:group_by()` that supports additional aggregation capabilities:

```prolog
# Basic counting (equivalent to group_by + count)
count_projects(Num) :-
  project(P) |> do fn:aggregate_by(), let Num = fn:group_size().

# Grouping with no key variables treats all data as one group
count_all_items(Total) :-
  item(X) |> do fn:aggregate_by(), let Total = fn:group_size().
```

### fn:group_map(KeyVar, ValueVar)

Creates a map from key expressions to collections of values, enabling complex grouping operations:

```prolog
# Group dependencies by project
project_dependencies(DepsMap) :-
  has_dependency(Project, Dependency)
  |> do fn:aggregate_by(), let DepsMap = fn:group_map(Project, Dependency).

# Result: {/project1: [/lib1, /lib2], /project2: [/lib1, /lib3]}
```

### fn:group_size()

Returns the count of items in the current aggregation group:

```prolog
# Count projects with vulnerable dependencies
vulnerable_project_count(Count) :-
  has_vulnerable_dependency(Project, _)
  |> do fn:aggregate_by(), let Count = fn:group_size().
```

## Examples Addressing Issue #52

The original issue suggested this syntax:

```
count_projects_with_vulnerable_log4j(Num) :-
  projects_with_vulnerable_log4j(P)
  |> { AllVulnProjects =>
    val groups = AllVulnProjects.groupBy { it.projectId }
    val Num = groups.values.sumOf { it.size }
  }
```

With the new aggregation functions, this can be expressed as:

```prolog
count_projects_with_vulnerable_log4j(Num) :-
  projects_with_vulnerable_log4j(P)
  |> do fn:aggregate_by(), let Num = fn:group_size().
```

Or for more complex grouping:

```prolog
projects_by_vulnerability_type(GroupedProjects) :-
  project_vulnerability(Project, VulnType)
  |> do fn:aggregate_by(), let GroupedProjects = fn:group_map(VulnType, Project).
```

## Benefits

1. **More Expressive**: The new functions enable complex grouping patterns that were difficult with the original syntax
2. **Backward Compatible**: All existing `fn:group_by()` usage continues to work unchanged
3. **Intuitive**: The `fn:group_map()` function provides a clear way to create grouped collections
4. **Efficient**: Built on the same underlying aggregation engine as the original functions

## Migration Guide

Existing code using `fn:group_by()` continues to work without changes. To adopt the new syntax:

1. Replace `fn:group_by()` with `fn:aggregate_by()` where desired
2. Use `fn:group_size()` instead of `fn:count()` for counting group members
3. Use `fn:group_map(Key, Value)` to create grouped collections

The new functions are fully compatible with all existing reduction functions like `fn:sum()`, `fn:max()`, `fn:collect()`, etc.