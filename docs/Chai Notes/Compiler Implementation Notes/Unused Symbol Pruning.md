# Unused Symbol Pruning

We need to decide whether to prune an unused symbol based on where it is used and the output mode so as to ensure the ABI of Chai object files matches up with user's expectations. 

Note that at the ABI level, we make the distinction between **exported** and **internal** symbols rather than public and private: a symbol which exported may not always be marked public (eg. `main`)

The general rule is that any symbol/definition which is not exported or referenced by an exported symbol/definition (possibly through multiple layers -- ie. indirectly) will be pruned.  However, the rules for what is considered "exported" may vary.

## Export Rules
These rules outline which symbols are considered exported in different output modes.

1.  If the output mode is an executable, then `_start` and any symbols in the root package marked `pub` will be exported.
2.  If the output mode is a static library, then only symbols in the root package marked `pub` or symbols annotated with `@static_export` in any package will be exported.
3.  If the output mode is a dynamic library, then only the DLL's main function (TBD on how to indicate this), any symbols in the root package marked `pub`, or symbols annotated with `@dynamic_export` in any package will be exported.
4.  If any other output mode is selected, then any symbol which is marked `pub` or annotated with either `@static_export` or `@dynamic_export`.  Furthermore, `_start`, and the indicated DLL main function (if any) will also be exported.
    
**TODO:** Should having the `--fPIC` option specified also be criteria of exporting `@dynamic_export` in case 4?