package knowledge

// TextChunker 文本分块器
type TextChunker struct {
	chunkSize    int
	chunkOverlap int
}

// NewTextChunker 创建文本分块器
func NewTextChunker(chunkSize, chunkOverlap int) *TextChunker {
	if chunkSize <= 0 {
		chunkSize = 500
	}
	if chunkOverlap < 0 {
		chunkOverlap = 0
	}
	if chunkOverlap >= chunkSize {
		chunkOverlap = chunkSize / 4
	}
	return &TextChunker{
		chunkSize:    chunkSize,
		chunkOverlap: chunkOverlap,
	}
}

// Chunk 按字符数分块，支持 overlap
func (c *TextChunker) Chunk(text string) []string {
	runes := []rune(text)
	totalLen := len(runes)

	if totalLen == 0 {
		return nil
	}

	if totalLen <= c.chunkSize {
		return []string{string(runes)}
	}

	var chunks []string
	step := c.chunkSize - c.chunkOverlap
	if step <= 0 {
		step = 1
	}

	for i := 0; i < totalLen; i += step {
		end := i + c.chunkSize
		if end > totalLen {
			end = totalLen
		}
		chunk := string(runes[i:end])
		chunks = append(chunks, chunk)

		// 如果已经到达末尾，停止
		if end >= totalLen {
			break
		}
	}

	return chunks
}
