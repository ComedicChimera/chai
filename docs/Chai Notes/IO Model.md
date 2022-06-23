# IO Model

This file documents the prospective type/functionality hierachy of Chai's IO model.

**Note:** All definitions below may exclude virtual (utility) methods defined on type classes.

## Reader and Writer

There are two fundamental classes of Chai's IO model: `Reader` and `Writer`.  They are defined in `io` and have the following definitions:

```
class Reader<E> R
	# Reads a single byte from the input stream.
	method read() Result<u8, E> end
end

class Writer<E> W
	# Write a single byte to the output stream.
	method write(b: u8) Result<(), E> end
end
```

These act as the most basic constructs in IO model: anything that wants to interact with Chai's IO layer needs to implement these two classes.

Since it is common for things to be both readable and writeable, there is another type class known as `Stream` which requires something to implement both methods

```
class Stream<E> S is Reader<E>, Writer<E> end
```

**Note:** This class syntax denotes that the class simply a combination of its inherited classes.

### Specialization

From these base class, many specialized readers and writers can be built up.  For example, if I wanted a buffered, encoded, and formatted reader, all I would need to do is implement all of the classes corresponding to those readers on my type of choice.  

Many of the specialized classes only add one or two requires methods (in some cases, they require none and just add functionality), and because Chai supports multiple inheritance at different points in the hierarchy, it is definitely possible to create fairly complex readers and writers easily.

### Aside: Text Encoding

By default, Chai assumes all text is UTF-8 encoded.  However, if you want/need to use an alternate encoding, there will be a separate back that wraps readers and writers a byte level to allow you to use different encodings.

## Buffered Reader and Writer
Many core IO constructs implement buffering in some form.  This is represented by the type classes `BufferedReader`, `BufferedWriter`, and `BufferedStream` which are all also defined in `io` and have the following definitions:

```
class BufferedReader<E> BR is Reader<E>
	# Fills the input buffer from the readable device.
	# This is often analagous to getting a single line.
	method fill() Result<(), E> end
end

class BufferedWriter<E> BW is Writer<E>
	# Flushes the output buffer to the writeable device.
	method flush() Result<(), E> end
end

class BufferedStream<E> BS is BufferedReader<E>, BufferedWriter<E> end
```

Note that the buffered IO classes should implement their `read` and `write` methods to appropriately interact with their internal buffers (possibly making use of `fill` and `flush` respectively).   

## Text Streams

`io` also provides tools for working with text, specifically UTF-8 encoded text (ie. Chai strings).

These tools come in the form of the `TextReader`, `TextWriter`, and `TextStream` type classes.  Their definitions are as follows:

```
class TextReader<E> is Reader<E>
	method read_rune() Result<rune, E>
		...
	end
	
	method read_string(len: isize) Result<string, E>
		...
	end
end

class TextWriter<E> is Writer<E>
	method write_rune(r: rune) Result<isize, E>
		...
	end
	
	method write_string(s: string) Result<isize, E>
		...
	end
end

class TextStream<E> is TextReader<E>, TextWriter<E> end
```

As you can see, these type classes only provide additional virtual methods on top of reader and writer: they are used to add functionality to readers and writers that deal with text.

## Formatted IO
Formatted IO is facilitated by several key type classes and constructs which are defined in the `io.fmt` sub-package.  The key players are: `FormattedReader`, `FormattedWriter`, `FormattedStream`, `Format`, and `Formatter`.  

The `FormattedReader`, `FormattedWriter`, and `FormattedStream` classes are used to mark something as formattable.  They don't introduce any new methods on their own, but rather add key methods to make readers and writers formattable.  Here are their definitions.

```
class FormattedReader<E> FR is TextReader<E>
	method scanf<...F: Format>(format: string, ...args: *F) Result<unit, E>
		...
	end
end

class FormattedWriter<E> FW is TextWriter<E>
	method printf<...F: Format>(format: string, ...args: F) Result<unit, E>
		...	
	end
end

class FormattedStream<E> is FormattedReader<E> FormattedWriter<E> end
```

The `Format` type class is used to make a type *formattable*.  It has the following definition:

```
class Format F is Show
	method print<T, E, FW: FormattedWriter>(w: FW<T, E>, arg: string) Option<E> end
	
	def scan<T, E, FR: FormattedReader>(r: FR<T, E>, arg: string) Result<F, E> end
	
	method show() string
		...
	end
end
```

Notice that all types which implement `Format` also implement `Show` by default.

Finally, there is `Formatter` which is a construct that acts as the formatting state machine.  It parses format strings and determine what should be done.  Here is its definition:

```
closed enum FormatAction = 
	| Text(string)
	| Value(u64, string)
	| Done
	
enum FormatError = ...

record Formatter = {
	...
}

space for Formatter
	def new(format: string) Formatter
		...
	end
	
	method action() Result<FormatAction, FormatError>
		...
	end
end
```

## Standard IO

The standard IO devices in Chai are stored in the package `io.std` and are all UTF-8 encoded, buffered, and formatted.  Here is an example subspace signature of `Stdin` IO device.

```
space for Stdin is (
	BufferedReader<OSError>,
	FormattedReader<OSError>
)
	...
end
```

The following functions are exported from `io.std`.

```
# Prints its arguments joined by spaces to stdout.
def print<...S: Show>(...args: S) end

# Prints its arguments joined by spaces ended with a newline to stdout.
def println<...S: Show>(...args: S) end

# Prints its arguments formatted into the format string to stdout.
def printf<...F: Format>(format: string, ...args: F) end

# Flushes the stdout buffer.
def flush() end

# Prints its arguments joined by spaces to stderr.
def error<...S: Show>(...args: S) end

# Prints its arguments joined by spaces ended with a newline to stderr.
def errorln<...S: Show>(...args: S) end

# Prints its arguments formatted into the format string to stderr.
def errorf<...F: Format>(format: string, ...args: F) end

# Reads a single character from stdin.
def scanc() rune end

# Reads a single token from stdin.
def scan() string end

# Reads a single line from stdin.
def scanln() string end

# Reads formatted input from stdin.
def scanf<...F: Format>(format: string, ...args: *F) Result<unit, EIOError>

# Prints its arguments joined by spaces to stderr and exits the program.
def fatal<...S: Show>(...args: S) end

# Prints its arguments joined by spaces ended with a newline to stderr
# and exits the program.
def fatalln<...S: Show>(...args: S) end

# Prints its arguments formatted into the format string to stderr
# and exits the program.
def fatalf<...F: Format>(format: string, ...args: F) end
```

**Note:** In the current model, most standard IO functions panic if they fail to perform their given task (with the exception of `scanf`).  

## Files

Files are always instances of `BufferedStream<OSError>` and are dealt with at a byte-level by default. 

Files are opened and created using the `File.open` function.  

Files are also instances of `FormattedStream` and thus can be used for reading and writing text as well.