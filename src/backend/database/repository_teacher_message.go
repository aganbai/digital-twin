package database

import (
	"database/sql"
	"fmt"
	"time"
)

// ==================== TeacherMessageRepository ====================

// TeacherMessageRepository 教师推送消息数据访问层
type TeacherMessageRepository struct {
	db *sql.DB
}

// NewTeacherMessageRepository 创建教师推送消息仓库
func NewTeacherMessageRepository(db *sql.DB) *TeacherMessageRepository {
	return &TeacherMessageRepository{db: db}
}

// Create 创建推送消息记录
func (r *TeacherMessageRepository) Create(msg *TeacherMessage) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO teacher_messages (teacher_id, target_type, target_id, content, status, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		msg.TeacherID, msg.TargetType, msg.TargetID, msg.Content, msg.Status, time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("创建推送消息记录失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取推送消息ID失败: %w", err)
	}

	return id, nil
}

// GetByTeacherID 获取教师的推送历史（分页）
func (r *TeacherMessageRepository) GetByTeacherID(teacherID int64, offset, limit int) ([]*TeacherMessage, int, error) {
	// 查询总数
	var total int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM teacher_messages WHERE teacher_id = ?`,
		teacherID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询推送消息总数失败: %w", err)
	}

	// 查询列表
	rows, err := r.db.Query(
		`SELECT id, teacher_id, target_type, target_id, content, status, created_at
		 FROM teacher_messages WHERE teacher_id = ?
		 ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		teacherID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("查询推送消息列表失败: %w", err)
	}
	defer rows.Close()

	var messages []*TeacherMessage
	for rows.Next() {
		msg := &TeacherMessage{}
		if err := rows.Scan(&msg.ID, &msg.TeacherID, &msg.TargetType, &msg.TargetID,
			&msg.Content, &msg.Status, &msg.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描推送消息记录失败: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, total, nil
}

// GetTodayPushCount 获取教师今日推送次数（用于频率限制）
func (r *TeacherMessageRepository) GetTodayPushCount(teacherID int64) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM teacher_messages WHERE teacher_id = ? AND DATE(created_at) = DATE('now')`,
		teacherID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("查询今日推送次数失败: %w", err)
	}

	return count, nil
}

// UpdateStatus 更新推送状态
func (r *TeacherMessageRepository) UpdateStatus(id int64, status string) error {
	_, err := r.db.Exec(
		`UPDATE teacher_messages SET status = ? WHERE id = ?`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("更新推送状态失败: %w", err)
	}

	return nil
}
