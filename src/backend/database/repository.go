package database

import (
	"database/sql"
	"fmt"
	"time"
)

// ==================== UserRepository ====================

// UserRepository 用户数据访问层
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(user *User) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO users (username, password, role, nickname, email, openid, created_at, updated_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		user.Username, user.Password, user.Role, user.Nickname, user.Email, user.OpenID,
		time.Now(), time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("创建用户失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取用户ID失败: %w", err)
	}

	return id, nil
}

// GetByOpenID 根据微信 openid 查询用户
func (r *UserRepository) GetByOpenID(openid string) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(
		`SELECT id, username, password, role, nickname, email, openid, created_at, updated_at 
		 FROM users WHERE openid = ?`,
		openid,
	).Scan(&user.ID, &user.Username, &user.Password, &user.Role,
		&user.Nickname, &user.Email, &user.OpenID, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("根据 openid 查询用户失败: %w", err)
	}

	return user, nil
}

// CreateWithOpenID 创建带 openid 的用户（微信登录用）
func (r *UserRepository) CreateWithOpenID(user *User) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO users (username, password, role, nickname, email, openid, created_at, updated_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		user.Username, user.Password, user.Role, user.Nickname, user.Email, user.OpenID,
		time.Now(), time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("创建微信用户失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取用户ID失败: %w", err)
	}

	return id, nil
}

// UpdateRoleAndNickname 更新用户角色和昵称（补全信息用）
func (r *UserRepository) UpdateRoleAndNickname(userID int64, role, nickname string) error {
	result, err := r.db.Exec(
		`UPDATE users SET role = ?, nickname = ?, updated_at = ? WHERE id = ?`,
		role, nickname, time.Now(), userID,
	)
	if err != nil {
		return fmt.Errorf("更新用户角色和昵称失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("用户不存在: id=%d", userID)
	}

	return nil
}

// GetByUsername 根据用户名查询用户
func (r *UserRepository) GetByUsername(username string) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(
		`SELECT id, username, password, role, nickname, email, openid, created_at, updated_at 
		 FROM users WHERE username = ?`,
		username,
	).Scan(&user.ID, &user.Username, &user.Password, &user.Role,
		&user.Nickname, &user.Email, &user.OpenID, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return user, nil
}

// GetByID 根据ID查询用户
func (r *UserRepository) GetByID(id int64) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(
		`SELECT id, username, password, role, nickname, email, openid, created_at, updated_at 
		 FROM users WHERE id = ?`,
		id,
	).Scan(&user.ID, &user.Username, &user.Password, &user.Role,
		&user.Nickname, &user.Email, &user.OpenID, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return user, nil
}

// ==================== DocumentRepository ====================

// DocumentRepository 文档数据访问层
type DocumentRepository struct {
	db *sql.DB
}

// NewDocumentRepository 创建文档仓库
func NewDocumentRepository(db *sql.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// Create 创建文档
func (r *DocumentRepository) Create(doc *Document) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO documents (teacher_id, title, content, doc_type, tags, status, created_at, updated_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		doc.TeacherID, doc.Title, doc.Content, doc.DocType, doc.Tags, doc.Status,
		time.Now(), time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("创建文档失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取文档ID失败: %w", err)
	}

	return id, nil
}

// GetByTeacherID 根据教师ID查询文档列表
func (r *DocumentRepository) GetByTeacherID(teacherID int64, offset, limit int) ([]*Document, int, error) {
	// 查询总数
	var total int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM documents WHERE teacher_id = ? AND status = 'active'`,
		teacherID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询文档总数失败: %w", err)
	}

	// 查询列表
	rows, err := r.db.Query(
		`SELECT id, teacher_id, title, content, doc_type, tags, status, created_at, updated_at 
		 FROM documents WHERE teacher_id = ? AND status = 'active' 
		 ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		teacherID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("查询文档列表失败: %w", err)
	}
	defer rows.Close()

	var docs []*Document
	for rows.Next() {
		doc := &Document{}
		if err := rows.Scan(&doc.ID, &doc.TeacherID, &doc.Title, &doc.Content,
			&doc.DocType, &doc.Tags, &doc.Status, &doc.CreatedAt, &doc.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描文档记录失败: %w", err)
		}
		docs = append(docs, doc)
	}

	return docs, total, nil
}

// GetByID 根据ID查询文档
func (r *DocumentRepository) GetByID(id int64) (*Document, error) {
	doc := &Document{}
	err := r.db.QueryRow(
		`SELECT id, teacher_id, title, content, doc_type, tags, status, created_at, updated_at 
		 FROM documents WHERE id = ?`,
		id,
	).Scan(&doc.ID, &doc.TeacherID, &doc.Title, &doc.Content,
		&doc.DocType, &doc.Tags, &doc.Status, &doc.CreatedAt, &doc.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询文档失败: %w", err)
	}

	return doc, nil
}

// Delete 删除文档（软删除，设置 status 为 archived）
func (r *DocumentRepository) Delete(id int64) error {
	result, err := r.db.Exec(
		`UPDATE documents SET status = 'archived', updated_at = ? WHERE id = ?`,
		time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("删除文档失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("文档不存在: id=%d", id)
	}

	return nil
}

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
		`INSERT INTO conversations (student_id, teacher_id, session_id, role, content, token_count, created_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		conv.StudentID, conv.TeacherID, conv.SessionID, conv.Role, conv.Content,
		conv.TokenCount, time.Now(),
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
		`SELECT id, student_id, teacher_id, session_id, role, content, token_count, created_at 
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
			&conv.Role, &conv.Content, &conv.TokenCount, &conv.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描对话记录失败: %w", err)
		}
		convs = append(convs, conv)
	}

	return convs, total, nil
}

// ==================== MemoryRepository ====================

// MemoryRepository 记忆数据访问层
type MemoryRepository struct {
	db *sql.DB
}

// NewMemoryRepository 创建记忆仓库
func NewMemoryRepository(db *sql.DB) *MemoryRepository {
	return &MemoryRepository{db: db}
}

// Create 创建记忆
func (r *MemoryRepository) Create(mem *Memory) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO memories (student_id, teacher_id, memory_type, content, importance, last_accessed, created_at, updated_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		mem.StudentID, mem.TeacherID, mem.MemoryType, mem.Content,
		mem.Importance, mem.LastAccessed, time.Now(), time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("创建记忆失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取记忆ID失败: %w", err)
	}

	return id, nil
}

// GetByStudentAndTeacher 根据学生ID和教师ID查询记忆列表
func (r *MemoryRepository) GetByStudentAndTeacher(studentID, teacherID int64, limit int) ([]*Memory, error) {
	rows, err := r.db.Query(
		`SELECT id, student_id, teacher_id, memory_type, content, importance, last_accessed, created_at, updated_at 
		 FROM memories WHERE student_id = ? AND teacher_id = ? 
		 ORDER BY importance DESC, created_at DESC LIMIT ?`,
		studentID, teacherID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("查询记忆列表失败: %w", err)
	}
	defer rows.Close()

	var memories []*Memory
	for rows.Next() {
		mem := &Memory{}
		if err := rows.Scan(&mem.ID, &mem.StudentID, &mem.TeacherID, &mem.MemoryType,
			&mem.Content, &mem.Importance, &mem.LastAccessed, &mem.CreatedAt, &mem.UpdatedAt); err != nil {
			return nil, fmt.Errorf("扫描记忆记录失败: %w", err)
		}
		memories = append(memories, mem)
	}

	return memories, nil
}

// UpdateLastAccessed 更新记忆的最后访问时间
func (r *MemoryRepository) UpdateLastAccessed(id int64) error {
	now := time.Now()
	result, err := r.db.Exec(
		`UPDATE memories SET last_accessed = ?, updated_at = ? WHERE id = ?`,
		now, now, id,
	)
	if err != nil {
		return fmt.Errorf("更新记忆访问时间失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("记忆不存在: id=%d", id)
	}

	return nil
}

// ==================== 教师列表查询 ====================

// GetTeachers 查询教师列表，LEFT JOIN documents 统计文档数
func (r *UserRepository) GetTeachers(offset, limit int) ([]TeacherWithDocCount, int, error) {
	// 查询总数
	var total int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM users WHERE role = 'teacher'`,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询教师总数失败: %w", err)
	}

	// 查询列表，LEFT JOIN documents 统计 active 文档数
	rows, err := r.db.Query(
		`SELECT u.id, u.username, u.nickname, u.role, 
		        COALESCE(COUNT(d.id), 0) AS document_count, u.created_at
		 FROM users u
		 LEFT JOIN documents d ON u.id = d.teacher_id AND d.status = 'active'
		 WHERE u.role = 'teacher'
		 GROUP BY u.id
		 ORDER BY u.created_at DESC
		 LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("查询教师列表失败: %w", err)
	}
	defer rows.Close()

	var teachers []TeacherWithDocCount
	for rows.Next() {
		t := TeacherWithDocCount{}
		if err := rows.Scan(&t.ID, &t.Username, &t.Nickname, &t.Role,
			&t.DocumentCount, &t.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描教师记录失败: %w", err)
		}
		teachers = append(teachers, t)
	}

	return teachers, total, nil
}

// ==================== 用户统计信息 ====================

// GetUserStats 获取用户统计信息
// 学生：conversation_count（对话数）+ memory_count（记忆数）
// 教师：document_count（文档数）+ conversation_count（被提问数）
func (r *UserRepository) GetUserStats(userID int64, role string) (map[string]int, error) {
	stats := make(map[string]int)

	if role == "student" {
		// 学生：对话数
		var convCount int
		err := r.db.QueryRow(
			`SELECT COUNT(*) FROM conversations WHERE student_id = ?`, userID,
		).Scan(&convCount)
		if err != nil {
			return nil, fmt.Errorf("查询学生对话数失败: %w", err)
		}
		stats["conversation_count"] = convCount

		// 学生：记忆数
		var memCount int
		err = r.db.QueryRow(
			`SELECT COUNT(*) FROM memories WHERE student_id = ?`, userID,
		).Scan(&memCount)
		if err != nil {
			return nil, fmt.Errorf("查询学生记忆数失败: %w", err)
		}
		stats["memory_count"] = memCount

	} else if role == "teacher" {
		// 教师：active 文档数
		var docCount int
		err := r.db.QueryRow(
			`SELECT COUNT(*) FROM documents WHERE teacher_id = ? AND status = 'active'`, userID,
		).Scan(&docCount)
		if err != nil {
			return nil, fmt.Errorf("查询教师文档数失败: %w", err)
		}
		stats["document_count"] = docCount

		// 教师：被提问数（作为 teacher_id 出现在 conversations 中的次数）
		var convCount int
		err = r.db.QueryRow(
			`SELECT COUNT(*) FROM conversations WHERE teacher_id = ?`, userID,
		).Scan(&convCount)
		if err != nil {
			return nil, fmt.Errorf("查询教师被提问数失败: %w", err)
		}
		stats["conversation_count"] = convCount
	}

	return stats, nil
}

// ==================== 对话历史增强查询 ====================

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
		`SELECT id, student_id, teacher_id, session_id, role, content, token_count, created_at 
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
			&conv.Role, &conv.Content, &conv.TokenCount, &conv.CreatedAt); err != nil {
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
		`SELECT id, student_id, teacher_id, session_id, role, content, token_count, created_at 
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
			&conv.Role, &conv.Content, &conv.TokenCount, &conv.CreatedAt); err != nil {
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
