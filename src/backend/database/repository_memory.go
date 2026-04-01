package database

import (
	"database/sql"
	"fmt"
	"time"
)

// ==================== MemoryRepository ====================

// MemoryRepository 记忆数据访问层
type MemoryRepository struct {
	db *sql.DB
}

// NewMemoryRepository 创建记忆仓库
func NewMemoryRepository(db *sql.DB) *MemoryRepository {
	return &MemoryRepository{db: db}
}

// Create 创建记忆
func (r *MemoryRepository) Create(mem *Memory) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO memories (student_id, teacher_id, memory_type, memory_layer, content, importance, last_accessed, created_at, updated_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		mem.StudentID, mem.TeacherID, mem.MemoryType, defaultMemoryLayer(mem.MemoryLayer),
		mem.Content, mem.Importance, mem.LastAccessed, time.Now(), time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("创建记忆失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取记忆ID失败: %w", err)
	}

	return id, nil
}

// GetByStudentAndTeacher 根据学生ID和教师ID查询记忆列表
func (r *MemoryRepository) GetByStudentAndTeacher(studentID, teacherID int64, limit int) ([]*Memory, error) {
	rows, err := r.db.Query(
		`SELECT id, student_id, teacher_id, memory_type, COALESCE(memory_layer, 'episodic'), content, importance, last_accessed, created_at, updated_at 
		 FROM memories WHERE student_id = ? AND teacher_id = ? 
		 ORDER BY importance DESC, created_at DESC LIMIT ?`,
		studentID, teacherID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("查询记忆列表失败: %w", err)
	}
	defer rows.Close()

	return scanMemories(rows)
}

// UpdateLastAccessed 更新记忆的最后访问时间
func (r *MemoryRepository) UpdateLastAccessed(id int64) error {
	now := time.Now()
	result, err := r.db.Exec(
		`UPDATE memories SET last_accessed = ?, updated_at = ? WHERE id = ?`,
		now, now, id,
	)
	if err != nil {
		return fmt.Errorf("更新记忆访问时间失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("记忆不存在: id=%d", id)
	}

	return nil
}

// ==================== 分身维度方法 ====================

// CreateWithPersonas 创建带分身维度的记忆
func (r *MemoryRepository) CreateWithPersonas(mem *Memory) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO memories (student_id, teacher_id, teacher_persona_id, student_persona_id, memory_type, memory_layer, content, importance, last_accessed, created_at, updated_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		mem.StudentID, mem.TeacherID, mem.TeacherPersonaID, mem.StudentPersonaID,
		mem.MemoryType, defaultMemoryLayer(mem.MemoryLayer),
		mem.Content, mem.Importance, mem.LastAccessed, time.Now(), time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("创建记忆失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取记忆ID失败: %w", err)
	}

	return id, nil
}

// GetByPersonas 按分身维度查询记忆（仅返回 core + episodic）
func (r *MemoryRepository) GetByPersonas(teacherPersonaID, studentPersonaID int64, limit int) ([]*Memory, error) {
	rows, err := r.db.Query(
		`SELECT id, student_id, teacher_id, COALESCE(teacher_persona_id, 0), COALESCE(student_persona_id, 0),
		        memory_type, COALESCE(memory_layer, 'episodic'), content, importance, last_accessed, created_at, updated_at 
		 FROM memories WHERE teacher_persona_id = ? AND student_persona_id = ? 
		 AND COALESCE(memory_layer, 'episodic') IN ('core', 'episodic')
		 ORDER BY importance DESC, updated_at DESC LIMIT ?`,
		teacherPersonaID, studentPersonaID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("查询记忆列表失败: %w", err)
	}
	defer rows.Close()

	return scanMemoriesWithPersonas(rows)
}

// ==================== V2.0 迭代6 新增方法 ====================

// GetByID 根据ID查询单条记忆
func (r *MemoryRepository) GetByID(id int64) (*Memory, error) {
	mem := &Memory{}
	err := r.db.QueryRow(
		`SELECT id, student_id, teacher_id, COALESCE(teacher_persona_id, 0), COALESCE(student_persona_id, 0),
		        memory_type, COALESCE(memory_layer, 'episodic'), content, importance, last_accessed, created_at, updated_at 
		 FROM memories WHERE id = ?`,
		id,
	).Scan(&mem.ID, &mem.StudentID, &mem.TeacherID, &mem.TeacherPersonaID, &mem.StudentPersonaID,
		&mem.MemoryType, &mem.MemoryLayer, &mem.Content, &mem.Importance, &mem.LastAccessed, &mem.CreatedAt, &mem.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询记忆失败: %w", err)
	}

	return mem, nil
}

// UpdateContent 更新记忆内容（教师编辑）
func (r *MemoryRepository) UpdateContent(id int64, content string, importance *float64, memoryLayer *string) error {
	now := time.Now()
	query := `UPDATE memories SET content = ?, updated_at = ?`
	args := []interface{}{content, now}

	if importance != nil {
		query += `, importance = ?`
		args = append(args, *importance)
	}
	if memoryLayer != nil {
		query += `, memory_layer = ?`
		args = append(args, *memoryLayer)
	}

	query += ` WHERE id = ?`
	args = append(args, id)

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("更新记忆失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("记忆不存在: id=%d", id)
	}

	return nil
}

// DeleteByID 删除记忆
func (r *MemoryRepository) DeleteByID(id int64) error {
	result, err := r.db.Exec(`DELETE FROM memories WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("删除记忆失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("记忆不存在: id=%d", id)
	}

	return nil
}

// UpdateMemoryLayer 批量更新记忆层级（用于摘要合并后标记 archived）
func (r *MemoryRepository) UpdateMemoryLayer(ids []int64, layer string) error {
	if len(ids) == 0 {
		return nil
	}

	// 构建 IN 子句
	placeholders := make([]byte, 0, len(ids)*2)
	args := make([]interface{}, 0, len(ids)+2)
	args = append(args, layer, time.Now())
	for i, id := range ids {
		if i > 0 {
			placeholders = append(placeholders, ',')
		}
		placeholders = append(placeholders, '?')
		args = append(args, id)
	}

	query := fmt.Sprintf(`UPDATE memories SET memory_layer = ?, updated_at = ? WHERE id IN (%s)`, string(placeholders))
	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("批量更新记忆层级失败: %w", err)
	}

	return nil
}

// GetCoreMemory 查询同 memory_type 的最近一条 core 记忆（用于更新覆盖）
func (r *MemoryRepository) GetCoreMemory(teacherPersonaID, studentPersonaID int64, memoryType string) (*Memory, error) {
	mem := &Memory{}
	err := r.db.QueryRow(
		`SELECT id, student_id, teacher_id, COALESCE(teacher_persona_id, 0), COALESCE(student_persona_id, 0),
		        memory_type, COALESCE(memory_layer, 'episodic'), content, importance, last_accessed, created_at, updated_at 
		 FROM memories 
		 WHERE teacher_persona_id = ? AND student_persona_id = ? AND memory_type = ? AND COALESCE(memory_layer, 'episodic') = 'core'
		 ORDER BY updated_at DESC LIMIT 1`,
		teacherPersonaID, studentPersonaID, memoryType,
	).Scan(&mem.ID, &mem.StudentID, &mem.TeacherID, &mem.TeacherPersonaID, &mem.StudentPersonaID,
		&mem.MemoryType, &mem.MemoryLayer, &mem.Content, &mem.Importance, &mem.LastAccessed, &mem.CreatedAt, &mem.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询核心记忆失败: %w", err)
	}

	return mem, nil
}

// UpdateCoreMemory 更新 core 记忆内容（存储时覆盖）
func (r *MemoryRepository) UpdateCoreMemory(id int64, content string, importance float64) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE memories SET content = ?, importance = ?, updated_at = ?, last_accessed = ? WHERE id = ?`,
		content, importance, now, now, id,
	)
	if err != nil {
		return fmt.Errorf("更新核心记忆失败: %w", err)
	}
	return nil
}

// CountEpisodicMemories 统计某学生分身的 episodic 记忆数量
func (r *MemoryRepository) CountEpisodicMemories(teacherPersonaID, studentPersonaID int64) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM memories 
		 WHERE teacher_persona_id = ? AND student_persona_id = ? AND COALESCE(memory_layer, 'episodic') = 'episodic'`,
		teacherPersonaID, studentPersonaID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("统计情景记忆数量失败: %w", err)
	}
	return count, nil
}

// ListEpisodicForSummarize 获取某学生分身的所有 episodic 记忆（用于摘要合并）
func (r *MemoryRepository) ListEpisodicForSummarize(teacherPersonaID, studentPersonaID int64) ([]*Memory, error) {
	rows, err := r.db.Query(
		`SELECT id, student_id, teacher_id, COALESCE(teacher_persona_id, 0), COALESCE(student_persona_id, 0),
		        memory_type, COALESCE(memory_layer, 'episodic'), content, importance, last_accessed, created_at, updated_at 
		 FROM memories 
		 WHERE teacher_persona_id = ? AND student_persona_id = ? AND COALESCE(memory_layer, 'episodic') = 'episodic'
		 ORDER BY updated_at DESC`,
		teacherPersonaID, studentPersonaID,
	)
	if err != nil {
		return nil, fmt.Errorf("查询情景记忆失败: %w", err)
	}
	defer rows.Close()

	return scanMemoriesWithPersonas(rows)
}

// ListMemoriesByPersonasWithFilter SQL层分页+层级筛选（替代应用层分页）
func (r *MemoryRepository) ListMemoriesByPersonasWithFilter(teacherPersonaID, studentPersonaID int64, layer string, page, pageSize int) ([]*Memory, int, error) {
	// 构建 WHERE 条件
	whereClause := `WHERE teacher_persona_id = ? AND student_persona_id = ?`
	args := []interface{}{teacherPersonaID, studentPersonaID}

	if layer != "" {
		whereClause += ` AND COALESCE(memory_layer, 'episodic') = ?`
		args = append(args, layer)
	} else {
		// 默认不返回 archived
		whereClause += ` AND COALESCE(memory_layer, 'episodic') IN ('core', 'episodic')`
	}

	// 查询总数
	var total int
	countQuery := `SELECT COUNT(*) FROM memories ` + whereClause
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询记忆总数失败: %w", err)
	}

	// 查询列表（SQL层分页）
	offset := (page - 1) * pageSize
	listQuery := `SELECT id, student_id, teacher_id, COALESCE(teacher_persona_id, 0), COALESCE(student_persona_id, 0),
	              memory_type, COALESCE(memory_layer, 'episodic'), content, importance, last_accessed, created_at, updated_at 
	              FROM memories ` + whereClause + ` ORDER BY importance DESC, updated_at DESC LIMIT ? OFFSET ?`
	listArgs := append(args, pageSize, offset)

	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询记忆列表失败: %w", err)
	}
	defer rows.Close()

	memories, err := scanMemoriesWithPersonas(rows)
	if err != nil {
		return nil, 0, err
	}

	return memories, total, nil
}

// ListAllStudentPersonaPairs 获取所有有 episodic 记忆的学生分身对（用于定时任务扫描）
func (r *MemoryRepository) ListAllStudentPersonaPairs() ([][2]int64, error) {
	rows, err := r.db.Query(
		`SELECT DISTINCT teacher_persona_id, student_persona_id 
		 FROM memories 
		 WHERE COALESCE(memory_layer, 'episodic') = 'episodic' 
		 AND teacher_persona_id > 0 AND student_persona_id > 0`,
	)
	if err != nil {
		return nil, fmt.Errorf("查询学生分身对失败: %w", err)
	}
	defer rows.Close()

	var pairs [][2]int64
	for rows.Next() {
		var teacherPersonaID, studentPersonaID int64
		if err := rows.Scan(&teacherPersonaID, &studentPersonaID); err != nil {
			return nil, fmt.Errorf("扫描学生分身对失败: %w", err)
		}
		pairs = append(pairs, [2]int64{teacherPersonaID, studentPersonaID})
	}

	return pairs, nil
}

// ==================== 辅助函数 ====================

// defaultMemoryLayer 如果 memory_layer 为空，返回默认值 "episodic"
func defaultMemoryLayer(layer string) string {
	if layer == "" {
		return "episodic"
	}
	return layer
}

// scanMemories 扫描不含分身ID的记忆列表
func scanMemories(rows *sql.Rows) ([]*Memory, error) {
	var memories []*Memory
	for rows.Next() {
		mem := &Memory{}
		if err := rows.Scan(&mem.ID, &mem.StudentID, &mem.TeacherID, &mem.MemoryType,
			&mem.MemoryLayer, &mem.Content, &mem.Importance, &mem.LastAccessed, &mem.CreatedAt, &mem.UpdatedAt); err != nil {
			return nil, fmt.Errorf("扫描记忆记录失败: %w", err)
		}
		memories = append(memories, mem)
	}
	return memories, nil
}

// scanMemoriesWithPersonas 扫描含分身ID的记忆列表
func scanMemoriesWithPersonas(rows *sql.Rows) ([]*Memory, error) {
	var memories []*Memory
	for rows.Next() {
		mem := &Memory{}
		if err := rows.Scan(&mem.ID, &mem.StudentID, &mem.TeacherID, &mem.TeacherPersonaID, &mem.StudentPersonaID,
			&mem.MemoryType, &mem.MemoryLayer, &mem.Content, &mem.Importance, &mem.LastAccessed, &mem.CreatedAt, &mem.UpdatedAt); err != nil {
			return nil, fmt.Errorf("扫描记忆记录失败: %w", err)
		}
		memories = append(memories, mem)
	}
	return memories, nil
}
