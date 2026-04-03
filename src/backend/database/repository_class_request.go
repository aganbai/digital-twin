package database

import (
	"database/sql"
	"fmt"
	"time"
)

// ClassJoinRequestRepository 班级加入申请数据访问
type ClassJoinRequestRepository struct {
	db *Database
}

// NewClassJoinRequestRepository 创建申请仓库
func NewClassJoinRequestRepository(db *Database) *ClassJoinRequestRepository {
	return &ClassJoinRequestRepository{db: db}
}

// CreateJoinRequest 创建加入申请
func (r *ClassJoinRequestRepository) CreateJoinRequest(req *ClassJoinRequest) (int64, error) {
	result, err := r.db.DB.Exec(`
		INSERT INTO class_join_requests (
			class_id, student_persona_id, student_id, status,
			request_message, student_age, student_gender, student_family_info,
			request_time, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(class_id, student_persona_id) DO UPDATE SET
			status = 'pending',
			request_message = excluded.request_message,
			request_time = excluded.request_time,
			updated_at = excluded.updated_at`,
		req.ClassID, req.StudentPersonaID, req.StudentID, req.Status,
		req.RequestMessage, req.StudentAge, req.StudentGender, req.StudentFamilyInfo,
		time.Now(), time.Now(), time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("创建加入申请失败: %w", err)
	}
	return result.LastInsertId()
}

// GetJoinRequestByID 根据ID获取申请
func (r *ClassJoinRequestRepository) GetJoinRequestByID(id int64) (*ClassJoinRequest, error) {
	row := r.db.DB.QueryRow(`
		SELECT id, class_id, student_persona_id, student_id, status,
			request_message, teacher_evaluation, student_age, student_gender,
			student_family_info, request_time, approval_time, created_at, updated_at
		FROM class_join_requests WHERE id = ?`, id)

	req := &ClassJoinRequest{}
	err := row.Scan(
		&req.ID, &req.ClassID, &req.StudentPersonaID, &req.StudentID, &req.Status,
		&req.RequestMessage, &req.TeacherEvaluation, &req.StudentAge, &req.StudentGender,
		&req.StudentFamilyInfo, &req.RequestTime, &req.ApprovalTime, &req.CreatedAt, &req.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询加入申请失败: %w", err)
	}
	return req, nil
}

// GetJoinRequestByClassAndStudent 根据班级和学生获取申请
func (r *ClassJoinRequestRepository) GetJoinRequestByClassAndStudent(classID, studentPersonaID int64) (*ClassJoinRequest, error) {
	row := r.db.DB.QueryRow(`
		SELECT id, class_id, student_persona_id, student_id, status,
			request_message, teacher_evaluation, student_age, student_gender,
			student_family_info, request_time, approval_time, created_at, updated_at
		FROM class_join_requests
		WHERE class_id = ? AND student_persona_id = ?`, classID, studentPersonaID)

	req := &ClassJoinRequest{}
	err := row.Scan(
		&req.ID, &req.ClassID, &req.StudentPersonaID, &req.StudentID, &req.Status,
		&req.RequestMessage, &req.TeacherEvaluation, &req.StudentAge, &req.StudentGender,
		&req.StudentFamilyInfo, &req.RequestTime, &req.ApprovalTime, &req.CreatedAt, &req.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询加入申请失败: %w", err)
	}
	return req, nil
}

// GetPendingRequestsByTeacher 获取教师的待审批列表
func (r *ClassJoinRequestRepository) GetPendingRequestsByTeacher(teacherID int64) ([]ClassJoinRequestItem, error) {
	rows, err := r.db.DB.Query(`
		SELECT 
			r.id, r.class_id, c.name as class_name,
			r.student_persona_id, p.nickname as student_nickname, p.avatar as student_avatar,
			r.status, r.request_message, r.request_time, r.approval_time
		FROM class_join_requests r
		JOIN classes c ON r.class_id = c.id
		JOIN personas p ON r.student_persona_id = p.id
		WHERE c.persona_id IN (SELECT id FROM personas WHERE user_id = ? AND role = 'teacher')
		AND r.status = 'pending'
		ORDER BY r.request_time DESC`, teacherID)
	if err != nil {
		return nil, fmt.Errorf("查询待审批列表失败: %w", err)
	}
	defer rows.Close()

	var items []ClassJoinRequestItem
	for rows.Next() {
		var item ClassJoinRequestItem
		err := rows.Scan(
			&item.ID, &item.ClassID, &item.ClassName,
			&item.StudentPersonaID, &item.StudentNickname, &item.StudentAvatar,
			&item.Status, &item.RequestMessage, &item.RequestTime, &item.ApprovalTime,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描待审批记录失败: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

// ApproveJoinRequest 审批通过
func (r *ClassJoinRequestRepository) ApproveJoinRequest(id int64, teacherEvaluation string) error {
	now := time.Now()
	_, err := r.db.DB.Exec(`
		UPDATE class_join_requests SET
			status = 'approved',
			teacher_evaluation = ?,
			approval_time = ?,
			updated_at = ?
		WHERE id = ?`,
		teacherEvaluation, now, now, id,
	)
	if err != nil {
		return fmt.Errorf("审批通过失败: %w", err)
	}
	return nil
}

// RejectJoinRequest 审批拒绝
func (r *ClassJoinRequestRepository) RejectJoinRequest(id int64, teacherEvaluation string) error {
	now := time.Now()
	_, err := r.db.DB.Exec(`
		UPDATE class_join_requests SET
			status = 'rejected',
			teacher_evaluation = ?,
			approval_time = ?,
			updated_at = ?
		WHERE id = ?`,
		teacherEvaluation, now, now, id,
	)
	if err != nil {
		return fmt.Errorf("审批拒绝失败: %w", err)
	}
	return nil
}

// GetStudentRequests 获取学生的申请记录
func (r *ClassJoinRequestRepository) GetStudentRequests(studentID int64) ([]ClassJoinRequestItem, error) {
	rows, err := r.db.DB.Query(`
		SELECT 
			r.id, r.class_id, c.name as class_name,
			r.student_persona_id, p.nickname as student_nickname, p.avatar as student_avatar,
			r.status, r.request_message, r.request_time, r.approval_time
		FROM class_join_requests r
		JOIN classes c ON r.class_id = c.id
		JOIN personas p ON r.student_persona_id = p.id
		WHERE r.student_id = ?
		ORDER BY r.created_at DESC`, studentID)
	if err != nil {
		return nil, fmt.Errorf("查询学生申请记录失败: %w", err)
	}
	defer rows.Close()

	var items []ClassJoinRequestItem
	for rows.Next() {
		var item ClassJoinRequestItem
		err := rows.Scan(
			&item.ID, &item.ClassID, &item.ClassName,
			&item.StudentPersonaID, &item.StudentNickname, &item.StudentAvatar,
			&item.Status, &item.RequestMessage, &item.RequestTime, &item.ApprovalTime,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描申请记录失败: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

// UpdateRequestStudentInfo 更新申请的学生信息
func (r *ClassJoinRequestRepository) UpdateRequestStudentInfo(id int64, age int, gender, familyInfo string) error {
	_, err := r.db.DB.Exec(`
		UPDATE class_join_requests SET
			student_age = ?,
			student_gender = ?,
			student_family_info = ?,
			updated_at = ?
		WHERE id = ?`,
		age, gender, familyInfo, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("更新学生信息失败: %w", err)
	}
	return nil
}
