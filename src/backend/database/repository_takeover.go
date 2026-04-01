package database

import (
	"database/sql"
	"fmt"
	"time"
)

// TakeoverRepository 教师接管数据访问层
type TakeoverRepository struct {
	db *sql.DB
}

// NewTakeoverRepository 创建接管仓库
func NewTakeoverRepository(db *sql.DB) *TakeoverRepository {
	return &TakeoverRepository{db: db}
}

// CreateOrGetActive 创建接管记录（如果已有 active 记录则返回已有的）
func (r *TakeoverRepository) CreateOrGetActive(teacherPersonaID, studentPersonaID int64, sessionID string) (*TeacherTakeover, error) {
	// 先查询是否已有 active 记录
	existing, err := r.GetActiveBySession(sessionID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	// 创建新记录
	now := time.Now()
	result, err := r.db.Exec(
		`INSERT INTO teacher_takeovers (teacher_persona_id, student_persona_id, session_id, status, started_at)
		 VALUES (?, ?, ?, 'active', ?)`,
		teacherPersonaID, studentPersonaID, sessionID, now,
	)
	if err != nil {
		return nil, fmt.Errorf("创建接管记录失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("获取接管记录ID失败: %w", err)
	}

	return &TeacherTakeover{
		ID:               id,
		TeacherPersonaID: teacherPersonaID,
		StudentPersonaID: studentPersonaID,
		SessionID:        sessionID,
		Status:           "active",
		StartedAt:        now,
	}, nil
}

// GetActiveBySession 查询会话的活跃接管记录
func (r *TakeoverRepository) GetActiveBySession(sessionID string) (*TeacherTakeover, error) {
	takeover := &TeacherTakeover{}
	err := r.db.QueryRow(
		`SELECT id, teacher_persona_id, student_persona_id, session_id, status, started_at, ended_at
		 FROM teacher_takeovers WHERE session_id = ? AND status = 'active' LIMIT 1`,
		sessionID,
	).Scan(&takeover.ID, &takeover.TeacherPersonaID, &takeover.StudentPersonaID,
		&takeover.SessionID, &takeover.Status, &takeover.StartedAt, &takeover.EndedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询接管记录失败: %w", err)
	}

	return takeover, nil
}

// EndTakeover 结束接管
func (r *TakeoverRepository) EndTakeover(sessionID string, teacherPersonaID int64) error {
	now := time.Now()
	result, err := r.db.Exec(
		`UPDATE teacher_takeovers SET status = 'ended', ended_at = ? WHERE session_id = ? AND teacher_persona_id = ? AND status = 'active'`,
		now, sessionID, teacherPersonaID,
	)
	if err != nil {
		return fmt.Errorf("结束接管失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("接管记录不存在或已结束")
	}

	return nil
}

// IsSessionTakenOver 检查会话是否被接管
func (r *TakeoverRepository) IsSessionTakenOver(sessionID string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM teacher_takeovers WHERE session_id = ? AND status = 'active'`,
		sessionID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查接管状态失败: %w", err)
	}
	return count > 0, nil
}
