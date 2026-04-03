package database

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestMemoryLayer(t *testing.T) {
	os.Remove("test_memory.db")
	db, err := NewDatabase("test_memory.db")
	if err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}
	defer os.Remove("test_memory.db")

	repo := NewMemoryRepository(db.DB)
	sID, tID := createTestUsers(t, db)

	mem := &Memory{
		StudentID:        sID,
		TeacherID:        tID,
		TeacherPersonaID: 10,
		StudentPersonaID: 20,
		MemoryType:       "concept",
		MemoryLayer:      "core",
		Content:          "Test Core Memory",
		Importance:       0.8,
	}

	id, err := repo.CreateWithPersonas(mem)
	if err != nil {
		t.Fatalf("Failed to create memory: %v", err)
	}

	savedMem, err := repo.GetByID(id)
	if err != nil {
		t.Fatalf("Failed to get memory: %v", err)
	}

	if savedMem.MemoryLayer != "core" {
		t.Errorf("Expected layer 'core', got '%s'", savedMem.MemoryLayer)
	}

	err = repo.UpdateMemoryLayer([]int64{id}, "archived")
	if err != nil {
		t.Fatalf("Failed to update memory layer: %v", err)
	}

	savedMem, _ = repo.GetByID(id)
	if savedMem.MemoryLayer != "archived" {
		t.Errorf("Expected layer 'archived', got '%s'", savedMem.MemoryLayer)
	}
}

// ==================== V2.0 迭代6 BE-M2 新增测试 ====================

// TestGetCoreMemoryByType 测试查询同类型的最新 core 记忆
func TestGetCoreMemoryByType(t *testing.T) {
	os.Remove("test_core_memory.db")
	db, err := NewDatabase("test_core_memory.db")
	if err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}
	defer os.Remove("test_core_memory.db")

	repo := NewMemoryRepository(db.DB)
	sID, tID := createTestUsers(t, db)

	teacherPersonaID := int64(100)
	studentPersonaID := int64(200)

	// 查询不存在的 core 记忆，应返回 nil
	mem, err := repo.GetCoreMemory(teacherPersonaID, studentPersonaID, "preference")
	if err != nil {
		t.Fatalf("查询不存在的 core 记忆不应报错: %v", err)
	}
	if mem != nil {
		t.Fatal("不存在的 core 记忆应返回 nil")
	}

	// 创建一条 core 记忆
	now := time.Now()
	coreMem := &Memory{
		StudentID:        sID,
		TeacherID:        tID,
		TeacherPersonaID: teacherPersonaID,
		StudentPersonaID: studentPersonaID,
		MemoryType:       "preference",
		MemoryLayer:      "core",
		Content:          "学生喜欢数学",
		Importance:       0.8,
		LastAccessed:     &now,
	}
	coreID, err := repo.CreateWithPersonas(coreMem)
	if err != nil {
		t.Fatalf("创建 core 记忆失败: %v", err)
	}

	// 查询应返回该条记忆
	found, err := repo.GetCoreMemory(teacherPersonaID, studentPersonaID, "preference")
	if err != nil {
		t.Fatalf("查询 core 记忆失败: %v", err)
	}
	if found == nil {
		t.Fatal("应找到 core 记忆")
	}
	if found.ID != coreID {
		t.Fatalf("期望 ID=%d, 实际=%d", coreID, found.ID)
	}
	if found.Content != "学生喜欢数学" {
		t.Fatalf("内容不匹配: %s", found.Content)
	}

	// 查询不同 memory_type 应返回 nil
	notFound, err := repo.GetCoreMemory(teacherPersonaID, studentPersonaID, "ability")
	if err != nil {
		t.Fatalf("查询不同类型不应报错: %v", err)
	}
	if notFound != nil {
		t.Fatal("不同 memory_type 应返回 nil")
	}

	// 创建一条 episodic 记忆，不应被 GetCoreMemory 返回
	episodicMem := &Memory{
		StudentID:        sID,
		TeacherID:        tID,
		TeacherPersonaID: teacherPersonaID,
		StudentPersonaID: studentPersonaID,
		MemoryType:       "preference",
		MemoryLayer:      "episodic",
		Content:          "某次提到喜欢物理",
		Importance:       0.5,
		LastAccessed:     &now,
	}
	repo.CreateWithPersonas(episodicMem)

	// 仍然只返回 core 记忆
	foundAgain, _ := repo.GetCoreMemory(teacherPersonaID, studentPersonaID, "preference")
	if foundAgain == nil || foundAgain.ID != coreID {
		t.Fatal("应仍然返回 core 记忆，不受 episodic 影响")
	}
}

// TestUpdateMemory 测试更新记忆内容（UpdateCoreMemory）
func TestUpdateMemory(t *testing.T) {
	os.Remove("test_update_memory.db")
	db, err := NewDatabase("test_update_memory.db")
	if err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}
	defer os.Remove("test_update_memory.db")

	repo := NewMemoryRepository(db.DB)
	sID, tID := createTestUsers(t, db)

	now := time.Now()
	mem := &Memory{
		StudentID:        sID,
		TeacherID:        tID,
		TeacherPersonaID: 100,
		StudentPersonaID: 200,
		MemoryType:       "preference",
		MemoryLayer:      "core",
		Content:          "原始内容",
		Importance:       0.5,
		LastAccessed:     &now,
	}
	id, err := repo.CreateWithPersonas(mem)
	if err != nil {
		t.Fatalf("创建记忆失败: %v", err)
	}

	// 更新内容和重要性
	err = repo.UpdateCoreMemory(id, "更新后的内容", 0.9)
	if err != nil {
		t.Fatalf("更新记忆失败: %v", err)
	}

	// 验证更新结果
	updated, err := repo.GetByID(id)
	if err != nil {
		t.Fatalf("查询更新后的记忆失败: %v", err)
	}
	if updated.Content != "更新后的内容" {
		t.Fatalf("内容未更新, 期望='更新后的内容', 实际='%s'", updated.Content)
	}
	if updated.Importance != 0.9 {
		t.Fatalf("重要性未更新, 期望=0.9, 实际=%f", updated.Importance)
	}

	// 验证 updated_at 已更新
	if !updated.UpdatedAt.After(updated.CreatedAt.Add(-time.Second)) {
		t.Fatal("updated_at 应该被更新")
	}
}

// TestCountEpisodicMemories 测试统计 episodic 记忆数量
func TestCountEpisodicMemories(t *testing.T) {
	os.Remove("test_count_episodic.db")
	db, err := NewDatabase("test_count_episodic.db")
	if err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}
	defer os.Remove("test_count_episodic.db")

	repo := NewMemoryRepository(db.DB)
	sID, tID := createTestUsers(t, db)

	teacherPersonaID := int64(100)
	studentPersonaID := int64(200)

	// 初始应为 0
	count, err := repo.CountEpisodicMemories(teacherPersonaID, studentPersonaID)
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}
	if count != 0 {
		t.Fatalf("初始 episodic 数量应为 0, 实际=%d", count)
	}

	now := time.Now()

	// 创建 3 条 episodic 记忆
	for i := 0; i < 3; i++ {
		repo.CreateWithPersonas(&Memory{
			StudentID:        sID,
			TeacherID:        tID,
			TeacherPersonaID: teacherPersonaID,
			StudentPersonaID: studentPersonaID,
			MemoryType:       "interaction",
			MemoryLayer:      "episodic",
			Content:          "情景记忆",
			Importance:       0.5,
			LastAccessed:     &now,
		})
	}

	// 创建 1 条 core 记忆（不应计入）
	repo.CreateWithPersonas(&Memory{
		StudentID:        sID,
		TeacherID:        tID,
		TeacherPersonaID: teacherPersonaID,
		StudentPersonaID: studentPersonaID,
		MemoryType:       "preference",
		MemoryLayer:      "core",
		Content:          "核心记忆",
		Importance:       0.9,
		LastAccessed:     &now,
	})

	// 创建 1 条 archived 记忆（不应计入）
	repo.CreateWithPersonas(&Memory{
		StudentID:        sID,
		TeacherID:        tID,
		TeacherPersonaID: teacherPersonaID,
		StudentPersonaID: studentPersonaID,
		MemoryType:       "interaction",
		MemoryLayer:      "archived",
		Content:          "已归档记忆",
		Importance:       0.3,
		LastAccessed:     &now,
	})

	count, err = repo.CountEpisodicMemories(teacherPersonaID, studentPersonaID)
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}
	if count != 3 {
		t.Fatalf("期望 episodic 数量=3, 实际=%d", count)
	}

	// 不同分身对的记忆不应计入
	count2, _ := repo.CountEpisodicMemories(999, 888)
	if count2 != 0 {
		t.Fatalf("不同分身对应为 0, 实际=%d", count2)
	}
}

// TestDeleteOldestEpisodicMemories 测试删除超出上限的最旧 episodic 记忆
func TestDeleteOldestEpisodicMemories(t *testing.T) {
	os.Remove("test_delete_oldest.db")
	db, err := NewDatabase("test_delete_oldest.db")
	if err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}
	defer os.Remove("test_delete_oldest.db")

	repo := NewMemoryRepository(db.DB)
	sID, tID := createTestUsers(t, db)

	teacherPersonaID := int64(100)
	studentPersonaID := int64(200)

	now := time.Now()

	// 创建 5 条 episodic 记忆（通过 time.Sleep 确保 updated_at 有区别）
	var ids []int64
	for i := 0; i < 5; i++ {
		id, err := repo.CreateWithPersonas(&Memory{
			StudentID:        sID,
			TeacherID:        tID,
			TeacherPersonaID: teacherPersonaID,
			StudentPersonaID: studentPersonaID,
			MemoryType:       "interaction",
			MemoryLayer:      "episodic",
			Content:          fmt.Sprintf("情景记忆%d", i+1),
			Importance:       0.5,
			LastAccessed:     &now,
		})
		if err != nil {
			t.Fatalf("创建记忆失败: %v", err)
		}
		ids = append(ids, id)
		time.Sleep(10 * time.Millisecond)
	}

	// 创建 1 条 core 记忆（不应被删除）
	coreID, _ := repo.CreateWithPersonas(&Memory{
		StudentID:        sID,
		TeacherID:        tID,
		TeacherPersonaID: teacherPersonaID,
		StudentPersonaID: studentPersonaID,
		MemoryType:       "preference",
		MemoryLayer:      "core",
		Content:          "核心记忆",
		Importance:       0.9,
		LastAccessed:     &now,
	})

	// 保留最新 3 条，删除最旧 2 条
	err = repo.DeleteOldestEpisodicMemories(teacherPersonaID, studentPersonaID, 3)
	if err != nil {
		t.Fatalf("删除最旧 episodic 记忆失败: %v", err)
	}

	// 验证 episodic 数量
	count, _ := repo.CountEpisodicMemories(teacherPersonaID, studentPersonaID)
	if count != 3 {
		t.Fatalf("期望剩余 3 条 episodic, 实际=%d", count)
	}

	// 验证最旧的 2 条已被删除
	for _, id := range ids[:2] {
		mem, _ := repo.GetByID(id)
		if mem != nil {
			t.Fatalf("最旧的记忆 ID=%d 应已被删除", id)
		}
	}

	// 验证最新的 3 条仍存在
	for _, id := range ids[2:] {
		mem, _ := repo.GetByID(id)
		if mem == nil {
			t.Fatalf("最新的记忆 ID=%d 不应被删除", id)
		}
	}

	// 验证 core 记忆未被删除
	coreMem, _ := repo.GetByID(coreID)
	if coreMem == nil {
		t.Fatal("core 记忆不应被删除")
	}

	// 测试 keepCount=0 的情况：删除所有 episodic
	err = repo.DeleteOldestEpisodicMemories(teacherPersonaID, studentPersonaID, 0)
	if err != nil {
		t.Fatalf("删除所有 episodic 记忆失败: %v", err)
	}
	count, _ = repo.CountEpisodicMemories(teacherPersonaID, studentPersonaID)
	if count != 0 {
		t.Fatalf("keepCount=0 后应无 episodic 记忆, 实际=%d", count)
	}

	// core 记忆仍在
	coreMem, _ = repo.GetByID(coreID)
	if coreMem == nil {
		t.Fatal("core 记忆不应被删除")
	}
}
