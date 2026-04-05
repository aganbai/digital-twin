package database

import (
	"database/sql"
	"fmt"
	"time"
)

// ==================== ConversationRepository ====================

// ConversationRepository 对话历史数据访问层
type ConversationRepository struct {
	db *sql.DB
}

// NewConversationRepository 创建对话仓库
func NewConversationRepository(db *sql.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

// Create 创建对话记录
func (r *ConversationRepository) Create(conv *Conversation) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO conversations (student_id, teacher_id, session_id, role, content, sender_type, token_count, created_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		conv.StudentID, conv.TeacherID, conv.SessionID, conv.Role, conv.Content,
		conv.SenderType, conv.TokenCount, time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("创建对话记录失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取对话记录ID失败: %w", err)
	}

	return id, nil
}

// GetByStudentAndTeacher 根据学生ID和教师ID查询对话历史（支持分页）
func (r *ConversationRepository) GetByStudentAndTeacher(studentID, teacherID int64, offset, limit int) ([]*Conversation, int, error) {
	// 查询总数
	var total int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM conversations WHERE student_id = ? AND teacher_id = ?`,
		studentID, teacherID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询对话总数失败: %w", err)
	}

	// 查询列表
	rows, err := r.db.Query(
		`SELECT id, student_id, teacher_id, session_id, role, content, token_count, COALESCE(sender_type, '') as sender_type, COALESCE(reply_to_id, 0) as reply_to_id, created_at 
		 FROM conversations WHERE student_id = ? AND teacher_id = ? 
		 ORDER BY created_at ASC LIMIT ? OFFSET ?`,
		studentID, teacherID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("查询对话列表失败: %w", err)
	}
	defer rows.Close()

	var convs []*Conversation
	for rows.Next() {
		conv := &Conversation{}
		if err := rows.Scan(&conv.ID, &conv.StudentID, &conv.TeacherID, &conv.SessionID,
			&conv.Role, &conv.Content, &conv.TokenCount, &conv.SenderType, &conv.ReplyToID, &conv.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描对话记录失败: %w", err)
		}
		convs = append(convs, conv)
	}

	return convs, total, nil
}

// GetConversationsByStudent 查询学生与所有教师的对话（不传 teacher_id 时使用）
func (r *ConversationRepository) GetConversationsByStudent(studentID int64, offset, limit int) ([]*Conversation, int, error) {
	// 查询总数
	var total int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM conversations WHERE student_id = ?`,
		studentID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询对话总数失败: %w", err)
	}

	// 查询列表
	rows, err := r.db.Query(
		`SELECT id, student_id, teacher_id, session_id, role, content, token_count, COALESCE(sender_type, '') as sender_type, COALESCE(reply_to_id, 0) as reply_to_id, created_at 
		 FROM conversations WHERE student_id = ? 
		 ORDER BY created_at ASC LIMIT ? OFFSET ?`,
		studentID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("查询对话列表失败: %w", err)
	}
	defer rows.Close()

	var convs []*Conversation
	for rows.Next() {
		conv := &Conversation{}
		if err := rows.Scan(&conv.ID, &conv.StudentID, &conv.TeacherID, &conv.SessionID,
			&conv.Role, &conv.Content, &conv.TokenCount, &conv.SenderType, &conv.ReplyToID, &conv.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描对话记录失败: %w", err)
		}
		convs = append(convs, conv)
	}

	return convs, total, nil
}

// GetConversationsBySession 按 session_id 筛选对话
func (r *ConversationRepository) GetConversationsBySession(studentID int64, sessionID string, offset, limit int) ([]*Conversation, int, error) {
	// 查询总数
	var total int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM conversations WHERE student_id = ? AND session_id = ?`,
		studentID, sessionID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询对话总数失败: %w", err)
	}

	// 查询列表
	rows, err := r.db.Query(
		`SELECT id, student_id, teacher_id, session_id, role, content, token_count, COALESCE(sender_type, '') as sender_type, COALESCE(reply_to_id, 0) as reply_to_id, created_at 
		 FROM conversations WHERE student_id = ? AND session_id = ? 
		 ORDER BY created_at ASC LIMIT ? OFFSET ?`,
		studentID, sessionID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("查询对话列表失败: %w", err)
	}
	defer rows.Close()

	var convs []*Conversation
	for rows.Next() {
		conv := &Conversation{}
		if err := rows.Scan(&conv.ID, &conv.StudentID, &conv.TeacherID, &conv.SessionID,
			&conv.Role, &conv.Content, &conv.TokenCount, &conv.SenderType, &conv.ReplyToID, &conv.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描对话记录失败: %w", err)
		}
		convs = append(convs, conv)
	}

	return convs, total, nil
}

// GetSessionsByStudent 获取学生的会话列表（按 session_id 分组，返回每个会话的摘要）
func (r *ConversationRepository) GetSessionsByStudent(studentID int64, offset, limit int) ([]SessionSummary, int, error) {
	// 查询总数（不同 session_id 的数量）
	var total int
	err := r.db.QueryRow(
		`SELECT COUNT(DISTINCT session_id) FROM conversations WHERE student_id = ?`,
		studentID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询会话总数失败: %w", err)
	}

	// 查询每个 session 的最新一条消息作为摘要
	rows, err := r.db.Query(
		`SELECT 
			c.session_id,
			c.teacher_id,
			u.nickname AS teacher_nickname,
			c.content AS last_message,
			c.role AS last_message_role,
			(SELECT COUNT(*) FROM conversations WHERE session_id = c.session_id) AS message_count,
			c.created_at AS updated_at
		FROM conversations c
		JOIN users u ON c.teacher_id = u.id
		WHERE c.student_id = ? 
			AND c.id IN (
				SELECT MAX(id) FROM conversations 
				WHERE student_id = ? 
				GROUP BY session_id
			)
		ORDER BY c.created_at DESC
		LIMIT ? OFFSET ?`,
		studentID, studentID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("查询会话列表失败: %w", err)
	}
	defer rows.Close()

	var sessions []SessionSummary
	for rows.Next() {
		s := SessionSummary{}
		if err := rows.Scan(&s.SessionID, &s.TeacherID, &s.TeacherNickname,
			&s.LastMessage, &s.LastMessageRole, &s.MessageCount, &s.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描会话摘要失败: %w", err)
		}
		// 截断 last_message 至 100 个字符（注意 UTF-8 字符边界）
		runes := []rune(s.LastMessage)
		if len(runes) > 100 {
			s.LastMessage = string(runes[:100]) + "..."
		}
		sessions = append(sessions, s)
	}

	return sessions, total, nil
}

// ==================== 分身维度方法 ====================

// CreateWithPersonas 创建带分身维度的对话记录
func (r *ConversationRepository) CreateWithPersonas(conv *Conversation) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO conversations (student_id, teacher_id, teacher_persona_id, student_persona_id, session_id, role, content, sender_type, reply_to_id, token_count, created_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		conv.StudentID, conv.TeacherID, conv.TeacherPersonaID, conv.StudentPersonaID,
		conv.SessionID, conv.Role, conv.Content, conv.SenderType, conv.ReplyToID, conv.TokenCount, time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("创建对话记录失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取对话记录ID失败: %w", err)
	}

	return id, nil
}

// GetByPersonas 按分身维度查询对话
func (r *ConversationRepository) GetByPersonas(teacherPersonaID, studentPersonaID int64, sessionID string, offset, limit int) ([]*Conversation, int, error) {
	// 查询总数
	countQuery := `SELECT COUNT(*) FROM conversations WHERE teacher_persona_id = ? AND student_persona_id = ?`
	args := []interface{}{teacherPersonaID, studentPersonaID}
	if sessionID != "" {
		countQuery += ` AND session_id = ?`
		args = append(args, sessionID)
	}

	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询对话总数失败: %w", err)
	}

	// 查询列表
	listQuery := `SELECT id, student_id, teacher_id, COALESCE(teacher_persona_id, 0), COALESCE(student_persona_id, 0), session_id, role, content, token_count, COALESCE(sender_type, '') as sender_type, COALESCE(reply_to_id, 0) as reply_to_id, created_at 
		 FROM conversations WHERE teacher_persona_id = ? AND student_persona_id = ?`
	listArgs := []interface{}{teacherPersonaID, studentPersonaID}
	if sessionID != "" {
		listQuery += ` AND session_id = ?`
		listArgs = append(listArgs, sessionID)
	}
	listQuery += ` ORDER BY created_at ASC LIMIT ? OFFSET ?`
	listArgs = append(listArgs, limit, offset)

	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询对话列表失败: %w", err)
	}
	defer rows.Close()

	var convs []*Conversation
	for rows.Next() {
		conv := &Conversation{}
		if err := rows.Scan(&conv.ID, &conv.StudentID, &conv.TeacherID, &conv.TeacherPersonaID, &conv.StudentPersonaID,
			&conv.SessionID, &conv.Role, &conv.Content, &conv.TokenCount, &conv.SenderType, &conv.ReplyToID, &conv.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描对话记录失败: %w", err)
		}
		convs = append(convs, conv)
	}

	return convs, total, nil
}

// GetSessionsByPersona 按学生分身维度查询会话列表
func (r *ConversationRepository) GetSessionsByPersona(studentPersonaID int64, offset, limit int) ([]SessionSummary, int, error) {
	// 查询总数
	var total int
	err := r.db.QueryRow(
		`SELECT COUNT(DISTINCT session_id) FROM conversations WHERE student_persona_id = ?`,
		studentPersonaID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询会话总数失败: %w", err)
	}

	rows, err := r.db.Query(
		`SELECT 
			c.session_id,
			c.teacher_id,
			COALESCE(u.nickname, '') AS teacher_nickname,
			c.content AS last_message,
			c.role AS last_message_role,
			(SELECT COUNT(*) FROM conversations WHERE session_id = c.session_id) AS message_count,
			c.created_at AS updated_at
		FROM conversations c
		JOIN users u ON c.teacher_id = u.id
		WHERE c.student_persona_id = ? 
			AND c.id IN (
				SELECT MAX(id) FROM conversations 
				WHERE student_persona_id = ? 
				GROUP BY session_id
			)
		ORDER BY c.created_at DESC
		LIMIT ? OFFSET ?`,
		studentPersonaID, studentPersonaID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("查询会话列表失败: %w", err)
	}
	defer rows.Close()

	var sessions []SessionSummary
	for rows.Next() {
		s := SessionSummary{}
		if err := rows.Scan(&s.SessionID, &s.TeacherID, &s.TeacherNickname,
			&s.LastMessage, &s.LastMessageRole, &s.MessageCount, &s.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描会话摘要失败: %w", err)
		}
		runes := []rune(s.LastMessage)
		if len(runes) > 100 {
			s.LastMessage = string(runes[:100]) + "..."
		}
		sessions = append(sessions, s)
	}

	return sessions, total, nil
}

// ======================== V2.0 迭代4 新增方法 ========================

// CreateWithSenderType 创建带 sender_type 的对话记录
func (r *ConversationRepository) CreateWithSenderType(conv *Conversation) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO conversations (student_id, teacher_id, teacher_persona_id, student_persona_id, session_id, role, content, token_count, sender_type, reply_to_id, created_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		conv.StudentID, conv.TeacherID, conv.TeacherPersonaID, conv.StudentPersonaID,
		conv.SessionID, conv.Role, conv.Content, conv.TokenCount,
		conv.SenderType, conv.ReplyToID, time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("创建对话记录失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取对话记录ID失败: %w", err)
	}

	return id, nil
}

// GetByIDSimple 根据ID查询单条对话记录（用于引用回复时获取原消息）
func (r *ConversationRepository) GetByIDSimple(id int64) (*Conversation, error) {
	conv := &Conversation{}
	err := r.db.QueryRow(
		`SELECT id, student_id, teacher_id, COALESCE(teacher_persona_id, 0), COALESCE(student_persona_id, 0), session_id, role, content, token_count, COALESCE(sender_type, ''), COALESCE(reply_to_id, 0), created_at 
		 FROM conversations WHERE id = ?`,
		id,
	).Scan(&conv.ID, &conv.StudentID, &conv.TeacherID, &conv.TeacherPersonaID, &conv.StudentPersonaID,
		&conv.SessionID, &conv.Role, &conv.Content, &conv.TokenCount,
		&conv.SenderType, &conv.ReplyToID, &conv.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询对话记录失败: %w", err)
	}

	return conv, nil
}

// GetByTeacherAndStudentPersonas 教师查看某学生的对话记录（按教师分身和学生分身维度）
func (r *ConversationRepository) GetByTeacherAndStudentPersonas(teacherPersonaID, studentPersonaID int64, sessionID string, offset, limit int) ([]*Conversation, int, error) {
	// 构建查询条件
	where := `teacher_persona_id = ? AND student_persona_id = ?`
	args := []interface{}{teacherPersonaID, studentPersonaID}
	if sessionID != "" {
		where += ` AND session_id = ?`
		args = append(args, sessionID)
	}

	// 查询总数
	var total int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM conversations WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询对话总数失败: %w", err)
	}

	// 查询列表
	listQuery := `SELECT id, student_id, teacher_id, COALESCE(teacher_persona_id, 0), COALESCE(student_persona_id, 0), session_id, role, content, token_count, COALESCE(sender_type, ''), COALESCE(reply_to_id, 0), created_at 
		 FROM conversations WHERE ` + where + ` ORDER BY created_at ASC LIMIT ? OFFSET ?`
	listArgs := append(args, limit, offset)

	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询对话列表失败: %w", err)
	}
	defer rows.Close()

	var convs []*Conversation
	for rows.Next() {
		conv := &Conversation{}
		if err := rows.Scan(&conv.ID, &conv.StudentID, &conv.TeacherID, &conv.TeacherPersonaID, &conv.StudentPersonaID,
			&conv.SessionID, &conv.Role, &conv.Content, &conv.TokenCount,
			&conv.SenderType, &conv.ReplyToID, &conv.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描对话记录失败: %w", err)
		}

		// 如果有引用回复，查询引用的消息内容
		if conv.ReplyToID > 0 {
			var replyContent string
			r.db.QueryRow(`SELECT content FROM conversations WHERE id = ?`, conv.ReplyToID).Scan(&replyContent)
			runes := []rune(replyContent)
			if len(runes) > 100 {
				replyContent = string(runes[:100]) + "..."
			}
			conv.ReplyToContent = replyContent
		}

		convs = append(convs, conv)
	}

	return convs, total, nil
}

// GetLatestSessionByPersonas 获取教师和学生分身之间的最新会话ID
func (r *ConversationRepository) GetLatestSessionByPersonas(teacherPersonaID, studentPersonaID int64) (string, error) {
	var sessionID string
	err := r.db.QueryRow(
		`SELECT session_id FROM conversations WHERE teacher_persona_id = ? AND student_persona_id = ? ORDER BY created_at DESC LIMIT 1`,
		teacherPersonaID, studentPersonaID,
	).Scan(&sessionID)

	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("查询最新会话ID失败: %w", err)
	}

	return sessionID, nil
}

// ======================== V2.0 迭代9 API-104 新增方法 ========================

// GetSessionsByTeacherPersona 按教师分身维度查询会话列表（支持按 teacher_persona_id 过滤）
// 返回带有标题、消息数、活跃状态的会话列表
func (r *ConversationRepository) GetSessionsByTeacherPersona(studentPersonaID, teacherPersonaID int64, offset, limit int) ([]SessionItem, int, error) {
	// 查询总数
	var total int
	countQuery := `SELECT COUNT(DISTINCT session_id) FROM conversations WHERE student_persona_id = ?`
	countArgs := []interface{}{studentPersonaID}
	if teacherPersonaID > 0 {
		countQuery += ` AND teacher_persona_id = ?`
		countArgs = append(countArgs, teacherPersonaID)
	}

	if err := r.db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询会话总数失败: %w", err)
	}

	if total == 0 {
		return []SessionItem{}, 0, nil
	}

	// 第一步：获取所有唯一的 session_id，按最新消息时间倒序
	sessionQuery := `
		SELECT session_id, MAX(created_at) as last_time
		FROM conversations
		WHERE student_persona_id = ?`
	sessionArgs := []interface{}{studentPersonaID}
	if teacherPersonaID > 0 {
		sessionQuery += ` AND teacher_persona_id = ?`
		sessionArgs = append(sessionArgs, teacherPersonaID)
	}
	sessionQuery += ` GROUP BY session_id ORDER BY last_time DESC LIMIT ? OFFSET ?`
	sessionArgs = append(sessionArgs, limit, offset)

	sessionRows, err := r.db.Query(sessionQuery, sessionArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询会话ID列表失败: %w", err)
	}
	defer sessionRows.Close()

	type sessionMeta struct {
		sessionID string
		updatedAt time.Time
	}
	var sessionMetas []sessionMeta
	for sessionRows.Next() {
		var sm sessionMeta
		var timeStr string
		if err := sessionRows.Scan(&sm.sessionID, &timeStr); err != nil {
			return nil, 0, fmt.Errorf("扫描会话元数据失败: %w", err)
		}
		// 解析时间字符串
		sm.updatedAt, _ = time.Parse("2006-01-02 15:04:05", timeStr)
		sessionMetas = append(sessionMetas, sm)
	}

	if len(sessionMetas) == 0 {
		return []SessionItem{}, total, nil
	}

	// 第二步：为每个会话获取详细信息
	var sessions []SessionItem
	for _, sm := range sessionMetas {
		// 获取最新消息
		var lastMsg, lastRole string
		var tpID int64
		err := r.db.QueryRow(`
			SELECT content, role, teacher_persona_id 
			FROM conversations 
			WHERE session_id = ? 
			ORDER BY created_at DESC LIMIT 1`, sm.sessionID).Scan(&lastMsg, &lastRole, &tpID)
		if err != nil {
			continue
		}

		// 获取消息数
		var msgCount int
		r.db.QueryRow(`SELECT COUNT(*) FROM conversations WHERE session_id = ?`, sm.sessionID).Scan(&msgCount)

		// 获取标题
		var title sql.NullString
		r.db.QueryRow(`SELECT title FROM session_titles WHERE session_id = ?`, sm.sessionID).Scan(&title)

		// 截断消息
		runes := []rune(lastMsg)
		if len(runes) > 100 {
			lastMsg = string(runes[:100]) + "..."
		}

		item := SessionItem{
			SessionID:       sm.sessionID,
			Title:           nil,
			LastMessage:     lastMsg,
			LastMessageRole: lastRole,
			MessageCount:    msgCount,
			UpdatedAt:       sm.updatedAt,
		}
		if title.Valid {
			item.Title = &title.String
		}
		sessions = append(sessions, item)
	}

	// 第三步：判断活跃会话：每个老师下最新更新的会话为活跃会话
	if len(sessions) > 0 {
		// 查询每个 teacher_persona_id 的最新会话
		activeQuery := `
			SELECT session_id 
			FROM conversations 
			WHERE id IN (
				SELECT MAX(id) FROM conversations 
				WHERE student_persona_id = ?`
		activeArgs := []interface{}{studentPersonaID}
		if teacherPersonaID > 0 {
			activeQuery += ` AND teacher_persona_id = ?`
			activeArgs = append(activeArgs, teacherPersonaID)
		}
		activeQuery += ` GROUP BY teacher_persona_id)`

		activeRows, err := r.db.Query(activeQuery, activeArgs...)
		if err == nil {
			defer activeRows.Close()
			activeSessions := make(map[string]bool)
			for activeRows.Next() {
				var sid string
				if err := activeRows.Scan(&sid); err == nil {
					activeSessions[sid] = true
				}
			}
			// 标记活跃会话
			for i := range sessions {
				sessions[i].IsActive = activeSessions[sessions[i].SessionID]
			}
		}
	}

	return sessions, total, nil
}

// ======================== V2.0 迭代11 M4 自测学生方法 ========================

// DeleteByStudentPersonaID 删除学生分身的所有对话记录
func (r *ConversationRepository) DeleteByStudentPersonaID(studentPersonaID int64) (int, error) {
	result, err := r.db.Exec(
		`DELETE FROM conversations WHERE student_persona_id = ?`,
		studentPersonaID,
	)
	if err != nil {
		return 0, fmt.Errorf("删除学生对话失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("获取影响行数失败: %w", err)
	}

	return int(affected), nil
}
