package database

import (
	"database/sql"
	"fmt"
	"time"
)

// ==================== RelationRepository ====================

type RelationRepository struct {
	db *sql.DB
}

func NewRelationRepository(db *sql.DB) *RelationRepository {
	return &RelationRepository{db: db}
}

func (r *RelationRepository) Create(rel *TeacherStudentRelation) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO teacher_student_relations (teacher_id, student_id, status, initiated_by, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		rel.TeacherID, rel.StudentID, rel.Status, rel.InitiatedBy,
		time.Now(), time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("创建师生关系失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取关系ID失败: %w", err)
	}

	return id, nil
}

func (r *RelationRepository) GetByID(id int64) (*TeacherStudentRelation, error) {
	rel := &TeacherStudentRelation{}
	err := r.db.QueryRow(
		`SELECT id, teacher_id, student_id, COALESCE(teacher_persona_id, 0), COALESCE(student_persona_id, 0), status, initiated_by, COALESCE(is_active, 1), created_at, updated_at
		 FROM teacher_student_relations WHERE id = ?`,
		id,
	).Scan(&rel.ID, &rel.TeacherID, &rel.StudentID, &rel.TeacherPersonaID, &rel.StudentPersonaID, &rel.Status, &rel.InitiatedBy, &rel.IsActive, &rel.CreatedAt, &rel.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询师生关系失败: %w", err)
	}

	return rel, nil
}

func (r *RelationRepository) GetByTeacherAndStudent(teacherID, studentID int64) (*TeacherStudentRelation, error) {
	rel := &TeacherStudentRelation{}
	err := r.db.QueryRow(
		`SELECT id, teacher_id, student_id, COALESCE(teacher_persona_id, 0), COALESCE(student_persona_id, 0), status, initiated_by, COALESCE(is_active, 1), created_at, updated_at
		 FROM teacher_student_relations WHERE teacher_id = ? AND student_id = ?`,
		teacherID, studentID,
	).Scan(&rel.ID, &rel.TeacherID, &rel.StudentID, &rel.TeacherPersonaID, &rel.StudentPersonaID, &rel.Status, &rel.InitiatedBy, &rel.IsActive, &rel.CreatedAt, &rel.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询师生关系失败: %w", err)
	}

	return rel, nil
}

func (r *RelationRepository) UpdateStatus(id int64, status string) error {
	result, err := r.db.Exec(
		`UPDATE teacher_student_relations SET status = ?, updated_at = ? WHERE id = ?`,
		status, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("更新师生关系状态失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("关系不存在: id=%d", id)
	}

	return nil
}

func (r *RelationRepository) ListByTeacher(teacherID int64, status string, offset, limit int) ([]*TeacherStudentRelation, int, error) {
	// 构建查询条件
	countQuery := `SELECT COUNT(*) FROM teacher_student_relations WHERE teacher_id = ?`
	listQuery := `SELECT id, teacher_id, student_id, status, initiated_by, created_at, updated_at
		 FROM teacher_student_relations WHERE teacher_id = ?`
	args := []interface{}{teacherID}

	if status != "" {
		countQuery += ` AND status = ?`
		listQuery += ` AND status = ?`
		args = append(args, status)
	}

	// 查询总数
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询教师关系总数失败: %w", err)
	}

	// 查询列表
	listQuery += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	listArgs := append(args, limit, offset)
	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询教师关系列表失败: %w", err)
	}
	defer rows.Close()

	var rels []*TeacherStudentRelation
	for rows.Next() {
		rel := &TeacherStudentRelation{}
		if err := rows.Scan(&rel.ID, &rel.TeacherID, &rel.StudentID, &rel.Status, &rel.InitiatedBy, &rel.CreatedAt, &rel.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描关系记录失败: %w", err)
		}
		rels = append(rels, rel)
	}

	return rels, total, nil
}

func (r *RelationRepository) ListByStudent(studentID int64, status string, offset, limit int) ([]*TeacherStudentRelation, int, error) {
	// 构建查询条件
	countQuery := `SELECT COUNT(*) FROM teacher_student_relations WHERE student_id = ?`
	listQuery := `SELECT id, teacher_id, student_id, status, initiated_by, created_at, updated_at
		 FROM teacher_student_relations WHERE student_id = ?`
	args := []interface{}{studentID}

	if status != "" {
		countQuery += ` AND status = ?`
		listQuery += ` AND status = ?`
		args = append(args, status)
	}

	// 查询总数
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询学生关系总数失败: %w", err)
	}

	// 查询列表
	listQuery += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	listArgs := append(args, limit, offset)
	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询学生关系列表失败: %w", err)
	}
	defer rows.Close()

	var rels []*TeacherStudentRelation
	for rows.Next() {
		rel := &TeacherStudentRelation{}
		if err := rows.Scan(&rel.ID, &rel.TeacherID, &rel.StudentID, &rel.Status, &rel.InitiatedBy, &rel.CreatedAt, &rel.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描关系记录失败: %w", err)
		}
		rels = append(rels, rel)
	}

	return rels, total, nil
}

// IsApproved 检查师生关系是否已授权
func (r *RelationRepository) IsApproved(teacherID, studentID int64) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM teacher_student_relations WHERE teacher_id = ? AND student_id = ? AND status = 'approved'`,
		teacherID, studentID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("查询授权关系失败: %w", err)
	}
	return count > 0, nil
}

// ListByTeacherWithStudent 教师视角：JOIN users 获取学生昵称
func (r *RelationRepository) ListByTeacherWithStudent(teacherID int64, status string, offset, limit int) ([]RelationWithStudent, int, error) {
	// 构建查询条件
	countQuery := `SELECT COUNT(*) FROM teacher_student_relations WHERE teacher_id = ?`
	listQuery := `SELECT r.id, r.student_id, COALESCE(u.nickname, '') AS student_nickname, COALESCE(r.teacher_persona_id, 0), COALESCE(r.student_persona_id, 0), r.status, r.initiated_by, r.created_at
		 FROM teacher_student_relations r
		 LEFT JOIN users u ON r.student_id = u.id
		 WHERE r.teacher_id = ?`
	args := []interface{}{teacherID}

	if status != "" {
		countQuery += ` AND status = ?`
		listQuery += ` AND r.status = ?`
		args = append(args, status)
	}

	// 查询总数
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询教师关系总数失败: %w", err)
	}

	// 查询列表
	listQuery += ` ORDER BY r.created_at DESC LIMIT ? OFFSET ?`
	listArgs := append(args, limit, offset)
	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询教师关系列表失败: %w", err)
	}
	defer rows.Close()

	var rels []RelationWithStudent
	for rows.Next() {
		var rel RelationWithStudent
		if err := rows.Scan(&rel.ID, &rel.StudentID, &rel.StudentNickname, &rel.TeacherPersonaID, &rel.StudentPersonaID, &rel.Status, &rel.InitiatedBy, &rel.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描关系记录失败: %w", err)
		}
		rels = append(rels, rel)
	}

	return rels, total, nil
}

// ListByStudentWithTeacher 学生视角：JOIN users 获取教师信息
func (r *RelationRepository) ListByStudentWithTeacher(studentID int64, status string, offset, limit int) ([]RelationWithTeacher, int, error) {
	// 构建查询条件
	countQuery := `SELECT COUNT(*) FROM teacher_student_relations WHERE student_id = ?`
	listQuery := `SELECT r.id, r.teacher_id, COALESCE(u.nickname, '') AS teacher_nickname,
			COALESCE(u.school, '') AS teacher_school, COALESCE(u.description, '') AS teacher_description,
			COALESCE(r.teacher_persona_id, 0), COALESCE(r.student_persona_id, 0),
			r.status, r.initiated_by, r.created_at
		 FROM teacher_student_relations r
		 LEFT JOIN users u ON r.teacher_id = u.id
		 WHERE r.student_id = ?`
	args := []interface{}{studentID}

	if status != "" {
		countQuery += ` AND status = ?`
		listQuery += ` AND r.status = ?`
		args = append(args, status)
	}

	// 查询总数
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询学生关系总数失败: %w", err)
	}

	// 查询列表
	listQuery += ` ORDER BY r.created_at DESC LIMIT ? OFFSET ?`
	listArgs := append(args, limit, offset)
	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询学生关系列表失败: %w", err)
	}
	defer rows.Close()

	var rels []RelationWithTeacher
	for rows.Next() {
		var rel RelationWithTeacher
		if err := rows.Scan(&rel.ID, &rel.TeacherID, &rel.TeacherNickname, &rel.TeacherSchool, &rel.TeacherDescription, &rel.TeacherPersonaID, &rel.StudentPersonaID, &rel.Status, &rel.InitiatedBy, &rel.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描关系记录失败: %w", err)
		}
		rels = append(rels, rel)
	}

	return rels, total, nil
}

// ListByTeacherPersona 教师分身维度：获取关系列表（含学生信息）
func (r *RelationRepository) ListByTeacherPersona(personaID int64, status string, offset, limit int) ([]RelationWithStudent, int, error) {
	countQuery := `SELECT COUNT(*) FROM teacher_student_relations WHERE teacher_persona_id = ?`
	// 优先从 personas 表获取昵称，回退到 users 表
	// last_chat_time: 优先用 student_persona_id 查询，回退到 student_id + teacher_persona_id
	listQuery := `SELECT r.id, r.student_id, COALESCE(p.nickname, u.nickname, '') AS student_nickname,
			COALESCE(r.teacher_persona_id, 0), COALESCE(r.student_persona_id, 0),
			r.status, r.initiated_by, COALESCE(r.is_active, 1), r.created_at,
			COALESCE(
				(SELECT MAX(created_at) FROM conversations WHERE student_persona_id = r.student_persona_id AND teacher_persona_id = r.teacher_persona_id AND r.student_persona_id > 0),
				(SELECT MAX(created_at) FROM conversations WHERE student_id = r.student_id AND teacher_persona_id = r.teacher_persona_id)
			) AS last_chat_time
		 FROM teacher_student_relations r
		 LEFT JOIN users u ON r.student_id = u.id
		 LEFT JOIN personas p ON r.student_persona_id = p.id AND r.student_persona_id > 0
		 WHERE r.teacher_persona_id = ?`
	args := []interface{}{personaID}

	if status != "" {
		countQuery += ` AND status = ?`
		listQuery += ` AND r.status = ?`
		args = append(args, status)
	}

	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询教师分身关系总数失败: %w", err)
	}

	listQuery += ` ORDER BY CASE WHEN last_chat_time IS NULL THEN 1 ELSE 0 END, last_chat_time DESC, r.created_at DESC LIMIT ? OFFSET ?`
	listArgs := append(args, limit, offset)
	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询教师分身关系列表失败: %w", err)
	}
	defer rows.Close()

	var rels []RelationWithStudent
	for rows.Next() {
		var rel RelationWithStudent
		var isActiveInt int
		if err := rows.Scan(&rel.ID, &rel.StudentID, &rel.StudentNickname, &rel.TeacherPersonaID, &rel.StudentPersonaID, &rel.Status, &rel.InitiatedBy, &isActiveInt, &rel.CreatedAt, &rel.LastChatTime); err != nil {
			return nil, 0, fmt.Errorf("扫描关系记录失败: %w", err)
		}
		rel.IsActive = isActiveInt == 1
		rels = append(rels, rel)
	}

	return rels, total, nil
}

// ListByStudentPersona 学生分身维度：获取关系列表（含教师信息）
func (r *RelationRepository) ListByStudentPersona(personaID int64, status string, offset, limit int) ([]RelationWithTeacher, int, error) {
	countQuery := `SELECT COUNT(*) FROM teacher_student_relations WHERE student_persona_id = ?`
	listQuery := `SELECT r.id, r.teacher_id, COALESCE(u.nickname, '') AS teacher_nickname,
			COALESCE(u.school, '') AS teacher_school, COALESCE(u.description, '') AS teacher_description,
			COALESCE(r.teacher_persona_id, 0), COALESCE(r.student_persona_id, 0),
			r.status, r.initiated_by, COALESCE(r.is_active, 1), r.created_at
		 FROM teacher_student_relations r
		 LEFT JOIN users u ON r.teacher_id = u.id
		 WHERE r.student_persona_id = ?`
	args := []interface{}{personaID}

	if status != "" {
		countQuery += ` AND status = ?`
		listQuery += ` AND r.status = ?`
		args = append(args, status)
	}

	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询学生分身关系总数失败: %w", err)
	}

	listQuery += ` ORDER BY r.created_at DESC LIMIT ? OFFSET ?`
	listArgs := append(args, limit, offset)
	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询学生分身关系列表失败: %w", err)
	}
	defer rows.Close()

	var rels []RelationWithTeacher
	for rows.Next() {
		var rel RelationWithTeacher
		var isActiveInt int
		if err := rows.Scan(&rel.ID, &rel.TeacherID, &rel.TeacherNickname, &rel.TeacherSchool, &rel.TeacherDescription, &rel.TeacherPersonaID, &rel.StudentPersonaID, &rel.Status, &rel.InitiatedBy, &isActiveInt, &rel.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描关系记录失败: %w", err)
		}
		rel.IsActive = isActiveInt == 1
		rels = append(rels, rel)
	}

	return rels, total, nil
}

// GetByPersonas 根据教师分身ID和学生分身ID查询关系
func (r *RelationRepository) GetByPersonas(teacherPersonaID, studentPersonaID int64) (*TeacherStudentRelation, error) {
	rel := &TeacherStudentRelation{}
	err := r.db.QueryRow(
		`SELECT id, teacher_id, student_id, teacher_persona_id, student_persona_id, status, initiated_by, COALESCE(is_active, 1), created_at, updated_at
		 FROM teacher_student_relations WHERE teacher_persona_id = ? AND student_persona_id = ?`,
		teacherPersonaID, studentPersonaID,
	).Scan(&rel.ID, &rel.TeacherID, &rel.StudentID, &rel.TeacherPersonaID, &rel.StudentPersonaID,
		&rel.Status, &rel.InitiatedBy, &rel.IsActive, &rel.CreatedAt, &rel.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询师生关系失败: %w", err)
	}

	return rel, nil
}

// IsApprovedByPersonas 检查分身维度的师生关系是否已授权
func (r *RelationRepository) IsApprovedByPersonas(teacherPersonaID, studentPersonaID int64) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM teacher_student_relations WHERE teacher_persona_id = ? AND student_persona_id = ? AND status = 'approved'`,
		teacherPersonaID, studentPersonaID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("查询授权关系失败: %w", err)
	}
	return count > 0, nil
}

// CreateWithPersonas 创建带分身维度的师生关系
func (r *RelationRepository) CreateWithPersonas(teacherID, studentID, teacherPersonaID, studentPersonaID int64, status, initiatedBy string) (int64, error) {
	now := time.Now()
	result, err := r.db.Exec(
		`INSERT INTO teacher_student_relations (teacher_id, student_id, teacher_persona_id, student_persona_id, status, initiated_by, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		teacherID, studentID, teacherPersonaID, studentPersonaID, status, initiatedBy, now, now,
	)
	if err != nil {
		return 0, fmt.Errorf("创建师生关系失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取关系ID失败: %w", err)
	}

	return id, nil
}

// ApproveWithDetails 审批通过并保存评语和班级
func (r *RelationRepository) ApproveWithDetails(id int64, comment string, classID *int64) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE teacher_student_relations SET status = 'approved', comment = ?, class_id = ?, updated_at = ? WHERE id = ?`,
		comment, classID, now, id,
	)
	if err != nil {
		return fmt.Errorf("审批更新失败: %w", err)
	}
	return nil
}

// ======================== V2.0 迭代3 新增方法 ========================

// ToggleRelationResult 关系启停结果
type ToggleRelationResult struct {
	ID               int64     `json:"id"`
	StudentPersonaID int64     `json:"student_persona_id"`
	StudentNickname  string    `json:"student_nickname"`
	IsActive         bool      `json:"is_active"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ToggleRelation 启用/停用师生关系
func (r *RelationRepository) ToggleRelation(relationID int64, isActive int) (*ToggleRelationResult, error) {
	now := time.Now()

	// 更新 is_active
	result, err := r.db.Exec(
		`UPDATE teacher_student_relations SET is_active = ?, updated_at = ? WHERE id = ?`,
		isActive, now, relationID,
	)
	if err != nil {
		return nil, fmt.Errorf("更新师生关系状态失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return nil, fmt.Errorf("关系不存在: id=%d", relationID)
	}

	// 查询关系信息
	var studentPersonaID int64
	var studentNickname string
	r.db.QueryRow(
		`SELECT COALESCE(r.student_persona_id, 0), COALESCE(u.nickname, '')
		 FROM teacher_student_relations r
		 LEFT JOIN users u ON r.student_id = u.id
		 WHERE r.id = ?`, relationID,
	).Scan(&studentPersonaID, &studentNickname)

	return &ToggleRelationResult{
		ID:               relationID,
		StudentPersonaID: studentPersonaID,
		StudentNickname:  studentNickname,
		IsActive:         isActive == 1,
		UpdatedAt:        now,
	}, nil
}

// ChatPermissionError 对话权限错误
type ChatPermissionError struct {
	Code    int
	Message string
}

func (e *ChatPermissionError) Error() string {
	return e.Message
}

// CheckChatPermission 检查学生对话权限（启停状态检查）
// 返回 nil 表示允许对话，返回 ChatPermissionError 表示不允许
func (r *RelationRepository) CheckChatPermission(teacherPersonaID, studentPersonaID int64) *ChatPermissionError {
	// 1. 查询师生关系
	var relStatus string
	var relIsActive int
	err := r.db.QueryRow(
		`SELECT status, COALESCE(is_active, 1) FROM teacher_student_relations
		 WHERE teacher_persona_id = ? AND student_persona_id = ?`,
		teacherPersonaID, studentPersonaID,
	).Scan(&relStatus, &relIsActive)
	if err != nil {
		return &ChatPermissionError{Code: 40007, Message: "未获得该教师授权，请先申请"}
	}

	// 检查 status
	if relStatus != "approved" {
		return &ChatPermissionError{Code: 40007, Message: "未获得该教师授权，请先申请"}
	}

	// 2. 检查 relation.is_active
	if relIsActive != 1 {
		return &ChatPermissionError{Code: 40027, Message: "您的访问权限已关闭，请联系教师"}
	}

	// 3. 检查教师分身 is_active
	var personaIsActive int
	err = r.db.QueryRow(
		`SELECT COALESCE(is_active, 1) FROM personas WHERE id = ?`,
		teacherPersonaID,
	).Scan(&personaIsActive)
	if err != nil || personaIsActive != 1 {
		return &ChatPermissionError{Code: 40025, Message: "教师分身已停用，无法发起对话"}
	}

	// 4. 查询学生在该教师分身下的班级，检查班级 is_active
	rows, err := r.db.Query(
		`SELECT c.is_active FROM class_members cm
		 JOIN classes c ON cm.class_id = c.id
		 WHERE c.persona_id = ? AND cm.student_persona_id = ?`,
		teacherPersonaID, studentPersonaID,
	)
	if err == nil {
		defer rows.Close()
		hasClass := false
		allInactive := true
		for rows.Next() {
			hasClass = true
			var classIsActive int
			if err := rows.Scan(&classIsActive); err == nil {
				if classIsActive == 1 {
					allInactive = false
					break
				}
			}
		}
		// 如果学生有班级，且所有班级都停用了
		if hasClass && allInactive {
			return &ChatPermissionError{Code: 40026, Message: "您所在的班级已停用，无法发起对话"}
		}
	}

	// 5. 所有检查通过
	return nil
}
