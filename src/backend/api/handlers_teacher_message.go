package api

import (
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"

	"digital-twin/src/backend/database"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ======================== V2.0 迭代7 教师消息推送 ========================

// HandlePushTeacherMessage 教师向班级/指定学生推送消息
// POST /api/teacher-messages
func (h *Handler) HandlePushTeacherMessage(c *gin.Context) {
	var req struct {
		TargetType string `json:"target_type" binding:"required"`
		TargetID   int64  `json:"target_id" binding:"required"`
		Content    string `json:"content" binding:"required"`
		PersonaID  int64  `json:"persona_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	// 校验 target_type
	if req.TargetType != "class" && req.TargetType != "student" {
		Error(c, http.StatusBadRequest, 40004, "target_type 只能为 class 或 student")
		return
	}

	// 校验内容长度
	if utf8.RuneCountInString(req.Content) > 500 {
		Error(c, http.StatusBadRequest, 40004, "消息内容不能超过500字")
		return
	}

	// 从 JWT 获取教师信息
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

	msgRepo := database.NewTeacherMessageRepository(db)
	convRepo := database.NewConversationRepository(db)

	// 检查每日推送频率限制（≤20条/天）
	todayCount, err := msgRepo.GetTodayPushCount(userIDInt64)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询推送频率失败: "+err.Error())
		return
	}
	if todayCount >= 20 {
		Error(c, http.StatusTooManyRequests, 40050, "今日推送次数已达上限（20条/天）")
		return
	}

	// 根据 target_type 查询目标学生列表
	var studentPersonaIDs []int64

	if req.TargetType == "class" {
		// 从 class_members 表查询班级所有成员的 student_persona_id
		rows, err := db.Query(
			`SELECT student_persona_id FROM class_members WHERE class_id = ?`,
			req.TargetID,
		)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "查询班级成员失败: "+err.Error())
			return
		}
		defer rows.Close()

		for rows.Next() {
			var spID int64
			if err := rows.Scan(&spID); err != nil {
				Error(c, http.StatusInternalServerError, 50001, "扫描班级成员失败: "+err.Error())
				return
			}
			studentPersonaIDs = append(studentPersonaIDs, spID)
		}

		if len(studentPersonaIDs) == 0 {
			Error(c, http.StatusBadRequest, 40004, "该班级没有成员")
			return
		}
	} else {
		// target_type="student": 直接使用 target_id 作为 student_persona_id
		studentPersonaIDs = append(studentPersonaIDs, req.TargetID)
	}

	// 单次推送人数上限100人
	if len(studentPersonaIDs) > 100 {
		Error(c, http.StatusBadRequest, 40004, "单次推送人数不能超过100人")
		return
	}

	// 写入 teacher_messages 表
	teacherMsg := &database.TeacherMessage{
		TeacherID:  userIDInt64,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		Content:    req.Content,
		Status:     "pending",
	}

	msgID, err := msgRepo.Create(teacherMsg)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "创建推送消息失败: "+err.Error())
		return
	}

	// 批量写入 conversations 表：每个目标学生一条记录
	successCount := 0
	failCount := 0

	for _, studentPersonaID := range studentPersonaIDs {
		// 查询该教师分身和学生分身之间的最新 session_id
		sessionID, err := convRepo.GetLatestSessionByPersonas(req.PersonaID, studentPersonaID)
		if err != nil || sessionID == "" {
			// 没有已有会话，生成新的 UUID
			sessionID = uuid.New().String()
		}

		// 查询学生分身对应的 user_id
		var studentUserID int64
		db.QueryRow(`SELECT COALESCE(user_id, 0) FROM personas WHERE id = ?`, studentPersonaID).Scan(&studentUserID)

		conv := &database.Conversation{
			StudentID:        studentUserID,
			TeacherID:        userIDInt64,
			TeacherPersonaID: req.PersonaID,
			StudentPersonaID: studentPersonaID,
			SessionID:        sessionID,
			Role:             "system",
			Content:          req.Content,
			SenderType:       "teacher_push",
			TokenCount:       0,
		}

		_, err = convRepo.CreateWithSenderType(conv)
		if err != nil {
			failCount++
			continue
		}
		successCount++
	}

	// 更新 teacher_messages 状态为 "sent"
	_ = msgRepo.UpdateStatus(msgID, "sent")

	Success(c, gin.H{
		"id":            msgID,
		"teacher_id":    userIDInt64,
		"target_type":   req.TargetType,
		"target_id":     req.TargetID,
		"content":       req.Content,
		"status":        "sent",
		"total_targets": len(studentPersonaIDs),
		"success_count": successCount,
		"fail_count":    failCount,
		"created_at":    time.Now(),
	})
}

// HandleGetTeacherMessageHistory 获取教师推送历史
// GET /api/teacher-messages/history
func (h *Handler) HandleGetTeacherMessageHistory(c *gin.Context) {
	// 从 JWT 获取教师信息
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

	msgRepo := database.NewTeacherMessageRepository(db)

	messages, total, err := msgRepo.GetByTeacherID(userIDInt64, offset, pageSize)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询推送历史失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"messages":  messages,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
