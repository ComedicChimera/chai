use std::fs::File;
use std::io::BufReader;
use utf8_chars::{BufReadCharsExt, Chars};

use crate::syntax::token::{Token, TokenData};
use crate::report::{TextPosition, CompileMessage};
use crate::report::display;

// Lexer takes in a file path and acts as a state machine that can be prompted
// repeatedly (via. next_token) to produce a token stream corresponding to the
// source text found in the file.
pub struct Lexer<'a, 'b> {
    fpath: &'a str,
    reader: Chars<'b, BufReader<File>>,
    line: i32,
    col: i32
}

impl<'a, 'b> Lexer<'a, 'b> {
    // new creates a new Lexer for the file at the given path
    pub fn new(fpath: &'a str) -> Lexer<'a, 'b> {
        Lexer {
            fpath: fpath, 
            reader: BufReader::new(File::open(fpath).expect(format!("unable to open file: `{}`", fpath))).chars(),
            line: 1, col: 0
        }
    }

    // next_token retrieves the next token from the Lexer
    pub fn next_token(&mut self) -> Result<Token, CompileMessage> {
        

        Ok(Token{data: TokenData::EndOfFile, position: TextPosition{start_line: 0, start_col: 0, end_line: 0, end_col: 0}})
    }

    // readChar reads a new character from input stream
    fn readChar(&mut self) -> Option<char> {
        let next_res = self.reader.next()?;

        match next_res {
            Ok(c) => {
                self.update_position(c);
                Some(c)
            },
            Err(e) => { 
                display::fatal_error(e.to_string().as_str());
                None
            }
        }
    }

    // update_position updates the position of the Lexer based on a char
    fn update_position(&mut self, c: char) {
        if c == '\n' {
            self.line += 1;
            self.col = 0;
        } else if c == '\t' {
            self.col += 4;
        } else {
            self.col += 1;
        }
    }
}