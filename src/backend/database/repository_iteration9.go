package database

import (
	"database/sql"
	"fmt"
	"time"
)

// ==================== CourseNotificationRepository ====================

// CourseNotificationRepository 课程推送通知数据访问层
type CourseNotificationRepository struct {
	db *sql.DB
}

// NewCourseNotificationRepository 创建课程推送通知仓库
func NewCourseNotificationRepository(db *sql.DB) *CourseNotificationRepository {
	return &CourseNotificationRepository{db: db}
}

// Create 创建课程推送通知记录
func (r *CourseNotificationRepository) Create(notification *CourseNotification) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO course_notifications (course_item_id, class_id, teacher_id, persona_id, push_type, status, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		notification.CourseItemID, notification.ClassID, notification.TeacherID,
		notification.PersonaID, notification.PushType, notification.Status, time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("创建课程推送通知失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取通知ID失败: %w", err)
	}

	return id, nil
}

// GetByID 根据ID查询课程推送通知
func (r *CourseNotificationRepository) GetByID(id int64) (*CourseNotification, error) {
	notification := &CourseNotification{}
	err := r.db.QueryRow(
		`SELECT id, course_item_id, class_id, teacher_id, persona_id, push_type, status, created_at
		 FROM course_notifications WHERE id = ?`,
		id,
	).Scan(&notification.ID, &notification.CourseItemID, &notification.ClassID,
		&notification.TeacherID, &notification.PersonaID, &notification.PushType,
		&notification.Status, &notification.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询课程推送通知失败: %w", err)
	}

	return notification, nil
}

// GetByClassID 根据班级ID查询推送通知列表
func (r *CourseNotificationRepository) GetByClassID(classID int64, offset, limit int) ([]CourseNotification, int, error) {
	// 查询总数
	var total int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM course_notifications WHERE class_id = ?`,
		classID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询课程推送通知总数失败: %w", err)
	}

	// 查询列表
	rows, err := r.db.Query(
		`SELECT id, course_item_id, class_id, teacher_id, persona_id, push_type, status, created_at
		 FROM course_notifications WHERE class_id = ?
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		classID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("查询课程推送通知列表失败: %w", err)
	}
	defer rows.Close()

	var notifications []CourseNotification
	for rows.Next() {
		var n CourseNotification
		if err := rows.Scan(&n.ID, &n.CourseItemID, &n.ClassID, &n.TeacherID,
			&n.PersonaID, &n.PushType, &n.Status, &n.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描课程推送通知记录失败: %w", err)
		}
		notifications = append(notifications, n)
	}

	return notifications, total, nil
}

// UpdateStatus 更新推送状态
func (r *CourseNotificationRepository) UpdateStatus(id int64, status string) error {
	result, err := r.db.Exec(
		`UPDATE course_notifications SET status = ? WHERE id = ?`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("更新推送状态失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("课程推送通知不存在: id=%d", id)
	}

	return nil
}

// GetPendingByPersona 获取待推送的通知（按教师分身）
func (r *CourseNotificationRepository) GetPendingByPersona(personaID int64, limit int) ([]CourseNotification, error) {
	rows, err := r.db.Query(
		`SELECT id, course_item_id, class_id, teacher_id, persona_id, push_type, status, created_at
		 FROM course_notifications WHERE persona_id = ? AND status = 'pending'
		 ORDER BY created_at ASC
		 LIMIT ?`,
		personaID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("查询待推送通知失败: %w", err)
	}
	defer rows.Close()

	var notifications []CourseNotification
	for rows.Next() {
		var n CourseNotification
		if err := rows.Scan(&n.ID, &n.CourseItemID, &n.ClassID, &n.TeacherID,
			&n.PersonaID, &n.PushType, &n.Status, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("扫描待推送通知记录失败: %w", err)
		}
		notifications = append(notifications, n)
	}

	return notifications, nil
}

// ==================== WxSubscriptionRepository ====================

// WxSubscriptionRepository 微信订阅状态数据访问层
type WxSubscriptionRepository struct {
	db *sql.DB
}

// NewWxSubscriptionRepository 创建微信订阅状态仓库
func NewWxSubscriptionRepository(db *sql.DB) *WxSubscriptionRepository {
	return &WxSubscriptionRepository{db: db}
}

// Upsert 创建或更新订阅状态
func (r *WxSubscriptionRepository) Upsert(userID int64, templateID string, isSubscribed bool) error {
	now := time.Now()
	var lastSubscribeTime *time.Time
	if isSubscribed {
		lastSubscribeTime = &now
	}

	_, err := r.db.Exec(
		`INSERT INTO wx_subscriptions (user_id, template_id, is_subscribed, last_subscribe_time, created_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(user_id, template_id) DO UPDATE SET
		 is_subscribed = excluded.is_subscribed,
		 last_subscribe_time = excluded.last_subscribe_time`,
		userID, templateID, isSubscribed, lastSubscribeTime, now,
	)
	if err != nil {
		return fmt.Errorf("更新微信订阅状态失败: %w", err)
	}

	return nil
}

// GetByUserAndTemplate 根据用户ID和模板ID查询订阅状态
func (r *WxSubscriptionRepository) GetByUserAndTemplate(userID int64, templateID string) (*WxSubscription, error) {
	sub := &WxSubscription{}
	var isSubscribed int
	err := r.db.QueryRow(
		`SELECT id, user_id, template_id, is_subscribed, last_subscribe_time, created_at
		 FROM wx_subscriptions WHERE user_id = ? AND template_id = ?`,
		userID, templateID,
	).Scan(&sub.ID, &sub.UserID, &sub.TemplateID, &isSubscribed, &sub.LastSubscribeTime, &sub.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询微信订阅状态失败: %w", err)
	}

	sub.IsSubscribed = isSubscribed == 1
	return sub, nil
}

// GetByUser 获取用户的所有订阅状态
func (r *WxSubscriptionRepository) GetByUser(userID int64) ([]WxSubscription, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, template_id, is_subscribed, last_subscribe_time, created_at
		 FROM wx_subscriptions WHERE user_id = ?`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("查询用户订阅列表失败: %w", err)
	}
	defer rows.Close()

	var subscriptions []WxSubscription
	for rows.Next() {
		var sub WxSubscription
		var isSubscribed int
		if err := rows.Scan(&sub.ID, &sub.UserID, &sub.TemplateID, &isSubscribed,
			&sub.LastSubscribeTime, &sub.CreatedAt); err != nil {
			return nil, fmt.Errorf("扫描订阅记录失败: %w", err)
		}
		sub.IsSubscribed = isSubscribed == 1
		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}

// IsSubscribed 检查用户是否订阅了某模板
func (r *WxSubscriptionRepository) IsSubscribed(userID int64, templateID string) (bool, error) {
	var isSubscribed int
	err := r.db.QueryRow(
		`SELECT is_subscribed FROM wx_subscriptions WHERE user_id = ? AND template_id = ?`,
		userID, templateID,
	).Scan(&isSubscribed)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("检查订阅状态失败: %w", err)
	}

	return isSubscribed == 1, nil
}

// ==================== SessionTitleRepository ====================

// SessionTitleRepository 会话标题数据访问层
type SessionTitleRepository struct {
	db *sql.DB
}

// NewSessionTitleRepository 创建会话标题仓库
func NewSessionTitleRepository(db *sql.DB) *SessionTitleRepository {
	return &SessionTitleRepository{db: db}
}

// Create 创建会话标题
func (r *SessionTitleRepository) Create(title *SessionTitle) (int64, error) {
	now := time.Now()
	result, err := r.db.Exec(
		`INSERT INTO session_titles (session_id, student_persona_id, teacher_persona_id, title, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		title.SessionID, title.StudentPersonaID, title.TeacherPersonaID, title.Title, now, now,
	)
	if err != nil {
		return 0, fmt.Errorf("创建会话标题失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取会话标题ID失败: %w", err)
	}

	return id, nil
}

// GetBySessionID 根据会话ID查询标题
func (r *SessionTitleRepository) GetBySessionID(sessionID string) (*SessionTitle, error) {
	title := &SessionTitle{}
	err := r.db.QueryRow(
		`SELECT id, session_id, student_persona_id, teacher_persona_id, title, created_at, updated_at
		 FROM session_titles WHERE session_id = ?`,
		sessionID,
	).Scan(&title.ID, &title.SessionID, &title.StudentPersonaID, &title.TeacherPersonaID,
		&title.Title, &title.CreatedAt, &title.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询会话标题失败: %w", err)
	}

	return title, nil
}

// Update 更新会话标题
func (r *SessionTitleRepository) Update(sessionID string, newTitle string) error {
	result, err := r.db.Exec(
		`UPDATE session_titles SET title = ?, updated_at = ? WHERE session_id = ?`,
		newTitle, time.Now(), sessionID,
	)
	if err != nil {
		return fmt.Errorf("更新会话标题失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("会话标题不存在: session_id=%s", sessionID)
	}

	return nil
}

// Upsert 创建或更新会话标题
func (r *SessionTitleRepository) Upsert(sessionID string, studentPersonaID, teacherPersonaID int64, title string) error {
	now := time.Now()
	_, err := r.db.Exec(
		`INSERT INTO session_titles (session_id, student_persona_id, teacher_persona_id, title, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(session_id) DO UPDATE SET title = excluded.title, updated_at = excluded.updated_at`,
		sessionID, studentPersonaID, teacherPersonaID, title, now, now,
	)
	if err != nil {
		return fmt.Errorf("创建或更新会话标题失败: %w", err)
	}

	return nil
}

// GetByStudentPersona 获取学生分身的所有会话标题
func (r *SessionTitleRepository) GetByStudentPersona(studentPersonaID int64, offset, limit int) ([]SessionTitle, int, error) {
	// 查询总数
	var total int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM session_titles WHERE student_persona_id = ?`,
		studentPersonaID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询会话标题总数失败: %w", err)
	}

	// 查询列表
	rows, err := r.db.Query(
		`SELECT id, session_id, student_persona_id, teacher_persona_id, title, created_at, updated_at
		 FROM session_titles WHERE student_persona_id = ?
		 ORDER BY updated_at DESC
		 LIMIT ? OFFSET ?`,
		studentPersonaID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("查询会话标题列表失败: %w", err)
	}
	defer rows.Close()

	var titles []SessionTitle
	for rows.Next() {
		var t SessionTitle
		if err := rows.Scan(&t.ID, &t.SessionID, &t.StudentPersonaID, &t.TeacherPersonaID,
			&t.Title, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描会话标题记录失败: %w", err)
		}
		titles = append(titles, t)
	}

	return titles, total, nil
}

// GetByTeacherPersona 获取教师分身的所有会话标题
func (r *SessionTitleRepository) GetByTeacherPersona(teacherPersonaID int64, offset, limit int) ([]SessionTitle, int, error) {
	// 查询总数
	var total int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM session_titles WHERE teacher_persona_id = ?`,
		teacherPersonaID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询会话标题总数失败: %w", err)
	}

	// 查询列表
	rows, err := r.db.Query(
		`SELECT id, session_id, student_persona_id, teacher_persona_id, title, created_at, updated_at
		 FROM session_titles WHERE teacher_persona_id = ?
		 ORDER BY updated_at DESC
		 LIMIT ? OFFSET ?`,
		teacherPersonaID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("查询会话标题列表失败: %w", err)
	}
	defer rows.Close()

	var titles []SessionTitle
	for rows.Next() {
		var t SessionTitle
		if err := rows.Scan(&t.ID, &t.SessionID, &t.StudentPersonaID, &t.TeacherPersonaID,
			&t.Title, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描会话标题记录失败: %w", err)
		}
		titles = append(titles, t)
	}

	return titles, total, nil
}

// Delete 删除会话标题
func (r *SessionTitleRepository) Delete(sessionID string) error {
	_, err := r.db.Exec(`DELETE FROM session_titles WHERE session_id = ?`, sessionID)
	if err != nil {
		return fmt.Errorf("删除会话标题失败: %w", err)
	}
	return nil
}
