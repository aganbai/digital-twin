package memory

import (
	"fmt"
	"time"

	"digital-twin/src/backend/database"
)

// MemoryStore 记忆存储辅助，封装 MemoryRepository 的高级操作
type MemoryStore struct {
	repo *database.MemoryRepository
}

// NewMemoryStore 创建记忆存储
func NewMemoryStore(repo *database.MemoryRepository) *MemoryStore {
	return &MemoryStore{repo: repo}
}

// RetrieveRelevant 检索相关记忆
// 按重要性和创建时间排序，返回最相关的记忆
func (s *MemoryStore) RetrieveRelevant(studentID, teacherID int64, limit int) ([]*database.Memory, error) {
	if studentID <= 0 || teacherID <= 0 {
		return nil, fmt.Errorf("无效的 studentID(%d) 或 teacherID(%d)", studentID, teacherID)
	}
	if limit <= 0 {
		limit = 10
	}

	memories, err := s.repo.GetByStudentAndTeacher(studentID, teacherID, limit)
	if err != nil {
		return nil, fmt.Errorf("检索记忆失败: %w", err)
	}

	// 更新已访问记忆的最后访问时间
	for _, mem := range memories {
		_ = s.repo.UpdateLastAccessed(mem.ID)
	}

	return memories, nil
}

// StoreMemory 存储记忆
func (s *MemoryStore) StoreMemory(studentID, teacherID int64, memoryType, content string, importance float64) (int64, error) {
	if studentID <= 0 || teacherID <= 0 {
		return 0, fmt.Errorf("无效的 studentID(%d) 或 teacherID(%d)", studentID, teacherID)
	}
	if content == "" {
		return 0, fmt.Errorf("记忆内容不能为空")
	}
	if memoryType == "" {
		memoryType = "general"
	}
	if importance <= 0 {
		importance = 0.5
	}

	now := time.Now()
	mem := &database.Memory{
		StudentID:    studentID,
		TeacherID:    teacherID,
		MemoryType:   memoryType,
		Content:      content,
		Importance:   importance,
		LastAccessed: &now,
	}

	id, err := s.repo.Create(mem)
	if err != nil {
		return 0, fmt.Errorf("存储记忆失败: %w", err)
	}

	return id, nil
}

// ListMemories 列表查询（支持按类型筛选和分页）
func (s *MemoryStore) ListMemories(studentID, teacherID int64, memoryType string, page, pageSize int) ([]*database.Memory, int, error) {
	if studentID <= 0 || teacherID <= 0 {
		return nil, 0, fmt.Errorf("无效的 studentID(%d) 或 teacherID(%d)", studentID, teacherID)
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 获取所有记忆（使用较大的 limit 来获取全部数据）
	allMemories, err := s.repo.GetByStudentAndTeacher(studentID, teacherID, 1000)
	if err != nil {
		return nil, 0, fmt.Errorf("查询记忆列表失败: %w", err)
	}

	// 按类型筛选
	var filtered []*database.Memory
	for _, mem := range allMemories {
		if memoryType == "" || mem.MemoryType == memoryType {
			filtered = append(filtered, mem)
		}
	}

	total := len(filtered)

	// 分页
	start := (page - 1) * pageSize
	if start >= total {
		return []*database.Memory{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}
