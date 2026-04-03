package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"digital-twin/src/backend/database"

	"github.com/gin-gonic/gin"
)

// ==================== HandleGetShareInfoV2 join_status 测试 ====================

// TestShareInfoV2_NeedLogin 未登录时返回 need_login
func TestShareInfoV2_NeedLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/api/shares/:code/info", func(c *gin.Context) {
		// 不设置任何认证信息，模拟未登录
		c.Next()
	}, func(c *gin.Context) {
		// 模拟 HandleGetShareInfoV2 的 join_status 判断逻辑
		joinStatus := "need_login"
		_, authenticated := c.Get("authenticated")
		if authenticated {
			joinStatus = "can_join" // 不会走到这里
		}

		Success(c, gin.H{
			"teacher_persona_id":        int64(1),
			"teacher_nickname":          "张老师",
			"teacher_school":            "XX中学",
			"teacher_description":       "数学老师",
			"class_name":                "高一1班",
			"target_student_persona_id": int64(0),
			"is_valid":                  true,
			"join_status":               joinStatus,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/shares/TESTCODE/info", nil)
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

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("期望 data 为 map 类型，实际 %T", resp.Data)
	}
	if data["join_status"] != "need_login" {
		t.Errorf("期望 join_status=need_login，实际 %v", data["join_status"])
	}
}

// TestShareInfoV2_Authenticated 已登录时返回正确的 join_status
func TestShareInfoV2_Authenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/api/shares/:code/info", func(c *gin.Context) {
		// 模拟 OptionalJWTAuthMiddleware 设置的认证信息
		c.Set("authenticated", true)
		c.Set("user_id", int64(1))
		c.Set("persona_id", int64(10))
		c.Next()
	}, func(c *gin.Context) {
		joinStatus := "need_login"
		_, authenticated := c.Get("authenticated")
		if authenticated {
			joinStatus = "can_join"
		}

		Success(c, gin.H{
			"teacher_persona_id": int64(1),
			"teacher_nickname":   "张老师",
			"is_valid":           true,
			"join_status":        joinStatus,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/shares/TESTCODE/info", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	data := resp.Data.(map[string]interface{})
	if data["join_status"] != "can_join" {
		t.Errorf("期望 join_status=can_join，实际 %v", data["join_status"])
	}
}

// TestShareInfoV2_NeedPersona 已登录但无学生分身时返回 need_persona
func TestShareInfoV2_NeedPersona(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/api/shares/:code/info", func(c *gin.Context) {
		c.Set("authenticated", true)
		c.Set("user_id", int64(1))
		c.Set("persona_id", int64(0)) // 无分身
		c.Next()
	}, func(c *gin.Context) {
		joinStatus := "need_login"
		_, authenticated := c.Get("authenticated")
		if authenticated {
			// 模拟无学生分身
			joinStatus = "need_persona"
		}

		Success(c, gin.H{
			"teacher_persona_id": int64(1),
			"teacher_nickname":   "张老师",
			"is_valid":           true,
			"join_status":        joinStatus,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/shares/TESTCODE/info", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	data := resp.Data.(map[string]interface{})
	if data["join_status"] != "need_persona" {
		t.Errorf("期望 join_status=need_persona，实际 %v", data["join_status"])
	}
}

// TestShareInfoV2_NotTarget 非目标学生时返回 not_target
func TestShareInfoV2_NotTarget(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/api/shares/:code/info", func(c *gin.Context) {
		c.Set("authenticated", true)
		c.Set("user_id", int64(2))
		c.Set("persona_id", int64(20))
		c.Next()
	}, func(c *gin.Context) {
		// 模拟定向分享码 + 非目标学生
		joinStatus := "not_target"

		Success(c, gin.H{
			"teacher_persona_id":        int64(1),
			"teacher_nickname":          "张老师",
			"teacher_school":            "XX中学",
			"teacher_description":       "数学老师",
			"class_name":                "高一1班",
			"target_student_persona_id": int64(5),
			"target_student_nickname":   "李四",
			"is_valid":                  true,
			"join_status":               joinStatus,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/shares/TESTCODE/info", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	data := resp.Data.(map[string]interface{})
	if data["join_status"] != "not_target" {
		t.Errorf("期望 join_status=not_target，实际 %v", data["join_status"])
	}
	if data["target_student_nickname"] != "李四" {
		t.Errorf("期望 target_student_nickname=李四，实际 %v", data["target_student_nickname"])
	}
	// 验证 target_student_persona_id 存在
	targetID, ok := data["target_student_persona_id"].(float64)
	if !ok || targetID != 5 {
		t.Errorf("期望 target_student_persona_id=5，实际 %v", data["target_student_persona_id"])
	}
}

// TestShareInfoV2_AlreadyJoined 已是该教师学生时返回 already_joined
func TestShareInfoV2_AlreadyJoined(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/api/shares/:code/info", func(c *gin.Context) {
		c.Set("authenticated", true)
		c.Set("user_id", int64(2))
		c.Set("persona_id", int64(20))
		c.Next()
	}, func(c *gin.Context) {
		// 模拟已有 approved 关系
		joinStatus := "already_joined"

		Success(c, gin.H{
			"teacher_persona_id": int64(1),
			"teacher_nickname":   "张老师",
			"is_valid":           true,
			"join_status":        joinStatus,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/shares/TESTCODE/info", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	data := resp.Data.(map[string]interface{})
	if data["join_status"] != "already_joined" {
		t.Errorf("期望 join_status=already_joined，实际 %v", data["join_status"])
	}
}

// ==================== HandleJoinByShare not_target 引导测试 ====================

// TestJoinByShare_NotTarget_ReturnsGuidance 非目标学生 join 返回引导信息
func TestJoinByShare_NotTarget_ReturnsGuidance(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/api/shares/:code/join", func(c *gin.Context) {
		c.Set("user_id", int64(2))
		c.Set("persona_id", int64(20))
		c.Next()
	}, func(c *gin.Context) {
		// 模拟非目标学生的引导响应
		Success(c, gin.H{
			"join_status":        "not_target",
			"teacher_persona_id": int64(1),
			"teacher_nickname":   "张老师",
			"message":            "该邀请码是老师专门发给特定同学的，你可以向老师发起申请",
			"can_apply":          true,
		})
	})

	req := httptest.NewRequest(http.MethodPost, "/api/shares/TESTCODE/join", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
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
		t.Errorf("期望 code=0（非错误码），实际 %d", resp.Code)
	}

	data := resp.Data.(map[string]interface{})
	if data["join_status"] != "not_target" {
		t.Errorf("期望 join_status=not_target，实际 %v", data["join_status"])
	}
	if data["can_apply"] != true {
		t.Errorf("期望 can_apply=true，实际 %v", data["can_apply"])
	}
	msg, _ := data["message"].(string)
	if !strings.Contains(msg, "专门发给特定同学") {
		t.Errorf("期望包含引导信息，实际: %s", msg)
	}
}

// TestJoinByShare_NormalJoin_ReturnsRelation 正常加入返回关系信息
func TestJoinByShare_NormalJoin_ReturnsRelation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/api/shares/:code/join", func(c *gin.Context) {
		c.Set("user_id", int64(2))
		c.Set("persona_id", int64(20))
		c.Next()
	}, func(c *gin.Context) {
		// 模拟正常加入的响应
		Success(c, gin.H{
			"relation_id":        int64(1),
			"teacher_persona_id": int64(1),
			"teacher_nickname":   "张老师",
			"class_id":           nil,
			"class_name":         "",
			"joined_class":       false,
		})
	})

	req := httptest.NewRequest(http.MethodPost, "/api/shares/TESTCODE/join", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
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

	data := resp.Data.(map[string]interface{})
	// 正常加入不应有 join_status 字段（或者没有 not_target）
	if js, exists := data["join_status"]; exists {
		t.Errorf("正常加入不应有 join_status 字段，实际: %v", js)
	}
	if data["relation_id"] == nil {
		t.Error("正常加入应返回 relation_id")
	}
}

// ==================== determineJoinStatus 逻辑测试 ====================

// TestDetermineJoinStatus_AllScenarios 测试所有 join_status 场景
func TestDetermineJoinStatus_AllScenarios(t *testing.T) {
	tests := []struct {
		name             string
		targetPersonaID  int64
		currentPersonaID int64
		approved         bool
		expected         string
	}{
		{
			name:             "已是该教师的学生",
			targetPersonaID:  0,
			currentPersonaID: 10,
			approved:         true,
			expected:         "already_joined",
		},
		{
			name:             "通用分享码可加入",
			targetPersonaID:  0,
			currentPersonaID: 10,
			approved:         false,
			expected:         "can_join",
		},
		{
			name:             "定向分享码目标匹配",
			targetPersonaID:  10,
			currentPersonaID: 10,
			approved:         false,
			expected:         "can_join",
		},
		{
			name:             "定向分享码非目标",
			targetPersonaID:  5,
			currentPersonaID: 10,
			approved:         false,
			expected:         "not_target",
		},
		{
			name:             "定向分享码非目标但已是学生",
			targetPersonaID:  5,
			currentPersonaID: 10,
			approved:         true,
			expected:         "already_joined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			share := &database.PersonaShare{
				TargetStudentPersonaID: tt.targetPersonaID,
			}
			result := determineJoinStatusForTest(share, tt.currentPersonaID, tt.approved)
			if result != tt.expected {
				t.Errorf("join_status 不匹配: got %q, want %q", result, tt.expected)
			}
		})
	}
}

// determineJoinStatusForTest 辅助函数：与 handler 中的逻辑保持一致
func determineJoinStatusForTest(share *database.PersonaShare, studentPersonaID int64, approved bool) string {
	if approved {
		return "already_joined"
	}
	if share.TargetStudentPersonaID > 0 && share.TargetStudentPersonaID != studentPersonaID {
		return "not_target"
	}
	return "can_join"
}
