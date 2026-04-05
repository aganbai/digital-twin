package api

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"digital-twin/src/backend/database"
	"digital-twin/src/harness/core"
	"digital-twin/src/plugins/dialogue"
	"digital-twin/src/plugins/knowledge"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ======================== 对话接口 ========================

// HandleChat 对话接口（通过管道编排执行）
// POST /api/chat
func (h *Handler) HandleChat(c *gin.Context) {
	var req struct {
		Message          string `json:"message" binding:"required"`
		TeacherID        int64  `json:"teacher_id"`         // 向后兼容，可选
		TeacherPersonaID int64  `json:"teacher_persona_id"` // 学生端可选，教师分身ID
		SessionID        string `json:"session_id"`
		AttachmentURL    string `json:"attachment_url"`  // V2.0 迭代5：附件文件路径
		AttachmentType   string `json:"attachment_type"` // V2.0 迭代5：附件类型
		AttachmentName   string `json:"attachment_name"` // V2.0 迭代5：附件原始文件名
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	// 兼容处理：优先使用 teacher_persona_id，回退到 teacher_id
	if req.TeacherPersonaID <= 0 && req.TeacherID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: teacher_persona_id 或 teacher_id 至少提供一个")
		return
	}
	if req.TeacherPersonaID > 0 && req.TeacherID <= 0 {
		req.TeacherID = req.TeacherPersonaID // 回退兼容
	}
	if req.TeacherID > 0 && req.TeacherPersonaID <= 0 {
		req.TeacherPersonaID = req.TeacherID // 反向兼容：teacher_id 在 V2.0 中实际就是 persona_id
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

	// 从 JWT 获取 persona_id
	personaID, _ := c.Get("persona_id")
	personaIDInt64, _ := personaID.(int64)

	// 🆕 V2.0 迭代9 M3：新会话指令检测
	isNewSessionCmd := isNewSessionCommand(req.Message)
	var newSessionID string
	if isNewSessionCmd {
		newSessionID = generateSessionID()
	}

	// 🆕 师生授权检查（支持分身维度 + 启停状态检查 + user_id 回退）
	if fmt.Sprintf("%v", role) == "student" {
		db := h.manager.GetDB()
		if db != nil {
			relationRepo := database.NewRelationRepository(db)
			authorized := false
			var chatPermErr *database.ChatPermissionError

			// 优先尝试分身维度授权检查
			if personaIDInt64 > 0 && req.TeacherPersonaID > 0 {
				chatPermErr = relationRepo.CheckChatPermission(req.TeacherPersonaID, personaIDInt64)
				if chatPermErr == nil {
					authorized = true
				} else if chatPermErr.Code != 40007 {
					// 非"关系不存在"的错误（如分身停用、班级停用、权限关闭），直接拒绝，不回退
					Error(c, http.StatusForbidden, chatPermErr.Code, chatPermErr.Message)
					return
				}
			}

			// 分身维度关系不存在时（40007），回退到 user_id 维度检查（兼容旧关系）
			if !authorized {
				approved, err := relationRepo.IsApproved(req.TeacherID, userIDInt64)
				if err != nil {
					Error(c, http.StatusInternalServerError, 50001, "查询授权关系失败")
					return
				}
				if approved {
					authorized = true
				}
			}

			if !authorized {
				Error(c, http.StatusForbidden, 40007, "未获得该教师授权，请先申请")
				return
			}
		}
	}

	// 🆕 V2.0 迭代4：接管状态检查
	if fmt.Sprintf("%v", role) == "student" && req.SessionID != "" {
		db := h.manager.GetDB()
		if db != nil {
			takeoverRepo := database.NewTakeoverRepository(db)
			isTakenOver, err := takeoverRepo.IsSessionTakenOver(req.SessionID)
			if err == nil && isTakenOver {
				// 🔧 修复外键约束：通过 teacher_persona_id 查询对应的 user_id
				actualTeacherID := req.TeacherID
				if req.TeacherPersonaID > 0 {
					personaRepo := database.NewPersonaRepository(db)
					teacherPersona, err := personaRepo.GetByID(req.TeacherPersonaID)
					if err == nil && teacherPersona != nil {
						actualTeacherID = teacherPersona.UserID
					}
				}

				// 接管中：保存学生消息但不调用 AI
				convRepo := database.NewConversationRepository(db)
				conv := &database.Conversation{
					StudentID:        userIDInt64,
					TeacherID:        actualTeacherID, // 使用正确的 user_id
					TeacherPersonaID: req.TeacherPersonaID,
					StudentPersonaID: personaIDInt64,
					SessionID:        req.SessionID,
					Role:             "user",
					Content:          req.Message,
					SenderType:       "student",
				}
				convID, _ := convRepo.CreateWithSenderType(conv)

				// 查询接管教师信息
				takeover, _ := takeoverRepo.GetActiveBySession(req.SessionID)
				var teacherNickname string
				var startedAt interface{}
				if takeover != nil {
					personaRepo := database.NewPersonaRepository(db)
					tp, _ := personaRepo.GetByID(takeover.TeacherPersonaID)
					if tp != nil {
						teacherNickname = tp.Nickname
					}
					startedAt = takeover.StartedAt
				}

				c.JSON(http.StatusOK, gin.H{
					"code":    40030,
					"message": "老师正在亲自回复中，请等待老师回复",
					"data": gin.H{
						"conversation_id": convID,
						"sender_type":     "student",
						"takeover_info": gin.H{
							"teacher_nickname": teacherNickname,
							"started_at":       startedAt,
						},
					},
				})
				return
			}
		}
	}

	// 🆕 V2.0 迭代9 M3：新会话指令触发处理
	if isNewSessionCmd {
		// 保存用户的指令消息到 conversations 表
		db := h.manager.GetDB()
		if db != nil {
			// 🔧 修复外键约束：通过 teacher_persona_id 查询对应的 user_id
			actualTeacherID := req.TeacherID
			if req.TeacherPersonaID > 0 {
				personaRepo := database.NewPersonaRepository(db)
				teacherPersona, err := personaRepo.GetByID(req.TeacherPersonaID)
				if err == nil && teacherPersona != nil {
					actualTeacherID = teacherPersona.UserID
				}
			}

			convRepo := database.NewConversationRepository(db)
			conv := &database.Conversation{
				StudentID:        userIDInt64,
				TeacherID:        actualTeacherID, // 使用正确的 user_id
				TeacherPersonaID: req.TeacherPersonaID,
				StudentPersonaID: personaIDInt64,
				SessionID:        newSessionID,
				Role:             "user",
				Content:          req.Message,
				SenderType:       "student",
			}
			convRepo.CreateWithPersonas(conv)
		}

		Success(c, gin.H{
			"type":       "new_session",
			"session_id": newSessionID,
			"message":    "已为您开启新会话，让我们开始新的话题吧！",
		})
		return
	}

	userContext := &core.UserContext{
		UserID:     fmt.Sprintf("%d", userIDInt64),
		Role:       fmt.Sprintf("%v", role),
		SessionID:  req.SessionID,
		Attributes: map[string]interface{}{"username": username},
	}

	// V2.0 迭代5：附件处理 — 解析附件内容拼接到 message
	message := req.Message
	if req.AttachmentURL != "" {
		attachmentContent := h.parseAttachment(req.AttachmentURL, req.AttachmentType, req.AttachmentName)
		if attachmentContent != "" {
			message = message + "\n" + attachmentContent
		}
	}

	// 构建管道输入
	input := &core.PluginInput{
		RequestID:   uuid.New().String(),
		UserContext: userContext,
		Data: map[string]interface{}{
			"action":     "chat",
			"message":    message,
			"teacher_id": req.TeacherID,
		},
		Context: c.Request.Context(),
	}

	// 传入分身维度信息
	if personaIDInt64 > 0 {
		input.Data["student_persona_id"] = personaIDInt64
	}
	if req.TeacherPersonaID > 0 {
		input.Data["teacher_persona_id"] = req.TeacherPersonaID
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

// HandleGetSessions 获取会话列表（V2.0 迭代9 增强：支持 teacher_persona_id 过滤）
// GET /api/conversations/sessions?teacher_persona_id=123&page=1&page_size=20
func (h *Handler) HandleGetSessions(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	// 从 JWT 获取 persona_id
	personaID, _ := c.Get("persona_id")
	personaIDInt64, _ := personaID.(int64)

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 解析 teacher_persona_id 过滤参数
	teacherPersonaIDStr := c.Query("teacher_persona_id")
	var teacherPersonaID int64
	if teacherPersonaIDStr != "" {
		var parseErr error
		teacherPersonaID, parseErr = strconv.ParseInt(teacherPersonaIDStr, 10, 64)
		if parseErr != nil || teacherPersonaID <= 0 {
			Error(c, http.StatusBadRequest, 40004, "无效的 teacher_persona_id 参数")
			return
		}
	}

	offset := (page - 1) * pageSize

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}
	convRepo := database.NewConversationRepository(db)

	// 分身维度查询（支持 teacher_persona_id 过滤）
	if personaIDInt64 > 0 {
		sessions, total, err := convRepo.GetSessionsByTeacherPersona(personaIDInt64, teacherPersonaID, offset, pageSize)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "查询会话列表失败: "+err.Error())
			return
		}
		SuccessPage(c, sessions, total, page, pageSize)
		return
	}

	// 向后兼容：user_id 维度（不支持 teacher_persona_id 过滤）
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

	// 从 JWT 获取 persona_id
	personaID, _ := c.Get("persona_id")
	personaIDInt64, _ := personaID.(int64)

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
	teacherPersonaIDStr := c.Query("teacher_persona_id")
	sessionID := c.Query("session_id")

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}
	convRepo := database.NewConversationRepository(db)

	// 分身维度查询（优先使用 teacher_persona_id 参数，回退到 teacher_id 参数）
	teacherPersonaIDForQuery := teacherPersonaIDStr
	if teacherPersonaIDForQuery == "" && teacherIDStr != "" {
		// V2.0 中 teacher_id 实际就是 teacher_persona_id，兼容处理
		teacherPersonaIDForQuery = teacherIDStr
	}
	if personaIDInt64 > 0 && teacherPersonaIDForQuery != "" {
		teacherPersonaID, parseErr := strconv.ParseInt(teacherPersonaIDForQuery, 10, 64)
		if parseErr != nil || teacherPersonaID <= 0 {
			Error(c, http.StatusBadRequest, 40004, "无效的 teacher_persona_id 参数")
			return
		}
		items, total, err := convRepo.GetByPersonas(teacherPersonaID, personaIDInt64, sessionID, offset, pageSize)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "查询对话历史失败: "+err.Error())
			return
		}
		SuccessPage(c, items, total, page, pageSize)
		return
	}

	// 向后兼容：user_id 维度
	var items []*database.Conversation
	var total int
	var err error

	if sessionID != "" {
		items, total, err = convRepo.GetConversationsBySession(userIDInt64, sessionID, offset, pageSize)
	} else if teacherIDStr != "" {
		teacherID, parseErr := strconv.ParseInt(teacherIDStr, 10, 64)
		if parseErr != nil || teacherID <= 0 {
			Error(c, http.StatusBadRequest, 40004, "无效的 teacher_id 参数")
			return
		}
		items, total, err = convRepo.GetByStudentAndTeacher(userIDInt64, teacherID, offset, pageSize)
	} else {
		items, total, err = convRepo.GetConversationsByStudent(userIDInt64, offset, pageSize)
	}

	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询对话历史失败: "+err.Error())
		return
	}

	SuccessPage(c, items, total, page, pageSize)
}

// ======================== 新会话关键词列表（V2.0 迭代9 M3） ========================

var newSessionKeywords = []string{
	"新话题",
	"开始新对话",
	"新会话",
	"换个话题",
	"重新开始",
}

// isNewSessionCommand 检测消息是否包含新会话关键词
func isNewSessionCommand(message string) bool {
	for _, keyword := range newSessionKeywords {
		if strings.Contains(message, keyword) {
			return true
		}
	}
	return false
}

// generateSessionID 生成新的 session_id（格式：sess_{uuid}）
func generateSessionID() string {
	return "sess_" + uuid.New().String()
}

// HandleChatStream SSE 流式对话
// POST /api/chat/stream
func (h *Handler) HandleChatStream(c *gin.Context) {
	var req struct {
		Message          string `json:"message" binding:"required"`
		TeacherID        int64  `json:"teacher_id"`         // 向后兼容，可选
		TeacherPersonaID int64  `json:"teacher_persona_id"` // 学生端可选，教师分身ID
		SessionID        string `json:"session_id"`
		AttachmentURL    string `json:"attachment_url"`  // V2.0 迭代5：附件文件路径
		AttachmentType   string `json:"attachment_type"` // V2.0 迭代5：附件类型
		AttachmentName   string `json:"attachment_name"` // V2.0 迭代5：附件原始文件名
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	// 兼容处理：优先使用 teacher_persona_id，回退到 teacher_id
	if req.TeacherPersonaID <= 0 && req.TeacherID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: teacher_persona_id 或 teacher_id 至少提供一个")
		return
	}
	if req.TeacherPersonaID > 0 && req.TeacherID <= 0 {
		req.TeacherID = req.TeacherPersonaID // 回退兼容
	}
	if req.TeacherID > 0 && req.TeacherPersonaID <= 0 {
		req.TeacherPersonaID = req.TeacherID // 反向兼容：teacher_id 在 V2.0 中实际就是 persona_id
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

	// 从 JWT 获取 persona_id
	personaID, _ := c.Get("persona_id")
	personaIDInt64, _ := personaID.(int64)

	// 🆕 V2.0 迭代9 M3：新会话指令检测
	isNewSessionCmd := isNewSessionCommand(req.Message)
	var newSessionID string
	if isNewSessionCmd {
		newSessionID = generateSessionID()
	}

	// 师生授权检查（支持分身维度 + 启停状态检查 + user_id 回退）
	if fmt.Sprintf("%v", role) == "student" {
		db := h.manager.GetDB()
		if db != nil {
			relationRepo := database.NewRelationRepository(db)
			authorized := false
			var chatPermErr *database.ChatPermissionError

			// 优先尝试分身维度授权检查
			if personaIDInt64 > 0 && req.TeacherPersonaID > 0 {
				chatPermErr = relationRepo.CheckChatPermission(req.TeacherPersonaID, personaIDInt64)
				if chatPermErr == nil {
					authorized = true
				} else if chatPermErr.Code != 40007 {
					// 非"关系不存在"的错误（如分身停用、班级停用、权限关闭），直接拒绝，不回退
					Error(c, http.StatusForbidden, chatPermErr.Code, chatPermErr.Message)
					return
				}
			}

			// 分身维度关系不存在时（40007），回退到 user_id 维度检查（兼容旧关系）
			if !authorized {
				approved, err := relationRepo.IsApproved(req.TeacherID, userIDInt64)
				if err != nil {
					Error(c, http.StatusInternalServerError, 50001, "查询授权关系失败")
					return
				}
				if approved {
					authorized = true
				}
			}

			if !authorized {
				Error(c, http.StatusForbidden, 40007, "未获得该教师授权，请先申请")
				return
			}
		}
	}

	// 🆕 V2.0 迭代9 M3：新会话指令触发处理
	if isNewSessionCmd {
		// 保存用户的指令消息到 conversations 表
		db := h.manager.GetDB()
		if db != nil {
			// 🔧 修复外键约束：通过 teacher_persona_id 查询对应的 user_id
			actualTeacherID := req.TeacherID
			if req.TeacherPersonaID > 0 {
				personaRepo := database.NewPersonaRepository(db)
				teacherPersona, err := personaRepo.GetByID(req.TeacherPersonaID)
				if err == nil && teacherPersona != nil {
					actualTeacherID = teacherPersona.UserID
				}
			}

			convRepo := database.NewConversationRepository(db)
			conv := &database.Conversation{
				StudentID:        userIDInt64,
				TeacherID:        actualTeacherID, // 使用正确的 user_id
				TeacherPersonaID: req.TeacherPersonaID,
				StudentPersonaID: personaIDInt64,
				SessionID:        newSessionID,
				Role:             "user",
				Content:          req.Message,
				SenderType:       "student",
			}
			convRepo.CreateWithPersonas(conv)
		}

		// 设置 SSE 响应头
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")

		// 发送 new_session 事件
		c.SSEvent("", gin.H{
			"type":       "new_session",
			"session_id": newSessionID,
			"message":    "已为您开启新会话，让我们开始新的话题吧！",
		})
		c.Writer.Flush()

		// 发送 done 事件结束 SSE 流
		c.SSEvent("", gin.H{"type": "done"})
		c.Writer.Flush()
		return
	}

	// 🆕 V2.0 迭代4：接管状态检查
	if fmt.Sprintf("%v", role) == "student" && req.SessionID != "" {
		db := h.manager.GetDB()
		if db != nil {
			takeoverRepo := database.NewTakeoverRepository(db)
			isTakenOver, err := takeoverRepo.IsSessionTakenOver(req.SessionID)
			if err == nil && isTakenOver {
				// 🔧 修复外键约束：通过 teacher_persona_id 查询对应的 user_id
				actualTeacherID := req.TeacherID
				if req.TeacherPersonaID > 0 {
					personaRepo := database.NewPersonaRepository(db)
					teacherPersona, err := personaRepo.GetByID(req.TeacherPersonaID)
					if err == nil && teacherPersona != nil {
						actualTeacherID = teacherPersona.UserID
					}
				}

				// 接管中：保存学生消息但不调用 AI，返回 JSON 而非 SSE
				convRepo := database.NewConversationRepository(db)
				conv := &database.Conversation{
					StudentID:        userIDInt64,
					TeacherID:        actualTeacherID, // 使用正确的 user_id
					TeacherPersonaID: req.TeacherPersonaID,
					StudentPersonaID: personaIDInt64,
					SessionID:        req.SessionID,
					Role:             "user",
					Content:          req.Message,
					SenderType:       "student",
				}
				convID, _ := convRepo.CreateWithSenderType(conv)

				// 查询接管教师信息
				streamTakeover, _ := takeoverRepo.GetActiveBySession(req.SessionID)
				var streamTeacherNickname string
				var streamStartedAt interface{}
				if streamTakeover != nil {
					personaRepo := database.NewPersonaRepository(db)
					tp, _ := personaRepo.GetByID(streamTakeover.TeacherPersonaID)
					if tp != nil {
						streamTeacherNickname = tp.Nickname
					}
					streamStartedAt = streamTakeover.StartedAt
				}

				c.JSON(http.StatusOK, gin.H{
					"code":    40030,
					"message": "老师正在亲自回复中，请等待老师回复",
					"data": gin.H{
						"conversation_id": convID,
						"sender_type":     "student",
						"takeover_info": gin.H{
							"teacher_nickname": streamTeacherNickname,
							"started_at":       streamStartedAt,
						},
					},
				})
				return
			}
		}
	}

	// 获取 dialogue 插件
	plugin, err := h.manager.GetPlugin("dialogue")
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "对话服务不可用")
		return
	}

	// session_id 处理
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// 发送 start 事件
	c.SSEvent("", gin.H{"type": "start", "session_id": sessionID})
	c.Writer.Flush()

	// 构建 sse_writer 回调（用于发送 delta 事件）
	sseWriter := func(content string) {
		c.SSEvent("", gin.H{"type": "delta", "content": content})
		c.Writer.Flush()
	}

	// 构建 thinking_step_writer 回调（用于发送 thinking_step 事件，V2.0 迭代9）
	thinkingStepWriter := func(eventType string, data map[string]interface{}) {
		c.SSEvent("", data)
		c.Writer.Flush()
	}

	// 构建用户上下文
	userContext := &core.UserContext{
		UserID:     fmt.Sprintf("%d", userIDInt64),
		Role:       fmt.Sprintf("%v", role),
		SessionID:  sessionID,
		Attributes: map[string]interface{}{"username": username},
	}

	// V2.0 迭代5：附件处理 — 解析附件内容拼接到 message
	streamMessage := req.Message
	if req.AttachmentURL != "" {
		attachmentContent := h.parseAttachment(req.AttachmentURL, req.AttachmentType, req.AttachmentName)
		if attachmentContent != "" {
			streamMessage = streamMessage + "\n" + attachmentContent
		}
	}

	// 构建插件输入
	input := &core.PluginInput{
		RequestID:   uuid.New().String(),
		UserContext: userContext,
		Data: map[string]interface{}{
			"action":               "chat_stream",
			"message":              streamMessage,
			"teacher_id":           req.TeacherID,
			"session_id":           sessionID,
			"sse_writer":           sseWriter,
			"thinking_step_writer": thinkingStepWriter, // V2.0 迭代9
		},
		Context: c.Request.Context(),
	}

	// 传入分身维度信息
	if personaIDInt64 > 0 {
		input.Data["student_persona_id"] = personaIDInt64
	}
	if req.TeacherPersonaID > 0 {
		input.Data["teacher_persona_id"] = req.TeacherPersonaID
	}

	// 调用 dialogue 插件的 chat_stream action
	output, err := plugin.Execute(c.Request.Context(), input)
	if err != nil {
		c.SSEvent("", gin.H{"type": "error", "code": 50002, "message": "大模型调用失败: " + err.Error()})
		c.Writer.Flush()
		return
	}

	if !output.Success {
		c.SSEvent("", gin.H{"type": "error", "code": 50002, "message": output.Error})
		c.Writer.Flush()
		return
	}

	// 发送 done 事件
	c.SSEvent("", gin.H{
		"type":            "done",
		"conversation_id": output.Data["conversation_id"],
		"token_usage":     output.Data["token_usage"],
	})
	c.Writer.Flush()
}

// ======================== 附件处理辅助方法（V2.0 迭代5） ========================

// parseAttachment 解析附件内容，返回格式化的附件文本
// 使用已有的 FileParser 解析文件内容，拼接格式：[附件: {filename}]\n{content}\n[/附件]
func (h *Handler) parseAttachment(attachmentURL, attachmentType, attachmentName string) string {
	if attachmentURL == "" {
		return ""
	}

	// 图片类型不解析内容，只标记
	ext := strings.ToLower(filepath.Ext(attachmentURL))
	if attachmentType == "image" || ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
		return fmt.Sprintf("[附件: %s]\n（图片文件，请根据文件名和上下文理解内容）\n[/附件]", attachmentName)
	}

	// 使用 FileParser 解析文本类附件
	// 附件 URL 格式为 /uploads/... 需要转为本地路径
	localPath := strings.TrimPrefix(attachmentURL, "/")
	parser := knowledge.NewFileParser()
	content, err := parser.Parse(localPath)
	if err != nil {
		log.Printf("[HandleChat] 解析附件失败: %v", err)
		return fmt.Sprintf("[附件: %s]\n（文件解析失败）\n[/附件]", attachmentName)
	}

	// 限制附件内容长度（防止超出 LLM 上下文窗口）
	maxContentLen := 8000
	contentRunes := []rune(content)
	if len(contentRunes) > maxContentLen {
		content = string(contentRunes[:maxContentLen]) + "\n...(内容已截断)"
	}

	return fmt.Sprintf("[附件: %s]\n%s\n[/附件]", attachmentName, content)
}

// ======================== V2.0 迭代9 M5：会话标题生成 ========================

// HandleGenerateSessionTitle 手动触发生成会话标题
// POST /api/conversations/sessions/:session_id/title
func (h *Handler) HandleGenerateSessionTitle(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		Error(c, http.StatusBadRequest, 40004, "缺少 session_id 参数")
		return
	}

	// 从 JWT 中间件获取用户信息
	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	// 从 JWT 获取 persona_id
	personaID, _ := c.Get("persona_id")
	personaIDInt64, _ := personaID.(int64)

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	// 验证会话归属（权限校验）
	var belongsToUser bool
	if personaIDInt64 > 0 {
		// 分身维度验证
		var count int
		err := db.QueryRow(
			`SELECT COUNT(*) FROM conversations WHERE session_id = ? AND student_persona_id = ?`,
			sessionID, personaIDInt64,
		).Scan(&count)
		if err == nil && count > 0 {
			belongsToUser = true
		}
	}
	if !belongsToUser {
		// user_id 维度验证（向后兼容）
		var count int
		err := db.QueryRow(
			`SELECT COUNT(*) FROM conversations WHERE session_id = ? AND student_id = ?`,
			sessionID, userIDInt64,
		).Scan(&count)
		if err == nil && count > 0 {
			belongsToUser = true
		}
	}

	if !belongsToUser {
		Error(c, http.StatusForbidden, 40007, "无权操作该会话")
		return
	}

	// 异步执行标题生成
	go h.generateSessionTitle(sessionID, personaIDInt64)

	Success(c, gin.H{
		"message": "标题生成任务已提交",
	})
}

// generateSessionTitle 生成会话标题的核心逻辑
func (h *Handler) generateSessionTitle(sessionID string, studentPersonaID int64) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[标题生成] panic recovered: %v\n", r)
		}
	}()

	db := h.manager.GetDB()
	if db == nil {
		log.Printf("[标题生成] 数据库服务不可用\n")
		return
	}

	titleRepo := database.NewSessionTitleRepository(db)

	// 幂等性检查：已存在标题则跳过
	existingTitle, err := titleRepo.GetBySessionID(sessionID)
	if err != nil {
		log.Printf("[标题生成] 查询标题失败: %v\n", err)
		return
	}
	if existingTitle != nil {
		log.Printf("[标题生成] 会话 %s 已有标题，跳过生成\n", sessionID)
		return
	}

	// 获取会话消息（最近10条）
	convRepo := database.NewConversationRepository(db)
	var messages []*database.Conversation
	var teacherPersonaID int64

	if studentPersonaID > 0 {
		// 分身维度查询
		messages, _, err = convRepo.GetByPersonas(0, studentPersonaID, sessionID, 0, 10)
		if len(messages) > 0 {
			teacherPersonaID = messages[0].TeacherPersonaID
		}
	}
	if len(messages) == 0 {
		// user_id 维度查询（向后兼容）
		var userID int64
		rows, queryErr := db.Query(
			`SELECT id, student_id, teacher_id, COALESCE(teacher_persona_id, 0), COALESCE(student_persona_id, 0), session_id, role, content, token_count, COALESCE(sender_type, ''), COALESCE(reply_to_id, 0), created_at 
			 FROM conversations WHERE session_id = ? 
			 ORDER BY created_at ASC LIMIT 10`,
			sessionID,
		)
		if queryErr != nil {
			log.Printf("[标题生成] 查询消息失败: %v\n", queryErr)
			return
		}
		defer rows.Close()
		for rows.Next() {
			conv := &database.Conversation{}
			if scanErr := rows.Scan(&conv.ID, &conv.StudentID, &conv.TeacherID, &conv.TeacherPersonaID, &conv.StudentPersonaID,
				&conv.SessionID, &conv.Role, &conv.Content, &conv.TokenCount, &conv.SenderType, &conv.ReplyToID, &conv.CreatedAt); scanErr != nil {
				continue
			}
			messages = append(messages, conv)
			if teacherPersonaID == 0 {
				teacherPersonaID = conv.TeacherPersonaID
			}
			if userID == 0 {
				userID = conv.StudentID
			}
		}
	}

	if len(messages) == 0 {
		log.Printf("[标题生成] 会话 %s 没有消息\n", sessionID)
		return
	}

	var title string

	// 判断消息数量：≤2条使用首条用户消息前20字
	if len(messages) <= 2 {
		// 查找首条用户消息
		for _, msg := range messages {
			if msg.Role == "user" {
				runes := []rune(msg.Content)
				if len(runes) > 20 {
					title = string(runes[:20])
				} else {
					title = msg.Content
				}
				break
			}
		}
		if title == "" {
			// 没有用户消息，使用第一条消息
			runes := []rune(messages[0].Content)
			if len(runes) > 20 {
				title = string(runes[:20])
			} else {
				title = messages[0].Content
			}
		}
	} else {
		// >2条消息，调用 LLM 生成标题
		title = h.generateTitleWithLLM(messages)
	}

	if title == "" {
		title = "新对话"
	}

	// 存储标题
	if studentPersonaID == 0 && len(messages) > 0 {
		studentPersonaID = messages[0].StudentPersonaID
	}
	if teacherPersonaID == 0 && len(messages) > 0 {
		teacherPersonaID = messages[0].TeacherPersonaID
	}

	err = titleRepo.Upsert(sessionID, studentPersonaID, teacherPersonaID, title)
	if err != nil {
		log.Printf("[标题生成] 存储标题失败: %v\n", err)
		return
	}

	log.Printf("[标题生成] 会话 %s 标题生成成功: %s\n", sessionID, title)
}

// generateTitleWithLLM 调用 LLM 生成会话标题
func (h *Handler) generateTitleWithLLM(messages []*database.Conversation) string {
	// 获取 dialogue 插件
	plugin, err := h.manager.GetPlugin("dialogue")
	if err != nil {
		log.Printf("[标题生成] 获取对话插件失败: %v\n", err)
		return ""
	}

	dialoguePlugin, ok := plugin.(*dialogue.DialoguePlugin)
	if !ok {
		log.Printf("[标题生成] 插件类型断言失败\n")
		return ""
	}

	llmClient := dialoguePlugin.GetLLMClient()
	if llmClient == nil {
		log.Printf("[标题生成] LLM 客户端不可用\n")
		return ""
	}

	// 构建消息内容摘要（最近10条）
	var contentBuilder strings.Builder
	maxMessages := 10
	if len(messages) < maxMessages {
		maxMessages = len(messages)
	}
	for i := len(messages) - maxMessages; i < len(messages); i++ {
		msg := messages[i]
		role := "用户"
		if msg.Role == "assistant" {
			role = "AI"
		}
		contentBuilder.WriteString(fmt.Sprintf("%s: %s\n", role, msg.Content))
	}

	// 构建 LLM prompt
	prompt := fmt.Sprintf(`请为以下对话生成一个简短的标题（10-20字），概括对话的主要话题：

%s

要求：
- 标题简洁明了
- 体现对话的核心话题
- 不超过20个字

请直接输出标题，不要包含其他内容。`, contentBuilder.String())

	chatMessages := []dialogue.ChatMessage{
		{Role: "user", Content: prompt},
	}

	resp, err := llmClient.Chat(chatMessages)
	if err != nil {
		log.Printf("[标题生成] LLM 调用失败: %v\n", err)
		return ""
	}

	// 清理响应（去除可能的引号、换行等）
	title := strings.TrimSpace(resp.Content)
	title = strings.Trim(title, "\"'「」【】")

	// 限制标题长度
	runes := []rune(title)
	if len(runes) > 20 {
		title = string(runes[:20])
	}

	return title
}
