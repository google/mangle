// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

use std::collections::HashMap;

use bumpalo::Bump;

// Immutable representation of syntax.
// We use references instead of a smart pointer,
// relying on an arena that holds everything.
// This way we can use pattern matching.

// Unit is a source unit.
// It consists of declarations and clauses.
// After parsing, all source units for a Mangle package
// are merged into one, so unit can also be seen
// as translation unit.
#[derive(Debug)]
pub struct Unit<'a> {
    pub decls: &'a [&'a Decl<'a>],
    pub clauses: &'a [&'a Clause<'a>],
}

// Predicate, package and use declarations.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct Decl<'a> {
    pub atom: &'a Atom<'a>,
    pub descr: &'a [&'a Atom<'a>],
    pub bounds: Option<&'a [&'a BoundDecl<'a>]>,
    pub constraints: Option<&'a Constraints<'a>>,
}

#[derive(Debug, PartialEq)]
pub struct BoundDecl<'a> {
    pub base_terms: &'a [&'a BaseTerm<'a>],
}

//
#[derive(Debug, Clone, PartialEq)]
pub struct Constraints<'a> {
    // All of these must hold.
    pub consequences: &'a [&'a Atom<'a>],
    // In addition to consequences, at least one of these must hold.
    pub alternatives: &'a [&'a [&'a Atom<'a>]],
}

#[derive(Debug)]
pub struct Clause<'a> {
    pub head: &'a Atom<'a>,
    pub premises: &'a [&'a Term<'a>],
    pub transform: &'a [&'a TransformStmt<'a>],
}

#[derive(Debug)]
pub struct TransformStmt<'a> {
    pub var: Option<&'a str>,
    pub app: &'a BaseTerm<'a>,
}

// Term that may appear on righthand-side of a clause.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Term<'a> {
    Atom(&'a Atom<'a>),
    NegAtom(&'a Atom<'a>),
    Eq(&'a BaseTerm<'a>, &'a BaseTerm<'a>),
    Ineq(&'a BaseTerm<'a>, &'a BaseTerm<'a>),
}

impl<'a> std::fmt::Display for Term<'a> {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Term::Atom(atom) => write!(f, "{atom}"),
            Term::NegAtom(atom) => write!(f, "!{atom}"),
            Term::Eq(left, right) => write!(f, "{left} = {right}"),
            Term::Ineq(left, right) => write!(f, "{left} != {right}"),
        }
    }
}

impl<'a> Term<'a> {
    pub fn apply_subst<'b>(
        &'a self,
        bump: &'b Bump,
        subst: &HashMap<&'a str, &'a BaseTerm<'a>>,
    ) -> &'b Term<'b> {
        &*bump.alloc(match self {
            Term::Atom(atom) => Term::Atom(atom.apply_subst(bump, subst)),
            Term::NegAtom(atom) => Term::NegAtom(atom.apply_subst(bump, subst)),
            Term::Eq(left, right) => Term::Eq(
                left.apply_subst(bump, subst),
                right.apply_subst(bump, subst),
            ),
            Term::Ineq(left, right) => Term::Ineq(
                left.apply_subst(bump, subst),
                right.apply_subst(bump, subst),
            ),
        })
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum BaseTerm<'a> {
    Const(Const<'a>),
    Variable(&'a str),
    ApplyFn(FunctionSym<'a>, &'a [&'a BaseTerm<'a>]),
}

impl<'a> BaseTerm<'a> {
    pub fn apply_subst<'b>(
        &'a self,
        bump: &'b Bump,
        subst: &HashMap<&'a str, &'a BaseTerm<'a>>,
    ) -> &'b BaseTerm<'b> {
        match self {
            BaseTerm::Const(_) => copy_base_term(bump, self),
            BaseTerm::Variable(v) => subst
                .get(v)
                .map_or(copy_base_term(bump, self), |b| copy_base_term(bump, b)),
            BaseTerm::ApplyFn(fun, args) => {
                let args: Vec<&'b BaseTerm<'b>> = args
                    .iter()
                    .map(|arg| arg.apply_subst(bump, subst))
                    .collect();
                copy_base_term(bump, &BaseTerm::ApplyFn(*fun, &args))
            }
        }
    }
}

impl<'a> std::fmt::Display for BaseTerm<'a> {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            BaseTerm::Const(c) => write!(f, "{c}"),
            BaseTerm::Variable(v) => write!(f, "{v}"),
            BaseTerm::ApplyFn(FunctionSym { name: n, .. }, args) => write!(
                f,
                "{n}({})",
                args.iter()
                    .map(|x| x.to_string())
                    .collect::<Vec<_>>()
                    .join(",")
            ),
        }
    }
}
#[derive(Debug, Clone, Copy, PartialEq)]
pub enum Const<'a> {
    Name(&'a str),
    Bool(bool),
    Number(i64),
    Float(f64),
    String(&'a str),
    Bytes(&'a [u8]),
    List(&'a [&'a Const<'a>]),
    Map {
        keys: &'a [&'a Const<'a>],
        values: &'a [&'a Const<'a>],
    },
    Struct {
        fields: &'a [&'a str],
        values: &'a [&'a Const<'a>],
    },
}

impl<'a> Eq for Const<'a> {}

impl<'a> std::fmt::Display for Const<'a> {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match *self {
            Const::Name(v) => write!(f, "{v}"),
            Const::Bool(v) => write!(f, "{v}"),
            Const::Number(v) => write!(f, "{v}"),
            Const::Float(v) => write!(f, "{v}"),
            Const::String(v) => write!(f, "{v}"),
            Const::Bytes(v) => write!(f, "{:?}", v),
            Const::List(v) => {
                write!(
                    f,
                    "[{}]",
                    v.iter()
                        .map(|x| x.to_string())
                        .collect::<Vec<_>>()
                        .join(", ")
                )
            }
            Const::Map { keys: _, values: _ } => write!(f, "{{...}}"),
            Const::Struct {
                fields: _,
                values: _,
            } => write!(f, "{{...}}"),
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct PredicateSym<'a> {
    pub name: &'a str,
    pub arity: Option<u8>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct FunctionSym<'a> {
    pub name: &'a str,
    pub arity: Option<u8>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Atom<'a> {
    pub sym: PredicateSym<'a>,

    pub args: &'a [&'a BaseTerm<'a>],
}

impl<'a> Atom<'a> {
    // A fact matches a query if there is a substitution s.t. subst(query) = fact.
    // We assume that self is a fact and query_args are the arguments of a query,
    // i.e. only variables and constants.
    pub fn matches(&'a self, query_args: &[&BaseTerm]) -> bool {
        for (fact_arg, query_arg) in self.args.iter().zip(query_args.iter()) {
            if let BaseTerm::Const(_) = query_arg {
                if fact_arg != query_arg {
                    return false;
                }
            }
        }
        true
    }

    pub fn apply_subst<'b>(
        &'a self,
        bump: &'b Bump,
        subst: &HashMap<&'a str, &'a BaseTerm<'a>>,
    ) -> &'b Atom<'b> {
        let args: Vec<&'b BaseTerm<'b>> = self
            .args
            .iter()
            .map(|arg| arg.apply_subst(bump, subst))
            .collect();
        let args = &*bump.alloc_slice_copy(&args);
        bump.alloc(Atom {
            sym: copy_predicate_sym(bump, self.sym),
            args,
        })
    }
}

impl<'a> std::fmt::Display for Atom<'a> {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}(", self.sym.name)?;
        for arg in self.args {
            write!(f, "{arg}")?;
        }
        write!(f, ")")
    }
}

pub fn copy_predicate_sym<'dest>(bump: &'dest Bump, p: PredicateSym) -> PredicateSym<'dest> {
    PredicateSym {
        name: bump.alloc_str(p.name),
        arity: p.arity,
    }
}

// Copies BaseTerm to another Bump
pub fn copy_atom<'dest, 'src>(bump: &'dest Bump, atom: &'src Atom<'src>) -> &'dest Atom<'dest> {
    let args: Vec<_> = atom
        .args
        .iter()
        .map(|arg| copy_base_term(bump, arg))
        .collect();
    let args = &*bump.alloc_slice_copy(&args);
    bump.alloc(Atom {
        sym: copy_predicate_sym(bump, atom.sym),
        args,
    })
}

// Copies BaseTerm to another Bump
pub fn copy_base_term<'dest, 'src>(
    bump: &'dest Bump,
    b: &'src BaseTerm<'src>,
) -> &'dest BaseTerm<'dest> {
    match b {
        BaseTerm::Const(c) =>
        // Should it be a reference?
        {
            bump.alloc(BaseTerm::Const(*copy_const(bump, c)))
        }
        BaseTerm::Variable(s) => bump.alloc(BaseTerm::Variable(bump.alloc_str(s))),
        BaseTerm::ApplyFn(fun, args) => {
            let fun = FunctionSym {
                name: bump.alloc_str(fun.name),
                arity: fun.arity,
            };
            let args: Vec<_> = args.iter().map(|a| copy_base_term(bump, a)).collect();
            let args = bump.alloc_slice_copy(&args);
            bump.alloc(BaseTerm::ApplyFn(fun, args))
        }
    }
}

// Copies Const to another Bump
pub fn copy_const<'dest, 'src>(bump: &'dest Bump, c: &'src Const<'src>) -> &'dest Const<'dest> {
    match c {
        Const::Name(name) => {
            let name = &*bump.alloc_str(name);
            bump.alloc(Const::Name(name))
        }
        Const::Bool(b) => bump.alloc(Const::Bool(*b)),
        Const::Number(n) => bump.alloc(Const::Number(*n)),
        Const::Float(f) => bump.alloc(Const::Float(*f)),
        Const::String(s) => {
            let s = &*bump.alloc_str(s);
            bump.alloc(Const::String(s))
        }
        Const::Bytes(b) => {
            let b = &*bump.alloc_slice_copy(b);
            bump.alloc(Const::Bytes(b))
        }
        Const::List(cs) => {
            let cs: Vec<_> = cs.iter().map(|c| copy_const(bump, c)).collect();
            let cs = &*bump.alloc_slice_copy(&cs);
            bump.alloc(Const::List(cs))
        }
        Const::Map { keys, values } => {
            let keys: Vec<_> = keys.iter().map(|c| copy_const(bump, c)).collect();
            let keys = &*bump.alloc_slice_copy(&keys);

            let values: Vec<_> = values.iter().map(|c| copy_const(bump, c)).collect();
            let values = &*bump.alloc_slice_copy(&values);

            bump.alloc(Const::Map { keys, values })
        }
        Const::Struct { fields, values } => {
            let fields: Vec<_> = fields.iter().map(|s| &*bump.alloc_str(s)).collect();
            let fields = &*bump.alloc_slice_copy(&fields);

            let values: Vec<_> = values.iter().map(|c| copy_const(bump, c)).collect();
            let values = &*bump.alloc_slice_copy(&values);

            bump.alloc(Const::Struct { fields, values })
        }
    }
}

pub fn copy_transform<'dest, 'src>(
    bump: &'dest Bump,
    stmt: &'src TransformStmt<'src>,
) -> &'dest TransformStmt<'dest> {
    let TransformStmt { var, app } = stmt;
    let var = var.map(|s| &*bump.alloc_str(s));
    let app = copy_base_term(bump, app);
    bump.alloc(TransformStmt { var, app })
}

pub fn copy_clause<'dest, 'src>(
    bump: &'dest Bump,
    clause: &'src Clause<'src>,
) -> &'dest Clause<'dest> {
    let Clause {
        head,
        premises,
        transform,
    } = clause;
    let premises: Vec<_> = premises.iter().map(|x| copy_term(bump, x)).collect();
    let transform: Vec<_> = transform.iter().map(|x| copy_transform(bump, x)).collect();
    bump.alloc(Clause {
        head: copy_atom(bump, head),
        premises: &*bump.alloc_slice_copy(&premises),
        transform: &*bump.alloc_slice_copy(&transform),
    })
}

fn copy_term<'dest, 'src>(bump: &'dest Bump, term: &'src Term<'src>) -> &'dest Term<'dest> {
    match term {
        Term::Atom(atom) => {
            let atom = copy_atom(bump, atom);
            bump.alloc(Term::Atom(atom))
        }
        Term::NegAtom(atom) => {
            let atom = copy_atom(bump, atom);
            bump.alloc(Term::NegAtom(atom))
        }
        Term::Eq(left, right) => {
            let left = copy_base_term(bump, left);
            let right = copy_base_term(bump, right);
            bump.alloc(Term::Eq(left, right))
        }
        Term::Ineq(left, right) => {
            let left = copy_base_term(bump, left);
            let right = copy_base_term(bump, right);
            bump.alloc(Term::Ineq(left, right))
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use bumpalo::Bump;
    use googletest::prelude::*;

    #[test]
    fn copying_atom_works() {
        let bump = Bump::new();
        let foo = &*bump.alloc(BaseTerm::Const(Const::Name("/foo")));
        let bar = bump.alloc(PredicateSym {
            name: "bar",
            arity: Some(1),
        });
        let bar_args = bump.alloc_slice_copy(&[foo]);
        let head = bump.alloc(Atom {
            sym: *bar,
            args: &*bar_args,
        });
        assert_that!("bar(/foo)", eq(head.to_string()));
    }

    #[test]
    fn atom_display_works() {
        let bar = BaseTerm::Const(Const::Name("/bar"));
        assert_that!(bar, displays_as(eq("/bar")));

        let atom = Atom {
            sym: PredicateSym {
                name: "foo",
                arity: Some(1),
            },
            args: &[&bar],
        };
        assert_that!(atom, displays_as(eq("foo(/bar)")));

        let tests = vec![
            (Term::Atom(&atom), "foo(/bar)"),
            (Term::NegAtom(&atom), "!foo(/bar)"),
            (Term::Eq(&bar, &bar), "/bar = /bar"),
            (Term::Ineq(&bar, &bar), "/bar != /bar"),
        ];
        for (term, s) in tests {
            assert_that!(term, displays_as(eq(s)));
        }
    }
}
