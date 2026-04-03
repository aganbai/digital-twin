package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"digital-twin/src/backend/database"

	"github.com/gin-gonic/gin"
)

// HandleCreateCurriculumConfig 创建教材配置
// POST /api/curriculum-configs
func (h *Handler) HandleCreateCurriculumConfig(c *gin.Context) {
	var req struct {
		PersonaID        int64                  `json:"persona_id" binding:"required"`
		GradeLevel       string                 `json:"grade_level"`
		Grade            string                 `json:"grade"`
		TextbookVersions []string               `json:"textbook_versions"`
		Subjects         []string               `json:"subjects"`
		CurrentProgress  map[string]interface{} `json:"current_progress"`
		Region           string                 `json:"region"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	// 自动推断学段
	gradeLevel := req.GradeLevel
	if gradeLevel == "" && req.Grade != "" {
		gradeLevel = inferGradeLevel(req.Grade)
	}

	// 验证学段枚举值（R1）
	if gradeLevel != "" && !validGradeLevels[gradeLevel] {
		Error(c, http.StatusBadRequest, 40041, "无效的学段类型: "+gradeLevel)
		return
	}

	// 序列化 JSON 字段
	textbookVersionsJSON, _ := json.Marshal(req.TextbookVersions)
	subjectsJSON, _ := json.Marshal(req.Subjects)
	// current_progress 为 JSON 字符串存储（前端传 object）
	currentProgressJSON := "{}"
	if req.CurrentProgress != nil {
		if data, err := json.Marshal(req.CurrentProgress); err == nil {
			currentProgressJSON = string(data)
		}
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	repo := database.NewCurriculumConfigRepository(db)
	config := &database.TeacherCurriculumConfig{
		TeacherID:        userIDInt64,
		PersonaID:        req.PersonaID,
		GradeLevel:       gradeLevel,
		Grade:            req.Grade,
		TextbookVersions: string(textbookVersionsJSON),
		Subjects:         string(subjectsJSON),
		CurrentProgress:  currentProgressJSON,
		Region:           req.Region,
	}

	id, err := repo.Create(config)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "创建教材配置失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"id":          id,
		"grade_level": gradeLevel,
	})
}

// HandleGetCurriculumConfigs 获取教材配置列表
// GET /api/curriculum-configs
func (h *Handler) HandleGetCurriculumConfigs(c *gin.Context) {
	personaIDStr := c.Query("persona_id")
	personaID, err := strconv.ParseInt(personaIDStr, 10, 64)
	if err != nil || personaID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "缺少或无效的 persona_id 参数")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	repo := database.NewCurriculumConfigRepository(db)
	configs, err := repo.ListByPersonaID(personaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询教材配置失败: "+err.Error())
		return
	}

	// 解析 JSON 字段
	var items []gin.H
	for _, cfg := range configs {
		var textbookVersions []string
		var subjects []string
		var currentProgress map[string]interface{}
		_ = json.Unmarshal([]byte(cfg.TextbookVersions), &textbookVersions)
		_ = json.Unmarshal([]byte(cfg.Subjects), &subjects)
		_ = json.Unmarshal([]byte(cfg.CurrentProgress), &currentProgress)

		items = append(items, gin.H{
			"id":                cfg.ID,
			"persona_id":        cfg.PersonaID,
			"grade_level":       cfg.GradeLevel,
			"grade":             cfg.Grade,
			"textbook_versions": textbookVersions,
			"subjects":          subjects,
			"current_progress":  currentProgress,
			"region":            cfg.Region,
			"is_active":         cfg.IsActive == 1,
			"created_at":        cfg.CreatedAt,
			"updated_at":        cfg.UpdatedAt,
		})
	}

	Success(c, gin.H{"items": items})
}

// HandleDeleteCurriculumConfig 删除教材配置
// DELETE /api/curriculum-configs/:id
func (h *Handler) HandleDeleteCurriculumConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的配置ID")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	repo := database.NewCurriculumConfigRepository(db)
	if err := repo.Delete(id); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "删除教材配置失败: "+err.Error())
		return
	}

	Success(c, gin.H{"message": "删除成功"})
}

// inferGradeLevel 根据年级自动推断学段
func inferGradeLevel(grade string) string {
	gradeMap := map[string]string{
		"学前班": "preschool", "幼儿园大班": "preschool",
		"一年级": "primary_lower", "二年级": "primary_lower", "三年级": "primary_lower",
		"1年级": "primary_lower", "2年级": "primary_lower", "3年级": "primary_lower",
		"四年级": "primary_upper", "五年级": "primary_upper", "六年级": "primary_upper",
		"4年级": "primary_upper", "5年级": "primary_upper", "6年级": "primary_upper",
		"七年级": "junior", "八年级": "junior", "九年级": "junior",
		"7年级": "junior", "8年级": "junior", "9年级": "junior",
		"初一": "junior", "初二": "junior", "初三": "junior",
		"十年级": "senior", "十一年级": "senior", "十二年级": "senior",
		"10年级": "senior", "11年级": "senior", "12年级": "senior",
		"高一": "senior", "高二": "senior", "高三": "senior",
		"大一": "university", "大二": "university", "大三": "university", "大四": "university",
		"研一": "university", "研二": "university", "研三": "university",
	}
	if level, ok := gradeMap[grade]; ok {
		return level
	}
	return ""
}
