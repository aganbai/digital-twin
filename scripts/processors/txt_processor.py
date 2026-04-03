"""TXT文件处理器"""


def process_txt(filepath):
    """解析TXT文件，返回文本内容"""
    with open(filepath, 'r', encoding='utf-8', errors='replace') as f:
        return f.read()
