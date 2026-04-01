package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"digital-twin/src/backend/database"
	"digital-twin/src/plugins/auth"

	"github.com/gin-gonic/gin"
)

// ======================== 分身管理接口 (V2.0 迭代2) ========================

// HandleCreatePersona 创建分身
// POST /api/personas
func (h *Handler) HandleCreatePersona(c *gin.Context) {
	var req struct {
		Role        string `json:"role" binding:"required"`
		Nickname    string `json:"nickname" binding:"required"`
		School      string `json:"school"`
		Description string `json:"description"`
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

	// 校验 role
	if req.Role != "teacher" && req.Role != "student" {
		Error(c, http.StatusBadRequest, 40004, "角色只能是 teacher 或 student")
		return
	}

	// 校验 nickname
	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" {
		Error(c, http.StatusBadRequest, 40004, "昵称不能为空")
		return
	}
	nicknameRunes := []rune(nickname)
	if len(nicknameRunes) > 20 {
		Error(c, http.StatusBadRequest, 40004, "昵称长度不能超过20个字符")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	personaRepo := database.NewPersonaRepository(db)
	userRepo := database.NewUserRepository(db)

	school := strings.TrimSpace(req.School)
	description := strings.TrimSpace(req.Description)

	// 教师分身额外校验
	if req.Role == "teacher" {
		if school == "" {
			Error(c, http.StatusBadRequest, 40004, "教师角色必须填写学校名称")
			return
		}
		if description == "" {
			Error(c, http.StatusBadRequest, 40004, "教师角色必须填写分身描述")
			return
		}
		// 检查同名+同校唯一性
		exists, err := personaRepo.CheckTeacherPersonaExists(nickname, school, 0)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "检查教师分身唯一性失败: "+err.Error())
			return
		}
		if exists {
			Error(c, http.StatusConflict, 40015, "该学校已有同名教师分身，请修改名称")
			return
		}
	}

	// 创建分身
	persona := &database.Persona{
		UserID:      userIDInt64,
		Role:        req.Role,
		Nickname:    nickname,
		School:      school,
		Description: description,
	}
	personaID, err := personaRepo.Create(persona)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "创建分身失败: "+err.Error())
		return
	}

	// 如果是第一个分身，自动设为 default_persona_id
	count, _ := personaRepo.CountByUserID(userIDInt64)
	if count <= 1 {
		userRepo.UpdateDefaultPersonaID(userIDInt64, personaID)
	}

	// 获取 JWT 管理器，生成新 token
	jwtManager := GetJWTManager(h.manager)
	if jwtManager == nil {
		Error(c, http.StatusInternalServerError, 50001, "认证服务不可用")
		return
	}

	username, _ := c.Get("username")
	usernameStr, _ := username.(string)
	token, expiresAt, err := jwtManager.GenerateToken(userIDInt64, usernameStr, req.Role, personaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "生成 token 失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"persona_id":  personaID,
		"role":        req.Role,
		"nickname":    nickname,
		"school":      school,
		"description": description,
		"token":       token,
		"expires_at":  expiresAt.Format(time.RFC3339),
	})
}

// HandleGetPersonas 获取分身列表
// GET /api/personas
func (h *Handler) HandleGetPersonas(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	roleFilter := c.Query("role")

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	personaRepo := database.NewPersonaRepository(db)
	userRepo := database.NewUserRepository(db)

	personas, err := personaRepo.ListByUserID(userIDInt64, roleFilter)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询分身列表失败: "+err.Error())
		return
	}

	// 获取当前默认分身ID
	user, err := userRepo.GetByID(userIDInt64)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询用户信息失败: "+err.Error())
		return
	}

	currentPersonaID := int64(0)
	if user != nil {
		currentPersonaID = user.DefaultPersonaID
	}

	Success(c, gin.H{
		"personas":           personas,
		"current_persona_id": currentPersonaID,
	})
}

// HandleEditPersona 编辑分身
// PUT /api/personas/:id
func (h *Handler) HandleEditPersona(c *gin.Context) {
	personaIDStr := c.Param("id")
	personaID, err := strconv.ParseInt(personaIDStr, 10, 64)
	if err != nil || personaID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的分身ID")
		return
	}

	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	var req struct {
		Nickname    string `json:"nickname"`
		School      string `json:"school"`
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

	personaRepo := database.NewPersonaRepository(db)

	// 校验分身存在且属于当前用户
	persona, err := personaRepo.GetByID(personaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询分身失败: "+err.Error())
		return
	}
	if persona == nil {
		Error(c, http.StatusNotFound, 40005, "分身不存在")
		return
	}
	if persona.UserID != userIDInt64 {
		Error(c, http.StatusForbidden, 40003, "无权操作此分身")
		return
	}

	// 使用原值填充未提供的字段
	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" {
		nickname = persona.Nickname
	}
	school := strings.TrimSpace(req.School)
	if school == "" {
		school = persona.School
	}
	description := strings.TrimSpace(req.Description)
	if description == "" {
		description = persona.Description
	}

	// 教师分身：校验 nickname + school 唯一（排除自身）
	if persona.Role == "teacher" {
		exists, err := personaRepo.CheckTeacherPersonaExists(nickname, school, personaID)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "检查教师分身唯一性失败: "+err.Error())
			return
		}
		if exists {
			Error(c, http.StatusConflict, 40015, "该学校已有同名教师分身，请修改名称")
			return
		}
	}

	// 更新分身
	if err := personaRepo.Update(personaID, nickname, school, description); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "更新分身失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"persona_id":  personaID,
		"role":        persona.Role,
		"nickname":    nickname,
		"school":      school,
		"description": description,
	})
}

// HandleActivatePersona 启用分身
// PUT /api/personas/:id/activate
func (h *Handler) HandleActivatePersona(c *gin.Context) {
	h.setPersonaActive(c, 1)
}

// HandleDeactivatePersona 停用分身
// PUT /api/personas/:id/deactivate
func (h *Handler) HandleDeactivatePersona(c *gin.Context) {
	h.setPersonaActive(c, 0)
}

// setPersonaActive 设置分身激活状态（内部公用方法）
func (h *Handler) setPersonaActive(c *gin.Context, isActive int) {
	personaIDStr := c.Param("id")
	personaID, err := strconv.ParseInt(personaIDStr, 10, 64)
	if err != nil || personaID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的分身ID")
		return
	}

	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	personaRepo := database.NewPersonaRepository(db)

	// 校验分身存在且属于当前用户
	persona, err := personaRepo.GetByID(personaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询分身失败: "+err.Error())
		return
	}
	if persona == nil {
		Error(c, http.StatusNotFound, 40005, "分身不存在")
		return
	}
	if persona.UserID != userIDInt64 {
		Error(c, http.StatusForbidden, 40003, "无权操作此分身")
		return
	}

	if err := personaRepo.SetActive(personaID, isActive); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "更新分身状态失败: "+err.Error())
		return
	}

	action := "activated"
	if isActive == 0 {
		action = "deactivated"
	}

	Success(c, gin.H{
		"persona_id": personaID,
		"is_active":  isActive == 1,
		"action":     action,
	})
}

// HandleSwitchPersona 切换当前分身
// PUT /api/personas/:id/switch
func (h *Handler) HandleSwitchPersona(c *gin.Context) {
	personaIDStr := c.Param("id")
	personaID, err := strconv.ParseInt(personaIDStr, 10, 64)
	if err != nil || personaID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的分身ID")
		return
	}

	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	personaRepo := database.NewPersonaRepository(db)
	userRepo := database.NewUserRepository(db)

	// 校验分身存在且属于当前用户且处于启用状态
	persona, err := personaRepo.GetByID(personaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询分身失败: "+err.Error())
		return
	}
	if persona == nil {
		Error(c, http.StatusNotFound, 40005, "分身不存在")
		return
	}
	if persona.UserID != userIDInt64 {
		Error(c, http.StatusForbidden, 40003, "无权操作此分身")
		return
	}
	if persona.IsActive != 1 {
		Error(c, http.StatusBadRequest, 40004, "该分身已停用，无法切换")
		return
	}

	// 更新 users.default_persona_id
	if err := userRepo.UpdateDefaultPersonaID(userIDInt64, personaID); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "切换分身失败: "+err.Error())
		return
	}

	// 获取 JWT 管理器，生成新 token
	jwtManager := GetJWTManager(h.manager)
	if jwtManager == nil {
		Error(c, http.StatusInternalServerError, 50001, "认证服务不可用")
		return
	}

	username, _ := c.Get("username")
	usernameStr, _ := username.(string)
	token, expiresAt, err := jwtManager.GenerateToken(userIDInt64, usernameStr, persona.Role, personaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "生成 token 失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"persona_id":  personaID,
		"role":        persona.Role,
		"nickname":    persona.Nickname,
		"school":      persona.School,
		"description": persona.Description,
		"token":       token,
		"expires_at":  expiresAt.Format(time.RFC3339),
	})
}

// GetJWTManagerFromHandler 从 Handler 获取 JWTManager（内部辅助方法）
func getJWTManagerFromHandler(h *Handler) *auth.JWTManager {
	return GetJWTManager(h.manager)
}

// ======================== V2.0 迭代3 新增方法 ========================

// HandleGetPersonaDashboard 教师分身仪表盘
// GET /api/personas/:id/dashboard
func (h *Handler) HandleGetPersonaDashboard(c *gin.Context) {
	personaIDStr := c.Param("id")
	personaID, err := strconv.ParseInt(personaIDStr, 10, 64)
	if err != nil || personaID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的分身ID")
		return
	}

	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	personaRepo := database.NewPersonaRepository(db)

	// 校验分身存在且属于当前用户
	persona, err := personaRepo.GetByID(personaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询分身失败: "+err.Error())
		return
	}
	if persona == nil {
		Error(c, http.StatusNotFound, 40013, "分身不存在")
		return
	}
	if persona.UserID != userIDInt64 {
		Error(c, http.StatusForbidden, 40014, "分身不属于当前用户")
		return
	}

	// 获取仪表盘聚合数据
	dashboard, err := personaRepo.GetPersonaDashboard(personaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "获取仪表盘数据失败: "+err.Error())
		return
	}

	// 构建响应
	result := gin.H{
		"persona": gin.H{
			"id":          dashboard.Persona.ID,
			"nickname":    dashboard.Persona.Nickname,
			"school":      dashboard.Persona.School,
			"description": dashboard.Persona.Description,
			"is_active":   dashboard.Persona.IsActive == 1,
		},
		"pending_count": dashboard.PendingCount,
		"classes":       dashboard.Classes,
		"latest_share":  dashboard.LatestShare,
		"stats":         dashboard.Stats,
	}

	Success(c, result)
}

// ======================== V2.0 迭代4 新增方法 ========================

// HandleGetMarketplace 获取分身广场列表
// GET /api/personas/marketplace
func (h *Handler) HandleGetMarketplace(c *gin.Context) {
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// 获取当前学生分身ID（如果有）
	personaID, _ := c.Get("persona_id")
	personaIDInt64, _ := personaID.(int64)

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	personaRepo := database.NewPersonaRepository(db)
	items, total, err := personaRepo.ListPublicPersonas(personaIDInt64, keyword, offset, pageSize)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询广场分身列表失败: "+err.Error())
		return
	}

	SuccessPage(c, items, total, page, pageSize)
}

// HandleSetVisibility 设置分身公开/私有
// PUT /api/personas/:id/visibility
func (h *Handler) HandleSetVisibility(c *gin.Context) {
	personaIDStr := c.Param("id")
	personaID, err := strconv.ParseInt(personaIDStr, 10, 64)
	if err != nil || personaID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的分身ID")
		return
	}

	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	var req struct {
		IsPublic bool `json:"is_public"`
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

	// 校验分身存在且属于当前用户
	persona, err := personaRepo.GetByID(personaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询分身失败: "+err.Error())
		return
	}
	if persona == nil {
		Error(c, http.StatusNotFound, 40013, "分身不存在")
		return
	}
	if persona.UserID != userIDInt64 {
		Error(c, http.StatusForbidden, 40014, "分身不属于当前用户")
		return
	}
	if persona.Role != "teacher" {
		Error(c, http.StatusBadRequest, 40032, "仅教师分身可设置公开状态")
		return
	}

	isPublic := 0
	if req.IsPublic {
		isPublic = 1
	}

	if err := personaRepo.UpdateVisibility(personaID, isPublic); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "更新分身公开状态失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"id":         personaID,
		"nickname":   persona.Nickname,
		"is_public":  req.IsPublic,
		"updated_at": time.Now().Format(time.RFC3339),
	})
}

// HandleSearchStudents 搜索已注册学生
// GET /api/students/search
func (h *Handler) HandleSearchStudents(c *gin.Context) {
	keyword := c.Query("keyword")
	if len([]rune(keyword)) < 2 {
		Error(c, http.StatusBadRequest, 40004, "搜索关键词至少需要2个字符")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	personaRepo := database.NewPersonaRepository(db)
	personas, total, err := personaRepo.SearchStudentPersonas(keyword, offset, pageSize)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "搜索学生失败: "+err.Error())
		return
	}

	// 转换为响应格式
	items := make([]gin.H, 0, len(personas))
	for _, p := range personas {
		items = append(items, gin.H{
			"persona_id": p.ID,
			"user_id":    p.UserID,
			"nickname":   p.Nickname,
			"created_at": p.CreatedAt,
		})
	}

	SuccessPage(c, items, total, page, pageSize)
}
