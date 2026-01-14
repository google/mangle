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

//! Intermediate Representation (IR) for Mangle.
//!
//! This IR uses a flat, indexed representation (similar to Carbon's SemIR).
//! Instructions are stored in a vector and referenced by `InstId`.
//!
//! # Relation Representation
//!
//! Relations (predicates) are primarily identified by `NameId`.
//! In the Logical IR, they appear in `Inst::Atom` and `Inst::Decl`.
//!
//! In the Physical IR (`physical::Op`), relations are abstract data sources.
//! Operations like `Scan` and `Insert` refer to relations by name/ID, but the
//! actual storage format (row-oriented, column-oriented, B-Tree, etc.) is
//! determined by the runtime `Host` implementation.
//!
//! # Physical Operations
//!
//! The `physical` module defines the imperative operations for execution:
//!
//! *   **Iterate/Scan**: Provides an iterator over a relation.
//! *   **Filter**: Selects tuples matching a condition.
//! *   **Insert**: Adds derived facts to a relation.
//! *   **Let**: Binds values to variables for projection or calculation.
//!
//! The Planner transforms declarative rules into trees of these operations.

pub mod physical;

use std::collections::HashMap;
use std::hash::Hash;
use std::num::NonZeroU32;

/// Index of an instruction in the IR.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord)]
pub struct InstId(pub NonZeroU32);

impl InstId {
    pub fn new(index: usize) -> Self {
        // We use 0-based index internally, but 1-based NonZeroU32 storage
        // to allow Option<InstId> to be same size as InstId.
        // index 0 -> 1
        InstId(NonZeroU32::new((index + 1) as u32).expect("index overflow"))
    }

    pub fn index(&self) -> usize {
        (self.0.get() - 1) as usize
    }
}

pub trait InternKey: Copy + Eq + Hash {
    fn new(index: usize) -> Self;
    fn index(&self) -> usize;
}

/// Index of a name (identifier) in the IR.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord)]
pub struct NameId(pub NonZeroU32);

impl InternKey for NameId {
    fn new(index: usize) -> Self {
        NameId(NonZeroU32::new((index + 1) as u32).expect("index overflow"))
    }

    fn index(&self) -> usize {
        (self.0.get() - 1) as usize
    }
}

/// Index of a string constant in the IR.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord)]
pub struct StringId(pub NonZeroU32);

impl InternKey for StringId {
    fn new(index: usize) -> Self {
        StringId(NonZeroU32::new((index + 1) as u32).expect("index overflow"))
    }

    fn index(&self) -> usize {
        (self.0.get() - 1) as usize
    }
}

/// A simple interner for strings.
#[derive(Debug)]
pub struct Store<K: InternKey> {
    map: HashMap<String, K>,
    vec: Vec<String>,
}

impl<K: InternKey> Default for Store<K> {
    fn default() -> Self {
        Self {
            map: HashMap::default(),
            vec: Vec::new(),
        }
    }
}

impl<K: InternKey> Store<K> {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn intern(&mut self, name: String) -> K {
        if let Some(&id) = self.map.get(&name) {
            return id;
        }
        let id = K::new(self.vec.len());
        self.vec.push(name.clone());
        self.map.insert(name, id);
        id
    }

    pub fn get(&self, id: K) -> &str {
        &self.vec[id.index()]
    }

    pub fn lookup(&self, name: &str) -> Option<K> {
        self.map.get(name).copied()
    }
}

/// The IR container.
#[derive(Default, Debug)]
pub struct Ir {
    pub insts: Vec<Inst>,
    pub name_store: Store<NameId>,
    pub string_store: Store<StringId>,
}

impl Ir {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn add_inst(&mut self, inst: Inst) -> InstId {
        let id = InstId::new(self.insts.len());
        self.insts.push(inst);
        id
    }

    pub fn get(&self, id: InstId) -> &Inst {
        &self.insts[id.index()]
    }

    pub fn intern_name(&mut self, name: impl Into<String>) -> NameId {
        self.name_store.intern(name.into())
    }

    pub fn resolve_name(&self, id: NameId) -> &str {
        self.name_store.get(id)
    }

    pub fn intern_string(&mut self, s: impl Into<String>) -> StringId {
        self.string_store.intern(s.into())
    }

    pub fn resolve_string(&self, id: StringId) -> &str {
        self.string_store.get(id)
    }
}

/// Instructions in the Mangle IR.
#[derive(Clone, Debug, PartialEq)]
pub enum Inst {
    // --- Constants ---
    Bool(bool),
    Number(i64),
    Float(f64),
    /// A string constant (e.g. "foo").
    String(StringId),
    Bytes(Vec<u8>),
    /// A name constant (e.g. /foo).
    Name(NameId),

    /// A list of values. args point to other Constant-like instructions.
    List(Vec<InstId>),
    /// A map of values. keys and values must have same length.
    Map {
        keys: Vec<InstId>,
        values: Vec<InstId>,
    },
    /// A struct. fields and values must have same length.
    Struct {
        fields: Vec<NameId>,
        values: Vec<InstId>,
    },

    // --- Variables ---
    /// A variable.
    Var(NameId),

    // --- Expressions (BaseTerm) ---
    /// Application of a function.
    ApplyFn {
        function: NameId,
        args: Vec<InstId>,
    },

    // --- Logical Formulas (Term / Atom) ---
    /// An atom (predicate application).
    Atom {
        predicate: NameId,
        args: Vec<InstId>,
    },
    /// Negation of an atom.
    NegAtom(InstId),
    /// Equality constraint (left = right).
    Eq(InstId, InstId),
    /// Inequality constraint (left != right).
    Ineq(InstId, InstId),

    // --- Transforms ---
    /// A transform statement: let var = app.
    Transform {
        var: Option<NameId>,
        app: InstId,
    },

    // --- Structure (Clauses / Decls) ---
    /// A Horn clause (Rule).
    Rule {
        head: InstId,           // Points to Atom
        premises: Vec<InstId>,  // Points to Atom, NegAtom, Eq, Ineq
        transform: Vec<InstId>, // Points to Transform
    },

    /// A Declaration.
    Decl {
        atom: InstId,
        descr: Vec<InstId>,          // Atoms
        bounds: Vec<InstId>,         // BoundDecls
        constraints: Option<InstId>, // Constraints
    },

    /// Bound Declaration.
    BoundDecl {
        base_terms: Vec<InstId>,
    },

    /// Constraints.
    Constraints {
        consequences: Vec<InstId>,      // Atoms
        alternatives: Vec<Vec<InstId>>, // List of List of Atoms
    },
}

#[cfg(test)]
mod compat_test;

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn basic_ir_construction() {
        let mut ir = Ir::new();

        // p(X) :- q(X).

        let x_name = ir.intern_name("X");
        let var_x = ir.add_inst(Inst::Var(x_name));

        let p_name = ir.intern_name("p");
        let atom_head = ir.add_inst(Inst::Atom {
            predicate: p_name,
            args: vec![var_x],
        });

        let q_name = ir.intern_name("q");
        let atom_body = ir.add_inst(Inst::Atom {
            predicate: q_name,
            args: vec![var_x],
        });

        let rule = ir.add_inst(Inst::Rule {
            head: atom_head,
            premises: vec![atom_body],
            transform: vec![],
        });

        assert_eq!(ir.insts.len(), 4);

        if let Inst::Rule { head, .. } = ir.get(rule) {
            assert_eq!(*head, atom_head);
        } else {
            panic!("Expected Rule");
        }
    }

    #[test]
    fn complex_types() {
        let mut ir = Ir::new();

        // Model: .Pair</string, .Struct<opt /a : /string>>

        // Constants / Symbols

        let n_string = ir.intern_name("/string");
        let s_string = ir.add_inst(Inst::Name(n_string));

        let n_a = ir.intern_name("/a");
        let s_a = ir.add_inst(Inst::Name(n_a));

        // Struct field type: opt /a : /string

        // In Mangle Go AST, this is ApplyFn("fn:opt", [/a, /string])

        let fn_opt_name = ir.intern_name("fn:opt");
        let fn_opt = ir.add_inst(Inst::ApplyFn {
            function: fn_opt_name,
            args: vec![s_a, s_string],
        });

        // Struct type: .Struct<...>

        // ApplyFn("fn:Struct", [fn_opt])

        let fn_struct_name = ir.intern_name("fn:Struct");
        let fn_struct = ir.add_inst(Inst::ApplyFn {
            function: fn_struct_name,
            args: vec![fn_opt],
        });

        // Pair type: .Pair</string, Struct...>

        let fn_pair_name = ir.intern_name("fn:Pair");
        let fn_pair = ir.add_inst(Inst::ApplyFn {
            function: fn_pair_name,
            args: vec![s_string, fn_struct],
        });

        // Check structure

        if let Inst::ApplyFn { function, args } = ir.get(fn_pair) {
            assert_eq!(ir.resolve_name(*function), "fn:Pair");

            assert_eq!(args.len(), 2);

            assert_eq!(args[0], s_string);

            assert_eq!(args[1], fn_struct);
        } else {
            panic!("Expected ApplyFn");
        }
    }
}
