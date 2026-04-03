package memory

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"digital-twin/src/backend/database"
)

// MemorySummarizer 记忆摘要合并定时任务
type MemorySummarizer struct {
	repo       *database.MemoryRepository
	store      *MemoryStore
	db         *sql.DB // V2.0 迭代7: 用于更新用户画像
	llmBaseURL string
	llmAPIKey  string
	llmModel   string
	stopCh     chan struct{}
}

// NewMemorySummarizer 创建记忆摘要合并器
func NewMemorySummarizer(repo *database.MemoryRepository, store *MemoryStore, db *sql.DB, llmBaseURL, llmAPIKey, llmModel string) *MemorySummarizer {
	return &MemorySummarizer{
		repo:       repo,
		store:      store,
		db:         db,
		llmBaseURL: llmBaseURL,
		llmAPIKey:  llmAPIKey,
		llmModel:   llmModel,
		stopCh:     make(chan struct{}),
	}
}

// summaryCoreMemory 摘要合并后的 core 记忆结构
type summaryCoreMemory struct {
	MemoryType string  `json:"memory_type"`
	Content    string  `json:"content"`
	Importance float64 `json:"importance"`
}

// StartMemorySummarizeScheduler 启动记忆摘要合并定时任务
// 每天凌晨 2:00 执行，扫描所有学生的 episodic 记忆，超过 50 条的自动触发摘要合并
func (s *MemorySummarizer) StartMemorySummarizeScheduler() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[MemorySummarizer] goroutine panic recovered: %v", r)
			}
		}()
		log.Println("[MemorySummarizer] 定时任务已启动，每天凌晨 2:00 执行")
		for {
			now := time.Now()
			// 计算下次凌晨 2:00 的时间
			next := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
			if now.After(next) {
				// 如果当前时间已过今天 2:00，则等到明天 2:00
				next = next.Add(24 * time.Hour)
			}
			waitDuration := next.Sub(now)

			log.Printf("[MemorySummarizer] 下次执行时间: %s（等待 %v）", next.Format("2006-01-02 15:04:05"), waitDuration)

			select {
			case <-time.After(waitDuration):
				s.runSummarize()
			case <-s.stopCh:
				log.Println("[MemorySummarizer] 定时任务已停止")
				return
			}
		}
	}()
}

// Stop 停止定时任务
func (s *MemorySummarizer) Stop() {
	close(s.stopCh)
}

// runSummarize 执行一次摘要合并
func (s *MemorySummarizer) runSummarize() {
	log.Println("[MemorySummarizer] 开始执行摘要合并任务...")

	// 1. 获取所有有 episodic 记忆的学生分身对
	pairs, err := s.repo.ListAllStudentPersonaPairs()
	if err != nil {
		log.Printf("[MemorySummarizer] 获取学生分身对失败: %v", err)
		return
	}

	summarizedCount := 0
	for _, pair := range pairs {
		teacherPersonaID, studentPersonaID := pair[0], pair[1]

		// 2. 统计 episodic 记忆数量
		count, err := s.repo.CountEpisodicMemories(teacherPersonaID, studentPersonaID)
		if err != nil {
			log.Printf("[MemorySummarizer] 统计 episodic 数量失败 (teacher=%d, student=%d): %v",
				teacherPersonaID, studentPersonaID, err)
			continue
		}

		if count <= 50 {
			continue
		}

		log.Printf("[MemorySummarizer] 学生分身对 (teacher=%d, student=%d) 有 %d 条 episodic 记忆，触发摘要合并",
			teacherPersonaID, studentPersonaID, count)

		// 3. 获取所有 episodic 记忆
		memories, err := s.repo.ListEpisodicForSummarize(teacherPersonaID, studentPersonaID)
		if err != nil {
			log.Printf("[MemorySummarizer] 获取 episodic 记忆失败: %v", err)
			continue
		}

		// 4. 调用 LLM 压缩为 core 记忆
		coreMemories, err := s.callLLMForSummarize(memories)
		if err != nil {
			log.Printf("[MemorySummarizer] LLM 摘要合并失败: %v", err)
			continue
		}

		// 5. 存储新的 core 记忆（通过 store 的 core 更新覆盖逻辑）
		// 需要获取 studentID 和 teacherID
		if len(memories) == 0 {
			continue
		}
		studentID := memories[0].StudentID
		teacherID := memories[0].TeacherID

		// 收集写入失败的 episodic ID，排除后再标记 archived
		failedEpisodicIDs := make(map[int64]bool)
		allCoreFailed := true
		for _, cm := range coreMemories {
			_, err := s.store.StoreMemoryWithLayer(studentID, teacherID, teacherPersonaID, studentPersonaID,
				cm.MemoryType, "core", cm.Content, cm.Importance)
			if err != nil {
				log.Printf("[MemorySummarizer] 存储 core 记忆失败: %v", err)
				// 将对应 memory_type 的 episodic 记忆 ID 标记为失败
				for _, m := range memories {
					if m.MemoryType == cm.MemoryType {
						failedEpisodicIDs[m.ID] = true
					}
				}
			} else {
				allCoreFailed = false
			}
		}

		// 如果所有 core 记忆写入都失败，跳过 archived 标记，避免数据丢失
		if allCoreFailed && len(coreMemories) > 0 {
			log.Printf("[MemorySummarizer] 所有 core 记忆写入失败，跳过 archived 标记 (teacher=%d, student=%d)",
				teacherPersonaID, studentPersonaID)
			continue
		}

		// 5.5 提炼用户画像并持久化（R10）—— 同步执行，避免 SQLite 并发写入冲突
		if studentID > 0 && s.db != nil {
			snapshot := s.extractProfileSnapshot(memories)
			if snapshot != "" {
				userRepo := database.NewUserRepository(s.db)
				if err := userRepo.UpdateProfileSnapshot(studentID, snapshot); err != nil {
					log.Printf("[MemorySummarizer] 更新用户画像失败: %v", err)
				} else {
					log.Printf("[MemorySummarizer] 用户画像已更新 (user_id=%d)", studentID)
				}
			}
		}

		// 6. 将原始 episodic 记忆标记为 archived（排除写入失败关联的 ID）
		ids := make([]int64, 0, len(memories))
		for _, m := range memories {
			if !failedEpisodicIDs[m.ID] {
				ids = append(ids, m.ID)
			}
		}
		if len(ids) == 0 {
			continue
		}
		if err := s.repo.UpdateMemoryLayer(ids, "archived"); err != nil {
			log.Printf("[MemorySummarizer] 标记 archived 失败: %v", err)
			continue
		}

		summarizedCount++
		log.Printf("[MemorySummarizer] 学生分身对 (teacher=%d, student=%d) 摘要合并完成: %d 条 episodic → %d 条 core",
			teacherPersonaID, studentPersonaID, len(memories), len(coreMemories))
	}

	log.Printf("[MemorySummarizer] 摘要合并任务完成，共处理 %d 个学生分身对", summarizedCount)
}

// callLLMForSummarize 调用 LLM 将多条 episodic 记忆压缩为 core 记忆
func (s *MemorySummarizer) callLLMForSummarize(memories []*database.Memory) ([]summaryCoreMemory, error) {
	if s.llmBaseURL == "" || s.llmAPIKey == "" {
		// 无 LLM 配置时，使用规则摘要
		return s.ruleSummarize(memories), nil
	}

	// 构建 episodic 记忆列表文本
	var memoryTexts []string
	for _, m := range memories {
		memoryTexts = append(memoryTexts, fmt.Sprintf("- [%s] %s (重要性: %.1f)", m.MemoryType, m.Content, m.Importance))
	}
	episodicList := strings.Join(memoryTexts, "\n")

	prompt := fmt.Sprintf(`你是一个学生画像分析专家。请将以下多条学生情景记忆合并为精炼的核心画像记忆。

学生的情景记忆列表：
%s

请将这些记忆合并为 1~3 条核心画像记忆，每条记忆应该是一个关于学生的持久性结论。
输出格式为 JSON 数组：
[
  {"memory_type": "learning_progress", "content": "...", "importance": 0.9},
  {"memory_type": "personality", "content": "...", "importance": 0.8}
]

memory_type 可选值：learning_progress（学习进度）、personality（性格特点）、learning_style（学习风格）、knowledge_gap（知识薄弱点）、strength（优势领域）`, episodicList)

	url := strings.TrimRight(s.llmBaseURL, "/") + "/chat/completions"
	model := s.llmModel
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]interface{}{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.3,
	}

	jsonData, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建 LLM 请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.llmAPIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("调用 LLM 失败: %w", err)
	}
	defer resp.Body.Close()

	var respData struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return nil, fmt.Errorf("解析 LLM 响应失败: %w", err)
	}

	if len(respData.Choices) == 0 {
		return nil, fmt.Errorf("LLM 返回空结果")
	}

	// 解析 JSON 数组
	content := strings.TrimSpace(respData.Choices[0].Message.Content)
	// 尝试提取 JSON 部分（LLM 可能返回额外文本）
	startIdx := strings.Index(content, "[")
	endIdx := strings.LastIndex(content, "]")
	if startIdx >= 0 && endIdx > startIdx {
		content = content[startIdx : endIdx+1]
	}

	var result []summaryCoreMemory
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("解析摘要结果失败: %w", err)
	}

	// 限制最多 3 条
	if len(result) > 3 {
		result = result[:3]
	}

	return result, nil
}

// ruleSummarize 基于规则的简单摘要（LLM 不可用时的回退方案）
func (s *MemorySummarizer) ruleSummarize(memories []*database.Memory) []summaryCoreMemory {
	// 按 memory_type 分组，取每组中重要性最高的记忆
	typeMap := make(map[string]*database.Memory)
	for _, m := range memories {
		existing, ok := typeMap[m.MemoryType]
		if !ok || m.Importance > existing.Importance {
			typeMap[m.MemoryType] = m
		}
	}

	var result []summaryCoreMemory
	for _, m := range typeMap {
		result = append(result, summaryCoreMemory{
			MemoryType: m.MemoryType,
			Content:    m.Content,
			Importance: m.Importance,
		})
		if len(result) >= 3 {
			break
		}
	}

	return result
}

// extractProfileSnapshot 从记忆中提炼用户画像
func (s *MemorySummarizer) extractProfileSnapshot(memories []*database.Memory) string {
	if s.llmBaseURL == "" || s.llmAPIKey == "" {
		return s.ruleExtractProfile(memories)
	}

	var memoryTexts []string
	for _, m := range memories {
		memoryTexts = append(memoryTexts, fmt.Sprintf("- [%s] %s", m.MemoryType, m.Content))
	}
	memoryList := strings.Join(memoryTexts, "\n")

	prompt := fmt.Sprintf(`你是一个学生画像分析专家。请从以下学生记忆中提炼出一份结构化的用户画像。

学生记忆列表：
%s

请输出JSON格式的用户画像，包含以下字段（如果信息不足则留空字符串）：
{
  "learning_style": "学习风格偏好（如视觉型/听觉型/动手型）",
  "knowledge_gaps": "知识薄弱点",
  "strengths": "优势领域",
  "interests": "兴趣领域",
  "personality": "性格特点",
  "learning_habits": "学习习惯",
  "summary": "一句话总结"
}

只返回JSON，不要其他内容。`, memoryList)

	url := strings.TrimRight(s.llmBaseURL, "/") + "/chat/completions"
	model := s.llmModel
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]interface{}{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.3,
	}

	jsonData, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return s.ruleExtractProfile(memories)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.llmAPIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return s.ruleExtractProfile(memories)
	}
	defer resp.Body.Close()

	var respData struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return s.ruleExtractProfile(memories)
	}

	if len(respData.Choices) == 0 {
		return s.ruleExtractProfile(memories)
	}

	content := strings.TrimSpace(respData.Choices[0].Message.Content)
	startIdx := strings.Index(content, "{")
	endIdx := strings.LastIndex(content, "}")
	if startIdx >= 0 && endIdx > startIdx {
		return content[startIdx : endIdx+1]
	}
	return s.ruleExtractProfile(memories)
}

// ruleExtractProfile 基于规则的简单画像提炼（LLM 不可用时的回退方案）
func (s *MemorySummarizer) ruleExtractProfile(memories []*database.Memory) string {
	profile := map[string]string{
		"learning_style":  "",
		"knowledge_gaps":  "",
		"strengths":       "",
		"interests":       "",
		"personality":     "",
		"learning_habits": "",
		"summary":         "",
	}

	for _, m := range memories {
		switch m.MemoryType {
		case "learning_style":
			profile["learning_style"] = m.Content
		case "knowledge_gap":
			profile["knowledge_gaps"] = m.Content
		case "strength":
			profile["strengths"] = m.Content
		case "personality", "personality_traits":
			profile["personality"] = m.Content
		case "learning_progress":
			if profile["summary"] == "" {
				profile["summary"] = m.Content
			}
		}
	}

	data, _ := json.Marshal(profile)
	return string(data)
}
