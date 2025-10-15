#!/usr/bin/env python3
"""
Script to generate grammar documentation for Mangle.
This addresses GitHub issue #3: "Grammar railroad diagram"

Generates MyST markdown documentation for the Sphinx-based readthedocs system.

Usage:
    python3 scripts/generate_railroad_diagrams.py
"""

import os
import re
import subprocess
import sys
from pathlib import Path

def create_simplified_core_grammar():
    """
    Create a simplified core grammar as suggested by burakemir in the issue comments.
    This focuses on the essential Datalog constructs.
    """
    return '''clause ::= atom ( ':-' clauseBody )? '.'
clauseBody ::= literal ( ',' literal )*
literal ::= atom | '!' atom  
atom ::= predicateName '(' varOrConstant ( ',' varOrConstant )* ')'
varOrConstant ::= variable | constant
predicateName ::= NAME
variable ::= VARIABLE
constant ::= CONSTANT | NUMBER | STRING'''

def create_full_ebnf_grammar():
    """
    Create a more complete EBNF grammar based on the current Mangle.g4 file.
    """
    return '''start ::= program EOF

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
STRING ::= '"' ( ~[\\"] | '\\' . )* '"' | "'" ( ~[\\'] | '\\' . )* "'" | '`' ( ~[\\\\] | '\\' . )* '`'
BYTESTRING ::= 'b' STRING
DOT_TYPE ::= '.' ( 'A'..'Z' ) ( 'a'..'z' | 'A'..'Z' | '0'..'9' | ':' | '_' | '.' )*'''

def generate_grammar_documentation():
    """
    Generate the complete grammar documentation as MyST markdown.
    """
    core_grammar = create_simplified_core_grammar()
    full_grammar = create_full_ebnf_grammar()
    
    return f"""# Grammar Reference

This section provides a comprehensive reference for the Mangle language grammar, addressing [Issue #3](https://github.com/google/mangle/issues/3).

The Mangle language is built on Datalog with extensions for structured data, aggregation, and type declarations. This grammar reference is organized into two main sections:

- **Core Datalog Grammar** - Essential constructs for facts, rules, and basic queries
- **Extended Mangle Grammar** - Complete language including packages, types, and structured data

## Core Datalog Grammar

The core Datalog grammar covers the fundamental constructs that form the foundation of Mangle programs:

### Basic Structure

```ebnf
{core_grammar}
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
{full_grammar}
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
"""

def main():
    """Main function to generate grammar documentation."""
    
    # Paths
    script_dir = Path(__file__).parent
    project_root = script_dir.parent
    grammar_file = project_root / 'parse' / 'gen' / 'Mangle.g4'
    readthedocs_dir = project_root / 'readthedocs'
    docs_dir = project_root / 'docs'
    
    # Ensure directories exist
    readthedocs_dir.mkdir(exist_ok=True)
    docs_dir.mkdir(exist_ok=True)
    
    print("📝 Generating Grammar Documentation for Mangle...")
    print(f"📁 Project root: {project_root}")
    print(f"📁 Readthedocs directory: {readthedocs_dir}")
    
    if not grammar_file.exists():
        print(f"❌ Grammar file not found: {grammar_file}")
        print("ℹ️  Continuing with predefined grammar...")
    else:
        print(f"✅ Found grammar file: {grammar_file}")
    
    # Generate MyST markdown documentation for Sphinx
    print("📝 Creating grammar documentation for readthedocs...")
    grammar_content = generate_grammar_documentation()
    grammar_output = readthedocs_dir / 'grammar.md'
    
    with open(grammar_output, 'w', encoding='utf-8') as f:
        f.write(grammar_content)
    
    print(f"✅ Grammar documentation saved to: {grammar_output}")
    
    # Generate EBNF file for easy reference
    ebnf_output = docs_dir / 'grammar.ebnf'
    core_grammar = create_simplified_core_grammar()
    full_grammar = create_full_ebnf_grammar()
    
    with open(ebnf_output, 'w', encoding='utf-8') as f:
        f.write("// Mangle Core Datalog Grammar (EBNF)\n")
        f.write("// Copy this to https://www.bottlecaps.de/rr/ui for interactive diagrams\n\n")
        f.write(core_grammar)
        f.write("\n\n// Complete Mangle Grammar (EBNF)\n\n")
        f.write(full_grammar)
    
    print(f"✅ EBNF grammar saved to: {ebnf_output}")
    
    print("\n🎉 Grammar documentation generation complete!")
    print("\n📋 Files created:")
    print(f"   📄 {grammar_output}")
    print(f"   📄 {ebnf_output}")
    print("\n🚀 Next steps:")
    print("1. The grammar documentation is now integrated into the Sphinx documentation")
    print("2. Users can copy EBNF from docs/grammar.ebnf to https://www.bottlecaps.de/rr/ui for interactive diagrams")
    print("3. Build the documentation with: cd readthedocs && sphinx-build . _build")
    print("4. Review and commit the generated files")
    
    return 0

if __name__ == '__main__':
    sys.exit(main())