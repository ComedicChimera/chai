# String Essentials

In this chapter, we will take a look at how Chai represents text using runes and
strings.  This will be a basic overview of some of the essential concepts for
working with strings.  We will revisit strings with a different lens several
times in later chapters.

**Table of Contents**

- [Runes and Rune Literals](#runes)
- [String Literals](#string-lits)
- [Raw Strings](#raw-strings)
- [Exercises](#exercises)

## <a name="runes"/> Runes and Rune Literals

A **rune** represents a single Unicode character of text.  Runes use the type
label `rune` and are essentially just numbers.  More specifically, a rune's
value corresponds to the numeric value of the Unicode code point for the
character being represented.

**Rune literals** are used to denote rune values without using their numeric
value explicitly.  They consist of a single character enclosed inside single
quotes.

    'A'  // a rune literal corresponding to the character A

Rune literals can also contain more complex Unicode characters:

    '£'
    'ý'
    '♠'

Rune literals cannot be empty and cannot contain a newlines or carriage returns. 

    ''  // ERROR!
    '
    '   // ERROR!

This restriction raises the obvious question of how to we denote such special
characters so that we can use their values in code.  

The solution to this problem is something called an **escape sequence**. Escape
sequences are special patterns which are used to denote special characters. They
always being with a backslash followed by the content of the sequence.  Here is
a table containing some commonly used ones:

| Sequence | Meaning |
| -------- | ------- |
| `\n` | Newline |
| `\r` | Carriage Return |
| `\t` | Tab |
| `\b` | Backspace |
| `\0` | Null Terminator |
| `\\` | Backslash |
| `\'` | Single Quote |
| `\"` | Double Quote

So, if we wanted to denote the newline character in a rune literal, we would
write:

    '\n'

Notice that we can use escape sequences inside rune literals even through they
are more than a single character.  This is because Chai counts all escape
sequences as a single character whenever they are encountered.

Another important point is that, as shown in the table above, we can use escape
sequences to denote things like backslashes and single quotes which otherwise can't
occur on their own in rune literals since they have special meanings:

    '\\'  // Rune literal for backslash
    '\''  // Rune literal for single quote

There are many escape sequences that we haven't covered here, but these should
be sufficient for this tutorial.  A complete account of all possible escape
sequences can be found at *insert link*.

The type label used for runes is `rune`.

    let r: rune = 'a';

Runes can be added and compared just like numbers and can be used in arithmetic
expressions with numbers:

    'a' < 'z'  // => true
    2 + '0'    // => '2'

> This is possible because rune literals are actually just integers.

## <a name="string-lits"/> Strings and String Literals

A **string** is a sequence of Unicode characters representing text.  In Chai,
strings are represented as a sequence of UTF-8 encoded bytes.

A **string literal** is used to represent a constant string value.
Semantically, string literals can be thought of as composed as a series of rune
literals denoting each of the characters that make up the string.

String literals are written as a simple text sequence enclosed in double quotes:

    "Hello, world!"

String literals cannot contain newlines.  If you want to write a string literal
to denote text the contains a newline, you need to use the newline escape code:

    "Line 1\nLine2"

Notice that strings can contain escape codes just like rune literals.  

Strings also cannot contain unescaped double quotes, but can contain single
quotes:

    "\"What's up?\", said Bob."

The type label used for strings is `string`.

    let name: string = "Joe";

Strings can be **concatenated** using the `+` operator.  Concatenation creates
a new string by placing two strings one after the other:

    let firstName = "Bob";
    let lastName = "Jones";

    let fullName = first_name + " " + last_name;  // fullName = "Bob Jones"

No other arithmetic operations are implemented for strings by default.

However, strings can be compared using all the native comparison operators.
`==` and `!=` return whether two strings are equal.  `<`, `>`, `<=, and `>=` all
compare strings lexicographically.

    "abc" == "ABC"    // => false
    "test" != "tset"  // => true
    "ab" < "bc"       // => true
    "123" >= "one"    // => false 

One final thing worth mentioning that we know a little bit more about strings is
another standard IO function: `Scanln`.  `Scanln` allows us to read a string in
from the user.

We can use `Scanln` to write a simple `echo` program.

    package echo;

    import Println, Scanln from io.std;

    func main() {
        let input = Scanln();
        Println("You said", input);
    }

## <a name="raw-strings"/> Raw String Literals

Chai actually supports two kinds of string literals.  The first kind are called
**standard string literals** are the string literals we have considered so far.

The second kind of string literal is called a **raw string literal**.  Raw string
literals are denoted as text enclosed inside backticks:

    `Hello, there!`

Unlike standard string literals, raw string literals can contain newlines:

    `Line 1
    Line 2`

This makes them convenient for denoting long multi-line strings such as
paragraphs or source code.

In addition, raw string literals also don't support regular escape codes.  This
means you can backslashes without any escaping in raw string literals which them
convenient for denoting regular expressions:

    `[a-zA-Z_]\w+`

These is one exception to this: a backslash is used to escape backticks inside
raw string literals.

    `undefined symbol: \`name\``

This also means that backslashes can escape other backslashes.  This means that
if you want to denote a double backslash in a raw string, you need to write:

    `\\\\`

Outside of these two special cases, any text or escape sequence can be used
inside a raw string literal.

## <a name="exercises"/> Exercises

1. Write a program that prompts the user to their first and last name and then
   prints out their full name.