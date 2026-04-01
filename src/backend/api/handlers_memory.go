package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"digital-twin/src/backend/database"
	"digital-twin/src/plugins/dialogue"
	"digital-twin/src/plugins/memory"

	"github.com/gin-gonic/gin"
)

// ======================== V2.0 迭代6 记忆管理接口 ========================

// HandleGetMemoriesV2 记忆列表 - 支持 layer 筛选 + SQL 层分页
// GET /api/memories (改造后)
func (h *Handler) HandleGetMemoriesV2(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	// 获取 persona_id
	personaID, _ := c.Get("persona_id")
	personaIDInt64, _ := personaID.(int64)

	// 获取 role
	role, _ := c.Get("role")
	roleStr := fmt.Sprintf("%v", role)

	// 查询参数
	teacherPersonaIDStr := c.Query("teacher_persona_id")
	studentPersonaIDStr := c.Query("student_persona_id")
	layer := c.Query("layer")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	// 验证 layer 参数
	if layer != "" && layer != "core" && layer != "episodic" && layer != "archived" {
		Error(c, http.StatusBadRequest, 40004, "无效的 layer 参数，可选值: core/episodic/archived")
		return
	}

	var teacherPersonaID, studentPersonaID int64

	// 如果提供了分身ID参数，优先使用
	if teacherPersonaIDStr != "" {
		teacherPersonaID, _ = strconv.ParseInt(teacherPersonaIDStr, 10, 64)
	}
	if studentPersonaIDStr != "" {
		studentPersonaID, _ = strconv.ParseInt(studentPersonaIDStr, 10, 64)
	}

	// 根据角色自动补全分身ID
	if roleStr == "teacher" && teacherPersonaID == 0 && personaIDInt64 > 0 {
		teacherPersonaID = personaIDInt64
	}
	if roleStr == "student" && studentPersonaID == 0 && personaIDInt64 > 0 {
		studentPersonaID = personaIDInt64
	}

	// 如果有分身ID，使用SQL层分页查询
	if teacherPersonaID > 0 && studentPersonaID > 0 {
		db := h.manager.GetDB()
		if db == nil {
			Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
			return
		}

		memRepo := database.NewMemoryRepository(db)
		store := memory.NewMemoryStore(memRepo)

		memories, total, err := store.ListMemoriesWithFilter(teacherPersonaID, studentPersonaID, layer, page, pageSize)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "查询记忆列表失败: "+err.Error())
			return
		}

		// 转换为响应格式
		items := make([]map[string]interface{}, 0, len(memories))
		for _, mem := range memories {
			memLayer := mem.MemoryLayer
			if memLayer == "" {
				memLayer = "episodic"
			}
			item := map[string]interface{}{
				"id":                 mem.ID,
				"memory_type":        mem.MemoryType,
				"memory_layer":       memLayer,
				"content":            mem.Content,
				"importance":         mem.Importance,
				"teacher_persona_id": mem.TeacherPersonaID,
				"student_persona_id": mem.StudentPersonaID,
				"created_at":         mem.CreatedAt.Format(time.RFC3339),
				"updated_at":         mem.UpdatedAt.Format(time.RFC3339),
			}
			if mem.LastAccessed != nil {
				item["last_accessed"] = mem.LastAccessed.Format(time.RFC3339)
			}
			items = append(items, item)
		}

		// 使用带 pagination 的响应格式
		Success(c, gin.H{
			"items": items,
			"pagination": gin.H{
				"page":      page,
				"page_size": pageSize,
				"total":     total,
			},
		})
		return
	}

	// 向后兼容：user_id 维度查询（旧逻辑）
	teacherIDStr := c.Query("teacher_id")
	teacherID, err := strconv.ParseInt(teacherIDStr, 10, 64)
	if err != nil || teacherID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "缺少或无效的 teacher_id 参数")
		return
	}

	memoryType := c.Query("memory_type")

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	memRepo := database.NewMemoryRepository(db)
	store := memory.NewMemoryStore(memRepo)

	memories, total, err := store.ListMemories(userIDInt64, teacherID, memoryType, page, pageSize)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询记忆列表失败: "+err.Error())
		return
	}

	items := make([]map[string]interface{}, 0, len(memories))
	for _, mem := range memories {
		memLayer := mem.MemoryLayer
		if memLayer == "" {
			memLayer = "episodic"
		}
		item := map[string]interface{}{
			"id":           mem.ID,
			"memory_type":  mem.MemoryType,
			"memory_layer": memLayer,
			"content":      mem.Content,
			"importance":   mem.Importance,
			"created_at":   mem.CreatedAt.Format(time.RFC3339),
			"updated_at":   mem.UpdatedAt.Format(time.RFC3339),
		}
		if mem.LastAccessed != nil {
			item["last_accessed"] = mem.LastAccessed.Format(time.RFC3339)
		}
		items = append(items, item)
	}

	SuccessPage(c, items, total, page, pageSize)
}

// HandleUpdateMemory 编辑记忆
// PUT /api/memories/:id
func (h *Handler) HandleUpdateMemory(c *gin.Context) {
	userID, _ := c.Get("user_id")
	teacherID, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	personaID, _ := c.Get("persona_id")
	teacherPersonaID, _ := personaID.(int64)

	// 获取记忆ID
	idStr := c.Param("id")
	memoryID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || memoryID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的记忆 ID")
		return
	}

	// 解析请求体
	var req struct {
		Content     string   `json:"content" binding:"required"`
		Importance  *float64 `json:"importance"`
		MemoryLayer *string  `json:"memory_layer"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	// 校验 memory_layer
	if req.MemoryLayer != nil {
		ml := *req.MemoryLayer
		if ml != "core" && ml != "episodic" {
			Error(c, http.StatusBadRequest, 40004, "memory_layer 必须为 core 或 episodic")
			return
		}
	}

	// 校验 importance
	if req.Importance != nil {
		if *req.Importance < 0.0 || *req.Importance > 1.0 {
			Error(c, http.StatusBadRequest, 40004, "importance 必须在 0.0-1.0 之间")
			return
		}
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	memRepo := database.NewMemoryRepository(db)

	// 查询记忆是否存在
	mem, err := memRepo.GetByID(memoryID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询记忆失败: "+err.Error())
		return
	}
	if mem == nil {
		Error(c, http.StatusNotFound, 40004, "记忆不存在")
		return
	}

	// 权限校验：必须是该学生的教师
	if !h.isTeacherOfMemory(mem, teacherID, teacherPersonaID) {
		Error(c, http.StatusForbidden, 40039, "无权操作该记忆")
		return
	}

	// 更新记忆
	if err := memRepo.UpdateContent(memoryID, req.Content, req.Importance, req.MemoryLayer); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "更新记忆失败: "+err.Error())
		return
	}

	// 返回更新后的记忆
	updatedMem, _ := memRepo.GetByID(memoryID)
	if updatedMem == nil {
		updatedMem = mem
		updatedMem.Content = req.Content
	}

	memLayer := updatedMem.MemoryLayer
	if memLayer == "" {
		memLayer = "episodic"
	}

	Success(c, gin.H{
		"id":           updatedMem.ID,
		"memory_type":  updatedMem.MemoryType,
		"memory_layer": memLayer,
		"content":      updatedMem.Content,
		"importance":   updatedMem.Importance,
		"updated_at":   updatedMem.UpdatedAt.Format(time.RFC3339),
	})
}

// HandleDeleteMemory 删除记忆
// DELETE /api/memories/:id
func (h *Handler) HandleDeleteMemory(c *gin.Context) {
	userID, _ := c.Get("user_id")
	teacherID, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	personaID, _ := c.Get("persona_id")
	teacherPersonaID, _ := personaID.(int64)

	// 获取记忆ID
	idStr := c.Param("id")
	memoryID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || memoryID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的记忆 ID")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	memRepo := database.NewMemoryRepository(db)

	// 查询记忆是否存在
	mem, err := memRepo.GetByID(memoryID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询记忆失败: "+err.Error())
		return
	}
	if mem == nil {
		Error(c, http.StatusNotFound, 40004, "记忆不存在")
		return
	}

	// 权限校验
	if !h.isTeacherOfMemory(mem, teacherID, teacherPersonaID) {
		Error(c, http.StatusForbidden, 40039, "无权操作该记忆")
		return
	}

	// 删除记忆
	if err := memRepo.DeleteByID(memoryID); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "删除记忆失败: "+err.Error())
		return
	}

	Success(c, nil)
}

// HandleSummarizeMemories 记忆摘要合并
// POST /api/memories/summarize
func (h *Handler) HandleSummarizeMemories(c *gin.Context) {
	userID, _ := c.Get("user_id")
	teacherID, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	personaID, _ := c.Get("persona_id")
	teacherPersonaID, _ := personaID.(int64)

	// 解析请求体
	var req struct {
		TeacherPersonaID int64 `json:"teacher_persona_id" binding:"required"`
		StudentPersonaID int64 `json:"student_persona_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	// 权限校验：请求的 teacher_persona_id 必须是当前教师的分身
	if teacherPersonaID > 0 && req.TeacherPersonaID != teacherPersonaID {
		Error(c, http.StatusForbidden, 40039, "无权操作该记忆")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	// 校验师生关系
	relationRepo := database.NewRelationRepository(db)
	approved, err := relationRepo.IsApprovedByPersonas(req.TeacherPersonaID, req.StudentPersonaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询授权关系失败: "+err.Error())
		return
	}
	if !approved {
		Error(c, http.StatusForbidden, 40039, "无权操作该记忆（非该学生的教师）")
		return
	}

	memRepo := database.NewMemoryRepository(db)

	// 查询 episodic 记忆
	episodicMemories, err := memRepo.ListEpisodicForSummarize(req.TeacherPersonaID, req.StudentPersonaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询情景记忆失败: "+err.Error())
		return
	}

	if len(episodicMemories) == 0 {
		Error(c, http.StatusBadRequest, 40038, "没有可合并的情景记忆")
		return
	}

	// 调用 LLM 进行摘要合并
	summaryResults, err := h.summarizeMemoriesWithLLM(episodicMemories, teacherID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "记忆摘要合并失败: "+err.Error())
		return
	}

	// 将摘要结果写入数据库
	store := memory.NewMemoryStore(memRepo)
	var newCoreMemories []map[string]interface{}

	for _, summary := range summaryResults {
		// 使用 StoreMemoryWithLayer（core 层级自动更新覆盖）
		memID, err := store.StoreMemoryWithLayer(
			episodicMemories[0].StudentID,
			episodicMemories[0].TeacherID,
			req.TeacherPersonaID,
			req.StudentPersonaID,
			summary.MemoryType,
			"core",
			summary.Content,
			summary.Importance,
		)
		if err != nil {
			continue // 跳过单条失败
		}

		newCoreMemories = append(newCoreMemories, map[string]interface{}{
			"id":           memID,
			"memory_type":  summary.MemoryType,
			"memory_layer": "core",
			"content":      summary.Content,
			"importance":   summary.Importance,
		})
	}

	// 将原始 episodic 记忆标记为 archived
	archivedIDs := make([]int64, 0, len(episodicMemories))
	for _, mem := range episodicMemories {
		archivedIDs = append(archivedIDs, mem.ID)
	}
	_ = memRepo.UpdateMemoryLayer(archivedIDs, "archived")

	Success(c, gin.H{
		"summarized_count":  len(episodicMemories),
		"new_core_memories": newCoreMemories,
		"archived_count":    len(archivedIDs),
	})
}

// ======================== 辅助方法 ========================

// isTeacherOfMemory 检查教师是否有权操作该记忆
func (h *Handler) isTeacherOfMemory(mem *database.Memory, teacherID, teacherPersonaID int64) bool {
	// 分身维度校验
	if teacherPersonaID > 0 && mem.TeacherPersonaID > 0 {
		return mem.TeacherPersonaID == teacherPersonaID
	}
	// user_id 维度校验
	return mem.TeacherID == teacherID
}

// memorySummaryResult LLM 摘要结果
type memorySummaryResult struct {
	MemoryType string  `json:"memory_type"`
	Content    string  `json:"content"`
	Importance float64 `json:"importance"`
}

// summarizeMemoriesWithLLM 调用 LLM 对 episodic 记忆进行摘要合并
func (h *Handler) summarizeMemoriesWithLLM(memories []*database.Memory, teacherID int64) ([]memorySummaryResult, error) {
	// 构建 LLM prompt
	var memoryTexts string
	for i, mem := range memories {
		memoryTexts += fmt.Sprintf("%d. [%s] %s (重要性: %.1f)\n", i+1, mem.MemoryType, mem.Content, mem.Importance)
	}

	prompt := fmt.Sprintf(`你是一个教育AI助手。请将以下学生的多条情景记忆合并为1~3条核心记忆摘要。

要求：
1. 每条核心记忆应该概括某一方面的学习情况
2. 保留最重要的信息，去除重复和冗余
3. 每条记忆需要指定 memory_type（如 learning_progress, preference, ability, interaction 等）
4. 每条记忆需要评估 importance（0.0~1.0）

以下是需要合并的情景记忆：
%s

请以 JSON 数组格式返回，每个元素包含 memory_type, content, importance 三个字段。
只返回 JSON 数组，不要有其他内容。`, memoryTexts)

	// 获取 dialogue 插件的 LLM 客户端
	plugin, err := h.manager.GetPlugin("dialogue-management")
	if err != nil {
		// 如果获取不到插件，使用简单的文本合并作为降级方案
		return h.simpleMergeMemories(memories), nil
	}

	dialoguePlugin, ok := plugin.(*dialogue.DialoguePlugin)
	if !ok {
		return h.simpleMergeMemories(memories), nil
	}

	llmClient := dialoguePlugin.GetLLMClient()
	if llmClient == nil {
		return h.simpleMergeMemories(memories), nil
	}

	messages := []dialogue.ChatMessage{
		{Role: "user", Content: prompt},
	}

	resp, err := llmClient.Chat(messages)
	if err != nil {
		// LLM 调用失败，使用降级方案
		return h.simpleMergeMemories(memories), nil
	}

	// 解析 LLM 响应
	var results []memorySummaryResult
	if err := json.Unmarshal([]byte(resp.Content), &results); err != nil {
		// 尝试提取 JSON
		jsonStr := extractJSONArray(resp.Content)
		if jsonStr != "" {
			if err2 := json.Unmarshal([]byte(jsonStr), &results); err2 == nil {
				return results, nil
			}
		}
		// JSON 解析失败，使用降级方案
		return h.simpleMergeMemories(memories), nil
	}

	return results, nil
}

// simpleMergeMemories 简单合并记忆（LLM 不可用时的降级方案）
func (h *Handler) simpleMergeMemories(memories []*database.Memory) []memorySummaryResult {
	// 按 memory_type 分组
	grouped := make(map[string][]*database.Memory)
	for _, mem := range memories {
		mt := mem.MemoryType
		if mt == "" {
			mt = "general"
		}
		grouped[mt] = append(grouped[mt], mem)
	}

	var results []memorySummaryResult
	for memType, mems := range grouped {
		// 取每组中重要性最高的记忆作为摘要
		var bestContent string
		var bestImportance float64
		for _, mem := range mems {
			if mem.Importance > bestImportance {
				bestImportance = mem.Importance
				bestContent = mem.Content
			}
		}
		if bestContent != "" {
			results = append(results, memorySummaryResult{
				MemoryType: memType,
				Content:    bestContent,
				Importance: bestImportance,
			})
		}
	}

	return results
}

// extractJSONArray 从文本中提取 JSON 数组
func extractJSONArray(text string) string {
	start := -1
	end := -1
	for i, r := range text {
		if r == '[' && start == -1 {
			start = i
		}
		if r == ']' {
			end = i
		}
	}
	if start >= 0 && end > start {
		return text[start : end+1]
	}
	return ""
}

// ======================== 定时任务 ========================

// StartMemorySummarizeScheduler 启动记忆摘要定时任务
// 每天凌晨 2:00 自动扫描所有学生的 episodic 记忆，超过 50 条的自动触发摘要合并
func StartMemorySummarizeScheduler(h *Handler) {
	go func() {
		for {
			// 计算到下一个凌晨 2:00 的时间
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 2, 0, 0, 0, now.Location())
			if now.Hour() < 2 {
				next = time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
			}
			duration := next.Sub(now)

			timer := time.NewTimer(duration)
			<-timer.C

			// 执行摘要任务
			h.runMemorySummarizeTask()
		}
	}()

	fmt.Println("[定时任务] 记忆摘要合并调度器已启动，每天凌晨 2:00 执行")
}

// runMemorySummarizeTask 执行记忆摘要定时任务
func (h *Handler) runMemorySummarizeTask() {
	db := h.manager.GetDB()
	if db == nil {
		fmt.Println("[定时任务] 记忆摘要: 数据库不可用")
		return
	}

	memRepo := database.NewMemoryRepository(db)

	// 获取所有有 episodic 记忆的学生分身对
	pairs, err := memRepo.ListAllStudentPersonaPairs()
	if err != nil {
		fmt.Printf("[定时任务] 记忆摘要: 查询学生分身对失败: %v\n", err)
		return
	}

	summarizedCount := 0
	for _, pair := range pairs {
		teacherPersonaID := pair[0]
		studentPersonaID := pair[1]

		// 统计 episodic 记忆数量
		count, err := memRepo.CountEpisodicMemories(teacherPersonaID, studentPersonaID)
		if err != nil {
			continue
		}

		// 超过 50 条才触发合并
		if count <= 50 {
			continue
		}

		// 获取 episodic 记忆
		episodicMemories, err := memRepo.ListEpisodicForSummarize(teacherPersonaID, studentPersonaID)
		if err != nil || len(episodicMemories) == 0 {
			continue
		}

		// 调用 LLM 摘要
		summaryResults, err := h.summarizeMemoriesWithLLM(episodicMemories, episodicMemories[0].TeacherID)
		if err != nil {
			continue
		}

		// 写入 core 记忆
		store := memory.NewMemoryStore(memRepo)
		for _, summary := range summaryResults {
			_, _ = store.StoreMemoryWithLayer(
				episodicMemories[0].StudentID,
				episodicMemories[0].TeacherID,
				teacherPersonaID,
				studentPersonaID,
				summary.MemoryType,
				"core",
				summary.Content,
				summary.Importance,
			)
		}

		// 标记 archived
		archivedIDs := make([]int64, 0, len(episodicMemories))
		for _, mem := range episodicMemories {
			archivedIDs = append(archivedIDs, mem.ID)
		}
		_ = memRepo.UpdateMemoryLayer(archivedIDs, "archived")

		summarizedCount++
	}

	if summarizedCount > 0 {
		fmt.Printf("[定时任务] 记忆摘要: 完成 %d 个学生的记忆合并\n", summarizedCount)
	}
}

// GetLLMConfigFromEnv 从环境变量获取 LLM 配置（供定时任务使用）
func GetLLMConfigFromEnv() (baseURL, apiKey, model string) {
	baseURL = os.Getenv("OPENAI_BASE_URL")
	apiKey = os.Getenv("OPENAI_API_KEY")
	model = os.Getenv("LLM_MODEL")
	if model == "" {
		model = "qwen-turbo"
	}
	return
}
