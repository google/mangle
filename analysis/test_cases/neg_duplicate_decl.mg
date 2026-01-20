# Test case for issue #25 - duplicate declarations should cause an error

Decl foo(X, Y, Z) descr [ extensional() ] bound [/x, /y, /z].

# This should cause an error - duplicate declaration
Decl foo(X, Y, Z) descr [ extensional() ].

foo(1, 2, 3).