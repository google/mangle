# Grammar Reference

This section provides a comprehensive reference for the Mangle language grammar, addressing [Issue #3](https://github.com/google/mangle/issues/3).

The Mangle language is built on Datalog with extensions for structured data, aggregation, and type declarations. This grammar reference is organized into two main sections:

- **Core Datalog Grammar** - Essential constructs for facts, rules, and basic queries
- **Extended Mangle Grammar** - Complete language including packages, types, and structured data

## Core Datalog Grammar

The core Datalog grammar covers the fundamental constructs that form the foundation of Mangle programs:

### Basic Structure

```ebnf
clause ::= atom ( ':-' clauseBody )? '.'
clauseBody ::= literal ( ',' literal )*
literal ::= atom | '!' atom  
atom ::= predicateName '(' varOrConstant ( ',' varOrConstant )* ')'
varOrConstant ::= variable | constant
predicateName ::= NAME
variable ::= VARIABLE
constant ::= CONSTANT | NUMBER | STRING
```

### Examples

Basic facts:
```mangle
parent(/john, /mary).
age(/john, 35).
```

Rules with variables:
```mangle
grandparent(X, Z) :- parent(X, Y), parent(Y, Z).
adult(Person) :- age(Person, Age), Age >= 18.
```

Negation:
```mangle
unmarried(Person) :- person(Person), !married(Person).
```

## Extended Mangle Grammar

The complete Mangle grammar includes additional constructs for building larger, structured programs:

```ebnf
start ::= program EOF

program ::= packageDecl? useDecl* ( decl | clause )*

packageDecl ::= 'Package' NAME atoms? '!'

useDecl ::= 'Use' NAME atoms? '!'

decl ::= 'Decl' atom descrBlock? boundsBlock* constraintsBlock? '.'

descrBlock ::= 'descr' atoms

boundsBlock ::= 'bound' '[' ( term ',' )* term? ']'

constraintsBlock ::= 'inclusion' atoms

clause ::= atom ( ( ':-' | '⟸' ) clauseBody )? '.'

clauseBody ::= literalOrFml ( ',' literalOrFml )* ','? ( '|>' transform )*

transform ::= 'do' term ( ',' letStmt ( ',' letStmt )* )?
            | letStmt ( ',' letStmt )*

letStmt ::= 'let' VARIABLE '=' term

literalOrFml ::= term ( ( '=' | '!=' | '<' | '<=' | '>' | '>=' ) term )?
               | '!' term

term ::= VARIABLE
       | CONSTANT  
       | NUMBER
       | FLOAT
       | STRING
       | BYTESTRING
       | '[' ( term ',' )* term? ']'
       | '[' ( term ':' term ',' )* ( term ':' term )? ']'
       | '{' ( term ':' term ',' )* ( term ':' term )? '}'
       | DOT_TYPE '<' ( member ',' )* ( member ','? )? '>'
       | NAME '(' ( term ',' )* term? ')'

member ::= term ( ':' term )?
         | 'opt' term ':' term

atom ::= term

atoms ::= '[' ( atom ',' )* atom? ']'

VARIABLE ::= '_' | ( 'A'..'Z' ( 'A'..'Z' | 'a'..'z' | '0'..'9' )* )
NAME ::= ':'? ( 'a'..'z' ) ( 'a'..'z' | 'A'..'Z' | '0'..'9' | ':' | '_' | '.' )*
CONSTANT ::= '/' ( 'a'..'z' | 'A'..'Z' | '0'..'9' | '.' | '-' | '_' | '~' | '%' )+ ( '/' ( 'a'..'z' | 'A'..'Z' | '0'..'9' | '.' | '-' | '_' | '~' | '%' )+ )*
NUMBER ::= '-'? ( '0'..'9' )+
FLOAT ::= '-'? ( '0'..'9' )+ '.' ( '0'..'9' )+ ( ( 'e' | 'E' ) ( '+' | '-' )? ( '0'..'9' )+ )?
STRING ::= '"' ( ~[\"] | '\' . )* '"' | "'" ( ~[\'] | '\' . )* "'" | '`' ( ~[\\] | '\' . )* '`'
BYTESTRING ::= 'b' STRING
DOT_TYPE ::= '.' ( 'A'..'Z' ) ( 'a'..'z' | 'A'..'Z' | '0'..'9' | ':' | '_' | '.' )*
```

## Usage Examples

### Simple Facts and Queries
```mangle
# Facts
parent(/alice, /bob).
parent(/bob, /charlie).
age(/alice, 45).

# Rules
grandparent(X, Z) :- parent(X, Y), parent(Y, Z).
older(X, Y) :- age(X, AgeX), age(Y, AgeY), AgeX > AgeY.
```

### Structured Data
```mangle
# Employee records
employee(.Employee<
  id: 1001,
  name: /john_doe,
  department: /engineering,
  salary: 75000
>).

# Processing with aggregation
dept_avg_salary(Dept, Avg) :-
  employee(.Employee<department: Dept, salary: Salary>),
  |> do fn:avg(Salary), let Avg = _.
```

### Type Declarations
```mangle
Decl employee(id:int, name:string, dept:string, salary:float)
  descr ["Employee information"]
  bound [1000..9999]
  inclusion [engineering, marketing, sales]
  .
```

## Grammar Rules Summary

### Core Rules
- **clause** - Facts and rules (with optional body)
- **atom** - Predicate applications with arguments
- **literal** - Positive or negative atoms
- **variable** - Uppercase identifiers for unknowns
- **constant** - Forward-slash prefixed identifiers

### Extended Rules  
- **program** - Complete Mangle programs with packages
- **decl** - Type and schema declarations
- **transform** - Aggregation and data processing pipelines
- **term** - All value expressions (variables, constants, structured data)
- **member** - Fields in structured types

## Regenerating This Documentation

This grammar reference is generated from the ANTLR grammar file. To regenerate after grammar changes:

```bash
cd /path/to/mangle
python3 scripts/generate_railroad_diagrams.py
```

The generator script will update this file and maintain consistency with the actual parser grammar.

---

*This grammar reference addresses [Issue #3](https://github.com/google/mangle/issues/3) by providing comprehensive syntax documentation integrated into the main documentation system.*
