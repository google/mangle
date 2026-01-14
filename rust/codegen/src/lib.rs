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

use std::collections::HashMap;

use mangle_analysis::{Planner, StratifiedProgram};
use mangle_ir::physical::{CmpOp, Condition, Constant, DataSource, Expr, Op, Operand};
use mangle_ir::{Inst, InstId, Ir, NameId};
use wasm_encoder::{
    CodeSection, EntityType, ExportKind, ExportSection, Function, FunctionSection, ImportSection,
    Instruction, MemorySection, Module, TypeSection, ValType,
};

/// Backend strategy for implementing physical operations.
pub trait Backend {
    /// Emits code to start a scan. Pushes iter_id (i32) to stack.
    fn emit_scan_start(&self, func: &mut Function, rel_name: &str);

    /// Emits code to start a delta scan (new facts only).
    fn emit_scan_delta_start(&self, func: &mut Function, rel_name: &str);

    /// Emits code to start an indexed scan.
    /// Expects key (i64) on stack. Pushes iter_id (i32) to stack.
    fn emit_scan_index_start(&self, func: &mut Function, rel_name: &str, col_idx: u32);

    /// Emits code to start an aggregation scan.
    /// Pushes iter_id (i32) to stack.
    fn emit_scan_aggregate_start(&self, func: &mut Function, rel_name: &str, ptr: i32, len: i32);

    /// Emits code to get next tuple. Takes iter_id from stack (or local).
    /// Pushes tuple_ptr (i32) to stack.
    fn emit_scan_next(&self, func: &mut Function, iter_local: u32);

    /// Emits code to get column value from tuple.
    /// Pushes value (i64) to stack.
    fn emit_get_col(&self, func: &mut Function, tuple_local: u32, col_idx: u32);

    /// Emits code to prepare insertion (e.g. push relation ID).
    fn emit_insert_start(&self, func: &mut Function, rel_name: &str);

    /// Emits code to perform insertion (e.g. call host function).
    fn emit_insert_end(&self, func: &mut Function);

    /// Emits code to merge deltas. Returns 1 if changes, 0 if not (on stack).
    fn emit_merge_deltas(&self, func: &mut Function);

    /// Emits code to log a value from WASM.
    fn emit_debuglog(&self, func: &mut Function, val_local: u32);
}

pub struct WasmImportsBackend;

impl Backend for WasmImportsBackend {
    fn emit_scan_start(&self, func: &mut Function, rel_name: &str) {
        let mut hash: u32 = 5381;
        for c in rel_name.bytes() {
            hash = ((hash << 5).wrapping_add(hash)).wrapping_add(c as u32);
        }
        func.instruction(&Instruction::I32Const(hash as i32));
        func.instruction(&Instruction::Call(0));
    }

    fn emit_scan_delta_start(&self, func: &mut Function, rel_name: &str) {
        let mut hash: u32 = 5381;
        for c in rel_name.bytes() {
            hash = ((hash << 5).wrapping_add(hash)).wrapping_add(c as u32);
        }
        func.instruction(&Instruction::I32Const(hash as i32));
        func.instruction(&Instruction::Call(4)); // scan_delta_start is import #4
    }

    fn emit_scan_index_start(&self, func: &mut Function, rel_name: &str, col_idx: u32) {
        // Key is already on stack (i64).
        // Save to Local 0 (i64 scratch) to reorder args.
        func.instruction(&Instruction::LocalSet(0));

        let mut hash: u32 = 5381;
        for c in rel_name.bytes() {
            hash = ((hash << 5).wrapping_add(hash)).wrapping_add(c as u32);
        }

        func.instruction(&Instruction::I32Const(hash as i32));
        func.instruction(&Instruction::I32Const(col_idx as i32));
        func.instruction(&Instruction::LocalGet(0));
        func.instruction(&Instruction::Call(7)); // scan_index_start is import #7
    }

    fn emit_scan_aggregate_start(&self, func: &mut Function, rel_name: &str, ptr: i32, len: i32) {
        let mut hash: u32 = 5381;
        for c in rel_name.bytes() {
            hash = ((hash << 5).wrapping_add(hash)).wrapping_add(c as u32);
        }
        func.instruction(&Instruction::I32Const(hash as i32));
        func.instruction(&Instruction::I32Const(ptr));
        func.instruction(&Instruction::I32Const(len));
        func.instruction(&Instruction::Call(8)); // scan_aggregate_start is import #8
    }

    fn emit_scan_next(&self, func: &mut Function, iter_local: u32) {
        func.instruction(&Instruction::LocalGet(iter_local));
        func.instruction(&Instruction::Call(1));
    }

    fn emit_get_col(&self, func: &mut Function, tuple_local: u32, col_idx: u32) {
        func.instruction(&Instruction::LocalGet(tuple_local));
        func.instruction(&Instruction::I32Const(col_idx as i32));
        func.instruction(&Instruction::Call(2));
    }

    fn emit_insert_start(&self, func: &mut Function, rel_name: &str) {
        let mut hash: u32 = 5381;
        for c in rel_name.bytes() {
            hash = ((hash << 5).wrapping_add(hash)).wrapping_add(c as u32);
        }
        func.instruction(&Instruction::I32Const(hash as i32));
    }

    fn emit_insert_end(&self, func: &mut Function) {
        func.instruction(&Instruction::Call(3));
    }

    fn emit_merge_deltas(&self, func: &mut Function) {
        func.instruction(&Instruction::Call(5)); // merge_deltas is import #5
    }

    fn emit_debuglog(&self, func: &mut Function, val_local: u32) {
        func.instruction(&Instruction::LocalGet(val_local));
        func.instruction(&Instruction::Call(6)); // debuglog is import #6
    }
}

pub struct Codegen<'a, B: Backend> {
    ir: &'a mut Ir,
    stratified: Option<&'a StratifiedProgram<'a>>,
    backend: B,
}

struct FuncContext {
    var_map: HashMap<NameId, u32>,
    next_local: u32,

    // Locals reserved for iterators (start index)
    iter_base: u32,
    // Current offset into iterator locals
    iter_offset: u32,
}

impl<'a, B: Backend> Codegen<'a, B> {
    pub fn new(ir: &'a mut Ir, backend: B) -> Self {
        Self {
            ir,
            stratified: None,
            backend,
        }
    }

    pub fn new_with_stratified(
        ir: &'a mut Ir,
        stratified: &'a StratifiedProgram<'a>,
        backend: B,
    ) -> Self {
        Self {
            ir,
            stratified: Some(stratified),
            backend,
        }
    }

    pub fn generate(&mut self) -> Vec<u8> {
        let mut module = Module::new();

        // 1. Types
        let mut types = TypeSection::new();
        types.ty().function(vec![], vec![]);
        types.ty().function(vec![ValType::I32], vec![ValType::I32]);
        types
            .ty()
            .function(vec![ValType::I32, ValType::I32], vec![ValType::I64]);
        types
            .ty()
            .function(vec![ValType::I32, ValType::I64], vec![]);
        // Import #7: scan_index_start(rel_id, col_idx, val) -> iter_id
        types.ty().function(
            vec![ValType::I32, ValType::I32, ValType::I64],
            vec![ValType::I32],
        );
        // Import #8: scan_aggregate_start(rel_id, ptr, len) -> iter_id
        types.ty().function(
            vec![ValType::I32, ValType::I32, ValType::I32],
            vec![ValType::I32],
        );
        // Import #4: scan_delta_start(rel_id) -> iter_id
        // Same signature as scan_start: (i32) -> i32. Already defined.
        // Import #5: merge_deltas() -> i32 (bool)
        types.ty().function(vec![], vec![ValType::I32]);
        // Import #6: debuglog(i64) -> ()
        types.ty().function(vec![ValType::I64], vec![]);

        module.section(&types);

        // 2. Imports
        let mut imports = ImportSection::new();
        if std::any::type_name::<B>() == std::any::type_name::<WasmImportsBackend>() {
            imports.import("env", "scan_start", EntityType::Function(1));
            imports.import("env", "scan_next", EntityType::Function(1));
            imports.import("env", "get_col", EntityType::Function(2));
            imports.import("env", "insert", EntityType::Function(3));
            imports.import("env", "scan_delta_start", EntityType::Function(1));
            imports.import("env", "merge_deltas", EntityType::Function(6));
            imports.import("env", "debuglog", EntityType::Function(7));
            imports.import("env", "scan_index_start", EntityType::Function(4));
            imports.import("env", "scan_aggregate_start", EntityType::Function(5));
        }
        module.section(&imports);

        // 3. Functions
        let mut functions = FunctionSection::new();
        functions.function(0);
        module.section(&functions);

        // 3b. Memory
        let mut memories = MemorySection::new();
        memories.memory(wasm_encoder::MemoryType {
            minimum: 1,
            maximum: None,
            memory64: false,
            shared: false,
            page_size_log2: None,
        });
        module.section(&memories);

        // 4. Exports
        let mut exports = ExportSection::new();
        exports.export("run", ExportKind::Func, 9); // Run is now function 9 (0-8 are imports)
        exports.export("memory", ExportKind::Memory, 0);
        module.section(&exports);

        // 5. Code
        let mut codes = CodeSection::new();

        // If stratified is present, we use it to determine order and loops.
        // Otherwise fallback to simple rule iteration (naive).

        let mut ops = Vec::new();
        let mut loop_ops = Vec::new(); // (start_idx, end_idx) of ops that should be wrapped in loop

        if let Some(stratified) = self.stratified {
            let arena = stratified.arena();
            for stratum in stratified.strata() {
                use fxhash::FxHashSet;
                let mut stratum_pred_names = FxHashSet::default();
                for pred in &stratum {
                    if let Some(name) = arena.predicate_name(*pred) {
                        stratum_pred_names.insert(name);
                    }
                }

                // Identify rules
                let mut rule_ids = Vec::new();
                for (i, inst) in self.ir.insts.iter().enumerate() {
                    if let Inst::Rule { head, .. } = inst
                        && let Inst::Atom { predicate, .. } = self.ir.get(*head)
                    {
                        let head_name = self.ir.resolve_name(*predicate);
                        if stratum_pred_names.contains(head_name) {
                            rule_ids.push(InstId::new(i));
                        }
                    }
                }

                if rule_ids.is_empty() {
                    continue;
                }

                let mut is_recursive = false;
                for &rule_id in &rule_ids {
                    if let Inst::Rule { premises, .. } = self.ir.get(rule_id) {
                        for &premise in premises {
                            if let Inst::Atom { predicate, .. } = self.ir.get(premise) {
                                let pred_name = self.ir.resolve_name(*predicate);
                                if stratum_pred_names.contains(pred_name) {
                                    is_recursive = true;
                                    break;
                                }
                            }
                        }
                    }
                    if is_recursive {
                        break;
                    }
                }

                if !is_recursive {
                    for rule_id in rule_ids {
                        let planner = Planner::new(self.ir);
                        if let Ok(op) = planner.plan_rule(rule_id) {
                            ops.push(op);
                        }
                    }
                    // No merge_deltas needed for non-recursive? Actually yes, to move to stable.
                    // But we can't emit just function call here easily, we need an Op.
                    // Or we emit special Op::MergeDeltas later.
                    // For now, assume host handles it or we add a special Op.
                } else {
                    // Recursive Stratum

                    // 1. Initial Step (Scan)
                    for &rule_id in &rule_ids {
                        let planner = Planner::new(self.ir);
                        if let Ok(op) = planner.plan_rule(rule_id) {
                            ops.push(op);
                        }
                    }

                    // Merge Deltas after initial step
                    // We need a marker in `ops` to say "Emit Merge Deltas here".
                    // Since Op enum is shared, maybe we can add Op::Custom or similar?
                    // Or we track indices.
                    let loop_start_idx = ops.len();

                    // 2. Iterative Step (ScanDelta)
                    for &rule_id in &rule_ids {
                        let premises = if let Inst::Rule { premises, .. } = self.ir.get(rule_id) {
                            premises.clone()
                        } else {
                            continue;
                        };

                        for &premise in &premises {
                            let (predicate, pred_name) =
                                if let Inst::Atom { predicate, .. } = self.ir.get(premise) {
                                    (*predicate, self.ir.resolve_name(*predicate).to_string())
                                } else {
                                    continue;
                                };

                            if stratum_pred_names.contains(pred_name.as_str()) {
                                let planner = Planner::new(self.ir).with_delta(predicate);
                                if let Ok(op) = planner.plan_rule(rule_id) {
                                    ops.push(op);
                                }
                            }
                        }
                    }
                    let loop_end_idx = ops.len();
                    loop_ops.push((loop_start_idx, loop_end_idx));
                }
            }
        } else {
            // Naive / No Stratification info provided
            let rule_ids: Vec<_> = self
                .ir
                .insts
                .iter()
                .enumerate()
                .filter_map(|(i, inst)| {
                    if let Inst::Rule { .. } = inst {
                        Some(mangle_ir::InstId::new(i))
                    } else {
                        None
                    }
                })
                .collect();

            for rule_id in &rule_ids {
                let planner = Planner::new(self.ir);
                if let Ok(op) = planner.plan_rule(*rule_id) {
                    ops.push(op);
                }
            }
        }

        let mut locals = vec![
            (1, ValType::I64), // Local 0: scratch / unused (for i64 values)
            (1, ValType::I32), // Local 1: tuple_ptr
        ];

        let mut ctx = FuncContext {
            var_map: HashMap::new(),
            next_local: 2,
            iter_base: 0,
            iter_offset: 0,
        };

        // Pass 1: Collect Vars
        let mut total_iter_count = 0;
        for op in &ops {
            total_iter_count += self.collect_vars(op, &mut ctx);
        }

        // Variable locals
        for _ in 0..ctx.var_map.len() {
            locals.push((1, ValType::I64));
        }

        // Iterator locals
        ctx.iter_base = ctx.next_local;
        for _ in 0..total_iter_count {
            locals.push((1, ValType::I32));
        }

        let mut run_func = Function::new(locals);

        // Pass 2: Emit
        let mut current_op_idx = 0;

        // Helper to check if we are entering a loop
        let mut loop_iter = loop_ops.into_iter();
        let mut next_loop = loop_iter.next();

        while current_op_idx < ops.len() {
            if let Some((start, end)) = next_loop
                && current_op_idx == start
            {
                // We are at start of recursive block (after initial step)

                // 1. Merge Deltas from initial step
                self.backend.emit_merge_deltas(&mut run_func);
                run_func.instruction(&Instruction::Drop); // ignore result of first merge

                // 2. Start Loop
                run_func.instruction(&Instruction::Loop(wasm_encoder::BlockType::Empty));

                // 3. Emit body (iterative step)
                while current_op_idx < end {
                    self.emit_op(&mut run_func, &ops[current_op_idx], &mut ctx);
                    current_op_idx += 1;
                }

                // 4. Merge Deltas & Check Termination
                self.backend.emit_merge_deltas(&mut run_func);
                // Stack has: i32 (1 if changes, 0 if no changes)
                run_func.instruction(&Instruction::BrIf(0)); // Branch to Loop start if changes

                run_func.instruction(&Instruction::End); // End Loop

                next_loop = loop_iter.next();
                continue;
            }

            self.emit_op(&mut run_func, &ops[current_op_idx], &mut ctx);
            // After non-recursive ops, we should technically merge deltas too?
            // Or we rely on the next loop doing it?
            // For non-recursive, we should probably merge to ensure stable is updated for next stratum.
            // But we didn't add logic to identify non-recursive boundaries precisely here unless we used the loop logic for everything.
            // For now, let's leave it. The next stratum will just use Scan which reads stable+delta.
            // Wait, Scan only reads Stable+Delta if the store supports it.
            // But Host Scan reads everything.
            // The issue is if we have S1 -> S2.
            // S1 produces facts. S2 consumes S1.
            // If S1 facts are in "New" bucket, does S2 Scan see them?
            // MemHost::scan combines stable + delta. Yes.
            // So it should be fine.

            current_op_idx += 1;
        }

        run_func.instruction(&Instruction::End);
        codes.function(&run_func);
        module.section(&codes);

        module.finish()
    }

    fn collect_vars(&self, op: &Op, ctx: &mut FuncContext) -> usize {
        let mut count = 0;
        match op {
            Op::Iterate { source, body } => {
                count += 1;
                match source {
                    DataSource::Scan { vars, .. }
                    | DataSource::ScanDelta { vars, .. }
                    | DataSource::IndexLookup { vars, .. } => {
                        for v in vars {
                            if !ctx.var_map.contains_key(v) {
                                ctx.var_map.insert(*v, ctx.next_local);
                                ctx.next_local += 1;
                            }
                        }
                    }
                }
                count += self.collect_vars(body, ctx);
            }
            Op::Let { var, body, .. } => {
                if !ctx.var_map.contains_key(var) {
                    ctx.var_map.insert(*var, ctx.next_local);
                    ctx.next_local += 1;
                }
                count += self.collect_vars(body, ctx);
            }
            Op::Filter { body, .. } => {
                count += self.collect_vars(body, ctx);
            }
            Op::Seq(ops) => {
                for o in ops {
                    count += self.collect_vars(o, ctx);
                }
            }
            Op::GroupBy {
                body,
                vars,
                aggregates,
                ..
            } => {
                // Collect vars from temp relation scan
                for v in vars {
                    if !ctx.var_map.contains_key(v) {
                        ctx.var_map.insert(*v, ctx.next_local);
                        ctx.next_local += 1;
                    }
                }
                // Collect vars from aggregates (outputs)
                for agg in aggregates {
                    if !ctx.var_map.contains_key(&agg.var) {
                        ctx.var_map.insert(agg.var, ctx.next_local);
                        ctx.next_local += 1;
                    }
                }
                // Recurse body? GroupBy body inserts head. No new loops?
                // Wait, GroupBy body might have Lets.
                count += self.collect_vars(body, ctx);
            }
            _ => {}
        }
        count
    }

    fn emit_op(&self, func: &mut Function, op: &Op, ctx: &mut FuncContext) {
        match op {
            Op::Iterate { source, body } => {
                match source {
                    DataSource::Scan { relation, vars }
                    | DataSource::ScanDelta { relation, vars } => {
                        let iter_local = ctx.iter_base + ctx.iter_offset;
                        ctx.iter_offset += 1;

                        let rel_name = self.ir.resolve_name(*relation);
                        if let DataSource::ScanDelta { .. } = source {
                            self.backend.emit_scan_delta_start(func, rel_name);
                        } else {
                            self.backend.emit_scan_start(func, rel_name);
                        }
                        func.instruction(&Instruction::LocalSet(iter_local));

                        // Block for break target
                        func.instruction(&Instruction::Block(wasm_encoder::BlockType::Empty));
                        // Loop
                        func.instruction(&Instruction::Loop(wasm_encoder::BlockType::Empty));

                        self.backend.emit_scan_next(func, iter_local);
                        func.instruction(&Instruction::LocalTee(1)); // tuple_ptr (Local 1)

                        // Break if null (0)
                        func.instruction(&Instruction::I32Eqz);
                        func.instruction(&Instruction::BrIf(1));

                        // Bind vars
                        for (i, var) in vars.iter().enumerate() {
                            if let Some(&local_idx) = ctx.var_map.get(var) {
                                self.backend.emit_get_col(func, 1, i as u32);
                                func.instruction(&Instruction::LocalSet(local_idx));
                            }
                        }

                        // Body
                        self.emit_op(func, body, ctx);

                        func.instruction(&Instruction::Br(0));
                        func.instruction(&Instruction::End); // End Loop
                        func.instruction(&Instruction::End); // End Block
                    }
                    DataSource::IndexLookup {
                        relation,
                        col_idx,
                        key,
                        vars,
                    } => {
                        let iter_local = ctx.iter_base + ctx.iter_offset;
                        ctx.iter_offset += 1;

                        let rel_name = self.ir.resolve_name(*relation);
                        self.emit_operand(func, key, ctx);
                        self.backend
                            .emit_scan_index_start(func, rel_name, *col_idx as u32);
                        func.instruction(&Instruction::LocalSet(iter_local));

                        // Block for break target
                        func.instruction(&Instruction::Block(wasm_encoder::BlockType::Empty));
                        // Loop
                        func.instruction(&Instruction::Loop(wasm_encoder::BlockType::Empty));

                        self.backend.emit_scan_next(func, iter_local);
                        func.instruction(&Instruction::LocalTee(1)); // tuple_ptr (Local 1)

                        // Break if null (0)
                        func.instruction(&Instruction::I32Eqz);
                        func.instruction(&Instruction::BrIf(1));

                        // Bind vars
                        for (i, var) in vars.iter().enumerate() {
                            if let Some(&local_idx) = ctx.var_map.get(var) {
                                self.backend.emit_get_col(func, 1, i as u32);
                                func.instruction(&Instruction::LocalSet(local_idx));
                            }
                        }

                        // Body
                        self.emit_op(func, body, ctx);

                        func.instruction(&Instruction::Br(0));
                        func.instruction(&Instruction::End); // End Loop
                        func.instruction(&Instruction::End); // End Block
                    }
                }
            }
            Op::GroupBy { .. } => {}
            Op::Filter { cond, body } => {
                self.emit_condition(func, cond, ctx);
                func.instruction(&Instruction::If(wasm_encoder::BlockType::Empty));
                self.emit_op(func, body, ctx);
                func.instruction(&Instruction::End);
            }
            Op::Let { var, expr, body } => {
                self.emit_expr(func, expr, ctx);
                if let Some(&local_idx) = ctx.var_map.get(var) {
                    func.instruction(&Instruction::LocalSet(local_idx));
                } else {
                    func.instruction(&Instruction::Drop);
                }
                self.emit_op(func, body, ctx);
            }
            Op::Insert { relation, args } => {
                self.backend
                    .emit_insert_start(func, self.ir.resolve_name(*relation));
                for arg in args {
                    self.emit_operand(func, arg, ctx);
                }
                self.backend.emit_insert_end(func);
            }
            Op::Seq(ops) => {
                for o in ops {
                    self.emit_op(func, o, ctx);
                }
            }
            _ => {}
        }
    }

    fn emit_condition(&self, func: &mut Function, cond: &Condition, ctx: &FuncContext) {
        match cond {
            Condition::Cmp { op, left, right } => {
                self.emit_operand(func, left, ctx);
                self.emit_operand(func, right, ctx);
                match op {
                    CmpOp::Eq => func.instruction(&Instruction::I64Eq),
                    CmpOp::Neq => func.instruction(&Instruction::I64Ne),
                    CmpOp::Lt => func.instruction(&Instruction::I64LtS),
                    CmpOp::Le => func.instruction(&Instruction::I64LeS),
                    CmpOp::Gt => func.instruction(&Instruction::I64GtS),
                    CmpOp::Ge => func.instruction(&Instruction::I64GeS),
                };
            }
            Condition::Call { .. } => {
                func.instruction(&Instruction::I32Const(1));
            }
            _ => {
                func.instruction(&Instruction::I32Const(1));
            }
        }
    }

    fn emit_expr(&self, func: &mut Function, expr: &Expr, ctx: &FuncContext) {
        match expr {
            Expr::Value(op) => self.emit_operand(func, op, ctx),
            Expr::Call { function, args } => {
                for arg in args {
                    self.emit_operand(func, arg, ctx);
                }
                let name = self.ir.resolve_name(*function);
                match name {
                    "fn:plus" => func.instruction(&Instruction::I64Add),
                    "fn:minus" => func.instruction(&Instruction::I64Sub),
                    _ => {
                        if !args.is_empty() {
                            func.instruction(&Instruction::Drop);
                        }
                        if args.len() > 1 {
                            func.instruction(&Instruction::Drop);
                        }
                        func.instruction(&Instruction::I64Const(0))
                    }
                };
            }
        }
    }

    fn emit_operand(&self, func: &mut Function, op: &Operand, ctx: &FuncContext) {
        match op {
            Operand::Var(v) => {
                if let Some(&idx) = ctx.var_map.get(v) {
                    func.instruction(&Instruction::LocalGet(idx));
                } else {
                    func.instruction(&Instruction::I64Const(0));
                }
            }
            Operand::Const(c) => match c {
                Constant::Number(n) => {
                    func.instruction(&Instruction::I64Const(*n));
                }
                Constant::String(_) => {
                    func.instruction(&Instruction::I64Const(0));
                }
            },
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use mangle_analysis::LoweringContext;
    use mangle_ast as ast;

    #[test]
    fn test_codegen_with_imports() {
        let arena = ast::Arena::new_with_global_interner();
        let foo = arena.predicate_sym("foo", Some(1));
        let bar = arena.predicate_sym("bar", Some(1));
        let x = arena.variable("X");

        let clause = ast::Clause {
            head: arena.atom(foo, &[x]),
            premises: arena
                .alloc_slice_copy(&[arena.alloc(ast::Term::Atom(arena.atom(bar, &[x])))]),
            transform: &[],
        };
        let unit = ast::Unit {
            decls: &[],
            clauses: arena.alloc_slice_copy(&[&clause]),
        };

        let ctx = LoweringContext::new(&arena);
        let mut ir = ctx.lower_unit(&unit);

        let mut codegen = Codegen::new(&mut ir, WasmImportsBackend);
        let wasm = codegen.generate();

        assert!(!wasm.is_empty());

        use wasmparser::Payload;
        let parser = wasmparser::Parser::new(0);
        let mut found_import = false;
        let mut found_code = false;

        for payload in parser.parse_all(&wasm) {
            match payload.expect("parsing failed") {
                Payload::ImportSection(reader) => {
                    for import in reader {
                        let import = import.expect("import failed");
                        if import.module == "env" && import.name == "scan_start" {
                            found_import = true;
                        }
                    }
                }
                Payload::CodeSectionEntry(_) => {
                    found_code = true;
                }
                _ => {}
            }
        }
        assert!(found_import, "scan_start import not found");
        assert!(found_code, "code section empty");
    }
}
