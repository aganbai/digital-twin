package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"digital-twin/src/backend/database"
	"digital-twin/src/harness/core"
	"digital-twin/src/harness/manager"
	"digital-twin/src/plugins/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler API 请求处理器
type Handler struct {
	manager *manager.HarnessManager
}

// NewHandler 创建请求处理器
func NewHandler(mgr *manager.HarnessManager) *Handler {
	return &Handler{manager: mgr}
}

// ======================== 认证接口 ========================

// HandleRegister 用户注册
// POST /api/auth/register
func (h *Handler) HandleRegister(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role"`
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	// 直接调用 auth 插件
	plugin, err := h.manager.GetPlugin("authentication")
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "认证服务不可用")
		return
	}

	input := &core.PluginInput{
		RequestID: uuid.New().String(),
		Data: map[string]interface{}{
			"action":   "register",
			"username": req.Username,
			"password": req.Password,
			"role":     req.Role,
			"nickname": req.Nickname,
			"email":    req.Email,
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(c.Request.Context(), input)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "注册失败: "+err.Error())
		return
	}

	if !output.Success {
		// 根据错误码判断 HTTP 状态码
		httpStatus := http.StatusBadRequest
		errorCode := 40004
		if code, ok := output.Data["error_code"]; ok {
			errorCode = toInt(code, 40004)
			if errorCode == 40006 {
				httpStatus = http.StatusConflict
			}
		}
		Error(c, httpStatus, errorCode, output.Error)
		return
	}

	Success(c, gin.H{
		"user_id":    output.Data["user_id"],
		"token":      output.Data["token"],
		"role":       output.Data["role"],
		"nickname":   output.Data["nickname"],
		"expires_at": output.Data["expires_at"],
	})
}

// HandleLogin 用户登录
// POST /api/auth/login
func (h *Handler) HandleLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	plugin, err := h.manager.GetPlugin("authentication")
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "认证服务不可用")
		return
	}

	input := &core.PluginInput{
		RequestID: uuid.New().String(),
		Data: map[string]interface{}{
			"action":   "login",
			"username": req.Username,
			"password": req.Password,
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(c.Request.Context(), input)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "登录失败: "+err.Error())
		return
	}

	if !output.Success {
		httpStatus := http.StatusUnauthorized
		errorCode := 40001
		if code, ok := output.Data["error_code"]; ok {
			errorCode = toInt(code, 40001)
		}
		Error(c, httpStatus, errorCode, output.Error)
		return
	}

	Success(c, gin.H{
		"user_id":    output.Data["user_id"],
		"token":      output.Data["token"],
		"role":       output.Data["role"],
		"nickname":   output.Data["nickname"],
		"expires_at": output.Data["expires_at"],
	})
}

// HandleWxLogin 微信登录
// POST /api/auth/wx-login
func (h *Handler) HandleWxLogin(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "缺少 code 参数")
		return
	}

	plugin, err := h.manager.GetPlugin("authentication")
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "认证服务不可用")
		return
	}

	input := &core.PluginInput{
		RequestID: uuid.New().String(),
		Data: map[string]interface{}{
			"action": "wx-login",
			"code":   req.Code,
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(c.Request.Context(), input)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "微信登录失败: "+err.Error())
		return
	}

	if !output.Success {
		httpStatus := http.StatusBadRequest
		errorCode := 40004
		if code, ok := output.Data["error_code"]; ok {
			errorCode = toInt(code, 40004)
			if errorCode == 50001 {
				httpStatus = http.StatusInternalServerError
			}
		}
		Error(c, httpStatus, errorCode, output.Error)
		return
	}

	Success(c, gin.H{
		"user_id":         output.Data["user_id"],
		"token":           output.Data["token"],
		"role":            output.Data["role"],
		"nickname":        output.Data["nickname"],
		"is_new_user":     output.Data["is_new_user"],
		"expires_at":      output.Data["expires_at"],
		"personas":        output.Data["personas"],        // V2.0 迭代2
		"current_persona": output.Data["current_persona"], // V2.0 迭代2
	})
}

// HandleCompleteProfile 新用户补全信息
// POST /api/auth/complete-profile
func (h *Handler) HandleCompleteProfile(c *gin.Context) {
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

	// 从 JWT 中间件获取用户信息
	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	plugin, err := h.manager.GetPlugin("authentication")
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "认证服务不可用")
		return
	}

	input := &core.PluginInput{
		RequestID: uuid.New().String(),
		Data: map[string]interface{}{
			"action":      "complete-profile",
			"user_id":     userIDInt64,
			"role":        req.Role,
			"nickname":    req.Nickname,
			"school":      req.School,
			"description": req.Description,
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(c.Request.Context(), input)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "补全信息失败: "+err.Error())
		return
	}

	if !output.Success {
		httpStatus := http.StatusBadRequest
		errorCode := 40004
		if code, ok := output.Data["error_code"]; ok {
			errorCode = toInt(code, 40004)
			if errorCode == 40005 {
				httpStatus = http.StatusNotFound
			} else if errorCode == 40008 || errorCode == 40015 {
				httpStatus = http.StatusConflict
			}
		}
		Error(c, httpStatus, errorCode, output.Error)
		return
	}

	Success(c, gin.H{
		"user_id":     output.Data["user_id"],
		"role":        output.Data["role"],
		"nickname":    output.Data["nickname"],
		"school":      output.Data["school"],
		"description": output.Data["description"],
		"persona_id":  output.Data["persona_id"], // V2.0 迭代2
		"token":       output.Data["token"],
		"expires_at":  output.Data["expires_at"],
	})
}

// HandleRefresh 刷新 token
// POST /api/auth/refresh
func (h *Handler) HandleRefresh(c *gin.Context) {
	// 从 Authorization 头获取 token
	token := extractToken(c)
	if token == "" {
		Error(c, http.StatusUnauthorized, 40001, "缺少认证令牌")
		return
	}

	plugin, err := h.manager.GetPlugin("authentication")
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "认证服务不可用")
		return
	}

	input := &core.PluginInput{
		RequestID: uuid.New().String(),
		Data: map[string]interface{}{
			"action": "refresh",
			"token":  token,
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(c.Request.Context(), input)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "刷新令牌失败: "+err.Error())
		return
	}

	if !output.Success {
		errorCode := 40001
		if code, ok := output.Data["error_code"]; ok {
			errorCode = toInt(code, 40001)
		}
		Error(c, http.StatusUnauthorized, errorCode, output.Error)
		return
	}

	Success(c, gin.H{
		"user_id":    output.Data["user_id"],
		"token":      output.Data["token"],
		"role":       output.Data["role"],
		"expires_at": output.Data["expires_at"],
	})
}

// ======================== 教师列表接口 ========================

// HandleGetTeachers 获取教师列表
// GET /api/teachers
// V2.0 迭代3 改造：学生角色仅返回已授权+启用的教师分身
func (h *Handler) HandleGetTeachers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	// V2.0 迭代3：判断当前用户角色
	role, _ := c.Get("role")
	roleStr, _ := role.(string)
	personaID, _ := c.Get("persona_id")
	personaIDInt64, _ := personaID.(int64)

	if roleStr == "student" && personaIDInt64 > 0 {
		// 学生角色：仅返回已授权+启用的教师分身
		userRepo := database.NewUserRepository(db)
		teachers, total, err := userRepo.ListTeachersForStudent(personaIDInt64, offset, pageSize)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "查询教师列表失败: "+err.Error())
			return
		}
		SuccessPage(c, teachers, total, page, pageSize)
		return
	}

	// 教师角色或其他：返回所有教师分身列表
	userRepo := database.NewUserRepository(db)
	teachers, total, err := userRepo.GetTeachers(offset, pageSize)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询教师列表失败: "+err.Error())
		return
	}

	SuccessPage(c, teachers, total, page, pageSize)
}

// ======================== 用户信息接口 ========================

// HandleGetUserProfile 获取当前用户信息
// GET /api/user/profile
// V2.0 迭代2 改造：返回分身信息
func (h *Handler) HandleGetUserProfile(c *gin.Context) {
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

	userRepo := database.NewUserRepository(db)
	personaRepo := database.NewPersonaRepository(db)

	// 查询用户基本信息
	user, err := userRepo.GetByID(userIDInt64)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询用户信息失败: "+err.Error())
		return
	}
	if user == nil {
		Error(c, http.StatusNotFound, 40005, "用户不存在")
		return
	}

	// V2.0 迭代2：查询分身列表
	personas, err := personaRepo.ListByUserID(userIDInt64, "")
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询分身列表失败: "+err.Error())
		return
	}

	// V2.0 迭代2：确定当前分身
	var currentPersona interface{}
	personaID, _ := c.Get("persona_id")
	personaIDInt64, _ := personaID.(int64)
	if personaIDInt64 > 0 {
		for _, p := range personas {
			if p.ID == personaIDInt64 {
				currentPersona = p
				break
			}
		}
	}

	// 确定角色和昵称
	role := ""
	nickname := user.Username
	if personaIDInt64 > 0 {
		for _, p := range personas {
			if p.ID == personaIDInt64 {
				role = p.Role
				nickname = p.Nickname
				break
			}
		}
	}
	if role == "" && len(personas) > 0 {
		role = personas[0].Role
		nickname = personas[0].Nickname
	}

	Success(c, gin.H{
		"id":              user.ID,
		"user_id":         user.ID,
		"nickname":        nickname,
		"role":            role,
		"username":        user.Username,
		"current_persona": currentPersona,
		"personas":        personas,
		"created_at":      user.CreatedAt,
	})
}

// ======================== 记忆接口 ========================

// HandleGetMemories 获取记忆列表
// GET /api/memories
func (h *Handler) HandleGetMemories(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	teacherIDStr := c.Query("teacher_id")
	teacherID, err := strconv.ParseInt(teacherIDStr, 10, 64)
	if err != nil || teacherID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "缺少或无效的 teacher_id 参数")
		return
	}

	memoryType := c.Query("memory_type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	plugin, err := h.manager.GetPlugin("memory-management")
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "记忆服务不可用")
		return
	}

	input := &core.PluginInput{
		RequestID: uuid.New().String(),
		Data: map[string]interface{}{
			"action":      "list",
			"student_id":  userIDInt64,
			"teacher_id":  teacherID,
			"memory_type": memoryType,
			"page":        page,
			"page_size":   pageSize,
		},
		Context: c.Request.Context(),
	}

	output, err := plugin.Execute(c.Request.Context(), input)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询记忆失败: "+err.Error())
		return
	}

	if !output.Success {
		errorCode := 50001
		if code, ok := output.Data["error_code"]; ok {
			errorCode = toInt(code, 50001)
		}
		Error(c, http.StatusInternalServerError, errorCode, output.Error)
		return
	}

	total := toInt(output.Data["total"], 0)
	SuccessPage(c, output.Data["memories"], total, page, pageSize)
}

// ======================== 系统接口 ========================

// HandleHealthCheck 健康检查
// GET /api/system/health
func (h *Handler) HandleHealthCheck(c *gin.Context) {
	health := h.manager.HealthCheck()
	Success(c, health)
}

// HandleGetPlugins 获取插件列表
// GET /api/system/plugins
func (h *Handler) HandleGetPlugins(c *gin.Context) {
	plugins := h.manager.ListPlugins()
	Success(c, plugins)
}

// HandleGetPipelines 获取管道列表
// GET /api/system/pipelines
func (h *Handler) HandleGetPipelines(c *gin.Context) {
	pipelines := h.manager.ListPipelines()
	Success(c, pipelines)
}

// ======================== 辅助函数 ========================

// extractToken 从 Authorization 头提取 Bearer token
func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}

// GetJWTManager 从 auth 插件获取 JWTManager
func GetJWTManager(mgr *manager.HarnessManager) *auth.JWTManager {
	plugin, err := mgr.GetPlugin("authentication")
	if err != nil {
		return nil
	}

	// 类型断言获取 AuthPlugin
	authPlugin, ok := plugin.(*auth.AuthPlugin)
	if !ok {
		return nil
	}

	return authPlugin.GetJWTManager()
}

// toInt 将 interface{} 转换为 int
func toInt(v interface{}, defaultVal int) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	default:
		return defaultVal
	}
}

// toInt64 将 interface{} 转换为 int64（暂不使用，预留）
func toInt64(v interface{}, defaultVal int64) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case float64:
		return int64(val)
	default:
		return defaultVal
	}
}
