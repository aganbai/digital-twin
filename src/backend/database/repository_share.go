package database

import (
	"database/sql"
	"fmt"
	"time"
)

// ShareRepository 分享码数据访问层
type ShareRepository struct {
	db *sql.DB
}

// NewShareRepository 创建分享码仓库
func NewShareRepository(db *sql.DB) *ShareRepository {
	return &ShareRepository{db: db}
}

// Create 创建分享码
func (r *ShareRepository) Create(share *PersonaShare) (int64, error) {
	now := time.Now()
	result, err := r.db.Exec(
		`INSERT INTO persona_shares (teacher_persona_id, share_code, class_id, target_student_persona_id, expires_at, max_uses, used_count, is_active, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		share.TeacherPersonaID, share.ShareCode, share.ClassID, share.TargetStudentPersonaID, share.ExpiresAt,
		share.MaxUses, 0, 1, now,
	)
	if err != nil {
		return 0, fmt.Errorf("创建分享码失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取分享码ID失败: %w", err)
	}

	return id, nil
}

// GetByCode 根据分享码查询
func (r *ShareRepository) GetByCode(code string) (*PersonaShare, error) {
	share := &PersonaShare{}
	err := r.db.QueryRow(
		`SELECT id, teacher_persona_id, share_code, class_id, COALESCE(target_student_persona_id, 0), expires_at, max_uses, used_count, is_active, created_at
		 FROM persona_shares WHERE share_code = ?`,
		code,
	).Scan(&share.ID, &share.TeacherPersonaID, &share.ShareCode, &share.ClassID,
		&share.TargetStudentPersonaID, &share.ExpiresAt, &share.MaxUses, &share.UsedCount, &share.IsActive, &share.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询分享码失败: %w", err)
	}

	return share, nil
}

// GetByID 根据ID查询分享码
func (r *ShareRepository) GetByID(id int64) (*PersonaShare, error) {
	share := &PersonaShare{}
	err := r.db.QueryRow(
		`SELECT id, teacher_persona_id, share_code, class_id, COALESCE(target_student_persona_id, 0), expires_at, max_uses, used_count, is_active, created_at
		 FROM persona_shares WHERE id = ?`,
		id,
	).Scan(&share.ID, &share.TeacherPersonaID, &share.ShareCode, &share.ClassID,
		&share.TargetStudentPersonaID, &share.ExpiresAt, &share.MaxUses, &share.UsedCount, &share.IsActive, &share.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询分享码失败: %w", err)
	}

	return share, nil
}

// IncrementUsedCount 增加使用次数
func (r *ShareRepository) IncrementUsedCount(id int64) error {
	_, err := r.db.Exec(
		`UPDATE persona_shares SET used_count = used_count + 1 WHERE id = ?`,
		id,
	)
	if err != nil {
		return fmt.Errorf("更新使用次数失败: %w", err)
	}
	return nil
}

// Deactivate 停用分享码
func (r *ShareRepository) Deactivate(id int64) error {
	result, err := r.db.Exec(
		`UPDATE persona_shares SET is_active = 0 WHERE id = ?`,
		id,
	)
	if err != nil {
		return fmt.Errorf("停用分享码失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("分享码不存在: id=%d", id)
	}

	return nil
}

// ListByPersonaID 获取教师分身的分享码列表
func (r *ShareRepository) ListByPersonaID(personaID int64) ([]ShareListItem, error) {
	rows, err := r.db.Query(
		`SELECT ps.id, ps.share_code, ps.class_id, COALESCE(c.name, '') AS class_name,
		        COALESCE(ps.target_student_persona_id, 0), COALESCE(tp.nickname, '') AS target_student_nickname,
		        ps.expires_at, ps.max_uses, ps.used_count, ps.is_active, ps.created_at
		 FROM persona_shares ps
		 LEFT JOIN classes c ON ps.class_id = c.id
		 LEFT JOIN personas tp ON ps.target_student_persona_id = tp.id
		 WHERE ps.teacher_persona_id = ?
		 ORDER BY ps.created_at DESC`,
		personaID,
	)
	if err != nil {
		return nil, fmt.Errorf("查询分享码列表失败: %w", err)
	}
	defer rows.Close()

	var items []ShareListItem
	for rows.Next() {
		var item ShareListItem
		var isActiveInt int
		var className string
		var targetStudentNickname string
		if err := rows.Scan(&item.ID, &item.ShareCode, &item.ClassID, &className,
			&item.TargetStudentPersonaID, &targetStudentNickname,
			&item.ExpiresAt, &item.MaxUses, &item.UsedCount, &isActiveInt, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("扫描分享码记录失败: %w", err)
		}
		item.IsActive = isActiveInt == 1
		item.ClassName = className
		item.TargetStudentNickname = targetStudentNickname
		items = append(items, item)
	}

	return items, nil
}

// GetShareInfo 获取分享码信息（预览用）
func (r *ShareRepository) GetShareInfo(code string) (*ShareInfo, error) {
	share, err := r.GetByCode(code)
	if err != nil {
		return nil, err
	}
	if share == nil {
		return &ShareInfo{IsValid: false, Reason: "分享码不存在"}, nil
	}

	// 检查是否有效
	if share.IsActive != 1 {
		return &ShareInfo{IsValid: false, Reason: "分享码已停用"}, nil
	}
	if share.ExpiresAt != nil && share.ExpiresAt.Before(time.Now()) {
		return &ShareInfo{IsValid: false, Reason: "分享码已过期"}, nil
	}
	if share.MaxUses > 0 && share.UsedCount >= share.MaxUses {
		return &ShareInfo{IsValid: false, Reason: "分享码使用次数已达上限"}, nil
	}

	// 查询教师分身信息
	info := &ShareInfo{
		TeacherPersonaID: share.TeacherPersonaID,
		IsValid:          true,
	}

	r.db.QueryRow(
		`SELECT nickname, school, description FROM personas WHERE id = ?`,
		share.TeacherPersonaID,
	).Scan(&info.TeacherNickname, &info.TeacherSchool, &info.TeacherDescription)

	// 查询班级名称
	if share.ClassID != nil {
		r.db.QueryRow(`SELECT name FROM classes WHERE id = ?`, *share.ClassID).Scan(&info.ClassName)
	}

	// 查询目标学生信息
	if share.TargetStudentPersonaID > 0 {
		info.TargetStudentPersonaID = share.TargetStudentPersonaID
		r.db.QueryRow(`SELECT nickname FROM personas WHERE id = ?`, share.TargetStudentPersonaID).Scan(&info.TargetStudentNickname)
	}

	return info, nil
}
