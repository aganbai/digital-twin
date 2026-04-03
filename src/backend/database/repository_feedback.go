package database

import (
	"database/sql"
	"fmt"
	"time"
)

// FeedbackRepository 反馈数据访问
type FeedbackRepository struct {
	db *sql.DB
}

// NewFeedbackRepository 创建反馈 Repository
func NewFeedbackRepository(db *sql.DB) *FeedbackRepository {
	return &FeedbackRepository{db: db}
}

// Create 创建反馈
func (r *FeedbackRepository) Create(feedback *Feedback) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO feedbacks (user_id, feedback_type, content, status, context_info)
		VALUES (?, ?, ?, 'pending', ?)`,
		feedback.UserID, feedback.FeedbackType, feedback.Content, feedback.ContextInfo)
	if err != nil {
		return 0, fmt.Errorf("创建反馈失败: %w", err)
	}
	return result.LastInsertId()
}

// ListAll 查询反馈列表（管理员用）
func (r *FeedbackRepository) ListAll(status string, offset, limit int) ([]*Feedback, int, error) {
	// 统计总数
	countSQL := "SELECT COUNT(*) FROM feedbacks"
	querySQL := "SELECT id, user_id, feedback_type, content, status, context_info, created_at, updated_at FROM feedbacks"
	var args []interface{}

	if status != "" {
		countSQL += " WHERE status = ?"
		querySQL += " WHERE status = ?"
		args = append(args, status)
	}

	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	if err := r.db.QueryRow(countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("统计反馈数量失败: %w", err)
	}

	querySQL += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query(querySQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询反馈列表失败: %w", err)
	}
	defer rows.Close()

	var feedbacks []*Feedback
	for rows.Next() {
		f := &Feedback{}
		err := rows.Scan(&f.ID, &f.UserID, &f.FeedbackType, &f.Content, &f.Status, &f.ContextInfo, &f.CreatedAt, &f.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("扫描反馈记录失败: %w", err)
		}
		feedbacks = append(feedbacks, f)
	}
	return feedbacks, total, nil
}

// UpdateStatus 更新反馈状态
func (r *FeedbackRepository) UpdateStatus(id int64, status string) error {
	_, err := r.db.Exec(`UPDATE feedbacks SET status = ?, updated_at = ? WHERE id = ?`, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("更新反馈状态失败: %w", err)
	}
	return nil
}
