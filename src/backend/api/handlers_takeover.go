package api

import (
	"net/http"
	"strconv"
	"time"

	"digital-twin/src/backend/database"

	"github.com/gin-gonic/gin"
)

// ======================== V2.0 迭代4 教师真人介入对话 ========================

// HandleTeacherReply 教师真人回复
// POST /api/chat/teacher-reply
func (h *Handler) HandleTeacherReply(c *gin.Context) {
	var req struct {
		StudentPersonaID int64  `json:"student_persona_id" binding:"required"`
		SessionID        string `json:"session_id" binding:"required"`
		Content          string `json:"content" binding:"required"`
		ReplyToID        int64  `json:"reply_to_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	// 从 JWT 获取教师信息
	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	personaID, _ := c.Get("persona_id")
	personaIDInt64, ok := personaID.(int64)
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
	relationRepo := database.NewRelationRepository(db)
	convRepo := database.NewConversationRepository(db)
	takeoverRepo := database.NewTakeoverRepository(db)

	// 校验教师分身
	teacherPersona, err := personaRepo.GetByID(personaIDInt64)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询教师分身失败: "+err.Error())
		return
	}
	if teacherPersona == nil || teacherPersona.Role != "teacher" {
		Error(c, http.StatusForbidden, 40032, "无权操作该会话")
		return
	}
	if teacherPersona.UserID != userIDInt64 {
		Error(c, http.StatusForbidden, 40032, "无权操作该会话")
		return
	}

	// 校验师生关系
	approved, err := relationRepo.IsApprovedByPersonas(personaIDInt64, req.StudentPersonaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询师生关系失败: "+err.Error())
		return
	}
	if !approved {
		Error(c, http.StatusForbidden, 40007, "未获得该学生的授权关系")
		return
	}

	// 查询学生分身获取 student_user_id
	studentPersona, err := personaRepo.GetByID(req.StudentPersonaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询学生分身失败: "+err.Error())
		return
	}
	if studentPersona == nil {
		Error(c, http.StatusNotFound, 40013, "学生分身不存在")
		return
	}

	// 获取引用消息内容
	var replyToContent string
	if req.ReplyToID > 0 {
		replyMsg, err := convRepo.GetByIDSimple(req.ReplyToID)
		if err == nil && replyMsg != nil {
			runes := []rune(replyMsg.Content)
			if len(runes) > 100 {
				replyToContent = string(runes[:100]) + "..."
			} else {
				replyToContent = replyMsg.Content
			}
		}
	}

	// 创建或获取接管记录（自动进入接管状态）
	takeover, err := takeoverRepo.CreateOrGetActive(personaIDInt64, req.StudentPersonaID, req.SessionID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "创建接管记录失败: "+err.Error())
		return
	}

	// 保存教师真人回复
	conv := &database.Conversation{
		StudentID:        studentPersona.UserID,
		TeacherID:        userIDInt64,
		TeacherPersonaID: personaIDInt64,
		StudentPersonaID: req.StudentPersonaID,
		SessionID:        req.SessionID,
		Role:             "assistant",
		Content:          req.Content,
		TokenCount:       0,
		SenderType:       "teacher",
		ReplyToID:        req.ReplyToID,
	}

	convID, err := convRepo.CreateWithSenderType(conv)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "保存教师回复失败: "+err.Error())
		return
	}

	conv.CreatedAt = time.Now()

	Success(c, gin.H{
		"conversation_id":  convID,
		"sender_type":      "teacher",
		"reply_to_id":      req.ReplyToID,
		"reply_to_content": replyToContent,
		"takeover_status":  takeover.Status,
		"created_at":       conv.CreatedAt,
	})
}

// HandleGetTakeoverStatus 查询接管状态
// GET /api/chat/takeover-status
func (h *Handler) HandleGetTakeoverStatus(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		Error(c, http.StatusBadRequest, 40004, "缺少 session_id 参数")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	takeoverRepo := database.NewTakeoverRepository(db)
	personaRepo := database.NewPersonaRepository(db)

	takeover, err := takeoverRepo.GetActiveBySession(sessionID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询接管状态失败: "+err.Error())
		return
	}

	if takeover == nil {
		Success(c, gin.H{
			"is_taken_over":      false,
			"teacher_persona_id": 0,
			"teacher_nickname":   "",
			"started_at":         nil,
		})
		return
	}

	// 查询教师昵称
	var teacherNickname string
	teacherPersona, err := personaRepo.GetByID(takeover.TeacherPersonaID)
	if err == nil && teacherPersona != nil {
		teacherNickname = teacherPersona.Nickname
	}

	Success(c, gin.H{
		"is_taken_over":      true,
		"teacher_persona_id": takeover.TeacherPersonaID,
		"teacher_nickname":   teacherNickname,
		"started_at":         takeover.StartedAt,
	})
}

// HandleEndTakeover 教师退出接管
// POST /api/chat/end-takeover
func (h *Handler) HandleEndTakeover(c *gin.Context) {
	var req struct {
		SessionID string `json:"session_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	personaID, _ := c.Get("persona_id")
	personaIDInt64, ok := personaID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	takeoverRepo := database.NewTakeoverRepository(db)

	if err := takeoverRepo.EndTakeover(req.SessionID, personaIDInt64); err != nil {
		Error(c, http.StatusBadRequest, 40031, "接管记录不存在或已结束")
		return
	}

	endedAt := time.Now()
	Success(c, gin.H{
		"session_id": req.SessionID,
		"status":     "ended",
		"ended_at":   endedAt,
	})
}

// HandleGetStudentConversations 教师查看学生对话记录
// GET /api/conversations/student/:student_persona_id
func (h *Handler) HandleGetStudentConversations(c *gin.Context) {
	studentPersonaIDStr := c.Param("student_persona_id")
	studentPersonaID, err := strconv.ParseInt(studentPersonaIDStr, 10, 64)
	if err != nil || studentPersonaID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的学生分身ID")
		return
	}

	personaID, _ := c.Get("persona_id")
	personaIDInt64, ok := personaID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	sessionID := c.Query("session_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	relationRepo := database.NewRelationRepository(db)
	convRepo := database.NewConversationRepository(db)
	personaRepo := database.NewPersonaRepository(db)
	takeoverRepo := database.NewTakeoverRepository(db)

	// 校验师生关系
	approved, err := relationRepo.IsApprovedByPersonas(personaIDInt64, studentPersonaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询师生关系失败: "+err.Error())
		return
	}

	// 如果分身维度关系不存在，尝试通过 student_id 维度查找
	// （兼容 student_persona_id 为 0 的旧关系，此时前端传的可能是 student_id）
	actualStudentPersonaID := studentPersonaID
	if !approved {
		// 尝试查找该教师分身下是否有以 studentPersonaID 作为 student_id 的关系
		var relStudentPersonaID int64
		db.QueryRow(
			`SELECT COALESCE(student_persona_id, 0) FROM teacher_student_relations WHERE teacher_persona_id = ? AND student_id = ? AND status = 'approved' LIMIT 1`,
			personaIDInt64, studentPersonaID,
		).Scan(&relStudentPersonaID)

		if relStudentPersonaID > 0 {
			actualStudentPersonaID = relStudentPersonaID
			approved = true
		} else {
			// 即使 student_persona_id 为 0，只要关系存在就允许查看
			var count int
			db.QueryRow(
				`SELECT COUNT(*) FROM teacher_student_relations WHERE teacher_persona_id = ? AND student_id = ? AND status = 'approved'`,
				personaIDInt64, studentPersonaID,
			).Scan(&count)
			if count > 0 {
				approved = true
				// actualStudentPersonaID 保持为传入的值（student_id），后续用 student_id 维度查询
			}
		}
	}

	if !approved {
		Error(c, http.StatusForbidden, 40007, "未获得该学生的授权关系")
		return
	}

	// 如果未指定 session_id，获取最新的
	if sessionID == "" {
		sessionID, _ = convRepo.GetLatestSessionByPersonas(personaIDInt64, actualStudentPersonaID)
		// 如果分身维度查不到，尝试用 student_id 维度
		if sessionID == "" && actualStudentPersonaID != studentPersonaID {
			sessionID, _ = convRepo.GetLatestSessionByPersonas(personaIDInt64, studentPersonaID)
		}
	}

	// 查询学生昵称
	var studentNickname string
	studentPersona, err := personaRepo.GetByID(actualStudentPersonaID)
	if err == nil && studentPersona != nil {
		studentNickname = studentPersona.Nickname
	}
	// 如果通过 persona 查不到昵称，尝试通过 user 查
	if studentNickname == "" {
		var userNickname string
		db.QueryRow(`SELECT COALESCE(nickname, '') FROM users WHERE id = ?`, studentPersonaID).Scan(&userNickname)
		if userNickname != "" {
			studentNickname = userNickname
		}
	}

	// 查询对话记录
	messages, total, err := convRepo.GetByTeacherAndStudentPersonas(personaIDInt64, actualStudentPersonaID, sessionID, offset, pageSize)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询对话记录失败: "+err.Error())
		return
	}

	// 如果分身维度查不到对话，尝试用 student_id 维度查询
	if total == 0 && actualStudentPersonaID != studentPersonaID {
		messages, total, err = convRepo.GetByTeacherAndStudentPersonas(personaIDInt64, studentPersonaID, sessionID, offset, pageSize)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "查询对话记录失败: "+err.Error())
			return
		}
	}

	// 查询接管状态
	takeoverStatus := "none"
	if sessionID != "" {
		takeover, err := takeoverRepo.GetActiveBySession(sessionID)
		if err == nil && takeover != nil {
			takeoverStatus = "active"
		}
	}

	Success(c, gin.H{
		"student_persona_id": actualStudentPersonaID,
		"student_nickname":   studentNickname,
		"session_id":         sessionID,
		"takeover_status":    takeoverStatus,
		"messages":           messages,
		"total":              total,
		"page":               page,
		"page_size":          pageSize,
	})
}
