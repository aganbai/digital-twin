package api

import (
	"digital-twin/src/backend/database"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetTeacherChatListResponse 教师端聊天列表响应
type GetTeacherChatListResponse struct {
	Classes []TeacherChatClassItem `json:"classes"`
	Total   int                    `json:"total"`
}

// TeacherChatClassItem 教师端班级聊天项
type TeacherChatClassItem struct {
	ClassID   int64                `json:"class_id"`
	ClassName string               `json:"class_name"`
	Subject   string               `json:"subject,omitempty"`
	Students  []TeacherChatStudent `json:"students"`
	IsPinned  bool                 `json:"is_pinned"`
	PinTime   *string              `json:"pin_time,omitempty"`
}

// TeacherChatStudent 教师端学生聊天信息
type TeacherChatStudent struct {
	StudentPersonaID int64   `json:"student_persona_id"`
	StudentNickname  string  `json:"student_nickname"`
	StudentAvatar    string  `json:"student_avatar"`
	LastMessage      string  `json:"last_message,omitempty"`
	LastMessageTime  *string `json:"last_message_time,omitempty"`
	UnreadCount      int     `json:"unread_count"`
	IsPinned         bool    `json:"is_pinned"`
}

// HandleGetTeacherChatList 获取教师端聊天列表（按班级组织）
func (h *Handler) HandleGetTeacherChatList(c *gin.Context) {
	userID, _ := c.Get("user_id")
	teacherID := userID.(int64)

	// V2.0 迭代9修复：直接使用token中的persona_id，保证与班级查询API的一致性
	personaIDInterface, exists := c.Get("persona_id")
	var personaID int64
	if exists {
		personaID, _ = personaIDInterface.(int64)
	}

	// 如果token中没有persona_id，则查询默认分身（兼容旧逻辑）
	if personaID == 0 {
		var err error
		personaID, err = h.getDefaultPersonaID(teacherID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "获取分身失败"})
			return
		}
	}

	// 获取置顶记录
	pinRepo := database.NewChatPinRepository(h.db)
	pinnedClasses, _ := pinRepo.GetPinnedTargets(teacherID, "teacher", "class", personaID)
	pinnedStudents, _ := pinRepo.GetPinnedTargets(teacherID, "teacher", "student", personaID)

	pinnedClassMap := make(map[int64]bool)
	for _, id := range pinnedClasses {
		pinnedClassMap[id] = true
	}
	pinnedStudentMap := make(map[int64]bool)
	for _, id := range pinnedStudents {
		pinnedStudentMap[id] = true
	}

	// 查询教师的班级列表
	rows, err := h.db.DB.Query(`
		SELECT id, name, subject
		FROM classes
		WHERE persona_id = ? AND is_active = 1
		ORDER BY created_at DESC`, personaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "查询班级失败"})
		return
	}
	defer rows.Close()

	var classes []TeacherChatClassItem
	for rows.Next() {
		var classItem TeacherChatClassItem
		if err := rows.Scan(&classItem.ClassID, &classItem.ClassName, &classItem.Subject); err != nil {
			continue
		}
		classItem.IsPinned = pinnedClassMap[classItem.ClassID]

		// 查询班级下的学生
		studentRows, err := h.db.DB.Query(`
			SELECT 
				cm.student_persona_id, p.nickname, p.avatar,
				(
					SELECT content FROM conversations 
					WHERE (student_persona_id = cm.student_persona_id AND teacher_persona_id = ?)
					ORDER BY created_at DESC LIMIT 1
				) as last_message,
				(
					SELECT created_at FROM conversations 
					WHERE (student_persona_id = cm.student_persona_id AND teacher_persona_id = ?)
					ORDER BY created_at DESC LIMIT 1
				) as last_message_time
			FROM class_members cm
			JOIN personas p ON cm.student_persona_id = p.id
			WHERE cm.class_id = ?
			ORDER BY last_message_time DESC NULLS LAST
			LIMIT 5`,
			personaID, personaID, classItem.ClassID)
		if err != nil {
			continue
		}

		var students []TeacherChatStudent
		for studentRows.Next() {
			var s TeacherChatStudent
			var lastMsgTime *string
			err := studentRows.Scan(
				&s.StudentPersonaID, &s.StudentNickname, &s.StudentAvatar,
				&s.LastMessage, &lastMsgTime,
			)
			if err != nil {
				continue
			}
			s.LastMessageTime = lastMsgTime
			s.IsPinned = pinnedStudentMap[s.StudentPersonaID]
			students = append(students, s)
		}
		studentRows.Close()

		classItem.Students = students
		classes = append(classes, classItem)
	}

	Success(c, GetTeacherChatListResponse{
		Classes: classes,
		Total:   len(classes),
	})
}

// GetStudentTeacherListResponse 学生端老师列表响应
type GetStudentTeacherListResponse struct {
	Teachers []StudentTeacherChatItem `json:"teachers"`
	Total    int                      `json:"total"`
}

// StudentTeacherChatItem 学生端老师聊天列表项
type StudentTeacherChatItem struct {
	TeacherPersonaID int64   `json:"teacher_persona_id"`
	TeacherNickname  string  `json:"teacher_nickname"`
	TeacherAvatar    string  `json:"teacher_avatar"`
	TeacherSchool    string  `json:"teacher_school,omitempty"`
	Subject          string  `json:"subject,omitempty"`
	LastMessage      string  `json:"last_message,omitempty"`
	LastMessageTime  *string `json:"last_message_time,omitempty"`
	UnreadCount      int     `json:"unread_count"`
	IsPinned         bool    `json:"is_pinned"`
}

// HandleGetStudentTeacherList 获取学生端老师聊天列表
func (h *Handler) HandleGetStudentTeacherList(c *gin.Context) {
	userID, _ := c.Get("user_id")
	studentID := userID.(int64)

	personaID, err := h.getDefaultPersonaID(studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "获取分身失败"})
		return
	}

	// 获取置顶记录
	pinRepo := database.NewChatPinRepository(h.db)
	pinnedTeachers, _ := pinRepo.GetPinnedTargets(studentID, "student", "teacher", personaID)
	pinnedTeacherMap := make(map[int64]bool)
	for _, id := range pinnedTeachers {
		pinnedTeacherMap[id] = true
	}

	// 查询学生的老师列表（通过班级关系）
	rows, err := h.db.DB.Query(`
		SELECT DISTINCT
			p.id as teacher_persona_id,
			p.nickname as teacher_nickname,
			p.avatar as teacher_avatar,
			p.school as teacher_school,
			c.subject,
			(
				SELECT content FROM conversations 
				WHERE (student_persona_id = ? AND teacher_persona_id = p.id)
				ORDER BY created_at DESC LIMIT 1
			) as last_message,
			(
				SELECT created_at FROM conversations 
				WHERE (student_persona_id = ? AND teacher_persona_id = p.id)
				ORDER BY created_at DESC LIMIT 1
			) as last_message_time
		FROM class_members cm
		JOIN classes c ON cm.class_id = c.id
		JOIN personas p ON c.persona_id = p.id
		WHERE cm.student_persona_id = ?
		ORDER BY last_message_time DESC NULLS LAST`,
		personaID, personaID, personaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "查询失败"})
		return
	}
	defer rows.Close()

	var teachers []StudentTeacherChatItem
	for rows.Next() {
		var t StudentTeacherChatItem
		var lastMsgTime *string
		err := rows.Scan(
			&t.TeacherPersonaID, &t.TeacherNickname, &t.TeacherAvatar,
			&t.TeacherSchool, &t.Subject, &t.LastMessage, &lastMsgTime,
		)
		if err != nil {
			continue
		}
		t.LastMessageTime = lastMsgTime
		t.IsPinned = pinnedTeacherMap[t.TeacherPersonaID]
		teachers = append(teachers, t)
	}

	Success(c, GetStudentTeacherListResponse{
		Teachers: teachers,
		Total:    len(teachers),
	})
}

// PinChatRequest 置顶聊天请求
type PinChatRequest struct {
	TargetType string `json:"target_type" binding:"required"` // teacher / student / class
	TargetID   int64  `json:"target_id" binding:"required"`
}

// HandlePinChat 置顶聊天
func (h *Handler) HandlePinChat(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(int64)

	userRole, _ := c.Get("role")
	role := userRole.(string)

	var req PinChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "请求参数错误"})
		return
	}

	personaID, err := h.getDefaultPersonaID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "获取分身失败"})
		return
	}

	pinRepo := database.NewChatPinRepository(h.db)
	pin := &database.ChatPin{
		UserID:     uid,
		UserRole:   role,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		PersonaID:  personaID,
	}

	pinID, err := pinRepo.CreateChatPin(pin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "置顶失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pin_id":  pinID,
		"success": true,
	})
}

// HandleUnpinChat 取消置顶
func (h *Handler) HandleUnpinChat(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(int64)

	userRole, _ := c.Get("role")
	role := userRole.(string)

	targetType := c.Param("type")
	targetIDStr := c.Param("id")
	targetID, err := strconv.ParseInt(targetIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "无效的ID"})
		return
	}

	personaID, err := h.getDefaultPersonaID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "获取分身失败"})
		return
	}

	pinRepo := database.NewChatPinRepository(h.db)
	if err := pinRepo.DeleteChatPin(uid, role, targetType, targetID, personaID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "取消置顶失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetPinnedChatsResponse 获取置顶列表响应
type GetPinnedChatsResponse struct {
	Pins  []PinnedChatItem `json:"pins"`
	Total int              `json:"total"`
}

// PinnedChatItem 置顶项
type PinnedChatItem struct {
	ID         int64  `json:"id"`
	TargetType string `json:"target_type"`
	TargetID   int64  `json:"target_id"`
	TargetName string `json:"target_name"`
	Avatar     string `json:"avatar,omitempty"`
	PinnedAt   string `json:"pinned_at"`
}

// HandleGetPinnedChats 获取置顶列表
func (h *Handler) HandleGetPinnedChats(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(int64)

	userRole, _ := c.Get("role")
	role := userRole.(string)

	personaID, err := h.getDefaultPersonaID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "获取分身失败"})
		return
	}

	pinRepo := database.NewChatPinRepository(h.db)
	pins, err := pinRepo.GetChatPinsByUser(uid, role, personaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "查询失败"})
		return
	}

	var result []PinnedChatItem
	for _, pin := range pins {
		item := PinnedChatItem{
			ID:         pin.ID,
			TargetType: pin.TargetType,
			TargetID:   pin.TargetID,
			PinnedAt:   pin.PinnedAt.Format("2006-01-02 15:04:05"),
		}

		// 根据类型查询名称
		switch pin.TargetType {
		case "class":
			var name string
			h.db.DB.QueryRow(`SELECT name FROM classes WHERE id = ?`, pin.TargetID).Scan(&name)
			item.TargetName = name
		case "student":
			var name, avatar string
			h.db.DB.QueryRow(`SELECT nickname, avatar FROM personas WHERE id = ?`, pin.TargetID).Scan(&name, &avatar)
			item.TargetName = name
			item.Avatar = avatar
		case "teacher":
			var name, avatar string
			h.db.DB.QueryRow(`SELECT nickname, avatar FROM personas WHERE id = ?`, pin.TargetID).Scan(&name, &avatar)
			item.TargetName = name
			item.Avatar = avatar
		}

		result = append(result, item)
	}

	c.JSON(http.StatusOK, GetPinnedChatsResponse{
		Pins:  result,
		Total: len(result),
	})
}

// NewSessionRequest 开启新会话请求
type NewSessionRequest struct {
	TeacherPersonaID int64  `json:"teacher_persona_id" binding:"required"`
	InitialMessage   string `json:"initial_message"`
}

// NewSessionResponse 开启新会话响应
type NewSessionResponse struct {
	SessionID string `json:"session_id"`
	CreatedAt string `json:"created_at"`
}

// HandleNewSession 开启新会话
func (h *Handler) HandleNewSession(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(int64)

	var req NewSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "请求参数错误"})
		return
	}

	// 获取当前用户的学生分身ID
	var studentPersonaID int64
	err := h.db.DB.QueryRow(`
		SELECT id FROM personas WHERE user_id = ? AND role = 'student' LIMIT 1`, uid).Scan(&studentPersonaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "获取学生分身失败"})
		return
	}

	// V2.0 迭代9 M5：异步为上一个活跃会话生成标题
	go h.generateTitleForLastSession(studentPersonaID, req.TeacherPersonaID)

	// 生成新的 session_id
	sessionID := fmt.Sprintf("sess_%s", uuid.New().String()[:12])
	createdAt := time.Now().Format("2006-01-02T15:04:05Z")

	c.JSON(http.StatusOK, NewSessionResponse{
		SessionID: sessionID,
		CreatedAt: createdAt,
	})
}

// generateTitleForLastSession 为上一个活跃会话生成标题
func (h *Handler) generateTitleForLastSession(studentPersonaID, teacherPersonaID int64) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[标题生成] generateTitleForLastSession panic recovered: %v\n", r)
		}
	}()

	db := h.manager.GetDB()
	if db == nil {
		return
	}

	// 查询上一个活跃会话（该教师分身与学生分身之间的最新会话）
	convRepo := database.NewConversationRepository(db)
	var lastSessionID string
	var queryErr error

	if teacherPersonaID > 0 && studentPersonaID > 0 {
		// 分身维度查询最新会话
		lastSessionID, queryErr = convRepo.GetLatestSessionByPersonas(teacherPersonaID, studentPersonaID)
	} else if studentPersonaID > 0 {
		// 仅学生分身维度
		err := db.QueryRow(
			`SELECT session_id FROM conversations WHERE student_persona_id = ? ORDER BY created_at DESC LIMIT 1`,
			studentPersonaID,
		).Scan(&lastSessionID)
		if err != nil {
			queryErr = err
		}
	}

	if queryErr != nil || lastSessionID == "" {
		return
	}

	// 调用标题生成逻辑
	h.generateSessionTitle(lastSessionID, studentPersonaID)
}

// QuickActionItem 快捷指令项
type QuickActionItem struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Action string `json:"action"`
}

// HandleGetQuickActions 获取快捷指令
func (h *Handler) HandleGetQuickActions(c *gin.Context) {
	teacherPersonaIDStr := c.Query("teacher_persona_id")
	if teacherPersonaIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "teacher_persona_id 为必填参数"})
		return
	}

	_, err := strconv.ParseInt(teacherPersonaIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "无效的 teacher_persona_id"})
		return
	}

	// 返回预设的快捷指令列表（硬编码，后续可配置化）
	actions := []QuickActionItem{
		{ID: "review", Label: "📚 回顾上次内容", Action: "回顾上次学习的内容"},
		{ID: "summarize", Label: "📝 总结已学知识", Action: "帮我总结一下已学的知识点"},
		{ID: "practice", Label: "✏️ 开始练习", Action: "我想开始练习"},
		{ID: "question", Label: "❓ 提个问题", Action: "我有一个问题"},
	}

	c.JSON(http.StatusOK, actions)
}
