// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

grammar Mangle;

// Grammar rules

start
    : program EOF
    ;

program
    : packageDecl? useDecl* (decl | clause)*
    ;

packageDecl
    : 'Package' NAME atoms? '!'
    ;

useDecl
    : 'Use' NAME atoms? '!'
    ;

decl
    : 'Decl' atom 'temporal'? descrBlock? boundsBlock* constraintsBlock? '.';

descrBlock
    : 'descr' atoms
    ;

boundsBlock
    : 'bound' '[' (term ',')* term? ']'
    ;

constraintsBlock
    : 'inclusion' atoms
    ;

clause
    : atom temporalAnnotation? ((':-'|LONGLEFTDOUBLEARROW) clauseBody)? '.'
    ;

// Temporal annotation for facts: @[start, end] or @[point]
temporalAnnotation
    : '@' '[' temporalBound (',' temporalBound)? ']'
    ;

// A temporal bound can be a timestamp, duration, variable (including _ for unbounded), or 'now'
temporalBound
    : TIMESTAMP           // 2024-01-15 or 2024-01-15T10:30:00
    | DURATION            // 7d, 24h, 30m, 1s
    | VARIABLE            // Bind to variable, or _ for unbounded
    | 'now'               // Current evaluation time
    ;

clauseBody
    : literalOrFml (',' literalOrFml)* ','? ('|>' transform)*
    ;

transform
    : 'do' term (',' letStmt (',' letStmt)*)?
    | letStmt (',' letStmt)*
    ;

letStmt
    : 'let' VARIABLE '=' term
    ;
    
literalOrFml
   : temporalOperator? term temporalAnnotation? ((EQ | BANGEQ | LESS | LESSEQ | GREATER | GREATEREQ) term)?
   | '!'term
   ;

// Temporal operators for querying the past/future
temporalOperator
   : DIAMONDMINUS '[' temporalBound ',' temporalBound ']'   // <-[0d, 7d] - sometime in past
   | BOXMINUS '[' temporalBound ',' temporalBound ']'       // [-[0d, 7d] - always in past
   | DIAMONDPLUS '[' temporalBound ',' temporalBound ']'    // <+[0d, 7d] - sometime in future
   | BOXPLUS '[' temporalBound ',' temporalBound ']'        // [+[0d, 7d] - always in future
   ;

term
   : VARIABLE # Var
   | CONSTANT # Const
   | NUMBER # Num
   | FLOAT # Float
   | STRING # Str
   | BYTESTRING # BStr
   | '[' (term ',')* term? ']'                     # List
   | '[' (term ':' term ',')* (term ':' term)? ']' # Map
   | '{' (term ':' term ',')* (term ':' term)? '}' # Struct
   | DOT_TYPE '<' (member ',')* (member ','?)? '>'  # DotType
   | NAME '(' (term ',')* term? ')'       # Appl
   ;

member
   : term (':' term)? 
   | 'opt' term ':' term
   ;

// Implementation enforces that this is an atom NAME(...)
atom
   : term
   ;

atoms
   :  '[' (atom ',')* atom? ']'
   ;

// lexer rules

WHITESPACE : ( '\t' | ' ' | '\r' | '\n'| '\u000C' )+ -> channel(HIDDEN) ;
COMMENT : '#' (~'\n')* -> channel(HIDDEN) ;

LONGLEFTDOUBLEARROW : '⟸';  // \u27F8

PACKAGE : 'Package';
USE : 'Use';
DECL : 'Decl';
BOUND : 'bound';
LET : 'let';
DO : 'do';
LPAREN : '(';
RPAREN : ')';
LBRACKET : '[';
RBRACKET : ']';
LBRACE : '{';
RBRACE : '}';
EQ : '=';
BANGEQ : '!=';
COMMA : ',';
BANG : '!';
LESSEQ : '<=';        // Must come before LESS
LESS : '<';
GREATEREQ : '>=';     // Must come before GREATER
GREATER : '>';
COLONDASH : ':-';
NEWLINE : '\n';
PIPEGREATER : '|>';
AT : '@';
DIAMONDMINUS : '<-';  // Must come before LESS
DIAMONDPLUS : '<+';   // Must come before LESS
BOXMINUS : '[-';      // Must come before LBRACKET
BOXPLUS : '[+';       // Must come before LBRACKET

fragment LETTER : 'A'..'Z' | 'a'..'z' ;
fragment DIGIT  : '0'..'9' ;

// Temporal tokens - must come before NUMBER to take precedence
// Timestamp: 2024-01-15 or 2024-01-15T10:30:00 or 2024-01-15T10:30:00Z
TIMESTAMP : DIGIT DIGIT DIGIT DIGIT '-' DIGIT DIGIT '-' DIGIT DIGIT
            ('T' DIGIT DIGIT ':' DIGIT DIGIT ':' DIGIT DIGIT ('.' DIGIT+)? 'Z'?)? ;

// Duration: 7d, 24h, 30m, 1s, 500ms
DURATION : DIGIT+ ('d' | 'h' | 'm' | 's' | 'ms') ;

NUMBER : '-'? DIGIT (DIGIT)*;
FLOAT : '-'? (DIGIT)+ '.' (DIGIT)+ EXPONENT?
      | '-'? '.' (DIGIT)+ EXPONENT?;

fragment
EXPONENT : ('e'|'E') ('+'|'-')? DIGIT+ ;

fragment VARIABLE_START : 'A'..'Z' ;
fragment VARIABLE_CHAR : LETTER | DIGIT ;

VARIABLE : '_' | (VARIABLE_START VARIABLE_CHAR*);

fragment NAME_CHAR : LETTER | DIGIT | ':' | '_' ;
NAME : ':'? ('a'..'z') ( NAME_CHAR | ('.' NAME_CHAR) )*;

TYPENAME : 'A'..'Z' ( NAME_CHAR | ('.' NAME_CHAR) )*;
DOT_TYPE : '.' TYPENAME;

fragment CONSTANT_CHAR : LETTER | DIGIT | '.' | '-' | '_' | '~' | '%';
CONSTANT : '/' CONSTANT_CHAR+ ('/' CONSTANT_CHAR+)*;

STRING : (SHORT_STRING | LONG_STRING);
BYTESTRING : 'b'STRING;

/// shortstring     ::=  "'" shortstringitem* "'" | '"' shortstringitem* '"'
/// shortstringitem ::=  shortstringchar | stringescapeseq
/// shortstringchar ::=  <any source character except "\" or newline or the quote>
fragment SHORT_STRING
 : '\'' ( STRING_ESCAPE_SEQ | ~[\\'] )* '\''
 | '"' ( STRING_ESCAPE_SEQ | ~[\\"] )* '"'
 ;
/// longstring      ::=  "`" longstringitem* "`"
fragment LONG_STRING
 : '`' LONG_STRING_ITEM*? '`'
 ;

/// longstringitem  ::=  longstringchar | stringescapeseq
fragment LONG_STRING_ITEM
 : LONG_STRING_CHAR
 | STRING_ESCAPE_SEQ
 ;

/// longstringchar  ::=  <any source character except "\">
fragment LONG_STRING_CHAR
 : ~'\\'
 ;

/// stringescapeseq ::=  "\[nt"'\]" | byteescape | unicodeescape | "\<newline>"
/// byteescape ::= "\x" hex hex
/// unicodeescape ::= "\u{" hex hex hex hex hex? hex? "}"
fragment STRING_ESCAPE_SEQ
 : '\\' 'n'
 | '\\' 't'
 | '\\' '"'
 | '\\' '\''
 | '\\' '\\'
 | '\\' 'x' HEXDIGIT HEXDIGIT
 | '\\' 'u' '{' HEXDIGIT HEXDIGIT HEXDIGIT HEXDIGIT HEXDIGIT? HEXDIGIT? '}' 
 | '\\' NEWLINE
 ;

fragment HEXDIGIT : 'a'..'f' | '0'..'9';

