package knowledge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// VectorClient 向量服务客户端，封装 HTTP 调用 Python LlamaIndex 服务
type VectorClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewVectorClient 创建向量服务客户端
func NewVectorClient() *VectorClient {
	baseURL := os.Getenv("KNOWLEDGE_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8100"
	}
	return &VectorClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// vectorAddRequest 存储文档向量请求
type vectorAddRequest struct {
	Collection string             `json:"collection"`
	DocID      int64              `json:"doc_id"`
	Title      string             `json:"title"`
	Chunks     []vectorChunkInput `json:"chunks"`
}

// vectorChunkInput 文本块输入
type vectorChunkInput struct {
	ID         string `json:"id"`
	Content    string `json:"content"`
	ChunkIndex int    `json:"chunk_index"`
}

// vectorAddResponse 存储文档向量响应
type vectorAddResponse struct {
	Success     bool   `json:"success"`
	ChunksCount int    `json:"chunks_count"`
	Error       string `json:"error,omitempty"`
}

// vectorSearchRequest 语义检索请求
type vectorSearchRequest struct {
	Collection string `json:"collection"`
	Query      string `json:"query"`
	TopK       int    `json:"top_k"`
}

// vectorSearchResultItem 检索结果项
type vectorSearchResultItem struct {
	Content string  `json:"content"`
	Score   float64 `json:"score"`
	DocID   int64   `json:"doc_id"`
	Title   string  `json:"title"`
	ChunkID string  `json:"chunk_id"`
}

// vectorSearchResponse 语义检索响应
type vectorSearchResponse struct {
	Results []vectorSearchResultItem `json:"results"`
}

// vectorDeleteResponse 删除文档向量响应
type vectorDeleteResponse struct {
	Success       bool   `json:"success"`
	DeletedChunks int    `json:"deleted_chunks"`
	Error         string `json:"error,omitempty"`
}

// AddDocuments 存储文档向量到 Python 服务
// 降级逻辑：Python 服务不可用时返回 nil（不报错）
func (vc *VectorClient) AddDocuments(collection string, docID int64, title string, chunks []VectorChunk) error {
	// 构建请求
	reqChunks := make([]vectorChunkInput, 0, len(chunks))
	for _, c := range chunks {
		reqChunks = append(reqChunks, vectorChunkInput{
			ID:         c.ID,
			Content:    c.Content,
			ChunkIndex: c.ChunkIndex,
		})
	}

	reqBody := vectorAddRequest{
		Collection: collection,
		DocID:      docID,
		Title:      title,
		Chunks:     reqChunks,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	resp, err := vc.httpClient.Post(
		vc.baseURL+"/api/v1/vectors/documents",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		// 降级：Python 服务不可用，记录日志但不报错
		log.Printf("[VectorClient] 向量服务不可用，降级处理（AddDocuments）: %v", err)
		return nil
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[VectorClient] 读取响应失败: %v", err)
		return nil
	}

	var result vectorAddResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Printf("[VectorClient] 解析响应失败: %v", err)
		return nil
	}

	if !result.Success {
		return fmt.Errorf("向量服务返回错误: %s", result.Error)
	}

	return nil
}

// Search 语义检索
// 降级逻辑：Python 服务不可用时返回空结果
func (vc *VectorClient) Search(collection string, query string, topK int) []SearchResult {
	reqBody := vectorSearchRequest{
		Collection: collection,
		Query:      query,
		TopK:       topK,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("[VectorClient] 序列化请求失败: %v", err)
		return nil
	}

	resp, err := vc.httpClient.Post(
		vc.baseURL+"/api/v1/vectors/search",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		// 降级：Python 服务不可用，返回空结果
		log.Printf("[VectorClient] 向量服务不可用，降级处理（Search）: %v", err)
		return nil
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[VectorClient] 读取响应失败: %v", err)
		return nil
	}

	var result vectorSearchResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Printf("[VectorClient] 解析响应失败: %v", err)
		return nil
	}

	// 转换为 SearchResult
	var searchResults []SearchResult
	for _, r := range result.Results {
		searchResults = append(searchResults, SearchResult{
			Content:    r.Content,
			Score:      r.Score,
			DocumentID: r.DocID,
			Title:      r.Title,
		})
	}

	return searchResults
}

// DeleteByDocumentID 删除指定文档的所有向量
// 降级逻辑：Python 服务不可用时返回 nil（不报错）
func (vc *VectorClient) DeleteByDocumentID(collection string, docID int64) error {
	url := fmt.Sprintf("%s/api/v1/vectors/documents/%d?collection=%s", vc.baseURL, docID, collection)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := vc.httpClient.Do(req)
	if err != nil {
		// 降级：Python 服务不可用，记录日志但不报错
		log.Printf("[VectorClient] 向量服务不可用，降级处理（DeleteByDocumentID）: %v", err)
		return nil
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[VectorClient] 读取响应失败: %v", err)
		return nil
	}

	var result vectorDeleteResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Printf("[VectorClient] 解析响应失败: %v", err)
		return nil
	}

	if !result.Success {
		return fmt.Errorf("向量服务返回错误: %s", result.Error)
	}

	return nil
}
