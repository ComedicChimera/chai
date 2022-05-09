## Idea
**NOTE:** *As of now, this concept is outdated as the idea of a "module file" has been removed.  However, this approach will be useful when implementing Chai's build tool, `kettle` -- something similar to this will likely be used there.*

Use the module file to handle import resolution for all non-standard packages as a way of providing package "namespacing" and version control.

## Implementation
The module file gets a new array of tables called `deps`.  Each `deps` entry appears as follows:

```
[[deps]]
name = "mydep"
vcs = "git"
url = "github.com/user/mydep"
version = "1.0.0"
```

The `name` field is how the module can be accessed via imports:

```
import mydep
```

The `vcs` can either be an actual VCS (eg. `git`) or it can be `local` designating that the package can be found locally.  For an VCS, the `url` is the path to the repository.  However, for `local`, the `url` is the path relative to the current module to the package.

`version` is used to faciliate version control: much like Go's decentralized package system, the version will be the name of the released version of the package it will look for (for a VCS).

This mechanism will be the only way to access non-standard packages.  The `pub` directory in `modules` will contain the downloaded source code for all installed modules.  The path will be based on the `url`.  

A non-standard module must have a `deps` entry in order to be imported in the project.