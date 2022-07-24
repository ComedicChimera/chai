package wintool

// The table mapping LLVM target architectures names to their VS instance
// architecture names: used to lookup VS instances using `vswhere.exe`.
var llvmArchToVSArch = map[string]string{
	"i386":    "x86.x64",
	"x86_64":  "x86.x64",
	"arm":     "ARM",
	"aarch64": "ARM64",
}

// The table mapping GOARCH values to their VC path host suffixes.
var hostArchToVCHostSuffix = map[string]string{
	"386":   "X86",
	"amd64": "X64",
	"arm":   "X86",
	"arm64": "X86",
}

// The table mapping LLVM architecture names to their VC 15+ subdirectory.
var llvmArchToVS15PlusSubDir = map[string]string{
	"i386":    "x86",
	"x86_64":  "x86.x64",
	"arm":     "arm",
	"aarch64": "arm64",
}

// The table mapping LLVM architecture names to their UCRT bin directory.
var llvmArchToUCRTHostDir = map[string]string{
	"i386":    "x86",
	"x86_64":  "x64",
	"arm":     "arm",
	"aarch64": "adm64",
}
