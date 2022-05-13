def merge_metaclasses(*cls_list):
    class MergedMeta(*[type(x) for x in cls_list]):
        pass

    return MergedMeta