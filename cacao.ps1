function prebuild {
    $llvm_config_path = "$PSScriptRoot/vendor/llvm-build/Debug/bin/llvm-config.exe"

    set "CGO_CPPFLAGS=$llvm_config_path --cppflags"
    set "CGO_CXXFLAGS=-std=c++14"
    set "CGO_LDFLAGS=$llvm_config_path --ldflags --libs --system-libs all"
    set "CGO_LDFLAGS_ALLOW=-Wl,(-search_paths_first|-headerpad_max_install_names)"
}

function updateLLVM {
    git submodule update
    cmake -DLLVM_ENABLE_BINDINGS=OFF -S vendor/llvm/llvm/ -B vendor/llvm-build/
    MSBuild vendor/llvm-build/LLVM.sln
}

switch ($args[0]) {
    "setup" {
        go env -w GOVCS "llvm.org:svn,private:all,public:git|hg"
        updateLLVM
    }
    "build" {
        prebuild

        cd src
        go build -tags byollvm -o ../bin/chai.exe main.go
        cd ../
    }
    "release" {
        # TODO
    }
    "update" {
        updateLLVM
    }
}