package database

import (
	"database/sql"
	"fmt"
	"time"
)

// PersonaRepository 分身数据访问层
type PersonaRepository struct {
	db *sql.DB
}

// NewPersonaRepository 创建分身仓库
func NewPersonaRepository(db *sql.DB) *PersonaRepository {
	return &PersonaRepository{db: db}
}

// Create 创建分身
func (r *PersonaRepository) Create(persona *Persona) (int64, error) {
	now := time.Now()
	result, err := r.db.Exec(
		`INSERT INTO personas (user_id, role, nickname, school, description, avatar, is_active, is_public, bound_class_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		persona.UserID, persona.Role, persona.Nickname, persona.School, persona.Description,
		persona.Avatar, 1, persona.IsPublic, persona.BoundClassID, now, now,
	)
	if err != nil {
		return 0, fmt.Errorf("创建分身失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取分身ID失败: %w", err)
	}

	return id, nil
}

// CreateWithTx 在事务中创建分身
func (r *PersonaRepository) CreateWithTx(tx *sql.Tx, persona *Persona) (int64, error) {
	now := time.Now()
	result, err := tx.Exec(
		`INSERT INTO personas (user_id, role, nickname, school, description, avatar, is_active, is_public, bound_class_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		persona.UserID, persona.Role, persona.Nickname, persona.School, persona.Description,
		persona.Avatar, 1, persona.IsPublic, persona.BoundClassID, now, now,
	)
	if err != nil {
		return 0, fmt.Errorf("创建分身失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取分身ID失败: %w", err)
	}

	return id, nil
}

// GetByID 根据ID查询分身
func (r *PersonaRepository) GetByID(id int64) (*Persona, error) {
	persona := &Persona{}
	err := r.db.QueryRow(
		`SELECT id, user_id, role, nickname, school, description, avatar, is_active, COALESCE(is_public, 0), bound_class_id, created_at, updated_at
		 FROM personas WHERE id = ?`,
		id,
	).Scan(&persona.ID, &persona.UserID, &persona.Role, &persona.Nickname, &persona.School,
		&persona.Description, &persona.Avatar, &persona.IsActive, &persona.IsPublic, &persona.BoundClassID, &persona.CreatedAt, &persona.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询分身失败: %w", err)
	}

	return persona, nil
}

// ListByUserID 获取用户的所有分身列表（含统计信息）
func (r *PersonaRepository) ListByUserID(userID int64, roleFilter string) ([]PersonaListItem, error) {
	query := `SELECT p.id, p.role, p.nickname, p.school, p.description, p.is_active, COALESCE(p.is_public, 0), p.bound_class_id, p.created_at
		 FROM personas p WHERE p.user_id = ?`
	args := []interface{}{userID}

	if roleFilter != "" {
		query += ` AND p.role = ?`
		args = append(args, roleFilter)
	}

	query += ` ORDER BY p.created_at ASC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询分身列表失败: %w", err)
	}
	defer rows.Close()

	var items []PersonaListItem
	for rows.Next() {
		var item PersonaListItem
		var isActiveInt int
		var isPublicInt int
		var boundClassID sql.NullInt64
		if err := rows.Scan(&item.ID, &item.Role, &item.Nickname, &item.School, &item.Description,
			&isActiveInt, &isPublicInt, &boundClassID, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("扫描分身记录失败: %w", err)
		}
		item.IsActive = isActiveInt == 1
		item.IsPublic = isPublicInt == 1

		// 处理bound_class_id
		if boundClassID.Valid {
			item.BoundClassID = &boundClassID.Int64
			// 查询班级名称
			var className string
			err := r.db.QueryRow(`SELECT name FROM classes WHERE id = ?`, boundClassID.Int64).Scan(&className)
			if err == nil {
				item.BoundClassName = className
			}
		}

		// 查询统计信息
		if item.Role == "teacher" {
			r.db.QueryRow(`SELECT COUNT(*) FROM teacher_student_relations WHERE teacher_persona_id = ? AND status = 'approved'`, item.ID).Scan(&item.StudentCount)
			r.db.QueryRow(`SELECT COUNT(*) FROM knowledge_items WHERE persona_id = ? AND status = 'active'`, item.ID).Scan(&item.DocumentCount)
			r.db.QueryRow(`SELECT COUNT(*) FROM classes WHERE persona_id = ?`, item.ID).Scan(&item.ClassCount)
		} else {
			r.db.QueryRow(`SELECT COUNT(*) FROM teacher_student_relations WHERE student_persona_id = ? AND status = 'approved'`, item.ID).Scan(&item.TeacherCount)
		}

		items = append(items, item)
	}

	return items, nil
}

// Update 更新分身信息
func (r *PersonaRepository) Update(id int64, nickname, school, description string) error {
	result, err := r.db.Exec(
		`UPDATE personas SET nickname = ?, school = ?, description = ?, updated_at = ? WHERE id = ?`,
		nickname, school, description, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("更新分身失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("分身不存在: id=%d", id)
	}

	return nil
}

// SetActive 设置分身激活状态
func (r *PersonaRepository) SetActive(id int64, isActive int) error {
	result, err := r.db.Exec(
		`UPDATE personas SET is_active = ?, updated_at = ? WHERE id = ?`,
		isActive, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("更新分身状态失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("分身不存在: id=%d", id)
	}

	return nil
}

// CheckTeacherPersonaExists 检查教师分身是否存在（按昵称+学校唯一性）
func (r *PersonaRepository) CheckTeacherPersonaExists(nickname, school string, excludeID int64) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM personas WHERE nickname = ? AND school = ? AND role = 'teacher' AND id != ?`,
		nickname, school, excludeID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查教师分身是否存在失败: %w", err)
	}
	return count > 0, nil
}

// CountByUserID 统计用户的分身数量
func (r *PersonaRepository) CountByUserID(userID int64) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM personas WHERE user_id = ?`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("统计分身数量失败: %w", err)
	}
	return count, nil
}

// ListStudentPersonasByUserID 获取用户的学生分身列表（简化版，用于分享码加入时选择）
func (r *PersonaRepository) ListStudentPersonasByUserID(userID int64) ([]Persona, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, role, nickname, school, description, avatar, is_active, created_at, updated_at
		 FROM personas WHERE user_id = ? AND role = 'student' AND is_active = 1`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("查询学生分身列表失败: %w", err)
	}
	defer rows.Close()

	var personas []Persona
	for rows.Next() {
		p := Persona{}
		if err := rows.Scan(&p.ID, &p.UserID, &p.Role, &p.Nickname, &p.School, &p.Description,
			&p.Avatar, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("扫描分身记录失败: %w", err)
		}
		personas = append(personas, p)
	}

	return personas, nil
}

// ======================== V2.0 迭代3 新增方法 ========================

// DashboardData 仪表盘聚合数据
type DashboardData struct {
	Persona      *Persona               `json:"persona"`
	PendingCount int                    `json:"pending_count"`
	Classes      []ClassWithMemberCount `json:"classes"`
	LatestShare  map[string]interface{} `json:"latest_share"`
	Stats        map[string]interface{} `json:"stats"`
}

// GetPersonaDashboard 获取教师分身仪表盘聚合数据
func (r *PersonaRepository) GetPersonaDashboard(personaID int64) (*DashboardData, error) {
	// 1. 查询分身基本信息
	persona := &Persona{}
	err := r.db.QueryRow(
		`SELECT id, user_id, role, nickname, school, description, avatar, is_active, created_at, updated_at
		 FROM personas WHERE id = ?`,
		personaID,
	).Scan(&persona.ID, &persona.UserID, &persona.Role, &persona.Nickname, &persona.School,
		&persona.Description, &persona.Avatar, &persona.IsActive, &persona.CreatedAt, &persona.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("查询分身失败: %w", err)
	}

	dashboard := &DashboardData{
		Persona: persona,
	}

	// 2. 查询待审批数
	err = r.db.QueryRow(
		`SELECT COUNT(*) FROM teacher_student_relations WHERE teacher_persona_id = ? AND status = 'pending'`,
		personaID,
	).Scan(&dashboard.PendingCount)
	if err != nil {
		dashboard.PendingCount = 0
	}

	// 3. 查询班级列表（含成员数和is_active）
	classRows, err := r.db.Query(
		`SELECT c.id, c.name, c.description,
		        COALESCE((SELECT COUNT(*) FROM class_members WHERE class_id = c.id), 0) AS member_count,
		        COALESCE(c.is_active, 1),
		        c.created_at
		 FROM classes c WHERE c.persona_id = ?
		 ORDER BY c.created_at ASC`,
		personaID,
	)
	if err == nil {
		defer classRows.Close()
		for classRows.Next() {
			var item ClassWithMemberCount
			var isActiveInt int
			if err := classRows.Scan(&item.ID, &item.Name, &item.Description, &item.MemberCount, &isActiveInt, &item.CreatedAt); err != nil {
				continue
			}
			item.IsActive = isActiveInt == 1
			dashboard.Classes = append(dashboard.Classes, item)
		}
	}
	if dashboard.Classes == nil {
		dashboard.Classes = []ClassWithMemberCount{}
	}

	// 4. 查询最新分享码
	var shareID int64
	var shareCode, className string
	var classID sql.NullInt64
	var expiresAt sql.NullTime
	var maxUses, usedCount, shareIsActive int
	var shareCreatedAt time.Time

	err = r.db.QueryRow(
		`SELECT ps.id, ps.share_code, ps.class_id, COALESCE(c.name, ''), ps.expires_at, ps.max_uses, ps.used_count, ps.is_active, ps.created_at
		 FROM persona_shares ps
		 LEFT JOIN classes c ON ps.class_id = c.id
		 WHERE ps.teacher_persona_id = ? AND ps.is_active = 1
		 ORDER BY ps.created_at DESC LIMIT 1`,
		personaID,
	).Scan(&shareID, &shareCode, &classID, &className, &expiresAt, &maxUses, &usedCount, &shareIsActive, &shareCreatedAt)
	if err == nil {
		latestShare := map[string]interface{}{
			"id":         shareID,
			"share_code": shareCode,
			"class_name": className,
			"max_uses":   maxUses,
			"used_count": usedCount,
			"is_active":  shareIsActive == 1,
			"created_at": shareCreatedAt,
		}
		if classID.Valid {
			latestShare["class_id"] = classID.Int64
		}
		if expiresAt.Valid {
			latestShare["expires_at"] = expiresAt.Time
		}
		dashboard.LatestShare = latestShare
	}

	// 5. 统计信息
	var totalStudents, totalDocuments, totalClasses int
	r.db.QueryRow(`SELECT COUNT(*) FROM teacher_student_relations WHERE teacher_persona_id = ? AND status = 'approved'`, personaID).Scan(&totalStudents)
	r.db.QueryRow(`SELECT COUNT(*) FROM knowledge_items WHERE persona_id = ? AND status = 'active'`, personaID).Scan(&totalDocuments)
	r.db.QueryRow(`SELECT COUNT(*) FROM classes WHERE persona_id = ?`, personaID).Scan(&totalClasses)

	dashboard.Stats = map[string]interface{}{
		"total_students":  totalStudents,
		"total_documents": totalDocuments,
		"total_classes":   totalClasses,
	}

	return dashboard, nil
}

// ======================== V2.0 迭代4 新增方法 ========================

// ListPublicPersonas 获取公开的教师分身列表（广场用）
// 排除当前学生已有 approved 关系的教师分身
func (r *PersonaRepository) ListPublicPersonas(studentPersonaID int64, keyword string, offset, limit int) ([]MarketplacePersona, int, error) {
	// 基础查询条件：公开 + 启用 + 教师角色
	baseWhere := `p.is_public = 1 AND p.is_active = 1 AND p.role = 'teacher'`
	args := []interface{}{}

	// 排除已有 approved 关系的教师分身
	if studentPersonaID > 0 {
		baseWhere += ` AND p.id NOT IN (SELECT teacher_persona_id FROM teacher_student_relations WHERE student_persona_id = ? AND status = 'approved')`
		args = append(args, studentPersonaID)
	}

	// 关键词搜索
	if keyword != "" {
		baseWhere += ` AND (p.nickname LIKE ? OR p.school LIKE ?)`
		kw := "%" + keyword + "%"
		args = append(args, kw, kw)
	}

	// 查询总数
	var total int
	countQuery := `SELECT COUNT(*) FROM personas p WHERE ` + baseWhere
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询广场分身总数失败: %w", err)
	}

	// 查询列表
	listQuery := `SELECT p.id, p.nickname, p.school, p.description,
		COALESCE((SELECT COUNT(*) FROM teacher_student_relations WHERE teacher_persona_id = p.id AND status = 'approved'), 0) AS student_count,
COALESCE((SELECT COUNT(*) FROM knowledge_items WHERE persona_id = p.id AND status = 'active'), 0) AS document_count
		FROM personas p WHERE ` + baseWhere + ` ORDER BY p.created_at DESC LIMIT ? OFFSET ?`
	listArgs := append(args, limit, offset)

	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询广场分身列表失败: %w", err)
	}
	defer rows.Close()

	var items []MarketplacePersona
	for rows.Next() {
		var item MarketplacePersona
		if err := rows.Scan(&item.ID, &item.Nickname, &item.School, &item.Description,
			&item.StudentCount, &item.DocumentCount); err != nil {
			return nil, 0, fmt.Errorf("扫描广场分身记录失败: %w", err)
		}

		// 查询当前学生对该教师的申请状态
		if studentPersonaID > 0 {
			var status string
			err := r.db.QueryRow(
				`SELECT status FROM teacher_student_relations WHERE teacher_persona_id = ? AND student_persona_id = ? ORDER BY created_at DESC LIMIT 1`,
				item.ID, studentPersonaID,
			).Scan(&status)
			if err == nil {
				item.ApplicationStatus = status
			}
		}

		items = append(items, item)
	}

	return items, total, nil
}

// UpdateVisibility 设置分身公开/私有
func (r *PersonaRepository) UpdateVisibility(personaID int64, isPublic int) error {
	result, err := r.db.Exec(
		`UPDATE personas SET is_public = ?, updated_at = ? WHERE id = ?`,
		isPublic, time.Now(), personaID,
	)
	if err != nil {
		return fmt.Errorf("更新分身公开状态失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("分身不存在: id=%d", personaID)
	}

	return nil
}

// SearchStudentPersonas 搜索已注册的学生分身（教师定向邀请用）
// V2.0 迭代11 M4：排除自测学生
func (r *PersonaRepository) SearchStudentPersonas(keyword string, offset, limit int) ([]Persona, int, error) {
	kw := "%" + keyword + "%"

	// 查询总数（排除自测学生）
	var total int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM personas p
		 JOIN users u ON p.user_id = u.id
		 WHERE p.role = 'student' AND p.is_active = 1 AND p.nickname LIKE ?
		 AND COALESCE(u.is_test_student, 0) = 0`,
		kw,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询学生分身总数失败: %w", err)
	}

	// 查询列表（排除自测学生）
	rows, err := r.db.Query(
		`SELECT p.id, p.user_id, p.role, p.nickname, p.school, p.description, p.avatar, p.is_active, p.created_at, p.updated_at
		 FROM personas p
		 JOIN users u ON p.user_id = u.id
		 WHERE p.role = 'student' AND p.is_active = 1 AND p.nickname LIKE ?
		 AND COALESCE(u.is_test_student, 0) = 0
		 ORDER BY p.created_at DESC LIMIT ? OFFSET ?`,
		kw, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("查询学生分身列表失败: %w", err)
	}
	defer rows.Close()

	var personas []Persona
	for rows.Next() {
		p := Persona{}
		if err := rows.Scan(&p.ID, &p.UserID, &p.Role, &p.Nickname, &p.School, &p.Description,
			&p.Avatar, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描学生分身记录失败: %w", err)
		}
		personas = append(personas, p)
	}

	return personas, total, nil
}

// ======================== V2.0 迭代11 M4 自测学生方法 ========================

// GetStudentPersonaByUserID 根据用户ID获取学生分身
func (r *PersonaRepository) GetStudentPersonaByUserID(userID int64) (*Persona, error) {
	persona := &Persona{}
	err := r.db.QueryRow(
		`SELECT id, user_id, role, nickname, school, description, avatar, is_active, created_at, updated_at
		 FROM personas WHERE user_id = ? AND role = 'student'`,
		userID,
	).Scan(&persona.ID, &persona.UserID, &persona.Role, &persona.Nickname, &persona.School,
		&persona.Description, &persona.Avatar, &persona.IsActive, &persona.CreatedAt, &persona.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询学生分身失败: %w", err)
	}

	return persona, nil
}
