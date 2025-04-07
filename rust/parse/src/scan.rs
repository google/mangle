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

use anyhow::{anyhow, Result};
use std::io;

use crate::error::{ErrorContext, ScanError};
use crate::quote::{unquote, DecodedSequence};
use crate::token::Token;

// Scanner turns a stream of bytes into a stream of tokens.
pub struct Scanner<R>
where
    R: io::Read,
{
    iter: io::Bytes<R>,

    // Peeked char.
    ch: Option<char>,

    pub line: usize,
    pub col: usize,
    /// Byte offset of start of current line.
    pub start_of_line: usize,

    // Text for the last token.
    text: String,

    // Path to file we are parsing. Used for error messages only.
    path: String,
}

impl<R> Scanner<R>
where
    R: io::Read,
{
    pub fn new<P: ToString>(reader: R, path: P) -> Self {
        Self {
            iter: reader.bytes(),
            ch: None,
            line: 1,
            col: 0,
            start_of_line: 0,
            text: String::new(),
            path: path.to_string(),
        }
    }

    pub fn get_error_context(&self) -> ErrorContext {
        ErrorContext {
            path: self.path.clone(),
            line: self.line,
            col: self.col,
            start_of_line: self.start_of_line,
        }
    }

    pub fn next_token(&mut self) -> Result<Token> {
        self.next_token_internal()
    }

    fn next_token_internal(&mut self) -> Result<Token> {
        match self.next_char_skip()? {
            Some('=') => Ok(Token::Eq),
            Some(';') => Ok(Token::Semi),
            Some(',') => Ok(Token::Comma),
            Some('!') => match self.peek()? {
                Some('=') => {
                    let _ = self.next_char()?;
                    Ok(Token::BangEq)
                }
                _ => Ok(Token::Bang),
            },
            Some('(') => Ok(Token::LParen),
            Some(')') => Ok(Token::RParen),
            Some('{') => Ok(Token::LBrace),
            Some('}') => Ok(Token::RBrace),
            Some('[') => Ok(Token::LBracket),
            Some(']') => Ok(Token::RBracket),
            Some('≤') => Ok(Token::Le), // unicode \u2264
            Some('<') => match self.peek()? {
                Some('=') => {
                    let _ = self.next_char()?;
                    Ok(Token::Le)
                }
                _ => Ok(Token::Lt),
            },
            Some('≥') => Ok(Token::Ge), // unicode \u2265
            Some('>') => match self.peek()? {
                Some('=') => {
                    let _ = self.next_char()?;
                    Ok(Token::Ge)
                }
                _ => Ok(Token::Gt),
            },
            Some(':') => match self.peek()? {
                Some('-') => {
                    let _ = self.next_char()?;
                    Ok(Token::ColonDash)
                }
                _ => Ok(Token::Colon),
            },
            Some('|') => match self.peek()? {
                Some('>') => {
                    let _ = self.next_char()?;
                    Ok(Token::PipeGt)
                }
                _ => Ok(Token::Pipe),
            },
            Some('.') => match self.peek()? {
                Some('A'..'Z') => {
                    let first = self.next_char()?.expect("could not get peeked character.");
                    self.ident_or_dot_ident(first, true)
                }
                _ => Ok(Token::Dot),
            },
            Some('/') => self.name(),
            Some('⟸') => Ok(Token::LongLeftDoubleArrow),
            Some(delim @ '\'') => self.string(delim, false),
            Some(delim @ '"') => self.string(delim, false),
            Some(first @ '0'..='9') => self.numeric(first),
            Some(ch) if is_ident_start(ch) => {
                if ch == 'b' {
                    if let Some(delim @ ('\'' | '"')) = self.peek()? {
                        let _ = self.next_char()?;
                        return self.string(delim, true);
                    }
                }
                self.ident(ch)
            }
            Some(ch) => Err(anyhow!(ScanError::Unexpected(self.get_error_context(), ch))),
            None => Ok(Token::Eof),
        }
    }

    fn name(&mut self) -> Result<Token> {
        self.text.clear();
        self.text.push('/');
        let mut seen_char = false;
        loop {
            match self.peek()? {
                Some(c) if is_name_char(c) => {
                    self.next_char()?;
                    self.text.push(c);
                    seen_char = true;
                }
                Some('/') => {
                    self.next_char()?;
                    if !seen_char {
                        anyhow::bail!("name constant: expected char after {}", self.text)
                    }
                    self.text.push('/');
                    seen_char = false;
                }
                _ => break,
            }
        }
        if !seen_char {
            anyhow::bail!("name constant: expected name char after {}", self.text)
        }
        Ok(Token::Name { name: self.text.to_string() })
    }

    // TODO: this only handles single-double quoted (short. not long).
    fn string(&mut self, delim: char, is_byte: bool) -> Result<Token> {
        self.text.clear();
        if is_byte {
            self.text.push('b');
        }
        self.text.push(delim); // TODO
        loop {
            match self.next_char()? {
                Some(c) if c == delim => break,
                Some(c) => self.text.push(c),
                _ => break,
            }
        }
        self.text.push(delim); // TODO
        match unquote(self.text.as_str())? {
            DecodedSequence::String(decoded) => Ok(Token::String { decoded }),
            DecodedSequence::Bytes(decoded) => Ok(Token::Bytes { decoded }),
        }
    }

    fn numeric(&mut self, first: char) -> Result<Token> {
        self.text.clear();
        self.text.push(first);
        let mut is_float = false;
        loop {
            match self.peek()? {
                Some(c @ '0'..='9') => {
                    self.next_char()?;
                    self.text.push(c)
                }
                Some(c @ '.') => {
                    self.next_char()?;
                    is_float = true;
                    self.text.push(c)
                }
                _ => break,
            }
        }
        if is_float {
            let num = self.text.parse::<f64>()?;
            return Ok(Token::Float { decoded: num });
        }
        let num = self.text.parse::<i64>()?;
        // if "0x" == self.text[0..2] {
        //    todo![]
        // }
        Ok(Token::Int { decoded: num })
    }

    fn ident(&mut self, first: char) -> Result<Token> {
        self.ident_or_dot_ident(first, false)
    }

    fn ident_or_dot_ident(&mut self, first: char, dot_ident: bool) -> Result<Token> {
        self.text.clear();
        self.text.push(first);
        loop {
            match self.peek()? {
                Some(ch) if is_ident(ch) => {
                    self.next_char()?;
                    self.text.push(ch);
                }
                Some(':') if self.text == "fn" => {
                    self.next_char()?;
                    self.text.push(':');
                }
                _ => {
                    return match self.text.as_str() {
                        "Package" => Ok(Token::Package),
                        "Use" => Ok(Token::Use),
                        "Decl" => Ok(Token::Decl),
                        "bound" => Ok(Token::Bound),
                        "inclusion" => Ok(Token::Inclusion),
                        "do" => Ok(Token::Do),
                        "descr" => Ok(Token::Descr),
                        "let" => Ok(Token::Let),
                        _ if dot_ident => {
                            let mut fn_name = String::new();
                            fn_name.push_str("fn:");
                            fn_name.push_str(&self.text);
                            Ok(Token::DotIdent { name: fn_name })
                        }
                        _ => Ok(Token::Ident { name: self.text.clone() }),
                    }
                }
            }
        }
    }

    #[inline]
    fn next_char(&mut self) -> Result<Option<char>> {
        if let Some(c) = self.ch.take() {
            return Ok(Some(c));
        }
        macro_rules! next_byte_or_incomplete {
            ($self:expr) => {
                $self
                    .next_byte()?
                    .ok_or_else(|| anyhow!(ScanError::IncompleteUtf8(self.get_error_context())))
            };
        }
        let b = self.next_byte()?;
        match b {
            None => Ok(None),
            Some(b @ 0x00..=0x7F) => Ok(Some(unsafe { char::from_u32_unchecked(b.into()) })),
            Some(b1 @ 0xC0..=0xDF) => {
                let b2 = next_byte_or_incomplete!(self)?;
                let bytes = [b1, b2];
                let s = std::str::from_utf8(&bytes)?;
                Ok(s.chars().next())
            }
            Some(b1 @ 0xE0..=0xEF) => {
                let b2 = next_byte_or_incomplete!(self)?;
                let b3 = next_byte_or_incomplete!(self)?;
                let bytes = [b1, b2, b3];
                let s = std::str::from_utf8(&bytes)?;
                Ok(s.chars().next())
            }
            Some(b1 @ 0xF0..=0xF4) => {
                let b2 = next_byte_or_incomplete!(self)?;
                let b3 = next_byte_or_incomplete!(self)?;
                let b4 = next_byte_or_incomplete!(self)?;
                let bytes = [b1, b2, b3, b4];
                let s = std::str::from_utf8(&bytes)?;
                Ok(s.chars().next())
            }
            _ => Err(anyhow!("invalid utf8")),
        }
    }

    /// Advance to next non-whitespace byte. Skip comments.
    #[inline]
    fn next_char_skip(&mut self) -> Result<Option<char>> {
        loop {
            match self.next_char()? {
                Some(' ' | '\t' | '\n') => {}
                Some('#') => self.skip_line()?,
                z => return Ok(z),
            };
        }
    }

    // Skip bytes until newline (included).
    fn skip_line(&mut self) -> Result<()> {
        loop {
            match self.next_byte()? {
                Some(b'\n') | None => return Ok(()),
                _ => {}
            }
        }
    }

    // Advance exactly one byte.
    fn next_byte(&mut self) -> Result<Option<u8>> {
        match self.iter.next() {
            None => Ok(None),
            Some(Ok(b'\n')) => {
                self.start_of_line += self.col + 1;
                self.line += 1;
                self.col = 0;
                Ok(Some(b'\n'))
            }
            Some(Ok(c)) => {
                self.col += 1;
                Ok(Some(c))
            }
            Some(Err(e)) => Err(e.into()),
        }
    }

    #[inline]
    pub fn peek(&mut self) -> Result<Option<char>> {
        Ok(match self.ch {
            Some(ch) => Some(ch),
            None => {
                self.ch = self.next_char()?;
                self.ch
            }
        })
    }
}

fn is_ident_start(c: char) -> bool {
    matches!(c, 'a'..='z' | 'A'..='Z' | '_' )
}

fn is_ident(c: char) -> bool {
    matches!(c, 'a'..='z' | 'A'..='Z' | '0'..='9' | '_' )
}

fn is_name_char(c: char) -> bool {
    matches!(c, 'a'..='z' | 'A'..='Z' | '0'..='9' | '_' | '~' | '.' | '%')
}

#[cfg(test)]
mod test {

    use super::*;

    #[test]
    fn test_ident() -> Result<()> {
        let mut sc = Scanner::new("hello".as_bytes(), "test");
        let token = sc.next_token()?;
        match token {
            Token::Ident { name } if name == "hello" => {}
            _ => assert!(false, "did not match"),
        }
        Ok(())
    }

    fn scan_all(s: &str) -> Result<Vec<Token>> {
        let mut sc = Scanner::new(s.as_bytes(), "test");
        let mut got = vec![];
        loop {
            let token = sc.next_token()?;
            if let Token::Eof = token {
                return Ok(got);
            }
            got.push(token.clone());
        }
    }

    #[test]
    fn test_keywords() -> Result<()> {
        let got = scan_all("do ⟸ let bound descr inclusion Package Use")?;
        use Token::*;
        let want = vec![Do, LongLeftDoubleArrow, Let, Bound, Descr, Inclusion, Package, Use];
        assert!(want == got, "want {:?} got {:?}", want, got);
        Ok(())
    }

    #[test]
    fn test_values() -> Result<()> {
        let got = scan_all("1 3.14 'foo🤖' b'foo👷‍♀️' \"bar\" b\"bar\" ")?;
        let want = vec![
            Token::Int { decoded: 1 },
            Token::Float { decoded: 3.14 },
            Token::String { decoded: "foo🤖".to_string() },
            Token::Bytes { decoded: "foo👷‍♀️".as_bytes().into() },
            Token::String { decoded: "bar".to_string() },
            Token::Bytes { decoded: "bar".as_bytes().into() },
        ];
        assert!(want == got, "want {:?} got {:?}", want, got);
        Ok(())
    }

    #[test]
    fn test_punctuation() -> Result<()> {
        let got = scan_all(".=!!=()[]{}::-|>")?;
        use Token::*;
        let want = vec![
            Dot, Eq, Bang, BangEq, LParen, RParen, LBracket, RBracket, LBrace, RBrace, Colon,
            ColonDash, PipeGt,
        ];
        assert!(want == got, "want {:?} got {:?}", want, got);
        Ok(())
    }

    #[test]
    fn test_names() -> Result<()> {
        let got = scan_all("/foo /foo/bar")?;
        let want = vec![
            Token::Name { name: "/foo".to_string() },
            Token::Name { name: "/foo/bar".to_string() },
        ];
        assert!(want == got, "want {:?} got {:?}", want, got);
        Ok(())
    }

    #[test]
    fn test_names_negative() -> Result<()> {
        scan_all("/").unwrap_err();
        scan_all("/foo/").unwrap_err();
        Ok(())
    }
}
