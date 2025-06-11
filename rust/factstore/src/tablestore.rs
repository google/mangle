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

use fxhash::FxHashMap;
use std::cell::RefCell;
use std::path::Path;

use crate::ast;
use crate::FactStore;
use crate::ReadOnlyFactStore;
use crate::{anyhow, Result};
use ast::Arena;

#[derive(Clone)]
pub enum TableConfig<'a> {
    InMemory,
    RowFile(&'a Path),
}

pub type TableStoreSchema<'a> = FxHashMap<ast::PredicateIndex, TableConfig<'a>>;

pub struct TableStoreImpl<'a> {
    schema: &'a TableStoreSchema<'a>,
    arena: Arena,
    tables: RefCell<FxHashMap<ast::PredicateIndex, Vec<&'a ast::Atom<'a>>>>,
}

impl<'a> ReadOnlyFactStore<'a> for TableStoreImpl<'a> {
    fn contains<'src>(&'a self, _src: &'src Arena, fact: &'src ast::Atom) -> Result<bool> {
        if let Some(table) = self.tables.borrow().get(&fact.sym) {
            return Ok(table.contains(&fact));
        }
        Err(anyhow!("unknown predicate index {}", fact.sym))
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
        Err(anyhow!("unknown predicate index {}", query.sym))
    }

    fn predicates(&'a self) -> Vec<ast::PredicateIndex> {
        self.tables.borrow().keys().copied().collect()
    }

    fn estimate_fact_count(&self) -> u32 {
        self.tables
            .borrow()
            .values()
            .fold(0, |x, y| x + y.len() as u32)
    }
}

impl<'a> FactStore<'a> for TableStoreImpl<'a> {
    fn add<'src>(&'a self, src: &'src Arena, fact: &'src ast::Atom) -> Result<bool> {
        let mut tables = self.tables.borrow_mut();
        let table = tables.get_mut(&fact.sym);
        match table {
            None => Err(anyhow!("no table for {:?}", fact.sym)),
            Some(table) => {
                Ok(if table.contains(&fact) {
                    false
                } else {
                    // We trust that `fact` is, in fact, a fact.
                    let fact = self.arena.copy_atom(src, fact);
                    table.push(fact);
                    true
                })
            }
        }
    }

    fn merge<'src, S>(&'a self, _src: &'src Arena, _store: &'src S)
    where
        S: crate::ReadOnlyFactStore<'src>,
    {
        todo!()
    }
}

impl<'a> TableStoreImpl<'a> {
    pub fn new(schema: &'a TableStoreSchema<'a>) -> Self {
        let mut tables = FxHashMap::default();
        for entry in schema.keys() {
            tables.insert(*entry, vec![]);
        }
        TableStoreImpl {
            schema,
            arena: Arena::new_global(),
            tables: RefCell::new(tables),
        }
    }
}
