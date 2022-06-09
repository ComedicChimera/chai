from typing import Tuple, List

# The debug name of this Chai compiler.
CHAI_NAME = 'pybs-chaic'

# The current Chai version.
CHAI_VERSION = '0.3.0'

# The target pointer size in bytes.
POINTER_SIZE = 8

def merge_metaclasses(*cls_list: List):
    '''
    Returns a merged meta class that is a child of all the metaclasses of the
    given classes.

    Params
    ------
    *cls_list: List
        The list of classes whose metaclasses to merge.
    '''

    class MergedMeta(*[type(x) for x in cls_list]):
        pass

    return MergedMeta

def trim_int_lit(value: str) -> Tuple[str, int, bool, bool]:
    '''
    Returns the integer literal value trimmed of all prefixes and suffixes along
    with all the information indicated by those affixes.

    Params
    ------
    value: str
        The literal int value to trim.
    
    Returns
    -------
    (trimmed_value: str, base: int, unsigned: bool, long: bool)
        The trimmed value and data extracted from the literal.
    '''

    value = value.replace('_', '')

    if len(value) > 2:
        match value[:2]:
            case '0b':
                base = 2
                value = value[2:]
            case '0o':
                base = 8
                value = value[2:]
            case '0x':
                base = 16
                value = value[2:]
            case _:
                base = 10
    else:
        base = 10

    if 'u' in value:
        is_unsigned = True
        value = value.replace('u', '')
    else:
        is_unsigned = False
    
    if value.endswith('l'):
        is_long = True
        value = value[:-1]
    else:
        is_long = False

    return value, base, is_unsigned, is_long