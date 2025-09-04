# Multi-line Continuation Fix

## Problem

The Mangle interpreter's multi-line continuation feature had a spacing issue. When users entered multi-line input that didn't end with "." or "!", the interpreter would prompt for continuation with ".. >" and concatenate the lines directly without proper spacing.

### Example of the issue:
```
mg > rule(X)
.. > :-
.. > pred(X).
```

This would be concatenated as: `rule(X):-pred(X).` (missing spaces)

In some cases, this could lead to parsing issues or unexpected behavior, especially with comments:
```
mg > # comment
.. > rule(X) :- pred(X).
```

This would become: `# commentrule(X) :- pred(X).` (rule becomes part of comment)

## Solution

Modified the continuation logic in `interpreter/interpreter.go` (lines 351-360) to add appropriate spacing between continued lines:

```go
// Add appropriate spacing between lines to avoid parsing issues
if nextLine != "" && !strings.HasSuffix(clauseText, " ") && !strings.HasPrefix(nextLine, " ") {
    clauseText = clauseText + " " + nextLine
} else {
    clauseText = clauseText + nextLine
}
```

### Logic:
- If neither the current text ends with a space nor the next line starts with a space, add a space between them
- Otherwise, concatenate directly (preserving existing spacing/formatting)

## Testing

Added comprehensive tests in `interpreter/interpreter_test.go` to verify:
1. Single line rules work correctly
2. Rules with proper spacing work correctly  
3. Rules without space after `:-` work correctly
4. All existing functionality remains intact

## Verification

All existing tests continue to pass, confirming no regressions were introduced while fixing the continuation spacing issue.