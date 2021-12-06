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

define internal void @build_test.putb(i8 %b) nounwind {
entry:
	%0 = alloca i8
	store i8 %b, i8* %0
	%1 = load i8*, i8** @stdout
	%2 = call i1 @WriteConsoleA(i8* %1, i8* %0, i32 1, i32* null, i8* null)
	ret void
}

define internal void @build_test.putd(i32 %d) nounwind {
entry:
	%0 = add i32 %d, 0
	%1 = trunc i32 %0 to i8
	call void @build_test.putb(i8 %1)
	ret void
}

define internal void @build_test.puti(i32 %a) nounwind {
entry:
	%0 = alloca i32
	store i32 %a, i32* %0
	%1 = load i32, i32* %0
	%2 = icmp slt i32 %1, 0
	br i1 %2, label %bb2, label %bb1

bb1:
	%3 = alloca i32
	store i32 1000000000, i32* %3
	%4 = alloca i1
	store i1 false, i1* %4
	br label %bb3

bb2:
	%5 = load i32, i32* %0
	%6 = sub i32 0, %5
	store i32 %6, i32* %0
	br label %bb1

bb3:
	%7 = load i32, i32* %3
	%8 = icmp sgt i32 %7, 0
	br i1 %8, label %bb4, label %bb5

bb4:
	%9 = load i32, i32* %0
	%10 = load i32, i32* %3
	%11 = srem i32 %9, %10
	%12 = icmp sgt i32 %11, 0
	br i1 %12, label %bb7, label %bb8

bb5:
	ret void

bb6:
	%13 = load i32, i32* %3
	%14 = sdiv i32 %13, 10
	store i32 %14, i32* %3
	br label %bb3

bb7:
	call void @build_test.putd(i32 %11)
	store i1 true, i1* %4
	br label %bb6

bb8:
	%15 = load i32, i32* %3
	%16 = icmp eq i32 %15, 1
	br i1 %16, label %bb11, label %bb10

bb9:
	call void @build_test.putd(i32 0)
	br label %bb6

bb10:
	%17 = load i1, i1* %4
	br label %bb11

bb11:
	%18 = phi i1 [ true, %bb8 ], [ %17, %bb10 ]
	br i1 %18, label %bb9, label %bb6
}

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
	call void @build_test.puti(i32 247)
	%4 = sub i32 0, 45
	call void @build_test.puti(i32 %4)
	call void @build_test.puti(i32 5002481)
	call void @build_test.puti(i32 50)
	ret void
}

declare external win64cc void @ExitProcess(i32 %code)

define external void @_start() nounwind {
entry:
	call void @__init()
	call void @build_test.main()
	call void @ExitProcess(i32 0)
	ret void
}

declare external win64cc i8* @GetStdHandle(i32 %nStdHandle)
