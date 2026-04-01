package knowledge

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewVectorClient 测试创建向量客户端
func TestNewVectorClient(t *testing.T) {
	client := NewVectorClient()
	if client == nil {
		t.Fatal("NewVectorClient 返回 nil")
	}
	if client.httpClient == nil {
		t.Fatal("httpClient 不应为 nil")
	}
}

// TestVectorClient_AddDocuments_Success 测试成功存储文档向量
func TestVectorClient_AddDocuments_Success(t *testing.T) {
	// 模拟 Python 服务
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/vectors/documents" {
			t.Errorf("期望路径 /api/v1/vectors/documents，实际 %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("期望 POST 方法，实际 %s", r.Method)
		}

		// 验证请求体
		var req vectorAddRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("解析请求体失败: %v", err)
		}
		if req.Collection != "teacher_1" {
			t.Errorf("期望 collection=teacher_1，实际 %s", req.Collection)
		}
		if req.DocID != 42 {
			t.Errorf("期望 doc_id=42，实际 %d", req.DocID)
		}
		if len(req.Chunks) != 2 {
			t.Errorf("期望 2 个 chunks，实际 %d", len(req.Chunks))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(vectorAddResponse{Success: true, ChunksCount: 2})
	}))
	defer server.Close()

	client := &VectorClient{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	chunks := []VectorChunk{
		{ID: "doc_42_chunk_0", Content: "内容1", ChunkIndex: 0},
		{ID: "doc_42_chunk_1", Content: "内容2", ChunkIndex: 1},
	}

	err := client.AddDocuments("teacher_1", 42, "测试文档", chunks)
	if err != nil {
		t.Fatalf("AddDocuments 失败: %v", err)
	}
}

// TestVectorClient_AddDocuments_Degradation 测试降级逻辑（服务不可用）
func TestVectorClient_AddDocuments_Degradation(t *testing.T) {
	client := &VectorClient{
		baseURL:    "http://localhost:19999", // 不存在的端口
		httpClient: http.DefaultClient,
	}

	chunks := []VectorChunk{
		{ID: "doc_1_chunk_0", Content: "内容", ChunkIndex: 0},
	}

	// 降级：应返回 nil 而不是报错
	err := client.AddDocuments("teacher_1", 1, "测试", chunks)
	if err != nil {
		t.Fatalf("降级时不应返回错误，实际: %v", err)
	}
}

// TestVectorClient_Search_Success 测试成功语义检索
func TestVectorClient_Search_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/vectors/search" {
			t.Errorf("期望路径 /api/v1/vectors/search，实际 %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(vectorSearchResponse{
			Results: []vectorSearchResultItem{
				{Content: "匹配内容", Score: 0.92, DocID: 42, Title: "文档标题", ChunkID: "doc_42_chunk_0"},
			},
		})
	}))
	defer server.Close()

	client := &VectorClient{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	results := client.Search("teacher_1", "查询内容", 5)
	if len(results) != 1 {
		t.Fatalf("期望 1 个结果，实际 %d", len(results))
	}
	if results[0].Content != "匹配内容" {
		t.Errorf("期望 Content='匹配内容'，实际 '%s'", results[0].Content)
	}
	if results[0].Score != 0.92 {
		t.Errorf("期望 Score=0.92，实际 %f", results[0].Score)
	}
	if results[0].DocumentID != 42 {
		t.Errorf("期望 DocumentID=42，实际 %d", results[0].DocumentID)
	}
}

// TestVectorClient_Search_Degradation 测试检索降级逻辑
func TestVectorClient_Search_Degradation(t *testing.T) {
	client := &VectorClient{
		baseURL:    "http://localhost:19999",
		httpClient: http.DefaultClient,
	}

	// 降级：应返回空结果
	results := client.Search("teacher_1", "查询", 5)
	if results != nil {
		t.Fatalf("降级时应返回 nil，实际: %v", results)
	}
}

// TestVectorClient_Search_EmptyResults 测试空检索结果
func TestVectorClient_Search_EmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(vectorSearchResponse{Results: []vectorSearchResultItem{}})
	}))
	defer server.Close()

	client := &VectorClient{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	results := client.Search("teacher_1", "不存在的内容", 5)
	if len(results) != 0 {
		t.Fatalf("期望 0 个结果，实际 %d", len(results))
	}
}

// TestVectorClient_DeleteByDocumentID_Success 测试成功删除文档向量
func TestVectorClient_DeleteByDocumentID_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/v1/vectors/documents/42"
		if r.URL.Path != expectedPath {
			t.Errorf("期望路径 %s，实际 %s", expectedPath, r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("期望 DELETE 方法，实际 %s", r.Method)
		}
		collection := r.URL.Query().Get("collection")
		if collection != "teacher_1" {
			t.Errorf("期望 collection=teacher_1，实际 %s", collection)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(vectorDeleteResponse{Success: true, DeletedChunks: 3})
	}))
	defer server.Close()

	client := &VectorClient{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	err := client.DeleteByDocumentID("teacher_1", 42)
	if err != nil {
		t.Fatalf("DeleteByDocumentID 失败: %v", err)
	}
}

// TestVectorClient_DeleteByDocumentID_Degradation 测试删除降级逻辑
func TestVectorClient_DeleteByDocumentID_Degradation(t *testing.T) {
	client := &VectorClient{
		baseURL:    "http://localhost:19999",
		httpClient: http.DefaultClient,
	}

	// 降级：应返回 nil 而不是报错
	err := client.DeleteByDocumentID("teacher_1", 42)
	if err != nil {
		t.Fatalf("降级时不应返回错误，实际: %v", err)
	}
}

// TestVectorClient_AddDocuments_ServerError 测试服务端返回错误
func TestVectorClient_AddDocuments_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(vectorAddResponse{Success: false, Error: "embedding API call failed"})
	}))
	defer server.Close()

	client := &VectorClient{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	chunks := []VectorChunk{
		{ID: "doc_1_chunk_0", Content: "内容", ChunkIndex: 0},
	}

	err := client.AddDocuments("teacher_1", 1, "测试", chunks)
	if err == nil {
		t.Fatal("服务端返回错误时应返回 error")
	}
	expectedMsg := "向量服务返回错误: embedding API call failed"
	if err.Error() != expectedMsg {
		t.Errorf("期望错误信息 '%s'，实际 '%s'", expectedMsg, err.Error())
	}
}

// TestVectorClient_DefaultURL 测试默认 URL
func TestVectorClient_DefaultURL(t *testing.T) {
	client := NewVectorClient()
	// 默认 URL 应该是 http://localhost:8100（或从环境变量读取）
	if client.baseURL == "" {
		t.Fatal("baseURL 不应为空")
	}
	fmt.Printf("VectorClient baseURL: %s\n", client.baseURL)
}
