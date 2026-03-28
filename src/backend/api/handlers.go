package api

import (
	"context"
	"fmt"
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
		"user_id":     output.Data["user_id"],
		"token":       output.Data["token"],
		"role":        output.Data["role"],
		"nickname":    output.Data["nickname"],
		"is_new_user": output.Data["is_new_user"],
		"expires_at":  output.Data["expires_at"],
	})
}

// HandleCompleteProfile 新用户补全信息
// POST /api/auth/complete-profile
func (h *Handler) HandleCompleteProfile(c *gin.Context) {
	var req struct {
		Role     string `json:"role" binding:"required"`
		Nickname string `json:"nickname" binding:"required"`
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
			"action":   "complete-profile",
			"user_id":  userIDInt64,
			"role":     req.Role,
			"nickname": req.Nickname,
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
			}
		}
		Error(c, httpStatus, errorCode, output.Error)
		return
	}

	Success(c, gin.H{
		"user_id":  output.Data["user_id"],
		"role":     output.Data["role"],
		"nickname": output.Data["nickname"],
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

// ======================== 对话接口 ========================

// HandleChat 对话接口（通过管道编排执行）
// POST /api/chat
func (h *Handler) HandleChat(c *gin.Context) {
	var req struct {
		Message   string `json:"message" binding:"required"`
		TeacherID int64  `json:"teacher_id" binding:"required"`
		SessionID string `json:"session_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	// 从 JWT 中间件获取用户信息
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	username, _ := c.Get("username")

	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	userContext := &core.UserContext{
		UserID:     fmt.Sprintf("%d", userIDInt64),
		Role:       fmt.Sprintf("%v", role),
		SessionID:  req.SessionID,
		Attributes: map[string]interface{}{"username": username},
	}

	// 构建管道输入
	input := &core.PluginInput{
		RequestID:   uuid.New().String(),
		UserContext: userContext,
		Data: map[string]interface{}{
			"action":     "chat",
			"message":    req.Message,
			"teacher_id": req.TeacherID,
		},
		Context: c.Request.Context(),
	}

	if req.SessionID != "" {
		input.Data["session_id"] = req.SessionID
	}

	// 通过管道编排执行
	output, err := h.manager.ExecutePipeline("student_chat", input)
	if err != nil {
		// 检查是否是超时错误
		if c.Request.Context().Err() != nil {
			Error(c, http.StatusGatewayTimeout, 50004, "管道执行超时")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			Error(c, http.StatusInternalServerError, 50001, err.Error())
			return
		}
		Error(c, http.StatusInternalServerError, 50001, "对话处理失败: "+err.Error())
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

	Success(c, gin.H{
		"reply":                output.Data["reply"],
		"session_id":           output.Data["session_id"],
		"conversation_id":      output.Data["conversation_id"],
		"token_usage":          output.Data["token_usage"],
		"pipeline_duration_ms": output.Data["pipeline_duration_ms"],
	})
}

// HandleGetSessions 获取会话列表
// GET /api/conversations/sessions
func (h *Handler) HandleGetSessions(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

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
	convRepo := database.NewConversationRepository(db)

	sessions, total, err := convRepo.GetSessionsByStudent(userIDInt64, offset, pageSize)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询会话列表失败: "+err.Error())
		return
	}

	SuccessPage(c, sessions, total, page, pageSize)
}

// HandleGetConversations 获取对话历史
// GET /api/conversations
func (h *Handler) HandleGetConversations(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	teacherIDStr := c.Query("teacher_id")
	sessionID := c.Query("session_id")

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}
	convRepo := database.NewConversationRepository(db)

	var items []*database.Conversation
	var total int
	var err error

	if sessionID != "" {
		// 按 session_id 筛选
		items, total, err = convRepo.GetConversationsBySession(userIDInt64, sessionID, offset, pageSize)
	} else if teacherIDStr != "" {
		// 传了 teacher_id，按教师筛选（向后兼容）
		teacherID, parseErr := strconv.ParseInt(teacherIDStr, 10, 64)
		if parseErr != nil || teacherID <= 0 {
			Error(c, http.StatusBadRequest, 40004, "无效的 teacher_id 参数")
			return
		}
		items, total, err = convRepo.GetByStudentAndTeacher(userIDInt64, teacherID, offset, pageSize)
	} else {
		// 不传 teacher_id，返回与所有教师的对话
		items, total, err = convRepo.GetConversationsByStudent(userIDInt64, offset, pageSize)
	}

	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询对话历史失败: "+err.Error())
		return
	}

	SuccessPage(c, items, total, page, pageSize)
}

// ======================== 教师列表接口 ========================

// HandleGetTeachers 获取教师列表
// GET /api/teachers
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

	// 查询统计信息
	stats, err := userRepo.GetUserStats(userIDInt64, user.Role)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询用户统计信息失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"nickname":   user.Nickname,
		"role":       user.Role,
		"email":      user.Email,
		"created_at": user.CreatedAt,
		"stats":      stats,
	})
}

// ======================== 知识库接口 ========================

// HandleAddDocument 添加文档
// POST /api/documents
func (h *Handler) HandleAddDocument(c *gin.Context) {
	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
		Tags    string `json:"tags"`
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

	plugin, err := h.manager.GetPlugin("knowledge-retrieval")
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "知识库服务不可用")
		return
	}

	userContext := &core.UserContext{
		UserID: fmt.Sprintf("%d", userIDInt64),
	}

	input := &core.PluginInput{
		RequestID:   uuid.New().String(),
		UserContext:  userContext,
		Data: map[string]interface{}{
			"action":     "add",
			"title":      req.Title,
			"content":    req.Content,
			"tags":       req.Tags,
			"teacher_id": userIDInt64,
		},
		Context: c.Request.Context(),
	}

	output, err := plugin.Execute(c.Request.Context(), input)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "添加文档失败: "+err.Error())
		return
	}

	if !output.Success {
		errorCode := 50001
		if code, ok := output.Data["error_code"]; ok {
			errorCode = toInt(code, 50001)
		}
		Error(c, http.StatusBadRequest, errorCode, output.Error)
		return
	}

	Success(c, gin.H{
		"document_id":  output.Data["document_id"],
		"chunks_count": output.Data["chunks_count"],
	})
}

// HandleGetDocuments 获取文档列表
// GET /api/documents
func (h *Handler) HandleGetDocuments(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	plugin, err := h.manager.GetPlugin("knowledge-retrieval")
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "知识库服务不可用")
		return
	}

	input := &core.PluginInput{
		RequestID: uuid.New().String(),
		Data: map[string]interface{}{
			"action":     "list",
			"teacher_id": userIDInt64,
			"page":       page,
			"page_size":  pageSize,
		},
		Context: c.Request.Context(),
	}

	output, err := plugin.Execute(c.Request.Context(), input)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询文档列表失败: "+err.Error())
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
	SuccessPage(c, output.Data["documents"], total, page, pageSize)
}

// HandleDeleteDocument 删除文档
// DELETE /api/documents/:id
func (h *Handler) HandleDeleteDocument(c *gin.Context) {
	idStr := c.Param("id")
	docID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || docID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的文档 ID")
		return
	}

	plugin, err := h.manager.GetPlugin("knowledge-retrieval")
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "知识库服务不可用")
		return
	}

	input := &core.PluginInput{
		RequestID: uuid.New().String(),
		Data: map[string]interface{}{
			"action":      "delete",
			"document_id": docID,
		},
		Context: c.Request.Context(),
	}

	output, err := plugin.Execute(c.Request.Context(), input)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "删除文档失败: "+err.Error())
		return
	}

	if !output.Success {
		errorCode := 50001
		httpStatus := http.StatusInternalServerError
		if code, ok := output.Data["error_code"]; ok {
			errorCode = toInt(code, 50001)
			if errorCode == 40005 {
				httpStatus = http.StatusNotFound
			}
		}
		Error(c, httpStatus, errorCode, output.Error)
		return
	}

	Success(c, gin.H{
		"document_id": docID,
		"deleted":     true,
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
