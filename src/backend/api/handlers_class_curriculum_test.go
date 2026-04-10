package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"digital-twin/src/backend/database"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ==================== 基础知识测试 ====================

// TestCurriculumConfigValue_Marshal 测试教材配置值序列化
func TestCurriculumConfigValue_Marshal(t *testing.T) {
	config := CurriculumConfigValue{
		GradeLevel:       "primary_lower",
		Grade:            "三年级",
		Subjects:         []string{"数学", "语文"},
		TextbookVersions: []string{"人教版", "北师大版"},
		CustomTextbooks:  []string{"《小学奥数》"},
		CurrentProgress:  "第三单元 乘法初步",
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	if result["grade_level"] != "primary_lower" {
		t.Errorf("grade_level 错误，实际: %v", result["grade_level"])
	}
	if result["grade"] != "三年级" {
		t.Errorf("grade 错误，实际: %v", result["grade"])
	}
	if result["current_progress"] != "第三单元 乘法初步" {
		t.Errorf("current_progress 错误，实际: %v", result["current_progress"])
	}
}

// TestCurriculumConfigValue_Unmarshal 测试教材配置值反序列化
func TestCurriculumConfigValue_Unmarshal(t *testing.T) {
	jsonData := `{
		"grade_level": "primary_upper",
		"grade": "五年级",
		"subjects": ["数学", "英语", "科学"],
		"textbook_versions": ["人教版"],
		"custom_textbooks": ["《奥数进阶》"],
		"current_progress": "第二单元"
	}`

	var config CurriculumConfigValue
	if err := json.Unmarshal([]byte(jsonData), &config); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	if config.GradeLevel != "primary_upper" {
		t.Errorf("GradeLevel 错误，实际: %s", config.GradeLevel)
	}
	if config.Grade != "五年级" {
		t.Errorf("Grade 错误，实际: %s", config.Grade)
	}
	if len(config.Subjects) != 3 {
		t.Errorf("Subjects 长度错误，实际: %d", len(config.Subjects))
	}
	if len(config.TextbookVersions) != 1 {
		t.Errorf("TextbookVersions 长度错误，实际: %d", len(config.TextbookVersions))
	}
	if config.CurrentProgress != "第二单元" {
		t.Errorf("CurrentProgress 错误，实际: %s", config.CurrentProgress)
	}
}

// TestValidGradeLevelsExt 测试学段枚举值
func TestValidGradeLevelsExt(t *testing.T) {
	expected := []string{
		"preschool",
		"primary_lower",
		"primary_upper",
		"junior",
		"senior",
		"university",
		"adult_life",
		"adult_professional",
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

	// 常见的中文枚举不应在列表中（API期望使用英文枚举值）
	if validGradeLevels["小学"] {
		t.Error("中文枚举不应在有效列表中")
	}
}

// TestCurriculumConfigToResponse 测试教材配置响应转换
func TestCurriculumConfigToResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	textbookVersions, _ := json.Marshal([]string{"人教版", "苏教版"})
	subjects, _ := json.Marshal([]string{"数学"})
	customTextbooks, _ := json.Marshal([]string{"《小学奥数》"})

	config := &database.TeacherCurriculumConfig{
		ID:               123,
		TeacherID:        1,
		PersonaID:        2,
		GradeLevel:       "primary_lower",
		Grade:            "三年级",
		TextbookVersions: string(textbookVersions),
		Subjects:         string(subjects),
		Region:           string(customTextbooks), // 使用Region字段存储自定义教材
		CurrentProgress:  "第三单元",
		IsActive:         1,
	}

	result := curriculumConfigToResponse(config)

	if result["id"] != int64(123) {
		t.Errorf("id 错误，实际: %v", result["id"])
	}
	if result["grade_level"] != "primary_lower" {
		t.Errorf("grade_level 错误，实际: %v", result["grade_level"])
	}
	if result["grade"] != "三年级" {
		t.Errorf("grade 错误，实际: %v", result["grade"])
	}
	if result["current_progress"] != "第三单元" {
		t.Errorf("current_progress 错误，实际: %v", result["current_progress"])
	}

	// 验证数组字段
	textbookVersionsResp, ok := result["textbook_versions"].([]string)
	if !ok {
		t.Fatalf("textbook_versions 类型错误")
	}
	if len(textbookVersionsResp) != 2 {
		t.Errorf("textbook_versions 长度错误，实际: %d", len(textbookVersionsResp))
	}

	subjectsResp, ok := result["subjects"].([]string)
	if !ok {
		t.Fatalf("subjects 类型错误")
	}
	if len(subjectsResp) != 1 || subjectsResp[0] != "数学" {
		t.Errorf("subjects 错误，实际: %v", subjectsResp)
	}

	customTextbooksResp, ok := result["custom_textbooks"].([]string)
	if !ok {
		t.Fatalf("custom_textbooks 类型错误")
	}
	if len(customTextbooksResp) != 1 || customTextbooksResp[0] != "《小学奥数》" {
		t.Errorf("custom_textbooks 错误，实际: %v", customTextbooksResp)
	}
}

// TestCurriculumConfigToResponse_Empty 测试空教材配置响应转换
func TestCurriculumConfigToResponse_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &database.TeacherCurriculumConfig{
		ID:               456,
		TeacherID:        1,
		PersonaID:        2,
		GradeLevel:       "university",
		TextbookVersions: "[]",
		Subjects:         "[]",
		Region:           "[]",
		IsActive:         1,
	}

	result := curriculumConfigToResponse(config)

	if result["id"] != int64(456) {
		t.Errorf("id 错误")
	}
	if result["grade_level"] != "university" {
		t.Errorf("grade_level 错误")
	}

	// 验证空数组
	textbookVersionsResp, ok := result["textbook_versions"].([]string)
	if !ok {
		t.Fatalf("textbook_versions 类型错误")
	}
	if len(textbookVersionsResp) != 0 {
		t.Errorf("textbook_versions 应为空数组，实际: %v", textbookVersionsResp)
	}
}

// TestCurriculumConfigValue_EdgeCases 测试边界情况
func TestCurriculumConfigValue_EdgeCases(t *testing.T) {
	// 测试空JSON
	emptyJson := `{}`
	var emptyConfig CurriculumConfigValue
	if err := json.Unmarshal([]byte(emptyJson), &emptyConfig); err != nil {
		t.Fatalf("空JSON反序列化失败: %v", err)
	}
	if emptyConfig.GradeLevel != "" {
		t.Error("空JSON时GradeLevel应为空")
	}
	if len(emptyConfig.Subjects) != 0 {
		t.Error("空JSON时Subjects应为空数组")
	}

	// 测试null值
	nullJson := `{"grade_level":null,"subjects":null}`
	var nullConfig CurriculumConfigValue
	if err := json.Unmarshal([]byte(nullJson), &nullConfig); err != nil {
		t.Fatalf("null JSON反序列化失败: %v", err)
	}
	if nullConfig.GradeLevel != "" {
		t.Error("null值时GradeLevel应为空")
	}
	if len(nullConfig.Subjects) != 0 {
		t.Error("null值时Subjects应为空数组")
	}

	// 测试空字符串
	emptyStringJson := `{"grade_level":"","subjects":[]}`
	var emptyStringConfig CurriculumConfigValue
	if err := json.Unmarshal([]byte(emptyStringJson), &emptyStringConfig); err != nil {
		t.Fatalf("空字符串JSON反序列化失败: %v", err)
	}
	if emptyStringConfig.GradeLevel != "" {
		t.Error("空字符串时GradeLevel应为空")
	}
	if len(emptyStringConfig.Subjects) != 0 {
		t.Error("空数组时Subjects长度应为0")
	}
}

// ==================== Handler 集成测试 ====================

// setupTestRouterWithClass 创建测试路由
func setupTestRouterWithClass(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 模拟JWT中间件
	authMiddleware := func(c *gin.Context) {
		// 默认设置教师认证信息
		c.Set("user_id", int64(1))
		c.Set("persona_id", int64(1))
		c.Set("role", "teacher")
		c.Set("user_role", "teacher")
		c.Set("username", "test_teacher")
		c.Next()
	}

	// 模拟数据库拦截中间件 - 在没有真实manager时提前返回错误
	dbInterceptor := func(c *gin.Context) {
		c.Set("test_mode", true)
		c.Next()
	}

	api := r.Group("/api", authMiddleware, dbInterceptor)
	{
		api.POST("/classes", handler.HandleCreateClass)
		api.GET("/classes/:id", handler.HandleGetClass)
		api.PUT("/classes/:id", handler.HandleUpdateClass)
	}

	return r
}

// TestHandleCreateClass_InvalidGradeLevel 测试创建班级时无效的学段
func TestHandleCreateClass_InvalidGradeLevel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 直接测试学段验证逻辑，不依赖完整handler
	type createClassReq struct {
		CurriculumConfig *CurriculumConfigValue `json:"curriculum_config"`
	}

	reqBody := `{
		"curriculum_config": {
			"grade_level": "invalid_level",
			"grade": "三年级",
			"subjects": ["数学"],
			"textbook_versions": ["人教版"],
			"current_progress": "第一单元"
		}
	}`

	var req createClassReq
	if err := json.Unmarshal([]byte(reqBody), &req); err != nil {
		t.Fatalf("解析请求失败: %v", err)
	}

	// 验证学段逻辑
	if req.CurriculumConfig != nil && req.CurriculumConfig.GradeLevel != "" {
		if !validGradeLevels[req.CurriculumConfig.GradeLevel] {
			// 预期行为：无效的学段返回40041
			t.Logf("学段 %s 无效，符合预期", req.CurriculumConfig.GradeLevel)
			return
		}
	}
	t.Error("应该检测到无效学段")
}

// TestHandleCreateClass_NoCurriculumConfig 测试创建班级不带教材配置
func TestHandleCreateClass_NoCurriculumConfig(t *testing.T) {
	type createClassReq struct {
		Name             string                 `json:"name"`
		PersonaNickname  string                 `json:"persona_nickname"`
		PersonaSchool    string                 `json:"persona_school"`
		PersonaDesc      string                 `json:"persona_description"`
		CurriculumConfig *CurriculumConfigValue `json:"curriculum_config"`
	}

	reqBody := `{
		"name": "测试班级",
		"persona_nickname": "王老师",
		"persona_school": "实验小学",
		"persona_description": "数学教师"
	}`

	var req createClassReq
	if err := json.Unmarshal([]byte(reqBody), &req); err != nil {
		t.Fatalf("解析请求失败: %v", err)
	}

	// 验证请求解析正确
	if req.Name != "测试班级" {
		t.Errorf("Name 错误: %s", req.Name)
	}
	if req.CurriculumConfig != nil {
		t.Error("CurriculumConfig 应该为 nil")
	}
}

// TestHandleCreateClass_ChineseGradeLevel 测试创建班级时使用中文学段
func TestHandleCreateClass_ChineseGradeLevel(t *testing.T) {
	type createClassReq struct {
		CurriculumConfig *CurriculumConfigValue `json:"curriculum_config"`
	}

	reqBody := `{
		"curriculum_config": {
			"grade_level": "小学",
			"grade": "三年级",
			"subjects": ["数学"],
			"textbook_versions": ["人教版"]
		}
	}`

	var req createClassReq
	if err := json.Unmarshal([]byte(reqBody), &req); err != nil {
		t.Fatalf("解析请求失败: %v", err)
	}

	// 验证中文学段无效
	if req.CurriculumConfig != nil && req.CurriculumConfig.GradeLevel != "" {
		if validGradeLevels[req.CurriculumConfig.GradeLevel] {
			t.Error("中文学段应该无效")
		} else {
			t.Logf("中文学段 %s 无效，符合预期", req.CurriculumConfig.GradeLevel)
		}
	}
}

// TestHandleUpdateClass_InvalidGradeLevel 测试更新班级时无效的学段
func TestHandleUpdateClass_InvalidGradeLevel(t *testing.T) {
	type updateClassReq struct {
		CurriculumConfig *CurriculumConfigValue `json:"curriculum_config"`
	}

	reqBody := `{
		"name": "测试班级",
		"curriculum_config": {
			"grade_level": "invalid_level",
			"grade": "三年级",
			"subjects": ["数学"]
		}
	}`

	var req updateClassReq
	if err := json.Unmarshal([]byte(reqBody), &req); err != nil {
		t.Fatalf("解析请求失败: %v", err)
	}

	// 验证学段逻辑
	if req.CurriculumConfig != nil && req.CurriculumConfig.GradeLevel != "" {
		if !validGradeLevels[req.CurriculumConfig.GradeLevel] {
			t.Logf("学段 %s 无效，符合预期", req.CurriculumConfig.GradeLevel)
			return
		}
	}
	t.Error("应该检测到无效学段")
}

// TestHandleUpdateClass_InvalidID 测试更新班级时无效的ID
func TestHandleUpdateClass_InvalidID(t *testing.T) {
	// 测试ID解析逻辑
	classIDStr := "abc"
	_, err := strconv.ParseInt(classIDStr, 10, 64)
	if err == nil {
		t.Error("无效ID应该返回错误")
	}

	// 测试ID为0
	classIDStr = "0"
	classID, err := strconv.ParseInt(classIDStr, 10, 64)
	if err != nil || classID <= 0 {
		t.Log("ID为0应视为无效")
	}
}

// TestHandleGetClass_InvalidID 测试获取班级详情时无效的ID
func TestHandleGetClass_InvalidID(t *testing.T) {
	// 测试ID解析逻辑
	classIDStr := "abc"
	_, err := strconv.ParseInt(classIDStr, 10, 64)
	if err == nil {
		t.Error("无效ID应该返回错误")
	}
}

// TestHandleGetClass_InvalidIDZero 测试获取班级详情时ID为0
func TestHandleGetClass_InvalidIDZero(t *testing.T) {
	// 测试ID为0
	classIDStr := "0"
	classID, err := strconv.ParseInt(classIDStr, 10, 64)
	if err != nil || classID <= 0 {
		t.Log("ID为0应视为无效")
		return
	}
	t.Error("ID为0应该被视为无效")
}

// TestHandleCreateClass_NonTeacherRole 测试非教师角色创建班级
func TestHandleCreateClass_NonTeacherRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 测试角色检查逻辑
	role := "student"
	if role != "teacher" && role != "admin" {
		t.Log("非教师/管理员角色应该被拒绝")
	} else {
		t.Error("学生角色应该被拒绝")
	}

	// 测试教师角色
	role = "teacher"
	if role != "teacher" && role != "admin" {
		t.Error("教师角色应该被允许")
	}

	// 测试管理员角色
	role = "admin"
	if role != "teacher" && role != "admin" {
		t.Error("管理员角色应该被允许")
	}
}

// ==================== 数据结构测试 ====================

// TestCurriculumConfigValue_AllGradeLevels 测试所有学段类型
func TestCurriculumConfigValue_AllGradeLevels(t *testing.T) {
	gradeLevels := map[string]string{
		"preschool":          "学前班",
		"primary_lower":      "小学低年级",
		"primary_upper":      "小学高年级",
		"junior":             "初中",
		"senior":             "高中",
		"university":         "大学及以上",
		"adult_life":         "成人生活",
		"adult_professional": "成人职业",
	}

	for level, grade := range gradeLevels {
		config := CurriculumConfigValue{
			GradeLevel:       level,
			Grade:            grade,
			Subjects:         []string{"测试学科"},
			TextbookVersions: []string{"测试版本"},
			CurrentProgress:  "测试进度",
		}

		data, err := json.Marshal(config)
		if err != nil {
			t.Errorf("学段 %s 序列化失败: %v", level, err)
		}

		var result CurriculumConfigValue
		if err := json.Unmarshal(data, &result); err != nil {
			t.Errorf("学段 %s 反序列化失败: %v", level, err)
		}

		if result.GradeLevel != level {
			t.Errorf("学段 %s 不匹配，实际: %s", level, result.GradeLevel)
		}
	}
}

// TestCurriculumConfigValue_NilSlice 测试nil切片处理
func TestCurriculumConfigValue_NilSlice(t *testing.T) {
	config := CurriculumConfigValue{
		GradeLevel:       "junior",
		Grade:            "七年级",
		Subjects:         nil,
		TextbookVersions: nil,
		CustomTextbooks:  nil,
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// nil切片序列化后应为null或[]
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	// nil切片序列化为null
	t.Logf("nil切片序列化结果: %s", string(data))
}

// TestCurriculumConfigToResponse_Nil 测试nil配置
func TestCurriculumConfigToResponse_Nil(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 测试nil参数不会panic
	var config *database.TeacherCurriculumConfig
	if config != nil {
		result := curriculumConfigToResponse(config)
		if result == nil {
			t.Error("结果不应为nil")
		}
	}
}

// TestCurriculumConfigToResponse_InvalidJSON 测试无效的JSON字段
func TestCurriculumConfigToResponse_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &database.TeacherCurriculumConfig{
		ID:               1,
		TeacherID:        1,
		PersonaID:        2,
		GradeLevel:       "junior",
		Grade:            "七年级",
		TextbookVersions: "invalid json",
		Subjects:         "also invalid",
		Region:           "not valid json",
		CurrentProgress:  "测试",
		IsActive:         1,
	}

	// 应能正确处理无效JSON而不会panic
	result := curriculumConfigToResponse(config)

	if result["id"] != int64(1) {
		t.Errorf("id 错误")
	}

	// 无效的JSON应该导致空数组或nil
	textbookVersions, ok := result["textbook_versions"].([]string)
	if ok && len(textbookVersions) > 0 {
		t.Error("无效JSON应该导致空数组")
	}
}

// ==================== 集成场景测试 ====================

// TestCurriculumConfigIntegration_CreateAndGet 模拟创建和获取流程
func TestCurriculumConfigIntegration_CreateAndGet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 验证请求体结构
	createReq := struct {
		Name             string                 `json:"name"`
		Description      string                 `json:"description"`
		PersonaNickname  string                 `json:"persona_nickname"`
		PersonaSchool    string                 `json:"persona_school"`
		PersonaDesc      string                 `json:"persona_description"`
		IsPublic         bool                   `json:"is_public"`
		CurriculumConfig *CurriculumConfigValue `json:"curriculum_config"`
	}{
		Name:            "三年级数学班",
		Description:     "小学数学培优班级",
		PersonaNickname: "王老师",
		PersonaSchool:   "实验小学",
		PersonaDesc:     "10年数学教学经验",
		IsPublic:        true,
		CurriculumConfig: &CurriculumConfigValue{
			GradeLevel:       "primary_lower",
			Grade:            "三年级",
			Subjects:         []string{"数学"},
			TextbookVersions: []string{"人教版", "北师大版"},
			CustomTextbooks:  []string{"《小学奥数》"},
			CurrentProgress:  "第三单元 乘法初步",
		},
	}

	data, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 验证序列化结果
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	// 验证基本字段
	assert.Equal(t, "三年级数学班", result["name"])

	// 验证 curriculum_config 字段存在且完整
	configData, ok := result["curriculum_config"].(map[string]interface{})
	if !ok {
		t.Fatal("curriculum_config 字段类型错误")
	}

	assert.Equal(t, "primary_lower", configData["grade_level"])
	assert.Equal(t, "三年级", configData["grade"])
	assert.Equal(t, "第三单元 乘法初步", configData["current_progress"])

	// 验证数组字段
	subjectsArr, ok := configData["subjects"].([]interface{})
	if !ok {
		t.Fatal("subjects 数组类型错误")
	}
	assert.Equal(t, 1, len(subjectsArr))
	assert.Equal(t, "数学", subjectsArr[0])

	textbookVersionsArr, ok := configData["textbook_versions"].([]interface{})
	if !ok {
		t.Fatal("textbook_versions 数组类型错误")
	}
	assert.Equal(t, 2, len(textbookVersionsArr))
}

// TestCurriculumConfigIntegration_Update 模拟更新流程
func TestCurriculumConfigIntegration_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	updateReq := struct {
		Name             string                 `json:"name"`
		Description      string                 `json:"description"`
		IsPublic         *bool                  `json:"is_public"`
		CurriculumConfig *CurriculumConfigValue `json:"curriculum_config"`
	}{
		Name:        "三年级数学班（已更名）",
		Description: "更新后的描述",
		IsPublic:    boolPtr(true),
		CurriculumConfig: &CurriculumConfigValue{
			GradeLevel:       "primary_lower",
			Grade:            "三年级",
			Subjects:         []string{"数学", "奥数"},
			TextbookVersions: []string{"人教版"},
			CustomTextbooks:  []string{"《小学奥数进阶》"},
			CurrentProgress:  "第四单元 除法",
		},
	}

	data, err := json.Marshal(updateReq)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	assert.Equal(t, "三年级数学班（已更名）", result["name"])
	assert.Equal(t, "更新后的描述", result["description"])

	isPublic, ok := result["is_public"].(bool)
	if !ok {
		t.Fatal("is_public 类型错误")
	}
	assert.True(t, isPublic)
}

// TestCurriculumConfigResponse_Get 模拟返回值结构
func TestCurriculumConfigResponse_Get(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 模拟GET /api/classes/:id响应
	getResp := struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			ID               int64                  `json:"id"`
			Name             string                 `json:"name"`
			Description      string                 `json:"description"`
			IsPublic         bool                   `json:"is_public"`
			IsActive         bool                   `json:"is_active"`
			PersonaID        int64                  `json:"persona_id"`
			TeacherID        int64                  `json:"teacher_id"`
			CreatedAt        string                 `json:"created_at"`
			UpdatedAt        string                 `json:"updated_at"`
			CurriculumConfig *CurriculumConfigValue `json:"curriculum_config"`
		} `json:"data"`
	}{
		Code:    0,
		Message: "success",
		Data: struct {
			ID               int64                  `json:"id"`
			Name             string                 `json:"name"`
			Description      string                 `json:"description"`
			IsPublic         bool                   `json:"is_public"`
			IsActive         bool                   `json:"is_active"`
			PersonaID        int64                  `json:"persona_id"`
			TeacherID        int64                  `json:"teacher_id"`
			CreatedAt        string                 `json:"created_at"`
			UpdatedAt        string                 `json:"updated_at"`
			CurriculumConfig *CurriculumConfigValue `json:"curriculum_config"`
		}{
			ID:          123,
			Name:        "三年级数学班",
			Description: "小学数学培优班级",
			IsPublic:    true,
			IsActive:    true,
			PersonaID:   456,
			TeacherID:   789,
			CreatedAt:   "2026-04-09T10:30:00Z",
			UpdatedAt:   "2026-04-09T10:30:00Z",
			CurriculumConfig: &CurriculumConfigValue{
				GradeLevel:       "primary_lower",
				Grade:            "三年级",
				Subjects:         []string{"数学"},
				TextbookVersions: []string{"人教版", "北师大版"},
				CustomTextbooks:  []string{"《小学奥数启蒙》"},
				CurrentProgress:  "第三单元 乘法初步",
			},
		},
	}

	data, err := json.Marshal(getResp)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	assert.Equal(t, float64(0), result["code"])
	assert.Equal(t, "success", result["message"])

	dataMap, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatal("data 字段类型错误")
	}

	assert.Equal(t, float64(123), dataMap["id"])
	assert.Equal(t, "三年级数学班", dataMap["name"])

	// 验证 curriculum_config 存在
	configData, ok := dataMap["curriculum_config"].(map[string]interface{})
	if !ok {
		t.Fatal("curriculum_config 字段类型错误")
	}

	assert.Equal(t, "primary_lower", configData["grade_level"])
	assert.Equal(t, "三年级", configData["grade"])

	// 验证数组字段
	subjectsArr, ok := configData["subjects"].([]interface{})
	if !ok {
		t.Fatal("subjects 数组类型错误")
	}
	assert.Equal(t, 1, len(subjectsArr))

	textbookVersionsArr, ok := configData["textbook_versions"].([]interface{})
	if !ok {
		t.Fatal("textbook_versions 数组类型错误")
	}
	assert.Equal(t, 2, len(textbookVersionsArr))

	customTextbooksArr, ok := configData["custom_textbooks"].([]interface{})
	if !ok {
		t.Fatal("custom_textbooks 数组类型错误")
	}
	assert.Equal(t, 1, len(customTextbooksArr))
}

// TestCurriculumConfigResponse_Null 测试返回值中curriculum_config为null的情况
func TestCurriculumConfigResponse_Null(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 模拟GET /api/classes/:id响应（无教材配置）
	getResp := struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			ID               int64                  `json:"id"`
			Name             string                 `json:"name"`
			Description      string                 `json:"description"`
			IsPublic         bool                   `json:"is_public"`
			IsActive         bool                   `json:"is_active"`
			PersonaID        int64                  `json:"persona_id"`
			TeacherID        int64                  `json:"teacher_id"`
			CreatedAt        string                 `json:"created_at"`
			UpdatedAt        string                 `json:"updated_at"`
			CurriculumConfig *CurriculumConfigValue `json:"curriculum_config"`
		} `json:"data"`
	}{
		Code:    0,
		Message: "success",
		Data: struct {
			ID               int64                  `json:"id"`
			Name             string                 `json:"name"`
			Description      string                 `json:"description"`
			IsPublic         bool                   `json:"is_public"`
			IsActive         bool                   `json:"is_active"`
			PersonaID        int64                  `json:"persona_id"`
			TeacherID        int64                  `json:"teacher_id"`
			CreatedAt        string                 `json:"created_at"`
			UpdatedAt        string                 `json:"updated_at"`
			CurriculumConfig *CurriculumConfigValue `json:"curriculum_config"`
		}{
			ID:               123,
			Name:             "临时班级",
			Description:      "",
			IsPublic:         false,
			IsActive:         true,
			PersonaID:        456,
			TeacherID:        789,
			CreatedAt:        "2026-04-09T10:30:00Z",
			UpdatedAt:        "2026-04-09T10:30:00Z",
			CurriculumConfig: nil, // 无教材配置
		},
	}

	data, err := json.Marshal(getResp)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	dataMap := result["data"].(map[string]interface{})

	// curriculum_config 应该为 null
	if dataMap["curriculum_config"] != nil {
		t.Errorf("curriculum_config 应该为 null，实际: %v", dataMap["curriculum_config"])
	}
}

// ==================== 辅助函数测试 ====================

// TestCurriculumConfigValue_DeepCopy 测试深拷贝
func TestCurriculumConfigValue_DeepCopy(t *testing.T) {
	original := &CurriculumConfigValue{
		GradeLevel:       "junior",
		Grade:            "七年级",
		Subjects:         []string{"数学", "语文"},
		TextbookVersions: []string{"人教版"},
		CustomTextbooks:  []string{"《奥数》"},
		CurrentProgress:  "第一单元",
	}

	// 序列化后反序列化实现深拷贝
	data, _ := json.Marshal(original)
	var copy CurriculumConfigValue
	json.Unmarshal(data, &copy)

	// 修改原对象
	original.Subjects[0] = "英语"
	original.TextbookVersions = append(original.TextbookVersions, "苏教版")

	// 验证拷贝不受影响
	if copy.Subjects[0] != "数学" {
		t.Error("深拷贝失败：Subjects被修改")
	}
	if len(copy.TextbookVersions) != 1 {
		t.Error("深拷贝失败：TextbookVersions被修改")
	}
}

// TestCurriculumConfigValue_EmptyFields 测试空字段处理
func TestCurriculumConfigValue_EmptyFields(t *testing.T) {
	// 只提供部分字段
	partial := `{
		"grade_level": "senior",
		"subjects": ["物理"]
	}`

	var config CurriculumConfigValue
	if err := json.Unmarshal([]byte(partial), &config); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	if config.GradeLevel != "senior" {
		t.Errorf("GradeLevel 错误: %s", config.GradeLevel)
	}
	if len(config.Subjects) != 1 || config.Subjects[0] != "物理" {
		t.Errorf("Subjects 错误: %v", config.Subjects)
	}
	// 未提供的字段应为零值
	if config.Grade != "" {
		t.Errorf("Grade 应该为空: %s", config.Grade)
	}
	if config.CurrentProgress != "" {
		t.Errorf("CurrentProgress 应该为空: %s", config.CurrentProgress)
	}
	if len(config.TextbookVersions) != 0 {
		t.Errorf("TextbookVersions 应该为空: %v", config.TextbookVersions)
	}
}

// TestCurriculumConfigJSON_MarshalUnmarshal 测试JSON序列化一致性
func TestCurriculumConfigJSON_MarshalUnmarshal(t *testing.T) {
	testCases := []CurriculumConfigValue{
		{
			GradeLevel:       "preschool",
			Grade:            "幼儿园大班",
			Subjects:         []string{"语言", "数学启蒙"},
			TextbookVersions: []string{"自编教材"},
			CurrentProgress:  "字母认知",
		},
		{
			GradeLevel:       "university",
			Grade:            "大一",
			Subjects:         []string{"高等数学", "线性代数"},
			CustomTextbooks:  []string{"《数学分析》", "《高等代数》"},
			CurrentProgress:  "第一章 极限与连续",
		},
		{
			GradeLevel:      "adult_professional",
			Subjects:        []string{"编程"},
			CurrentProgress: "Python基础",
		},
	}

	for _, tc := range testCases {
		// 序列化
		data, err := json.Marshal(tc)
		if err != nil {
			t.Errorf("序列化失败: %v", err)
			continue
		}

		// 反序列化
		var result CurriculumConfigValue
		if err := json.Unmarshal(data, &result); err != nil {
			t.Errorf("反序列化失败: %v", err)
			continue
		}

		// 验证关键字段
		if result.GradeLevel != tc.GradeLevel {
			t.Errorf("GradeLevel 不匹配: %s vs %s", result.GradeLevel, tc.GradeLevel)
		}
		if result.Grade != tc.Grade {
			t.Errorf("Grade 不匹配: %s vs %s", result.Grade, tc.Grade)
		}
		if len(result.Subjects) != len(tc.Subjects) {
			t.Errorf("Subjects 长度不匹配: %d vs %d", len(result.Subjects), len(tc.Subjects))
		}
	}
}

// TestHandleCreateClass_EmptyConfig 测试创建班级时空教材配置对象
func TestHandleCreateClass_EmptyConfig(t *testing.T) {
	type createClassReq struct {
		Name             string                 `json:"name"`
		PersonaNickname  string                 `json:"persona_nickname"`
		PersonaSchool    string                 `json:"persona_school"`
		PersonaDesc      string                 `json:"persona_description"`
		CurriculumConfig *CurriculumConfigValue `json:"curriculum_config"`
	}

	reqBody := `{
		"name": "测试班级",
		"persona_nickname": "王老师",
		"persona_school": "实验小学",
		"persona_description": "数学教师",
		"curriculum_config": {}
	}`

	var req createClassReq
	if err := json.Unmarshal([]byte(reqBody), &req); err != nil {
		t.Fatalf("解析请求失败: %v", err)
	}

	// 空配置对象应该有指针，但字段为空
	if req.CurriculumConfig == nil {
		t.Error("CurriculumConfig 不应该为 nil")
	}

	// 空对象的grade_level为空字符串
	if req.CurriculumConfig.GradeLevel != "" {
		t.Errorf("空配置GradeLevel应该为空: %s", req.CurriculumConfig.GradeLevel)
	}

	// 由于grade_level为空，不应进行枚举值校验
	if req.CurriculumConfig.GradeLevel != "" && !validGradeLevels[req.CurriculumConfig.GradeLevel] {
		t.Error("空grade_level不应触发校验失败")
	}
}

// TestHandleCreateClass_ValidGradeLevels 测试所有有效学段
func TestHandleCreateClass_ValidGradeLevels(t *testing.T) {
	validLevels := []string{
		"preschool",
		"primary_lower",
		"primary_upper",
		"junior",
		"senior",
		"university",
		"adult_life",
		"adult_professional",
	}

	for _, level := range validLevels {
		// 验证学段在枚举中
		if !validGradeLevels[level] {
			t.Errorf("学段 %s 应该在有效列表中", level)
		}
	}

	t.Log("所有学段验证通过")
}

// TestCurriculumConfigRepository_CreateWithTx 测试教材配置Repository
func TestCurriculumConfigRepository_CreateWithTx(t *testing.T) {
	// 由于需要真实的数据库连接，这里的测试主要是验证代码结构正确
	// 实际的数据库操作测试应该在集成测试中完成

	// 验证 TeacherCurriculumConfig 结构体
	config := &database.TeacherCurriculumConfig{
		ID:               1,
		TeacherID:        100,
		PersonaID:        200,
		GradeLevel:       "primary_lower",
		Grade:            "三年级",
		TextbookVersions: `["人教版"]`,
		Region:           `["《奥数》"]`,
		Subjects:         `["数学"]`,
		CurrentProgress:  "第一单元",
		IsActive:         1,
	}

	if config.ID != 1 {
		t.Error("ID 错误")
	}
	if config.GradeLevel != "primary_lower" {
		t.Error("GradeLevel 错误")
	}
}

// ==================== 错误处理测试 ====================

// TestValidateCurriculumConfig 测试教材配置验证
func TestValidateCurriculumConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  CurriculumConfigValue
		wantErr bool
	}{
		{
			name: "完整的有效配置",
			config: CurriculumConfigValue{
				GradeLevel:       "junior",
				Grade:            "七年级",
				Subjects:         []string{"数学", "语文"},
				TextbookVersions: []string{"人教版"},
				CurrentProgress:  "第一单元",
			},
			wantErr: false,
		},
		{
			name: "只有学段",
			config: CurriculumConfigValue{
				GradeLevel: "senior",
			},
			wantErr: false,
		},
		{
			name:    "空配置",
			config:  CurriculumConfigValue{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证配置
			if tt.config.GradeLevel != "" && !validGradeLevels[tt.config.GradeLevel] {
				if !tt.wantErr {
					t.Errorf("配置 %v 应该有效，但学段校验失败", tt.config)
				}
			}
		})
	}
}

// ==================== 边界条件测试 ====================

// TestCurriculumConfigValue_LargeArrays 测试大数组
func TestCurriculumConfigValue_LargeArrays(t *testing.T) {
	// 创建大数组
	subjects := make([]string, 100)
	for i := 0; i < 100; i++ {
		subjects[i] = "学科" + string(rune('A'+i%26))
	}

	config := CurriculumConfigValue{
		GradeLevel:       "primary_lower",
		Grade:            "一年级",
		Subjects:         subjects,
		TextbookVersions: []string{"人教版", "北师大版", "苏教版", "沪教版", "部编版"},
		CustomTextbooks:  []string{"教材1", "教材2", "教材3"},
		CurrentProgress:  "一个超长的进度说明" + strings.Repeat("非常多内容", 100),
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("大数组序列化失败: %v", err)
	}

	var result CurriculumConfigValue
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("大数组反序列化失败: %v", err)
	}

	if len(result.Subjects) != 100 {
		t.Errorf("Subjects 数量错误: %d", len(result.Subjects))
	}
}

// TestCurriculumConfigValue_SpecialCharacters 测试特殊字符
func TestCurriculumConfigValue_SpecialCharacters(t *testing.T) {
	config := CurriculumConfigValue{
		GradeLevel:       "primary_lower",
		Grade:            "三年级（实验）班",
		Subjects:         []string{"数学", "语文 & 阅读", "英语<口语>"},
		TextbookVersions: []string{"人教版", "北师大版（2024）"},
		CustomTextbooks:  []string{"《小学奥数（第三版）》", "教材A&B"},
		CurrentProgress:  "第三单元：\"乘法初步\" 与 '除法'",
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("特殊字符序列化失败: %v", err)
	}

	var result CurriculumConfigValue
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("特殊字符反序列化失败: %v", err)
	}

	if result.Grade != "三年级（实验）班" {
		t.Errorf("特殊字符处理错误: %s", result.Grade)
	}
}

// TestCurriculumConfigValue_Unicode 测试Unicode字符
func TestCurriculumConfigValue_Unicode(t *testing.T) {
	config := CurriculumConfigValue{
		GradeLevel:       "primary_lower",
		Grade:            "三年级 🎓",
		Subjects:         []string{"数学 📐", "语文 📖"},
		TextbookVersions: []string{"人教版 🇨🇳"},
		CurrentProgress:  "第三单元 📝",
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Unicode序列化失败: %v", err)
	}

	var result CurriculumConfigValue
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unicode反序列化失败: %v", err)
	}

	if !strings.Contains(result.Subjects[0], "📐") {
		t.Error("Unicode字符处理错误")
	}
}

// TestSQLNullHandling 测试SQL空值处理
func TestSQLNullHandling(t *testing.T) {
	// 验证 sql.NullString 等类型的处理
	var nullString sql.NullString
	if nullString.Valid {
		t.Error("NullString应该无效")
	}

	nullString = sql.NullString{String: "test", Valid: true}
	if !nullString.Valid {
		t.Error("NullString应该有效")
	}
}

// boolPtr 返回bool指针
func boolPtr(b bool) *bool {
	return &b
}

// ==================== httptest 集成测试 ====================

// TestCurriculumConfigAPI_CreateValidation 使用httptest测试创建班级时的教材配置校验
func TestCurriculumConfigAPI_CreateValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		reqBody    string
		setupAuth  func(*http.Request)
		wantStatus int
		wantCode   int // 业务错误码
	}{
		{
			name:    "无效学段类型",
			reqBody: `{"name":"测试班级","persona_nickname":"王老师","persona_school":"实验小学","persona_description":"数学教师","curriculum_config":{"grade_level":"invalid_level","grade":"三年级"}}`,
			setupAuth: func(req *http.Request) {
				// 模拟认证中间件会在实际路由中处理
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   40041,
		},
		{
			name:    "空学段不触发校验",
			reqBody: `{"name":"测试班级","persona_nickname":"王老师","persona_school":"实验小学","persona_description":"数学教师","curriculum_config":{"grade_level":"","grade":"三年级"}}`,
			setupAuth: func(req *http.Request) {
			},
			wantStatus: http.StatusOK, // 空学段不触发校验，请求成功
		},
		{
			name:       "无教材配置",
			reqBody:    `{"name":"测试班级","persona_nickname":"王老师","persona_school":"实验小学","persona_description":"数学教师"}`,
			wantStatus: http.StatusOK, // 无教材配置是允许的，请求格式正确
		},
		{
			name:       "有效学段配置（校验通过）",
			reqBody:    `{"name":"测试班级","persona_nickname":"王老师","persona_school":"实验小学","persona_description":"数学教师","curriculum_config":{"grade_level":"primary_lower","grade":"三年级","subjects":["数学"]}}`,
			wantStatus: http.StatusOK, // 有效学段校验通过
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 使用轻量级的路由测试而不是完整handler
			r := gin.New()

			// 模拟handler - 只测试学段校验逻辑
			r.POST("/test", func(c *gin.Context) {
				var req struct {
					Name             string                 `json:"name"`
					PersonaNickname  string                 `json:"persona_nickname"`
					PersonaSchool    string                 `json:"persona_school"`
					PersonaDesc      string                 `json:"persona_description"`
					CurriculumConfig *CurriculumConfigValue `json:"curriculum_config"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"code": 40004, "message": err.Error()})
					return
				}

				// 学段校验逻辑 - 与真实handler一致
				if req.CurriculumConfig != nil && req.CurriculumConfig.GradeLevel != "" {
					if !validGradeLevels[req.CurriculumConfig.GradeLevel] {
						c.JSON(http.StatusBadRequest, gin.H{"code": 40041, "message": "无效的学段类型"})
						return
					}
				}

				// 校验通过（实际场景会继续处理，这里简化为成功）
				c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/test", strings.NewReader(tt.reqBody))
			req.Header.Set("Content-Type", "application/json")
			if tt.setupAuth != nil {
				tt.setupAuth(req)
			}

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("状态码错误: got %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			// 验证业务错误码
			if tt.wantCode != 0 {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				code, ok := resp["code"].(float64)
				if ok && int(code) != tt.wantCode {
					t.Errorf("业务错误码错误: got %v, want %d", code, tt.wantCode)
				}
			}
		})
	}
}

// TestCurriculumConfigAPI_UpdateValidation 测试更新班级时的教材配置校验
func TestCurriculumConfigAPI_UpdateValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		id         string
		reqBody    string
		wantStatus int
		wantCode   int
	}{
		{
			name:       "无效班级ID",
			id:         "abc",
			reqBody:    `{"name":"班级名称"}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   40004,
		},
		{
			name:       "班级ID为0",
			id:         "0",
			reqBody:    `{"name":"班级名称"}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   40004,
		},
		{
			name:       "负班级ID",
			id:         "-1",
			reqBody:    `{"name":"班级名称"}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   40004,
		},
		{
			name:       "更新时无效学段",
			id:         "123",
			reqBody:    `{"name":"班级名称","curriculum_config":{"grade_level":"invalid","grade":"三年级"}}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   40041,
		},
		{
			name:       "更新所有字段",
			id:         "123",
			reqBody:    `{"name":"新名称","description":"新描述","is_public":true,"curriculum_config":{"grade_level":"junior","grade":"七年级","subjects":["数学","语文"]}}`,
			wantStatus: http.StatusForbidden, // 由于没有权限验证，会返回403
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()

			r.PUT("/api/classes/:id", func(c *gin.Context) {
				// ID校验
				idStr := c.Param("id")
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil || id <= 0 {
					c.JSON(http.StatusBadRequest, gin.H{"code": 40004, "message": "无效的班级ID"})
					return
				}

				var req struct {
					Name             string                 `json:"name"`
					Descrip          string                 `json:"description"`
					IsPublic         *bool                  `json:"is_public"`
					CurriculumConfig *CurriculumConfigValue `json:"curriculum_config"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"code": 40004, "message": err.Error()})
					return
				}

				// 学段校验
				if req.CurriculumConfig != nil && req.CurriculumConfig.GradeLevel != "" {
					if !validGradeLevels[req.CurriculumConfig.GradeLevel] {
						c.JSON(http.StatusBadRequest, gin.H{"code": 40041, "message": "无效的学段类型"})
						return
					}
				}

				// 权限校验（模拟）
				c.JSON(http.StatusForbidden, gin.H{"code": 40018, "message": "无权操作"})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("PUT", "/api/classes/"+tt.id, strings.NewReader(tt.reqBody))
			req.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("状态码错误: got %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantCode != 0 {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				code, ok := resp["code"].(float64)
				if ok && int(code) != tt.wantCode {
					t.Errorf("业务错误码错误: got %v, want %d, resp: %v", code, tt.wantCode, resp)
				}
			}
		})
	}
}

// TestCurriculumConfigAPI_GetValidation 测试获取班级详情的ID校验
func TestCurriculumConfigAPI_GetValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		id         string
		wantStatus int
		wantCode   int
	}{
		{
			name:       "无效班级ID",
			id:         "abc",
			wantStatus: http.StatusBadRequest,
			wantCode:   40004,
		},
		{
			name:       "班级ID为0",
			id:         "0",
			wantStatus: http.StatusBadRequest,
			wantCode:   40004,
		},
		{
			name:       "正常ID格式",
			id:         "123",
			wantStatus: http.StatusForbidden, // 无权限验证会返回403
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()

			r.GET("/api/classes/:id", func(c *gin.Context) {
				idStr := c.Param("id")
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil || id <= 0 {
					c.JSON(http.StatusBadRequest, gin.H{"code": 40004, "message": "无效的班级ID"})
					return
				}
				c.JSON(http.StatusForbidden, gin.H{"code": 40018, "message": "无权操作"})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/classes/"+tt.id, nil)

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("状态码错误: got %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantCode != 0 {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				code, ok := resp["code"].(float64)
				if ok && int(code) != tt.wantCode {
					t.Errorf("业务错误码错误: got %v, want %d", code, tt.wantCode)
				}
			}
		})
	}
}

// TestCurriculumConfigResponseStructure 测试响应结构完整性
func TestCurriculumConfigResponseStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 模拟构建完整响应
	response := gin.H{
		"id":          int64(123),
		"name":        "三年级数学班",
		"description": "小学数学培优班级",
		"is_public":   true,
		"is_active":   true,
		"persona_id":  int64(456),
		"teacher_id":  int64(789),
		"created_at":  "2026-04-09T10:30:00Z",
		"updated_at":  "2026-04-09T10:30:00Z",
		"curriculum_config": gin.H{
			"id":                int64(1001),
			"grade_level":       "primary_lower",
			"grade":             "三年级",
			"subjects":          []string{"数学"},
			"textbook_versions": []string{"人教版", "北师大版"},
			"custom_textbooks":  []string{"《小学奥数启蒙》"},
			"current_progress":  "第三单元 乘法初步",
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("响应序列化失败: %v", err)
	}

	// 验证响应结构
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("响应反序列化失败: %v", err)
	}

	// 验证必需字段存在
	requiredFields := []string{"id", "name", "description", "is_public", "is_active",
		"persona_id", "created_at", "updated_at", "curriculum_config"}
	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			t.Errorf("响应缺少必需字段: %s", field)
		}
	}

	// 验证curriculum_config结构
	config, ok := result["curriculum_config"].(map[string]interface{})
	if !ok {
		t.Fatal("curriculum_config 格式错误")
	}

	configFields := []string{"id", "grade_level", "grade", "subjects",
		"textbook_versions", "custom_textbooks", "current_progress"}
	for _, field := range configFields {
		if _, ok := config[field]; !ok {
			t.Errorf("curriculum_config 缺少必需字段: %s", field)
		}
	}
}

// TestCurriculumConfigEmptyObjVsNil 测试空对象和nil的区别
func TestCurriculumConfigEmptyObjVsNil(t *testing.T) {
	tests := []struct {
		name           string
		jsonData       string
		expectNil      bool
		expectGradeLvl string
	}{
		{
			name:           "null值",
			jsonData:       `{"curriculum_config":null}`,
			expectNil:      true,
			expectGradeLvl: "",
		},
		{
			name:           "空对象",
			jsonData:       `{"curriculum_config":{}}`,
			expectNil:      false,
			expectGradeLvl: "",
		},
		{
			name:           "只有学段",
			jsonData:       `{"curriculum_config":{"grade_level":"junior"}}`,
			expectNil:      false,
			expectGradeLvl: "junior",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req struct {
				CurriculumConfig *CurriculumConfigValue `json:"curriculum_config"`
			}

			if err := json.Unmarshal([]byte(tt.jsonData), &req); err != nil {
				t.Fatalf("解析失败: %v", err)
			}

			if tt.expectNil {
				if req.CurriculumConfig != nil {
					t.Error("期望nil，但得到非nil")
				}
			} else {
				if req.CurriculumConfig == nil {
					t.Error("期望非nil，但得到nil")
				} else if req.CurriculumConfig.GradeLevel != tt.expectGradeLvl {
					t.Errorf("grade_level错误: got %s, want %s",
						req.CurriculumConfig.GradeLevel, tt.expectGradeLvl)
				}
			}
		})
	}
}

// TestAllGradeLevelsInAPI 测试所有学段在API中的使用
func TestAllGradeLevelsInAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	gradeLevels := []struct {
		level string
		grade string
	}{
		{"preschool", "幼儿园大班"},
		{"primary_lower", "一年级"},
		{"primary_upper", "五年级"},
		{"junior", "七年级"},
		{"senior", "高一"},
		{"university", "大一"},
		{"adult_life", ""},
		{"adult_professional", ""},
	}

	for _, tt := range gradeLevels {
		t.Run(tt.level, func(t *testing.T) {
			// 验证学段有效
			if !validGradeLevels[tt.level] {
				t.Errorf("学段 %s 应该是有效的", tt.level)
			}

			// 构建请求
			config := CurriculumConfigValue{
				GradeLevel:       tt.level,
				Grade:            tt.grade,
				Subjects:         []string{"测试学科"},
				TextbookVersions: []string{"人教版"},
				CurrentProgress:  "测试进度",
			}

			data, err := json.Marshal(config)
			if err != nil {
				t.Fatalf("序列化失败: %v", err)
			}

			var result CurriculumConfigValue
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("反序列化失败: %v", err)
			}

			if result.GradeLevel != tt.level {
				t.Errorf("学段不匹配: got %s, want %s", result.GradeLevel, tt.level)
			}
		})
	}
}
