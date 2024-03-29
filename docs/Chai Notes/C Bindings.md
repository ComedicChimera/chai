# C Bindings
C bindings are absolute must for any compiled language.  They allow for quick library development and automatically make Chai compatible with a bunch of old toolchains that are industry standards.

## Accessing C Functions and Types
Use a special directive:

```
$ csrc `
	#include <stdio.h>
	
	#include "llvm/Analysis.h"
	
	void c_print(int n) {
		printf("%d\n", n);
	}
`
```

Then, you can use the `cmixin` module to pass data between Chai and C.  To access the definitions, add the special `import C` to your source file.

```
import cmixin;
import C;

func main() {
	C.c_print(10);
	
	C.printf(cmixin.CString("%d %c\n"), cmixin.VaArgs(12, 'a' as cmixin.char));
}
```

## Compiler Configuration
**NOTE:** *These options are not really compatible with tools like C-Make and MSBuild (ie. libraries which require a more complex build procedure) so we will need to rework our approach to compiler configuration.*

Special compiler Flags:
```
-cc, --ccompiler   specify the path to the C compiler to use
-ci, --cincludes   specify include paths to pass to the C compiler
-cf, --cflags      specify flags to pass to your C compiler
```

The Chai compiler will only use the C compiler to generate object files which it will combine into your executable.  Thus, all extra linking (eg. libraries like `glfw3.lib`, `openal.lib`, etc) can simply be added using standard Chai linking options.



