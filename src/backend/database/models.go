package database

import "time"

// User 用户模型
type User struct {
	ID               int64     `json:"id"`
	Username         string    `json:"username"`
	Password         string    `json:"-"` // JSON 序列化时隐藏密码
	Role             string    `json:"role"`
	Nickname         string    `json:"nickname,omitempty"`
	Email            string    `json:"email,omitempty"`
	OpenID           string    `json:"openid,omitempty"` // 微信 openid
	School           string    `json:"school,omitempty"`
	Description      string    `json:"description,omitempty"`
	DefaultPersonaID int64     `json:"default_persona_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Document 知识文档模型
type Document struct {
	ID              int64     `json:"id"`
	TeacherID       int64     `json:"teacher_id"`
	Title           string    `json:"title"`
	Content         string    `json:"content"`
	DocType         string    `json:"doc_type"`
	Tags            string    `json:"tags"` // JSON 数组格式存储
	Status          string    `json:"status"`
	Scope           string    `json:"scope"`             // global / class / student
	ScopeID         int64     `json:"scope_id"`          // scope=class 时为班级ID，scope=student 时为学生分身ID
	PersonaID       int64     `json:"persona_id"`        // 教师分身ID
	Summary         string    `json:"summary"`           // 文档摘要
	SourceSessionID string    `json:"source_session_id"` // V2.0 迭代6: 聊天记录导入的来源会话ID
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Conversation 对话历史模型
type Conversation struct {
	ID               int64     `json:"id"`
	StudentID        int64     `json:"student_id"`
	TeacherID        int64     `json:"teacher_id"`
	TeacherPersonaID int64     `json:"teacher_persona_id"`
	StudentPersonaID int64     `json:"student_persona_id"`
	SessionID        string    `json:"session_id"`
	Role             string    `json:"role"`
	Content          string    `json:"content"`
	TokenCount       int       `json:"token_count"`
	SenderType       string    `json:"sender_type"`
	ReplyToID        int64     `json:"reply_to_id"`
	ReplyToContent   string    `json:"reply_to_content,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// Memory 学生记忆模型
type Memory struct {
	ID               int64      `json:"id"`
	StudentID        int64      `json:"student_id"`
	TeacherID        int64      `json:"teacher_id"`
	TeacherPersonaID int64      `json:"teacher_persona_id"`
	StudentPersonaID int64      `json:"student_persona_id"`
	MemoryType       string     `json:"memory_type"`
	MemoryLayer      string     `json:"memory_layer"` // V2.0 迭代6: core / episodic / archived
	Content          string     `json:"content"`
	Importance       float64    `json:"importance"`
	LastAccessed     *time.Time `json:"last_accessed,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// TeacherWithDocCount 教师信息（含文档数量）
type TeacherWithDocCount struct {
	ID            int64     `json:"id"`
	Username      string    `json:"username"`
	Nickname      string    `json:"nickname"`
	Role          string    `json:"role"`
	School        string    `json:"school"`
	Description   string    `json:"description"`
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

// ======================== V2.0 迭代1 新增模型 ========================

// TeacherStudentRelation 师生授权关系
type TeacherStudentRelation struct {
	ID               int64     `json:"id"`
	TeacherID        int64     `json:"teacher_id"`
	StudentID        int64     `json:"student_id"`
	TeacherPersonaID int64     `json:"teacher_persona_id"`
	StudentPersonaID int64     `json:"student_persona_id"`
	Status           string    `json:"status"`       // pending / approved / rejected
	InitiatedBy      string    `json:"initiated_by"` // teacher / student
	IsActive         int       `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TeacherComment 教师评语
type TeacherComment struct {
	ID               int64     `json:"id"`
	TeacherID        int64     `json:"teacher_id"`
	StudentID        int64     `json:"student_id"`
	TeacherPersonaID int64     `json:"teacher_persona_id"`
	StudentPersonaID int64     `json:"student_persona_id"`
	Content          string    `json:"content"`
	ProgressSummary  string    `json:"progress_summary,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// StudentDialogueStyle 个性化问答风格
type StudentDialogueStyle struct {
	ID               int64     `json:"id"`
	TeacherID        int64     `json:"teacher_id"`
	StudentID        int64     `json:"student_id"`
	TeacherPersonaID int64     `json:"teacher_persona_id"`
	StudentPersonaID int64     `json:"student_persona_id"`
	StyleConfig      string    `json:"style_config"` // JSON 字符串
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// StyleConfig 风格配置（JSON 解析用）
type StyleConfig struct {
	Temperature      float64 `json:"temperature"`
	GuidanceLevel    string  `json:"guidance_level"` // low / medium / high
	TeachingStyle    string  `json:"teaching_style"` // V2.0 迭代6: socratic / explanatory / encouraging / strict / companion / custom
	StylePrompt      string  `json:"style_prompt"`
	MaxTurnsPerTopic int     `json:"max_turns_per_topic"`
}

// Assignment 学生作业
type Assignment struct {
	ID               int64     `json:"id"`
	StudentID        int64     `json:"student_id"`
	TeacherID        int64     `json:"teacher_id"`
	TeacherPersonaID int64     `json:"teacher_persona_id"`
	StudentPersonaID int64     `json:"student_persona_id"`
	Title            string    `json:"title"`
	Content          string    `json:"content,omitempty"`
	FilePath         string    `json:"file_path,omitempty"`
	FileType         string    `json:"file_type,omitempty"`
	Status           string    `json:"status"` // submitted / reviewed
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// AssignmentReview 作业点评
type AssignmentReview struct {
	ID           int64     `json:"id"`
	AssignmentID int64     `json:"assignment_id"`
	ReviewerType string    `json:"reviewer_type"` // ai / teacher
	ReviewerID   *int64    `json:"reviewer_id,omitempty"`
	Content      string    `json:"content"`
	Score        *float64  `json:"score,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// RelationWithStudent 关系+学生信息（教师视角）
type RelationWithStudent struct {
	ID               int64     `json:"id"`
	StudentID        int64     `json:"student_id"`
	StudentNickname  string    `json:"student_nickname"`
	TeacherPersonaID int64     `json:"teacher_persona_id"`
	StudentPersonaID int64     `json:"student_persona_id"`
	Status           string    `json:"status"`
	InitiatedBy      string    `json:"initiated_by"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
}

// RelationWithTeacher 关系+教师信息（学生视角）
type RelationWithTeacher struct {
	ID                 int64     `json:"id"`
	TeacherID          int64     `json:"teacher_id"`
	TeacherNickname    string    `json:"teacher_nickname"`
	TeacherSchool      string    `json:"teacher_school"`
	TeacherDescription string    `json:"teacher_description"`
	TeacherPersonaID   int64     `json:"teacher_persona_id"`
	StudentPersonaID   int64     `json:"student_persona_id"`
	Status             string    `json:"status"`
	InitiatedBy        string    `json:"initiated_by"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
}

// CommentWithNames 评语+用户名称（列表展示用）
type CommentWithNames struct {
	ID               int64     `json:"id"`
	TeacherID        int64     `json:"teacher_id"`
	TeacherNickname  string    `json:"teacher_nickname"`
	StudentID        int64     `json:"student_id"`
	StudentNickname  string    `json:"student_nickname"`
	TeacherPersonaID int64     `json:"teacher_persona_id"`
	StudentPersonaID int64     `json:"student_persona_id"`
	Content          string    `json:"content"`
	ProgressSummary  string    `json:"progress_summary,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// AssignmentListItem 作业列表项（含用户名称和统计）
type AssignmentListItem struct {
	ID               int64     `json:"id"`
	StudentID        int64     `json:"student_id"`
	StudentNickname  string    `json:"student_nickname"`
	TeacherID        int64     `json:"teacher_id"`
	TeacherNickname  string    `json:"teacher_nickname"`
	TeacherPersonaID int64     `json:"teacher_persona_id"`
	StudentPersonaID int64     `json:"student_persona_id"`
	Title            string    `json:"title"`
	Status           string    `json:"status"`
	HasFile          bool      `json:"has_file"`
	ReviewCount      int       `json:"review_count"`
	CreatedAt        time.Time `json:"created_at"`
}

// ======================== V2.0 迭代2 新增模型 ========================

// Persona 分身模型
type Persona struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Role        string    `json:"role"`
	Nickname    string    `json:"nickname"`
	School      string    `json:"school,omitempty"`
	Description string    `json:"description,omitempty"`
	Avatar      string    `json:"avatar,omitempty"`
	IsActive    int       `json:"is_active"`
	IsPublic    int       `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PersonaListItem 分身列表项（含统计信息）
type PersonaListItem struct {
	ID            int64     `json:"id"`
	Role          string    `json:"role"`
	Nickname      string    `json:"nickname"`
	School        string    `json:"school,omitempty"`
	Description   string    `json:"description,omitempty"`
	IsActive      bool      `json:"is_active"`
	IsPublic      bool      `json:"is_public"`
	StudentCount  int       `json:"student_count,omitempty"`  // 教师分身：学生数
	DocumentCount int       `json:"document_count,omitempty"` // 教师分身：文档数
	ClassCount    int       `json:"class_count,omitempty"`    // 教师分身：班级数
	TeacherCount  int       `json:"teacher_count,omitempty"`  // 学生分身：教师数
	CreatedAt     time.Time `json:"created_at"`
}

// Class 班级模型
type Class struct {
	ID          int64     `json:"id"`
	PersonaID   int64     `json:"persona_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	IsActive    int       `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ClassWithMemberCount 班级列表项（含成员数）
type ClassWithMemberCount struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	MemberCount int       `json:"member_count"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// ClassMember 班级成员模型
type ClassMember struct {
	ID               int64     `json:"id"`
	ClassID          int64     `json:"class_id"`
	StudentPersonaID int64     `json:"student_persona_id"`
	JoinedAt         time.Time `json:"joined_at"`
}

// ClassMemberItem 班级成员列表项（含学生昵称）
type ClassMemberItem struct {
	ID               int64     `json:"id"`
	StudentPersonaID int64     `json:"student_persona_id"`
	StudentNickname  string    `json:"student_nickname"`
	JoinedAt         time.Time `json:"joined_at"`
}

// PersonaShare 分身分享模型
type PersonaShare struct {
	ID                     int64      `json:"id"`
	TeacherPersonaID       int64      `json:"teacher_persona_id"`
	ShareCode              string     `json:"share_code"`
	ClassID                *int64     `json:"class_id,omitempty"`
	TargetStudentPersonaID int64      `json:"target_student_persona_id"`
	ExpiresAt              *time.Time `json:"expires_at,omitempty"`
	MaxUses                int        `json:"max_uses"`
	UsedCount              int        `json:"used_count"`
	IsActive               int        `json:"is_active"`
	CreatedAt              time.Time  `json:"created_at"`
}

// ShareInfo 分享码信息（预览用）
type ShareInfo struct {
	TeacherPersonaID       int64  `json:"teacher_persona_id"`
	TeacherNickname        string `json:"teacher_nickname"`
	TeacherSchool          string `json:"teacher_school"`
	TeacherDescription     string `json:"teacher_description"`
	ClassName              string `json:"class_name,omitempty"`
	TargetStudentPersonaID int64  `json:"target_student_persona_id"`
	TargetStudentNickname  string `json:"target_student_nickname,omitempty"`
	IsValid                bool   `json:"is_valid"`
	Reason                 string `json:"reason,omitempty"`
}

// ShareListItem 分享码列表项
type ShareListItem struct {
	ID                     int64      `json:"id"`
	ShareCode              string     `json:"share_code"`
	ClassID                *int64     `json:"class_id,omitempty"`
	ClassName              string     `json:"class_name,omitempty"`
	TargetStudentPersonaID int64      `json:"target_student_persona_id"`
	TargetStudentNickname  string     `json:"target_student_nickname,omitempty"`
	ExpiresAt              *time.Time `json:"expires_at,omitempty"`
	MaxUses                int        `json:"max_uses"`
	UsedCount              int        `json:"used_count"`
	IsActive               bool       `json:"is_active"`
	CreatedAt              time.Time  `json:"created_at"`
}

// ======================== V2.0 迭代4 新增模型 ========================

// TeacherTakeover 教师接管记录
type TeacherTakeover struct {
	ID               int64      `json:"id"`
	TeacherPersonaID int64      `json:"teacher_persona_id"`
	StudentPersonaID int64      `json:"student_persona_id"`
	SessionID        string     `json:"session_id"`
	Status           string     `json:"status"` // active / ended
	StartedAt        time.Time  `json:"started_at"`
	EndedAt          *time.Time `json:"ended_at,omitempty"`
}

// MarketplacePersona 广场分身列表项
type MarketplacePersona struct {
	ID                int64  `json:"id"`
	Nickname          string `json:"nickname"`
	School            string `json:"school"`
	Description       string `json:"description"`
	StudentCount      int    `json:"student_count"`
	DocumentCount     int    `json:"document_count"`
	ApplicationStatus string `json:"application_status"` // "" / "pending" / "approved"
}
