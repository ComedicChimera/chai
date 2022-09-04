# String Essentials

In this chapter, we will take a look at how Chai represents text using runes and
strings.  This will be a basic overview of some of the essential concepts for
working with strings.  We will revisit strings with a different lens several
times in later chapters.

**Table of Contents**

- [Runes and Rune Literals](#runes)
- [String Literals](#string-lits)
- [Operating on Strings](#string-ops)
- [Raw Strings](#raw-strings)

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

## <a name="string-lits"/> String Literals

## <a name="string-ops"/> Operating on Strings

## <a name="raw-strings"/> Raw Strings