use std::process::exit;

// fatal_error reports a fatal compilation error -- caused by some form of
// runtime error in the compiler.  These are generally various forms of IO
// errors.  This causes the program to exit immediately.
pub fn fatal_error(s: &str) {
    println!("fatal error: {}", s);
    exit(1);
}