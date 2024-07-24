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

use std::cell::RefCell;
use std::collections::HashMap;
use std::path::Path;

use crate::ast;
use crate::Bump;
use crate::FactStore;
use crate::ReadOnlyFactStore;
use crate::{anyhow, Result};

#[derive(Clone)]
pub enum TableConfig<'a> {
    InMemory,
    RowFile(&'a Path),
}

pub type TableStoreSchema<'a> = HashMap<&'a ast::PredicateSym<'a>, TableConfig<'a>>;

pub struct TableStoreImpl<'a> {
    schema: &'a TableStoreSchema<'a>,
    bump: Bump,
    tables: RefCell<HashMap<&'a ast::PredicateSym<'a>, Vec<&'a ast::Atom<'a>>>>,
}

impl<'a> ReadOnlyFactStore<'a> for TableStoreImpl<'a> {
    fn contains(&'a self, fact: &ast::Atom) -> Result<bool> {
        if let Some(table) = self.tables.borrow().get(&fact.sym) {
            return Ok(table.contains(&fact));
        }
        Err(anyhow!("unknown predicate {}", fact.sym.name))
    }

    fn get(
        &'a self,
        query: &ast::Atom,
        mut cb: impl FnMut(&'a ast::Atom<'a>) -> anyhow::Result<()>,
    ) -> anyhow::Result<()> {
        if let Some(table) = self.tables.borrow().get(&query.sym) {
            for fact in table {
                if fact.matches(query.args) {
                    cb(fact)?;
                }
            }
            return Ok(());
        }
        Err(anyhow!("unknown predicate {}", query.sym.name))
    }

    fn list_predicates(&'a self, mut cb: impl FnMut(&'a ast::PredicateSym)) {
        for pred in self.tables.borrow().keys() {
            cb(pred);
        }
    }

    fn estimate_fact_count(&self) -> u32 {
        self.tables
            .borrow()
            .values()
            .fold(0, |x, y| x + y.len() as u32)
    }
}

impl<'a> FactStore<'a> for TableStoreImpl<'a> {
    fn add(&'a self, fact: &ast::Atom) -> Result<bool> {
        {
            if let Some(table) = self.tables.borrow().get(&fact.sym) {
                if table.contains(&fact) {
                    return Ok(false);
                }
            } else {
                return Err(anyhow!("no table for {:?}", fact.sym.name));
            }
        }
        // Consider checking that `fact` is, in fact, a fact.
        let fact = ast::copy_atom(&self.bump, fact);
        self.tables
            .borrow_mut()
            .get_mut(&fact.sym)
            .unwrap()
            .push(fact);
        Ok(true)
    }

    fn merge<'other, S>(&'a self, _store: &'other S)
    where
        S: crate::ReadOnlyFactStore<'other>,
    {
        todo!()
    }
}

impl<'a> TableStoreImpl<'a> {
    pub fn new(schema: &'a TableStoreSchema<'a>) -> Self {
        let mut tables = HashMap::new();
        for entry in schema.keys() {
            tables.insert(*entry, vec![]);
        }
        TableStoreImpl {
            schema,
            bump: Bump::new(),
            tables: RefCell::new(tables),
        }
    }
}
