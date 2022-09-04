# Prelude Implementation
This file describes the implementation of the prelude.

## The `core` Package
All the "built-in" functions and types are implemented in `core`.  

This package is added to every package's import table (except its own), and its namespace is included in any symbol lookups from within the walker and resolver (without being explicitly added to any of the file import tables).  

The backend should automatically add all the symbols imported from this package as external definitions in the file it is compiling by virtue of it being present in the package's import table.  Importing `core` explicitly should probably be disabled: if a symbol import is used, we may have problems (maybe, we allow symbols in `core` to be shadowed by any package/file level symbols?)

## The `runtime` Package
All runtime support functions and types are defined in `runtime`.  

We can access them by simply including `runtime` in every package's import table (except its own).  However, it is not added to any file import tables and not made visible during look-ups.  

This should (as with `core`) ensure that the backend generates proper external definitions for `runtime` so that the support libraries can be used (although, it should be noted that we need to make sure the compiler marks the `runtime` symbol's it uses appropriately). 

Furthermore, if any package attempts to import `runtime` explicitly, this shouldn't cause any problems since the compiler will just use the definition it already has in the package import table, and any duplications between explicit and implicit usages should just compile as if it were explicitly used twice.

## Aside: Dependency Graph
Both `core` and `runtime` will need to be explicitly added to the dependency graph so that they are compiled with the project.  References to them will also need to be passed to multiple stages of the compiler for look-ups (eg. the parser, the walker, etc.)