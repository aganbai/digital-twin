package knowledge

import (
	"math"
	"strings"
	"sync"
)

// VectorChunk 文档块
type VectorChunk struct {
	ID         string `json:"id"`
	Content    string `json:"content"`
	DocumentID int64  `json:"document_id"`
	TeacherID  int64  `json:"teacher_id"`
	Title      string `json:"title"`
	ChunkIndex int    `json:"chunk_index"`
	Scope      string `json:"scope"`    // global / class / student
	ScopeID    int64  `json:"scope_id"` // scope=class 时为班级ID，scope=student 时为学生分身ID
}

// SearchResult 检索结果
type SearchResult struct {
	Content    string  `json:"content"`
	Score      float64 `json:"score"`
	DocumentID int64   `json:"document_id"`
	Title      string  `json:"title"`
}

// InMemoryVectorStore 内存向量存储
type InMemoryVectorStore struct {
	mu          sync.RWMutex
	collections map[string][]VectorChunk
}

// NewInMemoryVectorStore 创建内存向量存储
func NewInMemoryVectorStore() *InMemoryVectorStore {
	return &InMemoryVectorStore{
		collections: make(map[string][]VectorChunk),
	}
}

// AddDocuments 存储文档块
func (s *InMemoryVectorStore) AddDocuments(collectionName string, chunks []VectorChunk) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.collections[collectionName]; !exists {
		s.collections[collectionName] = make([]VectorChunk, 0)
	}

	s.collections[collectionName] = append(s.collections[collectionName], chunks...)
	return nil
}

// Search 简单关键词匹配检索（TF-IDF 简化版）
func (s *InMemoryVectorStore) Search(collectionName string, query string, limit int) []SearchResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	chunks, exists := s.collections[collectionName]
	if !exists || len(chunks) == 0 {
		return nil
	}

	// 分词（简单按空格和标点分割）
	queryTerms := tokenize(query)
	if len(queryTerms) == 0 {
		return nil
	}

	// 计算每个文档块的匹配分数
	type scoredResult struct {
		chunk VectorChunk
		score float64
	}

	var results []scoredResult
	for _, chunk := range chunks {
		score := calculateScore(chunk.Content, queryTerms)
		if score > 0 {
			results = append(results, scoredResult{chunk: chunk, score: score})
		}
	}

	// 按分数降序排序
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].score > results[i].score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// 限制返回数量
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	// 转换为 SearchResult
	var searchResults []SearchResult
	for _, r := range results {
		searchResults = append(searchResults, SearchResult{
			Content:    r.chunk.Content,
			Score:      r.score,
			DocumentID: r.chunk.DocumentID,
			Title:      r.chunk.Title,
		})
	}

	return searchResults
}

// DeleteByDocumentID 删除指定文档的所有块
func (s *InMemoryVectorStore) DeleteByDocumentID(collectionName string, documentID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	chunks, exists := s.collections[collectionName]
	if !exists {
		return nil
	}

	var remaining []VectorChunk
	for _, chunk := range chunks {
		if chunk.DocumentID != documentID {
			remaining = append(remaining, chunk)
		}
	}

	s.collections[collectionName] = remaining
	return nil
}

// tokenize 简单分词：按空格和常见标点分割，转为小写
func tokenize(text string) []string {
	// 将标点和空白字符替换为空格
	var builder strings.Builder
	for _, r := range text {
		if isPunctuation(r) || r == '\n' || r == '\r' || r == '\t' {
			builder.WriteRune(' ')
		} else {
			builder.WriteRune(r)
		}
	}
	text = strings.ToLower(builder.String())

	parts := strings.Fields(text)
	var tokens []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if len(p) > 0 {
			tokens = append(tokens, p)
		}
	}
	return tokens
}

// isPunctuation 判断是否为标点符号（包括中英文标点）
func isPunctuation(r rune) bool {
	// 英文标点
	if strings.ContainsRune(",.!?;:()[]{}\"'`~@#$%^&*-_=+<>/\\|", r) {
		return true
	}
	// 中文标点 Unicode 范围
	if r >= 0x3000 && r <= 0x303F {
		return true
	}
	// 全角标点
	if r >= 0xFF00 && r <= 0xFFEF {
		return true
	}
	return false
}

// calculateScore 计算匹配分数（TF-IDF 简化版）
func calculateScore(content string, queryTerms []string) float64 {
	contentLower := strings.ToLower(content)
	contentTokens := tokenize(content)

	if len(contentTokens) == 0 {
		return 0
	}

	var totalScore float64
	matchedTerms := 0

	for _, term := range queryTerms {
		termLower := strings.ToLower(term)

		// 方式1：直接字符串包含匹配
		if strings.Contains(contentLower, termLower) {
			// 计算词频 TF
			count := strings.Count(contentLower, termLower)
			tf := float64(count) / float64(len(contentTokens))
			totalScore += tf
			matchedTerms++
			continue
		}

		// 方式2：token 精确匹配
		for _, ct := range contentTokens {
			if ct == termLower {
				tf := 1.0 / float64(len(contentTokens))
				totalScore += tf
				matchedTerms++
				break
			}
		}
	}

	// 匹配覆盖率加权
	if matchedTerms > 0 {
		coverage := float64(matchedTerms) / float64(len(queryTerms))
		totalScore *= (1 + coverage)
	}

	// 归一化到 0-1 之间
	maxScore := float64(len(queryTerms)) * 2
	if maxScore > 0 {
		totalScore = math.Min(totalScore/maxScore, 1.0)
	}

	return totalScore
}
