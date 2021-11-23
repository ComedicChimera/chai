%string = type { i8*, i32 }

@__strlit.0 = global [13 x i8] c"Hello, world!"

declare external dllimport win64cc i8* @"GetStdHandle%1"(i32 %nStdHandle)

define internal i8* @build_test.get_stdout() nounwind {
entry:
	%0 = sub i32 0, 11
	%1 = call i8* @"GetStdHandle%1"(i32 %0)
	ret i8* %1
}

declare external dllimport win64cc i1 @"WriteConsole%5"(i8* %hConsoleOutput, i8* %lpBuffer, i32 %nNumberOfCharsToWrite, i32* %lpNumberOfCharsWritten, i8* %lpReserved)

define internal void @build_test.puts(i8* %h, %string* %str) nounwind {
entry:
	%0 = bitcast %string* %str to i8**
	%1 = load i8*, i8** %0
	%2 = getelementptr %string, %string* %str, i32 0, i32 1
	%3 = load i32, i32* %2
	%4 = call i1 @"WriteConsole%5"(i8* %h, i8* %1, i32 %3, i32* null, i8* null)
	ret void
}

define internal void @build_test.main() nounwind {
entry:
	%0 = call i8* @build_test.get_stdout()
	%1 = alloca %string
	%2 = getelementptr %string, %string* %1, i32 0, i32 0
	%3 = bitcast [13 x i8]* @__strlit.0 to i8*
	store i8* %3, i8** %2
	%4 = getelementptr %string, %string* %1, i32 0, i32 1
	store i32 13, i32* %4
	call void @build_test.puts(i8* %0, %string* %1)
	ret void
}

declare external dllimport win64cc void @"ExitProcess%1"(i32 %code)

define external void @_start() nounwind {
entry:
	call void @build_test.main()
	call void @"ExitProcess%1"(i32 0)
	ret void
}
