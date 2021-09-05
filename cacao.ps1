if ($args.Length -lt 1) {
    echo "expected a subcommand"
    exit
}

switch ($args[0]) {
    "build" {
        cmake -S . -B out
        MSBuild out/chai.sln
    }
    "run" {
        out/Debug/chai.exe
    }
}