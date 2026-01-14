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

use fxhash::FxHashMap;
use mangle_ast as ast;
use mangle_ir::{Inst, InstId, Ir};

pub struct LoweringContext<'a> {
    arena: &'a ast::Arena,
    ir: Ir,
    // Scope-specific maps
    vars: FxHashMap<ast::VariableIndex, InstId>,
}

impl<'a> LoweringContext<'a> {
    pub fn new(arena: &'a ast::Arena) -> Self {
        Self {
            arena,
            ir: Ir::new(),
            vars: FxHashMap::default(),
        }
    }

    pub fn lower_unit(mut self, unit: &ast::Unit) -> Ir {
        for decl in unit.decls {
            self.lower_decl(decl);
        }
        for clause in unit.clauses {
            self.lower_clause(clause);
        }
        self.ir
    }

    fn lower_decl(&mut self, decl: &ast::Decl) -> InstId {
        self.vars.clear();

        let atom = self.lower_atom(decl.atom);
        let descr: Vec<InstId> = decl.descr.iter().map(|a| self.lower_atom(a)).collect();
        let bounds: Vec<InstId> = if let Some(bs) = decl.bounds {
            bs.iter().map(|b| self.lower_bound_decl(b)).collect()
        } else {
            Vec::new()
        };
        let constraints = decl.constraints.map(|c| self.lower_constraints(c));

        self.ir.add_inst(Inst::Decl {
            atom,
            descr,
            bounds,
            constraints,
        })
    }

    fn lower_clause(&mut self, clause: &ast::Clause) -> InstId {
        self.vars.clear();

        let head = self.lower_atom(clause.head);
        let premises: Vec<InstId> = clause.premises.iter().map(|t| self.lower_term(t)).collect();
        let transform: Vec<InstId> = clause
            .transform
            .iter()
            .map(|t| self.lower_transform(t))
            .collect();

        self.ir.add_inst(Inst::Rule {
            head,
            premises,
            transform,
        })
    }

    fn lower_atom(&mut self, atom: &ast::Atom) -> InstId {
        let predicate_name = self
            .arena
            .predicate_name(atom.sym)
            .unwrap_or("unknown_pred")
            .to_string();
        let predicate = self.ir.intern_name(predicate_name);
        let args: Vec<InstId> = atom
            .args
            .iter()
            .map(|arg| self.lower_base_term(arg))
            .collect();
        self.ir.add_inst(Inst::Atom { predicate, args })
    }

    fn lower_term(&mut self, term: &ast::Term) -> InstId {
        match term {
            ast::Term::Atom(a) => self.lower_atom(a),
            ast::Term::NegAtom(a) => {
                let atom = self.lower_atom(a);
                self.ir.add_inst(Inst::NegAtom(atom))
            }
            ast::Term::Eq(l, r) => {
                let left = self.lower_base_term(l);
                let right = self.lower_base_term(r);
                self.ir.add_inst(Inst::Eq(left, right))
            }
            ast::Term::Ineq(l, r) => {
                let left = self.lower_base_term(l);
                let right = self.lower_base_term(r);
                self.ir.add_inst(Inst::Ineq(left, right))
            }
        }
    }

    fn lower_base_term(&mut self, term: &ast::BaseTerm) -> InstId {
        match term {
            ast::BaseTerm::Const(c) => self.lower_const(c),
            ast::BaseTerm::Variable(v) => {
                if let Some(id) = self.vars.get(v) {
                    *id
                } else {
                    let name_str = if v.0 == 0 {
                        "_".to_string()
                    } else {
                        self.arena
                            .lookup_name(v.0)
                            .unwrap_or("unknown_var")
                            .to_string()
                    };
                    let name = self.ir.intern_name(name_str);
                    let id = self.ir.add_inst(Inst::Var(name));
                    // Don't cache wildcard?
                    if v.0 != 0 {
                        self.vars.insert(*v, id);
                    }
                    id
                }
            }
            ast::BaseTerm::ApplyFn(f, args) => {
                let function_str = self
                    .arena
                    .function_name(*f)
                    .unwrap_or("unknown_fn")
                    .to_string();
                let function = self.ir.intern_name(function_str);
                let args = args.iter().map(|a| self.lower_base_term(a)).collect();
                self.ir.add_inst(Inst::ApplyFn { function, args })
            }
        }
    }

    fn lower_const(&mut self, c: &ast::Const) -> InstId {
        match c {
            ast::Const::Name(n) => {
                let name_str = self
                    .arena
                    .lookup_name(*n)
                    .unwrap_or("unknown_name")
                    .to_string();
                let name = self.ir.intern_name(name_str);
                self.ir.add_inst(Inst::Name(name))
            }
            ast::Const::Bool(b) => self.ir.add_inst(Inst::Bool(*b)),
            ast::Const::Number(n) => self.ir.add_inst(Inst::Number(*n)),
            ast::Const::Float(f) => self.ir.add_inst(Inst::Float(*f)),
            ast::Const::String(s) => {
                let id = self.ir.intern_string(*s);
                self.ir.add_inst(Inst::String(id))
            }
            ast::Const::Bytes(b) => self.ir.add_inst(Inst::Bytes(b.to_vec())),
            ast::Const::List(l) => {
                let args = l.iter().map(|c| self.lower_const(c)).collect();
                self.ir.add_inst(Inst::List(args))
            }
            ast::Const::Map { keys, values } => {
                let keys = keys.iter().map(|c| self.lower_const(c)).collect();
                let values = values.iter().map(|c| self.lower_const(c)).collect();
                self.ir.add_inst(Inst::Map { keys, values })
            }
            ast::Const::Struct { fields, values } => {
                let fields = fields
                    .iter()
                    .map(|s| self.ir.intern_name(s.to_string()))
                    .collect();
                let values = values.iter().map(|c| self.lower_const(c)).collect();
                self.ir.add_inst(Inst::Struct { fields, values })
            }
        }
    }

    fn lower_transform(&mut self, t: &ast::TransformStmt) -> InstId {
        let var = t.var.map(|s| self.ir.intern_name(s.to_string()));
        let app = self.lower_base_term(t.app);
        self.ir.add_inst(Inst::Transform { var, app })
    }

    fn lower_bound_decl(&mut self, b: &ast::BoundDecl) -> InstId {
        let base_terms = b
            .base_terms
            .iter()
            .map(|t| self.lower_base_term(t))
            .collect();
        self.ir.add_inst(Inst::BoundDecl { base_terms })
    }

    fn lower_constraints(&mut self, c: &ast::Constraints) -> InstId {
        let consequences = c.consequences.iter().map(|a| self.lower_atom(a)).collect();
        let alternatives = c
            .alternatives
            .iter()
            .map(|alt| alt.iter().map(|a| self.lower_atom(a)).collect())
            .collect();
        self.ir.add_inst(Inst::Constraints {
            consequences,
            alternatives,
        })
    }
}
