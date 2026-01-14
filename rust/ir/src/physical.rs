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

//! Physical Plan IR for Mangle.
//!
//! This represents the imperative execution logic (loops, joins, inserts)
//! derived from the declarative logical IR.

use crate::{NameId, StringId};

#[derive(Debug, Clone, PartialEq)]
pub enum Op {
    /// A no-op.
    Nop,

    /// Sequence of operations executed in order.
    Seq(Vec<Op>),

    /// Iterate over a data source.
    /// For each tuple yielded by `source`, `body` is executed.
    /// Variables defined in `source` are bound and available in `body`.
    Iterate { source: DataSource, body: Box<Op> },

    /// Filter / Check condition.
    /// If `cond` evaluates to true, `body` is executed.
    Filter { cond: Condition, body: Box<Op> },

    /// Insert a tuple into a relation.
    /// All variables in `args` must be bound.
    Insert {
        relation: NameId,
        args: Vec<Operand>,
    },

    /// Calculate a value and bind it to a variable.
    /// `let var = expr`
    Let {
        var: NameId,
        expr: Expr,
        body: Box<Op>,
    },

    /// GroupBy operation.
    /// Scans `source` (binding columns to `vars`), groups by `keys`, computes `aggregates`
    /// for each group, and then executes `body` for each group.
    GroupBy {
        source: NameId,    // Relation to scan
        vars: Vec<NameId>, // Variables to bind to source columns
        keys: Vec<NameId>, // Variables to group by (must be in `vars` or previously bound?)
        // Typically `keys` are subset of `vars`.
        aggregates: Vec<Aggregate>,
        body: Box<Op>,
    },
}

#[derive(Debug, Clone, PartialEq)]
pub struct Aggregate {
    pub var: NameId,
    pub func: NameId,
    pub args: Vec<Operand>,
}

#[derive(Debug, Clone, PartialEq)]
pub enum DataSource {
    /// Scan a relation (iterate over all tuples).
    /// Binds the variables in `vars` to the columns of the relation.
    Scan { relation: NameId, vars: Vec<NameId> },

    /// Scan only the "delta" set of a relation (new facts from last iteration).
    ScanDelta { relation: NameId, vars: Vec<NameId> },

    /// Lookup in an index.
    /// `col_idx`: The column index to lookup on.
    /// `key`: The value to look up.
    /// `vars`: Variables to bind to the *other* columns (or all columns?).
    /// For simplicity: `vars` maps to the relation columns. The column at `col_idx`
    /// is already bound (to `key`), but we might re-bind it or check it.
    IndexLookup {
        relation: NameId,
        col_idx: usize,
        key: Operand,
        vars: Vec<NameId>,
    },
}

#[derive(Debug, Clone, Copy, PartialEq)]
pub enum CmpOp {
    Eq,
    Neq,
    Lt,
    Le,
    Gt,
    Ge,
}

#[derive(Debug, Clone, PartialEq)]
pub enum Condition {
    /// Comparison of two operands.
    Cmp {
        op: CmpOp,
        left: Operand,
        right: Operand,
    },
    /// Negation check: !exists(...)
    Negation {
        relation: NameId,
        args: Vec<Operand>,
    },
    /// Call to a boolean function / predicate (e.g. starts_with).
    Call {
        function: NameId,
        args: Vec<Operand>,
    },
}

#[derive(Debug, Clone, PartialEq)]
pub enum Expr {
    // Basic value
    Value(Operand),
    // Function call (arithmetic or built-in)
    Call {
        function: NameId,
        args: Vec<Operand>,
    },
}

#[derive(Debug, Clone, PartialEq)]
pub enum Operand {
    Var(NameId),
    Const(Constant),
}

#[derive(Clone, Debug, PartialEq)]
pub enum Constant {
    Number(i64),
    String(StringId),
    // ...
}
