package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"digital-twin/src/backend/database"

	"github.com/gin-gonic/gin"
)

// validGradeLevels 有效的学段枚举值
var validGradeLevels = map[string]bool{
	"preschool":          true,
	"primary_lower":      true,
	"primary_upper":      true,
	"junior":             true,
	"senior":             true,
	"university":         true,
	"adult_life":         true,
	"adult_professional": true,
}

// HandleUpdateCurriculumConfig 更新教材配置
// PUT /api/curriculum-configs/:id
func (h *Handler) HandleUpdateCurriculumConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的配置ID")
		return
	}

	var req struct {
		GradeLevel       string   `json:"grade_level"`
		Grade            string   `json:"grade"`
		TextbookVersions []string `json:"textbook_versions"`
		Subjects         []string `json:"subjects"`
		CurrentProgress  string   `json:"current_progress"`
		Region           string   `json:"region"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	// 验证学段枚举值
	gradeLevel := req.GradeLevel
	if gradeLevel == "" && req.Grade != "" {
		gradeLevel = inferGradeLevel(req.Grade)
	}
	if gradeLevel != "" && !validGradeLevels[gradeLevel] {
		Error(c, http.StatusBadRequest, 40041, "无效的学段类型: "+gradeLevel)
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	repo := database.NewCurriculumConfigRepository(db)

	// 查询现有配置
	existing, err := repo.GetByID(id)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询配置失败: "+err.Error())
		return
	}
	if existing == nil {
		Error(c, http.StatusNotFound, 40004, "配置不存在")
		return
	}

	// 验证权限：只能更新自己的配置
	userID, _ := c.Get("user_id")
	userIDInt64, _ := userID.(int64)
	if existing.TeacherID != userIDInt64 {
		Error(c, http.StatusForbidden, 40003, "无权修改此配置")
		return
	}

	// 更新字段
	if gradeLevel != "" {
		existing.GradeLevel = gradeLevel
	}
	if req.Grade != "" {
		existing.Grade = req.Grade
	}
	if req.TextbookVersions != nil {
		textbookVersionsJSON, _ := json.Marshal(req.TextbookVersions)
		existing.TextbookVersions = string(textbookVersionsJSON)
	}
	if req.Subjects != nil {
		subjectsJSON, _ := json.Marshal(req.Subjects)
		existing.Subjects = string(subjectsJSON)
	}
	if req.CurrentProgress != "" {
		existing.CurrentProgress = req.CurrentProgress
	}
	if req.Region != "" {
		existing.Region = req.Region
	}

	if err := repo.Update(existing); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "更新教材配置失败: "+err.Error())
		return
	}

	// 解析JSON字段返回
	var textbookVersions []string
	var subjects []string
	_ = json.Unmarshal([]byte(existing.TextbookVersions), &textbookVersions)
	_ = json.Unmarshal([]byte(existing.Subjects), &subjects)

	Success(c, gin.H{
		"id":                existing.ID,
		"teacher_id":        existing.TeacherID,
		"persona_id":        existing.PersonaID,
		"grade_level":       existing.GradeLevel,
		"grade":             existing.Grade,
		"textbook_versions": textbookVersions,
		"subjects":          subjects,
		"current_progress":  existing.CurrentProgress,
		"region":            existing.Region,
		"is_active":         existing.IsActive == 1,
		"created_at":        existing.CreatedAt,
		"updated_at":        existing.UpdatedAt,
	})
}

// HandleGetCurriculumVersions 获取教材版本列表
// GET /api/curriculum-versions
func (h *Handler) HandleGetCurriculumVersions(c *gin.Context) {
	gradeLevel := c.Query("grade_level")

	// 成人学段不适用教材版本，返回空列表
	if gradeLevel == "adult_life" || gradeLevel == "adult_professional" {
		Success(c, gin.H{
			"versions":    []string{},
			"recommended": "",
		})
		return
	}

	// 所有可用教材版本
	allVersions := []string{"人教版", "北师大版", "苏教版", "沪教版", "部编版", "外研版"}

	// 按地区推荐默认版本
	region := c.Query("region")
	recommended := "人教版" // 默认推荐

	regionRecommendations := map[string]string{
		"北京": "人教版",
		"上海": "沪教版",
		"江苏": "苏教版",
		"南京": "苏教版",
		"广东": "人教版",
		"浙江": "人教版",
	}

	if region != "" {
		if rec, ok := regionRecommendations[region]; ok {
			recommended = rec
		}
	}

	Success(c, gin.H{
		"versions":    allVersions,
		"recommended": recommended,
	})
}
