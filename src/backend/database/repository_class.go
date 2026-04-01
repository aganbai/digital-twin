package database

import (
	"database/sql"
	"fmt"
	"time"
)

// ClassRepository 班级数据访问层
type ClassRepository struct {
	db *sql.DB
}

// NewClassRepository 创建班级仓库
func NewClassRepository(db *sql.DB) *ClassRepository {
	return &ClassRepository{db: db}
}

// Create 创建班级
func (r *ClassRepository) Create(class *Class) (int64, error) {
	now := time.Now()
	result, err := r.db.Exec(
		`INSERT INTO classes (persona_id, name, description, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?)`,
		class.PersonaID, class.Name, class.Description, now, now,
	)
	if err != nil {
		return 0, fmt.Errorf("创建班级失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取班级ID失败: %w", err)
	}

	return id, nil
}

// GetByID 根据ID查询班级
func (r *ClassRepository) GetByID(id int64) (*Class, error) {
	class := &Class{}
	err := r.db.QueryRow(
		`SELECT id, persona_id, name, description, COALESCE(is_active, 1), created_at, updated_at
		 FROM classes WHERE id = ?`,
		id,
	).Scan(&class.ID, &class.PersonaID, &class.Name, &class.Description, &class.IsActive, &class.CreatedAt, &class.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询班级失败: %w", err)
	}

	return class, nil
}

// ListByPersonaID 获取教师分身的班级列表（含成员数）
func (r *ClassRepository) ListByPersonaID(personaID int64) ([]ClassWithMemberCount, error) {
	rows, err := r.db.Query(
		`SELECT c.id, c.name, c.description,
		        COALESCE((SELECT COUNT(*) FROM class_members WHERE class_id = c.id), 0) AS member_count,
		        COALESCE(c.is_active, 1),
		        c.created_at
		 FROM classes c WHERE c.persona_id = ?
		 ORDER BY c.created_at ASC`,
		personaID,
	)
	if err != nil {
		return nil, fmt.Errorf("查询班级列表失败: %w", err)
	}
	defer rows.Close()

	var items []ClassWithMemberCount
	for rows.Next() {
		var item ClassWithMemberCount
		var isActiveInt int
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.MemberCount, &isActiveInt, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("扫描班级记录失败: %w", err)
		}
		item.IsActive = isActiveInt == 1
		items = append(items, item)
	}

	return items, nil
}

// Update 更新班级信息
func (r *ClassRepository) Update(id int64, name, description string) error {
	result, err := r.db.Exec(
		`UPDATE classes SET name = ?, description = ?, updated_at = ? WHERE id = ?`,
		name, description, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("更新班级失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("班级不存在: id=%d", id)
	}

	return nil
}

// Delete 删除班级
func (r *ClassRepository) Delete(id int64) error {
	result, err := r.db.Exec(`DELETE FROM classes WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("删除班级失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("班级不存在: id=%d", id)
	}

	return nil
}

// GetMemberCount 获取班级成员数
func (r *ClassRepository) GetMemberCount(classID int64) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM class_members WHERE class_id = ?`, classID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("查询班级成员数失败: %w", err)
	}
	return count, nil
}

// CheckNameExists 检查同一教师分身下班级名是否已存在
func (r *ClassRepository) CheckNameExists(personaID int64, name string, excludeID int64) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM classes WHERE persona_id = ? AND name = ? AND id != ?`,
		personaID, name, excludeID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查班级名是否存在失败: %w", err)
	}
	return count > 0, nil
}

// AddMember 添加班级成员
func (r *ClassRepository) AddMember(classID, studentPersonaID int64) (int64, error) {
	now := time.Now()
	result, err := r.db.Exec(
		`INSERT INTO class_members (class_id, student_persona_id, joined_at)
		 VALUES (?, ?, ?)`,
		classID, studentPersonaID, now,
	)
	if err != nil {
		return 0, fmt.Errorf("添加班级成员失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取成员ID失败: %w", err)
	}

	return id, nil
}

// RemoveMember 移除班级成员
func (r *ClassRepository) RemoveMember(memberID int64) error {
	result, err := r.db.Exec(`DELETE FROM class_members WHERE id = ?`, memberID)
	if err != nil {
		return fmt.Errorf("移除班级成员失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("成员记录不存在: id=%d", memberID)
	}

	return nil
}

// GetMemberByID 根据ID查询成员记录
func (r *ClassRepository) GetMemberByID(memberID int64) (*ClassMember, error) {
	member := &ClassMember{}
	err := r.db.QueryRow(
		`SELECT id, class_id, student_persona_id, joined_at FROM class_members WHERE id = ?`,
		memberID,
	).Scan(&member.ID, &member.ClassID, &member.StudentPersonaID, &member.JoinedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询成员记录失败: %w", err)
	}

	return member, nil
}

// ListMembers 获取班级成员列表（含学生昵称）
func (r *ClassRepository) ListMembers(classID int64, offset, limit int) ([]ClassMemberItem, int, error) {
	// 查询总数
	var total int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM class_members WHERE class_id = ?`, classID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询成员总数失败: %w", err)
	}

	// 查询列表
	rows, err := r.db.Query(
		`SELECT cm.id, cm.student_persona_id, COALESCE(p.nickname, '') AS student_nickname, cm.joined_at
		 FROM class_members cm
		 LEFT JOIN personas p ON cm.student_persona_id = p.id
		 WHERE cm.class_id = ?
		 ORDER BY cm.joined_at ASC
		 LIMIT ? OFFSET ?`,
		classID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("查询成员列表失败: %w", err)
	}
	defer rows.Close()

	var items []ClassMemberItem
	for rows.Next() {
		var item ClassMemberItem
		if err := rows.Scan(&item.ID, &item.StudentPersonaID, &item.StudentNickname, &item.JoinedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描成员记录失败: %w", err)
		}
		items = append(items, item)
	}

	return items, total, nil
}

// IsMember 检查学生是否已在班级中
func (r *ClassRepository) IsMember(classID, studentPersonaID int64) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM class_members WHERE class_id = ? AND student_persona_id = ?`,
		classID, studentPersonaID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查班级成员失败: %w", err)
	}
	return count > 0, nil
}

// GetClassIDsByStudentPersona 获取学生分身所在的班级ID列表（指定教师分身下）
func (r *ClassRepository) GetClassIDsByStudentPersona(teacherPersonaID, studentPersonaID int64) ([]int64, error) {
	rows, err := r.db.Query(
		`SELECT cm.class_id FROM class_members cm
		 JOIN classes c ON cm.class_id = c.id
		 WHERE c.persona_id = ? AND cm.student_persona_id = ?`,
		teacherPersonaID, studentPersonaID,
	)
	if err != nil {
		return nil, fmt.Errorf("查询学生所在班级失败: %w", err)
	}
	defer rows.Close()

	var classIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("扫描班级ID失败: %w", err)
		}
		classIDs = append(classIDs, id)
	}

	return classIDs, nil
}

// ======================== V2.0 迭代3 新增方法 ========================

// ToggleClassResult 班级启停结果
type ToggleClassResult struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	IsActive         bool      `json:"is_active"`
	AffectedStudents int       `json:"affected_students"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ToggleClass 启用/停用班级
func (r *ClassRepository) ToggleClass(classID int64, isActive int) (*ToggleClassResult, error) {
	now := time.Now()

	// 更新 is_active
	result, err := r.db.Exec(
		`UPDATE classes SET is_active = ?, updated_at = ? WHERE id = ?`,
		isActive, now, classID,
	)
	if err != nil {
		return nil, fmt.Errorf("更新班级状态失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return nil, fmt.Errorf("班级不存在: id=%d", classID)
	}

	// 查询班级信息和受影响学生数
	var name string
	r.db.QueryRow(`SELECT name FROM classes WHERE id = ?`, classID).Scan(&name)

	var studentCount int
	r.db.QueryRow(`SELECT COUNT(*) FROM class_members WHERE class_id = ?`, classID).Scan(&studentCount)

	return &ToggleClassResult{
		ID:               classID,
		Name:             name,
		IsActive:         isActive == 1,
		AffectedStudents: studentCount,
		UpdatedAt:        now,
	}, nil
}
