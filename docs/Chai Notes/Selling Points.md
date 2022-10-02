# Selling Points
## Easy to Learn

```
package hello;

import Println from io.std;

func main() {
	Println("Hello, world!");
}
```

## Expressive Type System

```
package exprs;

enum Expr =
	| Add(*Expr, *Expr)
	| Sub(*Expr, *Expr)
	| Mul(*Expr, *Expr)
	| Div(*Expr, *Expr)
	| Value(i64)
	;

func evaluate(expr: *Expr) Option<i64> =
	match *expr {
		case .Add(a, b) => evaluate(a)? + evaluate(b)?,
		case .Sub(a, b) => evaluate(a)? - evaluate(b)?,
		case .Mul(a, b) => evaluate(a)? * evaluate(b)?,
		case .Div(a, b) => 
			if let b <- evaluate(b); b != 0 => evaluate(a)? / b 
			else => None,
		case .Value(x) => Some(x)
	};
```

## Robust Error Handling

```
package fileio;

import File, FileMode from io.fs;

func countFileChars(filePath: string) Result<Map<rune, i64>, IOError> {
	let m: Map<rune, i64> = {};
	
	let file <- File.Open(filePath, FileMode.Read);
	defer file.Close();

	for c <- file.Runes() {
		m[c] = m.TryGet(c).OrElse(0) + 1;
	}

	return Ok(m);
}
```

## Powerful Generics

```
package numbers;

import Println from io.std;

func categorizeNumber<N: Real>(n: N) {
	match type n {
		case SInt:
			Println("signed integer");
		case UInt:
			Println("unsigned integer");
		case Float:
			Println("floating point");
	}
}
```

## Lightweight Concurrency

**TODO**

## Simple C Binding

```
package cbinding;

#csrc `
#include <stdio.h>

#include <cchai.h>

void helloFromC(const char* str, chai_i64 n) {
	for (int i = 0; i < n; i++)
		printf("C Says: %s\n", str);
}
`

import cmixin;
import C;

func cPrint(n: i64) {
	C.helloFromC(cmixin.CString("Hello!"), n);
}
```