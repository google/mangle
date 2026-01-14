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

//! # Mangle Engine (Legacy)
//!
//! This crate provides an AST-based interpreter for Mangle.
//! It is considered **legacy** and serves as a reference implementation for the
//! semantics of the language (specifically, the naive bottom-up evaluation).
//!
//! For the new, performance-oriented IR-based execution, see `mangle-interpreter` (Edge)
//! and `mangle-vm` (Server).

use mangle_ast as ast;

use anyhow::Result;
use mangle_analysis as analysis;
use mangle_factstore as factstore;

use analysis::StratifiedProgram;

mod naive;
pub use naive::Naive;

pub trait Engine<'e> {
    fn eval<'p>(
        &'e self,
        store: &'e impl factstore::FactStore<'e>,
        program: &'p StratifiedProgram<'p>,
    ) -> Result<()>;
}
