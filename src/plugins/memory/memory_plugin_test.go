package memory

import (
	"context"
	"fmt"
	"os"
	"testing"

	"digital-twin/src/backend/database"
	"digital-twin/src/harness/core"
)

// setupTestDB 创建测试用数据库
func setupTestDB(t *testing.T) *database.Database {
	t.Helper()
	tmpFile := t.TempDir() + "/test.db"
	db, err := database.NewDatabase(tmpFile)
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
		os.Remove(tmpFile)
	})
	return db
}

// createTestUsers 创建测试用的学生和教师
func createTestUsers(t *testing.T, db *database.Database) (studentID, teacherID int64) {
	t.Helper()
	userRepo := database.NewUserRepository(db.DB)

	teacherID, err := userRepo.Create(&database.User{
		Username: "teacher_test",
		Password: "hashed_password",
		Role:     "teacher",
		Nickname: "测试教师",
	})
	if err != nil {
		t.Fatalf("创建测试教师失败: %v", err)
	}

	studentID, err = userRepo.Create(&database.User{
		Username: "student_test",
		Password: "hashed_password",
		Role:     "student",
		Nickname: "测试学生",
	})
	if err != nil {
		t.Fatalf("创建测试学生失败: %v", err)
	}

	return studentID, teacherID
}

// setupMemoryPlugin 创建测试用记忆插件
func setupMemoryPlugin(t *testing.T) (*MemoryPlugin, *database.Database, int64, int64) {
	t.Helper()
	db := setupTestDB(t)
	studentID, teacherID := createTestUsers(t, db)

	plugin := NewMemoryPlugin("test-memory", db.DB)
	err := plugin.Init(map[string]interface{}{
		"retention.max_memories_per_user": 50,
	})
	if err != nil {
		t.Fatalf("初始化插件失败: %v", err)
	}

	return plugin, db, studentID, teacherID
}

// ==================== MemoryStore 测试 ====================

func TestMemoryStore_StoreAndRetrieve(t *testing.T) {
	db := setupTestDB(t)
	studentID, teacherID := createTestUsers(t, db)
	store := NewMemoryStore(database.NewMemoryRepository(db.DB))

	// 存储记忆
	memID, err := store.StoreMemory(studentID, teacherID, "concept", "学生理解了循环的概念", 0.8)
	if err != nil {
		t.Fatalf("存储记忆失败: %v", err)
	}
	if memID <= 0 {
		t.Fatal("memory_id 应 > 0")
	}

	// 检索记忆
	memories, err := store.RetrieveRelevant(studentID, teacherID, 10)
	if err != nil {
		t.Fatalf("检索记忆失败: %v", err)
	}
	if len(memories) != 1 {
		t.Fatalf("期望 1 条记忆, 实际=%d", len(memories))
	}
	if memories[0].Content != "学生理解了循环的概念" {
		t.Errorf("记忆内容不匹配: %s", memories[0].Content)
	}
}

func TestMemoryStore_StoreMemory_EmptyContent(t *testing.T) {
	db := setupTestDB(t)
	studentID, teacherID := createTestUsers(t, db)
	store := NewMemoryStore(database.NewMemoryRepository(db.DB))

	_, err := store.StoreMemory(studentID, teacherID, "concept", "", 0.5)
	if err == nil {
		t.Fatal("空内容应该返回错误")
	}
}

func TestMemoryStore_StoreMemory_InvalidIDs(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(database.NewMemoryRepository(db.DB))

	_, err := store.StoreMemory(0, 1, "concept", "内容", 0.5)
	if err == nil {
		t.Fatal("无效 studentID 应该返回错误")
	}

	_, err = store.StoreMemory(1, 0, "concept", "内容", 0.5)
	if err == nil {
		t.Fatal("无效 teacherID 应该返回错误")
	}
}

func TestMemoryStore_RetrieveRelevant_InvalidIDs(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(database.NewMemoryRepository(db.DB))

	_, err := store.RetrieveRelevant(0, 1, 10)
	if err == nil {
		t.Fatal("无效 studentID 应该返回错误")
	}
}

func TestMemoryStore_ListMemories(t *testing.T) {
	db := setupTestDB(t)
	studentID, teacherID := createTestUsers(t, db)
	store := NewMemoryStore(database.NewMemoryRepository(db.DB))

	// 存储多条记忆
	store.StoreMemory(studentID, teacherID, "concept", "概念记忆1", 0.8)
	store.StoreMemory(studentID, teacherID, "weakness", "薄弱点记忆1", 0.6)
	store.StoreMemory(studentID, teacherID, "concept", "概念记忆2", 0.9)

	// 列表查询（全部）
	memories, total, err := store.ListMemories(studentID, teacherID, "", 1, 10)
	if err != nil {
		t.Fatalf("列表查询失败: %v", err)
	}
	if total != 3 {
		t.Errorf("期望 total=3, 实际=%d", total)
	}
	if len(memories) != 3 {
		t.Errorf("期望 3 条记忆, 实际=%d", len(memories))
	}

	// 按类型筛选
	memories, total, err = store.ListMemories(studentID, teacherID, "concept", 1, 10)
	if err != nil {
		t.Fatalf("按类型查询失败: %v", err)
	}
	if total != 2 {
		t.Errorf("期望 total=2, 实际=%d", total)
	}

	// 分页
	memories, total, err = store.ListMemories(studentID, teacherID, "", 1, 2)
	if err != nil {
		t.Fatalf("分页查询失败: %v", err)
	}
	if total != 3 {
		t.Errorf("期望 total=3, 实际=%d", total)
	}
	if len(memories) != 2 {
		t.Errorf("期望 2 条记忆, 实际=%d", len(memories))
	}
}

func TestMemoryStore_ListMemories_InvalidIDs(t *testing.T) {
	db := setupTestDB(t)
	store := NewMemoryStore(database.NewMemoryRepository(db.DB))

	_, _, err := store.ListMemories(0, 1, "", 1, 10)
	if err == nil {
		t.Fatal("无效 studentID 应该返回错误")
	}
}

// ==================== MemoryPlugin 测试 ====================

func TestMemoryPlugin_NewAndInit(t *testing.T) {
	db := setupTestDB(t)
	plugin := NewMemoryPlugin("memory-test", db.DB)

	if plugin.Name() != "memory-test" {
		t.Errorf("期望名称=memory-test, 实际=%s", plugin.Name())
	}
	if plugin.Type() != core.PluginTypeMemory {
		t.Errorf("期望类型=memory, 实际=%s", plugin.Type())
	}

	err := plugin.Init(map[string]interface{}{
		"retention.max_memories_per_user": 100,
	})
	if err != nil {
		t.Fatalf("初始化失败: %v", err)
	}
}

func TestMemoryPlugin_StoreAction(t *testing.T) {
	plugin, _, studentID, teacherID := setupMemoryPlugin(t)

	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action":      "store",
			"student_id":  studentID,
			"teacher_id":  teacherID,
			"memory_type": "concept",
			"content":     "学生掌握了 Go 语言的基本语法",
			"importance":  0.8,
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行存储失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("存储应该成功, 错误: %s", output.Error)
	}
	if output.Data["memory_id"] == nil {
		t.Fatal("应返回 memory_id")
	}
}

func TestMemoryPlugin_StoreAction_MergeData(t *testing.T) {
	plugin, _, studentID, teacherID := setupMemoryPlugin(t)

	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action":       "store",
			"student_id":   studentID,
			"teacher_id":   teacherID,
			"memory_type":  "concept",
			"content":      "测试 merge",
			"importance":   0.5,
			"custom_field": "should_be_preserved",
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("存储应该成功: %s", output.Error)
	}

	// 验证上游 Data 被 merge
	if output.Data["custom_field"] != "should_be_preserved" {
		t.Error("上游数据 custom_field 应该被保留")
	}
}

func TestMemoryPlugin_StoreAction_MissingContent(t *testing.T) {
	plugin, _, studentID, teacherID := setupMemoryPlugin(t)

	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action":     "store",
			"student_id": studentID,
			"teacher_id": teacherID,
			"content":    "",
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if output.Success {
		t.Fatal("空内容应该失败")
	}
}

func TestMemoryPlugin_RetrieveAction(t *testing.T) {
	plugin, _, studentID, teacherID := setupMemoryPlugin(t)

	// 先存储
	storeInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":      "store",
			"student_id":  studentID,
			"teacher_id":  teacherID,
			"memory_type": "concept",
			"content":     "学生理解了递归的概念",
			"importance":  0.9,
		},
		Context: context.Background(),
	}
	_, err := plugin.Execute(context.Background(), storeInput)
	if err != nil {
		t.Fatalf("存储失败: %v", err)
	}

	// 检索
	retrieveInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":     "retrieve",
			"student_id": studentID,
			"teacher_id": teacherID,
			"limit":      10,
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), retrieveInput)
	if err != nil {
		t.Fatalf("检索失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("检索应该成功, 错误: %s", output.Error)
	}

	memories, ok := output.Data["memories"].([]map[string]interface{})
	if !ok {
		t.Fatal("应返回 memories 数组")
	}
	if len(memories) != 1 {
		t.Fatalf("期望 1 条记忆, 实际=%d", len(memories))
	}
}

func TestMemoryPlugin_RetrieveAction_MissingStudentID(t *testing.T) {
	plugin, _, _, teacherID := setupMemoryPlugin(t)

	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action":     "retrieve",
			"teacher_id": teacherID,
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if output.Success {
		t.Fatal("缺少 student_id 应该失败")
	}
}

func TestMemoryPlugin_ListAction(t *testing.T) {
	plugin, _, studentID, teacherID := setupMemoryPlugin(t)

	// 存储多条记忆
	for i := 0; i < 3; i++ {
		storeInput := &core.PluginInput{
			Data: map[string]interface{}{
				"action":      "store",
				"student_id":  studentID,
				"teacher_id":  teacherID,
				"memory_type": "concept",
				"content":     "记忆内容",
				"importance":  0.5,
			},
			Context: context.Background(),
		}
		plugin.Execute(context.Background(), storeInput)
	}

	// 列表查询
	listInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":     "list",
			"student_id": studentID,
			"teacher_id": teacherID,
			"page":       1,
			"page_size":  10,
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), listInput)
	if err != nil {
		t.Fatalf("列表查询失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("列表查询应该成功, 错误: %s", output.Error)
	}

	memories, ok := output.Data["memories"].([]map[string]interface{})
	if !ok {
		t.Fatal("应返回 memories 数组")
	}
	if len(memories) != 3 {
		t.Errorf("期望 3 条记忆, 实际=%d", len(memories))
	}

	total, ok := output.Data["total"].(int)
	if !ok || total != 3 {
		t.Errorf("期望 total=3, 实际=%v", output.Data["total"])
	}
}

func TestMemoryPlugin_ListAction_WithTypeFilter(t *testing.T) {
	plugin, _, studentID, teacherID := setupMemoryPlugin(t)

	// 存储不同类型的记忆
	types := []string{"concept", "weakness", "concept"}
	for _, mt := range types {
		storeInput := &core.PluginInput{
			Data: map[string]interface{}{
				"action":      "store",
				"student_id":  studentID,
				"teacher_id":  teacherID,
				"memory_type": mt,
				"content":     "记忆内容-" + mt,
				"importance":  0.5,
			},
			Context: context.Background(),
		}
		plugin.Execute(context.Background(), storeInput)
	}

	// 按类型筛选
	listInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":      "list",
			"student_id":  studentID,
			"teacher_id":  teacherID,
			"memory_type": "concept",
			"page":        1,
			"page_size":   10,
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), listInput)
	if err != nil {
		t.Fatalf("列表查询失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("列表查询应该成功: %s", output.Error)
	}

	total, _ := output.Data["total"].(int)
	if total != 2 {
		t.Errorf("期望 total=2, 实际=%d", total)
	}
}

func TestMemoryPlugin_PipelineMode(t *testing.T) {
	plugin, _, studentID, teacherID := setupMemoryPlugin(t)

	// 先存储一条记忆
	storeInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":      "store",
			"student_id":  studentID,
			"teacher_id":  teacherID,
			"memory_type": "concept",
			"content":     "学生已掌握基本数据结构",
			"importance":  0.7,
		},
		Context: context.Background(),
	}
	plugin.Execute(context.Background(), storeInput)

	// 管道模式（无 action）
	pipelineInput := &core.PluginInput{
		Data: map[string]interface{}{
			"teacher_id":    teacherID,
			"message":       "什么是二叉树？",
			"upstream_data": "should_be_preserved",
		},
		UserContext: &core.UserContext{
			UserID: fmt.Sprintf("%d", studentID),
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), pipelineInput)
	if err != nil {
		t.Fatalf("管道模式执行失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("管道模式应该成功, 错误: %s", output.Error)
	}

	// 验证记忆被注入
	memories, ok := output.Data["memories"].([]map[string]interface{})
	if !ok {
		t.Fatal("应注入 memories 数组")
	}
	if len(memories) != 1 {
		t.Errorf("期望 1 条记忆, 实际=%d", len(memories))
	}

	// 验证上游数据被保留
	if output.Data["upstream_data"] != "should_be_preserved" {
		t.Error("上游数据应该被保留")
	}
	if output.Data["message"] != "什么是二叉树？" {
		t.Error("上游 message 应该被保留")
	}
}

func TestMemoryPlugin_PipelineMode_NoUserContext(t *testing.T) {
	plugin, _, _, teacherID := setupMemoryPlugin(t)

	// 管道模式无 UserContext
	pipelineInput := &core.PluginInput{
		Data: map[string]interface{}{
			"teacher_id": teacherID,
			"message":    "测试",
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), pipelineInput)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("管道模式应该成功（即使没有用户信息）: %s", output.Error)
	}

	// 应返回空 memories
	memories, ok := output.Data["memories"].([]map[string]interface{})
	if !ok {
		t.Fatal("应返回 memories 数组")
	}
	if len(memories) != 0 {
		t.Errorf("无用户信息时应返回空 memories, 实际=%d", len(memories))
	}
}

func TestMemoryPlugin_InvalidAction(t *testing.T) {
	plugin, _, _, _ := setupMemoryPlugin(t)

	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action": "unknown",
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if output.Success {
		t.Fatal("未知 action 应该失败")
	}
}

func TestMemoryPlugin_HealthCheck(t *testing.T) {
	plugin, _, _, _ := setupMemoryPlugin(t)

	err := plugin.HealthCheck()
	if err != nil {
		t.Fatalf("健康检查失败: %v", err)
	}
}
