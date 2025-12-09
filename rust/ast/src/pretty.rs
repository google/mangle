// Copyright 2025 Google LLC
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

use crate::{
    Arena, Atom, BaseTerm, BoundDecl, Clause, Const, Constraints, Decl, FunctionIndex,
    PredicateIndex, Term, TransformStmt, Unit, VariableIndex,
};
use std::fmt;

/// Provides pretty-printing, including name lookup from arena.
/// Usage:
/// ```
/// clause.pretty(&arena).to_string()
/// ```
pub struct Pretty<'a, T: ?Sized> {
    arena: &'a Arena,
    inner: &'a T,
}

pub trait PrettyPrint {
    fn pretty<'a>(&'a self, arena: &'a Arena) -> Pretty<'a, Self>
    where
        Self: Sized,
    {
        Pretty { arena, inner: self }
    }
}

impl PrettyPrint for VariableIndex {}
impl PrettyPrint for PredicateIndex {}
impl PrettyPrint for FunctionIndex {}
impl<'a> PrettyPrint for Const<'a> {}
impl<'a> PrettyPrint for BaseTerm<'a> {}
impl<'a> PrettyPrint for Atom<'a> {}
impl<'a> PrettyPrint for Term<'a> {}
impl<'a> PrettyPrint for TransformStmt<'a> {}
impl<'a> PrettyPrint for Clause<'a> {}
impl<'a> PrettyPrint for Constraints<'a> {}
impl<'a> PrettyPrint for BoundDecl<'a> {}
impl<'a> PrettyPrint for Decl<'a> {}
impl<'a> PrettyPrint for Unit<'a> {}

impl<'a> fmt::Display for Pretty<'a, VariableIndex> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        if self.inner.0 == 0 {
            write!(f, "_")
        } else {
            match self.arena.lookup_name(self.inner.0) {
                Some(name) => write!(f, "{name}"),
                None => write!(f, "v${}", self.inner.0),
            }
        }
    }
}

impl<'a> fmt::Display for Pretty<'a, PredicateIndex> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self.arena.predicate_name(*self.inner) {
            Some(name) => write!(f, "{name}"),
            None => write!(f, "p${}", self.inner.0),
        }
    }
}

impl<'a> fmt::Display for Pretty<'a, FunctionIndex> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self.arena.function_name(*self.inner) {
            Some(name) => write!(f, "{name}"),
            None => write!(f, "f${}", self.inner.0),
        }
    }
}

impl<'a> fmt::Display for Pretty<'a, Const<'a>> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self.inner {
            Const::Name(n) => match self.arena.lookup_name(*n) {
                Some(name) => write!(f, "{name}"),
                None => write!(f, "n${}", n),
            },
            Const::Bool(b) => write!(f, "{b}"),
            Const::Number(n) => write!(f, "{n}"),
            Const::Float(fl) => write!(f, "{fl}"),
            Const::String(s) => write!(f, "{:?}", s), // Use Debug for quoting
            Const::Bytes(b) => write!(f, "{:?}", b),
            Const::List(l) => {
                write!(f, "[")?;
                for (i, c) in l.iter().enumerate() {
                    if i > 0 {
                        write!(f, ", ")?;
                    }
                    write!(f, "{}", c.pretty(self.arena))?;
                }
                write!(f, "]")
            }
            Const::Map { keys, values } => {
                if keys.is_empty() {
                    write!(f, "fn:map()")
                } else {
                    write!(f, "[")?;
                    for (i, (k, v)) in keys.iter().zip(values.iter()).enumerate() {
                        if i > 0 {
                            write!(f, ", ")?;
                        }
                        write!(f, "{}: {}", k.pretty(self.arena), v.pretty(self.arena))?;
                    }
                    write!(f, "]")
                }
            }
            Const::Struct { fields, values } => {
                write!(f, "{{")?;
                for (i, (field, val)) in fields.iter().zip(values.iter()).enumerate() {
                    if i > 0 {
                        write!(f, ", ")?;
                    }
                    write!(f, "{field}: {}", val.pretty(self.arena))?;
                }
                write!(f, "}}")
            }
        }
    }
}

impl<'a> fmt::Display for Pretty<'a, BaseTerm<'a>> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self.inner {
            BaseTerm::Const(c) => write!(f, "{}", c.pretty(self.arena)),
            BaseTerm::Variable(v) => write!(f, "{}", v.pretty(self.arena)),
            BaseTerm::ApplyFn(fun, args) => {
                write!(f, "{}(", fun.pretty(self.arena))?;
                for (i, arg) in args.iter().enumerate() {
                    if i > 0 {
                        write!(f, ", ")?;
                    }
                    write!(f, "{}", arg.pretty(self.arena))?;
                }
                write!(f, ")")
            }
        }
    }
}

impl<'a> fmt::Display for Pretty<'a, Atom<'a>> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}(", self.inner.sym.pretty(self.arena))?;
        for (i, arg) in self.inner.args.iter().enumerate() {
            if i > 0 {
                write!(f, ", ")?;
            }
            write!(f, "{}", arg.pretty(self.arena))?;
        }
        write!(f, ")")
    }
}

impl<'a> fmt::Display for Pretty<'a, Term<'a>> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self.inner {
            Term::Atom(a) => write!(f, "{}", a.pretty(self.arena)),
            Term::NegAtom(a) => write!(f, "!{}", a.pretty(self.arena)),
            Term::Eq(l, r) => {
                write!(f, "{} = {}", l.pretty(self.arena), r.pretty(self.arena))
            }
            Term::Ineq(l, r) => {
                write!(f, "{} != {}", l.pretty(self.arena), r.pretty(self.arena))
            }
        }
    }
}

impl<'a> fmt::Display for Pretty<'a, TransformStmt<'a>> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        if let Some(var) = self.inner.var {
            write!(f, "let {var} = ")?;
        }
        write!(f, "{}", self.inner.app.pretty(self.arena))
    }
}

impl<'a> fmt::Display for Pretty<'a, Clause<'a>> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.inner.head.pretty(self.arena))?;
        if !self.inner.premises.is_empty() || !self.inner.transform.is_empty() {
            write!(f, " :- ")?;
            let mut first = true;
            for premise in self.inner.premises {
                if !first {
                    write!(f, ", ")?;
                }
                write!(f, "{}", premise.pretty(self.arena))?;
                first = false;
            }
            for transform in self.inner.transform {
                if !first {
                    write!(f, ", ")?;
                }
                write!(f, "{}", transform.pretty(self.arena))?;
                first = false;
            }
        }
        write!(f, ".")
    }
}

impl<'a> fmt::Display for Pretty<'a, Constraints<'a>> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        if !self.inner.consequences.is_empty() {
            write!(f, " |> ")?;
            for (i, c) in self.inner.consequences.iter().enumerate() {
                if i > 0 {
                    write!(f, ", ")?;
                }
                write!(f, "{}", c.pretty(self.arena))?;
            }
        }
        if !self.inner.alternatives.is_empty() {
            for alt in self.inner.alternatives.iter() {
                write!(f, " | ")?;
                for (i, c) in alt.iter().enumerate() {
                    if i > 0 {
                        write!(f, ", ")?;
                    }
                    write!(f, "{}", c.pretty(self.arena))?;
                }
            }
        }
        Ok(())
    }
}

impl<'a> fmt::Display for Pretty<'a, BoundDecl<'a>> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        for (i, b) in self.inner.base_terms.iter().enumerate() {
            if i > 0 {
                write!(f, ", ")?;
            }
            write!(f, "{}", b.pretty(self.arena))?;
        }
        Ok(())
    }
}

impl<'a> fmt::Display for Pretty<'a, Decl<'a>> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.inner.atom.pretty(self.arena))?;
        if !self.inner.descr.is_empty() {
            write!(f, " [")?;
            for (i, d) in self.inner.descr.iter().enumerate() {
                if i > 0 {
                    write!(f, ", ")?;
                }
                write!(f, "{}", d.pretty(self.arena))?;
            }
            write!(f, "]")?;
        }
        if let Some(bounds) = self.inner.bounds {
            if !bounds.is_empty() {
                write!(f, " bound ")?;
                for (i, b) in bounds.iter().enumerate() {
                    if i > 0 {
                        write!(f, " | ")?;
                    }
                    write!(f, "{}", b.pretty(self.arena))?;
                }
            }
        }
        if let Some(constraints) = &self.inner.constraints {
            write!(f, "{}", constraints.pretty(self.arena))?;
        }
        write!(f, ".")
    }
}

impl<'a> fmt::Display for Pretty<'a, Unit<'a>> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        for decl in self.inner.decls {
            writeln!(f, "{}", decl.pretty(self.arena))?;
        }
        for clause in self.inner.clauses {
            writeln!(f, "{}", clause.pretty(self.arena))?;
        }
        Ok(())
    }
}
