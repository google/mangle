# Grammar

```

rule ::= atom '.'
       | atom  ':-' rhs

rhs ::= litOrFml (',' litOrFml)* '.'
      | litOrFml (',' litOrFml)* '|>' transforms

transforms ::= stmts '.'
             | 'do' apply-expr ',' stmts '.'

stmts ::= stmt (',' stmt)*
stmt ::= 'let' var '=' expr

litOrFml ::= atom | '!' atom | cmp

atom ::= pred-ident '(' exprs? ')'

expr ::= apply-expr
       | constant
       | var

exprs ::= expr (',' expr)*

apply-expr ::= fn-ident '(' exprs? ')'
             | [ exprs? ]

cmp ::= expr op expr

op ::= '=' | '!=' | '<' | '>'

constant   ::= name, number or string constant
fn-ident   ::= ident (starting with 'fn:')
pred-ident ::= ident (not starting with 'fn:')

```
