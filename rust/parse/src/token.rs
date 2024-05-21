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

#[derive(Debug, PartialEq, Clone)]
pub enum Token {
    Illegal,
    Eof,

    Name { name: String },      // /foo/bar
    Ident { name: String },     // x
    Int { decoded: i64 },       // 123
    Float { decoded: f64 },     // 1.23e45
    String { decoded: String }, // "foo" or 'foo' or '''foo''' or r'foo' or r"foo"
    Bytes { decoded: Vec<u8> }, // b"foo", etc

    Semi,      // ;
    Package,   // Package
    Decl,      // Decl
    Use,       // Use
    Bound,     // bound
    Descr,     // descr
    Inclusion, // inclusion
    Let,       // let
    Do,        // do
    LParen,    // (
    RParen,    // )
    LBracket,  // [
    RBracket,  // ]
    LBrace,    // {
    RBrace,    // }
    Colon,     // :
    ColonDash, // :-
    Eq,        // =
    Bang,      // !
    BangEq,    // !=
    Comma,     // ,
    Lt,        // <
    Le,        // <=
    Gt,        // >
    Ge,        // >=
    Pipe,      // |
    PipeGt,    // |>
    Dot,       // .
}

impl std::fmt::Display for Token {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        macro_rules! token_text {
            ( $fstr:expr ) => {
                write!(f, $fstr)
            };
            ( $fstr:expr, $( $x:expr ),* ) => {
                write!(f, $fstr, ($($x),+))
            };
        }

        match self {
            Token::Illegal => token_text!["illegal"],
            Token::Eof => token_text!["eof"],
            Token::Package => token_text!["Package"],
            Token::Decl => token_text!["Decl"],

            Token::Name { name } => token_text!["{}", name],
            Token::Ident { name } => token_text!["{}", name],
            Token::Int { decoded } => token_text!["{}", decoded],
            Token::Float { decoded } => token_text!["{}", decoded],
            Token::String { decoded } => token_text!["{}", crate::quote::quote(decoded.as_str())],
            Token::Bytes { decoded } => token_text!["{:?}", decoded],

            Token::Semi => token_text![";"],
            Token::Use => token_text!["Use"],
            Token::LParen => token_text!["("],
            Token::RParen => token_text![")"],
            Token::LBrace => token_text!["{{"],
            Token::RBrace => token_text!["}}"],
            Token::Colon => token_text![":"],
            Token::ColonDash => token_text![":-"],
            Token::Pipe => token_text!["|"],
            Token::PipeGt => token_text!["|>"],
            Token::Bound => token_text!["bound"],
            Token::Inclusion => token_text!["inclusion"],
            Token::Descr => token_text!["descr"],
            Token::Let => token_text!["let"],
            Token::Do => token_text!["do"],
            Token::LBracket => token_text!["["],
            Token::RBracket => token_text!["]"],
            Token::Eq => token_text!["="],
            Token::Bang => token_text!["!"],
            Token::BangEq => token_text!["!="],
            Token::Comma => token_text![","],
            Token::Lt => token_text!["<"],
            Token::Le => token_text!["<="],
            Token::Gt => token_text![">"],
            Token::Ge => token_text!["<="],
            Token::Dot => token_text!["."],
        }
    }
}
