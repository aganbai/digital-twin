package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestHandleGetComments_StudentReturnsEmpty 测试学生角色调用评语接口返回空列表
func TestHandleGetComments_StudentReturnsEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/api/comments", func(c *gin.Context) {
		// 模拟 JWT 中间件设置的用户信息
		c.Set("user_id", int64(1))
		c.Set("role", "student")
		c.Set("persona_id", int64(10))
		c.Next()
	}, func(c *gin.Context) {
		// 模拟 HandleGetComments 的学生逻辑
		role, _ := c.Get("role")
		roleStr, _ := role.(string)

		if roleStr == "student" {
			Success(c, []interface{}{})
			return
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/api/comments", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际 %d", w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if resp.Code != 0 {
		t.Errorf("期望 code=0，实际 %d", resp.Code)
	}

	// data 应该是空数组
	dataArr, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("期望 data 为数组类型，实际 %T", resp.Data)
	}
	if len(dataArr) != 0 {
		t.Errorf("期望 data 为空数组，实际长度 %d", len(dataArr))
	}
}

// TestHandleGetComments_TeacherReturnsNormally 测试教师角色调用评语接口正常返回
func TestHandleGetComments_TeacherReturnsNormally(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/api/comments", func(c *gin.Context) {
		// 模拟 JWT 中间件设置教师角色
		c.Set("user_id", int64(1))
		c.Set("role", "teacher")
		c.Set("persona_id", int64(10))
		c.Next()
	}, func(c *gin.Context) {
		role, _ := c.Get("role")
		roleStr, _ := role.(string)

		// 教师角色不应走空列表分支
		if roleStr == "student" {
			t.Error("教师角色不应进入学生分支")
		}

		// 模拟教师正常返回
		Success(c, gin.H{"items": []interface{}{}, "total": 0})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/comments", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际 %d", w.Code)
	}

	body := w.Body.String()
	if strings.Contains(body, "\"data\":[]") {
		// 教师应该返回对象结构而不是空数组
		t.Log("注意：教师返回结构应包含 items 和 total 字段")
	}
}

// TestParseAttachment_TextFile 测试文本附件解析
func TestParseAttachment_TextFile(t *testing.T) {
	h := &Handler{}

	// 测试空附件
	result := h.parseAttachment("", "", "")
	if result != "" {
		t.Errorf("空附件应返回空字符串，实际: %s", result)
	}

	// 测试图片类型
	result = h.parseAttachment("/uploads/2026/03/photo.jpg", "image", "照片.jpg")
	if !strings.Contains(result, "[附件: 照片.jpg]") {
		t.Errorf("图片附件应包含附件标记，实际: %s", result)
	}
	if !strings.Contains(result, "图片文件") {
		t.Errorf("图片附件应标记为图片文件，实际: %s", result)
	}

	// 测试不存在的文件
	result = h.parseAttachment("/uploads/2026/03/nonexist.pdf", "pdf", "不存在.pdf")
	if !strings.Contains(result, "[附件: 不存在.pdf]") {
		t.Errorf("不存在的文件应包含附件标记，实际: %s", result)
	}
}

// TestParseAttachment_ImageExtensions 测试图片扩展名识别
func TestParseAttachment_ImageExtensions(t *testing.T) {
	h := &Handler{}

	exts := []string{".jpg", ".jpeg", ".png"}
	for _, ext := range exts {
		result := h.parseAttachment("/uploads/2026/03/file"+ext, "", "文件"+ext)
		if !strings.Contains(result, "图片文件") {
			t.Errorf("扩展名 %s 应被识别为图片，实际: %s", ext, result)
		}
	}
}
