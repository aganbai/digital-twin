package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandleGetCurriculumVersions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &Handler{}

	// 测试不带 region 参数
	t.Run("无region参数", func(t *testing.T) {
		r := gin.New()
		r.GET("/api/curriculum-versions", handler.HandleGetCurriculumVersions)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/curriculum-versions", nil)
		r.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("期望200，实际 %d", w.Code)
		}

		var resp struct {
			Code int `json:"code"`
			Data struct {
				Versions    []string `json:"versions"`
				Recommended string   `json:"recommended"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("解析响应失败: %v", err)
		}

		if len(resp.Data.Versions) == 0 {
			t.Fatal("版本列表不应为空")
		}
		if resp.Data.Recommended != "人教版" {
			t.Errorf("默认推荐应为人教版，实际: %s", resp.Data.Recommended)
		}
	})

	// 测试带 region 参数
	t.Run("上海region", func(t *testing.T) {
		r := gin.New()
		r.GET("/api/curriculum-versions", handler.HandleGetCurriculumVersions)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/curriculum-versions?region=上海", nil)
		r.ServeHTTP(w, req)

		var resp struct {
			Code int `json:"code"`
			Data struct {
				Recommended string `json:"recommended"`
			} `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)

		if resp.Data.Recommended != "沪教版" {
			t.Errorf("上海应推荐沪教版，实际: %s", resp.Data.Recommended)
		}
	})

	// 测试成人学段返回空列表
	t.Run("成人学段_adult_life", func(t *testing.T) {
		r := gin.New()
		r.GET("/api/curriculum-versions", handler.HandleGetCurriculumVersions)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/curriculum-versions?grade_level=adult_life", nil)
		r.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("期望200，实际 %d", w.Code)
		}

		var resp struct {
			Code int `json:"code"`
			Data struct {
				Versions    []string `json:"versions"`
				Recommended string   `json:"recommended"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("解析响应失败: %v", err)
		}

		if len(resp.Data.Versions) != 0 {
			t.Errorf("成人学段版本列表应为空，实际: %v", resp.Data.Versions)
		}
		if resp.Data.Recommended != "" {
			t.Errorf("成人学段推荐应为空，实际: %s", resp.Data.Recommended)
		}
	})

	// 测试成人学段 adult_professional
	t.Run("成人学段_adult_professional", func(t *testing.T) {
		r := gin.New()
		r.GET("/api/curriculum-versions", handler.HandleGetCurriculumVersions)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/curriculum-versions?grade_level=adult_professional", nil)
		r.ServeHTTP(w, req)

		var resp struct {
			Code int `json:"code"`
			Data struct {
				Versions []string `json:"versions"`
			} `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)

		if len(resp.Data.Versions) != 0 {
			t.Errorf("成人学段版本列表应为空，实际: %v", resp.Data.Versions)
		}
	})

	// 测试非成人学段带grade_level参数
	t.Run("junior学段", func(t *testing.T) {
		r := gin.New()
		r.GET("/api/curriculum-versions", handler.HandleGetCurriculumVersions)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/curriculum-versions?grade_level=junior", nil)
		r.ServeHTTP(w, req)

		var resp struct {
			Code int `json:"code"`
			Data struct {
				Versions    []string `json:"versions"`
				Recommended string   `json:"recommended"`
			} `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)

		if len(resp.Data.Versions) != 6 {
			t.Errorf("非成人学段应返回6个版本，实际: %d", len(resp.Data.Versions))
		}
		if resp.Data.Recommended != "人教版" {
			t.Errorf("默认推荐应为人教版，实际: %s", resp.Data.Recommended)
		}
	})
}

func TestValidGradeLevels(t *testing.T) {
	expected := []string{
		"preschool", "primary_lower", "primary_upper", "junior",
		"senior", "university", "adult_life", "adult_professional",
	}

	for _, level := range expected {
		if !validGradeLevels[level] {
			t.Errorf("学段 %s 应该在有效列表中", level)
		}
	}

	// 无效学段
	if validGradeLevels["invalid_level"] {
		t.Error("无效学段不应在列表中")
	}
}

func TestInferGradeLevel(t *testing.T) {
	tests := []struct {
		grade    string
		expected string
	}{
		{"初一", "junior"},
		{"高一", "senior"},
		{"一年级", "primary_lower"},
		{"五年级", "primary_upper"},
		{"大一", "university"},
		{"学前班", "preschool"},
		{"未知", ""},
	}

	for _, tt := range tests {
		result := inferGradeLevel(tt.grade)
		if result != tt.expected {
			t.Errorf("inferGradeLevel(%s) = %s, 期望 %s", tt.grade, result, tt.expected)
		}
	}
}

// TestHandleUpdateCurriculumConfig_InvalidID 测试无效ID
func TestHandleUpdateCurriculumConfig_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &Handler{}

	r := gin.New()
	r.PUT("/api/curriculum-configs/:id", handler.HandleUpdateCurriculumConfig)

	w := httptest.NewRecorder()
	body := `{"grade_level":"junior"}`
	req, _ := http.NewRequest("PUT", "/api/curriculum-configs/abc", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatalf("无效ID应返回400，实际 %d", w.Code)
	}
}
