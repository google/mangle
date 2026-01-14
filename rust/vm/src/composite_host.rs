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

use crate::Host;
use std::collections::HashMap;

/// A Host implementation that aggregates multiple sub-hosts.
///
/// It routes relation scans and inserts to the appropriate backend based on
/// explicit routing rules.
pub struct CompositeHost {
    hosts: Vec<Box<dyn Host + Send>>,

    /// rel_id -> index in `hosts`
    routes: HashMap<i32, usize>,

    /// iter_id -> (host_index, real_iter_id)
    active_iters: HashMap<i32, (usize, i32)>,

    next_iter_id: i32,
}

impl Default for CompositeHost {
    fn default() -> Self {
        Self::new()
    }
}

impl CompositeHost {
    pub fn new() -> Self {
        Self {
            hosts: Vec::new(),
            routes: HashMap::new(),
            active_iters: HashMap::new(),
            next_iter_id: 1,
        }
    }

    /// Adds a sub-host and returns its index.
    pub fn add_host(&mut self, host: Box<dyn Host + Send>) -> usize {
        let idx = self.hosts.len();
        self.hosts.push(host);
        idx
    }

    /// Routes a relation to a specific sub-host.
    pub fn route_relation(&mut self, rel_name: &str, host_index: usize) {
        let id = hash_name(rel_name);
        self.routes.insert(id, host_index);
    }
}

fn hash_name(name: &str) -> i32 {
    let mut hash: u32 = 5381;
    for c in name.bytes() {
        hash = ((hash << 5).wrapping_add(hash)).wrapping_add(c as u32);
    }
    hash as i32
}

impl Host for CompositeHost {
    fn scan_start(&mut self, rel_id: i32) -> i32 {
        if let Some(&h_idx) = self.routes.get(&rel_id) {
            let real_id = self.hosts[h_idx].scan_start(rel_id);
            if real_id != 0 {
                let id = self.next_iter_id;
                self.next_iter_id += 1;
                self.active_iters.insert(id, (h_idx, real_id));
                return id;
            }
        }
        0
    }

    fn scan_next(&mut self, iter_id: i32) -> i32 {
        if let Some(&(h_idx, real_id)) = self.active_iters.get(&iter_id) {
            let ptr = self.hosts[h_idx].scan_next(real_id);
            if ptr == 0 {
                return 0;
            }

            // Tag pointer with host index (using top 6 bits)
            // This assumes sub-hosts don't use more than 26 bits for their pointers.
            return ptr | ((h_idx as i32 + 1) << 26);
        }
        0
    }

    fn get_col(&mut self, tuple_ptr: i32, col_idx: i32) -> i64 {
        let h_idx_plus_1 = (tuple_ptr >> 26) & 0x3F;
        if h_idx_plus_1 == 0 {
            return 0;
        }

        let h_idx = (h_idx_plus_1 - 1) as usize;
        let real_ptr = tuple_ptr & !(0x3F << 26);

        if h_idx < self.hosts.len() {
            return self.hosts[h_idx].get_col(real_ptr, col_idx);
        }
        0
    }

    fn insert(&mut self, rel_id: i32, val: i64) {
        if let Some(&h_idx) = self.routes.get(&rel_id) {
            self.hosts[h_idx].insert(rel_id, val);
        }
    }

    fn scan_delta_start(&mut self, rel_id: i32) -> i32 {
        if let Some(&h_idx) = self.routes.get(&rel_id) {
            let real_id = self.hosts[h_idx].scan_delta_start(rel_id);
            if real_id != 0 {
                let id = self.next_iter_id;
                self.next_iter_id += 1;
                self.active_iters.insert(id, (h_idx, real_id));
                return id;
            }
        }
        0
    }

    fn scan_index_start(&mut self, rel_id: i32, col_idx: i32, val: i64) -> i32 {
        if let Some(&h_idx) = self.routes.get(&rel_id) {
            let real_id = self.hosts[h_idx].scan_index_start(rel_id, col_idx, val);
            if real_id != 0 {
                let id = self.next_iter_id;
                self.next_iter_id += 1;
                self.active_iters.insert(id, (h_idx, real_id));
                return id;
            }
        }
        0
    }

    fn scan_aggregate_start(&mut self, rel_id: i32, description: Vec<i32>) -> i32 {
        if let Some(&h_idx) = self.routes.get(&rel_id) {
            let real_id = self.hosts[h_idx].scan_aggregate_start(rel_id, description);
            if real_id != 0 {
                let id = self.next_iter_id;
                self.next_iter_id += 1;
                self.active_iters.insert(id, (h_idx, real_id));
                return id;
            }
        }
        0
    }

    fn merge_deltas(&mut self) -> i32 {
        let mut changes = 0;
        for host in &mut self.hosts {
            changes |= host.merge_deltas();
        }
        changes
    }

    fn debuglog(&mut self, _val: i64) {}
}
