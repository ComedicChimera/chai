use crate::report::{TextPosition};

// Token represents a single lexical element of Chai source text.
pub struct Token {
    pub data: TokenData,
    pub position: TextPosition
}

// TokenData is the actual data stored by the token based on lexical category.
// It is stored in such a manner that pattern matching is possible.
pub enum TokenData {
    Identifer(String),

    EndOfFile
}