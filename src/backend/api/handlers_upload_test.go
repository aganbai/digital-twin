package api

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupUploadTestRouter 创建测试用路由
func setupUploadTestRouter() *gin.Engine {
	r := gin.New()
	handler := &Handler{}
	r.POST("/api/upload", handler.HandleUpload)
	return r
}

// TestHandleUpload_Success 测试成功上传文件
func TestHandleUpload_Success(t *testing.T) {
	// 创建临时文件
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("测试文件内容"), 0644)

	// 构建 multipart 请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatal(err)
	}
	fileContent := []byte("测试文件内容")
	part.Write(fileContent)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router := setupUploadTestRouter()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际 %d，响应: %s", w.Code, w.Body.String())
	}

	respBody := w.Body.String()
	if !strings.Contains(respBody, "\"url\"") {
		t.Errorf("响应应包含 url 字段，实际: %s", respBody)
	}
	if !strings.Contains(respBody, "\"original_name\":\"test.txt\"") {
		t.Errorf("响应应包含 original_name=test.txt，实际: %s", respBody)
	}
	if !strings.Contains(respBody, "text/plain") {
		t.Errorf("响应应包含 mime_type=text/plain，实际: %s", respBody)
	}

	// 清理上传的文件
	os.RemoveAll("uploads")
}

// TestHandleUpload_UnsupportedType 测试不支持的文件类型
func TestHandleUpload_UnsupportedType(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.exe")
	part.Write([]byte("binary content"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router := setupUploadTestRouter()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望状态码 400，实际 %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "40035") {
		t.Errorf("期望错误码 40035，实际: %s", w.Body.String())
	}
}

// TestHandleUpload_FileTooLarge 测试文件过大
func TestHandleUpload_FileTooLarge(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "large.txt")
	// 写入 11MB 数据
	largeData := make([]byte, 11<<20)
	part.Write(largeData)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router := setupUploadTestRouter()
	router.ServeHTTP(w, req)

	// 应该返回错误（400 或 413）
	if w.Code == http.StatusOK {
		t.Fatal("上传过大文件应失败")
	}
}

// TestHandleUpload_NoFile 测试缺少文件
func TestHandleUpload_NoFile(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router := setupUploadTestRouter()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望状态码 400，实际 %d", w.Code)
	}
}

// TestHandleUpload_PDFFile 测试上传 PDF 文件
func TestHandleUpload_PDFFile(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "document.pdf")
	part.Write([]byte("%PDF-1.4 test content"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router := setupUploadTestRouter()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际 %d，响应: %s", w.Code, w.Body.String())
	}

	if !strings.Contains(w.Body.String(), "application/pdf") {
		t.Errorf("响应应包含 application/pdf，实际: %s", w.Body.String())
	}

	// 清理
	os.RemoveAll("uploads")
}

// TestHandleUpload_ImageFile 测试上传图片文件
func TestHandleUpload_ImageFile(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "photo.jpg")
	part.Write([]byte("\xff\xd8\xff\xe0 fake jpg"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router := setupUploadTestRouter()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际 %d，响应: %s", w.Code, w.Body.String())
	}

	if !strings.Contains(w.Body.String(), "image/jpeg") {
		t.Errorf("响应应包含 image/jpeg，实际: %s", w.Body.String())
	}

	// 清理
	os.RemoveAll("uploads")
}

// TestSanitizeFilename 测试文件名清理
func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "helloworld"},
		{"文件名", "文件名"},
		{"test_file-name", "test_file-name"},
		{"bad<>file", "badfile"},
		{"", ""},
		{"abc123", "abc123"},
	}

	for _, tt := range tests {
		result := sanitizeFilename(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeFilename(%q) = %q，期望 %q", tt.input, result, tt.expected)
		}
	}
}

// 防止 unused import
var _ = io.Discard
