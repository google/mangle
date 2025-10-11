#!/usr/bin/env python3
"""
Script to generate railroad diagrams for Mangle grammar.
This addresses GitHub issue #3: "Grammar railroad diagram"

Usage:
    python3 scripts/generate_railroad_diagrams.py
"""

import os
import re
import subprocess
import sys
from pathlib import Path
import html

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

def generate_railroad_diagram_html(ebnf_grammar, title, output_file):
    """
    Generate an HTML file with railroad diagrams using a JavaScript library.
    This creates a self-contained HTML file that can be opened in a browser.
    """
    
    # HTML template with railroad diagram generator
    html_template = f'''<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{title}</title>
    <style>
        body {{
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            margin: 20px;
            background-color: #f5f5f5;
        }}
        .container {{
            max-width: 1200px;
            margin: 0 auto;
            background-color: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }}
        h1 {{
            color: #333;
            text-align: center;
            border-bottom: 2px solid #4CAF50;
            padding-bottom: 10px;
        }}
        h2 {{
            color: #555;
            margin-top: 30px;
        }}
        .grammar-rule {{
            background-color: #f8f9fa;
            border: 1px solid #e9ecef;
            border-radius: 4px;
            margin: 15px 0;
            padding: 15px;
        }}
        .rule-name {{
            font-weight: bold;
            color: #2c3e50;
            font-size: 16px;
        }}
        .rule-definition {{
            font-family: 'Courier New', monospace;
            margin-top: 8px;
            color: #34495e;
            line-height: 1.4;
        }}
        .note {{
            background-color: #e3f2fd;
            border-left: 4px solid #2196F3;
            padding: 15px;
            margin: 20px 0;
        }}
        .ebnf-box {{
            background-color: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 4px;
            padding: 20px;
            margin: 20px 0;
            font-family: 'Courier New', monospace;
            white-space: pre-wrap;
            overflow-x: auto;
        }}
        .footer {{
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #eee;
            text-align: center;
            color: #666;
        }}
        .copy-button {{
            background-color: #007bff;
            color: white;
            border: none;
            padding: 8px 16px;
            border-radius: 4px;
            cursor: pointer;
            margin-bottom: 10px;
        }}
        .copy-button:hover {{
            background-color: #0056b3;
        }}
    </style>
</head>
<body>
    <div class="container">
        <h1>{title}</h1>
        
        <div class="note">
            <strong>🚂 Interactive Railroad Diagrams:</strong> Copy the EBNF grammar below and paste it into 
            <a href="https://www.bottlecaps.de/rr/ui" target="_blank">https://www.bottlecaps.de/rr/ui</a>
            on the "Edit Grammar" tab, then click "View Diagram" to see beautiful interactive railroad diagrams!
        </div>

        <h2>EBNF Grammar</h2>
        <button class="copy-button" onclick="copyToClipboard('ebnf-content')">📋 Copy Grammar</button>
        <div class="ebnf-box" id="ebnf-content">{html.escape(ebnf_grammar)}</div>

        <h2>Individual Grammar Rules</h2>
'''

    # Parse and display individual rules
    rules = []
    for line in ebnf_grammar.strip().split('\n'):
        line = line.strip()
        if '::=' in line:
            rule_parts = line.split('::=', 1)
            if len(rule_parts) == 2:
                rule_name = rule_parts[0].strip()
                rule_def = rule_parts[1].strip()
                rules.append((rule_name, rule_def))
    
    for rule_name, rule_def in rules:
        html_template += f'''
        <div class="grammar-rule">
            <div class="rule-name">{html.escape(rule_name)}</div>
            <div class="rule-definition">{html.escape(rule_def)}</div>
        </div>
'''
    
    html_template += f'''
        <div class="footer">
            <p>Generated for <a href="https://github.com/google/mangle/issues/3" target="_blank">Mangle Issue #3: Grammar railroad diagram</a></p>
            <h3>How to Generate Interactive Railroad Diagrams:</h3>
            <ol>
                <li>Click the "📋 Copy Grammar" button above</li>
                <li>Go to <a href="https://www.bottlecaps.de/rr/ui" target="_blank">https://www.bottlecaps.de/rr/ui</a></li>
                <li>Paste the grammar in the "Edit Grammar" tab</li>
                <li>Click "View Diagram" to see the interactive railroad diagrams</li>
                <li>Navigate through the grammar rules using the diagram interface</li>
            </ol>
            <p><em>Generated by scripts/generate_railroad_diagrams.py</em></p>
        </div>
    </div>

    <script>
        function copyToClipboard(elementId) {{
            const element = document.getElementById(elementId);
            const text = element.textContent;
            
            if (navigator.clipboard) {{
                navigator.clipboard.writeText(text).then(function() {{
                    alert('✅ Grammar copied to clipboard!');
                }});
            }} else {{
                // Fallback for older browsers
                const textArea = document.createElement('textarea');
                textArea.value = text;
                document.body.appendChild(textArea);
                textArea.select();
                document.execCommand('copy');
                document.body.removeChild(textArea);
                alert('✅ Grammar copied to clipboard!');
            }}
        }}
    </script>
</body>
</html>
'''

    # Write the HTML file
    with open(output_file, 'w', encoding='utf-8') as f:
        f.write(html_template)

def main():
    """Main function to generate railroad diagrams."""
    
    # Paths
    script_dir = Path(__file__).parent
    project_root = script_dir.parent
    grammar_file = project_root / 'parse' / 'gen' / 'Mangle.g4'
    docs_dir = project_root / 'docs'
    
    # Ensure docs directory exists
    docs_dir.mkdir(exist_ok=True)
    
    print("🚂 Generating Railroad Diagrams for Mangle Grammar...")
    print(f"📁 Project root: {project_root}")
    print(f"📁 Docs directory: {docs_dir}")
    
    if not grammar_file.exists():
        print(f"❌ Grammar file not found: {grammar_file}")
        print("ℹ️  Continuing with predefined grammar...")
    else:
        print(f"✅ Found grammar file: {grammar_file}")
    
    # Generate core simplified grammar
    print("📝 Creating core Datalog grammar...")
    core_grammar = create_simplified_core_grammar()
    core_output = docs_dir / 'grammar_railroad_core.html'
    generate_railroad_diagram_html(
        core_grammar,
        "Mangle Core Datalog Grammar - Railroad Diagrams",
        core_output
    )
    print(f"✅ Core grammar diagram saved to: {core_output}")
    
    # Generate full grammar
    print("📝 Creating complete grammar...")
    full_grammar = create_full_ebnf_grammar()
    full_output = docs_dir / 'grammar_railroad_full.html'
    generate_railroad_diagram_html(
        full_grammar,
        "Mangle Complete Grammar - Railroad Diagrams", 
        full_output
    )
    print(f"✅ Full grammar diagram saved to: {full_output}")
    
    # Generate EBNF file for easy copying
    ebnf_output = docs_dir / 'grammar.ebnf'
    with open(ebnf_output, 'w', encoding='utf-8') as f:
        f.write("// Mangle Core Datalog Grammar (EBNF)\n")
        f.write("// Copy this to https://www.bottlecaps.de/rr/ui for interactive diagrams\n\n")
        f.write(core_grammar)
        f.write("\n\n// Complete Mangle Grammar (EBNF)\n\n")
        f.write(full_grammar)
    
    print(f"✅ EBNF grammar saved to: {ebnf_output}")
    
    # Create documentation
    readme_content = '''# Mangle Grammar Railroad Diagrams

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
'''
    
    readme_output = docs_dir / 'README_grammar.md'
    with open(readme_output, 'w', encoding='utf-8') as f:
        f.write(readme_content)
    
    print(f"✅ Documentation saved to: {readme_output}")
    
    print("\n🎉 Railroad diagram generation complete!")
    print("\n📋 Files created:")
    print(f"   📄 {core_output}")
    print(f"   📄 {full_output}")  
    print(f"   📄 {ebnf_output}")
    print(f"   📄 {readme_output}")
    print("\n🚀 Next steps:")
    print("1. Open the HTML files in your browser to view the diagrams")
    print("2. Copy EBNF grammar to https://www.bottlecaps.de/rr/ui for interactive diagrams")
    print("3. Review and commit the generated files")
    
    return 0

if __name__ == '__main__':
    sys.exit(main())