# Using the interactive Interpreter

The interactive interpreter makes it easier to play with Mangle sources.

You can start the interpreter like this: 

```bash
go run interpreter/main/main.go
```

If you want to load files relative to a different directory from your
working directory, you can pass the intended root as a flag:

```bash
go run interpreter/main/main.go --root=$PWD/examples
```

Commands supported:

```
<decl>.            adds declaration to interactive buffer
<useDecl>.         adds package-use declaration to interactive buffer
<clause>.          adds clause to interactive buffer, evaluates.
?<predicate>       looks up predicate name and queries all facts
?<goal>            queries all facts that match goal
::load <path>      pops interactive buffer and loads source file at <path>
::help             display this help text
::pop              reset state to before interactive defs. or last load command
::show <predicate> shows information about predicate
::show all         shows information about all available predicates
<Ctrl-D>           quit
```

Typing `::load` on a interpreter shell is the fastest way to check
syntax and perform some analysis that works on source that are contained
in a single file. There are some shortcomings multiple files and packages,
so for those, it may be better to write a customized interpreter.

Loading will *forget* interactive definitions, load and evaluate sources
and update the store. If you want to keep your definitions, store them in a
file.

`::pop` will forget what was added last, restoring the store to the state before
the last source was loaded. If interactive definitions were entered, they are
always the last set of definitions. This is useful if you made a
mistake and want to undo the changes.

## Command line flags

The interpreter shell supports a number of command-line flags.

When you pass a query to `--exec` (without the leading `?`), the interpreter
will execute the query, print results and exit. If the query returned at least
one result, it prints "#PASS" and the exit code will be 0, else it will
print "#FAIL" and the exit code is non-zero. 

You can "preload" sources by passing them to the `--load` flag as a 
comma-separated list. The preloaded sources will be analyzed as one atomic
set of definitions, which can make a difference if you have multiple files:
a source that is loaded at the prompt is analyzed by itself, it can use all 
previously seen definitions but one is not permitted to add a clause to a
previously defined predicate.

## Notes on the Interactive Buffer

You can either load files, or enter definitions interactively. Every time you
enter something at the interpreter prompt, you are modifying an invisible
buffer.

When you enter a definition and there is no error, all previous state from is
forgotten, all definitions are parsed again. Then all definitions are evaluated
and results can be queried. So you never need to worry about "data dependencies"
when entering definitions.

Adding multiple rules for a predicate within this buffer is possible, however
one cannot add rules for predicates that were defined in files: all rules for a
predicate must come from the same "unit" (either file, or interactive buffer).
This enables arity checking (and type checking).

The interpreter checks that the predicates that one references in the body of
the rule were defined previously and have the right arity. This helps with
catching typos and arity errors. Therefore rules must either be entered in the
order determined by the dependency of the predicates, or one has to declare a
predicate like this:

```
Decl foo(Arg1, Arg2).
mr >Decl foo(X,Y).
defined [foo(A, B)].
mr >bar(X) :- foo(X, _).
defined [foo(A, B) bar(A)].
mr >foo(1,1).
defined [foo(A, B) bar(A)].
mr >?bar
bar(1)
Found 1 entries for bar.
```
