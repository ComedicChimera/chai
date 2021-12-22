package report

// Message is a generic interface representing all different kinds of possible
// messages that can be reported as errors and warnings. The `display` method is
// private since it should only be called by the reporter directly.
type Message interface {
	// display should invoke the appropriate mechanism to print this message to
	// the console.
	display()

	// isError indicates if this should be counted as an error or a status
	// update/warning. This determines how the reporter should handle the
	// message.
	isError() bool
}

// CompileMessage is a message that is specifically generated by the compiler to
// alert the user of some fault or possible issue with their code.  This is a
// standard compiler error.  These always correspond to a specific position
// within user source code.
type CompileMessage struct {
	Context  *CompilationContext
	Message  string
	Position *TextPosition
	IsError  bool
}

// CompilationContext is the context in which the file is being compiled for
// the purposes of locating and displaying the source file.
type CompilationContext struct {
	ModName     string
	ModAbsPath  string
	FileRelPath string
}

func (cm *CompileMessage) isError() bool {
	return cm.IsError
}

// ModuleMessage is an error or warning that occurs while loading and parsing a
// module. This can be due to some error in the actual source of the module or
// because is irresolvably conflicts with the state of the compiler.
type ModuleMessage struct {
	ModName string
	Message string
	IsError bool
}

func (mm *ModuleMessage) isError() bool {
	return mm.IsError
}

// PackageError is an error that occurs while initializing a package.
type PackageError struct {
	ModRelPath string
	Message    string
}

func (pe *PackageError) isError() bool {
	return true
}
