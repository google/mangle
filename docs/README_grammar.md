# Mangle Grammar Railroad Diagrams

This directory contains railroad diagram documentation for the Mangle language grammar, addressing [Issue #3](https://github.com/google/mangle/issues/3).

## Files

- **`grammar_railroad_core.html`** - Interactive railroad diagrams for core Datalog syntax
- **`grammar_railroad_full.html`** - Interactive railroad diagrams for complete Mangle grammar  
- **`grammar.ebnf`** - EBNF grammar file for use with external tools
- **`README_grammar.md`** - This documentation file

## Quick Start

1. **View Static Diagrams**: Open the HTML files in your web browser
2. **Interactive Diagrams**: 
   - Open either HTML file in your browser
   - Click the "📋 Copy Grammar" button
   - Go to [https://www.bottlecaps.de/rr/ui](https://www.bottlecaps.de/rr/ui)
   - Paste the grammar in the "Edit Grammar" tab
   - Click "View Diagram" for beautiful interactive railroad diagrams!

## Core Datalog Grammar

The **core grammar** (`grammar_railroad_core.html`) focuses on essential Datalog constructs:
- **Clauses and atoms** - Basic facts and rules
- **Literals** - Positive and negative conditions  
- **Variables and constants** - Data elements
- **Predicates** - Relation names

Example core syntax:
```prolog
parent(john, mary).           % Fact
grandparent(X, Z) :- parent(X, Y), parent(Y, Z).  % Rule with variables
```

## Complete Grammar  

The **full grammar** (`grammar_railroad_full.html`) includes all Mangle language features:
- **Package declarations** - Module system
- **Type declarations** - Schema definitions with bounds and constraints
- **Aggregation transforms** - Data processing pipelines
- **Structured data** - Lists, maps, and structured types
- **Comparison operators** - Numeric and string comparisons

## Grammar Rules Overview

### Core Rules
- `clause` - Top-level facts and rules
- `atom` - Predicate applications  
- `literal` - Atoms with optional negation
- `varOrConstant` - Variable or constant values

### Extended Rules
- `program` - Complete Mangle programs
- `decl` - Type and schema declarations
- `transform` - Aggregation and data processing
- `term` - All value expressions (atoms, lists, maps, etc.)

## Regenerating Diagrams

To regenerate these diagrams after grammar changes:

```bash
cd /path/to/mangle
python3 scripts/generate_railroad_diagrams.py
```

## References

- 🎫 [Original Issue #3](https://github.com/google/mangle/issues/3) - Request for railroad diagrams
- 🚂 [Bottlecaps Railroad Diagram Generator](https://www.bottlecaps.de/rr/ui) - Interactive diagram tool
- 🔄 [ANTLR to EBNF Converter](https://www.bottlecaps.de/convert/) - Grammar conversion tool
- 📚 [Mangle Documentation](https://github.com/google/mangle) - Main project documentation

## Contributing

When modifying the Mangle grammar:
1. Update the ANTLR grammar file (`parse/gen/Mangle.g4`)
2. Run the railroad diagram generator
3. Review the generated diagrams for clarity
4. Update documentation as needed

---
*Generated automatically by `scripts/generate_railroad_diagrams.py`*
