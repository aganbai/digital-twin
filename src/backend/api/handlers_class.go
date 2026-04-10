package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"digital-twin/src/backend/database"

	"github.com/gin-gonic/gin"
)

// CurriculumConfigValue 教材配置值对象（用于班级API）
type CurriculumConfigValue struct {
	GradeLevel       string   `json:"grade_level"`
	Grade            string   `json:"grade"`
	Subjects         []string `json:"subjects"`
	TextbookVersions []string `json:"textbook_versions"`
	CustomTextbooks  []string `json:"custom_textbooks"`
	CurrentProgress  string   `json:"current_progress"`
}

// curriculumConfigToResponse 将数据库教材配置转换为响应格式
func curriculumConfigToResponse(config *database.TeacherCurriculumConfig) gin.H {
	var textbookVersions []string
	var subjects []string
	var customTextbooks []string
	_ = json.Unmarshal([]byte(config.TextbookVersions), &textbookVersions)
	_ = json.Unmarshal([]byte(config.Subjects), &subjects)
	_ = json.Unmarshal([]byte(config.Region), &customTextbooks)

	return gin.H{
		"id":                config.ID,
		"grade_level":       config.GradeLevel,
		"grade":             config.Grade,
		"subjects":          subjects,
		"textbook_versions": textbookVersions,
		"custom_textbooks":  customTextbooks,
		"current_progress":  config.CurrentProgress,
	}
}

// ======================== 班级管理接口 (V2.0 迭代2) ========================

// HandleCreateClass 创建班级
// POST /api/classes
// V2.0 迭代11重构：创建班级时同步创建分身
func (h *Handler) HandleCreateClass(c *gin.Context) {
	// 角色校验：仅教师可创建班级（防御深度，配合路由层 RoleRequired 中间件）
	role, _ := c.Get("role")
	roleStr, _ := role.(string)
	if roleStr != "teacher" && roleStr != "admin" {
		Error(c, http.StatusForbidden, 40301, "仅教师角色可创建班级")
		return
	}

	var req struct {
		Name               string                 `json:"name" binding:"required"`
		Description        string                 `json:"description"`
		PersonaNickname    string                 `json:"persona_nickname" binding:"required"`
		PersonaSchool      string                 `json:"persona_school" binding:"required"`
		PersonaDescription string                 `json:"persona_description" binding:"required"`
		IsPublic           *bool                  `json:"is_public"`
		CurriculumConfig   *CurriculumConfigValue `json:"curriculum_config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	// 获取当前用户信息
	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok || userIDInt64 <= 0 {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	// 校验班级名称
	name := strings.TrimSpace(req.Name)
	if name == "" {
		Error(c, http.StatusBadRequest, 40004, "班级名称不能为空")
		return
	}

	// 校验分身信息
	personaNickname := strings.TrimSpace(req.PersonaNickname)
	if personaNickname == "" {
		Error(c, http.StatusBadRequest, 40004, "分身昵称不能为空")
		return
	}
	personaSchool := strings.TrimSpace(req.PersonaSchool)
	if personaSchool == "" {
		Error(c, http.StatusBadRequest, 40004, "学校名称不能为空")
		return
	}
	personaDescription := strings.TrimSpace(req.PersonaDescription)
	if personaDescription == "" {
		Error(c, http.StatusBadRequest, 40004, "分身描述不能为空")
		return
	}

	// 设置默认值
	isPublic := 1
	if req.IsPublic != nil && !*req.IsPublic {
		isPublic = 0
	}

	personaRepo := database.NewPersonaRepository(db)
	classRepo := database.NewClassRepository(db)
	curriculumRepo := database.NewCurriculumConfigRepository(db)

	// 检查教师是否已有同名班级（使用用户ID而非分身ID）
	// 在新模式下，每个班级对应一个分身，所以班级名称在用户维度下唯一
	exists, err := classRepo.CheckNameExistsByUserID(userIDInt64, name, 0)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "检查班级名称失败: "+err.Error())
		return
	}
	if exists {
		Error(c, http.StatusConflict, 40030, "该班级名称已存在")
		return
	}

	// 开启事务
	tx, err := db.Begin()
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "开启事务失败: "+err.Error())
		return
	}
	defer tx.Rollback()

	// 步骤1: 验证教材配置学段（如有提供）
	if req.CurriculumConfig != nil && req.CurriculumConfig.GradeLevel != "" {
		if !validGradeLevels[req.CurriculumConfig.GradeLevel] {
			Error(c, http.StatusBadRequest, 40041, "无效的学段类型: "+req.CurriculumConfig.GradeLevel)
			return
		}
	}

	// 步骤2: 先创建班级专属分身（需要在创建班级之前，因为班级需要persona_id）
	persona := &database.Persona{
		UserID:       userIDInt64,
		Role:         "teacher",
		Nickname:     personaNickname,
		School:       personaSchool,
		Description:  personaDescription,
		IsPublic:     isPublic, // 分身的公开状态继承自班级
		BoundClassID: nil,      // 暂时为nil，创建班级后更新
	}
	personaID, err := personaRepo.CreateWithTx(tx, persona)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "创建分身失败: "+err.Error())
		return
	}

	// 步骤2: 创建班级记录（使用创建的分身ID）
	class := &database.Class{
		PersonaID:   personaID,
		Name:        name,
		Description: strings.TrimSpace(req.Description),
		IsPublic:    isPublic,
	}
	classID, err := classRepo.CreateWithTx(tx, class)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "创建班级失败: "+err.Error())
		return
	}

	// 步骤3: 更新分身的bound_class_id
	_, err = tx.Exec(`UPDATE personas SET bound_class_id = ? WHERE id = ?`, classID, personaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "更新分身班级关联失败: "+err.Error())
		return
	}

	// 步骤4: 如果教师有自测学生，自动将自测学生加入该班级
	// 注意：这部分功能依赖M4模块，当前先预留
	testStudentPersonaID, _ := h.getTestStudentPersonaID(userIDInt64, db)
	if testStudentPersonaID > 0 {
		// 自动加入班级
		_, err = tx.Exec(
			`INSERT INTO class_members (class_id, student_persona_id, joined_at) VALUES (?, ?, ?)`,
			classID, testStudentPersonaID, time.Now(),
		)
		if err != nil {
			// 加入失败不影响班级创建，仅记录日志
			// TODO: 添加日志记录
		}
	}

	// 步骤5: 如提供了curriculum_config，创建教材配置
	if req.CurriculumConfig != nil {
		textbookVersionsJSON, _ := json.Marshal(req.CurriculumConfig.TextbookVersions)
		subjectsJSON, _ := json.Marshal(req.CurriculumConfig.Subjects)
		customTextbooksJSON, _ := json.Marshal(req.CurriculumConfig.CustomTextbooks)

		config := &database.TeacherCurriculumConfig{
			TeacherID:        userIDInt64,
			PersonaID:        personaID,
			GradeLevel:       req.CurriculumConfig.GradeLevel,
			Grade:            req.CurriculumConfig.Grade,
			TextbookVersions: string(textbookVersionsJSON),
			Subjects:         string(subjectsJSON),
			CurrentProgress:  req.CurriculumConfig.CurrentProgress,
		}
		// 使用custom_textbooks字段存储自定义教材
		// 注意：由于数据库模型中没有custom_textbooks字段，我们将其存储在region字段中
		// 这是一种临时方案，后续如果数据库表结构更新可以调整
		config.Region = string(customTextbooksJSON)

		_, curriculumErr := curriculumRepo.CreateWithTx(tx, config)
		if curriculumErr != nil {
			// 教材配置创建失败不影响班级创建，仅记录日志
			// TODO: 添加日志记录
			_ = curriculumErr
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "提交事务失败: "+err.Error())
		return
	}

	// 查询创建后的班级信息
	created, err := classRepo.GetByID(classID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询班级信息失败: "+err.Error())
		return
	}

	// 查询教材配置（如有）
	var curriculumConfigData gin.H
	if req.CurriculumConfig != nil {
		curriculumConfig, _ := curriculumRepo.GetActiveByPersonaID(personaID)
		if curriculumConfig != nil {
			var textbookVersions []string
			var subjects []string
			var customTextbooks []string
			_ = json.Unmarshal([]byte(curriculumConfig.TextbookVersions), &textbookVersions)
			_ = json.Unmarshal([]byte(curriculumConfig.Subjects), &subjects)
			_ = json.Unmarshal([]byte(curriculumConfig.Region), &customTextbooks)

			curriculumConfigData = gin.H{
				"id":                curriculumConfig.ID,
				"grade_level":       curriculumConfig.GradeLevel,
				"grade":             curriculumConfig.Grade,
				"subjects":          subjects,
				"textbook_versions": textbookVersions,
				"custom_textbooks":  customTextbooks,
				"current_progress":  curriculumConfig.CurrentProgress,
			}
		}
	}

	// 生成包含新分身的token
	jwtManager := GetJWTManager(h.manager)
	if jwtManager != nil {
		userRole, _ := c.Get("user_role")
		userRoleStr, _ := userRole.(string)
		username, _ := c.Get("username")
		usernameStr, _ := username.(string)
		newToken, _, err := jwtManager.GenerateTokenWithUserRole(userIDInt64, usernameStr, "teacher", userRoleStr, personaID)
		if err == nil {
			resp := gin.H{
				"id":                  created.ID,
				"name":                created.Name,
				"description":         created.Description,
				"is_public":           created.IsPublic == 1,
				"persona_id":          personaID,
				"persona_nickname":    personaNickname,
				"persona_school":      personaSchool,
				"persona_description": personaDescription,
				"teacher_id":          userIDInt64,
				"created_at":          created.CreatedAt,
				"token":               newToken,
			}
			if curriculumConfigData != nil {
				resp["curriculum_config"] = curriculumConfigData
			}
			Success(c, resp)
			return
		}
	}

	// token生成失败或JWT管理器不可用，返回不含token的响应
	resp := gin.H{
		"id":                  created.ID,
		"name":                created.Name,
		"description":         created.Description,
		"is_public":           created.IsPublic == 1,
		"persona_id":          personaID,
		"persona_nickname":    personaNickname,
		"persona_school":      personaSchool,
		"persona_description": personaDescription,
		"teacher_id":          userIDInt64,
		"created_at":          created.CreatedAt,
	}
	if curriculumConfigData != nil {
		resp["curriculum_config"] = curriculumConfigData
	}
	Success(c, resp)
}

// getTestStudentPersonaID 获取教师的自测学生分身ID
// V2.0 迭代11 M4：实现自测学生自动加入班级
func (h *Handler) getTestStudentPersonaID(teacherID int64, db *sql.DB) (int64, error) {
	userRepo := database.NewUserRepository(db)
	personaRepo := database.NewPersonaRepository(db)

	// 查询自测学生
	testStudent, err := userRepo.FindByTestTeacherID(teacherID)
	if err != nil {
		return 0, err
	}
	if testStudent == nil {
		return 0, nil
	}

	// 获取自测学生的学生分身
	testPersona, err := personaRepo.GetStudentPersonaByUserID(testStudent.ID)
	if err != nil {
		return 0, err
	}
	if testPersona == nil {
		return 0, nil
	}

	return testPersona.ID, nil
}

// HandleGetClasses 获取班级列表
// GET /api/classes
func (h *Handler) HandleGetClasses(c *gin.Context) {
	personaID, _ := c.Get("persona_id")
	personaIDInt64, ok := personaID.(int64)
	if !ok || personaIDInt64 <= 0 {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	personaRepo := database.NewPersonaRepository(db)
	classRepo := database.NewClassRepository(db)

	// 校验当前分身是教师分身
	persona, err := personaRepo.GetByID(personaIDInt64)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询分身失败: "+err.Error())
		return
	}
	if persona == nil {
		Error(c, http.StatusNotFound, 40013, "分身不存在")
		return
	}
	if persona.Role != "teacher" {
		Error(c, http.StatusForbidden, 40018, "只有教师分身才能查看班级列表")
		return
	}

	// 获取班级列表（含成员数）
	classes, err := classRepo.ListByPersonaID(personaIDInt64)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询班级列表失败: "+err.Error())
		return
	}

	Success(c, classes)
}

// HandleUpdateClass 更新班级信息
// PUT /api/classes/:id
// V2.0 迭代13扩展：支持更新教材配置
func (h *Handler) HandleUpdateClass(c *gin.Context) {
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

	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok || userIDInt64 <= 0 {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	var req struct {
		Name             string                 `json:"name"`
		Description      string                 `json:"description"`
		IsPublic         *bool                  `json:"is_public"`
		CurriculumConfig *CurriculumConfigValue `json:"curriculum_config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	// 验证学段枚举值（如有提供教材配置）
	if req.CurriculumConfig != nil && req.CurriculumConfig.GradeLevel != "" {
		if !validGradeLevels[req.CurriculumConfig.GradeLevel] {
			Error(c, http.StatusBadRequest, 40041, "无效的学段类型: "+req.CurriculumConfig.GradeLevel)
			return
		}
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	classRepo := database.NewClassRepository(db)
	curriculumRepo := database.NewCurriculumConfigRepository(db)

	// 校验班级存在且属于当前教师分身
	class, err := classRepo.GetByID(classID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询班级失败: "+err.Error())
		return
	}
	if class == nil {
		Error(c, http.StatusNotFound, 40017, "班级不存在")
		return
	}
	if class.PersonaID != personaIDInt64 {
		Error(c, http.StatusForbidden, 40018, "无权操作此班级")
		return
	}

	// 使用原值填充未提供的字段
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = class.Name
	}
	description := strings.TrimSpace(req.Description)
	if description == "" {
		description = class.Description
	}

	// 校验同一教师分身下班级名不重复（排除自身）
	exists, err := classRepo.CheckNameExists(personaIDInt64, name, classID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "检查班级名称失败: "+err.Error())
		return
	}
	if exists {
		Error(c, http.StatusConflict, 40016, "班级名称已存在")
		return
	}

	// 开启事务
	tx, err := db.Begin()
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "开启事务失败: "+err.Error())
		return
	}
	defer tx.Rollback()

	// 更新班级信息
	isPublic := class.IsPublic
	isPublicChanged := false
	if req.IsPublic != nil {
		rawIsPublic := 0
		if *req.IsPublic {
			rawIsPublic = 1
		}
		if rawIsPublic != class.IsPublic {
			isPublic = rawIsPublic
			isPublicChanged = true
		}
	}

	_, err = tx.Exec(`UPDATE classes SET name = ?, description = ?, is_public = ?, updated_at = ? WHERE id = ?`,
		name, description, isPublic, time.Now(), classID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "更新班级失败: "+err.Error())
		return
	}

	// 如果公开状态变更，同步更新分身状态
	if isPublicChanged {
		_, err = tx.Exec(`UPDATE personas SET is_public = ? WHERE id = ?`, isPublic, class.PersonaID)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "更新分身公开状态失败: "+err.Error())
			return
		}
	}

	// 如果提供了教材配置，更新或创建教材配置
	if req.CurriculumConfig != nil {
		// 序列化JSON字段
		textbookVersionsJSON, _ := json.Marshal(req.CurriculumConfig.TextbookVersions)
		subjectsJSON, _ := json.Marshal(req.CurriculumConfig.Subjects)
		customTextbooksJSON, _ := json.Marshal(req.CurriculumConfig.CustomTextbooks)

		config := &database.TeacherCurriculumConfig{
			TeacherID:        userIDInt64,
			PersonaID:        class.PersonaID,
			GradeLevel:       req.CurriculumConfig.GradeLevel,
			Grade:            req.CurriculumConfig.Grade,
			TextbookVersions: string(textbookVersionsJSON),
			Subjects:         string(subjectsJSON),
			CurrentProgress:  req.CurriculumConfig.CurrentProgress,
			Region:           string(customTextbooksJSON),
		}

		// 使用UpsertByPersonaID来更新或创建配置
		_, curriculumErr := curriculumRepo.UpsertByPersonaID(tx, config)
		if curriculumErr != nil {
			// 教材配置更新失败不影响班级更新，仅记录日志
			// TODO: 添加日志记录
			_ = curriculumErr
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "提交事务失败: "+err.Error())
		return
	}

	// 如果更新了教材配置，查询最新的配置信息并返回
	var configResponse gin.H
	if req.CurriculumConfig != nil {
		curriculumConfig, _ := curriculumRepo.GetActiveByPersonaID(class.PersonaID)
		if curriculumConfig != nil {
			configResponse = curriculumConfigToResponse(curriculumConfig)
		}
	}

	resp := gin.H{
		"id":          classID,
		"name":        name,
		"description": description,
		"is_public":   isPublic == 1,
		"persona_id":  personaIDInt64,
	}
	if configResponse != nil {
		resp["curriculum_config"] = configResponse
	}
	Success(c, resp)
}

// HandleDeleteClass 删除班级
// DELETE /api/classes/:id
func (h *Handler) HandleDeleteClass(c *gin.Context) {
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

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	classRepo := database.NewClassRepository(db)

	// 校验班级存在且属于当前教师分身
	class, err := classRepo.GetByID(classID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询班级失败: "+err.Error())
		return
	}
	if class == nil {
		Error(c, http.StatusNotFound, 40017, "班级不存在")
		return
	}
	if class.PersonaID != personaIDInt64 {
		Error(c, http.StatusForbidden, 40018, "无权操作此班级")
		return
	}

	// 检查班级是否还有成员，有成员时不允许删除
	memberCount, err := classRepo.GetMemberCount(classID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询班级成员数失败: "+err.Error())
		return
	}
	if memberCount > 0 {
		Error(c, http.StatusBadRequest, 40024, "班级有成员，无法删除，请先移除所有成员")
		return
	}

	// 删除班级
	if err := classRepo.Delete(classID); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "删除班级失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"message": "班级已删除",
	})
}

// HandleGetClassMembers 获取班级成员列表
// GET /api/classes/:id/members
func (h *Handler) HandleGetClassMembers(c *gin.Context) {
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

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	classRepo := database.NewClassRepository(db)

	// 校验班级存在且属于当前教师分身
	class, err := classRepo.GetByID(classID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询班级失败: "+err.Error())
		return
	}
	if class == nil {
		Error(c, http.StatusNotFound, 40017, "班级不存在")
		return
	}
	if class.PersonaID != personaIDInt64 {
		Error(c, http.StatusForbidden, 40018, "无权操作此班级")
		return
	}

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// 获取成员列表
	members, total, err := classRepo.ListMembers(classID, offset, pageSize)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询班级成员失败: "+err.Error())
		return
	}

	SuccessPage(c, members, total, page, pageSize)
}

// HandleAddClassMember 添加班级成员
// POST /api/classes/:id/members
func (h *Handler) HandleAddClassMember(c *gin.Context) {
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

	var req struct {
		StudentPersonaID int64 `json:"student_persona_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	personaRepo := database.NewPersonaRepository(db)
	classRepo := database.NewClassRepository(db)

	// 校验班级存在且属于当前教师分身
	class, err := classRepo.GetByID(classID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询班级失败: "+err.Error())
		return
	}
	if class == nil {
		Error(c, http.StatusNotFound, 40017, "班级不存在")
		return
	}
	if class.PersonaID != personaIDInt64 {
		Error(c, http.StatusForbidden, 40018, "无权操作此班级")
		return
	}

	// 校验 student_persona_id 是有效的学生分身
	studentPersona, err := personaRepo.GetByID(req.StudentPersonaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询学生分身失败: "+err.Error())
		return
	}
	if studentPersona == nil {
		Error(c, http.StatusNotFound, 40013, "学生分身不存在")
		return
	}
	if studentPersona.Role != "student" {
		Error(c, http.StatusBadRequest, 40004, "指定的分身不是学生角色")
		return
	}

	// 校验该学生未在班级中
	isMember, err := classRepo.IsMember(classID, req.StudentPersonaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "检查班级成员失败: "+err.Error())
		return
	}
	if isMember {
		Error(c, http.StatusConflict, 40019, "学生已在班级中")
		return
	}

	// 添加成员
	memberID, err := classRepo.AddMember(classID, req.StudentPersonaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "添加班级成员失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"member_id":          memberID,
		"class_id":           classID,
		"student_persona_id": req.StudentPersonaID,
		"message":            "成员添加成功",
	})
}

// HandleRemoveClassMember 移除班级成员
// DELETE /api/classes/:id/members/:member_id
func (h *Handler) HandleRemoveClassMember(c *gin.Context) {
	classIDStr := c.Param("id")
	classID, err := strconv.ParseInt(classIDStr, 10, 64)
	if err != nil || classID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的班级ID")
		return
	}

	memberIDStr := c.Param("member_id")
	memberID, err := strconv.ParseInt(memberIDStr, 10, 64)
	if err != nil || memberID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的成员ID")
		return
	}

	personaID, _ := c.Get("persona_id")
	personaIDInt64, ok := personaID.(int64)
	if !ok || personaIDInt64 <= 0 {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	classRepo := database.NewClassRepository(db)

	// 校验班级存在且属于当前教师分身
	class, err := classRepo.GetByID(classID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询班级失败: "+err.Error())
		return
	}
	if class == nil {
		Error(c, http.StatusNotFound, 40017, "班级不存在")
		return
	}
	if class.PersonaID != personaIDInt64 {
		Error(c, http.StatusForbidden, 40018, "无权操作此班级")
		return
	}

	// 校验成员记录存在且属于该班级
	member, err := classRepo.GetMemberByID(memberID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询成员记录失败: "+err.Error())
		return
	}
	if member == nil {
		Error(c, http.StatusNotFound, 40020, "成员不存在")
		return
	}
	if member.ClassID != classID {
		Error(c, http.StatusBadRequest, 40020, "成员不属于该班级")
		return
	}

	// 移除成员
	if err := classRepo.RemoveMember(memberID); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "移除班级成员失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"message": "成员已移除",
	})
}

// ======================== V2.0 迭代3 新增方法 ========================

// HandleToggleClass 启用/停用班级
// PUT /api/classes/:id/toggle
func (h *Handler) HandleToggleClass(c *gin.Context) {
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

	var req struct {
		IsActive *bool `json:"is_active" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	classRepo := database.NewClassRepository(db)

	// 校验班级存在且属于当前教师分身
	class, err := classRepo.GetByID(classID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询班级失败: "+err.Error())
		return
	}
	if class == nil {
		Error(c, http.StatusNotFound, 40016, "班级不存在")
		return
	}
	if class.PersonaID != personaIDInt64 {
		Error(c, http.StatusForbidden, 40017, "班级不属于当前教师分身")
		return
	}

	// 执行启停
	isActiveInt := 0
	if *req.IsActive {
		isActiveInt = 1
	}

	result, err := classRepo.ToggleClass(classID, isActiveInt)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "更新班级状态失败: "+err.Error())
		return
	}

	Success(c, result)
}

// ======================== V2.0 迭代13 新增方法 ========================

// HandleGetClass 获取班级详情
// GET /api/classes/:id
// V2.0 迭代13新增：支持返回教材配置
func (h *Handler) HandleGetClass(c *gin.Context) {
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

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	classRepo := database.NewClassRepository(db)
	curriculumRepo := database.NewCurriculumConfigRepository(db)

	// 查询班级信息
	class, err := classRepo.GetByID(classID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询班级失败: "+err.Error())
		return
	}
	if class == nil {
		Error(c, http.StatusNotFound, 40017, "班级不存在")
		return
	}

	// 校验权限：班级属于当前教师分身
	if class.PersonaID != personaIDInt64 {
		Error(c, http.StatusForbidden, 40018, "无权操作此班级")
		return
	}

	// 查询关联的教材配置
	curriculumConfig, err := curriculumRepo.GetActiveByPersonaID(class.PersonaID)
	if err != nil {
		// 查询配置失败不影响返回班级信息
		curriculumConfig = nil
	}

	// 组装教材配置响应
	var configResponse gin.H
	if curriculumConfig != nil {
		var subjects, textbookVersions, customTextbooks []string
		_ = json.Unmarshal([]byte(curriculumConfig.Subjects), &subjects)
		_ = json.Unmarshal([]byte(curriculumConfig.TextbookVersions), &textbookVersions)
		_ = json.Unmarshal([]byte(curriculumConfig.Region), &customTextbooks)

		configResponse = gin.H{
			"id":                curriculumConfig.ID,
			"grade_level":       curriculumConfig.GradeLevel,
			"grade":             curriculumConfig.Grade,
			"subjects":          subjects,
			"textbook_versions": textbookVersions,
			"custom_textbooks":  customTextbooks,
			"current_progress":  curriculumConfig.CurrentProgress,
		}
	}

	Success(c, gin.H{
		"id":                class.ID,
		"name":              class.Name,
		"description":       class.Description,
		"is_public":         class.IsPublic == 1,
		"is_active":         class.IsActive == 1,
		"persona_id":        class.PersonaID,
		"created_at":        class.CreatedAt,
		"updated_at":        class.UpdatedAt,
		"curriculum_config": configResponse,
	})
}
