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

//! # Mangle Interpreter
//!
//! A pure Rust interpreter for the Mangle Intermediate Representation (IR).
//!
//! This crate enables the **Edge Mode** of execution, allowing Mangle programs
//! to run on devices or in environments where a WebAssembly runtime is not available
//! or desired.
//!
//! It executes the Physical IR operations (`Op`) directly.
//!
//! ## Usage
//!
//! See `mangle-driver` for the high-level API to compile and execute source code.

use anyhow::{Result, anyhow};
use mangle_ir::physical::{Aggregate, CmpOp, Condition, Constant, DataSource, Expr, Op, Operand};
use mangle_ir::{Ir, NameId};
use std::collections::HashMap;

#[derive(Debug, Clone, PartialEq, PartialOrd, Eq, Hash)]
pub enum Value {
    Number(i64),
    String(String),
    Null, // Used for iteration end or missing
}

/// Abstract interface for relation storage.
pub trait Store {
    /// Returns an iterator over all tuples in the relation.
    /// Returns an error if the relation does not exist.
    fn scan(&self, relation: &str) -> Result<Box<dyn Iterator<Item = Vec<Value>> + '_>>;

    /// Returns an iterator over only the new tuples added in the last iteration.
    fn scan_delta(&self, relation: &str) -> Result<Box<dyn Iterator<Item = Vec<Value>> + '_>>;

    /// Returns an iterator over tuples being collected for the next iteration.
    fn scan_next_delta(&self, relation: &str) -> Result<Box<dyn Iterator<Item = Vec<Value>> + '_>>;

    /// Returns an iterator over tuples in the relation matching a key in a column.
    fn scan_index(
        &self,
        relation: &str,
        col_idx: usize,
        key: &Value,
    ) -> Result<Box<dyn Iterator<Item = Vec<Value>> + '_>>;

    /// Returns an iterator over delta tuples matching a key in a column.
    fn scan_delta_index(
        &self,
        relation: &str,
        col_idx: usize,
        key: &Value,
    ) -> Result<Box<dyn Iterator<Item = Vec<Value>> + '_>>;

    /// Inserts a tuple into the relation (specifically into the delta/new set).
    /// Returns true if it was new.
    fn insert(&mut self, relation: &str, tuple: Vec<Value>) -> Result<bool>;

    /// Merges current deltas into the stable set of facts.
    fn merge_deltas(&mut self);

    /// Ensures a relation exists in the store.
    fn create_relation(&mut self, relation: &str);
}

/// A simple in-memory implementation of `Store`.
/// Supports semi-naive evaluation by tracking "stable" and "delta" facts.
#[derive(Default)]
pub struct MemStore {
    // Stable facts from previous iterations
    stable: HashMap<String, Vec<Vec<Value>>>,
    // New facts from the current iteration
    delta: HashMap<String, Vec<Vec<Value>>>,
    // Facts being collected for the next iteration
    next_delta: HashMap<String, Vec<Vec<Value>>>,

    // Secondary indexes: (relation_name, col_idx) -> { Value -> [row_indices] }
    // These only index stable facts for simplicity, or we re-build them.
    // Actually, let's index ALL facts (stable + delta).
    stable_indexes: HashMap<(String, usize), HashMap<Value, Vec<usize>>>,
    delta_indexes: HashMap<(String, usize), HashMap<Value, Vec<usize>>>,
}

impl MemStore {
    pub fn new() -> Self {
        Self::default()
    }

    /// Registers a relation (creating it if absent) to allow scanning it.
    pub fn create_relation(&mut self, relation: &str) {
        self.stable.entry(relation.to_string()).or_default();
    }

    /// Add a fact manually (for testing/setup). Auto-creates relation in stable.
    pub fn add_fact(&mut self, relation: &str, args: Vec<Value>) {
        let table = self.stable.entry(relation.to_string()).or_default();
        if !table.contains(&args) {
            let row_idx = table.len();
            table.push(args.clone());
            // Update stable index
            for (col_idx, val) in args.into_iter().enumerate() {
                self.stable_indexes
                    .entry((relation.to_string(), col_idx))
                    .or_default()
                    .entry(val)
                    .or_default()
                    .push(row_idx);
            }
        }
    }

    pub fn get_facts(&self, relation: &str) -> Vec<Vec<Value>> {
        let mut all = self.stable.get(relation).cloned().unwrap_or_default();
        if let Some(d) = self.delta.get(relation) {
            all.extend(d.iter().cloned());
        }
        all
    }
}

impl Store for MemStore {
    fn scan(&self, relation: &str) -> Result<Box<dyn Iterator<Item = Vec<Value>> + '_>> {
        let s = self.stable.get(relation).into_iter().flatten().cloned();
        let d = self.delta.get(relation).into_iter().flatten().cloned();
        Ok(Box::new(s.chain(d)))
    }

    fn scan_delta(&self, relation: &str) -> Result<Box<dyn Iterator<Item = Vec<Value>> + '_>> {
        match self.delta.get(relation) {
            Some(tuples) => Ok(Box::new(tuples.iter().cloned())),
            None => Ok(Box::new(std::iter::empty())),
        }
    }

    fn scan_next_delta(&self, relation: &str) -> Result<Box<dyn Iterator<Item = Vec<Value>> + '_>> {
        match self.next_delta.get(relation) {
            Some(tuples) => Ok(Box::new(tuples.iter().cloned())),
            None => Ok(Box::new(std::iter::empty())),
        }
    }

    fn scan_index(
        &self,
        relation: &str,
        col_idx: usize,
        key: &Value,
    ) -> Result<Box<dyn Iterator<Item = Vec<Value>> + '_>> {
        let mut results: Vec<Vec<Value>> = Vec::new();

        if let Some(idx_map) = self.stable_indexes.get(&(relation.to_string(), col_idx))
            && let Some(row_indices) = idx_map.get(key)
            && let Some(table) = self.stable.get(relation)
        {
            for &i in row_indices {
                results.push(table[i].clone());
            }
        }

        if let Some(idx_map) = self.delta_indexes.get(&(relation.to_string(), col_idx))
            && let Some(row_indices) = idx_map.get(key)
            && let Some(table) = self.delta.get(relation)
        {
            for &i in row_indices {
                results.push(table[i].clone());
            }
        }

        Ok(Box::new(results.into_iter()))
    }

    fn scan_delta_index(
        &self,
        relation: &str,
        col_idx: usize,
        key: &Value,
    ) -> Result<Box<dyn Iterator<Item = Vec<Value>> + '_>> {
        let mut results: Vec<Vec<Value>> = Vec::new();

        if let Some(idx_map) = self.delta_indexes.get(&(relation.to_string(), col_idx))
            && let Some(row_indices) = idx_map.get(key)
            && let Some(table) = self.delta.get(relation)
        {
            for &i in row_indices {
                results.push(table[i].clone());
            }
        }

        Ok(Box::new(results.into_iter()))
    }

    fn insert(&mut self, relation: &str, tuple: Vec<Value>) -> Result<bool> {
        // Check if fact is already in stable, delta, or next_delta
        if self
            .stable
            .get(relation)
            .is_some_and(|v| v.contains(&tuple))
            || self.delta.get(relation).is_some_and(|v| v.contains(&tuple))
            || self
                .next_delta
                .get(relation)
                .is_some_and(|v| v.contains(&tuple))
        {
            return Ok(false);
        }

        self.next_delta
            .entry(relation.to_string())
            .or_default()
            .push(tuple);
        Ok(true)
    }

    fn merge_deltas(&mut self) {
        // 1. Move current delta to stable
        for (rel_name, mut tuples) in self.delta.drain() {
            let table = self.stable.entry(rel_name.clone()).or_default();
            for tuple in tuples.drain(..) {
                let row_idx = table.len();
                // Update stable index
                for (col_idx, val) in tuple.iter().enumerate() {
                    self.stable_indexes
                        .entry((rel_name.clone(), col_idx))
                        .or_default()
                        .entry(val.clone())
                        .or_default()
                        .push(row_idx);
                }
                table.push(tuple);
            }
        }
        self.delta_indexes.clear();

        // 2. Move next_delta to delta and build delta index
        self.delta = std::mem::take(&mut self.next_delta);
        for (rel_name, tuples) in &self.delta {
            for (row_idx, tuple) in tuples.iter().enumerate() {
                for (col_idx, val) in tuple.iter().enumerate() {
                    self.delta_indexes
                        .entry((rel_name.clone(), col_idx))
                        .or_default()
                        .entry(val.clone())
                        .or_default()
                        .push(row_idx);
                }
            }
        }
    }

    fn create_relation(&mut self, relation: &str) {
        self.stable.entry(relation.to_string()).or_default();
    }
}

/// A pure Rust interpreter for Mangle IR.
pub struct Interpreter<'a> {
    ir: &'a Ir,
    store: Box<dyn Store + 'a>,
}

struct Env {
    vars: HashMap<NameId, Value>,
}

impl Env {
    fn new() -> Self {
        Self {
            vars: HashMap::new(),
        }
    }
}

impl<'a> Interpreter<'a> {
    pub fn new(ir: &'a Ir, store: Box<dyn Store + 'a>) -> Self {
        Self { ir, store }
    }

    /// Helper to get the underlying store (e.g. to inspect results).
    pub fn store(&self) -> &dyn Store {
        &*self.store
    }

    /// Helper to get the underlying store mutably.
    pub fn store_mut(&mut self) -> &mut dyn Store {
        &mut *self.store
    }

    /// Executes the operation and returns the number of facts inserted.
    pub fn execute(&mut self, op: &Op) -> Result<usize> {
        let mut env = Env::new();
        self.exec_op(op, &mut env)
    }

    fn exec_op(&mut self, op: &Op, env: &mut Env) -> Result<usize> {
        match op {
            Op::Nop => Ok(0),
            Op::Seq(ops) => {
                let mut count = 0;
                for o in ops {
                    count += self.exec_op(o, env)?;
                }
                Ok(count)
            }
            Op::Iterate { source, body } => {
                let mut count = 0;
                match source {
                    DataSource::Scan { relation, vars } => {
                        let rel_name = self.ir.resolve_name(*relation);
                        let iter = self.store.scan(rel_name)?;
                        let tuples: Vec<_> = iter.collect();

                        for tuple in tuples {
                            if tuple.len() != vars.len() {
                                continue;
                            }
                            for (i, var) in vars.iter().enumerate() {
                                env.vars.insert(*var, tuple[i].clone());
                            }
                            count += self.exec_op(body, env)?;
                        }
                    }
                    DataSource::ScanDelta { relation, vars } => {
                        let rel_name = self.ir.resolve_name(*relation);
                        let iter = self.store.scan_delta(rel_name)?;
                        let tuples: Vec<_> = iter.collect();

                        for tuple in tuples {
                            if tuple.len() != vars.len() {
                                continue;
                            }
                            for (i, var) in vars.iter().enumerate() {
                                env.vars.insert(*var, tuple[i].clone());
                            }
                            count += self.exec_op(body, env)?;
                        }
                    }
                    DataSource::IndexLookup {
                        relation,
                        col_idx,
                        key,
                        vars,
                    } => {
                        let rel_name = self.ir.resolve_name(*relation);
                        let key_val = self.eval_operand(key, env)?;

                        let iter = self.store.scan_index(rel_name, *col_idx, &key_val)?;
                        let tuples: Vec<_> = iter.collect();

                        for tuple in tuples {
                            if tuple.len() != vars.len() {
                                continue;
                            }
                            for (i, var) in vars.iter().enumerate() {
                                env.vars.insert(*var, tuple[i].clone());
                            }
                            count += self.exec_op(body, env)?;
                        }
                    }
                }
                Ok(count)
            }
            Op::Filter { cond, body } => {
                if self.eval_cond(cond, env)? {
                    self.exec_op(body, env)
                } else {
                    Ok(0)
                }
            }
            Op::Insert { relation, args } => {
                let rel_name = self.ir.resolve_name(*relation);
                let mut tuple = Vec::new();
                for arg in args {
                    tuple.push(self.eval_operand(arg, env)?);
                }
                if self.store.insert(rel_name, tuple)? {
                    Ok(1)
                } else {
                    Ok(0)
                }
            }
            Op::Let { var, expr, body } => {
                let val = self.eval_expr(expr, env)?;
                env.vars.insert(*var, val);
                self.exec_op(body, env)
            }
            Op::GroupBy {
                source,
                vars,
                keys,
                aggregates,
                body,
            } => {
                let rel_name = self.ir.resolve_name(*source);

                // For GroupBy, we must scan ALL available facts including ones just produced in this stratum
                // if we want to match Go implementation's behavior for non-recursive strata.
                let iter = self.store.scan(rel_name)?;
                let mut tuples: Vec<_> = iter.collect();

                // Also scan next_delta if it's the same relation
                if let Ok(nd_iter) = self.store.scan_next_delta(rel_name) {
                    tuples.extend(nd_iter);
                }

                let mut groups: HashMap<Vec<Value>, Vec<Vec<Value>>> = HashMap::new();

                for tuple in tuples {
                    if tuple.len() != vars.len() {
                        continue;
                    }
                    // Temporarily bind variables to extract key
                    for (i, var) in vars.iter().enumerate() {
                        env.vars.insert(*var, tuple[i].clone());
                    }

                    let mut key = Vec::new();
                    for k in keys {
                        if let Some(val) = env.vars.get(k) {
                            key.push(val.clone());
                        } else {
                            // Should not happen if well-typed
                            key.push(Value::Null);
                        }
                    }
                    groups.entry(key).or_default().push(tuple);
                }

                let mut count = 0;
                for (key, group_tuples) in groups {
                    // Bind keys
                    for (i, k) in keys.iter().enumerate() {
                        env.vars.insert(*k, key[i].clone());
                    }

                    // Compute aggregates
                    for agg in aggregates {
                        let val = self.eval_aggregate(agg, &group_tuples, vars, env)?;
                        env.vars.insert(agg.var, val);
                    }

                    count += self.exec_op(body, env)?;
                }
                Ok(count)
            }
        }
    }

    fn eval_aggregate(
        &self,
        agg: &Aggregate,
        group: &[Vec<Value>],
        vars: &[NameId],
        env: &mut Env,
    ) -> Result<Value> {
        let fn_name = self.ir.resolve_name(agg.func);
        match fn_name {
            "fn:count" => Ok(Value::Number(group.len() as i64)),
            "fn:sum" => {
                let mut sum = 0;
                // Assuming single argument for sum
                let arg = agg
                    .args
                    .first()
                    .ok_or_else(|| anyhow!("fn:sum requires 1 argument"))?;

                for tuple in group {
                    // We need to re-bind vars for each tuple to evaluate arg
                    for (i, var) in vars.iter().enumerate() {
                        env.vars.insert(*var, tuple[i].clone());
                    }
                    let val = self.eval_operand(arg, env)?;
                    if let Value::Number(n) = val {
                        sum += n;
                    }
                }
                Ok(Value::Number(sum))
            }
            "fn:max" => {
                let mut max_val = None;
                let arg = agg
                    .args
                    .first()
                    .ok_or_else(|| anyhow!("fn:max requires 1 argument"))?;

                for tuple in group {
                    for (i, var) in vars.iter().enumerate() {
                        env.vars.insert(*var, tuple[i].clone());
                    }
                    let val = self.eval_operand(arg, env)?;
                    match max_val {
                        None => max_val = Some(val),
                        Some(ref m) => {
                            if val > *m {
                                max_val = Some(val);
                            }
                        }
                    }
                }
                max_val.ok_or_else(|| anyhow!("fn:max on empty group"))
            }
            "fn:min" => {
                let mut min_val = None;
                let arg = agg
                    .args
                    .first()
                    .ok_or_else(|| anyhow!("fn:min requires 1 argument"))?;

                for tuple in group {
                    for (i, var) in vars.iter().enumerate() {
                        env.vars.insert(*var, tuple[i].clone());
                    }
                    let val = self.eval_operand(arg, env)?;
                    match min_val {
                        None => min_val = Some(val),
                        Some(ref m) => {
                            if val < *m {
                                min_val = Some(val);
                            }
                        }
                    }
                }
                min_val.ok_or_else(|| anyhow!("fn:min on empty group"))
            }
            _ => Err(anyhow!("Unknown aggregation function: {}", fn_name)),
        }
    }

    fn eval_cond(&self, cond: &Condition, env: &Env) -> Result<bool> {
        match cond {
            Condition::Cmp { op, left, right } => {
                let l = self.eval_operand(left, env)?;
                let r = self.eval_operand(right, env)?;
                match op {
                    CmpOp::Eq => Ok(l == r),
                    CmpOp::Neq => Ok(l != r),
                    CmpOp::Lt => Ok(l < r),
                    CmpOp::Le => Ok(l <= r),
                    CmpOp::Gt => Ok(l > r),
                    CmpOp::Ge => Ok(l >= r),
                }
            }
            Condition::Negation { relation, args } => {
                let rel_name = self.ir.resolve_name(*relation);
                let iter = self.store.scan(rel_name)?;
                for tuple in iter {
                    let mut mat = true;
                    if tuple.len() != args.len() {
                        continue;
                    }
                    for (i, arg) in args.iter().enumerate() {
                        let val = self.eval_operand(arg, env)?;
                        if tuple[i] != val {
                            mat = false;
                            break;
                        }
                    }
                    if mat {
                        return Ok(false); // Found match, negation fails
                    }
                }
                Ok(true) // No match found
            }
            Condition::Call { .. } => {
                // TODO: Implement boolean calls
                Ok(true)
            }
        }
    }

    fn eval_expr(&self, expr: &Expr, env: &Env) -> Result<Value> {
        match expr {
            Expr::Value(op) => self.eval_operand(op, env),
            Expr::Call { function, args } => {
                let fn_name = self.ir.resolve_name(*function);
                let mut vals = Vec::new();
                for arg in args {
                    vals.push(self.eval_operand(arg, env)?);
                }
                match fn_name {
                    "fn:plus" => {
                        if let (Value::Number(a), Value::Number(b)) = (&vals[0], &vals[1]) {
                            Ok(Value::Number(a + b))
                        } else {
                            Err(anyhow!("Type mismatch for fn:plus"))
                        }
                    }
                    "fn:minus" => {
                        if let (Value::Number(a), Value::Number(b)) = (&vals[0], &vals[1]) {
                            Ok(Value::Number(a - b))
                        } else {
                            Err(anyhow!("Type mismatch for fn:minus"))
                        }
                    }
                    _ => Err(anyhow!("Unknown function: {}", fn_name)),
                }
            }
        }
    }

    fn eval_operand(&self, op: &Operand, env: &Env) -> Result<Value> {
        match op {
            Operand::Var(v) => env
                .vars
                .get(v)
                .cloned()
                .ok_or_else(|| anyhow!("Variable not found")),
            Operand::Const(c) => match c {
                Constant::Number(n) => Ok(Value::Number(*n)),
                Constant::String(sid) => {
                    Ok(Value::String(self.ir.resolve_string(*sid).to_string()))
                }
            },
        }
    }
}
