# Standard IO

```
def print[...S: Show](...s: S)
def println[..S: Show](...s: S)
def printf[...T](fmt: string, ...args: T)

def error[...S: Show](...s: S)
def errorln[...S: Show](...s: S)
def errorf[...T](fmt: string, ...args: T)

def scan() Result[rune, Error]
def scanln() Result[string, Error]
def scanf[...T](fmt: string, ...args: &T) Result[(), Error]

def flush() bool
