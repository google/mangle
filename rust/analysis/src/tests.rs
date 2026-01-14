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

use crate::{LoweringContext, TypeChecker};
use mangle_ast as ast;
use mangle_ir::Inst;

#[test]
fn test_lowering_and_type_check_basic() {
    let arena = ast::Arena::new_with_global_interner();

    // Decl foo(X) bound [/number].
    let foo_sym = arena.predicate_sym("foo", Some(1));
    let var_x = arena.variable("X");
    let atom_foo_x = arena.atom(foo_sym, &[var_x]);

    let num_type = arena.const_(arena.name("/number"));
    let bound_decl = ast::BoundDecl {
        base_terms: arena.alloc_slice_copy(&[num_type]),
    };
    let bound_ref = arena.alloc(bound_decl);

    let decl = ast::Decl {
        atom: atom_foo_x,
        descr: &[],
        bounds: Some(arena.alloc_slice_copy(&[bound_ref])),
        constraints: None,
    };

    // foo(42).
    let const_42 = arena.const_(ast::Const::Number(42));
    let atom_foo_42 = arena.atom(foo_sym, &[const_42]);
    let clause = ast::Clause {
        head: atom_foo_42,
        premises: &[],
        transform: &[],
    };

    let unit = ast::Unit {
        decls: arena.alloc_slice_copy(&[&decl]),
        clauses: arena.alloc_slice_copy(&[&clause]),
    };

    let ctx = LoweringContext::new(&arena);
    let ir = ctx.lower_unit(&unit);

    // Verify IR contains Decl and Rule
    let has_decl = ir.insts.iter().any(|i| matches!(i, Inst::Decl { .. }));
    let has_rule = ir.insts.iter().any(|i| matches!(i, Inst::Rule { .. }));
    assert!(has_decl, "IR missing Decl");
    assert!(has_rule, "IR missing Rule");

    // Type Check
    let mut checker = TypeChecker::new(&ir);
    assert!(
        checker.check().is_ok(),
        "Type check failed for valid program"
    );
}

#[test]
fn test_type_check_arity_ismatch() {
    let arena = ast::Arena::new_with_global_interner();

    // Decl foo(X) bound [/number].
    let foo_sym = arena.predicate_sym("foo", Some(1));
    let var_x = arena.variable("X");
    let atom_foo_x = arena.atom(foo_sym, &[var_x]);

    let num_type = arena.const_(arena.name("/number"));
    let bound_decl = ast::BoundDecl {
        base_terms: arena.alloc_slice_copy(&[num_type]),
    };
    let bound_ref = arena.alloc(bound_decl);

    let decl = ast::Decl {
        atom: atom_foo_x,
        descr: &[],
        bounds: Some(arena.alloc_slice_copy(&[bound_ref])),
        constraints: None,
    };

    // foo(42, 43). -> Arity mismatch (defined as 1, used as 2)
    let const_42 = arena.const_(ast::Const::Number(42));
    let const_43 = arena.const_(ast::Const::Number(43));
    let atom_foo_bad = arena.atom(foo_sym, &[const_42, const_43]); // AST allows this construction
    let clause = ast::Clause {
        head: atom_foo_bad,
        premises: &[],
        transform: &[],
    };

    let unit = ast::Unit {
        decls: arena.alloc_slice_copy(&[&decl]),
        clauses: arena.alloc_slice_copy(&[&clause]),
    };

    let ctx = LoweringContext::new(&arena);
    let ir = ctx.lower_unit(&unit);

    let mut checker = TypeChecker::new(&ir);
    let result = checker.check();
    assert!(result.is_err());
    let err = result.err().unwrap().to_string();
    assert!(err.contains("Arity mismatch"), "Unexpected error: {}", err);
}

#[test]
fn test_type_check_type_mismatch() {
    let arena = ast::Arena::new_with_global_interner();

    // Decl foo(X) bound [/number].
    let foo_sym = arena.predicate_sym("foo", Some(1));
    let var_x = arena.variable("X");
    let atom_foo_x = arena.atom(foo_sym, &[var_x]);

    let num_type = arena.const_(arena.name("/number"));
    let bound_decl = ast::BoundDecl {
        base_terms: arena.alloc_slice_copy(&[num_type]),
    };
    let bound_ref = arena.alloc(bound_decl);

    let decl = ast::Decl {
        atom: atom_foo_x,
        descr: &[],
        bounds: Some(arena.alloc_slice_copy(&[bound_ref])),
        constraints: None,
    };

    // foo("string"). -> Type mismatch
    let const_string = arena.const_(ast::Const::String("hello"));
    let atom_foo_bad = arena.atom(foo_sym, &[const_string]);
    let clause = ast::Clause {
        head: atom_foo_bad,
        premises: &[],
        transform: &[],
    };

    let unit = ast::Unit {
        decls: arena.alloc_slice_copy(&[&decl]),
        clauses: arena.alloc_slice_copy(&[&clause]),
    };

    let ctx = LoweringContext::new(&arena);
    let ir = ctx.lower_unit(&unit);

    let mut checker = TypeChecker::new(&ir);
    let result = checker.check();
    assert!(result.is_err());
    let err = result.err().unwrap().to_string();
    assert!(err.contains("Type mismatch"), "Unexpected error: {}", err);
}

#[test]
fn test_planner_basic() {
    let arena = ast::Arena::new_with_global_interner();
    // Rule: p(X) :- q(X).
    let p = arena.predicate_sym("p", Some(1));
    let q = arena.predicate_sym("q", Some(1));
    let x = arena.variable("X");

    let head = arena.atom(p, &[x]);
    let premise = arena.atom(q, &[x]);

    let clause = ast::Clause {
        head,
        premises: arena.alloc_slice_copy(&[arena.alloc(ast::Term::Atom(premise))]),
        transform: &[],
    };

    let unit = ast::Unit {
        decls: &[],
        clauses: arena.alloc_slice_copy(&[&clause]),
    };

    let ctx = LoweringContext::new(&arena);
    let mut ir = ctx.lower_unit(&unit);

    // Find rule
    let rule_id = ir
        .insts
        .iter()
        .position(|i| matches!(i, Inst::Rule { .. }))
        .unwrap();
    let rule_inst = mangle_ir::InstId::new(rule_id);

    use crate::Planner;
    let planner = Planner::new(&mut ir);
    let op = planner.plan_rule(rule_inst).unwrap();

    // Check if Op is Iterate -> Insert
    use mangle_ir::physical::Op;
    if let Op::Iterate { body, .. } = op {
        if let Op::Insert { relation, .. } = *body {
            assert_eq!(ir.resolve_name(relation), "p");
        } else {
            panic!("Expected inner Insert");
        }
    } else {
        panic!("Expected outer Iterate");
    }
}
