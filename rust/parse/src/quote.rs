// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

use anyhow::anyhow;

// unesc maps single-letter chars following \ to their actual values.
fn unesc(c: &char) -> u8 {
    match c {
        'a' => b'\x07',
        'b' => b'\x08',
        'f' => b'\x0C',
        'n' => b'\x0A',
        'r' => b'\x0D',
        't' => b'\x09',
        'v' => b'\x0B',
        '\\' => b'\\',
        '\'' => b'\'',
        '"' => b'"',
        _ => unreachable!(),
    }
}

pub enum DecodedSequence {
    String(String),
    Bytes(Vec<u8>),
}

// unquote unquotes the quoted string, returning the actual
// string value, whether the original was triple-quoted,
// whether it was a byte string, and an error describing invalid input.
// TODO
pub fn unquote(quoted_str: &str) -> anyhow::Result<DecodedSequence> {
    let mut quoted = quoted_str;
    let mut is_byte = false;
    // Check for raw prefix: means don't interpret the inner \.
    let mut raw = false;
    if quoted.starts_with('r') {
        raw = true;
        quoted = &quoted[1..]
    }
    // Check for bytes prefix.
    if quoted.starts_with('b') {
        is_byte = true;
        quoted = &quoted[1..]
    }
    if quoted.len() < 2 {
        return Err(anyhow!("string literal too short"));
    }

    let first = quoted.chars().next().unwrap();
    if first != '"' && first != '\'' || first != quoted.chars().last().unwrap() {
        return Err(anyhow!("string literal {quoted} has invalid quotes"));
    }

    // Check for triple quoted string.
    let quote = quoted.chars().next().unwrap();
    if quoted.len() >= 6
        && quoted.chars().nth(1).unwrap() == quote
        && quoted.chars().nth(2).unwrap() == quote
        && quoted[..3] == quoted[quoted.len() - 3..]
    {
        quoted = &quoted[3..quoted.len() - 3]
    } else {
        quoted = &quoted[1..quoted.len() - 1]
    }

    // Now quoted is the quoted data, but no quotes.
    // If we're in raw mode or there are no escapes or
    // carriage returns, we're done.
    let unquote_chars = if raw { "\r" } else { "\\\r" };
    if !quoted.chars().any(|x| unquote_chars.contains(x)) {
        // TODO
        return if is_byte {
            Ok(DecodedSequence::Bytes(quoted.into()))
        } else {
            Ok(DecodedSequence::String(quoted.to_string()))
        };
    }

    // Otherwise process quoted string.
    // Each iteration processes one escape sequence along with the
    // plain text leading up to it.
    let mut buf: Vec<u8> = vec![];
    loop {
        // Remove prefix before escape sequence.
        match quoted.chars().position(|c| unquote_chars.contains(c)) {
            Some(i) => {
                (quoted[..i]).chars().for_each(|c| buf.push(c as u8));
                quoted = &quoted[i..];
            }
            _ => {
                quoted.chars().for_each(|c| buf.push(c as u8));
                break;
            }
        }

        // Process carriage return.
        if quoted.starts_with('\r') {
            buf.push(b'\n');
            quoted = if quoted.len() > 1 && quoted.chars().nth(1).unwrap() == '\n' {
                &quoted[2..]
            } else {
                &quoted[1..]
            };
            continue;
        }

        // Process escape sequence.
        if quoted.len() == 1 {
            return Err(anyhow!("truncated escape sequence \\"));
        }

        match quoted.chars().nth(1) {
            Some('\n') =>
            // Ignore the escape and the line break.
            {
                quoted = &quoted[2..]
            }

            Some('a' | 'b' | 'f' | 'n' | 'r' | 't' | 'v' | '\\' | '\'' | '"') => {
                // One-char escape.
                // Escapes are allowed for both kinds of quotation
                // mark, not just the kind in use.
                buf.push(unesc(&quoted.chars().nth(1).unwrap()));
                quoted = &quoted[2..]
            }

            Some('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7') => {
                // Octal escape, up to 3 digits, \OOO.
                let mut n = quoted.chars().nth(1).unwrap().to_digit(8).unwrap();
                quoted = &quoted[2..];
                for i in 1..3 {
                    if quoted.is_empty()
                        || quoted.chars().nth(i).unwrap() < '0'
                        || '7' < quoted.chars().next().unwrap()
                    {
                        break;
                    }
                    n = n * 8 + quoted.chars().next().unwrap().to_digit(8).unwrap();
                    quoted = &quoted[1..];
                }
                if !is_byte && n > 127 {
                    return Err(anyhow!(
                        "non-ASCII octal escape \\{n:o} (use \\u{n:04x} for the UTF-8 encoding of U+{n:04x})",
                    ));
                }
                if n >= 256 {
                    // NOTE: Python silently discards the high bit,
                    // so that '\541' == '\141' == 'a'.
                    // Let's see if we can avoid doing that in BUILD files.
                    return Err(anyhow!("invalid escape sequence \\{n:03}o"));
                }
                buf.push(0 /* char::from_u32(n).unwrap() */) // TODO
            }
            Some('x') => {
                // Hexadecimal escape, exactly 2 digits, \xXX. [0-127]
                if quoted.len() < 4 {
                    return Err(anyhow!("truncated escape sequence {quoted}"));
                }
                match u32::from_str_radix(&quoted[2..4], 16) {
                    Ok(n) => {
                        if !is_byte && n > 127 {
                            return Err(anyhow!(
                                "non-ASCII hex escape {} (use \\u{n:04X} for the UTF-8 encoding of U+{n:04x})",
                                &quoted[..4],
                            ));
                        }
                        let decoded_ch = char::from_u32(n);
                        if decoded_ch.is_none() {
                            return Err(anyhow!("invalid Unicode code point U{n:04x}"));
                        }
                        let mut tmp: [u8; 4] = [0; 4];
                        let encoded = char::encode_utf8(decoded_ch.unwrap(), &mut tmp);
                        encoded.as_bytes().iter().for_each(|b| buf.push(*b));
                        quoted = &quoted[4..]
                    }
                    _ => return Err(anyhow!("could not parse unicode codepoint {quoted}")),
                }
            }
            Some('u' | 'U') => {
                // Unicode code point, 4 (\uXXXX) or 8 (\UXXXXXXXX) hex digits.
                let mut sz = 6;
                if quoted.chars().nth(1).unwrap() == 'U' {
                    sz = 10
                }
                if quoted.len() < sz {
                    return Err(anyhow!("truncated escape sequence {quoted}"));
                }

                match u32::from_str_radix(&quoted[2..sz], 16) {
                    Ok(n) => {
                        // As in Rust, surrogates are disallowed.
                        if (0xd800u32..0xe000u32).contains(&n) {
                            return Err(anyhow!("invalid Unicode code point U{n:04x}"));
                        }
                        if n > 0x10FFFFu32 {
                            return Err(anyhow!(
                                "code point out of range: {} (max \\U{:08x})",
                                &quoted[..sz],
                                n
                            ));
                        }
                        let decoded_ch = char::from_u32(n);
                        if decoded_ch.is_none() {
                            return Err(anyhow!("invalid Unicode code point U{n:04x}"));
                        }
                        let mut tmp: [u8; 8] = [0; 8];
                        let encoded = char::encode_utf8(decoded_ch.unwrap(), &mut tmp); // from_u32(n).unwrap());
                        encoded.as_bytes().iter().for_each(|b| buf.push(*b));
                        quoted = &quoted[sz..]
                    }
                    _ => return Err(anyhow!("failed to parse unicode code point: {quoted}")),
                }
            }
            _ =>
            // In Starlark, like Go, a backslash must escape something.
            // (Python still treats unnecessary backslashes literally,
            // but since 3.6 has emitted a deprecation warning.)
            {
                return Err(anyhow!("invalid escape sequence \\{quoted}"));
            }
        }
    }

    if is_byte {
        return Ok(DecodedSequence::Bytes(buf));
    }

    let buf_utf8 = String::from_utf8(buf)?;
    Ok(DecodedSequence::String(buf_utf8))
}

// Quote returns a Starlark literal that denotes s.
// If b, it returns a bytes literal.
// Not handling invalid literals.
pub fn quote(s: &str) -> String {
    let mut buf = "\"".to_string();
    for c in s.chars() {
        if c == '\'' {
            buf.push(c);
            continue;
        }
        c.escape_default().for_each(|c| buf.push(c));
    }
    buf.push('"');
    buf
}
