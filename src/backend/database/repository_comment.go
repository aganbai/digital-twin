package database

import (
	"database/sql"
	"fmt"
	"time"
)

// ==================== CommentRepository ====================

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(comment *TeacherComment) (int64, error) {
	now := time.Now()
	result, err := r.db.Exec(
		`INSERT INTO teacher_comments (teacher_id, student_id, content, progress_summary, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		comment.TeacherID, comment.StudentID, comment.Content, comment.ProgressSummary,
		now, now,
	)
	if err != nil {
		return 0, fmt.Errorf("创建评语失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取评语ID失败: %w", err)
	}

	return id, nil
}

func (r *CommentRepository) ListByTeacher(teacherID int64, studentID *int64, offset, limit int) ([]*TeacherComment, int, error) {
	// 构建查询条件
	countQuery := `SELECT COUNT(*) FROM teacher_comments WHERE teacher_id = ?`
	listQuery := `SELECT id, teacher_id, student_id, content, progress_summary, created_at, updated_at
		 FROM teacher_comments WHERE teacher_id = ?`
	args := []interface{}{teacherID}

	if studentID != nil {
		countQuery += ` AND student_id = ?`
		listQuery += ` AND student_id = ?`
		args = append(args, *studentID)
	}

	// 查询总数
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询评语总数失败: %w", err)
	}

	// 查询列表
	listQuery += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	listArgs := append(args, limit, offset)
	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询评语列表失败: %w", err)
	}
	defer rows.Close()

	var comments []*TeacherComment
	for rows.Next() {
		c := &TeacherComment{}
		if err := rows.Scan(&c.ID, &c.TeacherID, &c.StudentID, &c.Content,
			&c.ProgressSummary, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描评语记录失败: %w", err)
		}
		comments = append(comments, c)
	}

	return comments, total, nil
}

func (r *CommentRepository) ListByStudent(studentID int64, teacherID *int64, offset, limit int) ([]*TeacherComment, int, error) {
	// 构建查询条件
	countQuery := `SELECT COUNT(*) FROM teacher_comments WHERE student_id = ?`
	listQuery := `SELECT id, teacher_id, student_id, content, progress_summary, created_at, updated_at
		 FROM teacher_comments WHERE student_id = ?`
	args := []interface{}{studentID}

	if teacherID != nil {
		countQuery += ` AND teacher_id = ?`
		listQuery += ` AND teacher_id = ?`
		args = append(args, *teacherID)
	}

	// 查询总数
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询评语总数失败: %w", err)
	}

	// 查询列表
	listQuery += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	listArgs := append(args, limit, offset)
	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询评语列表失败: %w", err)
	}
	defer rows.Close()

	var comments []*TeacherComment
	for rows.Next() {
		c := &TeacherComment{}
		if err := rows.Scan(&c.ID, &c.TeacherID, &c.StudentID, &c.Content,
			&c.ProgressSummary, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描评语记录失败: %w", err)
		}
		comments = append(comments, c)
	}

	return comments, total, nil
}

// ListByTeacherWithNames 教师视角：JOIN users 获取 student_nickname 和 teacher_nickname
func (r *CommentRepository) ListByTeacherWithNames(teacherID int64, studentID *int64, offset, limit int) ([]CommentWithNames, int, error) {
	// 构建查询条件
	countQuery := `SELECT COUNT(*) FROM teacher_comments WHERE teacher_id = ?`
	listQuery := `SELECT c.id, c.teacher_id, COALESCE(t.nickname, '') AS teacher_nickname,
			c.student_id, COALESCE(s.nickname, '') AS student_nickname,
			c.content, c.progress_summary, c.created_at
		 FROM teacher_comments c
		 LEFT JOIN users t ON c.teacher_id = t.id
		 LEFT JOIN users s ON c.student_id = s.id
		 WHERE c.teacher_id = ?`
	args := []interface{}{teacherID}

	if studentID != nil {
		countQuery += ` AND student_id = ?`
		listQuery += ` AND c.student_id = ?`
		args = append(args, *studentID)
	}

	// 查询总数
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询评语总数失败: %w", err)
	}

	// 查询列表
	listQuery += ` ORDER BY c.created_at DESC LIMIT ? OFFSET ?`
	listArgs := append(args, limit, offset)
	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询评语列表失败: %w", err)
	}
	defer rows.Close()

	var comments []CommentWithNames
	for rows.Next() {
		var cmt CommentWithNames
		if err := rows.Scan(&cmt.ID, &cmt.TeacherID, &cmt.TeacherNickname,
			&cmt.StudentID, &cmt.StudentNickname,
			&cmt.Content, &cmt.ProgressSummary, &cmt.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描评语记录失败: %w", err)
		}
		comments = append(comments, cmt)
	}

	return comments, total, nil
}

// ListByStudentWithNames 学生视角：JOIN users 获取 teacher_nickname 和 student_nickname
func (r *CommentRepository) ListByStudentWithNames(studentID int64, teacherID *int64, offset, limit int) ([]CommentWithNames, int, error) {
	// 构建查询条件
	countQuery := `SELECT COUNT(*) FROM teacher_comments WHERE student_id = ?`
	listQuery := `SELECT c.id, c.teacher_id, COALESCE(t.nickname, '') AS teacher_nickname,
			c.student_id, COALESCE(s.nickname, '') AS student_nickname,
			c.content, c.progress_summary, c.created_at
		 FROM teacher_comments c
		 LEFT JOIN users t ON c.teacher_id = t.id
		 LEFT JOIN users s ON c.student_id = s.id
		 WHERE c.student_id = ?`
	args := []interface{}{studentID}

	if teacherID != nil {
		countQuery += ` AND teacher_id = ?`
		listQuery += ` AND c.teacher_id = ?`
		args = append(args, *teacherID)
	}

	// 查询总数
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询评语总数失败: %w", err)
	}

	// 查询列表
	listQuery += ` ORDER BY c.created_at DESC LIMIT ? OFFSET ?`
	listArgs := append(args, limit, offset)
	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询评语列表失败: %w", err)
	}
	defer rows.Close()

	var comments []CommentWithNames
	for rows.Next() {
		var cmt CommentWithNames
		if err := rows.Scan(&cmt.ID, &cmt.TeacherID, &cmt.TeacherNickname,
			&cmt.StudentID, &cmt.StudentNickname,
			&cmt.Content, &cmt.ProgressSummary, &cmt.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描评语记录失败: %w", err)
		}
		comments = append(comments, cmt)
	}

	return comments, total, nil
}

// ==================== 分身维度方法 ====================

// CreateWithPersonas 创建带分身维度的评语
func (r *CommentRepository) CreateWithPersonas(comment *TeacherComment) (int64, error) {
	now := time.Now()
	result, err := r.db.Exec(
		`INSERT INTO teacher_comments (teacher_id, student_id, teacher_persona_id, student_persona_id, content, progress_summary, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		comment.TeacherID, comment.StudentID, comment.TeacherPersonaID, comment.StudentPersonaID,
		comment.Content, comment.ProgressSummary, now, now,
	)
	if err != nil {
		return 0, fmt.Errorf("创建评语失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取评语ID失败: %w", err)
	}

	return id, nil
}

// ListByTeacherPersona 按教师分身查询评语列表
func (r *CommentRepository) ListByTeacherPersona(teacherPersonaID int64, studentPersonaID *int64, offset, limit int) ([]CommentWithNames, int, error) {
	countQuery := `SELECT COUNT(*) FROM teacher_comments WHERE teacher_persona_id = ?`
	listQuery := `SELECT c.id, c.teacher_id, COALESCE(t.nickname, '') AS teacher_nickname,
			c.student_id, COALESCE(s.nickname, '') AS student_nickname,
			COALESCE(c.teacher_persona_id, 0), COALESCE(c.student_persona_id, 0),
			c.content, c.progress_summary, c.created_at
		 FROM teacher_comments c
		 LEFT JOIN users t ON c.teacher_id = t.id
		 LEFT JOIN users s ON c.student_id = s.id
		 WHERE c.teacher_persona_id = ?`
	args := []interface{}{teacherPersonaID}

	if studentPersonaID != nil {
		countQuery += ` AND student_persona_id = ?`
		listQuery += ` AND c.student_persona_id = ?`
		args = append(args, *studentPersonaID)
	}

	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询评语总数失败: %w", err)
	}

	listQuery += ` ORDER BY c.created_at DESC LIMIT ? OFFSET ?`
	listArgs := append(args, limit, offset)
	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询评语列表失败: %w", err)
	}
	defer rows.Close()

	var comments []CommentWithNames
	for rows.Next() {
		var cmt CommentWithNames
		if err := rows.Scan(&cmt.ID, &cmt.TeacherID, &cmt.TeacherNickname,
			&cmt.StudentID, &cmt.StudentNickname,
			&cmt.TeacherPersonaID, &cmt.StudentPersonaID,
			&cmt.Content, &cmt.ProgressSummary, &cmt.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描评语记录失败: %w", err)
		}
		comments = append(comments, cmt)
	}

	return comments, total, nil
}

// ListByStudentPersona 按学生分身查询评语列表
func (r *CommentRepository) ListByStudentPersona(studentPersonaID int64, teacherPersonaID *int64, offset, limit int) ([]CommentWithNames, int, error) {
	countQuery := `SELECT COUNT(*) FROM teacher_comments WHERE student_persona_id = ?`
	listQuery := `SELECT c.id, c.teacher_id, COALESCE(t.nickname, '') AS teacher_nickname,
			c.student_id, COALESCE(s.nickname, '') AS student_nickname,
			COALESCE(c.teacher_persona_id, 0), COALESCE(c.student_persona_id, 0),
			c.content, c.progress_summary, c.created_at
		 FROM teacher_comments c
		 LEFT JOIN users t ON c.teacher_id = t.id
		 LEFT JOIN users s ON c.student_id = s.id
		 WHERE c.student_persona_id = ?`
	args := []interface{}{studentPersonaID}

	if teacherPersonaID != nil {
		countQuery += ` AND teacher_persona_id = ?`
		listQuery += ` AND c.teacher_persona_id = ?`
		args = append(args, *teacherPersonaID)
	}

	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询评语总数失败: %w", err)
	}

	listQuery += ` ORDER BY c.created_at DESC LIMIT ? OFFSET ?`
	listArgs := append(args, limit, offset)
	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询评语列表失败: %w", err)
	}
	defer rows.Close()

	var comments []CommentWithNames
	for rows.Next() {
		var cmt CommentWithNames
		if err := rows.Scan(&cmt.ID, &cmt.TeacherID, &cmt.TeacherNickname,
			&cmt.StudentID, &cmt.StudentNickname,
			&cmt.TeacherPersonaID, &cmt.StudentPersonaID,
			&cmt.Content, &cmt.ProgressSummary, &cmt.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描评语记录失败: %w", err)
		}
		comments = append(comments, cmt)
	}

	return comments, total, nil
}
