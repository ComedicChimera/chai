from dataclasses import dataclass


@dataclass
class Package:
    pkg_id: int
    pkg_name: str
    abs_path: str