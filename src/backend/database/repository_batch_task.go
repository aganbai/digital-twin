package database

import (
	"database/sql"
	"fmt"
	"time"
)

// BatchTaskRepository 批量任务数据访问
type BatchTaskRepository struct {
	db *sql.DB
}

// NewBatchTaskRepository 创建批量任务 Repository
func NewBatchTaskRepository(db *sql.DB) *BatchTaskRepository {
	return &BatchTaskRepository{db: db}
}

// Create 创建批量任务
func (r *BatchTaskRepository) Create(task *BatchTask) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO batch_tasks (task_id, persona_id, knowledge_base_id, status, total_files, success_files, failed_files, result_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		task.TaskID, task.PersonaID, task.KnowledgeBaseID, task.Status,
		task.TotalFiles, task.SuccessFiles, task.FailedFiles, task.ResultJSON)
	if err != nil {
		return 0, fmt.Errorf("创建批量任务失败: %w", err)
	}
	return result.LastInsertId()
}

// GetByTaskID 根据任务ID查询
func (r *BatchTaskRepository) GetByTaskID(taskID string) (*BatchTask, error) {
	row := r.db.QueryRow(`
		SELECT id, task_id, persona_id, knowledge_base_id, status, total_files, success_files, failed_files, result_json, created_at, updated_at
		FROM batch_tasks WHERE task_id = ?`, taskID)

	task := &BatchTask{}
	err := row.Scan(&task.ID, &task.TaskID, &task.PersonaID, &task.KnowledgeBaseID,
		&task.Status, &task.TotalFiles, &task.SuccessFiles, &task.FailedFiles,
		&task.ResultJSON, &task.CreatedAt, &task.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询批量任务失败: %w", err)
	}
	return task, nil
}

// UpdateStatus 更新任务状态
func (r *BatchTaskRepository) UpdateStatus(taskID, status string, successFiles, failedFiles int, resultJSON string) error {
	_, err := r.db.Exec(`
		UPDATE batch_tasks SET status = ?, success_files = ?, failed_files = ?, result_json = ?, updated_at = ?
		WHERE task_id = ?`,
		status, successFiles, failedFiles, resultJSON, time.Now(), taskID)
	if err != nil {
		return fmt.Errorf("更新批量任务状态失败: %w", err)
	}
	return nil
}
