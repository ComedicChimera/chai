// Top level module definitions
mod syntax;
mod report;

use syntax::lexer::{Lexer};

fn main() {
    let lexer = Lexer::new("../tests/scan_test.chai");
    println!("Hello, world!");
}
