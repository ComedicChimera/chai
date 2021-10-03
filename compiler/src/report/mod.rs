pub mod display;

// TextPosition is used to demarkate a specific position with user source text
// for purposes of error/warning reporting
pub struct TextPosition {
    pub start_line: i32,
    pub start_col: i32,
    pub end_line: i32,
    pub end_col: i32
}

// CompileMessage is a construct used to represent an error during compilation
pub struct CompileMessage {
    pub message: String,
    pub position: TextPosition,
    pub rel_file_path: String
}