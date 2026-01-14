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

use anyhow::{Context, Result, anyhow};
use mangle_ast as ast;
use mangle_parse::Parser;
use std::collections::HashMap;
use std::io::{BufRead, BufReader, Read};

/// A simple in-memory representation of the loaded data.
pub struct SimpleColumnData<'a> {
    /// Map predicate name -> List of facts (args)
    pub tables: HashMap<String, Vec<Vec<&'a ast::BaseTerm<'a>>>>,
}

pub fn read_from_bytes<'a>(arena: &'a ast::Arena, data: &[u8]) -> Result<SimpleColumnData<'a>> {
    let reader = BufReader::new(data);
    read_simple_column(arena, reader)
}

pub fn read_from_reader<'a, R: Read>(
    arena: &'a ast::Arena,
    reader: R,
) -> Result<SimpleColumnData<'a>> {
    let reader = BufReader::new(reader);
    read_simple_column(arena, reader)
}

fn read_simple_column<'a, R: BufRead>(
    arena: &'a ast::Arena,
    mut reader: R,
) -> Result<SimpleColumnData<'a>> {
    let mut line = String::new();

    // 1. Num Predicates
    reader.read_line(&mut line)?;
    let num_preds: usize = line.trim().parse().context("parsing num_preds")?;
    line.clear();

    struct PredInfo {
        name: String,
        arity: usize,
        num_facts: usize,
    }
    let mut preds = Vec::with_capacity(num_preds);

    // 2. Predicate Headers
    for _ in 0..num_preds {
        reader.read_line(&mut line)?;
        let parts: Vec<&str> = line.split_whitespace().collect();
        if parts.len() != 3 {
            return Err(anyhow!("Invalid predicate header: {}", line));
        }
        let name = parts[0].to_string();
        let arity: usize = parts[1].parse().context("parsing arity")?;
        let num_facts: usize = parts[2].parse().context("parsing num_facts")?;
        preds.push(PredInfo {
            name,
            arity,
            num_facts,
        });
        line.clear();
    }

    let mut tables = HashMap::new();

    // 3. Columns
    for pred in preds {
        if pred.arity == 0 {
            // Flag fact
            if pred.num_facts > 0 {
                tables.insert(pred.name, vec![vec![]]);
            }
            continue;
        }

        // Initialize facts with empty rows
        let mut facts: Vec<Vec<&'a ast::BaseTerm<'a>>> =
            vec![Vec::with_capacity(pred.arity); pred.num_facts];

        // Read columns
        for col_idx in 0..pred.arity {
            for row_idx in 0..pred.num_facts {
                line.clear();
                if reader.read_line(&mut line)? == 0 {
                    return Err(anyhow!(
                        "Unexpected EOF reading column {} row {}",
                        col_idx,
                        row_idx
                    ));
                }

                let text = line.trim();
                let term_str = if text.starts_with('/') {
                    percent_unescape(text)?
                } else {
                    text.to_string()
                };

                // Parse term
                // We need a fresh parser for each term?
                // Parser::new takes full input. `parse_base_term` parses one term.
                // It should stop after the term.
                // Since line contains just the term, it should work.
                let mut parser = Parser::new(arena, term_str.as_bytes(), "simplecolumn");
                parser.next_token()?;
                let term = parser
                    .parse_base_term()
                    .context(format!("parsing term: {}", term_str))?;

                facts[row_idx].push(term);
            }
        }
        tables.insert(pred.name, facts);
    }

    Ok(SimpleColumnData { tables })
}

fn percent_unescape(s: &str) -> Result<String> {
    // Go implementation uses url.QueryUnescape which handles %XX and + -> space.
    // Rust `url` crate? I don't have it in dependencies.
    // I can do basic unescaping manually or assume basic format.
    // The previous implementation used `url` crate in Go.
    // I'll check if I can use a crate or implement simple unescape.
    // Given constraints, I'll implement a simple one handling %XX.

    let mut out = Vec::new();
    let bytes = s.as_bytes();
    let mut i = 0;
    while i < bytes.len() {
        if bytes[i] == b'%' {
            if i + 2 >= bytes.len() {
                return Err(anyhow!("Invalid escape sequence"));
            }
            let hex = &s[i + 1..i + 3];
            let byte = u8::from_str_radix(hex, 16)?;
            out.push(byte);
            i += 3;
        } else if bytes[i] == b'+' {
            out.push(b' '); // url decoding usually treats + as space
            i += 1;
        } else {
            out.push(bytes[i]);
            i += 1;
        }
    }
    Ok(String::from_utf8(out)?)
}

// --- Edge Mode Support ---
#[cfg(feature = "edge")]
pub mod store {
    use super::*;
    use mangle_ast as ast;
    use mangle_interpreter::{MemStore, Store, Value};

    pub struct SimpleColumnStore {
        mem: MemStore,
    }

    impl SimpleColumnStore {
        pub fn from_bytes(arena: &ast::Arena, data: &[u8]) -> Result<Self> {
            let sc_data = read_from_bytes(arena, data)?;
            let mut mem = MemStore::new();

            for (pred, facts) in sc_data.tables {
                mem.create_relation(&pred);
                for row in facts {
                    let tuple: Vec<Value> = row.iter().map(|t| term_to_value(t)).collect();
                    mem.insert(&pred, tuple)?;
                }
            }
            Ok(Self { mem })
        }
    }

    impl Store for SimpleColumnStore {
        fn scan(&self, relation: &str) -> Result<Box<dyn Iterator<Item = Vec<Value>> + '_>> {
            self.mem.scan(relation)
        }

        fn scan_delta(&self, relation: &str) -> Result<Box<dyn Iterator<Item = Vec<Value>> + '_>> {
            self.mem.scan_delta(relation)
        }

        fn scan_next_delta(
            &self,
            relation: &str,
        ) -> Result<Box<dyn Iterator<Item = Vec<Value>> + '_>> {
            self.mem.scan_next_delta(relation)
        }

        fn scan_index(
            &self,
            relation: &str,
            col_idx: usize,
            key: &Value,
        ) -> Result<Box<dyn Iterator<Item = Vec<Value>> + '_>> {
            self.mem.scan_index(relation, col_idx, key)
        }

        fn scan_delta_index(
            &self,
            relation: &str,
            col_idx: usize,
            key: &Value,
        ) -> Result<Box<dyn Iterator<Item = Vec<Value>> + '_>> {
            self.mem.scan_delta_index(relation, col_idx, key)
        }

        fn insert(&mut self, relation: &str, tuple: Vec<Value>) -> Result<bool> {
            self.mem.insert(relation, tuple)
        }

        fn merge_deltas(&mut self) {
            self.mem.merge_deltas()
        }

        fn create_relation(&mut self, relation: &str) {
            self.mem.create_relation(relation)
        }
    }

    fn term_to_value(term: &ast::BaseTerm) -> Value {
        match term {
            ast::BaseTerm::Const(ast::Const::Number(n)) => Value::Number(*n),
            ast::BaseTerm::Const(ast::Const::String(s)) => Value::String(s.to_string()),
            // TODO: Handle other types mapping
            _ => Value::String(format!("{:?}", term)), // Fallback
        }
    }
}

// --- Server Mode Support ---
#[cfg(feature = "server")]
pub mod host {
    use super::*;
    use mangle_ast::Arena;
    use mangle_vm::Host;
    use std::collections::HashMap;
    use std::fs::File;
    use std::path::Path;

    pub struct SimpleColumnHost {
        // We'll reuse the internal logic by loading into a structure similar to MemHost
        // Since we need an Arena to parse, we must own or borrow one.
        // For simplicity, we create one internally?
        // But `read_simple_column` requires `&Arena`.
        // We'll create one in `new`.
        arena: Arena,
        data: HashMap<i32, Vec<Vec<i64>>>, // Only supports i64 for now as per VM limitation
        iters: HashMap<i32, (i32, usize)>,
        next_iter_id: i32,
    }

    impl SimpleColumnHost {
        pub fn new() -> Self {
            Self {
                arena: Arena::new_with_global_interner(),
                data: HashMap::new(),
                iters: HashMap::new(),
                next_iter_id: 1,
            }
        }

        pub fn load_file(&mut self, _rel_name: &str, path: &Path) -> Result<()> {
            let file = File::open(path)?;
            let sc_data = read_simple_column(&self.arena, std::io::BufReader::new(file))?;

            // Assume the file contains the relation we asked for (or others)
            // But sc_data has names from file.
            // We map them to hash IDs.

            for (pred, facts) in sc_data.tables {
                let id = hash_name(&pred);
                let mut numeric_facts = Vec::new();
                for row in facts {
                    let mut tuple = Vec::new();
                    for term in row {
                        if let ast::BaseTerm::Const(ast::Const::Number(n)) = term {
                            tuple.push(*n);
                        } else {
                            // Skip non-numeric for now? Or error?
                            // eprintln!("Skipping non-numeric term in SimpleColumnHost");
                            tuple.push(0);
                        }
                    }
                    numeric_facts.push(tuple);
                }
                self.data.insert(id, numeric_facts);
            }
            Ok(())
        }
    }

    fn hash_name(name: &str) -> i32 {
        let mut hash: u32 = 5381;
        for c in name.bytes() {
            hash = ((hash << 5).wrapping_add(hash)).wrapping_add(c as u32);
        }
        hash as i32
    }

    impl Host for SimpleColumnHost {
        fn scan_start(&mut self, rel_id: i32) -> i32 {
            let id = self.next_iter_id;
            self.next_iter_id += 1;
            self.iters.insert(id, (rel_id, 0));
            id
        }

        fn scan_next(&mut self, iter_id: i32) -> i32 {
            if let Some((rel_id, idx)) = self.iters.get_mut(&iter_id) {
                if let Some(tuples) = self.data.get(rel_id) {
                    if *idx < tuples.len() {
                        let ptr = (iter_id << 16) | (*idx as i32 + 1);
                        *idx += 1;
                        return ptr;
                    }
                }
            }
            0
        }

        fn get_col(&mut self, ptr: i32, col_idx: i32) -> i64 {
            let iter_id = ptr >> 16;
            let tuple_idx = (ptr & 0xFFFF) - 1;

            if let Some((rel_id, _)) = self.iters.get(&iter_id) {
                if let Some(tuples) = self.data.get(rel_id) {
                    if let Some(row) = tuples.get(tuple_idx as usize) {
                        return row.get(col_idx as usize).copied().unwrap_or(0);
                    }
                }
            }
            0
        }

        fn insert(&mut self, rel_id: i32, val: i64) {
            // SimpleColumnHost is primarily read-only EDB, but if we insert, we assume single-column?
            // Or we just append a row with one value.
            self.data.entry(rel_id).or_default().push(vec![val]);
        }

        fn scan_delta_start(&mut self, _rel_id: i32) -> i32 {
            // SimpleColumnHost is for testing simple EDBs, no delta/recursion logic yet.
            // Just return 0 (no deltas).
            0
        }

        fn scan_index_start(&mut self, _rel_id: i32, _col_idx: i32, _val: i64) -> i32 {
            0
        }

        fn scan_aggregate_start(&mut self, _rel_id: i32, _description: Vec<i32>) -> i32 {
            0
        }

        fn merge_deltas(&mut self) -> i32 {
            0
        }

        fn debuglog(&mut self, _val: i64) {}
    }
}
