package memory

import (
	"testing"
	"time"

	"digital-twin/src/backend/database"
)

// v6TestData 迭代6测试辅助数据
type v6TestData struct {
	db               *database.Database
	repo             *database.MemoryRepository
	store            *MemoryStore
	studentID        int64
	teacherID        int64
	teacherPersonaID int64
	studentPersonaID int64
}

// setupV6Test 创建迭代6测试所需的完整测试数据
func setupV6Test(t *testing.T) *v6TestData {
	t.Helper()
	db := setupTestDB(t)

	// 创建用户
	studentID, teacherID := createTestUsers(t, db)

	// 创建分身
	personaRepo := database.NewPersonaRepository(db.DB)
	teacherPersonaID, err := personaRepo.Create(&database.Persona{
		UserID:   teacherID,
		Role:     "teacher",
		Nickname: "V6测试教师分身",
		IsActive: 1,
	})
	if err != nil {
		t.Fatalf("创建教师分身失败: %v", err)
	}

	studentPersonaID, err := personaRepo.Create(&database.Persona{
		UserID:   studentID,
		Role:     "student",
		Nickname: "V6测试学生分身",
		IsActive: 1,
	})
	if err != nil {
		t.Fatalf("创建学生分身失败: %v", err)
	}

	repo := database.NewMemoryRepository(db.DB)
	store := NewMemoryStore(repo)

	return &v6TestData{
		db:               db,
		repo:             repo,
		store:            store,
		studentID:        studentID,
		teacherID:        teacherID,
		teacherPersonaID: teacherPersonaID,
		studentPersonaID: studentPersonaID,
	}
}

// TestStoreMemoryWithLayer_Episodic 测试 episodic 层级存储
func TestStoreMemoryWithLayer_Episodic(t *testing.T) {
	td := setupV6Test(t)

	id, err := td.store.StoreMemoryWithLayer(td.studentID, td.teacherID, td.teacherPersonaID, td.studentPersonaID, "interaction", "episodic", "学生问了一个关于数学的问题", 0.5)
	if err != nil {
		t.Fatalf("存储 episodic 记忆失败: %v", err)
	}
	if id <= 0 {
		t.Fatal("返回的 ID 应大于 0")
	}

	// 验证存储结果
	mem, err := td.repo.GetByID(id)
	if err != nil {
		t.Fatalf("查询记忆失败: %v", err)
	}
	if mem == nil {
		t.Fatal("记忆不应为 nil")
	}
	if mem.MemoryLayer != "episodic" {
		t.Fatalf("期望 memory_layer=episodic, 实际=%s", mem.MemoryLayer)
	}
	if mem.MemoryType != "interaction" {
		t.Fatalf("期望 memory_type=interaction, 实际=%s", mem.MemoryType)
	}
}

// TestStoreMemoryWithLayer_CoreUpdateOverride 测试 core 层级更新覆盖
func TestStoreMemoryWithLayer_CoreUpdateOverride(t *testing.T) {
	td := setupV6Test(t)

	// 第一次存储 core 记忆
	id1, err := td.store.StoreMemoryWithLayer(td.studentID, td.teacherID, td.teacherPersonaID, td.studentPersonaID, "preference", "core", "学生喜欢数学", 0.8)
	if err != nil {
		t.Fatalf("第一次存储 core 记忆失败: %v", err)
	}

	// 第二次存储同类 core 记忆 → 应该 UPDATE 而非 INSERT
	id2, err := td.store.StoreMemoryWithLayer(td.studentID, td.teacherID, td.teacherPersonaID, td.studentPersonaID, "preference", "core", "学生非常喜欢数学和物理", 0.9)
	if err != nil {
		t.Fatalf("第二次存储 core 记忆失败: %v", err)
	}

	// 验证返回的是同一个 ID（UPDATE 而非 INSERT）
	if id1 != id2 {
		t.Fatalf("core 记忆应该更新覆盖，期望 ID=%d, 实际=%d", id1, id2)
	}

	// 验证内容已更新
	mem, _ := td.repo.GetByID(id1)
	if mem == nil {
		t.Fatal("记忆不应为 nil")
	}
	if mem.Content != "学生非常喜欢数学和物理" {
		t.Fatalf("core 记忆内容应已更新, 实际=%s", mem.Content)
	}
	if mem.Importance != 0.9 {
		t.Fatalf("core 记忆重要性应已更新, 期望=0.9, 实际=%f", mem.Importance)
	}
}

// TestStoreMemoryWithLayer_DifferentTypesCreateSeparateCore 不同 memory_type 创建独立 core
func TestStoreMemoryWithLayer_DifferentTypesCreateSeparateCore(t *testing.T) {
	td := setupV6Test(t)

	id1, _ := td.store.StoreMemoryWithLayer(td.studentID, td.teacherID, td.teacherPersonaID, td.studentPersonaID, "preference", "core", "喜欢数学", 0.8)
	id2, _ := td.store.StoreMemoryWithLayer(td.studentID, td.teacherID, td.teacherPersonaID, td.studentPersonaID, "ability", "core", "擅长编程", 0.9)

	// 不同 memory_type 应创建不同的记忆
	if id1 == id2 {
		t.Fatal("不同 memory_type 的 core 记忆不应覆盖")
	}
}

// TestListMemoriesWithFilter_LayerFilter 测试按层级筛选
func TestListMemoriesWithFilter_LayerFilter(t *testing.T) {
	td := setupV6Test(t)

	// 创建不同层级的记忆
	td.store.StoreMemoryWithLayer(td.studentID, td.teacherID, td.teacherPersonaID, td.studentPersonaID, "preference", "core", "核心记忆1", 0.9)
	td.store.StoreMemoryWithLayer(td.studentID, td.teacherID, td.teacherPersonaID, td.studentPersonaID, "interaction", "episodic", "情景记忆1", 0.5)
	td.store.StoreMemoryWithLayer(td.studentID, td.teacherID, td.teacherPersonaID, td.studentPersonaID, "interaction", "episodic", "情景记忆2", 0.6)

	// 筛选 core
	coreMemories, total, err := td.store.ListMemoriesWithFilter(td.teacherPersonaID, td.studentPersonaID, "core", 1, 20)
	if err != nil {
		t.Fatalf("筛选 core 记忆失败: %v", err)
	}
	if total != 1 {
		t.Fatalf("期望 core 记忆数量=1, 实际=%d", total)
	}
	if len(coreMemories) != 1 {
		t.Fatalf("期望返回 1 条 core 记忆, 实际=%d", len(coreMemories))
	}
	if coreMemories[0].MemoryLayer != "core" {
		t.Fatalf("期望 memory_layer=core, 实际=%s", coreMemories[0].MemoryLayer)
	}

	// 筛选 episodic
	episodicMemories, total, err := td.store.ListMemoriesWithFilter(td.teacherPersonaID, td.studentPersonaID, "episodic", 1, 20)
	if err != nil {
		t.Fatalf("筛选 episodic 记忆失败: %v", err)
	}
	if total != 2 {
		t.Fatalf("期望 episodic 记忆数量=2, 实际=%d", total)
	}
	if len(episodicMemories) != 2 {
		t.Fatalf("期望返回 2 条 episodic 记忆, 实际=%d", len(episodicMemories))
	}

	// 不传 layer → 返回 core + episodic
	allMemories, total, err := td.store.ListMemoriesWithFilter(td.teacherPersonaID, td.studentPersonaID, "", 1, 20)
	if err != nil {
		t.Fatalf("查询全部记忆失败: %v", err)
	}
	if total != 3 {
		t.Fatalf("期望全部记忆数量=3, 实际=%d", total)
	}
	if len(allMemories) != 3 {
		t.Fatalf("期望返回 3 条记忆, 实际=%d", len(allMemories))
	}
}

// TestListMemoriesWithFilter_Pagination SQL层分页
func TestListMemoriesWithFilter_Pagination(t *testing.T) {
	td := setupV6Test(t)

	// 创建 5 条 episodic 记忆
	for i := 0; i < 5; i++ {
		td.store.StoreMemoryWithLayer(td.studentID, td.teacherID, td.teacherPersonaID, td.studentPersonaID, "interaction", "episodic",
			"情景记忆"+string(rune('A'+i)), float64(i+1)*0.1)
	}

	// 第1页（每页2条）
	page1, total, _ := td.store.ListMemoriesWithFilter(td.teacherPersonaID, td.studentPersonaID, "episodic", 1, 2)
	if total != 5 {
		t.Fatalf("期望总数=5, 实际=%d", total)
	}
	if len(page1) != 2 {
		t.Fatalf("期望第1页返回2条, 实际=%d", len(page1))
	}

	// 第3页（每页2条，应只有1条）
	page3, _, _ := td.store.ListMemoriesWithFilter(td.teacherPersonaID, td.studentPersonaID, "episodic", 3, 2)
	if len(page3) != 1 {
		t.Fatalf("期望第3页返回1条, 实际=%d", len(page3))
	}
}

// TestMemoryRepository_V6_GetByID 测试根据 ID 查询记忆
func TestMemoryRepository_V6_GetByID(t *testing.T) {
	td := setupV6Test(t)

	now := time.Now()
	mem := &database.Memory{
		StudentID:        td.studentID,
		TeacherID:        td.teacherID,
		TeacherPersonaID: td.teacherPersonaID,
		StudentPersonaID: td.studentPersonaID,
		MemoryType:       "test",
		MemoryLayer:      "episodic",
		Content:          "测试记忆",
		Importance:       0.7,
		LastAccessed:     &now,
	}
	id, _ := td.repo.CreateWithPersonas(mem)

	// 查询
	found, err := td.repo.GetByID(id)
	if err != nil {
		t.Fatalf("查询记忆失败: %v", err)
	}
	if found == nil {
		t.Fatal("记忆不应为 nil")
	}
	if found.Content != "测试记忆" {
		t.Fatalf("内容不匹配: %s", found.Content)
	}

	// 查询不存在的 ID
	notFound, err := td.repo.GetByID(99999)
	if err != nil {
		t.Fatalf("查询不存在的记忆不应报错: %v", err)
	}
	if notFound != nil {
		t.Fatal("不存在的记忆应返回 nil")
	}
}

// TestMemoryRepository_V6_UpdateContent 测试更新记忆内容
func TestMemoryRepository_V6_UpdateContent(t *testing.T) {
	td := setupV6Test(t)

	now := time.Now()
	mem := &database.Memory{
		StudentID:        td.studentID,
		TeacherID:        td.teacherID,
		TeacherPersonaID: td.teacherPersonaID,
		StudentPersonaID: td.studentPersonaID,
		MemoryType:       "test",
		MemoryLayer:      "episodic",
		Content:          "原始内容",
		Importance:       0.5,
		LastAccessed:     &now,
	}
	id, _ := td.repo.CreateWithPersonas(mem)

	// 更新内容
	newImportance := 0.9
	newLayer := "core"
	err := td.repo.UpdateContent(id, "更新后的内容", &newImportance, &newLayer)
	if err != nil {
		t.Fatalf("更新记忆失败: %v", err)
	}

	// 验证更新
	updated, _ := td.repo.GetByID(id)
	if updated.Content != "更新后的内容" {
		t.Fatalf("内容未更新: %s", updated.Content)
	}
	if updated.Importance != 0.9 {
		t.Fatalf("重要性未更新: %f", updated.Importance)
	}
	if updated.MemoryLayer != "core" {
		t.Fatalf("层级未更新: %s", updated.MemoryLayer)
	}
}

// TestMemoryRepository_V6_DeleteByID 测试删除记忆
func TestMemoryRepository_V6_DeleteByID(t *testing.T) {
	td := setupV6Test(t)

	now := time.Now()
	mem := &database.Memory{
		StudentID:        td.studentID,
		TeacherID:        td.teacherID,
		TeacherPersonaID: td.teacherPersonaID,
		StudentPersonaID: td.studentPersonaID,
		MemoryType:       "test",
		Content:          "待删除记忆",
		Importance:       0.5,
		LastAccessed:     &now,
	}
	id, _ := td.repo.CreateWithPersonas(mem)

	// 删除
	err := td.repo.DeleteByID(id)
	if err != nil {
		t.Fatalf("删除记忆失败: %v", err)
	}

	// 验证已删除
	deleted, _ := td.repo.GetByID(id)
	if deleted != nil {
		t.Fatal("记忆应已被删除")
	}

	// 删除不存在的记忆
	err = td.repo.DeleteByID(99999)
	if err == nil {
		t.Fatal("删除不存在的记忆应报错")
	}
}

// TestMemoryRepository_V6_UpdateMemoryLayer 测试批量更新层级
func TestMemoryRepository_V6_UpdateMemoryLayer(t *testing.T) {
	td := setupV6Test(t)

	// 创建 3 条 episodic 记忆
	id1, _ := td.store.StoreMemoryWithLayer(td.studentID, td.teacherID, td.teacherPersonaID, td.studentPersonaID, "interaction", "episodic", "记忆1", 0.5)
	id2, _ := td.store.StoreMemoryWithLayer(td.studentID, td.teacherID, td.teacherPersonaID, td.studentPersonaID, "interaction", "episodic", "记忆2", 0.6)
	id3, _ := td.store.StoreMemoryWithLayer(td.studentID, td.teacherID, td.teacherPersonaID, td.studentPersonaID, "interaction", "episodic", "记忆3", 0.7)

	// 批量标记为 archived
	err := td.repo.UpdateMemoryLayer([]int64{id1, id2, id3}, "archived")
	if err != nil {
		t.Fatalf("批量更新层级失败: %v", err)
	}

	// 验证
	mem1, _ := td.repo.GetByID(id1)
	if mem1.MemoryLayer != "archived" {
		t.Fatalf("记忆1层级应为 archived, 实际=%s", mem1.MemoryLayer)
	}

	// archived 不应出现在默认查询中
	memories, total, _ := td.store.ListMemoriesWithFilter(td.teacherPersonaID, td.studentPersonaID, "", 1, 20)
	if total != 0 {
		t.Fatalf("archived 记忆不应出现在默认查询中, total=%d", total)
	}
	if len(memories) != 0 {
		t.Fatalf("archived 记忆不应出现在默认查询中, len=%d", len(memories))
	}
}

// TestMemoryRepository_V6_CountEpisodicMemories 测试统计 episodic 数量
func TestMemoryRepository_V6_CountEpisodicMemories(t *testing.T) {
	td := setupV6Test(t)

	// 创建混合层级记忆
	td.store.StoreMemoryWithLayer(td.studentID, td.teacherID, td.teacherPersonaID, td.studentPersonaID, "preference", "core", "核心记忆", 0.9)
	td.store.StoreMemoryWithLayer(td.studentID, td.teacherID, td.teacherPersonaID, td.studentPersonaID, "interaction", "episodic", "情景1", 0.5)
	td.store.StoreMemoryWithLayer(td.studentID, td.teacherID, td.teacherPersonaID, td.studentPersonaID, "interaction", "episodic", "情景2", 0.6)

	count, err := td.repo.CountEpisodicMemories(td.teacherPersonaID, td.studentPersonaID)
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}
	if count != 2 {
		t.Fatalf("期望 episodic 数量=2, 实际=%d", count)
	}
}

// TestDefaultMemoryLayer_V6 测试默认层级
func TestDefaultMemoryLayer_V6(t *testing.T) {
	td := setupV6Test(t)

	// 不指定 layer，应默认 episodic
	id, _ := td.store.StoreMemoryWithLayer(td.studentID, td.teacherID, td.teacherPersonaID, td.studentPersonaID, "test", "", "默认层级记忆", 0.5)
	mem, _ := td.repo.GetByID(id)
	if mem.MemoryLayer != "episodic" {
		t.Fatalf("默认层级应为 episodic, 实际=%s", mem.MemoryLayer)
	}
}

// TestClassifyMemoryLayer_V6 测试记忆层级分类
func TestClassifyMemoryLayer_V6(t *testing.T) {
	plugin := &MemoryPlugin{}

	tests := []struct {
		memoryType string
		expected   string
	}{
		{"preference", "core"},
		{"ability", "core"},
		{"learning_goal", "core"},
		{"interaction", "episodic"},
		{"question", "episodic"},
		{"feedback", "episodic"},
		{"unknown", "episodic"},
		{"", "episodic"},
	}

	for _, tt := range tests {
		result := plugin.classifyMemoryLayer("some content", tt.memoryType)
		if result != tt.expected {
			t.Errorf("classifyMemoryLayer(%q) = %q, 期望 %q", tt.memoryType, result, tt.expected)
		}
	}
}
