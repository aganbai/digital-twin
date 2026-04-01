package database

import (
	"database/sql"
	"fmt"
	"time"
)

// ==================== StyleRepository ====================

type StyleRepository struct {
	db *sql.DB
}

func NewStyleRepository(db *sql.DB) *StyleRepository {
	return &StyleRepository{db: db}
}

func (r *StyleRepository) Upsert(style *StudentDialogueStyle) (int64, error) {
	now := time.Now()
	result, err := r.db.Exec(
		`INSERT INTO student_dialogue_styles (teacher_id, student_id, style_config, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(teacher_id, student_id) DO UPDATE SET
		   style_config = excluded.style_config,
		   updated_at = excluded.updated_at`,
		style.TeacherID, style.StudentID, style.StyleConfig, now, now,
	)
	if err != nil {
		return 0, fmt.Errorf("upsert 个性化风格失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取风格ID失败: %w", err)
	}

	return id, nil
}

func (r *StyleRepository) GetByTeacherAndStudent(teacherID, studentID int64) (*StudentDialogueStyle, error) {
	style := &StudentDialogueStyle{}
	err := r.db.QueryRow(
		`SELECT id, teacher_id, student_id, style_config, created_at, updated_at
		 FROM student_dialogue_styles WHERE teacher_id = ? AND student_id = ?`,
		teacherID, studentID,
	).Scan(&style.ID, &style.TeacherID, &style.StudentID, &style.StyleConfig, &style.CreatedAt, &style.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询个性化风格失败: %w", err)
	}

	return style, nil
}

// ==================== 分身维度方法 ====================

// UpsertWithPersonas 按分身维度 Upsert 风格
func (r *StyleRepository) UpsertWithPersonas(style *StudentDialogueStyle) (int64, error) {
	now := time.Now()
	result, err := r.db.Exec(
		`INSERT INTO student_dialogue_styles (teacher_id, student_id, teacher_persona_id, student_persona_id, style_config, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(teacher_id, student_id) DO UPDATE SET
		   style_config = excluded.style_config,
		   teacher_persona_id = excluded.teacher_persona_id,
		   student_persona_id = excluded.student_persona_id,
		   updated_at = excluded.updated_at`,
		style.TeacherID, style.StudentID, style.TeacherPersonaID, style.StudentPersonaID,
		style.StyleConfig, now, now,
	)
	if err != nil {
		return 0, fmt.Errorf("upsert 个性化风格失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取风格ID失败: %w", err)
	}

	return id, nil
}

// GetByPersonas 按分身维度查询风格
func (r *StyleRepository) GetByPersonas(teacherPersonaID, studentPersonaID int64) (*StudentDialogueStyle, error) {
	style := &StudentDialogueStyle{}
	err := r.db.QueryRow(
		`SELECT id, teacher_id, student_id, COALESCE(teacher_persona_id, 0), COALESCE(student_persona_id, 0), style_config, created_at, updated_at
		 FROM student_dialogue_styles WHERE teacher_persona_id = ? AND student_persona_id = ?`,
		teacherPersonaID, studentPersonaID,
	).Scan(&style.ID, &style.TeacherID, &style.StudentID, &style.TeacherPersonaID, &style.StudentPersonaID,
		&style.StyleConfig, &style.CreatedAt, &style.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询个性化风格失败: %w", err)
	}

	return style, nil
}
