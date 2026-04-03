"""Markdown文件处理器"""


def process_md(filepath):
    """解析Markdown文件，返回文本内容（保留Markdown格式）"""
    with open(filepath, 'r', encoding='utf-8', errors='replace') as f:
        return f.read()
