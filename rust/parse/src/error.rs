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

use thiserror::Error;

use crate::token::Token;

#[derive(Error, Debug)]
pub enum ParseError {
    #[error("{0}: expected `{1}` got `{2}`")]
    Unexpected(ErrorContext, Token, Token),
}

#[derive(Error, Debug)]
pub enum ScanError {
    #[error("{0}: incomplete UTF8")]
    IncompleteUtf8(ErrorContext),

    #[error("{0}: unexpected character `{1}`")]
    Unexpected(ErrorContext, char),
}

// Represents an error encountered during parsing.
pub struct ErrorContext {
    pub path: String,
    pub line: usize,
    pub col: usize,
    pub start_of_line: usize,
}

impl std::fmt::Display for ErrorContext {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}:{}:{}", self.path, self.line, self.col)
    }
}

impl std::fmt::Debug for ErrorContext {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}:{}:{}", self.path, self.line, self.col)
    }
}


