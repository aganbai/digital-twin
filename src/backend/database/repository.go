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
		`SELECT id, username, password, role, nickname, email, openid, COALESCE(status, 'active'), COALESCE(wx_unionid, ''), school, description, default_persona_id, COALESCE(profile_snapshot, '{}'), created_at, updated_at 
		 FROM users WHERE openid = ?`,
		openid,
	).Scan(&user.ID, &user.Username, &user.Password, &user.Role,
		&user.Nickname, &user.Email, &user.OpenID, &user.Status, &user.WxUnionID, &user.School, &user.Description, &user.DefaultPersonaID, &user.ProfileSnapshot, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("根据 openid 查询用户失败: %w", err)
	}

	// V2.0 迭代9: 画像隐私保护，清空用户画像
	user.ProfileSnapshot = ""

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
		`SELECT id, username, password, role, nickname, email, openid, COALESCE(status, 'active'), COALESCE(wx_unionid, ''), school, description, default_persona_id, COALESCE(profile_snapshot, '{}'), created_at, updated_at 
		 FROM users WHERE username = ?`,
		username,
	).Scan(&user.ID, &user.Username, &user.Password, &user.Role,
		&user.Nickname, &user.Email, &user.OpenID, &user.Status, &user.WxUnionID, &user.School, &user.Description, &user.DefaultPersonaID, &user.ProfileSnapshot, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// V2.0 迭代9: 画像隐私保护，清空用户画像
	user.ProfileSnapshot = ""

	return user, nil
}

// GetByID 根据ID查询用户
func (r *UserRepository) GetByID(id int64) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(
		`SELECT id, username, password, role, nickname, email, openid, COALESCE(status, 'active'), COALESCE(wx_unionid, ''), school, description, default_persona_id, COALESCE(profile_snapshot, '{}'), created_at, updated_at 
		 FROM users WHERE id = ?`,
		id,
	).Scan(&user.ID, &user.Username, &user.Password, &user.Role,
		&user.Nickname, &user.Email, &user.OpenID, &user.Status, &user.WxUnionID, &user.School, &user.Description, &user.DefaultPersonaID, &user.ProfileSnapshot, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// V2.0 迭代9: 画像隐私保护，清空用户画像
	user.ProfileSnapshot = ""

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
		 ORDER BY u.created_at DESC
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

// ======================== V2.0 迭代10 新增方法 ========================

// UpdateWxUnionID 更新用户的微信 UnionID
func (r *UserRepository) UpdateWxUnionID(userID int64, unionID string) error {
	_, err := r.db.Exec(
		`UPDATE users SET wx_unionid = ?, updated_at = ? WHERE id = ?`,
		unionID, time.Now(), userID,
	)
	if err != nil {
		return fmt.Errorf("更新用户 UnionID 失败: %w", err)
	}
	return nil
}

// UpdateUserRole 更新用户角色
func (r *UserRepository) UpdateUserRole(userID int64, role string) error {
	result, err := r.db.Exec(
		`UPDATE users SET role = ?, updated_at = ? WHERE id = ?`,
		role, time.Now(), userID,
	)
	if err != nil {
		return fmt.Errorf("更新用户角色失败: %w", err)
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

// UpdateUserStatus 更新用户状态（启用/禁用）
func (r *UserRepository) UpdateUserStatus(userID int64, status string) error {
	result, err := r.db.Exec(
		`UPDATE users SET status = ?, updated_at = ? WHERE id = ?`,
		status, time.Now(), userID,
	)
	if err != nil {
		return fmt.Errorf("更新用户状态失败: %w", err)
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

// ListUsers 分页查询用户列表（管理员用）
func (r *UserRepository) ListUsers(role, status string, offset, limit int) ([]User, int, error) {
	// 构建查询条件
	whereClauses := []string{"1=1"}
	args := []interface{}{}

	if role != "" {
		whereClauses = append(whereClauses, "role = ?")
		args = append(args, role)
	}
	if status != "" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, status)
	}

	whereClause := ""
	for i, clause := range whereClauses {
		if i == 0 {
			whereClause = clause
		} else {
			whereClause += " AND " + clause
		}
	}

	// 查询总数
	var total int
	err := r.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM users WHERE %s", whereClause), args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询用户总数失败: %w", err)
	}

	// 查询列表
	query := fmt.Sprintf(`
		SELECT id, username, password, role, nickname, email, openid, COALESCE(status, 'active'), COALESCE(wx_unionid, ''),
		       school, description, default_persona_id, created_at, updated_at
		FROM users WHERE %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, limit, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询用户列表失败: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(
			&u.ID, &u.Username, &u.Password, &u.Role, &u.Nickname, &u.Email, &u.OpenID,
			&u.Status, &u.WxUnionID, &u.School, &u.Description, &u.DefaultPersonaID,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("扫描用户记录失败: %w", err)
		}
		users = append(users, u)
	}

	return users, total, nil
}

// CountUsers 统计用户总数
func (r *UserRepository) CountUsers(role string) (int, error) {
	var count int
	var err error
	if role != "" {
		err = r.db.QueryRow(`SELECT COUNT(*) FROM users WHERE role = ?`, role).Scan(&count)
	} else {
		err = r.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	}
	if err != nil {
		return 0, fmt.Errorf("统计用户数失败: %w", err)
	}
	return count, nil
}

// CountUsersByStatus 按状态统计用户数
func (r *UserRepository) CountUsersByStatus(status string) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM users WHERE COALESCE(status, 'active') = ?`, status).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("统计用户数失败: %w", err)
	}
	return count, nil
}

// GetActiveUsers 获取活跃用户排行（按最近登录/活动时间）
func (r *UserRepository) GetActiveUsers(limit int) ([]User, error) {
	rows, err := r.db.Query(`
		SELECT u.id, u.username, u.role, u.nickname, u.school, u.created_at
		FROM users u
		WHERE u.status = 'active' OR u.status IS NULL
		ORDER BY u.updated_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("查询活跃用户失败: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.Nickname, &u.School, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("扫描用户记录失败: %w", err)
		}
		users = append(users, u)
	}

	return users, nil
}

// ======================== V2.0 迭代11 M4 自测学生方法 ========================

// FindByTestTeacherID 根据教师ID查询自测学生
func (r *UserRepository) FindByTestTeacherID(teacherID int64) (*User, error) {
	user := &User{}
	var email, openid, status, wxUnionID, school, description sql.NullString
	var defaultPersonaID sql.NullInt64
	err := r.db.QueryRow(
		`SELECT id, username, password, role, nickname, email, openid, COALESCE(status, 'active'), wx_unionid, school, description, default_persona_id, COALESCE(is_test_student, 0), COALESCE(test_teacher_id, 0), created_at, updated_at 
		 FROM users WHERE test_teacher_id = ? AND is_test_student = 1`,
		teacherID,
	).Scan(&user.ID, &user.Username, &user.Password, &user.Role,
		&user.Nickname, &email, &openid, &status, &wxUnionID, &school, &description, &defaultPersonaID, &user.IsTestStudent, &user.TestTeacherID, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询自测学生失败: %w", err)
	}

	// 处理可能为 NULL 的字段
	user.Email = email.String
	user.OpenID = openid.String
	user.Status = status.String
	user.WxUnionID = wxUnionID.String
	user.School = school.String
	user.Description = description.String
	if defaultPersonaID.Valid {
		user.DefaultPersonaID = defaultPersonaID.Int64
	}

	return user, nil
}

// CreateTestStudent 创建自测学生用户
func (r *UserRepository) CreateTestStudent(user *User) (int64, error) {
	now := time.Now()
	result, err := r.db.Exec(
		`INSERT INTO users (username, password, role, nickname, is_test_student, test_teacher_id, created_at, updated_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		user.Username, user.Password, user.Role, user.Nickname, 1, user.TestTeacherID, now, now,
	)
	if err != nil {
		return 0, fmt.Errorf("创建自测学生失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取用户ID失败: %w", err)
	}

	return id, nil
}
