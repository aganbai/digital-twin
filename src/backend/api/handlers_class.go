package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"digital-twin/src/backend/database"

	"github.com/gin-gonic/gin"
)

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
		Name               string `json:"name" binding:"required"`
		Description        string `json:"description"`
		PersonaNickname    string `json:"persona_nickname" binding:"required"`
		PersonaSchool      string `json:"persona_school" binding:"required"`
		PersonaDescription string `json:"persona_description" binding:"required"`
		IsPublic           *bool  `json:"is_public"`
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

	// 步骤1: 先创建班级专属分身（需要在创建班级之前，因为班级需要persona_id）
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

	// 生成包含新分身的token
	jwtManager := GetJWTManager(h.manager)
	if jwtManager != nil {
		userRole, _ := c.Get("user_role")
		userRoleStr, _ := userRole.(string)
		username, _ := c.Get("username")
		usernameStr, _ := username.(string)
		newToken, _, err := jwtManager.GenerateTokenWithUserRole(userIDInt64, usernameStr, "teacher", userRoleStr, personaID)
		if err == nil {
			Success(c, gin.H{
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
			})
			return
		}
	}

	// token生成失败或JWT管理器不可用，返回不含token的响应
	Success(c, gin.H{
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
	})
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

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
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

	// 更新班级信息
	if err := classRepo.Update(classID, name, description); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "更新班级失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"id":          classID,
		"name":        name,
		"description": description,
		"persona_id":  personaIDInt64,
	})
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
