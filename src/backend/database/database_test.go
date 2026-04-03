package database

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) *Database {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// ==================== Database 基础测试 ====================

func TestNewDatabase_Success(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "data", "test.db")

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("创建数据库失败: %v", err)
	}
	defer db.Close()

	// 验证目录已自动创建
	dirPath := filepath.Join(tmpDir, "data")
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		t.Error("数据库目录未自动创建")
	}

	// 验证数据库文件存在
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("数据库文件未创建")
	}
}

func TestNewDatabase_AutoCreateDir(t *testing.T) {
	tmpDir := t.TempDir()
	// 多层嵌套目录
	dbPath := filepath.Join(tmpDir, "a", "b", "c", "test.db")

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("创建数据库失败: %v", err)
	}
	defer db.Close()

	dirPath := filepath.Join(tmpDir, "a", "b", "c")
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		t.Error("多层嵌套目录未自动创建")
	}
}

func TestNewDatabase_TablesCreated(t *testing.T) {
	db := setupTestDB(t)

	// 验证表已创建
	tables := []string{"users", "documents", "conversations", "memories"}
	for _, table := range tables {
		var count int
		err := db.DB.QueryRow(
			"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&count)
		if err != nil {
			t.Fatalf("查询表 %s 失败: %v", table, err)
		}
		if count == 0 {
			t.Errorf("表 %s 未创建", table)
		}
	}
}

func TestDatabase_Close(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("创建数据库失败: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("关闭数据库失败: %v", err)
	}

	// 关闭后操作应失败
	err = db.DB.Ping()
	if err == nil {
		t.Error("数据库关闭后 Ping 应失败")
	}
}

// ==================== UserRepository 测试 ====================

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db.DB)

	user := &User{
		Username: "testuser",
		Password: "hashed_password",
		Role:     "student",
		Nickname: "测试用户",
		Email:    "test@example.com",
	}

	id, err := repo.Create(user)
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	if id <= 0 {
		t.Errorf("用户ID应大于0: got %d", id)
	}
}

func TestUserRepository_CreateDuplicateUsername(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db.DB)

	user := &User{
		Username: "duplicate_user",
		Password: "password",
		Role:     "student",
	}

	_, err := repo.Create(user)
	if err != nil {
		t.Fatalf("第一次创建用户失败: %v", err)
	}

	// 重复用户名应失败
	_, err = repo.Create(user)
	if err == nil {
		t.Fatal("重复用户名应返回错误")
	}
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db.DB)

	// 先创建用户
	user := &User{
		Username: "findme",
		Password: "password123",
		Role:     "teacher",
		Nickname: "查找我",
	}
	_, err := repo.Create(user)
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	// 查询用户
	found, err := repo.GetByUsername("findme")
	if err != nil {
		t.Fatalf("查询用户失败: %v", err)
	}
	if found == nil {
		t.Fatal("应找到用户")
	}
	if found.Username != "findme" {
		t.Errorf("用户名不匹配: got %q", found.Username)
	}
	if found.Role != "teacher" {
		t.Errorf("角色不匹配: got %q", found.Role)
	}
}

func TestUserRepository_GetByUsername_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db.DB)

	found, err := repo.GetByUsername("nonexistent")
	if err != nil {
		t.Fatalf("查询不应返回错误: %v", err)
	}
	if found != nil {
		t.Fatal("不存在的用户应返回 nil")
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db.DB)

	user := &User{
		Username: "iduser",
		Password: "password",
		Role:     "admin",
	}
	id, err := repo.Create(user)
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	found, err := repo.GetByID(id)
	if err != nil {
		t.Fatalf("查询用户失败: %v", err)
	}
	if found == nil {
		t.Fatal("应找到用户")
	}
	if found.ID != id {
		t.Errorf("用户ID不匹配: got %d, want %d", found.ID, id)
	}
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db.DB)

	found, err := repo.GetByID(99999)
	if err != nil {
		t.Fatalf("查询不应返回错误: %v", err)
	}
	if found != nil {
		t.Fatal("不存在的用户应返回 nil")
	}
}

// ==================== DocumentRepository 测试 ====================

func createTestTeacher(t *testing.T, db *Database) int64 {
	t.Helper()
	repo := NewUserRepository(db.DB)
	id, err := repo.Create(&User{
		Username: "teacher_" + time.Now().Format("150405.000"),
		Password: "password",
		Role:     "teacher",
	})
	if err != nil {
		t.Fatalf("创建测试教师失败: %v", err)
	}
	return id
}

func TestDocumentRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	teacherID := createTestTeacher(t, db)
	repo := NewDocumentRepository(db.DB)

	doc := &Document{
		TeacherID: teacherID,
		Title:     "测试文档",
		Content:   "这是测试内容",
		DocType:   "text",
		Tags:      `["测试","文档"]`,
		Status:    "active",
	}

	id, err := repo.Create(doc)
	if err != nil {
		t.Fatalf("创建文档失败: %v", err)
	}
	if id <= 0 {
		t.Errorf("文档ID应大于0: got %d", id)
	}
}

func TestDocumentRepository_GetByTeacherID(t *testing.T) {
	db := setupTestDB(t)
	teacherID := createTestTeacher(t, db)
	repo := NewDocumentRepository(db.DB)

	// 创建多个文档
	for i := 0; i < 3; i++ {
		doc := &Document{
			TeacherID: teacherID,
			Title:     "文档" + string(rune('A'+i)),
			Content:   "内容",
			DocType:   "text",
			Status:    "active",
		}
		_, err := repo.Create(doc)
		if err != nil {
			t.Fatalf("创建文档失败: %v", err)
		}
	}

	docs, total, err := repo.GetByTeacherID(teacherID, 0, 10)
	if err != nil {
		t.Fatalf("查询文档列表失败: %v", err)
	}
	if total != 3 {
		t.Errorf("文档总数不匹配: got %d, want 3", total)
	}
	if len(docs) != 3 {
		t.Errorf("文档列表长度不匹配: got %d, want 3", len(docs))
	}
}

func TestDocumentRepository_GetByTeacherID_Pagination(t *testing.T) {
	db := setupTestDB(t)
	teacherID := createTestTeacher(t, db)
	repo := NewDocumentRepository(db.DB)

	for i := 0; i < 5; i++ {
		doc := &Document{
			TeacherID: teacherID,
			Title:     "文档" + string(rune('A'+i)),
			Content:   "内容",
			DocType:   "text",
			Status:    "active",
		}
		_, err := repo.Create(doc)
		if err != nil {
			t.Fatalf("创建文档失败: %v", err)
		}
	}

	// 第一页
	docs, total, err := repo.GetByTeacherID(teacherID, 0, 2)
	if err != nil {
		t.Fatalf("查询第一页失败: %v", err)
	}
	if total != 5 {
		t.Errorf("总数不匹配: got %d, want 5", total)
	}
	if len(docs) != 2 {
		t.Errorf("第一页数量不匹配: got %d, want 2", len(docs))
	}
}

func TestDocumentRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	teacherID := createTestTeacher(t, db)
	repo := NewDocumentRepository(db.DB)

	doc := &Document{
		TeacherID: teacherID,
		Title:     "查找文档",
		Content:   "内容",
		DocType:   "text",
		Status:    "active",
	}
	id, err := repo.Create(doc)
	if err != nil {
		t.Fatalf("创建文档失败: %v", err)
	}

	found, err := repo.GetByID(id)
	if err != nil {
		t.Fatalf("查询文档失败: %v", err)
	}
	if found == nil {
		t.Fatal("应找到文档")
	}
	if found.Title != "查找文档" {
		t.Errorf("文档标题不匹配: got %q", found.Title)
	}
}

func TestDocumentRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	teacherID := createTestTeacher(t, db)
	repo := NewDocumentRepository(db.DB)

	doc := &Document{
		TeacherID: teacherID,
		Title:     "待删除文档",
		Content:   "内容",
		DocType:   "text",
		Status:    "active",
	}
	id, err := repo.Create(doc)
	if err != nil {
		t.Fatalf("创建文档失败: %v", err)
	}

	// 删除文档（软删除）
	if err := repo.Delete(id); err != nil {
		t.Fatalf("删除文档失败: %v", err)
	}

	// 验证状态变为 archived
	found, err := repo.GetByID(id)
	if err != nil {
		t.Fatalf("查询文档失败: %v", err)
	}
	if found.Status != "archived" {
		t.Errorf("文档状态应为 archived: got %q", found.Status)
	}

	// 验证 GetByTeacherID 不返回已删除的文档
	docs, total, err := repo.GetByTeacherID(teacherID, 0, 10)
	if err != nil {
		t.Fatalf("查询文档列表失败: %v", err)
	}
	if total != 0 {
		t.Errorf("已删除文档不应出现在列表中: total=%d", total)
	}
	if len(docs) != 0 {
		t.Errorf("已删除文档不应出现在列表中: len=%d", len(docs))
	}
}

func TestDocumentRepository_Delete_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDocumentRepository(db.DB)

	err := repo.Delete(99999)
	if err == nil {
		t.Fatal("删除不存在的文档应返回错误")
	}
}

// ==================== ConversationRepository 测试 ====================

func createTestStudentAndTeacher(t *testing.T, db *Database) (int64, int64) {
	t.Helper()
	userRepo := NewUserRepository(db.DB)

	suffix := time.Now().Format("150405.000000")
	teacherID, err := userRepo.Create(&User{
		Username: "teacher_" + suffix,
		Password: "password",
		Role:     "teacher",
	})
	if err != nil {
		t.Fatalf("创建教师失败: %v", err)
	}

	studentID, err := userRepo.Create(&User{
		Username: "student_" + suffix,
		Password: "password",
		Role:     "student",
	})
	if err != nil {
		t.Fatalf("创建学生失败: %v", err)
	}

	return studentID, teacherID
}

func TestConversationRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	studentID, teacherID := createTestStudentAndTeacher(t, db)
	repo := NewConversationRepository(db.DB)

	conv := &Conversation{
		StudentID:  studentID,
		TeacherID:  teacherID,
		SessionID:  "test-session-001",
		Role:       "user",
		Content:    "你好，请问什么是牛顿第一定律？",
		TokenCount: 15,
	}

	id, err := repo.Create(conv)
	if err != nil {
		t.Fatalf("创建对话记录失败: %v", err)
	}
	if id <= 0 {
		t.Errorf("对话记录ID应大于0: got %d", id)
	}
}

func TestConversationRepository_GetByStudentAndTeacher(t *testing.T) {
	db := setupTestDB(t)
	studentID, teacherID := createTestStudentAndTeacher(t, db)
	repo := NewConversationRepository(db.DB)

	// 创建多条对话
	for i := 0; i < 4; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		conv := &Conversation{
			StudentID:  studentID,
			TeacherID:  teacherID,
			SessionID:  "session-001",
			Role:       role,
			Content:    "消息 " + string(rune('A'+i)),
			TokenCount: 10,
		}
		_, err := repo.Create(conv)
		if err != nil {
			t.Fatalf("创建对话记录失败: %v", err)
		}
	}

	convs, total, err := repo.GetByStudentAndTeacher(studentID, teacherID, 0, 10)
	if err != nil {
		t.Fatalf("查询对话历史失败: %v", err)
	}
	if total != 4 {
		t.Errorf("对话总数不匹配: got %d, want 4", total)
	}
	if len(convs) != 4 {
		t.Errorf("对话列表长度不匹配: got %d, want 4", len(convs))
	}
}

func TestConversationRepository_GetByStudentAndTeacher_Pagination(t *testing.T) {
	db := setupTestDB(t)
	studentID, teacherID := createTestStudentAndTeacher(t, db)
	repo := NewConversationRepository(db.DB)

	for i := 0; i < 5; i++ {
		conv := &Conversation{
			StudentID:  studentID,
			TeacherID:  teacherID,
			SessionID:  "session-001",
			Role:       "user",
			Content:    "消息",
			TokenCount: 5,
		}
		_, err := repo.Create(conv)
		if err != nil {
			t.Fatalf("创建对话记录失败: %v", err)
		}
	}

	// 分页查询
	convs, total, err := repo.GetByStudentAndTeacher(studentID, teacherID, 0, 2)
	if err != nil {
		t.Fatalf("分页查询失败: %v", err)
	}
	if total != 5 {
		t.Errorf("总数不匹配: got %d, want 5", total)
	}
	if len(convs) != 2 {
		t.Errorf("分页数量不匹配: got %d, want 2", len(convs))
	}
}

// ==================== MemoryRepository 测试 ====================

func TestMemoryRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	studentID, teacherID := createTestStudentAndTeacher(t, db)
	repo := NewMemoryRepository(db.DB)

	mem := &Memory{
		StudentID:  studentID,
		TeacherID:  teacherID,
		MemoryType: "conversation",
		Content:    "学生对牛顿定律有基本了解",
		Importance: 0.8,
	}

	id, err := repo.Create(mem)
	if err != nil {
		t.Fatalf("创建记忆失败: %v", err)
	}
	if id <= 0 {
		t.Errorf("记忆ID应大于0: got %d", id)
	}
}

func TestMemoryRepository_GetByStudentAndTeacher(t *testing.T) {
	db := setupTestDB(t)
	studentID, teacherID := createTestStudentAndTeacher(t, db)
	repo := NewMemoryRepository(db.DB)

	// 创建多条记忆
	types := []string{"conversation", "learning_progress", "personality_traits"}
	importances := []float64{0.5, 0.9, 0.3}
	for i, mt := range types {
		mem := &Memory{
			StudentID:  studentID,
			TeacherID:  teacherID,
			MemoryType: mt,
			Content:    "记忆内容 " + mt,
			Importance: importances[i],
		}
		_, err := repo.Create(mem)
		if err != nil {
			t.Fatalf("创建记忆失败: %v", err)
		}
	}

	memories, err := repo.GetByStudentAndTeacher(studentID, teacherID, 10)
	if err != nil {
		t.Fatalf("查询记忆失败: %v", err)
	}
	if len(memories) != 3 {
		t.Errorf("记忆数量不匹配: got %d, want 3", len(memories))
	}

	// 验证按重要性降序排列
	if memories[0].Importance < memories[1].Importance {
		t.Error("记忆应按重要性降序排列")
	}
}

func TestMemoryRepository_GetByStudentAndTeacher_Limit(t *testing.T) {
	db := setupTestDB(t)
	studentID, teacherID := createTestStudentAndTeacher(t, db)
	repo := NewMemoryRepository(db.DB)

	for i := 0; i < 5; i++ {
		mem := &Memory{
			StudentID:  studentID,
			TeacherID:  teacherID,
			MemoryType: "conversation",
			Content:    "记忆",
			Importance: 0.5,
		}
		_, err := repo.Create(mem)
		if err != nil {
			t.Fatalf("创建记忆失败: %v", err)
		}
	}

	memories, err := repo.GetByStudentAndTeacher(studentID, teacherID, 3)
	if err != nil {
		t.Fatalf("查询记忆失败: %v", err)
	}
	if len(memories) != 3 {
		t.Errorf("Limit 限制不生效: got %d, want 3", len(memories))
	}
}

func TestMemoryRepository_UpdateLastAccessed(t *testing.T) {
	db := setupTestDB(t)
	studentID, teacherID := createTestStudentAndTeacher(t, db)
	repo := NewMemoryRepository(db.DB)

	mem := &Memory{
		StudentID:  studentID,
		TeacherID:  teacherID,
		MemoryType: "conversation",
		Content:    "需要更新访问时间的记忆",
		Importance: 0.7,
	}
	id, err := repo.Create(mem)
	if err != nil {
		t.Fatalf("创建记忆失败: %v", err)
	}

	// 更新访问时间
	if err := repo.UpdateLastAccessed(id); err != nil {
		t.Fatalf("更新访问时间失败: %v", err)
	}

	// 验证访问时间已更新
	memories, err := repo.GetByStudentAndTeacher(studentID, teacherID, 10)
	if err != nil {
		t.Fatalf("查询记忆失败: %v", err)
	}
	if len(memories) == 0 {
		t.Fatal("应找到记忆")
	}
	if memories[0].LastAccessed == nil {
		t.Error("LastAccessed 应已更新")
	}
}

func TestMemoryRepository_UpdateLastAccessed_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMemoryRepository(db.DB)

	err := repo.UpdateLastAccessed(99999)
	if err == nil {
		t.Fatal("更新不存在的记忆应返回错误")
	}
}

// ==================== GetByOpenID 测试 ====================

func TestGetByOpenID_Found(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db.DB)

	// 先创建带 openid 的用户
	user := &User{
		Username: "wx_user_found",
		Password: "hashed_password",
		Role:     "student",
		Nickname: "微信用户",
		OpenID:   "openid_test_found_123",
	}
	createdID, err := repo.Create(user)
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	// 根据 openid 查询
	found, err := repo.GetByOpenID("openid_test_found_123")
	if err != nil {
		t.Fatalf("GetByOpenID 查询失败: %v", err)
	}
	if found == nil {
		t.Fatal("应找到用户")
	}
	if found.ID != createdID {
		t.Errorf("用户 ID 不匹配: got %d, want %d", found.ID, createdID)
	}
	if found.Username != "wx_user_found" {
		t.Errorf("用户名不匹配: got %q, want %q", found.Username, "wx_user_found")
	}
	if found.OpenID != "openid_test_found_123" {
		t.Errorf("OpenID 不匹配: got %q", found.OpenID)
	}
	if found.Nickname != "微信用户" {
		t.Errorf("昵称不匹配: got %q, want %q", found.Nickname, "微信用户")
	}
}

func TestGetByOpenID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db.DB)

	// 查询不存在的 openid
	found, err := repo.GetByOpenID("nonexistent_openid_xyz")
	if err != nil {
		t.Fatalf("查询不应返回错误: %v", err)
	}
	if found != nil {
		t.Fatal("不存在的 openid 应返回 nil")
	}
}

// ==================== CreateWithOpenID 测试 ====================

func TestCreateWithOpenID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db.DB)

	user := &User{
		Username: "wx_用户_abc123",
		Password: "hashed_random_password",
		Role:     "", // 新微信用户 role 为空
		Nickname: "", // 新微信用户 nickname 为空
		OpenID:   "openid_create_test_456",
	}

	id, err := repo.CreateWithOpenID(user)
	if err != nil {
		t.Fatalf("CreateWithOpenID 失败: %v", err)
	}
	if id <= 0 {
		t.Errorf("用户 ID 应大于 0: got %d", id)
	}

	// 验证可以通过 openid 查回来
	found, err := repo.GetByOpenID("openid_create_test_456")
	if err != nil {
		t.Fatalf("查询创建的用户失败: %v", err)
	}
	if found == nil {
		t.Fatal("应找到刚创建的用户")
	}
	if found.ID != id {
		t.Errorf("用户 ID 不匹配: got %d, want %d", found.ID, id)
	}
	if found.Username != "wx_用户_abc123" {
		t.Errorf("用户名不匹配: got %q", found.Username)
	}
	if found.Role != "" {
		t.Errorf("新微信用户 role 应为空: got %q", found.Role)
	}
	if found.OpenID != "openid_create_test_456" {
		t.Errorf("OpenID 不匹配: got %q", found.OpenID)
	}
}

// ==================== UpdateRoleAndNickname 测试 ====================

func TestUpdateRoleAndNickname(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db.DB)

	// 先创建一个 role 为空的用户（模拟微信新用户）
	user := &User{
		Username: "wx_update_test",
		Password: "hashed_password",
		Role:     "",
		Nickname: "",
		OpenID:   "openid_update_test_789",
	}
	id, err := repo.Create(user)
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	// 更新角色和昵称
	err = repo.UpdateRoleAndNickname(id, "teacher", "王老师")
	if err != nil {
		t.Fatalf("UpdateRoleAndNickname 失败: %v", err)
	}

	// 验证更新结果
	found, err := repo.GetByID(id)
	if err != nil {
		t.Fatalf("查询更新后的用户失败: %v", err)
	}
	if found == nil {
		t.Fatal("应找到用户")
	}
	if found.Role != "teacher" {
		t.Errorf("角色更新不正确: got %q, want %q", found.Role, "teacher")
	}
	if found.Nickname != "王老师" {
		t.Errorf("昵称更新不正确: got %q, want %q", found.Nickname, "王老师")
	}
}

func TestUpdateRoleAndNickname_UserNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db.DB)

	// 更新不存在的用户
	err := repo.UpdateRoleAndNickname(99999, "student", "不存在的用户")
	if err == nil {
		t.Fatal("更新不存在的用户应返回错误")
	}
}

// ==================== GetTeachers 教师列表测试 ====================

// TestGetTeachers_WithDocuments 有教师有文档，验证 document_count 正确
func TestGetTeachers_WithDocuments(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)

	// 创建教师
	teacherID, err := userRepo.Create(&User{
		Username: "teacher_with_docs",
		Password: "password",
		Role:     "teacher",
		Nickname: "有文档的老师",
	})
	if err != nil {
		t.Fatalf("创建教师失败: %v", err)
	}

	// 创建教师分身（满足 knowledge_items 外键约束）
	personaID, err := personaRepo.Create(&Persona{UserID: teacherID, Role: "teacher", Nickname: "教师分身", School: "测试学校"})
	if err != nil {
		t.Fatalf("创建分身失败: %v", err)
	}

	// 为教师创建 3 个 active 知识条目 + 1 个 archived 知识条目（迭代8迁移到 knowledge_items 表）
	for i := 0; i < 3; i++ {
		_, err := db.DB.Exec(
			`INSERT INTO knowledge_items (teacher_id, persona_id, title, content, item_type, status) VALUES (?, ?, ?, ?, 'text', 'active')`,
			teacherID, personaID, fmt.Sprintf("文档%d", i+1), "内容",
		)
		if err != nil {
			t.Fatalf("创建知识条目失败: %v", err)
		}
	}
	// 创建一个 archived 知识条目，不应计入 document_count
	_, err = db.DB.Exec(
		`INSERT INTO knowledge_items (teacher_id, persona_id, title, content, item_type, status) VALUES (?, ?, ?, ?, 'text', 'archived')`,
		teacherID, personaID, "已归档文档", "内容",
	)
	if err != nil {
		t.Fatalf("创建归档知识条目失败: %v", err)
	}

	// 查询教师列表
	teachers, total, err := userRepo.GetTeachers(0, 10)
	if err != nil {
		t.Fatalf("GetTeachers 失败: %v", err)
	}
	if total != 1 {
		t.Errorf("教师总数不匹配: got %d, want 1", total)
	}
	if len(teachers) != 1 {
		t.Fatalf("教师列表长度不匹配: got %d, want 1", len(teachers))
	}
	if teachers[0].ID != teacherID {
		t.Errorf("教师 ID 不匹配: got %d, want %d", teachers[0].ID, teacherID)
	}
	if teachers[0].Nickname != "有文档的老师" {
		t.Errorf("教师昵称不匹配: got %q", teachers[0].Nickname)
	}
	// 只统计 active 文档，应为 3
	if teachers[0].DocumentCount != 3 {
		t.Errorf("文档数不匹配: got %d, want 3", teachers[0].DocumentCount)
	}
}

// TestGetTeachers_Empty 没有教师，返回空列表
func TestGetTeachers_Empty(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db.DB)

	// 只创建学生，不创建教师
	_, err := userRepo.Create(&User{
		Username: "student_only",
		Password: "password",
		Role:     "student",
	})
	if err != nil {
		t.Fatalf("创建学生失败: %v", err)
	}

	teachers, total, err := userRepo.GetTeachers(0, 10)
	if err != nil {
		t.Fatalf("GetTeachers 失败: %v", err)
	}
	if total != 0 {
		t.Errorf("教师总数应为 0: got %d", total)
	}
	if len(teachers) != 0 {
		t.Errorf("教师列表应为空: got %d", len(teachers))
	}
}

// TestGetTeachers_Pagination 分页测试，验证 offset/limit 正确
func TestGetTeachers_Pagination(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db.DB)

	// 创建 5 个教师
	for i := 0; i < 5; i++ {
		_, err := userRepo.Create(&User{
			Username: fmt.Sprintf("teacher_page_%d", i),
			Password: "password",
			Role:     "teacher",
			Nickname: fmt.Sprintf("教师%d", i),
		})
		if err != nil {
			t.Fatalf("创建教师失败: %v", err)
		}
	}

	// 第一页：offset=0, limit=2
	teachers1, total, err := userRepo.GetTeachers(0, 2)
	if err != nil {
		t.Fatalf("查询第一页失败: %v", err)
	}
	if total != 5 {
		t.Errorf("总数不匹配: got %d, want 5", total)
	}
	if len(teachers1) != 2 {
		t.Errorf("第一页数量不匹配: got %d, want 2", len(teachers1))
	}

	// 第二页：offset=2, limit=2
	teachers2, total2, err := userRepo.GetTeachers(2, 2)
	if err != nil {
		t.Fatalf("查询第二页失败: %v", err)
	}
	if total2 != 5 {
		t.Errorf("总数不匹配: got %d, want 5", total2)
	}
	if len(teachers2) != 2 {
		t.Errorf("第二页数量不匹配: got %d, want 2", len(teachers2))
	}

	// 第三页：offset=4, limit=2（只剩 1 个）
	teachers3, _, err := userRepo.GetTeachers(4, 2)
	if err != nil {
		t.Fatalf("查询第三页失败: %v", err)
	}
	if len(teachers3) != 1 {
		t.Errorf("第三页数量不匹配: got %d, want 1", len(teachers3))
	}

	// 验证三页的教师 ID 互不重复
	idSet := make(map[int64]bool)
	for _, ts := range [][]TeacherWithDocCount{teachers1, teachers2, teachers3} {
		for _, t := range ts {
			if idSet[t.ID] {
				// 使用 fmt 而非 t.Errorf 避免与循环变量冲突
				panic(fmt.Sprintf("教师 ID %d 在分页中重复出现", t.ID))
			}
			idSet[t.ID] = true
		}
	}
	if len(idSet) != 5 {
		t.Errorf("分页后教师总数不匹配: got %d, want 5", len(idSet))
	}
}

// TestGetTeachers_OnlyTeacherRole 只返回 role=teacher 的用户，不返回 student
func TestGetTeachers_OnlyTeacherRole(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db.DB)

	// 创建 2 个教师 + 2 个学生
	for i := 0; i < 2; i++ {
		_, err := userRepo.Create(&User{
			Username: fmt.Sprintf("teacher_role_%d", i),
			Password: "password",
			Role:     "teacher",
			Nickname: fmt.Sprintf("角色教师%d", i),
		})
		if err != nil {
			t.Fatalf("创建教师失败: %v", err)
		}
	}
	for i := 0; i < 2; i++ {
		_, err := userRepo.Create(&User{
			Username: fmt.Sprintf("student_role_%d", i),
			Password: "password",
			Role:     "student",
		})
		if err != nil {
			t.Fatalf("创建学生失败: %v", err)
		}
	}

	teachers, total, err := userRepo.GetTeachers(0, 10)
	if err != nil {
		t.Fatalf("GetTeachers 失败: %v", err)
	}
	if total != 2 {
		t.Errorf("教师总数不匹配: got %d, want 2", total)
	}
	if len(teachers) != 2 {
		t.Errorf("教师列表长度不匹配: got %d, want 2", len(teachers))
	}
	// 验证所有返回的用户都是 teacher 角色
	for _, teacher := range teachers {
		if teacher.Role != "teacher" {
			t.Errorf("返回了非教师角色: id=%d, role=%q", teacher.ID, teacher.Role)
		}
	}
}

// ==================== GetUserStats 用户统计测试 ====================

// TestGetUserStats_Student 学生统计：conversation_count + memory_count
func TestGetUserStats_Student(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db.DB)
	convRepo := NewConversationRepository(db.DB)
	memRepo := NewMemoryRepository(db.DB)

	// 创建学生和教师
	studentID, err := userRepo.Create(&User{
		Username: "stats_student",
		Password: "password",
		Role:     "student",
	})
	if err != nil {
		t.Fatalf("创建学生失败: %v", err)
	}
	teacherID, err := userRepo.Create(&User{
		Username: "stats_teacher",
		Password: "password",
		Role:     "teacher",
	})
	if err != nil {
		t.Fatalf("创建教师失败: %v", err)
	}

	// 创建 3 条对话
	for i := 0; i < 3; i++ {
		_, err := convRepo.Create(&Conversation{
			StudentID:  studentID,
			TeacherID:  teacherID,
			SessionID:  "session-stats",
			Role:       "user",
			Content:    fmt.Sprintf("对话%d", i),
			TokenCount: 10,
		})
		if err != nil {
			t.Fatalf("创建对话失败: %v", err)
		}
	}

	// 创建 2 条记忆
	for i := 0; i < 2; i++ {
		_, err := memRepo.Create(&Memory{
			StudentID:  studentID,
			TeacherID:  teacherID,
			MemoryType: "conversation",
			Content:    fmt.Sprintf("记忆%d", i),
			Importance: 0.5,
		})
		if err != nil {
			t.Fatalf("创建记忆失败: %v", err)
		}
	}

	// 查询学生统计
	stats, err := userRepo.GetUserStats(studentID, "student")
	if err != nil {
		t.Fatalf("GetUserStats 失败: %v", err)
	}
	if stats["conversation_count"] != 3 {
		t.Errorf("对话数不匹配: got %d, want 3", stats["conversation_count"])
	}
	if stats["memory_count"] != 2 {
		t.Errorf("记忆数不匹配: got %d, want 2", stats["memory_count"])
	}
}

// TestGetUserStats_Teacher 教师统计：document_count + conversation_count
func TestGetUserStats_Teacher(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db.DB)
	convRepo := NewConversationRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)

	// 创建教师和学生
	teacherID, err := userRepo.Create(&User{
		Username: "stats_teacher_t",
		Password: "password",
		Role:     "teacher",
	})
	if err != nil {
		t.Fatalf("创建教师失败: %v", err)
	}
	studentID, err := userRepo.Create(&User{
		Username: "stats_student_t",
		Password: "password",
		Role:     "student",
	})
	if err != nil {
		t.Fatalf("创建学生失败: %v", err)
	}

	// 创建教师分身（满足 knowledge_items 外键约束）
	personaID, err := personaRepo.Create(&Persona{UserID: teacherID, Role: "teacher", Nickname: "统计教师分身", School: "统计学校"})
	if err != nil {
		t.Fatalf("创建分身失败: %v", err)
	}

	// 创建 4 个 active 知识条目 + 1 个 archived 知识条目（迭代8迁移到 knowledge_items 表）
	for i := 0; i < 4; i++ {
		_, err := db.DB.Exec(
			`INSERT INTO knowledge_items (teacher_id, persona_id, title, content, item_type, status) VALUES (?, ?, ?, ?, 'text', 'active')`,
			teacherID, personaID, fmt.Sprintf("文档%d", i), "内容",
		)
		if err != nil {
			t.Fatalf("创建知识条目失败: %v", err)
		}
	}
	_, err = db.DB.Exec(
		`INSERT INTO knowledge_items (teacher_id, persona_id, title, content, item_type, status) VALUES (?, ?, ?, ?, 'text', 'archived')`,
		teacherID, personaID, "归档文档", "内容",
	)
	if err != nil {
		t.Fatalf("创建归档知识条目失败: %v", err)
	}

	// 创建 5 条对话（学生向该教师提问）
	for i := 0; i < 5; i++ {
		_, err := convRepo.Create(&Conversation{
			StudentID:  studentID,
			TeacherID:  teacherID,
			SessionID:  "session-teacher-stats",
			Role:       "user",
			Content:    fmt.Sprintf("提问%d", i),
			TokenCount: 10,
		})
		if err != nil {
			t.Fatalf("创建对话失败: %v", err)
		}
	}

	// 查询教师统计
	stats, err := userRepo.GetUserStats(teacherID, "teacher")
	if err != nil {
		t.Fatalf("GetUserStats 失败: %v", err)
	}
	// 只统计 active 文档
	if stats["document_count"] != 4 {
		t.Errorf("文档数不匹配: got %d, want 4", stats["document_count"])
	}
	if stats["conversation_count"] != 5 {
		t.Errorf("被提问数不匹配: got %d, want 5", stats["conversation_count"])
	}
}

// TestGetUserStats_NoData 用户没有任何数据，统计全为 0
func TestGetUserStats_NoData(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db.DB)

	// 创建学生，不创建任何对话或记忆
	studentID, err := userRepo.Create(&User{
		Username: "stats_empty_student",
		Password: "password",
		Role:     "student",
	})
	if err != nil {
		t.Fatalf("创建学生失败: %v", err)
	}

	stats, err := userRepo.GetUserStats(studentID, "student")
	if err != nil {
		t.Fatalf("GetUserStats 失败: %v", err)
	}
	if stats["conversation_count"] != 0 {
		t.Errorf("对话数应为 0: got %d", stats["conversation_count"])
	}
	if stats["memory_count"] != 0 {
		t.Errorf("记忆数应为 0: got %d", stats["memory_count"])
	}

	// 创建教师，不创建任何文档或对话
	teacherID, err := userRepo.Create(&User{
		Username: "stats_empty_teacher",
		Password: "password",
		Role:     "teacher",
	})
	if err != nil {
		t.Fatalf("创建教师失败: %v", err)
	}

	statsT, err := userRepo.GetUserStats(teacherID, "teacher")
	if err != nil {
		t.Fatalf("GetUserStats 教师查询失败: %v", err)
	}
	if statsT["document_count"] != 0 {
		t.Errorf("文档数应为 0: got %d", statsT["document_count"])
	}
	if statsT["conversation_count"] != 0 {
		t.Errorf("被提问数应为 0: got %d", statsT["conversation_count"])
	}
}

// ==================== GetConversationsByStudent 对话历史增强测试 ====================

// TestGetConversationsByStudent 查询学生所有对话（不传 teacher_id）
func TestGetConversationsByStudent(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db.DB)
	convRepo := NewConversationRepository(db.DB)

	// 创建 1 个学生 + 2 个教师
	studentID, err := userRepo.Create(&User{
		Username: "conv_student",
		Password: "password",
		Role:     "student",
	})
	if err != nil {
		t.Fatalf("创建学生失败: %v", err)
	}
	teacher1ID, err := userRepo.Create(&User{
		Username: "conv_teacher1",
		Password: "password",
		Role:     "teacher",
		Nickname: "对话教师1",
	})
	if err != nil {
		t.Fatalf("创建教师1失败: %v", err)
	}
	teacher2ID, err := userRepo.Create(&User{
		Username: "conv_teacher2",
		Password: "password",
		Role:     "teacher",
		Nickname: "对话教师2",
	})
	if err != nil {
		t.Fatalf("创建教师2失败: %v", err)
	}

	// 与教师1对话 2 条
	for i := 0; i < 2; i++ {
		_, err := convRepo.Create(&Conversation{
			StudentID:  studentID,
			TeacherID:  teacher1ID,
			SessionID:  "session-t1",
			Role:       "user",
			Content:    fmt.Sprintf("对话T1-%d", i),
			TokenCount: 10,
		})
		if err != nil {
			t.Fatalf("创建对话失败: %v", err)
		}
	}

	// 与教师2对话 3 条
	for i := 0; i < 3; i++ {
		_, err := convRepo.Create(&Conversation{
			StudentID:  studentID,
			TeacherID:  teacher2ID,
			SessionID:  "session-t2",
			Role:       "user",
			Content:    fmt.Sprintf("对话T2-%d", i),
			TokenCount: 10,
		})
		if err != nil {
			t.Fatalf("创建对话失败: %v", err)
		}
	}

	// 查询学生所有对话，应返回 5 条
	convs, total, err := convRepo.GetConversationsByStudent(studentID, 0, 20)
	if err != nil {
		t.Fatalf("GetConversationsByStudent 失败: %v", err)
	}
	if total != 5 {
		t.Errorf("对话总数不匹配: got %d, want 5", total)
	}
	if len(convs) != 5 {
		t.Errorf("对话列表长度不匹配: got %d, want 5", len(convs))
	}

	// 验证所有对话都属于该学生
	for _, conv := range convs {
		if conv.StudentID != studentID {
			t.Errorf("对话学生ID不匹配: got %d, want %d", conv.StudentID, studentID)
		}
	}
}

// TestGetConversationsByStudent_Empty 学生没有对话，返回空列表
func TestGetConversationsByStudent_Empty(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db.DB)
	convRepo := NewConversationRepository(db.DB)

	// 创建学生，不创建任何对话
	studentID, err := userRepo.Create(&User{
		Username: "conv_empty_student",
		Password: "password",
		Role:     "student",
	})
	if err != nil {
		t.Fatalf("创建学生失败: %v", err)
	}

	convs, total, err := convRepo.GetConversationsByStudent(studentID, 0, 10)
	if err != nil {
		t.Fatalf("GetConversationsByStudent 失败: %v", err)
	}
	if total != 0 {
		t.Errorf("对话总数应为 0: got %d", total)
	}
	if len(convs) != 0 {
		t.Errorf("对话列表应为空: got %d", len(convs))
	}
}

// TestGetConversationsBySession 按 session_id 筛选对话
func TestGetConversationsBySession(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db.DB)
	convRepo := NewConversationRepository(db.DB)

	studentID, err := userRepo.Create(&User{
		Username: "session_student",
		Password: "password",
		Role:     "student",
	})
	if err != nil {
		t.Fatalf("创建学生失败: %v", err)
	}
	teacherID, err := userRepo.Create(&User{
		Username: "session_teacher",
		Password: "password",
		Role:     "teacher",
	})
	if err != nil {
		t.Fatalf("创建教师失败: %v", err)
	}

	// session-A 创建 3 条对话
	for i := 0; i < 3; i++ {
		_, err := convRepo.Create(&Conversation{
			StudentID:  studentID,
			TeacherID:  teacherID,
			SessionID:  "session-A",
			Role:       "user",
			Content:    fmt.Sprintf("A消息%d", i),
			TokenCount: 10,
		})
		if err != nil {
			t.Fatalf("创建对话失败: %v", err)
		}
	}

	// session-B 创建 2 条对话
	for i := 0; i < 2; i++ {
		_, err := convRepo.Create(&Conversation{
			StudentID:  studentID,
			TeacherID:  teacherID,
			SessionID:  "session-B",
			Role:       "user",
			Content:    fmt.Sprintf("B消息%d", i),
			TokenCount: 10,
		})
		if err != nil {
			t.Fatalf("创建对话失败: %v", err)
		}
	}

	// 按 session-A 筛选，应返回 3 条
	convs, total, err := convRepo.GetConversationsBySession(studentID, "session-A", 0, 10)
	if err != nil {
		t.Fatalf("GetConversationsBySession 失败: %v", err)
	}
	if total != 3 {
		t.Errorf("对话总数不匹配: got %d, want 3", total)
	}
	if len(convs) != 3 {
		t.Errorf("对话列表长度不匹配: got %d, want 3", len(convs))
	}
	// 验证所有返回的对话都是 session-A
	for _, conv := range convs {
		if conv.SessionID != "session-A" {
			t.Errorf("SessionID 不匹配: got %q, want %q", conv.SessionID, "session-A")
		}
	}

	// 按 session-B 筛选，应返回 2 条
	convsB, totalB, err := convRepo.GetConversationsBySession(studentID, "session-B", 0, 10)
	if err != nil {
		t.Fatalf("GetConversationsBySession session-B 失败: %v", err)
	}
	if totalB != 2 {
		t.Errorf("session-B 对话总数不匹配: got %d, want 2", totalB)
	}
	if len(convsB) != 2 {
		t.Errorf("session-B 对话列表长度不匹配: got %d, want 2", len(convsB))
	}
}

// TestGetConversationsBySession_NotFound session_id 不存在，返回空列表
func TestGetConversationsBySession_NotFound(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db.DB)
	convRepo := NewConversationRepository(db.DB)

	studentID, err := userRepo.Create(&User{
		Username: "session_notfound_student",
		Password: "password",
		Role:     "student",
	})
	if err != nil {
		t.Fatalf("创建学生失败: %v", err)
	}

	// 查询不存在的 session_id
	convs, total, err := convRepo.GetConversationsBySession(studentID, "nonexistent-session", 0, 10)
	if err != nil {
		t.Fatalf("GetConversationsBySession 失败: %v", err)
	}
	if total != 0 {
		t.Errorf("对话总数应为 0: got %d", total)
	}
	if len(convs) != 0 {
		t.Errorf("对话列表应为空: got %d", len(convs))
	}
}

// ==================== GetSessionsByStudent 会话列表测试 ====================

// TestGetSessionsByStudent_WithSessions 学生有多个会话（与不同教师），验证摘要信息正确
func TestGetSessionsByStudent_WithSessions(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db.DB)
	convRepo := NewConversationRepository(db.DB)

	// 创建学生
	studentID, err := userRepo.Create(&User{
		Username: "sess_student",
		Password: "password",
		Role:     "student",
	})
	if err != nil {
		t.Fatalf("创建学生失败: %v", err)
	}

	// 创建教师1
	teacher1ID, err := userRepo.Create(&User{
		Username: "sess_teacher1",
		Password: "password",
		Role:     "teacher",
		Nickname: "张老师",
	})
	if err != nil {
		t.Fatalf("创建教师1失败: %v", err)
	}

	// 创建教师2
	teacher2ID, err := userRepo.Create(&User{
		Username: "sess_teacher2",
		Password: "password",
		Role:     "teacher",
		Nickname: "李老师",
	})
	if err != nil {
		t.Fatalf("创建教师2失败: %v", err)
	}

	// session-1: 与教师1的会话，2条消息
	_, err = convRepo.Create(&Conversation{
		StudentID: studentID, TeacherID: teacher1ID,
		SessionID: "session-1", Role: "user",
		Content: "你好张老师", TokenCount: 5,
	})
	if err != nil {
		t.Fatalf("创建对话失败: %v", err)
	}
	_, err = convRepo.Create(&Conversation{
		StudentID: studentID, TeacherID: teacher1ID,
		SessionID: "session-1", Role: "assistant",
		Content: "你好同学，有什么问题吗？", TokenCount: 10,
	})
	if err != nil {
		t.Fatalf("创建对话失败: %v", err)
	}

	// session-2: 与教师2的会话，3条消息
	_, err = convRepo.Create(&Conversation{
		StudentID: studentID, TeacherID: teacher2ID,
		SessionID: "session-2", Role: "user",
		Content: "李老师请教一个问题", TokenCount: 8,
	})
	if err != nil {
		t.Fatalf("创建对话失败: %v", err)
	}
	_, err = convRepo.Create(&Conversation{
		StudentID: studentID, TeacherID: teacher2ID,
		SessionID: "session-2", Role: "assistant",
		Content: "请说", TokenCount: 3,
	})
	if err != nil {
		t.Fatalf("创建对话失败: %v", err)
	}
	_, err = convRepo.Create(&Conversation{
		StudentID: studentID, TeacherID: teacher2ID,
		SessionID: "session-2", Role: "user",
		Content: "什么是量子力学？", TokenCount: 7,
	})
	if err != nil {
		t.Fatalf("创建对话失败: %v", err)
	}

	// 查询会话列表
	sessions, total, err := convRepo.GetSessionsByStudent(studentID, 0, 10)
	if err != nil {
		t.Fatalf("GetSessionsByStudent 失败: %v", err)
	}

	// 验证总数
	if total != 2 {
		t.Errorf("会话总数不匹配: got %d, want 2", total)
	}
	if len(sessions) != 2 {
		t.Fatalf("会话列表长度不匹配: got %d, want 2", len(sessions))
	}

	// 按 updated_at 降序排列，session-2 最后插入的消息更晚，应排在第一位
	if sessions[0].SessionID != "session-2" {
		t.Errorf("第一个会话应为 session-2: got %q", sessions[0].SessionID)
	}
	if sessions[1].SessionID != "session-1" {
		t.Errorf("第二个会话应为 session-1: got %q", sessions[1].SessionID)
	}

	// 验证 teacher_nickname
	if sessions[0].TeacherNickname != "李老师" {
		t.Errorf("session-2 教师昵称不匹配: got %q, want %q", sessions[0].TeacherNickname, "李老师")
	}
	if sessions[1].TeacherNickname != "张老师" {
		t.Errorf("session-1 教师昵称不匹配: got %q, want %q", sessions[1].TeacherNickname, "张老师")
	}

	// 验证 last_message 是该会话最新的消息
	if sessions[0].LastMessage != "什么是量子力学？" {
		t.Errorf("session-2 last_message 不匹配: got %q, want %q", sessions[0].LastMessage, "什么是量子力学？")
	}
	if sessions[1].LastMessage != "你好同学，有什么问题吗？" {
		t.Errorf("session-1 last_message 不匹配: got %q, want %q", sessions[1].LastMessage, "你好同学，有什么问题吗？")
	}

	// 验证 message_count
	if sessions[0].MessageCount != 3 {
		t.Errorf("session-2 message_count 不匹配: got %d, want 3", sessions[0].MessageCount)
	}
	if sessions[1].MessageCount != 2 {
		t.Errorf("session-1 message_count 不匹配: got %d, want 2", sessions[1].MessageCount)
	}
}

// TestGetSessionsByStudent_Empty 学生没有任何对话，返回空列表和 total=0
func TestGetSessionsByStudent_Empty(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db.DB)
	convRepo := NewConversationRepository(db.DB)

	// 创建学生，不创建任何对话
	studentID, err := userRepo.Create(&User{
		Username: "sess_empty_student",
		Password: "password",
		Role:     "student",
	})
	if err != nil {
		t.Fatalf("创建学生失败: %v", err)
	}

	sessions, total, err := convRepo.GetSessionsByStudent(studentID, 0, 10)
	if err != nil {
		t.Fatalf("GetSessionsByStudent 失败: %v", err)
	}
	if total != 0 {
		t.Errorf("会话总数应为 0: got %d", total)
	}
	if len(sessions) != 0 {
		t.Errorf("会话列表应为空: got %d", len(sessions))
	}
}

// TestGetSessionsByStudent_Pagination 分页测试，创建3个会话，每页1个，验证分页正确
func TestGetSessionsByStudent_Pagination(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db.DB)
	convRepo := NewConversationRepository(db.DB)

	// 创建学生
	studentID, err := userRepo.Create(&User{
		Username: "sess_page_student",
		Password: "password",
		Role:     "student",
	})
	if err != nil {
		t.Fatalf("创建学生失败: %v", err)
	}

	// 创建教师
	teacherID, err := userRepo.Create(&User{
		Username: "sess_page_teacher",
		Password: "password",
		Role:     "teacher",
		Nickname: "分页老师",
	})
	if err != nil {
		t.Fatalf("创建教师失败: %v", err)
	}

	// 创建 3 个不同 session 的对话
	for i := 0; i < 3; i++ {
		_, err := convRepo.Create(&Conversation{
			StudentID: studentID, TeacherID: teacherID,
			SessionID:  fmt.Sprintf("page-session-%d", i),
			Role:       "user",
			Content:    fmt.Sprintf("分页消息%d", i),
			TokenCount: 5,
		})
		if err != nil {
			t.Fatalf("创建对话失败: %v", err)
		}
	}

	// 第一页：offset=0, limit=1
	sessions1, total, err := convRepo.GetSessionsByStudent(studentID, 0, 1)
	if err != nil {
		t.Fatalf("查询第一页失败: %v", err)
	}
	if total != 3 {
		t.Errorf("总数不匹配: got %d, want 3", total)
	}
	if len(sessions1) != 1 {
		t.Errorf("第一页数量不匹配: got %d, want 1", len(sessions1))
	}

	// 第二页：offset=1, limit=1
	sessions2, total2, err := convRepo.GetSessionsByStudent(studentID, 1, 1)
	if err != nil {
		t.Fatalf("查询第二页失败: %v", err)
	}
	if total2 != 3 {
		t.Errorf("总数不匹配: got %d, want 3", total2)
	}
	if len(sessions2) != 1 {
		t.Errorf("第二页数量不匹配: got %d, want 1", len(sessions2))
	}

	// 第三页：offset=2, limit=1
	sessions3, _, err := convRepo.GetSessionsByStudent(studentID, 2, 1)
	if err != nil {
		t.Fatalf("查询第三页失败: %v", err)
	}
	if len(sessions3) != 1 {
		t.Errorf("第三页数量不匹配: got %d, want 1", len(sessions3))
	}

	// 超出范围：offset=3, limit=1
	sessions4, _, err := convRepo.GetSessionsByStudent(studentID, 3, 1)
	if err != nil {
		t.Fatalf("查询超出范围页失败: %v", err)
	}
	if len(sessions4) != 0 {
		t.Errorf("超出范围页应为空: got %d", len(sessions4))
	}

	// 验证三页的 session_id 互不重复
	idSet := make(map[string]bool)
	allSessions := [][]SessionSummary{sessions1, sessions2, sessions3}
	for _, page := range allSessions {
		for _, s := range page {
			if idSet[s.SessionID] {
				t.Errorf("session_id %q 在分页中重复出现", s.SessionID)
			}
			idSet[s.SessionID] = true
		}
	}
	if len(idSet) != 3 {
		t.Errorf("分页后会话总数不匹配: got %d, want 3", len(idSet))
	}
}

// TestGetSessionsByStudent_LongMessage last_message 超过 100 字符时被截断
func TestGetSessionsByStudent_LongMessage(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db.DB)
	convRepo := NewConversationRepository(db.DB)

	// 创建学生
	studentID, err := userRepo.Create(&User{
		Username: "sess_long_student",
		Password: "password",
		Role:     "student",
	})
	if err != nil {
		t.Fatalf("创建学生失败: %v", err)
	}

	// 创建教师
	teacherID, err := userRepo.Create(&User{
		Username: "sess_long_teacher",
		Password: "password",
		Role:     "teacher",
		Nickname: "长消息老师",
	})
	if err != nil {
		t.Fatalf("创建教师失败: %v", err)
	}

	// 构造超过 100 个字符的消息（使用中文，每个字符是一个 rune）
	longMsg := ""
	for i := 0; i < 120; i++ {
		longMsg += "测"
	}

	_, err = convRepo.Create(&Conversation{
		StudentID: studentID, TeacherID: teacherID,
		SessionID: "long-session", Role: "user",
		Content: longMsg, TokenCount: 120,
	})
	if err != nil {
		t.Fatalf("创建对话失败: %v", err)
	}

	sessions, total, err := convRepo.GetSessionsByStudent(studentID, 0, 10)
	if err != nil {
		t.Fatalf("GetSessionsByStudent 失败: %v", err)
	}
	if total != 1 {
		t.Errorf("会话总数不匹配: got %d, want 1", total)
	}
	if len(sessions) != 1 {
		t.Fatalf("会话列表长度不匹配: got %d, want 1", len(sessions))
	}

	// 验证 last_message 被截断到 100 个 rune + "..."
	runes := []rune(sessions[0].LastMessage)
	// 截断后应该是 100 个 "测" + "..."，即 103 个 rune
	expectedRunes := 100 + len([]rune("..."))
	if len(runes) != expectedRunes {
		t.Errorf("截断后 rune 长度不匹配: got %d, want %d", len(runes), expectedRunes)
	}
	// 验证以 "..." 结尾
	if sessions[0].LastMessage[len(sessions[0].LastMessage)-3:] != "..." {
		t.Errorf("截断消息应以 '...' 结尾: got %q", sessions[0].LastMessage[len(sessions[0].LastMessage)-10:])
	}

	// 额外验证：恰好 100 个字符的消息不应被截断
	exactMsg := ""
	for i := 0; i < 100; i++ {
		exactMsg += "正"
	}
	// 创建另一个 session 使用恰好 100 字符的消息
	_, err = convRepo.Create(&Conversation{
		StudentID: studentID, TeacherID: teacherID,
		SessionID: "exact-session", Role: "user",
		Content: exactMsg, TokenCount: 100,
	})
	if err != nil {
		t.Fatalf("创建对话失败: %v", err)
	}

	sessions2, _, err := convRepo.GetSessionsByStudent(studentID, 0, 10)
	if err != nil {
		t.Fatalf("GetSessionsByStudent 失败: %v", err)
	}
	// 找到 exact-session
	for _, s := range sessions2 {
		if s.SessionID == "exact-session" {
			if s.LastMessage != exactMsg {
				t.Errorf("恰好 100 字符的消息不应被截断")
			}
			return
		}
	}
	t.Error("未找到 exact-session")
}

// TestGetSessionsByStudent_MessageCount 同一个 session 有多条消息，验证 message_count 正确统计
func TestGetSessionsByStudent_MessageCount(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db.DB)
	convRepo := NewConversationRepository(db.DB)

	// 创建学生
	studentID, err := userRepo.Create(&User{
		Username: "sess_count_student",
		Password: "password",
		Role:     "student",
	})
	if err != nil {
		t.Fatalf("创建学生失败: %v", err)
	}

	// 创建教师
	teacherID, err := userRepo.Create(&User{
		Username: "sess_count_teacher",
		Password: "password",
		Role:     "teacher",
		Nickname: "计数老师",
	})
	if err != nil {
		t.Fatalf("创建教师失败: %v", err)
	}

	// session-count-1: 创建 5 条消息（模拟多轮对话）
	for i := 0; i < 5; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		_, err := convRepo.Create(&Conversation{
			StudentID: studentID, TeacherID: teacherID,
			SessionID:  "session-count-1",
			Role:       role,
			Content:    fmt.Sprintf("消息%d", i),
			TokenCount: 5,
		})
		if err != nil {
			t.Fatalf("创建对话失败: %v", err)
		}
	}

	// session-count-2: 创建 1 条消息
	_, err = convRepo.Create(&Conversation{
		StudentID: studentID, TeacherID: teacherID,
		SessionID: "session-count-2", Role: "user",
		Content: "唯一的消息", TokenCount: 5,
	})
	if err != nil {
		t.Fatalf("创建对话失败: %v", err)
	}

	// 查询会话列表
	sessions, total, err := convRepo.GetSessionsByStudent(studentID, 0, 10)
	if err != nil {
		t.Fatalf("GetSessionsByStudent 失败: %v", err)
	}
	if total != 2 {
		t.Errorf("会话总数不匹配: got %d, want 2", total)
	}
	if len(sessions) != 2 {
		t.Fatalf("会话列表长度不匹配: got %d, want 2", len(sessions))
	}

	// 按 updated_at DESC 排序，session-count-2 最后插入应排第一
	// 构建 session_id -> SessionSummary 的映射方便验证
	sessionMap := make(map[string]SessionSummary)
	for _, s := range sessions {
		sessionMap[s.SessionID] = s
	}

	// 验证 session-count-1 的 message_count 为 5
	if s, ok := sessionMap["session-count-1"]; ok {
		if s.MessageCount != 5 {
			t.Errorf("session-count-1 message_count 不匹配: got %d, want 5", s.MessageCount)
		}
		// 验证 last_message 是最后一条消息
		if s.LastMessage != "消息4" {
			t.Errorf("session-count-1 last_message 不匹配: got %q, want %q", s.LastMessage, "消息4")
		}
		// 验证 last_message_role 是最后一条的角色（i=4, 偶数, role=user）
		if s.LastMessageRole != "user" {
			t.Errorf("session-count-1 last_message_role 不匹配: got %q, want %q", s.LastMessageRole, "user")
		}
	} else {
		t.Error("未找到 session-count-1")
	}

	// 验证 session-count-2 的 message_count 为 1
	if s, ok := sessionMap["session-count-2"]; ok {
		if s.MessageCount != 1 {
			t.Errorf("session-count-2 message_count 不匹配: got %d, want 1", s.MessageCount)
		}
		if s.LastMessage != "唯一的消息" {
			t.Errorf("session-count-2 last_message 不匹配: got %q, want %q", s.LastMessage, "唯一的消息")
		}
	} else {
		t.Error("未找到 session-count-2")
	}
}
