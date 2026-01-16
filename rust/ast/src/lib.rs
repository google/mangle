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

pub mod pretty;
pub use pretty::PrettyPrint;

use bumpalo::Bump;
use fxhash::FxHashMap;
use std::cell::RefCell;
use std::mem;
use std::num::NonZeroUsize;
use std::sync::{Arc, Mutex};

const INTERNER_DEFAULT_CAPACITY: NonZeroUsize = NonZeroUsize::new(4096).unwrap();

/// A simple way to intern strings and refer to them using u32 indices.
/// This is following the blog post here:
/// <https://matklad.github.io/2020/03/22/fast-simple-rust-interner.html>
pub struct Interner {
    map: FxHashMap<&'static str, u32>,
    vec: Vec<&'static str>,
    buf: String,
    full: Vec<String>,
}

impl Interner {
    fn new_global_interner() -> Arc<Mutex<Interner>> {
        Arc::new(Mutex::new(Interner::with_capacity(
            INTERNER_DEFAULT_CAPACITY.into(),
        )))
    }

    pub fn with_capacity(cap: usize) -> Interner {
        let cap = cap.next_power_of_two();
        let mut interner = Interner {
            map: FxHashMap::default(),
            vec: Vec::new(),
            buf: String::with_capacity(cap),
            full: Vec::new(),
        };
        interner.intern("_");
        interner
    }

    /// Interns the argument and returns a unique index.
    pub fn intern(&mut self, name: &str) -> u32 {
        if let Some(&id) = self.map.get(name) {
            return id;
        }
        let name = unsafe { self.alloc(name) };
        let id = self.map.len() as u32;
        self.map.insert(name, id);
        self.vec.push(name);
        debug_assert!(self.lookup(id).expect("expected to find name") == name);
        debug_assert!(self.intern(name) == id);
        id
    }

    pub fn lookup_name_index(&self, name: &str) -> Option<u32> {
        self.map.get(name).copied()
    }

    fn lookup(&self, id: u32) -> Option<&'static str> {
        self.vec.get(id as usize).copied()
    }

    /// Copies a str into our internal buffer and returns a reference.
    ///
    /// # Safety
    ///
    /// The reference is only valid for the lifetime of this object.
    unsafe fn alloc(&mut self, name: &str) -> &'static str {
        let cap = self.buf.capacity();
        if cap < self.buf.len() + name.len() {
            let new_cap = (cap.max(name.len()) + 1).next_power_of_two();
            let new_buf = String::with_capacity(new_cap);
            let old_buf = mem::replace(&mut self.buf, new_buf);
            self.full.push(old_buf);
        }
        let interned = {
            let start = self.buf.len();
            self.buf.push_str(name);
            &self.buf[start..]
        };
        // SAFETY: This is reference to our internal buffer, which will
        // be valid as long as this object is valid.
        unsafe { &*(interned as *const str) }
    }
}

pub struct Arena {
    pub(crate) bump: Bump,
    pub(crate) interner: Arc<Mutex<Interner>>,
    pub(crate) predicate_syms: RefCell<Vec<PredicateSym>>,
    pub(crate) function_syms: RefCell<Vec<FunctionSym>>,
}

impl<'arena> Arena {
    pub fn new(interner: Arc<Mutex<Interner>>) -> Self {
        Self {
            bump: Bump::new(),
            interner,
            predicate_syms: RefCell::new(Vec::new()),
            function_syms: RefCell::new(Vec::new()),
        }
    }

    pub fn new_with_global_interner() -> Self {
        Self::new(Interner::new_global_interner())
    }

    pub fn intern(&'arena self, name: &str) -> u32 {
        self.interner.lock().unwrap().intern(name)
    }

    // Returns index for name, if it exists.
    pub fn lookup_opt(&'arena self, name: &str) -> Option<u32> {
        self.interner.lock().unwrap().map.get(name).copied()
    }

    pub fn name(&'arena self, name: &str) -> Const<'arena> {
        Const::Name(self.intern(name))
    }

    pub fn variable(&'arena self, name: &str) -> &'arena BaseTerm<'arena> {
        self.alloc(BaseTerm::Variable(self.variable_sym(name)))
    }

    pub fn const_(&'arena self, c: Const<'arena>) -> &'arena BaseTerm<'arena> {
        self.alloc(BaseTerm::Const(c))
    }

    pub fn atom(
        &'arena self,
        p: PredicateIndex,
        args: &[&'arena BaseTerm<'arena>],
    ) -> &'arena Atom<'arena> {
        self.alloc(Atom {
            sym: p,
            args: self.alloc_slice_copy(args),
        })
    }

    pub fn apply_fn(
        &'arena self,
        fun: FunctionIndex,
        args: &[&'arena BaseTerm<'arena>],
    ) -> &'arena BaseTerm<'arena> {
        //let fun = &self.function_syms.borrow()[fun.0];
        //let name = self.lookup_name(fun.name);
        let args = self.alloc_slice_copy(args);
        self.alloc(BaseTerm::ApplyFn(fun, args))
    }

    pub fn alloc<T>(&self, x: T) -> &mut T {
        self.bump.alloc(x)
    }

    pub fn alloc_slice_copy<T: Copy>(&self, x: &[T]) -> &[T] {
        self.bump.alloc_slice_copy(x)
    }

    pub fn alloc_str(&'arena self, s: &str) -> &'arena str {
        self.bump.alloc_str(s)
    }

    pub fn new_query(&'arena self, p: PredicateIndex) -> Atom<'arena> {
        let arity = self.predicate_syms.borrow()[p.0].arity;
        let args: Vec<_> = match arity {
            Some(arity) => (0..arity).map(|_i| &ANY_VAR_TERM).collect(),
            None => Vec::new(),
        };

        let args = self.alloc_slice_copy(&args);
        Atom { sym: p, args }
    }

    /// Given a name index, returns the name.
    pub fn lookup_name(&self, name_index: u32) -> Option<&'static str> {
        self.interner.lock().unwrap().lookup(name_index)
    }

    /// Given a name, returns the index of the name if it exists in the interner.
    pub fn lookup_name_index(&self, name: &str) -> Option<u32> {
        self.interner.lock().unwrap().lookup_name_index(name)
    }

    /// Given predicate index, returns name of predicate symbol.
    pub fn predicate_name(&self, predicate_index: PredicateIndex) -> Option<&'static str> {
        let syms = self.predicate_syms.borrow();
        let i = predicate_index.0;
        if i >= syms.len() {
            return None;
        }
        let n = syms[i].name;
        self.interner.lock().unwrap().lookup(n)
    }

    /// Given predicate index, returns arity of predicate symbol.
    pub fn predicate_arity(&self, predicate_index: PredicateIndex) -> Option<u8> {
        self.predicate_syms
            .borrow()
            .get(predicate_index.0)
            .and_then(|s| s.arity)
    }

    /// Given function index, returns name of function symbol.
    pub fn function_name(&self, function_index: FunctionIndex) -> Option<&'static str> {
        let syms = self.function_syms.borrow();
        let i = function_index.0;
        if i >= syms.len() {
            return None;
        }
        let n = syms[i].name;
        self.interner.lock().unwrap().lookup(n)
    }

    /// Returns index for this predicate symbol.
    pub fn lookup_predicate_sym(&'arena self, predicate_name: u32) -> Option<PredicateIndex> {
        for (index, p) in self.predicate_syms.borrow().iter().enumerate() {
            if p.name == predicate_name {
                return Some(PredicateIndex(index));
            }
        }
        None
    }

    /// Constructs a new variable symbol.
    pub fn variable_sym(&'arena self, name: &str) -> VariableIndex {
        let n = self.interner.lock().unwrap().intern(name);
        VariableIndex(n)
    }

    // Looks up a function sym, copies if it doesn't exist.
    pub fn function_sym(&'arena self, name: &str, arity: Option<u8>) -> FunctionIndex {
        let n = self.interner.lock().unwrap().intern(name);
        let f = FunctionSym { name: n, arity };
        for (index, f) in self.function_syms.borrow().iter().enumerate() {
            if f.name == n {
                return FunctionIndex(index);
            }
        }

        self.function_syms.borrow_mut().push(f);
        FunctionIndex(self.function_syms.borrow().len() - 1)
    }

    // Looks up or creates a predicate_sym, copies if it doesn't exist.
    pub fn predicate_sym(&'arena self, name: &str, arity: Option<u8>) -> PredicateIndex {
        let n = self.interner.lock().unwrap().intern(name);
        let p = PredicateSym { name: n, arity };
        for (index, p) in self.predicate_syms.borrow().iter().enumerate() {
            if p.name == n {
                return PredicateIndex(index);
            }
        }

        self.predicate_syms.borrow_mut().push(p);
        PredicateIndex(self.predicate_syms.borrow().len() - 1)
    }

    pub fn copy_function_sym<'src>(
        &'arena self,
        src: &'src Arena,
        f: FunctionIndex,
    ) -> FunctionIndex {
        let function_sym = &src.function_syms.borrow()[f.0];
        let name = src
            .lookup_name(function_sym.name)
            .expect("expected to find name");
        self.function_sym(name, function_sym.arity)
    }

    pub fn copy_predicate_sym<'src>(
        &'arena self,
        src: &'src Arena,
        p: PredicateIndex, //predicate_sym: &'src PredicateSym,
    ) -> PredicateIndex {
        let predicate_sym = &src.predicate_syms.borrow()[p.0];
        let name = src
            .lookup_name(predicate_sym.name)
            .expect("expected to find name");
        self.predicate_sym(name, predicate_sym.arity)
    }

    // Copies BaseTerm from another [`Arena`].
    pub fn copy_atom<'src>(
        &'arena self,
        src: &'src Arena,
        atom: &'src Atom<'src>,
    ) -> &'arena Atom<'arena> {
        let args: Vec<_> = atom
            .args
            .iter()
            .map(|arg| self.copy_base_term(src, arg))
            .collect();
        let args = self.alloc_slice_copy(&args);
        // TODO: have to look up predicate syms from !source!
        self.alloc(Atom {
            sym: self.copy_predicate_sym(src, atom.sym),
            args,
        })
    }

    // Copies BaseTerm to another Arena
    pub fn copy_base_term<'src>(
        &'arena self,
        src: &'src Arena,
        b: &'src BaseTerm<'src>,
    ) -> &'arena BaseTerm<'arena> {
        match b {
            BaseTerm::Const(c) =>
            // Should it be a reference?
            {
                self.alloc(BaseTerm::Const(*self.copy_const(src, c)))
            }
            BaseTerm::Variable(v) => {
                let name = src
                    .interner
                    .lock()
                    .unwrap()
                    .lookup(v.0)
                    .expect("expected to find name")
                    .to_string();
                let v = self.variable_sym(&name);
                self.alloc(BaseTerm::Variable(v))
            }
            BaseTerm::ApplyFn(fun, args) => {
                let fun = self.copy_function_sym(src, *fun);
                //let fun = FunctionSym { name: self.alloc_str(fun.name), arity: fun.arity };
                let args: Vec<_> = args.iter().map(|a| self.copy_base_term(src, a)).collect();
                let args = self.alloc_slice_copy(&args);
                self.alloc(BaseTerm::ApplyFn(fun, args))
            }
        }
    }

    // Copies Const to another Arena
    pub fn copy_const<'src>(
        &'arena self,
        src: &'src Arena,
        c: &'src Const<'src>,
    ) -> &'arena Const<'arena> {
        match c {
            Const::Name(name) => {
                let name = src
                    .interner
                    .lock()
                    .unwrap()
                    .lookup(*name)
                    .expect("expected to find name");
                let name = self.interner.lock().unwrap().intern(name);
                self.alloc(Const::Name(name))
            }
            Const::Bool(b) => self.alloc(Const::Bool(*b)),
            Const::Number(n) => self.alloc(Const::Number(*n)),
            Const::Float(f) => self.alloc(Const::Float(*f)),
            Const::String(s) => {
                let s = self.alloc_str(s);
                self.alloc(Const::String(s))
            }
            Const::Bytes(b) => {
                let b = self.alloc_slice_copy(b);
                self.alloc(Const::Bytes(b))
            }
            Const::List(cs) => {
                let cs: Vec<_> = cs.iter().map(|c| self.copy_const(src, c)).collect();
                let cs = self.alloc_slice_copy(&cs);
                self.alloc(Const::List(cs))
            }
            Const::Map { keys, values } => {
                let keys: Vec<_> = keys.iter().map(|c| self.copy_const(src, c)).collect();
                let keys = self.alloc_slice_copy(&keys);

                let values: Vec<_> = values.iter().map(|c| self.copy_const(src, c)).collect();
                let values = self.alloc_slice_copy(&values);

                self.alloc(Const::Map { keys, values })
            }
            Const::Struct { fields, values } => {
                let fields: Vec<_> = fields.iter().map(|s| self.alloc_str(s)).collect();
                let fields = self.alloc_slice_copy(&fields);

                let values: Vec<_> = values.iter().map(|c| self.copy_const(src, c)).collect();
                let values = self.alloc_slice_copy(&values);

                self.alloc(Const::Struct { fields, values })
            }
        }
    }

    pub fn copy_transform<'src>(
        &'arena self,
        src: &'src Arena,
        stmt: &'src TransformStmt<'src>,
    ) -> &'arena TransformStmt<'arena> {
        let TransformStmt { var, app } = stmt;
        let var = var.map(|s| self.alloc_str(s));
        let app = self.copy_base_term(src, app);
        self.alloc(TransformStmt { var, app })
    }

    pub fn copy_clause<'src>(
        &'arena self,
        src: &'src Arena,
        src_clause: &'src Clause<'src>,
    ) -> &'arena Clause<'arena> {
        let Clause {
            head,
            premises,
            transform,
        } = src_clause;
        let premises: Vec<_> = premises.iter().map(|x| self.copy_term(src, x)).collect();
        let transform: Vec<_> = transform
            .iter()
            .map(|x| self.copy_transform(src, x))
            .collect();
        self.alloc(Clause {
            head: self.copy_atom(src, head),
            premises: self.alloc_slice_copy(&premises),
            transform: self.alloc_slice_copy(&transform),
        })
    }

    fn copy_term<'src>(
        &'arena self,
        src: &'src Arena,
        term: &'src Term<'src>,
    ) -> &'arena Term<'arena> {
        match term {
            Term::Atom(atom) => {
                let atom = self.copy_atom(src, atom);
                self.alloc(Term::Atom(atom))
            }
            Term::NegAtom(atom) => {
                let atom = self.copy_atom(src, atom);
                self.alloc(Term::NegAtom(atom))
            }
            Term::Eq(left, right) => {
                let left = self.copy_base_term(src, left);
                let right = self.copy_base_term(src, right);
                self.alloc(Term::Eq(left, right))
            }
            Term::Ineq(left, right) => {
                let left = self.copy_base_term(src, left);
                let right = self.copy_base_term(src, right);
                self.alloc(Term::Ineq(left, right))
            }
        }
    }
}

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

impl std::fmt::Display for Term<'_> {
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
    pub fn apply_subst(
        &'a self,
        arena: &'a Arena,
        subst: &FxHashMap<u32, &'a BaseTerm<'a>>,
    ) -> &'a Term<'a> {
        &*arena.alloc(match self {
            Term::Atom(atom) => Term::Atom(atom.apply_subst(arena, subst)),
            Term::NegAtom(atom) => Term::NegAtom(atom.apply_subst(arena, subst)),
            Term::Eq(left, right) => Term::Eq(
                left.apply_subst(arena, subst),
                right.apply_subst(arena, subst),
            ),
            Term::Ineq(left, right) => Term::Ineq(
                left.apply_subst(arena, subst),
                right.apply_subst(arena, subst),
            ),
        })
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum BaseTerm<'a> {
    Const(Const<'a>),
    Variable(VariableIndex),
    ApplyFn(FunctionIndex, &'a [&'a BaseTerm<'a>]),
}

impl<'arena> BaseTerm<'arena> {
    pub fn apply_subst(
        &'arena self,
        arena: &'arena Arena,
        subst: &FxHashMap<u32, &'arena BaseTerm<'arena>>,
    ) -> &'arena BaseTerm<'arena> {
        match self {
            BaseTerm::Const(_) => self,
            BaseTerm::Variable(v) => subst.get(&v.0).unwrap_or(&self),
            BaseTerm::ApplyFn(fun, args) => {
                let args: Vec<&'arena BaseTerm<'arena>> = args
                    .iter()
                    .map(|arg| arg.apply_subst(arena, subst))
                    .collect();
                let args = arena.alloc_slice_copy(&args);
                arena.alloc(BaseTerm::ApplyFn(*fun, args))
            }
        }
    }
}

impl std::fmt::Display for BaseTerm<'_> {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            BaseTerm::Const(c) => write!(f, "{c}"),
            BaseTerm::Variable(v) => write!(f, "{v}"),
            BaseTerm::ApplyFn(fun, args) => {
                write!(
                    f,
                    "{fun}({})",
                    args.iter()
                        .map(|x| x.to_string())
                        .collect::<Vec<_>>()
                        .join(",")
                )
            }
        }
    }
}
#[derive(Debug, Clone, Copy, PartialEq)]
pub enum Const<'a> {
    Name(u32),
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

impl Eq for Const<'_> {}

impl std::fmt::Display for Const<'_> {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match *self {
            Const::Name(v) => write!(f, "n${v}"),
            Const::Bool(v) => write!(f, "{v}"),
            Const::Number(v) => write!(f, "{v}"),
            Const::Float(v) => write!(f, "{v}"),
            Const::String(v) => write!(f, "{v}"),
            Const::Bytes(v) => write!(f, "{v:?}"),
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
            Const::Map { keys, values } => {
                if keys.is_empty() {
                    write!(f, "fn:map()")
                } else {
                    write!(f, "[")?;
                    for (i, (k, v)) in keys.iter().zip(values.iter()).enumerate() {
                        if i > 0 {
                            write!(f, ", ")?;
                        }
                        write!(f, "{k}: {v}")?;
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
                    write!(f, "{field}: {val}")?;
                }
                write!(f, "}}")
            }
        }
    }
}

#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub struct PredicateSym {
    pub name: u32,
    pub arity: Option<u8>,
}

#[derive(Debug, Copy, Clone, PartialEq, Eq, Hash, PartialOrd, Ord)]
pub struct PredicateIndex(usize);

impl std::fmt::Display for PredicateIndex {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "p${}", self.0)
    }
}

#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub struct FunctionSym {
    pub name: u32,
    pub arity: Option<u8>,
}

#[derive(Debug, Copy, Clone, PartialEq, Eq, Hash)]
pub struct FunctionIndex(usize);

impl std::fmt::Display for FunctionIndex {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "f${}", self.0)
    }
}

#[derive(Debug, Copy, Clone, PartialEq, Eq, Hash)]
pub struct VariableIndex(pub u32);

impl std::fmt::Display for VariableIndex {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        if self.0 == 0 {
            write!(f, "_")
        } else {
            write!(f, "v${}", self.0)
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Atom<'a> {
    pub sym: PredicateIndex,
    pub args: &'a [&'a BaseTerm<'a>],
}

impl<'a> Atom<'a> {
    // A fact matches a query if there is a substitution s.t. subst(query) = fact.
    // We assume that self is a fact and query_args are the arguments of a query,
    // i.e. only variables and constants.
    pub fn matches(&'a self, query_args: &[&BaseTerm]) -> bool {
        for (fact_arg, query_arg) in self.args.iter().zip(query_args.iter()) {
            if let BaseTerm::Const(_) = query_arg
                && fact_arg != query_arg
            {
                return false;
            }
        }
        true
    }

    pub fn apply_subst(
        &'a self,
        arena: &'a Arena,
        subst: &FxHashMap<u32, &'a BaseTerm<'a>>,
    ) -> &'a Atom<'a> {
        let args: Vec<&'a BaseTerm<'a>> = self
            .args
            .iter()
            .map(|arg| arg.apply_subst(arena, subst))
            .collect();
        arena.atom(self.sym, &args)
    }
}

impl std::fmt::Display for Atom<'_> {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}(", self.sym)?;
        for (i, arg) in self.args.iter().enumerate() {
            if i > 0 {
                write!(f, ", ")?;
            }
            write!(f, "{arg}")?;
        }
        write!(f, ")")
    }
}

static ANY_VAR_TERM: BaseTerm = BaseTerm::Variable(VariableIndex(0));

#[cfg(test)]
mod tests {
    use super::*;
    use googletest::prelude::*;

    #[test]
    fn copying_atom_works() {
        let arena = Arena::new_with_global_interner();
        let foo = arena.const_(arena.name("/foo"));
        let bar = arena.predicate_sym("bar", Some(1));
        let head = arena.atom(bar, &[foo]);
        assert_that!(head.to_string(), eq("p$0(n$1)"));
    }

    #[test]
    fn atom_display_works() {
        let arena = Arena::new_with_global_interner();
        let bar = arena.const_(arena.name("/bar"));
        let sym = arena.predicate_sym("foo", Some(1));
        let atom = Atom { sym, args: &[bar] };
        assert_that!(atom, displays_as(eq("p$0(n$1)")));

        let tests = vec![
            (Term::Atom(&atom), "p$0(n$1)"),
            (Term::NegAtom(&atom), "!p$0(n$1)"),
            (Term::Eq(bar, bar), "n$1 = n$1"),
            (Term::Ineq(bar, bar), "n$1 != n$1"),
        ];
        for (term, s) in tests {
            assert_that!(term, displays_as(eq(s)));
        }
    }

    #[test]
    fn new_query_works() {
        let arena = Arena::new_with_global_interner();

        let pred = arena.predicate_sym("foo", Some(1));
        let query = arena.new_query(pred);
        assert_that!(query, displays_as(eq("p$0(_)")));

        let pred = arena.predicate_sym("bar", Some(2));
        let query = arena.new_query(pred);
        assert_that!(query, displays_as(eq("p$1(_, _)")));

        let pred = arena.predicate_sym("frob", None);
        let query = arena.new_query(pred);
        assert_that!(query, displays_as(eq("p$2()")));
    }

    #[test]
    fn subst_works() {
        let arena = Arena::new_with_global_interner();
        let atom = arena.atom(arena.predicate_sym("foo", Some(1)), &[arena.variable("x")]);

        let mut subst = FxHashMap::default();
        subst.insert(arena.variable_sym("x").0, arena.const_(arena.name("/bar")));

        let subst_atom = atom.apply_subst(&arena, &subst);
        assert_that!(arena.name("/bar"), displays_as(eq("n$3")));
        assert_that!(subst_atom, displays_as(eq("p$0(n$3)")));
    }

    #[test]
    fn do_intern_beyond_initial_capacity() {
        let arena = Arena::new_with_global_interner();

        let p = arena.predicate_sym("/foo", Some(1));
        let mut name = "".to_string();
        for _ in 0..INTERNER_DEFAULT_CAPACITY.into() {
            name += "a";
        }
        arena.interner.lock().unwrap().intern(&name);
        assert_that!(arena.predicate_name(p), eq(Some("/foo")));
    }

    #[test]
    fn pretty_print_works() {
        let arena = Arena::new_with_global_interner();
        let foo = arena.const_(arena.name("/foo"));
        let bar_pred = arena.predicate_sym("bar", Some(1));
        let x_var = arena.variable("X");
        let head = arena.atom(bar_pred, &[x_var]); // bar(X)

        let premise = Term::Eq(x_var, foo);
        let premise_ref = arena.alloc(premise);

        let clause = Clause {
            head,
            premises: arena.alloc_slice_copy(&[premise_ref]),
            transform: &[],
        };

        assert_that!(clause.pretty(&arena).to_string(), eq("bar(X) :- X = /foo."));

        let fun = arena.function_sym("f", Some(1));
        let app = arena.apply_fn(fun, &[x_var]);
        assert_that!(app.pretty(&arena).to_string(), eq("f(X)"));
    }
}
