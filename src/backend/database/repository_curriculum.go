package database

import (
	"database/sql"
	"fmt"
	"time"
)

// CurriculumConfigRepository 教材配置数据访问
type CurriculumConfigRepository struct {
	db *sql.DB
}

// NewCurriculumConfigRepository 创建教材配置 Repository
func NewCurriculumConfigRepository(db *sql.DB) *CurriculumConfigRepository {
	return &CurriculumConfigRepository{db: db}
}

// Create 创建教材配置
func (r *CurriculumConfigRepository) Create(config *TeacherCurriculumConfig) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO teacher_curriculum_configs (teacher_id, persona_id, grade_level, grade, textbook_versions, region, subjects, current_progress, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		config.TeacherID, config.PersonaID, config.GradeLevel, config.Grade,
		config.TextbookVersions, config.Region, config.Subjects, config.CurrentProgress, 1)
	if err != nil {
		return 0, fmt.Errorf("创建教材配置失败: %w", err)
	}
	return result.LastInsertId()
}

// Update 更新教材配置（含所有权验证）
func (r *CurriculumConfigRepository) Update(config *TeacherCurriculumConfig) error {
	result, err := r.db.Exec(`
		UPDATE teacher_curriculum_configs
		SET grade_level = ?, grade = ?, textbook_versions = ?, region = ?, subjects = ?, current_progress = ?, updated_at = ?
		WHERE id = ? AND teacher_id = ?`,
		config.GradeLevel, config.Grade, config.TextbookVersions, config.Region,
		config.Subjects, config.CurrentProgress, time.Now(), config.ID, config.TeacherID)
	if err != nil {
		return fmt.Errorf("更新教材配置失败: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("配置不存在或无权修改")
	}
	return nil
}

// GetByID 根据ID查询
func (r *CurriculumConfigRepository) GetByID(id int64) (*TeacherCurriculumConfig, error) {
	row := r.db.QueryRow(`SELECT id, teacher_id, persona_id, grade_level, grade, textbook_versions, region, subjects, current_progress, is_active, created_at, updated_at FROM teacher_curriculum_configs WHERE id = ?`, id)
	return r.scanConfig(row)
}

// GetActiveByPersonaID 获取分身的活跃教材配置
func (r *CurriculumConfigRepository) GetActiveByPersonaID(personaID int64) (*TeacherCurriculumConfig, error) {
	row := r.db.QueryRow(`SELECT id, teacher_id, persona_id, grade_level, grade, textbook_versions, region, subjects, current_progress, is_active, created_at, updated_at FROM teacher_curriculum_configs WHERE persona_id = ? AND is_active = 1 ORDER BY updated_at DESC LIMIT 1`, personaID)
	return r.scanConfig(row)
}

// ListByPersonaID 获取分身的所有教材配置
func (r *CurriculumConfigRepository) ListByPersonaID(personaID int64) ([]*TeacherCurriculumConfig, error) {
	rows, err := r.db.Query(`SELECT id, teacher_id, persona_id, grade_level, grade, textbook_versions, region, subjects, current_progress, is_active, created_at, updated_at FROM teacher_curriculum_configs WHERE persona_id = ? ORDER BY is_active DESC, updated_at DESC`, personaID)
	if err != nil {
		return nil, fmt.Errorf("查询教材配置列表失败: %w", err)
	}
	defer rows.Close()

	var configs []*TeacherCurriculumConfig
	for rows.Next() {
		config := &TeacherCurriculumConfig{}
		err := rows.Scan(&config.ID, &config.TeacherID, &config.PersonaID, &config.GradeLevel, &config.Grade,
			&config.TextbookVersions, &config.Region, &config.Subjects, &config.CurrentProgress,
			&config.IsActive, &config.CreatedAt, &config.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描教材配置失败: %w", err)
		}
		configs = append(configs, config)
	}
	return configs, nil
}

// Delete 删除教材配置
func (r *CurriculumConfigRepository) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM teacher_curriculum_configs WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("删除教材配置失败: %w", err)
	}
	return nil
}

// scanConfig 扫描单行结果
func (r *CurriculumConfigRepository) scanConfig(row *sql.Row) (*TeacherCurriculumConfig, error) {
	config := &TeacherCurriculumConfig{}
	err := row.Scan(&config.ID, &config.TeacherID, &config.PersonaID, &config.GradeLevel, &config.Grade,
		&config.TextbookVersions, &config.Region, &config.Subjects, &config.CurrentProgress,
		&config.IsActive, &config.CreatedAt, &config.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("扫描教材配置失败: %w", err)
	}
	return config, nil
}
