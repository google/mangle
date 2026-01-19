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

use anyhow::{Result, anyhow, bail};
use ast::{Arena, Constraints};
use mangle_ast as ast;
use std::io;

mod error;
mod quote;
mod scan;
mod token;

pub use error::{ErrorContext, ParseError};
use token::Token;

pub struct Parser<'arena, R>
where
    R: io::Read,
{
    sc: scan::Scanner<R>,
    token: crate::token::Token,
    arena: &'arena Arena,
}

fn package_sym(arena: &Arena) -> ast::PredicateIndex {
    arena.predicate_sym("Package", Some(0))
}

fn name_sym(arena: &Arena) -> ast::PredicateIndex {
    arena.predicate_sym("name", Some(1))
}

fn use_sym(arena: &Arena) -> ast::PredicateIndex {
    arena.predicate_sym("Use", Some(0))
}

fn lt_sym(arena: &Arena) -> ast::PredicateIndex {
    arena.predicate_sym(":lt", Some(2))
}

fn le_sym(arena: &Arena) -> ast::PredicateIndex {
    arena.predicate_sym(":le", Some(2))
}

fn fn_list_sym(arena: &Arena) -> ast::FunctionIndex {
    arena.function_sym("fn:list", None)
}

fn fn_map_sym(arena: &Arena) -> ast::FunctionIndex {
    arena.function_sym("fn:map", None)
}

fn fn_struct_sym(arena: &Arena) -> ast::FunctionIndex {
    arena.function_sym("fn:struct", None)
}

fn fn_list_type_sym(arena: &Arena) -> ast::FunctionIndex {
    arena.function_sym("fn:List", None)
}

fn fn_option_type_sym(arena: &Arena) -> ast::FunctionIndex {
    arena.function_sym("fn:Option", None)
}

fn empty_package_decl(arena: &Arena) -> ast::Decl<'_> {
    ast::Decl {
        atom: arena.alloc(ast::Atom {
            sym: package_sym(arena),
            args: &[],
        }),
        descr: arena.alloc_slice_copy(&[arena.alloc(ast::Atom {
            sym: name_sym(arena),
            args: arena
                .alloc_slice_copy(&[arena.alloc(ast::BaseTerm::Const(ast::Const::String("")))]),
        })]),
        bounds: None,
        constraints: None,
    }
}

macro_rules! alloc {
    ($self:expr, $e:expr) => {
        &*$self.arena.alloc($e)
    };
}

macro_rules! alloc_str {
    ($self:expr, $e:expr) => {
        &*$self.arena.alloc_str($e)
    };
}

macro_rules! alloc_slice {
    ($self:expr, $e:expr) => {
        &*$self.arena.alloc_slice_copy($e)
    };
}

impl<'arena, R> Parser<'arena, R>
where
    R: io::Read,
{
    pub fn new<P: ToString>(arena: &'arena Arena, reader: R, path: P) -> Self
    where
        R: io::Read,
    {
        Self {
            sc: scan::Scanner::new(reader, path),
            token: token::Token::Illegal,
            arena,
        }
    }

    pub fn next_token(&mut self) -> Result<()> {
        self.token = self.sc.next_token()?;
        Ok(())
    }

    // Check that token is the expected one and advance.
    fn expect(&mut self, expected: Token) -> Result<()> {
        if expected != self.token {
            let error = ParseError::Unexpected(
                self.sc.get_error_context(),
                expected.clone(),
                self.token.clone(),
            );
            return Err(anyhow!(error));
        }
        self.next_token()
    }

    pub fn parse_unit(&mut self) -> Result<&'arena ast::Unit<'arena>> {
        let package = if matches!(self.token.clone(), Token::Package) {
            self.parse_package_decl()?
        } else {
            self.arena.alloc(empty_package_decl(self.arena))
        };
        let mut decls = vec![package];
        while let Token::Use = self.token {
            decls.push(self.parse_use_decl()?);
        }
        let mut clauses = vec![];
        loop {
            match self.token {
                Token::Eof => break,
                Token::Decl => decls.push(self.parse_decl()?),
                _ => clauses.push(self.parse_clause()?),
            }
        }
        let decls: &'arena [&'arena ast::Decl<'arena>] = self.arena.alloc_slice_copy(&decls);
        let clauses: &'arena [&'arena ast::Clause<'arena>] = self.arena.alloc_slice_copy(&clauses);
        let unit: &'arena ast::Unit<'arena> = &*self.arena.alloc(ast::Unit { clauses, decls });
        Ok(unit)
    }

    /// package_decl ::= `package` name (`[` `]`)? `!`
    pub fn parse_package_decl(&mut self) -> Result<&'arena ast::Decl<'arena>> {
        self.expect(Token::Package)?;
        let package_name: &'arena str = if let Token::Ident { name } = &self.token {
            self.arena.alloc_str(name.as_str())
        } else {
            bail!("expected identifer got {}", self.token);
        };

        let name_atom: &'arena ast::Atom<'arena> = self.arena.alloc(ast::Atom {
            sym: name_sym(self.arena),
            args: self.arena.alloc_slice_copy(&[self
                .arena
                .alloc(ast::BaseTerm::Const(ast::Const::String(package_name)))]),
        });
        let mut descr_atoms: Vec<&'arena ast::Atom<'arena>> = vec![name_atom];
        self.next_token()?;
        if Token::LBracket == self.token {
            self.parse_bracket_atoms(&mut descr_atoms)?;
        }
        let descr = alloc_slice!(self, &descr_atoms);

        self.expect(Token::Bang)?;

        let package_atom = alloc!(
            self,
            ast::Atom {
                sym: package_sym(self.arena),
                args: &[]
            }
        );

        //let descr_atoms = ;
        let decl: &'arena ast::Decl = alloc!(
            self,
            ast::Decl {
                atom: package_atom,
                bounds: None,
                descr,
                constraints: None
            }
        );
        Ok(decl)
    }

    fn parse_use_decl(&mut self) -> Result<&'arena ast::Decl<'arena>> {
        self.expect(Token::Use)?;
        let use_atom = alloc!(
            self,
            ast::Atom {
                sym: use_sym(self.arena),
                args: &[]
            }
        );

        let name = match &self.token {
            Token::Ident { name } => name.as_str(),
            _ => bail!("parse_use_decl: expected identifer got {}", self.token),
        };

        let name: &'arena str = alloc_str!(self, name);
        let name = alloc!(self, ast::BaseTerm::Const(ast::Const::String(name)));
        let args = alloc_slice!(self, &[name]);

        let mut descr_atoms: Vec<&ast::Atom> = vec![self.arena.alloc(ast::Atom {
            sym: name_sym(self.arena),
            args,
        })];
        self.next_token()?;
        if Token::LBracket == self.token {
            self.parse_bracket_atoms(&mut descr_atoms)?;
        }
        self.expect(Token::Bang)?;

        let descr_atoms = alloc_slice!(self, &descr_atoms);
        Ok(alloc!(
            self,
            ast::Decl {
                atom: use_atom,
                descr: descr_atoms,
                bounds: None,
                constraints: None
            }
        ))
    }

    fn parse_decl(&mut self) -> Result<&'arena ast::Decl<'arena>> {
        self.expect(Token::Decl)?;
        let atom = self.parse_atom()?;
        let mut descr_atoms = vec![];
        if Token::Descr == self.token {
            self.next_token()?;
            self.parse_bracket_atoms(&mut descr_atoms)?;
        }
        let mut bound_decls = vec![];
        loop {
            if Token::Bound != self.token {
                break;
            }
            bound_decls.push(self.parse_bounds_decl()?);
        }
        let bounds = if bound_decls.is_empty() {
            None
        } else {
            Some(alloc_slice!(self, &bound_decls))
        };
        let constraints = if Token::Inclusion == self.token {
            Some(self.parse_inclusion_constraint()?)
        } else {
            None
        };
        self.expect(Token::Dot)?;
        Ok(alloc!(
            self,
            ast::Decl {
                atom,
                descr: alloc_slice!(self, &descr_atoms),
                bounds,
                constraints
            }
        ))
    }

    /// bound_decl ::= `bound` `[` base_term {`,` base_term} `]`
    fn parse_bounds_decl(&mut self) -> Result<&'arena ast::BoundDecl<'arena>> {
        self.expect(Token::Bound)?;
        self.expect(Token::LBracket)?;
        let mut base_terms = vec![];
        self.parse_base_terms(&mut base_terms)?;
        self.expect(Token::RBracket)?;
        let base_terms = alloc_slice!(self, &base_terms);
        let bound_decl = alloc!(self, ast::BoundDecl { base_terms });
        Ok(bound_decl)
    }

    fn parse_inclusion_constraint(&mut self) -> Result<&'arena ast::Constraints<'arena>> {
        self.expect(Token::Inclusion)?;
        let mut consequences = vec![];
        self.parse_bracket_atoms(&mut consequences)?;
        let consequences = alloc_slice!(self, &consequences);
        Ok(alloc!(
            self,
            Constraints {
                consequences,
                alternatives: &[]
            }
        ))
    }

    pub fn parse_clause(&mut self) -> Result<&'arena ast::Clause<'arena>> {
        let head = self.parse_atom()?;
        let mut premises = vec![];
        let mut transform = vec![];
        match self.token {
            Token::ColonDash | Token::LongLeftDoubleArrow => {
                self.next_token()?;
                self.parse_terms(&mut premises)?;
                if let Token::PipeGt = self.token {
                    self.next_token()?;
                    self.parse_transforms(&mut transform)?;
                }
            }
            _ => {}
        }
        self.expect(Token::Dot)?;
        let premises = alloc_slice!(self, &premises);
        let transform = alloc_slice!(self, &transform);
        Ok(alloc!(
            self,
            ast::Clause {
                head,
                premises,
                transform
            }
        ))
    }

    /// terms ::= term { , term }
    fn parse_terms(&mut self, terms: &mut Vec<&'arena ast::Term<'arena>>) -> Result<()> {
        terms.push(self.parse_term()?);
        loop {
            if Token::Comma != self.token {
                return Ok(());
            }
            self.next_token()?;
            terms.push(self.parse_term()?);
        }
    }

    pub fn parse_term(&mut self) -> Result<&'arena ast::Term<'arena>> {
        match &self.token {
            Token::Bang => {
                self.next_token()?;
                let atom = self.parse_atom()?;
                Ok(alloc!(self, ast::Term::NegAtom(atom)))
            }
            t if base_term_start(t) => {
                let left_base_term = self.parse_base_term()?;
                let op = self.token.clone();
                match op {
                    Token::Eq | Token::BangEq | Token::Lt | Token::Le => self.next_token()?,
                    _ => bail!("parse_terms: expected `=` or `!=` got {}", self.token),
                };
                let right_base_term = self.parse_base_term()?;
                let term = match op {
                    Token::Eq => ast::Term::Eq(left_base_term, right_base_term),
                    Token::BangEq => ast::Term::Ineq(left_base_term, right_base_term),
                    Token::Lt => ast::Term::Atom(alloc!(
                        self,
                        ast::Atom {
                            sym: lt_sym(self.arena),
                            args: alloc_slice!(self, &[left_base_term, right_base_term]),
                        }
                    )),
                    Token::Le => ast::Term::Atom(self.arena.alloc(ast::Atom {
                        sym: le_sym(self.arena),
                        args: alloc_slice!(self, &[left_base_term, right_base_term]),
                    })),
                    _ => unreachable!(),
                };
                Ok(alloc!(self, term))
            }
            Token::Ident { .. } => {
                let atom = self.parse_atom()?;
                Ok(alloc!(self, ast::Term::Atom(atom)))
            }
            _ => bail!("parse_term: unexpected token {:?}", self.token),
        }
    }

    // bracket_atoms ::= `[` [ atom {`,` atom } ] `]`
    fn parse_bracket_atoms(&mut self, atoms: &mut Vec<&'arena ast::Atom<'arena>>) -> Result<()> {
        self.expect(Token::LBracket)?;
        self.parse_atoms(atoms)?;
        self.expect(Token::RBracket)?;
        Ok(())
    }

    // `atoms ::= [ atom {`,` atom } ]
    fn parse_atoms(&mut self, atoms: &mut Vec<&'arena ast::Atom<'arena>>) -> Result<()> {
        if let Token::Ident { .. } = self.token {
            atoms.push(self.parse_atom()?);
            loop {
                if Token::Comma != self.token {
                    break;
                }
                self.next_token()?;
                let atom = self.parse_atom()?;
                atoms.push(atom);
            }
        }
        Ok(())
    }

    // atom ::= name `(` args `)`
    fn parse_atom(&mut self) -> Result<&'arena ast::Atom<'arena>> {
        let name = match &self.token {
            Token::Ident { name } => name.as_str(),
            _ => bail!("parse_atom: expected identifer got {}", self.token),
        };
        let name = self.arena.alloc_str(name);

        self.next_token()?;
        self.expect(Token::LParen)?;
        let mut args = vec![];
        if Token::RParen != self.token {
            self.parse_base_terms(&mut args)?;
        }
        self.expect(Token::RParen)?;
        let args = alloc_slice!(self, &args);
        Ok(alloc!(
            self,
            ast::Atom {
                sym: self.arena.predicate_sym(name, None),
                args
            }
        ))
    }

    fn parse_transforms(
        &mut self,
        transforms: &mut Vec<&'arena ast::TransformStmt<'arena>>,
    ) -> Result<()> {
        if Token::Do == self.token {
            self.next_token()?;
            let expr = self.parse_base_term()?;
            transforms.push(alloc!(
                self,
                ast::TransformStmt {
                    var: None,
                    app: expr
                }
            ));
            self.expect(Token::Semi)?;
        }
        loop {
            if Token::Let != self.token {
                break;
            }
            self.next_token()?;
            if let Token::Ident { name } = &self.token {
                let name = alloc_str!(self, name.as_str());
                self.next_token()?;
                self.expect(Token::Eq)?;
                let expr = self.parse_base_term()?;
                transforms.push(alloc!(
                    self,
                    ast::TransformStmt {
                        var: Some(name),
                        app: expr
                    }
                ))
            }
            if let Token::Dot = self.token {
                break;
            }
            self.expect(Token::Semi)?;
        }
        Ok(())
    }

    // base_term ::= var
    //             | fun`(`[base_term {',' base_term}`)`
    //             | string_constant
    //             | bytes_constant
    //             | number_constant
    //             | float_constant
    //             | name_constant
    pub fn parse_base_term(&mut self) -> Result<&'arena ast::BaseTerm<'arena>> {
        match &self.token {
            Token::LBracket => return self.parse_list_or_map(),
            Token::LBrace => return self.parse_struct(),
            _ => {}
        }

        let mut is_type = false;
        let mut base_term = match &self.token {
            Token::Ident { name } if is_variable(name) => {
                ast::BaseTerm::Variable(self.arena.variable_sym(name))
            }
            Token::Ident { name } if is_fn(name) => {
                let name = self.arena.alloc_str(name);
                // Arguments parsed below.
                ast::BaseTerm::ApplyFn(self.arena.function_sym(name, None), &[])
            }
            Token::DotIdent { name } => {
                let name = self.arena.alloc_str(name);
                is_type = true;
                // Arguments parsed below.
                ast::BaseTerm::ApplyFn(self.arena.function_sym(name, None), &[])
            }
            Token::String { decoded } => {
                let value = self.arena.alloc_str(decoded.as_str());
                ast::BaseTerm::Const(ast::Const::String(value))
            }
            Token::Bytes { decoded } => {
                let value = self.arena.alloc_slice_copy(decoded);
                ast::BaseTerm::Const(ast::Const::Bytes(value))
            }
            Token::Int { decoded } => ast::BaseTerm::Const(ast::Const::Number(*decoded)),
            Token::Float { decoded } => ast::BaseTerm::Const(ast::Const::Float(*decoded)),
            Token::Name { name } => {
                let name = self.arena.intern(name);
                ast::BaseTerm::Const(ast::Const::Name(name))
            }
            _ => bail!("parse_base_term: unexpected token {:?}", self.token),
        };
        self.next_token()?;
        if let ast::BaseTerm::ApplyFn(fn_sym, _) = base_term {
            let mut fn_args = vec![];
            if is_type {
                self.parse_langle_base_terms(&mut fn_args)?;
            } else {
                self.parse_paren_base_terms(&mut fn_args)?;
            }
            let fn_args = self.arena.alloc_slice_copy(&fn_args);
            base_term = ast::BaseTerm::ApplyFn(fn_sym, fn_args);
        }
        let base_term = alloc!(self, base_term);
        Ok(base_term)
    }

    fn parse_list_or_map(&mut self) -> Result<&'arena ast::BaseTerm<'arena>> {
        self.expect(Token::LBracket)?;
        if Token::RBracket == self.token {
            self.next_token()?;
            return Ok(alloc!(
                self,
                ast::BaseTerm::ApplyFn(fn_list_sym(self.arena), &[])
            ));
        }
        let first = self.parse_base_term()?;
        let expr = if Token::Colon != self.token {
            self.expect(Token::Comma)?;
            let mut items = vec![first];
            self.parse_base_terms(&mut items)?;
            ast::BaseTerm::ApplyFn(fn_list_sym(self.arena), alloc_slice!(self, &items))
        } else {
            self.expect(Token::Colon)?; // is a map
            let first_val = self.parse_base_term()?;
            let mut items = vec![first, first_val];
            loop {
                if Token::Comma != self.token {
                    break;
                }
                self.next_token()?;
                items.push(self.parse_base_term()?);
                self.expect(Token::Colon)?;
                items.push(self.parse_base_term()?);
            }
            ast::BaseTerm::ApplyFn(fn_map_sym(self.arena), alloc_slice!(self, &items))
        };
        self.expect(Token::RBracket)?;
        Ok(alloc!(self, expr))
    }

    fn parse_struct(&mut self) -> Result<&'arena ast::BaseTerm<'arena>> {
        self.expect(Token::LBrace)?;
        if Token::RBrace == self.token {
            self.next_token()?;
            return Ok(alloc!(
                self,
                ast::BaseTerm::ApplyFn(fn_struct_sym(self.arena), &[])
            ));
        }
        let mut items = vec![];
        let name = self.parse_base_term()?;
        if let ast::BaseTerm::Const(ast::Const::Name { .. }) = name {
            items.push(name)
        } else {
            bail!("parse_base_term: expected name in struct expression {{ ... }} got {name:?}",);
        }
        self.expect(Token::Colon)?;
        items.push(self.parse_base_term()?);
        loop {
            if Token::Comma != self.token {
                break;
            }
            let name = self.parse_base_term()?;
            if let ast::BaseTerm::Const(ast::Const::Name { .. }) = name {
                items.push(name)
            } else {
                bail!("parse_base_term: expected name in struct expression {{ ... }} got {name:?}");
            }
            self.expect(Token::Colon)?;
            items.push(self.parse_base_term()?);
        }
        self.expect(Token::RBrace)?;
        Ok(alloc!(
            self,
            ast::BaseTerm::ApplyFn(fn_struct_sym(self.arena), alloc_slice!(self, &items))
        ))
    }

    /// paren_langle_terms ::=  `<` [base_terms] `>`
    fn parse_langle_base_terms(
        &mut self,
        base_terms: &mut Vec<&'arena ast::BaseTerm<'arena>>,
    ) -> Result<()> {
        self.expect(Token::Lt)?;
        if Token::Gt != self.token {
            self.parse_base_terms(base_terms)?;
        }
        self.expect(Token::Gt)?;
        Ok(())
    }

    /// paren_base_terms ::=  `(` [base_terms] `)`
    fn parse_paren_base_terms(
        &mut self,
        base_terms: &mut Vec<&'arena ast::BaseTerm<'arena>>,
    ) -> Result<()> {
        self.expect(Token::LParen)?;
        if Token::RParen != self.token {
            self.parse_base_terms(base_terms)?;
        }
        self.expect(Token::RParen)?;
        Ok(())
    }

    /// base_terms ::= base_term { `,` base_term }
    fn parse_base_terms(
        &mut self,
        base_terms: &mut Vec<&'arena ast::BaseTerm<'arena>>,
    ) -> Result<()> {
        base_terms.push(self.parse_base_term()?);
        while let Token::Comma = self.token {
            self.next_token()?;
            base_terms.push(self.parse_base_term()?);
        }

        Ok(())
    }
}

fn is_variable(name: &str) -> bool {
    name.chars().next().unwrap().is_ascii_uppercase()
}

fn is_fn(name: &str) -> bool {
    name.starts_with("fn:")
}

fn base_term_start(t: &Token) -> bool {
    match t {
        Token::Name { .. }
        | Token::Int { .. }
        | Token::Float { .. }
        | Token::String { .. }
        | Token::Bytes { .. }
        | Token::LBracket
        | Token::LBrace
        | Token::DotIdent { .. } => true,
        Token::Ident { name } => is_variable(name) || is_fn(name),
        _ => false,
    }
}

#[cfg(test)]
mod test {

    use super::*;
    use googletest::prelude::{eq, gtest, verify_that};

    fn make_parser<'arena>(
        arena: &'arena Arena,
        input: &'arena str,
    ) -> Parser<'arena, &'arena [u8]> {
        let mut p = Parser::new(arena, input.as_bytes(), "test");
        p.next_token().unwrap();
        p
    }

    #[test]
    fn test_empty_unit() -> Result<()> {
        let arena = Arena::new_with_global_interner();
        let mut p = make_parser(&arena, "");
        match p.parse_unit()? {
            &ast::Unit { decls: &[pkg], .. } => {
                assert_eq!(pkg, &empty_package_decl(&arena));
            }
            z => panic!("unexpected: {:?}", z),
        }
        Ok(())
    }

    #[test]
    fn test_package_use() -> Result<()> {
        let arena = Arena::new_with_global_interner();
        let input = "Package foo[bar()]! Use baz[bar()]!";

        let mut p = make_parser(&arena, input);
        let unit = p.parse_unit()?;
        match unit.decls {
            &[
                &ast::Decl {
                    atom:
                        &ast::Atom {
                            sym: got_package_sym,
                            ..
                        },
                    descr:
                        &[
                            &ast::Atom {
                                sym: got_name_sym1,
                                args: &[ast::BaseTerm::Const(ast::Const::String("foo"))],
                            },
                            &ast::Atom {
                                sym: got_bar_sym1,
                                args: &[],
                            },
                        ],
                    ..
                },
                &ast::Decl {
                    atom:
                        &ast::Atom {
                            sym: got_use_sym, ..
                        },
                    descr:
                        &[
                            &ast::Atom {
                                sym: got_name_sym2,
                                args: &[ast::BaseTerm::Const(ast::Const::String("baz"))],
                            },
                            &ast::Atom {
                                sym: got_bar_sym2,
                                args: &[],
                            },
                        ],
                    ..
                },
            ] => {
                assert_eq!(got_use_sym, use_sym(&arena));
                assert_eq!(got_package_sym, package_sym(&arena));
                assert_eq!(got_name_sym1, name_sym(&arena));
                assert_eq!(got_name_sym2, name_sym(&arena));
                assert_eq!(got_bar_sym1, arena.predicate_sym("bar", None));
                assert_eq!(got_bar_sym2, arena.predicate_sym("bar", None));
            }
            z => panic!("unexpected {z:?}"),
        }
        Ok(())
    }

    #[test]
    fn test_decl() -> Result<()> {
        let arena = Arena::new_with_global_interner();
        let input = "Decl foo(X, Y).";
        let mut p = make_parser(&arena, input);
        match p.parse_decl()? {
            &ast::Decl {
                atom:
                    &ast::Atom {
                        sym: got_foo_sym,
                        args:
                            &[
                                &ast::BaseTerm::Variable(x_sym),
                                &ast::BaseTerm::Variable(y_sym),
                            ],
                    },
                ..
            } => {
                assert_eq!(got_foo_sym, arena.predicate_sym("foo", None));
                assert_eq!(x_sym, arena.variable_sym("X"));
                assert_eq!(y_sym, arena.variable_sym("Y"))
            }
            decl => panic!("got {:?}", decl),
        };
        Ok(())
    }

    #[test]
    fn test_base_term() -> googletest::Result<()> {
        let arena = Arena::new_with_global_interner();
        let input = "X 3 1.5 'foo' /foo fn:list() fn:list(/a) fn:list(/a, 3)"; //.as_bytes();
        let mut p = make_parser(&arena, input);
        let mut got_base_terms = vec![];
        loop {
            if Token::Eof == p.token {
                break;
            }
            // TODO: "err_to_test_failure".
            let base_term = p.parse_base_term().unwrap();
            got_base_terms.push(base_term);
        }
        let expected = vec![
            arena.variable("X"),
            arena.const_(ast::Const::Number(3)),
            arena.const_(ast::Const::Float(1.5)),
            arena.const_(ast::Const::String("foo")),
            arena.const_(arena.name("/foo")),
            arena.apply_fn(fn_list_sym(&arena), &[]),
            arena.apply_fn(fn_list_sym(&arena), &[arena.const_(arena.name("/a"))]),
            arena.apply_fn(
                fn_list_sym(&arena),
                &[
                    arena.const_(arena.name("/a")),
                    arena.const_(ast::Const::Number(3)),
                ],
            ),
        ];
        verify_that!(got_base_terms, eq(&expected))
    }

    #[test]
    fn test_term() -> googletest::Result<()> {
        let arena = Arena::new_with_global_interner();
        let input = "foo(/bar) !bar() X = Z X != 3 3 < 1 3 <= 1";
        let mut p = make_parser(&arena, input);
        let mut got_terms = vec![];
        loop {
            if Token::Eof == p.token {
                break;
            }
            // TODO: "err_to_test_failure".
            got_terms.push(p.parse_term().unwrap());
        }
        let expected = [
            &ast::Term::Atom(arena.atom(
                arena.predicate_sym("foo", None),
                &[arena.const_(arena.name("/bar"))],
            )),
            &ast::Term::NegAtom(arena.atom(arena.predicate_sym("bar", None), &[])),
            &ast::Term::Eq(arena.variable("X"), arena.variable("Z")),
            &ast::Term::Ineq(
                arena.variable("X"),
                arena.alloc(ast::BaseTerm::Const(ast::Const::Number(3))),
            ),
            &ast::Term::Atom(arena.atom(
                arena.predicate_sym(":lt", Some(2)),
                &[
                    arena.const_(ast::Const::Number(3)),
                    arena.const_(ast::Const::Number(1)),
                ],
            )),
            &ast::Term::Atom(arena.atom(
                arena.predicate_sym(":le", Some(2)),
                &[
                    arena.const_(ast::Const::Number(3)),
                    arena.const_(ast::Const::Number(1)),
                ],
            )),
        ];
        verify_that!(got_terms, eq(&expected))
    }

    #[gtest]
    fn test_structured_data_and_types() -> googletest::Result<()> {
        let arena = Arena::new_with_global_interner();
        let input =
            "[] [1,2,3] [1: 'one', 2: 'two'] {} {/foo: /bar} .List<.Option</name>, /string>";
        let mut p = make_parser(&arena, input);
        let mut got_base_terms = vec![];
        loop {
            if Token::Eof == p.token {
                break;
            }
            // TODO: "err_to_test_failure".
            let base_term = p.parse_base_term().unwrap();
            got_base_terms.push(base_term);
        }
        let expected = vec![
            arena.apply_fn(fn_list_sym(&arena), &[]),
            arena.apply_fn(
                fn_list_sym(&arena),
                &[
                    arena.const_(ast::Const::Number(1)),
                    arena.const_(ast::Const::Number(2)),
                    arena.const_(ast::Const::Number(3)),
                ],
            ),
            arena.apply_fn(
                fn_map_sym(&arena),
                &[
                    arena.const_(ast::Const::Number(1)),
                    arena.const_(ast::Const::String("one")),
                    arena.const_(ast::Const::Number(2)),
                    arena.const_(ast::Const::String("two")),
                ],
            ),
            arena.apply_fn(fn_struct_sym(&arena), &[]),
            arena.apply_fn(
                fn_struct_sym(&arena),
                &[
                    arena.const_(arena.name("/foo")),
                    arena.const_(arena.name("/bar")),
                ],
            ),
            arena.apply_fn(
                fn_list_type_sym(&arena),
                &[
                    arena.apply_fn(
                        fn_option_type_sym(&arena),
                        &[arena.const_(arena.name("/name"))],
                    ),
                    arena.const_(arena.name("/string")),
                ],
            ),
        ];
        verify_that!(got_base_terms, eq(&expected))
    }

    #[test]
    fn test_clause() -> Result<()> {
        let arena = Arena::new_with_global_interner();
        let mut p = make_parser(&arena, "foo(X).");
        let clause = p.parse_clause()?;
        match clause {
            &ast::Clause {
                head:
                    &ast::Atom {
                        args: &[ast::BaseTerm::Variable(x_sym)],
                        ..
                    },
                premises: &[],
                transform: &[],
            } => {
                assert_eq!(*x_sym, arena.variable_sym("X"));
                assert_eq!(clause.head.sym, arena.predicate_sym("foo", None));
            }
            _ => panic!("unexpected: {:?}", clause),
        }
        let mut p = make_parser(&arena, "foo(X) :- !bar(X).");
        let clause = p.parse_clause()?;
        match clause {
            &ast::Clause {
                head:
                    &ast::Atom {
                        sym: foo_sym,
                        args: _,
                    },
                premises:
                    &[
                        &ast::Term::NegAtom(&ast::Atom {
                            sym: bar_sym,
                            args: _,
                        }),
                    ],
                transform: &[],
            } => {
                assert_eq!(foo_sym, arena.predicate_sym("foo", None));
                assert_eq!(bar_sym, arena.predicate_sym("bar", None));
            }
            _ => panic!("unexpected: {:?}", clause),
        };
        let mut p = make_parser(
            &arena,
            "foo(Z) ⟸ bar(Y) |> do fn:group_by(); let X = fn:count(Y).",
        );

        let clause = p.parse_clause()?;
        match clause {
            &ast::Clause {
                head: &ast::Atom { .. },
                premises: &[&ast::Term::Atom(ast::Atom { .. })],
                transform:
                    &[
                        &ast::TransformStmt {
                            var: None,
                            app: ast::BaseTerm::ApplyFn(first_sym, _),
                        },
                        &ast::TransformStmt {
                            var: Some("X"),
                            app: ast::BaseTerm::ApplyFn(second_sym, _),
                        },
                    ],
            } => {
                assert_eq!(clause.head.sym, arena.predicate_sym("foo", None));
                assert_eq!(clause.transform.len(), 2);
                assert_eq!(*first_sym, arena.function_sym("fn:group_by", None));
                assert_eq!(*second_sym, arena.function_sym("fn:count", None));
            }
            _ => panic!("unexpected: {:?}", clause),
        }

        Ok(())
    }
}
