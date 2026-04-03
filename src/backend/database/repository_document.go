// Deprecated: 迭代8已废弃documents表，请使用repository_knowledge.go中的KnowledgeRepository。
// 本文件仍被handlers_chat_import.go和handlers_knowledge.go引用，待后续迁移后移除。
package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

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
		`INSERT INTO documents (teacher_id, title, content, doc_type, tags, status, scope, scope_id, persona_id, summary, source_session_id, created_at, updated_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		doc.TeacherID, doc.Title, doc.Content, doc.DocType, doc.Tags, doc.Status,
		doc.Scope, doc.ScopeID, doc.PersonaID, doc.Summary, doc.SourceSessionID,
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
		`SELECT id, teacher_id, title, content, doc_type, tags, status,
		        COALESCE(scope, 'global'), COALESCE(scope_id, 0), COALESCE(persona_id, 0),
		        COALESCE(summary, ''),
		        created_at, updated_at 
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
			&doc.DocType, &doc.Tags, &doc.Status,
			&doc.Scope, &doc.ScopeID, &doc.PersonaID,
			&doc.Summary,
			&doc.CreatedAt, &doc.UpdatedAt); err != nil {
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
		`SELECT id, teacher_id, title, content, doc_type, tags, status,
		        COALESCE(scope, 'global'), COALESCE(scope_id, 0), COALESCE(persona_id, 0),
		        COALESCE(summary, ''),
		        created_at, updated_at 
		 FROM documents WHERE id = ?`,
		id,
	).Scan(&doc.ID, &doc.TeacherID, &doc.Title, &doc.Content,
		&doc.DocType, &doc.Tags, &doc.Status,
		&doc.Scope, &doc.ScopeID, &doc.PersonaID,
		&doc.Summary,
		&doc.CreatedAt, &doc.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询文档失败: %w", err)
	}

	return doc, nil
}

// GetByPersonaID 获取教师分身的文档列表（可按 scope 筛选）
func (r *DocumentRepository) GetByPersonaID(personaID int64, scope string) ([]Document, error) {
	query := `SELECT id, teacher_id, title, content, doc_type, tags, status,
	                 COALESCE(scope, 'global'), COALESCE(scope_id, 0), COALESCE(persona_id, 0),
	                 COALESCE(summary, ''),
	                 created_at, updated_at
	          FROM documents WHERE persona_id = ? AND status = 'active'`
	args := []interface{}{personaID}

	if scope != "" {
		query += ` AND scope = ?`
		args = append(args, scope)
	}

	query += ` ORDER BY created_at DESC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询教师分身文档列表失败: %w", err)
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var doc Document
		if err := rows.Scan(&doc.ID, &doc.TeacherID, &doc.Title, &doc.Content,
			&doc.DocType, &doc.Tags, &doc.Status,
			&doc.Scope, &doc.ScopeID, &doc.PersonaID,
			&doc.Summary,
			&doc.CreatedAt, &doc.UpdatedAt); err != nil {
			return nil, fmt.Errorf("扫描文档记录失败: %w", err)
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

// GetByStudentScope 获取学生可见的文档（global + 所在班级 + 指定给该学生的）
func (r *DocumentRepository) GetByStudentScope(teacherPersonaID, studentPersonaID int64, classIDs []int64) ([]Document, error) {
	// 基础条件：属于该教师分身且状态为 active
	query := `SELECT id, teacher_id, title, content, doc_type, tags, status,
	                 COALESCE(scope, 'global'), COALESCE(scope_id, 0), COALESCE(persona_id, 0),
	                 COALESCE(summary, ''),
	                 created_at, updated_at
	          FROM documents WHERE persona_id = ? AND status = 'active'
	          AND (`
	args := []interface{}{teacherPersonaID}

	// scope = 'global'
	query += `scope = 'global' OR scope IS NULL`

	// scope = 'student' AND scope_id = studentPersonaID
	query += ` OR (scope = 'student' AND scope_id = ?)`
	args = append(args, studentPersonaID)

	// scope = 'class' AND scope_id IN (classIDs...)
	if len(classIDs) > 0 {
		placeholders := make([]string, len(classIDs))
		for i, cid := range classIDs {
			placeholders[i] = "?"
			args = append(args, cid)
		}
		query += fmt.Sprintf(` OR (scope = 'class' AND scope_id IN (%s))`, strings.Join(placeholders, ","))
	}

	query += `) ORDER BY created_at DESC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询学生可见文档失败: %w", err)
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var doc Document
		if err := rows.Scan(&doc.ID, &doc.TeacherID, &doc.Title, &doc.Content,
			&doc.DocType, &doc.Tags, &doc.Status,
			&doc.Scope, &doc.ScopeID, &doc.PersonaID,
			&doc.Summary,
			&doc.CreatedAt, &doc.UpdatedAt); err != nil {
			return nil, fmt.Errorf("扫描文档记录失败: %w", err)
		}
		docs = append(docs, doc)
	}

	return docs, nil
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
