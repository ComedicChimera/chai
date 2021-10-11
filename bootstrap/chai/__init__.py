from dataclasses import dataclass

# The current version number of Chai
CHAI_VERSION = '0.1.0'

# The conventional Chai file extension
CHAI_FILE_EXT = '.chai'

# TextPosition represents a position in user source code.
@dataclass
class TextPosition:
    start_line: int
    start_col: int
    end_line: int
    end_col: int

# ChaiCompileError is an error in user source code.
@dataclass
class ChaiCompileError(Exception):
    rel_path: str
    position: TextPosition
    message: str

    def report(self) -> str:
        return f'{self.rel_path}:{self.position.start_line}:{self.position.start_col}: {self.message}'

# ChaiModuleError is an error loading a module.
@dataclass
class ChaiModuleError(Exception):
    module_name: str
    message: str

    def report(self) -> str:
        return f'error loading module `{self.module_name}`: {self.message}'
