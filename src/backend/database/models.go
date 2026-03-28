package database

import "time"

// User 用户模型
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // JSON 序列化时隐藏密码
	Role      string    `json:"role"`
	Nickname  string    `json:"nickname,omitempty"`
	Email     string    `json:"email,omitempty"`
	OpenID    string    `json:"openid,omitempty"` // 微信 openid
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Document 知识文档模型
type Document struct {
	ID        int64     `json:"id"`
	TeacherID int64     `json:"teacher_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	DocType   string    `json:"doc_type"`
	Tags      string    `json:"tags"` // JSON 数组格式存储
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Conversation 对话历史模型
type Conversation struct {
	ID         int64     `json:"id"`
	StudentID  int64     `json:"student_id"`
	TeacherID  int64     `json:"teacher_id"`
	SessionID  string    `json:"session_id"`
	Role       string    `json:"role"`
	Content    string    `json:"content"`
	TokenCount int       `json:"token_count"`
	CreatedAt  time.Time `json:"created_at"`
}

// Memory 学生记忆模型
type Memory struct {
	ID           int64      `json:"id"`
	StudentID    int64      `json:"student_id"`
	TeacherID    int64      `json:"teacher_id"`
	MemoryType   string     `json:"memory_type"`
	Content      string     `json:"content"`
	Importance   float64    `json:"importance"`
	LastAccessed *time.Time `json:"last_accessed,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// TeacherWithDocCount 教师信息（含文档数量）
type TeacherWithDocCount struct {
	ID            int64     `json:"id"`
	Username      string    `json:"username"`
	Nickname      string    `json:"nickname"`
	Role          string    `json:"role"`
	DocumentCount int       `json:"document_count"`
	CreatedAt     time.Time `json:"created_at"`
}

// SessionSummary 会话摘要（会话列表用）
type SessionSummary struct {
	SessionID       string    `json:"session_id"`
	TeacherID       int64     `json:"teacher_id"`
	TeacherNickname string    `json:"teacher_nickname"`
	LastMessage     string    `json:"last_message"`
	LastMessageRole string    `json:"last_message_role"`
	MessageCount    int       `json:"message_count"`
	UpdatedAt       time.Time `json:"updated_at"`
}
