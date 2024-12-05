# Type Expressions

This section describes type expressions in more detail and introduces
more ways to construct type expressions.

For basic types like `/number`, we have used the same syntax as for names,
whereas for constructed types, we have introduced a syntax that involves
a *type constructor* `fn:Pair(/number, /string)`.

## Type variables

A type expressions may contain *type variables*. These start with a
capital letter.

**Examples of types containing type variables:*
```
X
fn:Pair(Y, /string)
```

## Any type

There is a type expression `/any`. Any datum has the type `/any`.

## Singleton types

For every name that appears in a program, there is a unique
type `fn:Singleton(`{math}`name``)`.

**Examples**

```
fn:Singleton(/foo)
```

## Union types

If {math}`T_1,...,T_n` are type expressions, then
`fn:Union(`{math}`T_1,...,T_n``)` is a union type.

There is an empty union type `fn:Union()`.

**Examples**

```
fn:Union()
fn:Union(/name, /string)
fn:Union(fn:Singleton(/foo), fn:Singleton(/bar))
```

# Definition of Type Expressions

Type expressions are the expressions that we can build from type variables,
basic types, `/any` and type constructors.

