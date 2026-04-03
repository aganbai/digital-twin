package database

import (
	"database/sql"
	"fmt"
	"time"
)

// ==================== UserRepository ====================

// UserRepository 用户数据访问层
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(user *User) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO users (username, password, role, nickname, email, openid, created_at, updated_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		user.Username, user.Password, user.Role, user.Nickname, user.Email, user.OpenID,
		time.Now(), time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("创建用户失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取用户ID失败: %w", err)
	}

	return id, nil
}

// GetByOpenID 根据微信 openid 查询用户
func (r *UserRepository) GetByOpenID(openid string) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(
		`SELECT id, username, password, role, nickname, email, openid, school, description, default_persona_id, COALESCE(profile_snapshot, '{}'), created_at, updated_at 
		 FROM users WHERE openid = ?`,
		openid,
	).Scan(&user.ID, &user.Username, &user.Password, &user.Role,
		&user.Nickname, &user.Email, &user.OpenID, &user.School, &user.Description, &user.DefaultPersonaID, &user.ProfileSnapshot, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("根据 openid 查询用户失败: %w", err)
	}

	return user, nil
}

// CreateWithOpenID 创建带 openid 的用户（微信登录用）
func (r *UserRepository) CreateWithOpenID(user *User) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO users (username, password, role, nickname, email, openid, created_at, updated_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		user.Username, user.Password, user.Role, user.Nickname, user.Email, user.OpenID,
		time.Now(), time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("创建微信用户失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取用户ID失败: %w", err)
	}

	return id, nil
}

// UpdateRoleAndNickname 更新用户角色和昵称（补全信息用）
func (r *UserRepository) UpdateRoleAndNickname(userID int64, role, nickname string) error {
	result, err := r.db.Exec(
		`UPDATE users SET role = ?, nickname = ?, updated_at = ? WHERE id = ?`,
		role, nickname, time.Now(), userID,
	)
	if err != nil {
		return fmt.Errorf("更新用户角色和昵称失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("用户不存在: id=%d", userID)
	}

	return nil
}

// GetByUsername 根据用户名查询用户
func (r *UserRepository) GetByUsername(username string) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(
		`SELECT id, username, password, role, nickname, email, openid, school, description, default_persona_id, COALESCE(profile_snapshot, '{}'), created_at, updated_at 
		 FROM users WHERE username = ?`,
		username,
	).Scan(&user.ID, &user.Username, &user.Password, &user.Role,
		&user.Nickname, &user.Email, &user.OpenID, &user.School, &user.Description, &user.DefaultPersonaID, &user.ProfileSnapshot, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return user, nil
}

// GetByID 根据ID查询用户
func (r *UserRepository) GetByID(id int64) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(
		`SELECT id, username, password, role, nickname, email, openid, school, description, default_persona_id, COALESCE(profile_snapshot, '{}'), created_at, updated_at 
		 FROM users WHERE id = ?`,
		id,
	).Scan(&user.ID, &user.Username, &user.Password, &user.Role,
		&user.Nickname, &user.Email, &user.OpenID, &user.School, &user.Description, &user.DefaultPersonaID, &user.ProfileSnapshot, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return user, nil
}

// UpdateProfileSnapshot 更新用户画像快照（V2.0 迭代7: R10）
func (r *UserRepository) UpdateProfileSnapshot(userID int64, snapshot string) error {
	_, err := r.db.Exec(`UPDATE users SET profile_snapshot = ? WHERE id = ?`, snapshot, userID)
	if err != nil {
		return fmt.Errorf("更新用户画像失败: %w", err)
	}
	return nil
}

// ==================== 教师列表查询 ====================

// GetTeachers 查询教师列表，LEFT JOIN knowledge_items 统计文档数
func (r *UserRepository) GetTeachers(offset, limit int) ([]TeacherWithDocCount, int, error) {
	// 查询总数
	var total int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM users WHERE role = 'teacher'`,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询教师总数失败: %w", err)
	}

	// 查询列表，LEFT JOIN knowledge_items 统计 active 文档数
	rows, err := r.db.Query(
		`SELECT u.id, u.username, u.nickname, u.role, u.school, u.description,
		        COALESCE(COUNT(d.id), 0) AS document_count, u.created_at
		 FROM users u
		 LEFT JOIN knowledge_items d ON u.id = d.teacher_id AND d.status = 'active'
		 WHERE u.role = 'teacher'
		 GROUP BY u.id
		 ORDER BY u.created_at DESC
		 LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("查询教师列表失败: %w", err)
	}
	defer rows.Close()

	var teachers []TeacherWithDocCount
	for rows.Next() {
		t := TeacherWithDocCount{}
		if err := rows.Scan(&t.ID, &t.Username, &t.Nickname, &t.Role, &t.School, &t.Description,
			&t.DocumentCount, &t.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描教师记录失败: %w", err)
		}
		teachers = append(teachers, t)
	}

	return teachers, total, nil
}

// ==================== 用户统计信息 ====================

// GetUserStats 获取用户统计信息
// 学生：conversation_count（对话数）+ memory_count（记忆数）
// 教师：document_count（文档数）+ conversation_count（被提问数）
func (r *UserRepository) GetUserStats(userID int64, role string) (map[string]int, error) {
	stats := make(map[string]int)

	if role == "student" {
		// 学生：对话数
		var convCount int
		err := r.db.QueryRow(
			`SELECT COUNT(*) FROM conversations WHERE student_id = ?`, userID,
		).Scan(&convCount)
		if err != nil {
			return nil, fmt.Errorf("查询学生对话数失败: %w", err)
		}
		stats["conversation_count"] = convCount

		// 学生：记忆数
		var memCount int
		err = r.db.QueryRow(
			`SELECT COUNT(*) FROM memories WHERE student_id = ?`, userID,
		).Scan(&memCount)
		if err != nil {
			return nil, fmt.Errorf("查询学生记忆数失败: %w", err)
		}
		stats["memory_count"] = memCount

	} else if role == "teacher" {
		// 教师：active 文档数
		var docCount int
		err := r.db.QueryRow(
			`SELECT COUNT(*) FROM knowledge_items WHERE teacher_id = ? AND status = 'active'`, userID,
		).Scan(&docCount)
		if err != nil {
			return nil, fmt.Errorf("查询教师文档数失败: %w", err)
		}
		stats["document_count"] = docCount

		// 教师：被提问数（作为 teacher_id 出现在 conversations 中的次数）
		var convCount int
		err = r.db.QueryRow(
			`SELECT COUNT(*) FROM conversations WHERE teacher_id = ?`, userID,
		).Scan(&convCount)
		if err != nil {
			return nil, fmt.Errorf("查询教师被提问数失败: %w", err)
		}
		stats["conversation_count"] = convCount
	}

	return stats, nil
}

// ==================== UserRepository 新增方法 ====================

// CheckTeacherExists 检查教师是否存在（按昵称+学校唯一性）
func (r *UserRepository) CheckTeacherExists(nickname, school string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM users WHERE nickname = ? AND school = ? AND role = 'teacher'`,
		nickname, school,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查教师是否存在失败: %w", err)
	}
	return count > 0, nil
}

// UpdateProfile 更新用户资料（角色、昵称、学校、描述）
func (r *UserRepository) UpdateProfile(userID int64, role, nickname, school, description string) error {
	result, err := r.db.Exec(
		`UPDATE users SET role = ?, nickname = ?, school = ?, description = ?, updated_at = ? WHERE id = ?`,
		role, nickname, school, description, time.Now(), userID,
	)
	if err != nil {
		return fmt.Errorf("更新用户信息失败: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("用户不存在: id=%d", userID)
	}
	return nil
}

// UpdateDefaultPersonaID 更新用户的默认分身ID
func (r *UserRepository) UpdateDefaultPersonaID(userID, personaID int64) error {
	result, err := r.db.Exec(
		`UPDATE users SET default_persona_id = ?, updated_at = ? WHERE id = ?`,
		personaID, time.Now(), userID,
	)
	if err != nil {
		return fmt.Errorf("更新默认分身ID失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("用户不存在: id=%d", userID)
	}

	return nil
}

// ======================== V2.0 迭代3 新增方法 ========================

// ListTeachersForStudent 查询与指定学生分身有 approved + is_active 关系的教师分身列表
func (r *UserRepository) ListTeachersForStudent(studentPersonaID int64, offset, limit int) ([]TeacherWithDocCount, int, error) {
	// 查询总数
	var total int
	err := r.db.QueryRow(
		`SELECT COUNT(DISTINCT u.id) FROM users u
		 JOIN teacher_student_relations r ON u.id = r.teacher_id
		 JOIN personas p ON r.teacher_persona_id = p.id
		 WHERE r.student_persona_id = ? AND r.status = 'approved' AND COALESCE(r.is_active, 1) = 1 AND COALESCE(p.is_active, 1) = 1`,
		studentPersonaID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询教师总数失败: %w", err)
	}

	// 查询列表
	rows, err := r.db.Query(
		`SELECT u.id, r.teacher_persona_id, u.username, u.nickname, u.role, u.school, u.description,
		        COALESCE(COUNT(d.id), 0) AS document_count, u.created_at
		 FROM users u
		 JOIN teacher_student_relations r ON u.id = r.teacher_id
		 JOIN personas p ON r.teacher_persona_id = p.id
		 LEFT JOIN knowledge_items d ON u.id = d.teacher_id AND d.status = 'active'
		 WHERE r.student_persona_id = ? AND r.status = 'approved' AND COALESCE(r.is_active, 1) = 1 AND COALESCE(p.is_active, 1) = 1
		 GROUP BY u.id, r.teacher_persona_id
		 ORDER BY r.updated_at DESC
		 LIMIT ? OFFSET ?`,
		studentPersonaID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("查询教师列表失败: %w", err)
	}
	defer rows.Close()

	var teachers []TeacherWithDocCount
	for rows.Next() {
		t := TeacherWithDocCount{}
		if err := rows.Scan(&t.ID, &t.PersonaID, &t.Username, &t.Nickname, &t.Role, &t.School, &t.Description,
			&t.DocumentCount, &t.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描教师记录失败: %w", err)
		}
		teachers = append(teachers, t)
	}

	return teachers, total, nil
}
