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

use anyhow::{Result, anyhow};
use fxhash::FxHashSet;
use mangle_ir::physical::{self, Aggregate, CmpOp, Condition, DataSource, Expr, Op, Operand};
use mangle_ir::{Inst, InstId, Ir, NameId};

pub struct Planner<'a> {
    ir: &'a mut Ir,
    delta_pred: Option<NameId>,
}

impl<'a> Planner<'a> {
    pub fn new(ir: &'a mut Ir) -> Self {
        Self {
            ir,
            delta_pred: None,
        }
    }

    pub fn with_delta(mut self, delta_pred: NameId) -> Self {
        self.delta_pred = Some(delta_pred);
        self
    }

    pub fn plan_rule(mut self, rule_id: InstId) -> Result<Op> {
        let (head, premises, transform) = match self.ir.get(rule_id) {
            Inst::Rule {
                head,
                premises,
                transform,
            } => (*head, premises.clone(), transform.clone()),
            _ => return Err(anyhow!("Not a rule")),
        };

        // Split transforms into blocks by 'do' statements
        let blocks = self.split_transforms(transform);
        let num_blocks = blocks.len();

        let mut ops = Vec::new();
        let mut current_source: Option<(NameId, Vec<NameId>)> = None;
        let mut bound_vars = FxHashSet::default();

        for (i, block) in blocks.into_iter().enumerate() {
            let is_last = i == num_blocks - 1;

            if i == 0 {
                // Block 0: Premises + Lets
                if is_last {
                    // Only one block, no aggregations
                    let op = self.plan_join_sequence(
                        premises.clone(),
                        &mut bound_vars,
                        |planner, vars| {
                            planner.plan_transforms_sequence(&block, vars, |p, v| {
                                p.plan_head_insert(head, v)
                            })
                        },
                    )?;
                    ops.push(op);
                } else {
                    // Materialize to temp
                    let temp_rel = self.fresh_var("temp_grp");
                    let mut capture_vars: Vec<NameId> = Vec::new(); // Will be populated by continuation

                    let op = self.plan_join_sequence(
                        premises.clone(),
                        &mut bound_vars,
                        |planner, vars| {
                            planner.plan_transforms_sequence(&block, vars, |_, v| {
                                let mut sorted_vars: Vec<NameId> = v.iter().cloned().collect();
                                sorted_vars.sort();
                                capture_vars = sorted_vars.clone();
                                let args =
                                    sorted_vars.iter().map(|&var| Operand::Var(var)).collect();
                                Ok(Op::Insert {
                                    relation: temp_rel,
                                    args,
                                })
                            })
                        },
                    )?;
                    ops.push(op);
                    current_source = Some((temp_rel, capture_vars));
                }
            } else {
                // Block i > 0: Starts with 'do'
                let (src_rel, src_vars) = current_source.take().expect("No source for aggregation");

                if is_last {
                    let op = self.plan_block_k(src_rel, src_vars, &block, |p, v| {
                        p.plan_head_insert(head, v)
                    })?;
                    ops.push(op);
                } else {
                    let next_temp = self.fresh_var("temp_grp");
                    let mut next_vars: Vec<NameId> = Vec::new();

                    let op = self.plan_block_k(src_rel, src_vars, &block, |_, v| {
                        let mut sorted_vars: Vec<NameId> = v.iter().cloned().collect();
                        sorted_vars.sort();
                        next_vars = sorted_vars.clone();
                        let args = sorted_vars.iter().map(|&var| Operand::Var(var)).collect();
                        Ok(Op::Insert {
                            relation: next_temp,
                            args,
                        })
                    })?;
                    ops.push(op);
                    current_source = Some((next_temp, next_vars));
                }
            }
        }

        if ops.len() == 1 {
            Ok(ops.remove(0))
        } else {
            Ok(Op::Seq(ops))
        }
    }

    fn split_transforms(&self, transforms: Vec<InstId>) -> Vec<Vec<InstId>> {
        let mut blocks = Vec::new();
        let mut current = Vec::new();
        for t in transforms {
            let inst = self.ir.get(t);
            if let Inst::Transform { var: None, .. } = inst {
                blocks.push(current);
                current = Vec::new();
            }
            current.push(t);
        }
        blocks.push(current);
        blocks
    }

    fn plan_block_k<F>(
        &mut self,
        source_rel: NameId,
        source_vars: Vec<NameId>,
        block: &[InstId],
        continuation: F,
    ) -> Result<Op>
    where
        F: FnOnce(&mut Self, &mut FxHashSet<NameId>) -> Result<Op>,
    {
        let do_stmt = block[0];
        let rest = &block[1..];

        let keys_insts = self.get_transform_app_args(do_stmt)?;
        let mut keys = Vec::new();
        for k in keys_insts {
            if let Inst::Var(v) = self.ir.get(k) {
                keys.push(*v);
            } else {
                return Err(anyhow!("GroupBy keys must be variables"));
            }
        }

        let mut aggregates = Vec::new();
        let mut lets = Vec::new();
        for &t in rest {
            if let Some(agg) = self.try_parse_aggregate(t)? {
                aggregates.push(agg);
            } else {
                lets.push(t);
            }
        }

        let mut inner_vars = FxHashSet::default();
        for &k in &keys {
            inner_vars.insert(k);
        }
        for agg in &aggregates {
            inner_vars.insert(agg.var);
        }

        let body = self.plan_transforms_sequence(&lets, &mut inner_vars, continuation)?;

        Ok(Op::GroupBy {
            source: source_rel,
            vars: source_vars,
            keys,
            aggregates,
            body: Box::new(body),
        })
    }

    fn plan_transforms_sequence<F>(
        &mut self,
        transforms: &[InstId],
        bound_vars: &mut FxHashSet<NameId>,
        continuation: F,
    ) -> Result<Op>
    where
        F: FnOnce(&mut Self, &mut FxHashSet<NameId>) -> Result<Op>,
    {
        if transforms.is_empty() {
            return continuation(self, bound_vars);
        }

        let t_id = transforms[0];
        let rest = &transforms[1..];

        let inst = self.ir.get(t_id).clone();
        if let Inst::Transform {
            var: Some(var),
            app,
        } = inst
        {
            self.inst_to_expr(app, |planner, expr| {
                bound_vars.insert(var);
                let body = planner.plan_transforms_sequence(rest, bound_vars, continuation)?;
                Ok(Op::Let {
                    var,
                    expr,
                    body: Box::new(body),
                })
            })
        } else {
            // Should not happen if split_transforms is correct
            Err(anyhow!("Unexpected transform in sequence"))
        }
    }

    fn fresh_var(&mut self, prefix: &str) -> NameId {
        let name = format!("${}_{}", prefix, self.ir.insts.len());
        self.ir.intern_name(name)
    }

    fn plan_join_sequence<F>(
        &mut self,
        mut premises: Vec<InstId>,
        bound_vars: &mut FxHashSet<NameId>,
        continuation: F,
    ) -> Result<Op>
    where
        F: FnOnce(&mut Self, &mut FxHashSet<NameId>) -> Result<Op>,
    {
        if premises.is_empty() {
            return continuation(self, bound_vars);
        }

        let current_premise = premises.remove(0);
        let inst = self.ir.get(current_premise).clone();

        match inst {
            Inst::Atom { predicate, args } => {
                let mut scan_vars = Vec::new();
                let mut new_bindings = Vec::new();

                // Look for a potential index lookup
                let mut index_lookup: Option<(usize, Operand)> = None;

                for (i, arg) in args.iter().enumerate() {
                    let arg_inst = self.ir.get(*arg).clone();
                    match arg_inst {
                        Inst::Var(v) if bound_vars.contains(&v) => {
                            if index_lookup.is_none() {
                                index_lookup = Some((i, Operand::Var(v)));
                            }
                        }
                        Inst::Number(n) => {
                            if index_lookup.is_none() {
                                index_lookup =
                                    Some((i, Operand::Const(physical::Constant::Number(n))));
                            }
                        }
                        Inst::String(s) => {
                            if index_lookup.is_none() {
                                index_lookup =
                                    Some((i, Operand::Const(physical::Constant::String(s))));
                            }
                        }
                        _ => {}
                    }
                }

                for arg in &args {
                    if let Inst::Var(v) = self.ir.get(*arg)
                        && !bound_vars.contains(v)
                    {
                        scan_vars.push(*v);
                        new_bindings.push(*v);
                        continue;
                    }
                    let tmp = self.fresh_var("scan");
                    scan_vars.push(tmp);
                    new_bindings.push(tmp);
                }

                for v in &new_bindings {
                    bound_vars.insert(*v);
                }

                let body = self.plan_join_sequence(premises, bound_vars, continuation)?;
                let wrapped_body = self.apply_constraints(&args, &scan_vars, body)?;

                let source = if let Some((col_idx, key)) = index_lookup {
                    DataSource::IndexLookup {
                        relation: predicate,
                        col_idx,
                        key,
                        vars: scan_vars,
                    }
                } else if Some(predicate) == self.delta_pred {
                    DataSource::ScanDelta {
                        relation: predicate,
                        vars: scan_vars,
                    }
                } else {
                    DataSource::Scan {
                        relation: predicate,
                        vars: scan_vars,
                    }
                };
                Ok(Op::Iterate {
                    source,
                    body: Box::new(wrapped_body),
                })
            }
            Inst::Eq(l, r) => {
                let body = self.plan_join_sequence(premises, bound_vars, continuation)?;
                self.wrap_eq_check(l, r, body)
            }
            _ => Err(anyhow!("Unsupported premise type")),
        }
    }

    fn apply_constraints(
        &mut self,
        args: &[InstId],
        scan_vars: &[NameId],
        mut body: Op,
    ) -> Result<Op> {
        for (i, arg) in args.iter().enumerate().rev() {
            let scan_var = scan_vars[i];
            let arg_inst = self.ir.get(*arg).clone();
            match arg_inst {
                Inst::Var(v) => {
                    if v == scan_var {
                        continue;
                    }
                    body = Op::Filter {
                        cond: Condition::Cmp {
                            op: CmpOp::Eq,
                            left: Operand::Var(scan_var),
                            right: Operand::Var(v),
                        },
                        body: Box::new(body),
                    };
                }
                _ => {
                    body = self.wrap_eval_check(*arg, Operand::Var(scan_var), body)?;
                }
            }
        }
        Ok(body)
    }

    fn wrap_eq_check(&mut self, l: InstId, r: InstId, body: Op) -> Result<Op> {
        self.with_eval(l, |this, op_l| {
            this.with_eval(r, |_this, op_r| {
                Ok(Op::Filter {
                    cond: Condition::Cmp {
                        op: CmpOp::Eq,
                        left: op_l,
                        right: op_r,
                    },
                    body: Box::new(body),
                })
            })
        })
    }

    fn wrap_eval_check(&mut self, inst: InstId, target: Operand, body: Op) -> Result<Op> {
        self.with_eval(inst, |_this, op| {
            Ok(Op::Filter {
                cond: Condition::Cmp {
                    op: CmpOp::Eq,
                    left: target,
                    right: op,
                },
                body: Box::new(body),
            })
        })
    }

    fn with_eval<F>(&mut self, inst: InstId, f: F) -> Result<Op>
    where
        F: FnOnce(&mut Self, Operand) -> Result<Op>,
    {
        let i = self.ir.get(inst).clone();
        match i {
            Inst::Var(v) => f(self, Operand::Var(v)),
            Inst::String(s) => f(self, Operand::Const(physical::Constant::String(s))),
            Inst::Number(n) => f(self, Operand::Const(physical::Constant::Number(n))),
            Inst::ApplyFn { function, args } => self.with_eval_args(
                &args,
                0,
                Vec::new(),
                Box::new(|this, ops| {
                    let tmp = this.fresh_var("call");
                    let inner = f(this, Operand::Var(tmp))?;
                    Ok(Op::Let {
                        var: tmp,
                        expr: Expr::Call {
                            function,
                            args: ops,
                        },
                        body: Box::new(inner),
                    })
                }),
            ),
            _ => Err(anyhow!("Unsupported expression in evaluation")),
        }
    }

    fn inst_to_expr<F>(&mut self, inst: InstId, f: F) -> Result<Op>
    where
        F: FnOnce(&mut Self, Expr) -> Result<Op>,
    {
        let i = self.ir.get(inst).clone();
        match i {
            Inst::ApplyFn { function, args } => self.with_eval_args(
                &args,
                0,
                Vec::new(),
                Box::new(|this, ops| {
                    f(
                        this,
                        Expr::Call {
                            function,
                            args: ops,
                        },
                    )
                }),
            ),
            _ => self.with_eval(inst, |this, op| f(this, Expr::Value(op))),
        }
    }

    fn with_eval_args(
        &mut self,
        args: &[InstId],
        index: usize,
        mut acc: Vec<Operand>,
        f: Box<dyn FnOnce(&mut Self, Vec<Operand>) -> Result<Op> + '_>,
    ) -> Result<Op> {
        if index >= args.len() {
            return f(self, acc);
        }
        self.with_eval(args[index], |this, op| {
            acc.push(op);
            this.with_eval_args(args, index + 1, acc, f)
        })
    }

    fn plan_head_insert(
        &mut self,
        head: InstId,
        _bound_vars: &mut FxHashSet<NameId>,
    ) -> Result<Op> {
        let inst = self.ir.get(head).clone();
        if let Inst::Atom { predicate, args } = inst {
            self.with_eval_args(
                &args,
                0,
                Vec::new(),
                Box::new(|_this, ops| {
                    Ok(Op::Insert {
                        relation: predicate,
                        args: ops,
                    })
                }),
            )
        } else {
            Err(anyhow!("Head must be an atom"))
        }
    }

    fn get_transform_app_args(&self, t_id: InstId) -> Result<Vec<InstId>> {
        if let Inst::Transform { app, .. } = self.ir.get(t_id)
            && let Inst::ApplyFn { args, .. } = self.ir.get(*app)
        {
            return Ok(args.clone());
        }
        Err(anyhow!("Invalid transform structure"))
    }

    fn try_parse_aggregate(&mut self, t_id: InstId) -> Result<Option<Aggregate>> {
        let inst = self.ir.get(t_id).clone();
        if let Inst::Transform {
            var: Some(var),
            app,
        } = inst
            && let Inst::ApplyFn { function, args } = self.ir.get(app).clone()
        {
            let func_name = self.ir.resolve_name(function);
            if matches!(
                func_name,
                "fn:sum" | "fn:count" | "fn:max" | "fn:min" | "fn:collect"
            ) {
                let mut op_args = Vec::new();
                for arg in args {
                    let arg_inst = self.ir.get(arg).clone();
                    match arg_inst {
                        Inst::Var(v) => op_args.push(Operand::Var(v)),
                        Inst::Number(n) => {
                            op_args.push(Operand::Const(physical::Constant::Number(n)))
                        }
                        _ => {
                            return Err(anyhow!(
                                "Complex expressions in aggregates not supported yet"
                            ));
                        }
                    }
                }
                return Ok(Some(Aggregate {
                    var,
                    func: function,
                    args: op_args,
                }));
            }
        }
        Ok(None)
    }
}
