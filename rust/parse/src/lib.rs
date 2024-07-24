#![allow(dead_code)]
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

use ast::Constraints;
use bumpalo::Bump;
/// Open the file in read-only mode with buffer.
///
/// let file = std::fs::File::open(path)?;
/// let reader = std::io::BufReader::new(file);
///
use std::io;

use anyhow::{anyhow, bail, Result};
use mangle_ast as ast;

mod error;
mod quote;
mod scan;
mod token;

pub use error::{ErrorContext, ParseError};
use token::Token;

pub struct Parser<'b, R>
where
    R: io::Read,
{
    sc: scan::Scanner<R>,
    token: crate::token::Token,
    bump: &'b Bump,
}

const PACKAGE_SYM: ast::PredicateSym = ast::PredicateSym {
    name: "Package",
    arity: Some(0),
};

const NAME_SYM: ast::PredicateSym = ast::PredicateSym {
    name: "name",
    arity: Some(1),
};

const USE_SYM: ast::PredicateSym = ast::PredicateSym {
    name: "Use",
    arity: Some(0),
};

const LT_SYM: ast::PredicateSym = ast::PredicateSym {
    name: ":lt",
    arity: Some(2),
};

const LE_SYM: ast::PredicateSym = ast::PredicateSym {
    name: ":le",
    arity: Some(2),
};

const FN_LIST_SYM: ast::FunctionSym = ast::FunctionSym {
    name: "fn:list",
    arity: None,
};
const FN_MAP_SYM: ast::FunctionSym = ast::FunctionSym {
    name: "fn:map",
    arity: None,
};
const FN_STRUCT_SYM: ast::FunctionSym = ast::FunctionSym {
    name: "fn:struct",
    arity: None,
};

const EMPTY_PACKAGE: ast::Decl<'_> = ast::Decl {
    atom: &ast::Atom {
        sym: PACKAGE_SYM,
        args: &[],
    },
    descr: &[&ast::Atom {
        sym: NAME_SYM,
        args: &[&ast::BaseTerm::Const(ast::Const::String(""))],
    }],
    bounds: None,
    constraints: None,
};

macro_rules! alloc {
    ($self:expr, $e:expr) => {
        &*$self.bump.alloc($e)
    };
}

macro_rules! alloc_str {
    ($self:expr, $e:expr) => {
        &*$self.bump.alloc_str($e)
    };
}

macro_rules! alloc_slice {
    ($self:expr, $e:expr) => {
        &*$self.bump.alloc_slice_copy($e)
    };
}

impl<'b, R> Parser<'b, R>
where
    R: io::Read,
{
    pub fn new<P: ToString>(bump: &'b Bump, reader: R, path: P) -> Self
    where
        R: io::Read,
    {
        Self {
            sc: scan::Scanner::new(reader, path),
            token: token::Token::Illegal,
            bump,
        }
    }

    fn next_token(&mut self) -> Result<()> {
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

    pub fn parse_unit(&mut self) -> Result<&'b ast::Unit<'b>> {
        let package = if matches!(self.token.clone(), Token::Package) {
            self.parse_package_decl()?
        } else {
            &EMPTY_PACKAGE
        };
        let mut decls = vec![package];
        while let Token::Use = self.token {
            decls.push(self.parse_use_decl()?);
        }
        let decls: &'b [&'b ast::Decl<'b>] = &*self.bump.alloc_slice_copy(&decls);
        let unit: &'b ast::Unit<'b> = &*self.bump.alloc(ast::Unit {
            clauses: &[],
            decls,
        });
        Ok(unit)
    }

    /// package_decl ::= `package` name (`[` `]`)? `!`
    pub fn parse_package_decl(&mut self) -> Result<&'b ast::Decl<'b>> {
        self.expect(Token::Package)?;
        let name = match &self.token {
            Token::Ident { name } => name.as_str(),
            _ => bail!("expected identifer got {}", self.token),
        };

        let name_atom = ast::Atom {
            sym: NAME_SYM,
            args: &[&ast::BaseTerm::Const(ast::Const::String(name))],
        };
        let mut descr_atoms: Vec<&ast::Atom> = vec![ast::copy_atom(self.bump, &name_atom)];
        self.next_token()?;
        if Token::LBracket == self.token {
            self.parse_bracket_atoms(&mut descr_atoms)?;
        }

        self.expect(Token::Bang)?;

        let package_atom = alloc!(
            self,
            ast::Atom {
                sym: PACKAGE_SYM,
                args: &[],
            }
        );

        //let descr_atoms = ;
        let decl: &'b ast::Decl = alloc!(
            self,
            ast::Decl {
                atom: package_atom,
                bounds: None,
                descr: alloc_slice!(self, &descr_atoms),
                constraints: None,
            }
        );
        Ok(decl)
    }

    fn parse_use_decl(&mut self) -> Result<&'b ast::Decl<'b>> {
        self.expect(Token::Use)?;
        let use_atom = alloc!(
            self,
            ast::Atom {
                sym: USE_SYM,
                args: &[],
            }
        );

        let name = match &self.token {
            Token::Ident { name } => name.as_str(),
            _ => bail!("parse_use_decl: expected identifer got {}", self.token),
        };

        let name: &'b str = alloc_str!(self, name);
        let name = alloc!(self, ast::BaseTerm::Const(ast::Const::String(name)));
        let args = alloc_slice!(self, &[name]);

        let mut descr_atoms: Vec<&ast::Atom> = vec![self.bump.alloc(ast::Atom {
            sym: NAME_SYM,
            args,
        })];
        self.next_token()?;
        if Token::LBracket == self.token {
            self.parse_bracket_atoms(&mut descr_atoms)?;
        }

        let descr_atoms = alloc_slice!(self, &descr_atoms);
        Ok(alloc!(
            self,
            ast::Decl {
                atom: use_atom,
                descr: descr_atoms,
                bounds: None,
                constraints: None,
            }
        ))
    }

    fn parse_decl(&mut self) -> Result<&'b ast::Decl<'b>> {
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
                constraints,
            }
        ))
    }

    /// bound_decl ::= `bound` `[` base_term {`,` base_term} `]`
    fn parse_bounds_decl(&mut self) -> Result<&'b ast::BoundDecl<'b>> {
        self.expect(Token::Bound)?;
        self.expect(Token::LBracket)?;
        let mut base_terms = vec![];
        self.parse_base_terms(&mut base_terms)?;
        self.expect(Token::RBracket)?;
        let base_terms = alloc_slice!(self, &base_terms);
        let bound_decl = alloc!(self, ast::BoundDecl { base_terms });
        Ok(bound_decl)
    }

    fn parse_inclusion_constraint(&mut self) -> Result<&'b ast::Constraints<'b>> {
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

    fn parse_clause(&mut self) -> Result<&'b ast::Clause<'b>> {
        let head = self.parse_atom()?;
        let mut premises = vec![];
        let mut transform = vec![];
        if let Token::ColonDash = self.token {
            self.next_token()?;
            self.parse_terms(&mut premises)?;
            if let Token::PipeGt = self.token {
                self.next_token()?;
                self.parse_transforms(&mut transform)?;
            }
        }
        self.expect(Token::Dot)?;
        let premises = alloc_slice!(self, &premises);
        let transform = alloc_slice!(self, &transform);
        Ok(alloc!(
            self,
            ast::Clause {
                head,
                premises,
                transform,
            }
        ))
    }

    /// terms ::= term { , term }
    fn parse_terms(&mut self, terms: &mut Vec<&'b ast::Term<'b>>) -> Result<()> {
        terms.push(self.parse_term()?);
        loop {
            if Token::Comma != self.token {
                return Ok(());
            }
            self.next_token()?;
            terms.push(self.parse_term()?);
        }
    }

    fn parse_term(&mut self) -> Result<&'b ast::Term<'b>> {
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
                            sym: LT_SYM,
                            args: alloc_slice!(self, &[left_base_term, right_base_term]),
                        }
                    )),
                    Token::Le => ast::Term::Atom(self.bump.alloc(ast::Atom {
                        sym: LE_SYM,
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
    fn parse_bracket_atoms(&mut self, atoms: &mut Vec<&'b ast::Atom<'b>>) -> Result<()> {
        self.expect(Token::LBracket)?;
        self.parse_atoms(atoms)?;
        self.expect(Token::RBracket)?;
        Ok(())
    }

    // `atoms ::= [ atom {`,` atom } ]
    fn parse_atoms(&mut self, atoms: &mut Vec<&'b ast::Atom<'b>>) -> Result<()> {
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
    fn parse_atom(&mut self) -> Result<&'b ast::Atom<'b>> {
        let name = match &self.token {
            Token::Ident { name } => name.as_str(),
            _ => bail!("parse_atom: expected identifer got {}", self.token),
        };
        let name = &*self.bump.alloc_str(name);

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
                sym: ast::PredicateSym { name, arity: None },
                args,
            }
        ))
    }

    fn parse_transforms(&mut self, transforms: &mut Vec<&'b ast::TransformStmt<'b>>) -> Result<()> {
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
    fn parse_base_term(&mut self) -> Result<&'b ast::BaseTerm<'b>> {
        match &self.token {
            Token::LBracket => return self.parse_list_or_map(),
            Token::LBrace => return self.parse_struct(),
            _ => {}
        }

        let mut base_term = match &self.token {
            Token::Ident { name } if is_variable(name) => {
                let name = alloc_str!(self, &name);
                ast::BaseTerm::Variable(name)
            }
            Token::Ident { name } if is_fn(name) => {
                let name = self.bump.alloc_str(name);
                // Arguments parsed below.
                ast::BaseTerm::ApplyFn(ast::FunctionSym { name, arity: None }, &[])
            }
            Token::String { decoded } => {
                let value = self.bump.alloc_str(decoded.as_str());
                ast::BaseTerm::Const(ast::Const::String(value))
            }
            Token::Bytes { decoded } => {
                let value = self.bump.alloc_slice_copy(decoded);
                ast::BaseTerm::Const(ast::Const::Bytes(value))
            }
            Token::Int { decoded } => ast::BaseTerm::Const(ast::Const::Number(*decoded)),
            Token::Float { decoded } => ast::BaseTerm::Const(ast::Const::Float(*decoded)),
            Token::Name { name } => {
                let name = self.bump.alloc_str(name.as_str());
                ast::BaseTerm::Const(ast::Const::Name(name))
            }
            _ => bail!("parse_base_term: unexpected token {:?}", self.token),
        };
        self.next_token()?;
        if let ast::BaseTerm::ApplyFn(fn_sym, _) = base_term {
            let mut fn_args = vec![];
            self.parse_paren_base_terms(&mut fn_args)?;
            let fn_args = self.bump.alloc_slice_copy(&fn_args);
            base_term = ast::BaseTerm::ApplyFn(fn_sym, fn_args);
        }
        let base_term = alloc!(self, base_term);
        Ok(base_term)
    }

    fn parse_list_or_map(&mut self) -> Result<&'b ast::BaseTerm<'b>> {
        self.expect(Token::LBracket)?;
        if Token::RBracket == self.token {
            self.next_token()?;
            return Ok(alloc!(self, ast::BaseTerm::ApplyFn(FN_MAP_SYM, &[])));
        }
        let first = self.parse_base_term()?;
        let expr = if Token::Colon != self.token {
            self.expect(Token::Comma)?;
            let mut items = vec![first];
            self.parse_base_terms(&mut items)?;
            ast::BaseTerm::ApplyFn(FN_LIST_SYM, alloc_slice!(self, &items))
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
            ast::BaseTerm::ApplyFn(FN_MAP_SYM, alloc_slice!(self, &items))
        };
        self.expect(Token::RBracket)?;
        Ok(alloc!(self, expr))
    }

    fn parse_struct(&mut self) -> Result<&'b ast::BaseTerm<'b>> {
        self.expect(Token::LBrace)?;
        if Token::RBrace == self.token {
            self.next_token()?;
            return Ok(alloc!(self, ast::BaseTerm::ApplyFn(FN_STRUCT_SYM, &[])));
        }
        let mut items = vec![];
        let name = self.parse_base_term()?;
        if let ast::BaseTerm::Const(ast::Const::Name { .. }) = name {
            items.push(name)
        } else {
            bail!(
                "parse_base_term: expected name in struct expression {{ ... }} got {:?}",
                name
            );
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
                bail!(
                    "parse_base_term: expected name in struct expression {{ ... }} got {:?}",
                    name
                );
            }
            self.expect(Token::Colon)?;
            items.push(self.parse_base_term()?);
        }
        self.expect(Token::RBrace)?;
        Ok(alloc!(
            self,
            ast::BaseTerm::ApplyFn(FN_STRUCT_SYM, alloc_slice!(self, &items))
        ))
    }

    /// paren_base_terms ::=  `(` [base_terms] `)`
    fn parse_paren_base_terms(
        &mut self,
        base_terms: &mut Vec<&'b ast::BaseTerm<'b>>,
    ) -> Result<()> {
        self.expect(Token::LParen)?;
        if Token::RParen != self.token {
            self.parse_base_terms(base_terms)?;
        }
        self.expect(Token::RParen)?;
        Ok(())
    }

    /// base_terms ::= base_term { `,` base_term }
    fn parse_base_terms(&mut self, base_terms: &mut Vec<&'b ast::BaseTerm<'b>>) -> Result<()> {
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
        | Token::LBrace => true,
        Token::Ident { name } => is_variable(name) || is_fn(name),
        _ => false,
    }
}

#[cfg(test)]
mod test {

    use super::*;

    fn make_parser<'b>(bump: &'b Bump, input: &'b str) -> Parser<'b, &'b [u8]> {
        let mut p = Parser::new(bump, input.as_bytes(), "test");
        p.next_token().unwrap();
        p
    }

    #[test]
    fn test_empty_unit() -> Result<()> {
        let bump = Bump::new();
        let mut p = make_parser(&bump, "");
        match p.parse_unit()? {
            ast::Unit { decls: &[pkg], .. } if *pkg == EMPTY_PACKAGE => {}
            z => panic!("unexpected: {:?}", z),
        }
        Ok(())
    }

    #[test]
    fn test_package_use() -> Result<()> {
        let bump = Bump::new();
        let input = "Package foo[bar()]! Use baz[bar()]!";

        //let mut p = Parser::new(&bump, input);
        let mut p = make_parser(&bump, input);
        let unit = p.parse_unit()?;
        match unit.decls {
            &[&ast::Decl {
                atom: &ast::Atom {
                    sym: PACKAGE_SYM, ..
                },
                descr:
                    &[&ast::Atom {
                        sym: NAME_SYM,
                        args: &[ast::BaseTerm::Const(ast::Const::String("foo"))],
                    }, &ast::Atom {
                        sym:
                            ast::PredicateSym {
                                name: "bar",
                                arity: None,
                            },
                        args: &[],
                    }],
                ..
            }, &ast::Decl {
                atom: &ast::Atom { sym: USE_SYM, .. },
                descr:
                    &[&ast::Atom {
                        sym: NAME_SYM,
                        args: &[ast::BaseTerm::Const(ast::Const::String("baz"))],
                    }, &ast::Atom {
                        sym:
                            ast::PredicateSym {
                                name: "bar",
                                arity: None,
                            },
                        args: &[],
                    }],
                ..
            }] => {}
            z => panic!("unexpcted {z:?}"),
        }
        Ok(())
    }

    #[test]
    fn test_decl() -> Result<()> {
        let bump = Bump::new();
        let input = "Decl foo(X, Y).";
        let mut p = make_parser(&bump, input);
        match p.parse_decl()? {
            ast::Decl {
                atom:
                    &ast::Atom {
                        sym:
                            ast::PredicateSym {
                                name: "foo",
                                arity: None,
                            },
                        args: &[&ast::BaseTerm::Variable("X"), &ast::BaseTerm::Variable("Y")],
                    },
                ..
            } => {}
            decl => panic!("got {:?}", decl),
        };
        Ok(())
    }

    #[test]
    fn test_base_term() -> Result<()> {
        let bump = Bump::new();
        let input = "X 3 1.5 'foo' /foo fn:list() fn:list(/a) fn:list(/a, 3)"; //.as_bytes();
        let mut p = make_parser(&bump, input);
        let mut got_base_terms = vec![];
        loop {
            if Token::Eof == p.token {
                break;
            }
            match p.parse_base_term() {
                Ok(base_term) => got_base_terms.push(base_term),
                Err(err) => {
                    println!("err: {err:?}");
                    return Err(err);
                }
            }
        }
        let expected = vec![
            &ast::BaseTerm::Variable("X"),
            &ast::BaseTerm::Const(ast::Const::Number(3)),
            &ast::BaseTerm::Const(ast::Const::Float(1.5)),
            &ast::BaseTerm::Const(ast::Const::String("foo")),
            &ast::BaseTerm::Const(ast::Const::Name("/foo")),
            &ast::BaseTerm::ApplyFn(
                ast::FunctionSym {
                    name: "fn:list",
                    arity: None,
                },
                &[],
            ),
            &ast::BaseTerm::ApplyFn(
                ast::FunctionSym {
                    name: "fn:list",
                    arity: None,
                },
                &[&ast::BaseTerm::Const(ast::Const::Name("/a"))],
            ),
            &ast::BaseTerm::ApplyFn(
                ast::FunctionSym {
                    name: "fn:list",
                    arity: None,
                },
                &[
                    &ast::BaseTerm::Const(ast::Const::Name("/a")),
                    &ast::BaseTerm::Const(ast::Const::Number(3)),
                ],
            ),
        ];
        assert!(
            expected == got_base_terms,
            "want: {expected:?}\n got: {got_base_terms:?}"
        );
        Ok(())
    }

    #[test]
    fn test_term() -> Result<()> {
        let bump = Bump::new();
        let input = "foo(/bar) !bar() X = Z X != 3 3 < 1 3 <= 1";
        let mut p = make_parser(&bump, input);
        let mut got_terms = vec![];
        loop {
            if Token::Eof == p.token {
                break;
            }
            match p.parse_term() {
                Ok(term) => got_terms.push(term),
                Err(err) => {
                    println!("err: {err:?}");
                    return Err(err);
                }
            }
        }
        let expected = vec![
            &ast::Term::Atom(&ast::Atom {
                sym: ast::PredicateSym {
                    name: "foo",
                    arity: None,
                },
                args: &[&ast::BaseTerm::Const(ast::Const::Name("/bar"))],
            }),
            &ast::Term::NegAtom(&ast::Atom {
                sym: ast::PredicateSym {
                    name: "bar",
                    arity: None,
                },
                args: &[],
            }),
            &ast::Term::Eq(&ast::BaseTerm::Variable("X"), &ast::BaseTerm::Variable("Z")),
            &ast::Term::Ineq(
                &ast::BaseTerm::Variable("X"),
                &ast::BaseTerm::Const(ast::Const::Number(3)),
            ),
            &ast::Term::Atom(&ast::Atom {
                sym: ast::PredicateSym {
                    name: ":lt",
                    arity: Some(2),
                },
                args: &[
                    &ast::BaseTerm::Const(ast::Const::Number(3)),
                    &ast::BaseTerm::Const(ast::Const::Number(1)),
                ],
            }),
            &ast::Term::Atom(&ast::Atom {
                sym: ast::PredicateSym {
                    name: ":le",
                    arity: Some(2),
                },
                args: &[
                    &ast::BaseTerm::Const(ast::Const::Number(3)),
                    &ast::BaseTerm::Const(ast::Const::Number(1)),
                ],
            }),
        ];
        if expected != got_terms {
            for (l, r) in expected.iter().zip(got_terms.iter()) {
                println!("{:?} == {:?} ? {}", l, r, l == r)
            }
        }
        assert!(
            expected == got_terms,
            "want: {expected:?}\n got: {got_terms:?}"
        );
        Ok(())
    }

    #[test]
    fn test_base_terms() -> Result<()> {
        let bump = Bump::new();
        let input = "[] [1,2,3] [1: 'one', 2: 'two'] {} {/foo: /bar}";
        let mut p = make_parser(&bump, input);
        let mut got_base_terms = vec![];
        loop {
            if Token::Eof == p.token {
                break;
            }
            match p.parse_base_term() {
                Ok(term) => got_base_terms.push(term),
                Err(err) => {
                    println!("err: {err:?}");
                    return Err(err);
                }
            }
        }
        Ok(())
    }

    #[test]
    fn test_clause() -> Result<()> {
        let bump = Bump::new();
        let mut p = make_parser(&bump, "foo(X).");
        let clause = p.parse_clause()?;
        assert!(matches!(
            clause,
            ast::Clause {
                head: ast::Atom {
                    sym: ast::PredicateSym {
                        name: "foo",
                        arity: None
                    },
                    args: &[ast::BaseTerm::Variable("X")]
                },
                premises: &[],
                transform: &[],
            }
        ));
        let mut p = make_parser(&bump, "foo(X) :- !bar(X).");
        let clause = p.parse_clause()?;
        assert!(matches!(
            clause,
            ast::Clause {
                head: ast::Atom {
                    sym: ast::PredicateSym { name: "foo", .. },
                    args: _
                },
                premises: &[&ast::Term::NegAtom(ast::Atom {
                    sym: ast::PredicateSym { name: "bar", .. },
                    args: _
                })],
                transform: &[],
            }
        ));
        let mut p = make_parser(
            &bump,
            "foo(Z) :- bar(Y) |> do fn:group_by(); let X = fn:count(Y).",
        );

        let clause = p.parse_clause()?;
        assert!(matches!(
            clause,
            ast::Clause {
                head: ast::Atom {
                    sym: ast::PredicateSym {
                        name: "foo",
                        arity: None
                    },
                    args: _
                },
                premises: &[&ast::Term::Atom(ast::Atom { .. })],
                transform: &[
                    &ast::TransformStmt {
                        var: None,
                        app: ast::BaseTerm::ApplyFn(
                            ast::FunctionSym {
                                name: "fn:group_by",
                                arity: None
                            },
                            _
                        )
                    },
                    &ast::TransformStmt {
                        var: Some("X"),
                        app: ast::BaseTerm::ApplyFn(
                            ast::FunctionSym {
                                name: "fn:count",
                                arity: None
                            },
                            _
                        )
                    }
                ],
            }
        ));

        Ok(())
    }
}
