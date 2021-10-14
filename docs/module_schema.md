# Module Schema

This file documents the various options (required and optional) that modules can
accept: ie. the module's schema.

*Note: other supported operating systems and architectures will be added in the future.*

## Table of Contents

- [Top Level Configuration](#top-conf)
- [Build Profiles](#build-prof)
  * [Supported Operating System](#supported-os)
  * [Supported Architectures](#supported-arch)
  * [Output Formats](#output-fmt)
  * [Profile Selection](#profile-selection)
  * [Profile Elision](#profile-elision)

## <a name="top-conf"> Top Level Configuration

The following are all the top level module configuration fields:

| Field | Type | Purpose | Required |
| ----- | ---- | ------- | -------- |
| `name` | string | specifiy the module name | Y |
| `version` | string | version specifies the module's version (mostly for dependency version control) | N |
| `chai-version` | string | specify the Chai version the module was created on in the form: `Major.Minor.Build` | Y |
| `caching` | bool | enable compilation caching | N, default = `false` |
| `enable-profile-elision` | bool | enables [profile elision](#profile-elision) | N, default = `true` |

## <a name="build-prof"> Build Profiles

Build profiles are a mechanism for specifying compilation options such as target
operating system, link objects, etc.  They are essentially just platform
specific configurations for a project.

Build profiles are specified in an array of tables called `profiles`.

The primary options for each build profile entry are:

| Option | Type | Purpose | Required |
| ------ | ---- | ------- | -------- |
| `name` | string | the name for the profile (as it can be specified from the CLI) | Y |
| `target-os` | string | the target operating system (see [supported OSs](#supported-os)) | Y |
| `target-arch` | string | the target architecture (see [supported architectures](#supported-arch) | Y |
| `debug` | bool | indicates whether the profile should be built in debug mode + debug info | Y |
| `format` | string | indicates the output format (see [output formats](#output-fmt)) | Y |
| `output-dir` | string | the path to write output (directory of file depending on output format) | Y |
| `link-objects` | array of strings | array of paths to additional objects or static libraries to link | N |
| `default` | bool | flag indicates that this profile is the [default](#profile-selection) | N, default = `false` |
| `base-only` | bool | flag indicates that this profile should only be considered if it is being selected as the [base profile](#profile-selection) | N, default = `false` |

### <a name="supported-os"> Supported Operating Systems

| OS | Alias |
| -- | ----- |
| Windows 10 | `windows` |

### <a name="supported-arch"> Supported Architectures

| Arch | Alias |
| ---- | ----- |
| x64 | `amd64` |
| x86 | `i386` |

### <a name="output-fmt"> Output Formats

The output format specifies what type of output should be produced by the compiler.

| Name | Alias | Output Kind |
| ---- | ----- | ----------- |
| Executable | `bin` | Single File |
| Library | `lib` | Single File |
| Objects | `obj` | Directory of Object Files |
| Assembly | `asm` | Directory of Assembly Files |
| LLVM IR | `llvm` | Directory of LLVM IR source files |
| Chai MIR | `mir` | Directory of "textualized" Chai MIR files |

*Note: for all output formats that produce multiple files, the convention is one file per package.*

### <a name="profile-selection"> Profile Selection

The **base profile** is the profile of the root module being built.

Chai determines the profile to build based on (in order of precedence):

1. Specified input profile (with `-p` or `--profile` compiler argument)
  - This only determines the base profile
2. Single profile matching base profile
  - Profile shares same: target os, target architecture, and debug flag
  - The profile's output format is overridden by the base profile's output
    format
  - The profile's output directory is also overidden
  - If the profile being determined is the base profile, then the target os and
    arch are chosen based on the current system and the debug profile is
    preferred.
3. Matching default profile
  - This occurs when there are multiple matching profiles
  - The profile with the `default` option is selected: called the **default
    profile**
  - If multiple matching profiles are marked as default, the first in order of
    definition is selected.
  - Similarly, if there is no default profile and multiple matching profiles,
    the first in order of definition is selected.

If there is no matching profile, then refer to [profile elision](#profile-elision)

### <a name="profile-elision"> Profile Elision

It is common for sub-modules that don't introduce any additional linking
dependencies to not provide a build profile (or to provide only profiles that
are `base-only`).  This is to allow them to simple assume the profile of the
base module since Chai code is generally cross platform with little additional
configuration.

**Profile elision** is the mechanism that allows for this lack of build profile.
When enabled via the top level option, the module is allowed to assume the
configuration of the base profile is no matching profile is encountered.

If profile elision is not enabled and no matching profile can be found, then an
error is thrown during compilation.
