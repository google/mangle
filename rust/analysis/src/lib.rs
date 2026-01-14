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

//! Analysis and Transformation Pipeline for Mangle.
//!
//! This crate provides the core analysis passes and transformations that turn
//! a parsed Mangle AST into an executable plan.
//!
//! # Transformation Stages
//!
//! 1.  **Program Structure**: The raw AST is wrapped in a [`Program`] abstraction
//!     which distinguishes between extensional (data) and intensional (rules) predicates.
//!
//! 2.  **Stratification**: The program is analyzed for dependencies and stratified
//!     to handle negation correctly. This produces a [`StratifiedProgram`], where
//!     predicates are grouped into layers (strata) that can be evaluated sequentially.
//!
//! 3.  **Lowering**: The AST (or stratified program parts) is lowered into the
//!     Intermediate Representation (IR). See [`LoweringContext`].
//!
//! 4.  **Type Checking**: The IR is checked for type consistency and safety.
//!     See [`TypeChecker`].
//!
//! 5.  **Planning**: The Logical IR rules are transformed into Physical Operations
//!     (like nested-loop joins) ready for execution or codegen.
//!     See [`Planner`].

use mangle_ast as ast;

mod type_check;

pub use type_check::TypeChecker;

#[cfg(test)]
mod tests;

mod lowering;

pub use lowering::LoweringContext;

mod rename;

pub use rename::rewrite_unit;

mod planner;

pub use planner::Planner;

mod stratification;

pub use stratification::{Program, StratifiedProgram};

/// A set of predicate symbols, typically used to represent a stratum or a
/// collection of EDB/IDB predicates.
pub type PredicateSet = fxhash::FxHashSet<ast::PredicateIndex>;
