"""
文本分块器
按段落/固定长度分块，每块 500-1000 字符，块间 100 字符重叠
"""


def chunk_text(text, chunk_size=800, overlap=100, min_chunk_size=100):
    """
    将文本分块

    Args:
        text: 原始文本
        chunk_size: 目标块大小（字符数），默认800
        overlap: 块间重叠字符数，默认100
        min_chunk_size: 最小块大小，小于此值的块会合并到上一块

    Returns:
        list[dict]: 分块列表，每项包含 chunk_index, content, char_count
    """
    if not text or not text.strip():
        return []

    # 先尝试按段落分割
    paragraphs = [p.strip() for p in text.split('\n') if p.strip()]

    if not paragraphs:
        return []

    # 合并段落到目标大小的块
    chunks = []
    current_chunk = ""

    for para in paragraphs:
        # 如果当前段落本身就超过 chunk_size，需要按字符切分
        if len(para) > chunk_size:
            # 先把已累积的内容作为一个块
            if current_chunk:
                chunks.append(current_chunk)
                current_chunk = ""

            # 按固定长度切分大段落
            start = 0
            while start < len(para):
                end = min(start + chunk_size, len(para))
                chunks.append(para[start:end])
                start = end - overlap if end < len(para) else end
            continue

        # 如果加上当前段落会超过 chunk_size
        if len(current_chunk) + len(para) + 1 > chunk_size:
            if current_chunk:
                chunks.append(current_chunk)
            current_chunk = para
        else:
            if current_chunk:
                current_chunk += "\n" + para
            else:
                current_chunk = para

    # 处理最后一个块
    if current_chunk:
        # 如果最后一块太小，合并到前一块
        if len(current_chunk) < min_chunk_size and chunks:
            chunks[-1] += "\n" + current_chunk
        else:
            chunks.append(current_chunk)

    # 添加重叠（对非段落边界的块）
    result_chunks = []
    for i, chunk in enumerate(chunks):
        # 对于第二个及之后的块，如果上一块末尾和当前块没有自然重叠，添加重叠
        if i > 0 and overlap > 0:
            prev_chunk = chunks[i - 1]
            overlap_text = prev_chunk[-overlap:] if len(prev_chunk) > overlap else prev_chunk
            # 只在不是段落自然分割时添加重叠
            if not chunk.startswith(overlap_text[:20] if len(overlap_text) >= 20 else overlap_text):
                chunk = overlap_text + "\n" + chunk

        result_chunks.append({
            "chunk_index": i,
            "content": chunk,
            "char_count": len(chunk),
        })

    return result_chunks
