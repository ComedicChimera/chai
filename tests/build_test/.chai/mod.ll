%string = type { i8*, i32 }

@stdout = global i8* null
@__strlit.0 = global [13 x i8] c"Hello, world!"

define void @__init() {
entry:
	%0 = sub i32 0, 11
	%1 = call i8* @GetStdHandle(i32 %0)
	store i8* %1, i8** @stdout
	ret void
}

declare external win64cc i1 @WriteConsoleA(i8* %hConsoleOutput, i8* %lpBuffer, i32 %nNumberOfCharsToWrite, i32* %lpNumberOfCharsWritten, i8* %lpReserved)

define internal void @build_test.puts(%string* %str) nounwind {
entry:
	%0 = load i8*, i8** @stdout
	%1 = bitcast %string* %str to i8**
	%2 = load i8*, i8** %1
	%3 = getelementptr %string, %string* %str, i32 0, i32 1
	%4 = load i32, i32* %3
	%5 = call i1 @WriteConsoleA(i8* %0, i8* %2, i32 %4, i32* null, i8* null)
	ret void
}

define internal void @build_test.main() nounwind {
entry:
	%0 = alloca %string
	%1 = getelementptr %string, %string* %0, i32 0, i32 0
	%2 = bitcast [13 x i8]* @__strlit.0 to i8*
	store i8* %2, i8** %1
	%3 = getelementptr %string, %string* %0, i32 0, i32 1
	store i32 13, i32* %3
	call void @build_test.puts(%string* %0)
	ret void
}

declare dllimport win64cc void @"ExitProcess%1"(i32 %code)

define external void @_start() nounwind {
entry:
	call void @__init()
	call void @build_test.main()
	call void @"ExitProcess%1"(i32 0)
	ret void
}

declare external win64cc i8* @GetStdHandle(i32 %nStdHandle)
