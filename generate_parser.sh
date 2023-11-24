#!/bin/sh
#
# Uncomment if you want to fetch the file:
#wget http://www.antlr.org/download/antlr-4.13.1-complete.jar
#
# Or update path to antlr jar, if necessary.
alias antlr4='java -jar antlr-4.13.1-complete.jar'
antlr4 -Dlanguage=Go -package gen -o ./ parse/gen/Mangle.g4 -visitor
