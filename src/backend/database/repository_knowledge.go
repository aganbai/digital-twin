package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// KnowledgeRepository 知识库数据访问
type KnowledgeRepository struct {
	db *Database
}

// NewKnowledgeRepository 创建知识库仓库
func NewKnowledgeRepository(db *Database) *KnowledgeRepository {
	return &KnowledgeRepository{db: db}
}

// CreateKnowledgeItem 创建知识库条目
func (r *KnowledgeRepository) CreateKnowledgeItem(item *KnowledgeItem) (int64, error) {
	result, err := r.db.DB.Exec(`
		INSERT INTO knowledge_items (
			teacher_id, persona_id, title, content, item_type,
			source_url, file_url, file_name, file_size, tags,
			status, summary, scope, scope_id, source_session_id,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.TeacherID, item.PersonaID, item.Title, item.Content, item.ItemType,
		item.SourceURL, item.FileURL, item.FileName, item.FileSize, item.Tags,
		item.Status, item.Summary, item.Scope, item.ScopeID, item.SourceSessionID,
		time.Now(), time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("创建知识库条目失败: %w", err)
	}
	return result.LastInsertId()
}

// GetKnowledgeItemByID 根据ID获取知识库条目
func (r *KnowledgeRepository) GetKnowledgeItemByID(id int64) (*KnowledgeItem, error) {
	row := r.db.DB.QueryRow(`
		SELECT id, teacher_id, persona_id, title, content, item_type,
			source_url, file_url, file_name, file_size, tags,
			status, summary, scope, scope_id, source_session_id,
			created_at, updated_at
		FROM knowledge_items WHERE id = ?`, id)

	item := &KnowledgeItem{}
	err := row.Scan(
		&item.ID, &item.TeacherID, &item.PersonaID, &item.Title, &item.Content, &item.ItemType,
		&item.SourceURL, &item.FileURL, &item.FileName, &item.FileSize, &item.Tags,
		&item.Status, &item.Summary, &item.Scope, &item.ScopeID, &item.SourceSessionID,
		&item.CreatedAt, &item.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询知识库条目失败: %w", err)
	}
	return item, nil
}

// SearchKnowledgeItems 搜索知识库条目
func (r *KnowledgeRepository) SearchKnowledgeItems(teacherID int64, keyword string, itemType string, scope string, limit, offset int) ([]KnowledgeItemListItem, int64, error) {
	whereClause := "WHERE teacher_id = ?"
	args := []interface{}{teacherID}

	if keyword != "" {
		whereClause += " AND (title LIKE ? OR content LIKE ? OR tags LIKE ?)"
		likePattern := "%" + keyword + "%"
		args = append(args, likePattern, likePattern, likePattern)
	}

	if itemType != "" {
		whereClause += " AND item_type = ?"
		args = append(args, itemType)
	}

	if scope != "" {
		whereClause += " AND scope = ?"
		args = append(args, scope)
	}

	// 查询总数
	var total int64
	countQuery := "SELECT COUNT(*) FROM knowledge_items " + whereClause
	if err := r.db.DB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询总数失败: %w", err)
	}

	// 查询列表
	query := `
		SELECT id, title, item_type, file_name, file_size, tags, status, summary, scope, created_at
		FROM knowledge_items ` + whereClause + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.db.DB.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询知识库列表失败: %w", err)
	}
	defer rows.Close()

	var items []KnowledgeItemListItem
	for rows.Next() {
		var item KnowledgeItemListItem
		err := rows.Scan(
			&item.ID, &item.Title, &item.ItemType, &item.FileName, &item.FileSize,
			&item.Tags, &item.Status, &item.Summary, &item.Scope, &item.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("扫描知识库条目失败: %w", err)
		}
		items = append(items, item)
	}

	return items, total, nil
}

// UpdateKnowledgeItem 更新知识库条目
func (r *KnowledgeRepository) UpdateKnowledgeItem(item *KnowledgeItem) error {
	_, err := r.db.DB.Exec(`
		UPDATE knowledge_items SET
			title = ?, content = ?, tags = ?, summary = ?, scope = ?, scope_id = ?,
			status = ?, updated_at = ?
		WHERE id = ? AND teacher_id = ?`,
		item.Title, item.Content, item.Tags, item.Summary, item.Scope, item.ScopeID,
		item.Status, time.Now(),
		item.ID, item.TeacherID,
	)
	if err != nil {
		return fmt.Errorf("更新知识库条目失败: %w", err)
	}
	return nil
}

// DeleteKnowledgeItem 删除知识库条目
func (r *KnowledgeRepository) DeleteKnowledgeItem(id, teacherID int64) error {
	_, err := r.db.DB.Exec(`
		DELETE FROM knowledge_items WHERE id = ? AND teacher_id = ?`,
		id, teacherID,
	)
	if err != nil {
		return fmt.Errorf("删除知识库条目失败: %w", err)
	}
	return nil
}

// UpdateKnowledgeItemStatus 更新知识库条目状态
func (r *KnowledgeRepository) UpdateKnowledgeItemStatus(id int64, status string) error {
	_, err := r.db.DB.Exec(`
		UPDATE knowledge_items SET status = ?, updated_at = ? WHERE id = ?`,
		status, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("更新知识库状态失败: %w", err)
	}
	return nil
}

// GetKnowledgeItemsByPersonaID 获取分身的所有知识库条目
func (r *KnowledgeRepository) GetKnowledgeItemsByPersonaID(personaID int64) ([]KnowledgeItem, error) {
	rows, err := r.db.DB.Query(`
		SELECT id, teacher_id, persona_id, title, content, item_type,
			source_url, file_url, file_name, file_size, tags,
			status, summary, scope, scope_id, source_session_id,
			created_at, updated_at
		FROM knowledge_items
		WHERE persona_id = ? AND status = 'active'
		ORDER BY created_at DESC`, personaID)
	if err != nil {
		return nil, fmt.Errorf("查询知识库条目失败: %w", err)
	}
	defer rows.Close()

	var items []KnowledgeItem
	for rows.Next() {
		var item KnowledgeItem
		err := rows.Scan(
			&item.ID, &item.TeacherID, &item.PersonaID, &item.Title, &item.Content, &item.ItemType,
			&item.SourceURL, &item.FileURL, &item.FileName, &item.FileSize, &item.Tags,
			&item.Status, &item.Summary, &item.Scope, &item.ScopeID, &item.SourceSessionID,
			&item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描知识库条目失败: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

// BatchCreateKnowledgeItems 批量创建知识库条目
func (r *KnowledgeRepository) BatchCreateKnowledgeItems(items []KnowledgeItem) error {
	tx, err := r.db.DB.Begin()
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO knowledge_items (
			teacher_id, persona_id, title, content, item_type,
			source_url, file_url, file_name, file_size, tags,
			status, summary, scope, scope_id,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("准备语句失败: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, item := range items {
		_, err := stmt.Exec(
			item.TeacherID, item.PersonaID, item.Title, item.Content, item.ItemType,
			item.SourceURL, item.FileURL, item.FileName, item.FileSize, item.Tags,
			item.Status, item.Summary, item.Scope, item.ScopeID,
			now, now,
		)
		if err != nil {
			return fmt.Errorf("批量插入知识库条目失败: %w", err)
		}
	}

	return tx.Commit()
}

// ParseTags 解析标签JSON字符串
func ParseTags(tagsJSON string) ([]string, error) {
	if tagsJSON == "" {
		return []string{}, nil
	}
	var tags []string
	if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil {
		return nil, err
	}
	return tags, nil
}

// TagsToJSON 将标签数组转为JSON字符串
func TagsToJSON(tags []string) string {
	if len(tags) == 0 {
		return "[]"
	}
	data, _ := json.Marshal(tags)
	return string(data)
}

// BuildSearchContent 构建搜索内容（标题+内容）
func BuildSearchContent(title, content string) string {
	var parts []string
	if title != "" {
		parts = append(parts, title)
	}
	if content != "" {
		parts = append(parts, content)
	}
	return strings.Join(parts, " ")
}
