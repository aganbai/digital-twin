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
				// 接管中：保存学生消息但不调用 AI
				convRepo := database.NewConversationRepository(db)
				conv := &database.Conversation{
					StudentID:        userIDInt64,
					TeacherID:        req.TeacherID,
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

// HandleGetSessions 获取会话列表
// GET /api/conversations/sessions
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

	// 分身维度查询
	if personaIDInt64 > 0 {
		sessions, total, err := convRepo.GetSessionsByPersona(personaIDInt64, offset, pageSize)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "查询会话列表失败: "+err.Error())
			return
		}
		SuccessPage(c, sessions, total, page, pageSize)
		return
	}

	// 向后兼容：user_id 维度
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

	// 🆕 V2.0 迭代4：接管状态检查
	if fmt.Sprintf("%v", role) == "student" && req.SessionID != "" {
		db := h.manager.GetDB()
		if db != nil {
			takeoverRepo := database.NewTakeoverRepository(db)
			isTakenOver, err := takeoverRepo.IsSessionTakenOver(req.SessionID)
			if err == nil && isTakenOver {
				// 接管中：保存学生消息但不调用 AI，返回 JSON 而非 SSE
				convRepo := database.NewConversationRepository(db)
				conv := &database.Conversation{
					StudentID:        userIDInt64,
					TeacherID:        req.TeacherID,
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

	// 构建 sse_writer 回调
	sseWriter := func(content string) {
		c.SSEvent("", gin.H{"type": "delta", "content": content})
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
			"action":     "chat_stream",
			"message":    streamMessage,
			"teacher_id": req.TeacherID,
			"session_id": sessionID,
			"sse_writer": sseWriter,
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
