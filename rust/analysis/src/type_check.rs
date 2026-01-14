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

use anyhow::{Result, anyhow};
use fxhash::FxHashMap;
use mangle_ir::{Inst, InstId, Ir, NameId};

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Type {
    Any,
    Bool,
    Number,
    Float,
    String,
    Bytes,
    List(Box<Type>),
    #[allow(dead_code)]
    Map(Box<Type>, Box<Type>),
    #[allow(dead_code)]
    Struct, // Simplified
            // TODO: More precise types
}

pub struct TypeChecker<'a> {
    ir: &'a Ir,
    // Predicate name -> Arg types
    signatures: FxHashMap<NameId, Vec<Type>>,
}

impl<'a> TypeChecker<'a> {
    pub fn new(ir: &'a Ir) -> Self {
        Self {
            ir,
            signatures: FxHashMap::default(),
        }
    }

    pub fn check(&mut self) -> Result<()> {
        // Pass 1: Collect signatures from Decls
        for inst in &self.ir.insts {
            if let Inst::Decl { atom, bounds, .. } = inst {
                self.collect_signature(*atom, bounds)?;
            }
        }

        // Pass 2: Check Rules
        for inst in &self.ir.insts {
            if let Inst::Rule {
                head,
                premises,
                transform,
            } = inst
            {
                self.check_rule(*head, premises, transform)?;
            }
        }
        Ok(())
    }

    fn collect_signature(&mut self, atom_id: InstId, bounds: &[InstId]) -> Result<()> {
        let atom = self.ir.get(atom_id);
        if let Inst::Atom { predicate, args } = atom {
            let mut types = Vec::new();
            if !bounds.is_empty() {
                // Bounds map to args?
                // Mangle syntax: bound [/type1, /type2]
                // If there are multiple bound decls, it's intersection or union?
                // Usually one bound decl per rule/pred?
                // Assuming first bound decl defines signature for now.
                if let Some(first_bound_id) = bounds.first()
                    && let Inst::BoundDecl { base_terms } = self.ir.get(*first_bound_id)
                {
                    for term_id in base_terms {
                        types.push(self.resolve_type(*term_id)?);
                    }
                }
            } else {
                // Default to Any
                for _ in args {
                    types.push(Type::Any);
                }
            }
            self.signatures.insert(*predicate, types);
        }
        Ok(())
    }

    fn resolve_type(&self, type_term_id: InstId) -> Result<Type> {
        let inst = self.ir.get(type_term_id);
        match inst {
            Inst::Name(s) => match self.ir.resolve_name(*s) {
                "/string" => Ok(Type::String),
                "/number" => Ok(Type::Number),
                "/float" => Ok(Type::Float),
                "/bool" => Ok(Type::Bool),
                "/bytes" => Ok(Type::Bytes),
                _ => Ok(Type::Any), // Unknown type name
            },
            Inst::ApplyFn { function, args } => {
                match self.ir.resolve_name(*function) {
                    "fn:List" | "fn:list" => {
                        let inner = if let Some(arg) = args.first() {
                            self.resolve_type(*arg)?
                        } else {
                            Type::Any
                        };
                        Ok(Type::List(Box::new(inner)))
                    }
                    // TODO: Map, Struct
                    _ => Ok(Type::Any),
                }
            }
            _ => Ok(Type::Any),
        }
    }

    fn check_rule(&self, head: InstId, premises: &[InstId], _transform: &[InstId]) -> Result<()> {
        let mut var_types: FxHashMap<NameId, Type> = FxHashMap::default();

        // Check premises
        for premise in premises {
            self.check_premise(*premise, &mut var_types)?;
        }

        // Check head
        self.check_atom(head, &mut var_types)?;

        Ok(())
    }

    fn check_premise(
        &self,
        premise: InstId,
        var_types: &mut FxHashMap<NameId, Type>,
    ) -> Result<()> {
        match self.ir.get(premise) {
            Inst::Atom { .. } => self.check_atom(premise, var_types),
            Inst::NegAtom(a) => self.check_atom(*a, var_types),
            Inst::Eq(l, r) => {
                // Unify types of l and r
                let t_l = self.infer_type(*l, var_types)?;
                let t_r = self.infer_type(*r, var_types)?;
                self.unify(t_l, t_r).map(|_| ())
            }
            // ...
            _ => Ok(()),
        }
    }

    fn check_atom(&self, atom_id: InstId, var_types: &mut FxHashMap<NameId, Type>) -> Result<()> {
        if let Inst::Atom { predicate, args } = self.ir.get(atom_id)
            && let Some(sig) = self.signatures.get(predicate)
        {
            if sig.len() != args.len() {
                return Err(anyhow!(
                    "Arity mismatch for {}: expected {}, got {}",
                    self.ir.resolve_name(*predicate),
                    sig.len(),
                    args.len()
                ));
            }
            for (i, arg) in args.iter().enumerate() {
                let expected_type = &sig[i];
                // Update variable type if it was Any/Unknown
                // Or check if it matches
                self.unify_arg(*arg, expected_type.clone(), var_types)?;
            }
        }
        Ok(())
    }

    fn infer_type(&self, term: InstId, var_types: &FxHashMap<NameId, Type>) -> Result<Type> {
        match self.ir.get(term) {
            Inst::Var(name) => Ok(var_types.get(name).cloned().unwrap_or(Type::Any)),
            Inst::Number(_) => Ok(Type::Number),
            Inst::String(_) => Ok(Type::String),
            Inst::Bool(_) => Ok(Type::Bool),
            Inst::Float(_) => Ok(Type::Float),
            Inst::Bytes(_) => Ok(Type::Bytes),
            // ...
            _ => Ok(Type::Any),
        }
    }

    fn unify_arg(
        &self,
        term: InstId,
        expected: Type,
        var_types: &mut FxHashMap<NameId, Type>,
    ) -> Result<()> {
        if let Inst::Var(name) = self.ir.get(term) {
            if let Some(current) = var_types.get(name) {
                let new_type = self.unify(current.clone(), expected)?;
                var_types.insert(*name, new_type);
            } else {
                var_types.insert(*name, expected);
            }
        } else {
            let actual = self.infer_type(term, var_types)?;
            self.unify(actual, expected)?;
        }
        Ok(())
    }

    fn unify(&self, t1: Type, t2: Type) -> Result<Type> {
        match (t1, t2) {
            (Type::Any, t) => Ok(t),
            (t, Type::Any) => Ok(t),
            (t1, t2) if t1 == t2 => Ok(t1),
            (t1, t2) => Err(anyhow!("Type mismatch: {:?} vs {:?}", t1, t2)),
        }
    }
}
