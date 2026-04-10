package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"digital-twin/src/backend/database"

	"github.com/gin-gonic/gin"
)

// ======================== V2.0 迭代9 M6: 头像点击 API ========================

// HandleGetClassForStudent 学生查看班级详情
// GET /api/classes/:id
// 权限：学生必须是班级成员才能查看
// V2.0 迭代13: 扩展返回教材配置信息
func (h *Handler) HandleGetClassForStudent(c *gin.Context) {
	classIDStr := c.Param("id")
	classID, err := strconv.ParseInt(classIDStr, 10, 64)
	if err != nil || classID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的班级ID")
		return
	}

	personaID, _ := c.Get("persona_id")
	personaIDInt64, ok := personaID.(int64)
	if !ok || personaIDInt64 <= 0 {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	role, _ := c.Get("role")
	roleStr, _ := role.(string)
	if roleStr != "student" {
		Error(c, http.StatusForbidden, 40018, "只有学生分身才能查看班级详情")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	classRepo := database.NewClassRepository(db)

	// 校验学生是该班级成员
	isMember, err := classRepo.IsStudentInClass(classID, personaIDInt64)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "检查班级成员失败: "+err.Error())
		return
	}
	if !isMember {
		Error(c, http.StatusForbidden, 40018, "您不是该班级的成员，无法查看")
		return
	}

	// 获取班级详情
	detail, err := classRepo.GetClassDetailForStudent(classID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询班级详情失败: "+err.Error())
		return
	}
	if detail == nil {
		Error(c, http.StatusNotFound, 40017, "班级不存在")
		return
	}

	// 从班级名称推导科目（简单实现：取第一个关键词）
	subject := extractSubject(detail.Name)

	// 查询关联的教师分身ID
	var teacherPersonaID int64
	err = db.QueryRow("SELECT persona_id FROM classes WHERE id = ?", classID).Scan(&teacherPersonaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询班级教师信息失败: "+err.Error())
		return
	}

	// 查询教材配置
	curriculumRepo := database.NewCurriculumConfigRepository(db)
	curriculumConfig, _ := curriculumRepo.GetActiveByPersonaID(teacherPersonaID)

	resp := gin.H{
		"id":           detail.ID,
		"name":         detail.Name,
		"subject":      subject,
		"description":  detail.Description,
		"teacher_name": detail.TeacherName,
		"member_count": detail.MemberCount,
		"created_at":   detail.CreatedAt,
	}

	// 如有教材配置，解析并返回
	if curriculumConfig != nil {
		var textbookVersions []string
		var subjects []string
		var customTextbooks []string
		_ = json.Unmarshal([]byte(curriculumConfig.TextbookVersions), &textbookVersions)
		_ = json.Unmarshal([]byte(curriculumConfig.Subjects), &subjects)
		_ = json.Unmarshal([]byte(curriculumConfig.Region), &customTextbooks)

		resp["curriculum_config"] = gin.H{
			"id":                curriculumConfig.ID,
			"grade_level":       curriculumConfig.GradeLevel,
			"grade":             curriculumConfig.Grade,
			"subjects":          subjects,
			"textbook_versions": textbookVersions,
			"custom_textbooks":  customTextbooks,
			"current_progress":  curriculumConfig.CurrentProgress,
		}
	}

	Success(c, resp)
}

// HandleGetStudentProfile 教师查看学生详情
// GET /api/students/:id/profile
// 权限：教师必须与该学生有师生关系才能查看
func (h *Handler) HandleGetStudentProfile(c *gin.Context) {
	studentPersonaIDStr := c.Param("id")
	studentPersonaID, err := strconv.ParseInt(studentPersonaIDStr, 10, 64)
	if err != nil || studentPersonaID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的学生ID")
		return
	}

	personaID, _ := c.Get("persona_id")
	teacherPersonaID, ok := personaID.(int64)
	if !ok || teacherPersonaID <= 0 {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	role, _ := c.Get("role")
	roleStr, _ := role.(string)
	if roleStr != "teacher" {
		Error(c, http.StatusForbidden, 40018, "只有教师分身才能查看学生详情")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	relationRepo := database.NewRelationRepository(db)
	personaRepo := database.NewPersonaRepository(db)

	// 校验教师与该学生有师生关系
	hasRelation, err := relationRepo.IsApprovedByPersonas(teacherPersonaID, studentPersonaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "检查师生关系失败: "+err.Error())
		return
	}
	if !hasRelation {
		Error(c, http.StatusForbidden, 40018, "您与该学生没有师生关系，无法查看")
		return
	}

	// 获取学生分身信息
	studentPersona, err := personaRepo.GetByID(studentPersonaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询学生信息失败: "+err.Error())
		return
	}
	if studentPersona == nil {
		Error(c, http.StatusNotFound, 40013, "学生分身不存在")
		return
	}

	// 获取学生在该教师班级中的详细信息（年龄、性别、家庭情况、评语）
	profile, err := getStudentProfileFromDB(db, teacherPersonaID, studentPersonaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询学生画像失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"id":                 studentPersonaID,
		"nickname":           studentPersona.Nickname,
		"age":                profile.Age,
		"gender":             profile.Gender,
		"family_info":        profile.FamilyInfo,
		"teacher_evaluation": profile.TeacherEvaluation,
		"class_name":         profile.ClassName,
	})
}

// HandleUpdateStudentEvaluation 教师更新学生评语
// PUT /api/students/:id/evaluation
// 权限：教师必须与该学生有师生关系才能修改
func (h *Handler) HandleUpdateStudentEvaluation(c *gin.Context) {
	studentPersonaIDStr := c.Param("id")
	studentPersonaID, err := strconv.ParseInt(studentPersonaIDStr, 10, 64)
	if err != nil || studentPersonaID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的学生ID")
		return
	}

	personaID, _ := c.Get("persona_id")
	teacherPersonaID, ok := personaID.(int64)
	if !ok || teacherPersonaID <= 0 {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	role, _ := c.Get("role")
	roleStr, _ := role.(string)
	if roleStr != "teacher" {
		Error(c, http.StatusForbidden, 40018, "只有教师分身才能更新学生评语")
		return
	}

	var req struct {
		Evaluation string `json:"evaluation" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	evaluation := strings.TrimSpace(req.Evaluation)
	if evaluation == "" {
		Error(c, http.StatusBadRequest, 40004, "评语内容不能为空")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	relationRepo := database.NewRelationRepository(db)

	// 校验教师与该学生有师生关系
	hasRelation, err := relationRepo.IsApprovedByPersonas(teacherPersonaID, studentPersonaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "检查师生关系失败: "+err.Error())
		return
	}
	if !hasRelation {
		Error(c, http.StatusForbidden, 40018, "您与该学生没有师生关系，无法修改评语")
		return
	}

	// 更新学生评语（存储在 teacher_student_relations 表的 comment 字段）
	// 或者存储在 class_members 表的 teacher_evaluation 字段
	// 这里使用 class_members 表，因为评语是与班级关联的
	err = updateStudentEvaluationInDB(db, teacherPersonaID, studentPersonaID, evaluation)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "更新学生评语失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"message": "评语更新成功",
	})
}

// ======================== 辅助函数 ========================

// extractSubject 从班级名称中提取科目
func extractSubject(className string) string {
	// 简单实现：匹配常见科目关键词
	subjects := []string{"美术", "音乐", "数学", "语文", "英语", "物理", "化学", "生物", "历史", "地理", "政治"}
	classNameLower := strings.ToLower(className)
	for _, subject := range subjects {
		if strings.Contains(classNameLower, subject) {
			return subject
		}
	}
	// 默认返回空字符串
	return ""
}

// getStudentProfileFromDB 从数据库获取学生画像详情
func getStudentProfileFromDB(db *sql.DB, teacherPersonaID, studentPersonaID int64) (*database.StudentProfileDetail, error) {
	profile := &database.StudentProfileDetail{
		ID: studentPersonaID,
	}

	// 从 class_members 表获取学生在该教师班级中的信息
	query := `SELECT cm.age, COALESCE(cm.gender, ''), COALESCE(cm.family_info, ''), 
		COALESCE(cm.teacher_evaluation, ''), COALESCE(c.name, '')
		FROM class_members cm
		JOIN classes c ON cm.class_id = c.id
		WHERE c.persona_id = ? AND cm.student_persona_id = ?
		LIMIT 1`

	var age sql.NullInt64
	var gender, familyInfo, teacherEval, className string
	err := db.QueryRow(query, teacherPersonaID, studentPersonaID).Scan(&age, &gender, &familyInfo, &teacherEval, &className)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if age.Valid {
		profile.Age = int(age.Int64)
	}
	profile.Gender = gender
	profile.FamilyInfo = familyInfo
	profile.TeacherEvaluation = teacherEval
	profile.ClassName = className

	return profile, nil
}

// updateStudentEvaluationInDB 更新学生评语到数据库
func updateStudentEvaluationInDB(db *sql.DB, teacherPersonaID, studentPersonaID int64, evaluation string) error {
	// 更新 class_members 表中的 teacher_evaluation 字段
	query := `UPDATE class_members SET teacher_evaluation = ? 
		WHERE class_id IN (SELECT id FROM classes WHERE persona_id = ?) 
		AND student_persona_id = ?`

	result, err := db.Exec(query, evaluation, teacherPersonaID, studentPersonaID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		// 如果没有更新到 class_members，尝试更新 teacher_student_relations 表
		query2 := `UPDATE teacher_student_relations SET comment = ? 
			WHERE teacher_persona_id = ? AND student_persona_id = ?`
		_, err = db.Exec(query2, evaluation, teacherPersonaID, studentPersonaID)
		if err != nil {
			return err
		}
	}

	return nil
}
