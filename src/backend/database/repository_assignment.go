package database

import (
	"database/sql"
	"fmt"
	"time"
)

// ==================== AssignmentRepository ====================

type AssignmentRepository struct {
	db *sql.DB
}

func NewAssignmentRepository(db *sql.DB) *AssignmentRepository {
	return &AssignmentRepository{db: db}
}

func (r *AssignmentRepository) Create(asg *Assignment) (int64, error) {
	now := time.Now()
	result, err := r.db.Exec(
		`INSERT INTO assignments (student_id, teacher_id, title, content, file_path, file_type, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		asg.StudentID, asg.TeacherID, asg.Title, asg.Content, asg.FilePath, asg.FileType, asg.Status,
		now, now,
	)
	if err != nil {
		return 0, fmt.Errorf("创建作业失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取作业ID失败: %w", err)
	}

	return id, nil
}

func (r *AssignmentRepository) GetByID(id int64) (*Assignment, error) {
	asg := &Assignment{}
	err := r.db.QueryRow(
		`SELECT id, student_id, teacher_id, title, COALESCE(content, '') AS content,
		        COALESCE(file_path, '') AS file_path, COALESCE(file_type, '') AS file_type,
		        status, created_at, updated_at
		 FROM assignments WHERE id = ?`,
		id,
	).Scan(&asg.ID, &asg.StudentID, &asg.TeacherID, &asg.Title, &asg.Content,
		&asg.FilePath, &asg.FileType, &asg.Status, &asg.CreatedAt, &asg.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询作业失败: %w", err)
	}

	return asg, nil
}

func (r *AssignmentRepository) ListByTeacher(teacherID int64, studentID *int64, status string, offset, limit int) ([]*Assignment, int, error) {
	// 构建查询条件
	countQuery := `SELECT COUNT(*) FROM assignments WHERE teacher_id = ?`
	listQuery := `SELECT id, student_id, teacher_id, title, COALESCE(content, '') AS content,
		        COALESCE(file_path, '') AS file_path, COALESCE(file_type, '') AS file_type,
		        status, created_at, updated_at
		 FROM assignments WHERE teacher_id = ?`
	args := []interface{}{teacherID}

	if studentID != nil {
		countQuery += ` AND student_id = ?`
		listQuery += ` AND student_id = ?`
		args = append(args, *studentID)
	}
	if status != "" {
		countQuery += ` AND status = ?`
		listQuery += ` AND status = ?`
		args = append(args, status)
	}

	// 查询总数
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询作业总数失败: %w", err)
	}

	// 查询列表
	listQuery += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	listArgs := append(args, limit, offset)
	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询作业列表失败: %w", err)
	}
	defer rows.Close()

	var assignments []*Assignment
	for rows.Next() {
		asg := &Assignment{}
		if err := rows.Scan(&asg.ID, &asg.StudentID, &asg.TeacherID, &asg.Title, &asg.Content,
			&asg.FilePath, &asg.FileType, &asg.Status, &asg.CreatedAt, &asg.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描作业记录失败: %w", err)
		}
		assignments = append(assignments, asg)
	}

	return assignments, total, nil
}

// ListByTeacherWithDetails 教师视角：JOIN users 获取 nickname、review_count、has_file
func (r *AssignmentRepository) ListByTeacherWithDetails(teacherID int64, studentID *int64, status string, offset, limit int) ([]AssignmentListItem, int, error) {
	// 构建查询条件
	countQuery := `SELECT COUNT(*) FROM assignments WHERE teacher_id = ?`
	listQuery := `SELECT a.id, a.student_id, COALESCE(s.nickname, '') AS student_nickname,
		        a.teacher_id, COALESCE(t.nickname, '') AS teacher_nickname,
		        a.title, a.status,
		        CASE WHEN COALESCE(a.file_path, '') != '' THEN 1 ELSE 0 END AS has_file,
		        (SELECT COUNT(*) FROM assignment_reviews WHERE assignment_id = a.id) AS review_count,
		        a.created_at
		 FROM assignments a
		 LEFT JOIN users s ON a.student_id = s.id
		 LEFT JOIN users t ON a.teacher_id = t.id
		 WHERE a.teacher_id = ?`
	args := []interface{}{teacherID}
	listArgs := []interface{}{teacherID}

	if studentID != nil {
		countQuery += ` AND student_id = ?`
		listQuery += ` AND a.student_id = ?`
		args = append(args, *studentID)
		listArgs = append(listArgs, *studentID)
	}
	if status != "" {
		countQuery += ` AND status = ?`
		listQuery += ` AND a.status = ?`
		args = append(args, status)
		listArgs = append(listArgs, status)
	}

	// 查询总数
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询作业总数失败: %w", err)
	}

	// 查询列表
	listQuery += ` ORDER BY a.created_at DESC LIMIT ? OFFSET ?`
	listArgs = append(listArgs, limit, offset)
	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询作业列表失败: %w", err)
	}
	defer rows.Close()

	var items []AssignmentListItem
	for rows.Next() {
		var item AssignmentListItem
		var hasFileInt int
		if err := rows.Scan(&item.ID, &item.StudentID, &item.StudentNickname,
			&item.TeacherID, &item.TeacherNickname,
			&item.Title, &item.Status, &hasFileInt, &item.ReviewCount,
			&item.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描作业记录失败: %w", err)
		}
		item.HasFile = hasFileInt == 1
		items = append(items, item)
	}

	return items, total, nil
}

func (r *AssignmentRepository) ListByStudent(studentID int64, teacherID *int64, status string, offset, limit int) ([]*Assignment, int, error) {
	// 构建查询条件
	countQuery := `SELECT COUNT(*) FROM assignments WHERE student_id = ?`
	listQuery := `SELECT id, student_id, teacher_id, title, COALESCE(content, '') AS content,
		        COALESCE(file_path, '') AS file_path, COALESCE(file_type, '') AS file_type,
		        status, created_at, updated_at
		 FROM assignments WHERE student_id = ?`
	args := []interface{}{studentID}

	if teacherID != nil {
		countQuery += ` AND teacher_id = ?`
		listQuery += ` AND teacher_id = ?`
		args = append(args, *teacherID)
	}
	if status != "" {
		countQuery += ` AND status = ?`
		listQuery += ` AND status = ?`
		args = append(args, status)
	}

	// 查询总数
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询作业总数失败: %w", err)
	}

	// 查询列表
	listQuery += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	listArgs := append(args, limit, offset)
	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询作业列表失败: %w", err)
	}
	defer rows.Close()

	var assignments []*Assignment
	for rows.Next() {
		asg := &Assignment{}
		if err := rows.Scan(&asg.ID, &asg.StudentID, &asg.TeacherID, &asg.Title, &asg.Content,
			&asg.FilePath, &asg.FileType, &asg.Status, &asg.CreatedAt, &asg.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描作业记录失败: %w", err)
		}
		assignments = append(assignments, asg)
	}

	return assignments, total, nil
}

// ListByStudentWithDetails 学生视角：JOIN users 获取 nickname、review_count、has_file
func (r *AssignmentRepository) ListByStudentWithDetails(studentID int64, teacherID *int64, status string, offset, limit int) ([]AssignmentListItem, int, error) {
	// 构建查询条件
	countQuery := `SELECT COUNT(*) FROM assignments WHERE student_id = ?`
	listQuery := `SELECT a.id, a.student_id, COALESCE(s.nickname, '') AS student_nickname,
		        a.teacher_id, COALESCE(t.nickname, '') AS teacher_nickname,
		        a.title, a.status,
		        CASE WHEN COALESCE(a.file_path, '') != '' THEN 1 ELSE 0 END AS has_file,
		        (SELECT COUNT(*) FROM assignment_reviews WHERE assignment_id = a.id) AS review_count,
		        a.created_at
		 FROM assignments a
		 LEFT JOIN users s ON a.student_id = s.id
		 LEFT JOIN users t ON a.teacher_id = t.id
		 WHERE a.student_id = ?`
	args := []interface{}{studentID}
	listArgs := []interface{}{studentID}

	if teacherID != nil {
		countQuery += ` AND teacher_id = ?`
		listQuery += ` AND a.teacher_id = ?`
		args = append(args, *teacherID)
		listArgs = append(listArgs, *teacherID)
	}
	if status != "" {
		countQuery += ` AND status = ?`
		listQuery += ` AND a.status = ?`
		args = append(args, status)
		listArgs = append(listArgs, status)
	}

	// 查询总数
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询作业总数失败: %w", err)
	}

	// 查询列表
	listQuery += ` ORDER BY a.created_at DESC LIMIT ? OFFSET ?`
	listArgs = append(listArgs, limit, offset)
	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询作业列表失败: %w", err)
	}
	defer rows.Close()

	var items []AssignmentListItem
	for rows.Next() {
		var item AssignmentListItem
		var hasFileInt int
		if err := rows.Scan(&item.ID, &item.StudentID, &item.StudentNickname,
			&item.TeacherID, &item.TeacherNickname,
			&item.Title, &item.Status, &hasFileInt, &item.ReviewCount,
			&item.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描作业记录失败: %w", err)
		}
		item.HasFile = hasFileInt == 1
		items = append(items, item)
	}

	return items, total, nil
}

func (r *AssignmentRepository) UpdateStatus(id int64, status string) error {
	result, err := r.db.Exec(
		`UPDATE assignments SET status = ?, updated_at = ? WHERE id = ?`,
		status, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("更新作业状态失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("作业不存在: id=%d", id)
	}

	return nil
}

// ==================== 分身维度方法 ====================

// CreateWithPersonas 创建带分身维度的作业
func (r *AssignmentRepository) CreateWithPersonas(asg *Assignment) (int64, error) {
	now := time.Now()
	result, err := r.db.Exec(
		`INSERT INTO assignments (student_id, teacher_id, teacher_persona_id, student_persona_id, title, content, file_path, file_type, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		asg.StudentID, asg.TeacherID, asg.TeacherPersonaID, asg.StudentPersonaID,
		asg.Title, asg.Content, asg.FilePath, asg.FileType, asg.Status, now, now,
	)
	if err != nil {
		return 0, fmt.Errorf("创建作业失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取作业ID失败: %w", err)
	}

	return id, nil
}

// ListByTeacherPersona 按教师分身查询作业列表
func (r *AssignmentRepository) ListByTeacherPersona(teacherPersonaID int64, studentPersonaID *int64, status string, offset, limit int) ([]AssignmentListItem, int, error) {
	countQuery := `SELECT COUNT(*) FROM assignments WHERE teacher_persona_id = ?`
	listQuery := `SELECT a.id, a.student_id, COALESCE(s.nickname, '') AS student_nickname,
		        a.teacher_id, COALESCE(t.nickname, '') AS teacher_nickname,
		        COALESCE(a.teacher_persona_id, 0), COALESCE(a.student_persona_id, 0),
		        a.title, a.status,
		        CASE WHEN COALESCE(a.file_path, '') != '' THEN 1 ELSE 0 END AS has_file,
		        (SELECT COUNT(*) FROM assignment_reviews WHERE assignment_id = a.id) AS review_count,
		        a.created_at
		 FROM assignments a
		 LEFT JOIN users s ON a.student_id = s.id
		 LEFT JOIN users t ON a.teacher_id = t.id
		 WHERE a.teacher_persona_id = ?`
	args := []interface{}{teacherPersonaID}
	listArgs := []interface{}{teacherPersonaID}

	if studentPersonaID != nil {
		countQuery += ` AND student_persona_id = ?`
		listQuery += ` AND a.student_persona_id = ?`
		args = append(args, *studentPersonaID)
		listArgs = append(listArgs, *studentPersonaID)
	}
	if status != "" {
		countQuery += ` AND status = ?`
		listQuery += ` AND a.status = ?`
		args = append(args, status)
		listArgs = append(listArgs, status)
	}

	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询作业总数失败: %w", err)
	}

	listQuery += ` ORDER BY a.created_at DESC LIMIT ? OFFSET ?`
	listArgs = append(listArgs, limit, offset)
	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询作业列表失败: %w", err)
	}
	defer rows.Close()

	var items []AssignmentListItem
	for rows.Next() {
		var item AssignmentListItem
		var hasFileInt int
		if err := rows.Scan(&item.ID, &item.StudentID, &item.StudentNickname,
			&item.TeacherID, &item.TeacherNickname,
			&item.TeacherPersonaID, &item.StudentPersonaID,
			&item.Title, &item.Status, &hasFileInt, &item.ReviewCount,
			&item.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描作业记录失败: %w", err)
		}
		item.HasFile = hasFileInt == 1
		items = append(items, item)
	}

	return items, total, nil
}

// ListByStudentPersona 按学生分身查询作业列表
func (r *AssignmentRepository) ListByStudentPersona(studentPersonaID int64, teacherPersonaID *int64, status string, offset, limit int) ([]AssignmentListItem, int, error) {
	countQuery := `SELECT COUNT(*) FROM assignments WHERE student_persona_id = ?`
	listQuery := `SELECT a.id, a.student_id, COALESCE(s.nickname, '') AS student_nickname,
		        a.teacher_id, COALESCE(t.nickname, '') AS teacher_nickname,
		        COALESCE(a.teacher_persona_id, 0), COALESCE(a.student_persona_id, 0),
		        a.title, a.status,
		        CASE WHEN COALESCE(a.file_path, '') != '' THEN 1 ELSE 0 END AS has_file,
		        (SELECT COUNT(*) FROM assignment_reviews WHERE assignment_id = a.id) AS review_count,
		        a.created_at
		 FROM assignments a
		 LEFT JOIN users s ON a.student_id = s.id
		 LEFT JOIN users t ON a.teacher_id = t.id
		 WHERE a.student_persona_id = ?`
	args := []interface{}{studentPersonaID}
	listArgs := []interface{}{studentPersonaID}

	if teacherPersonaID != nil {
		countQuery += ` AND teacher_persona_id = ?`
		listQuery += ` AND a.teacher_persona_id = ?`
		args = append(args, *teacherPersonaID)
		listArgs = append(listArgs, *teacherPersonaID)
	}
	if status != "" {
		countQuery += ` AND status = ?`
		listQuery += ` AND a.status = ?`
		args = append(args, status)
		listArgs = append(listArgs, status)
	}

	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询作业总数失败: %w", err)
	}

	listQuery += ` ORDER BY a.created_at DESC LIMIT ? OFFSET ?`
	listArgs = append(listArgs, limit, offset)
	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询作业列表失败: %w", err)
	}
	defer rows.Close()

	var items []AssignmentListItem
	for rows.Next() {
		var item AssignmentListItem
		var hasFileInt int
		if err := rows.Scan(&item.ID, &item.StudentID, &item.StudentNickname,
			&item.TeacherID, &item.TeacherNickname,
			&item.TeacherPersonaID, &item.StudentPersonaID,
			&item.Title, &item.Status, &hasFileInt, &item.ReviewCount,
			&item.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描作业记录失败: %w", err)
		}
		item.HasFile = hasFileInt == 1
		items = append(items, item)
	}

	return items, total, nil
}

// ==================== ReviewRepository ====================

type ReviewRepository struct {
	db *sql.DB
}

func NewReviewRepository(db *sql.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) Create(review *AssignmentReview) (int64, error) {
	now := time.Now()
	result, err := r.db.Exec(
		`INSERT INTO assignment_reviews (assignment_id, reviewer_type, reviewer_id, content, score, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		review.AssignmentID, review.ReviewerType, review.ReviewerID, review.Content, review.Score,
		now,
	)
	if err != nil {
		return 0, fmt.Errorf("创建点评失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取点评ID失败: %w", err)
	}

	return id, nil
}

func (r *ReviewRepository) ListByAssignment(assignmentID int64) ([]*AssignmentReview, error) {
	rows, err := r.db.Query(
		`SELECT id, assignment_id, reviewer_type, reviewer_id, content, score, created_at
		 FROM assignment_reviews WHERE assignment_id = ?
		 ORDER BY created_at ASC`,
		assignmentID,
	)
	if err != nil {
		return nil, fmt.Errorf("查询点评列表失败: %w", err)
	}
	defer rows.Close()

	var reviews []*AssignmentReview
	for rows.Next() {
		review := &AssignmentReview{}
		if err := rows.Scan(&review.ID, &review.AssignmentID, &review.ReviewerType,
			&review.ReviewerID, &review.Content, &review.Score, &review.CreatedAt); err != nil {
			return nil, fmt.Errorf("扫描点评记录失败: %w", err)
		}
		reviews = append(reviews, review)
	}

	return reviews, nil
}
