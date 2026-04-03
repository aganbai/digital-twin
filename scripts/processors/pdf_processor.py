"""PDF文件处理器"""

try:
    from PyPDF2 import PdfReader
    HAS_PYPDF2 = True
except ImportError:
    HAS_PYPDF2 = False


def process_pdf(filepath):
    """解析PDF文件，提取文本内容"""
    if not HAS_PYPDF2:
        raise ImportError("PyPDF2 未安装，请运行: pip install PyPDF2")

    reader = PdfReader(filepath)
    text_parts = []
    for page in reader.pages:
        page_text = page.extract_text()
        if page_text:
            text_parts.append(page_text)

    content = "\n".join(text_parts)
    if not content.strip():
        raise ValueError("PDF文件中未能提取到文本内容")
    return content
