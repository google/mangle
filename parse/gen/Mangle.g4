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
    : 'Decl' atom descrBlock? boundsBlock* constraintsBlock? '.';

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
    : atom (':-' clauseBody)? '.'
    ;

clauseBody
    : literalOrFml (',' literalOrFml)* ','? ('|>' transform)?
    ;

transform
    : 'do' term (',' letStmt (',' letStmt)*)?
    | letStmt (',' letStmt)*
    ;

letStmt
    : 'let' VARIABLE '=' term
    ;
    
literalOrFml
   : term (('=' | '!=' | '<' | '<=') term)?
   | '!'term
   ;

term
   : VARIABLE # Var
   | CONSTANT # Const
   | NUMBER # Num
   | FLOAT # Float
   | STRING # Str
   | NAME '(' (term ',')* term? ')' # Appl
   | '[' (term ',')* term? ']' # List
   | '[' (term ':' term ',')* (term ':' term)? ']' # Map
   | '{' (term ':' term ',')* (term ':' term)? '}' # Struct
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
EQ : '=';
BANGEQ : '!=';
COMMA : ',';
BANG : '!';
LESS : '<';
LESSEQ : '<=';
GREATER : '>';
GREATEREQ : '>=';
COLONDASH : ':-';
NEWLINE : '\n';
PIPEGREATER : '|>';

fragment LETTER : 'A'..'Z' | 'a'..'z' ;
fragment DIGIT  : '0'..'9' ;

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


fragment CONSTANT_CHAR : LETTER | DIGIT | '.' | '-' | '_' | '~' | '%';
CONSTANT : '/' CONSTANT_CHAR+ ('/' CONSTANT_CHAR+)*;


STRING : (SHORT_STRING | LONG_STRING);

/// shortstring     ::=  "'" shortstringitem* "'" | '"' shortstringitem* '"'
/// shortstringitem ::=  shortstringchar | stringescapeseq
/// shortstringchar ::=  <any source character except "\" or newline or the quote>
fragment SHORT_STRING
 : '\'' ( STRING_ESCAPE_SEQ | ~[\\\r\n\f'] )* '\''
 | '"' ( STRING_ESCAPE_SEQ | ~[\\\r\n\f"] )* '"'
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

/// stringescapeseq ::=  "\" <any source character>
fragment STRING_ESCAPE_SEQ
 : '\\' .
 | '\\' NEWLINE
 ;


