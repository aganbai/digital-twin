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
	db         *sql.DB // V2.0 迭代7: 用于更新用户画像; V2.0 迭代9: 用于老师画像生成
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

	// V2.0 迭代9: 执行老师画像生成任务
	s.runTeacherProfileGeneration()
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

// ======================== V2.0 迭代9: 老师画像生成 ========================

// TeacherProfile 老师画像数据结构
type TeacherProfile struct {
	TeachingStyle          string   `json:"teaching_style"`
	ResponseFrequency      string   `json:"response_frequency"`
	AvgResponseLength      string   `json:"avg_response_length"`
	CoursePublishFrequency string   `json:"course_publish_frequency"`
	ActiveStudentCount     int      `json:"active_student_count"`
	KnowledgeContentTags   []string `json:"knowledge_content_tags"`
	LastUpdated            string   `json:"last_updated"`
	AvgResponseTime        string   `json:"avg_response_time,omitempty"`
	TotalConversations     int      `json:"total_conversations,omitempty"`
	TotalCoursesPublished  int      `json:"total_courses_published,omitempty"`
}

// generateTeacherProfile 分析老师行为数据并生成画像
func (s *MemorySummarizer) generateTeacherProfile(teacherID int64) (*TeacherProfile, error) {
	if s.db == nil {
		return nil, fmt.Errorf("数据库连接不可用")
	}

	profile := &TeacherProfile{
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
	}

	// 1. 分析过去30天的对话记录
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	// 1.1 回复频率和平均长度
	var totalReplies int
	var totalLength int
	var replyCount int

	err := s.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(LENGTH(content)), 0), COUNT(CASE WHEN LENGTH(content) > 0 THEN 1 END)
		FROM conversations 
		WHERE teacher_id = ? AND role = 'assistant' AND created_at >= ?`,
		teacherID, thirtyDaysAgo,
	).Scan(&totalReplies, &totalLength, &replyCount)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("[MemorySummarizer] 查询回复数据失败 (teacher_id=%d): %v", teacherID, err)
	} else if replyCount > 0 {
		avgLength := totalLength / replyCount

		// 回复频率分级
		dailyAvg := float64(totalReplies) / 30.0
		switch {
		case dailyAvg >= 50:
			profile.ResponseFrequency = fmt.Sprintf("高频（日均回复%.0f条）", dailyAvg)
		case dailyAvg >= 20:
			profile.ResponseFrequency = fmt.Sprintf("中频（日均回复%.0f条）", dailyAvg)
		default:
			profile.ResponseFrequency = fmt.Sprintf("低频（日均回复%.0f条）", dailyAvg)
		}

		// 平均回复长度分级
		switch {
		case avgLength >= 200:
			profile.AvgResponseLength = fmt.Sprintf("较长（平均%d字）", avgLength)
		case avgLength >= 100:
			profile.AvgResponseLength = fmt.Sprintf("中等（平均%d字）", avgLength)
		default:
			profile.AvgResponseLength = fmt.Sprintf("简洁（平均%d字）", avgLength)
		}
	}

	profile.TotalConversations = totalReplies

	// 1.2 计算平均响应时间（学生发问到老师回复的时间差）
	var avgResponseTimeMinutes float64
	err = s.db.QueryRow(`
		SELECT AVG(
			EXTRACT(EPOCH FROM (teacher_msg.created_at - student_msg.created_at)) / 60
		)
		FROM conversations teacher_msg
		JOIN conversations student_msg ON teacher_msg.session_id = student_msg.session_id
			AND student_msg.role = 'user'
			AND teacher_msg.role = 'assistant'
			AND teacher_msg.created_at > student_msg.created_at
		WHERE teacher_msg.teacher_id = ?
			AND teacher_msg.created_at >= ?
			AND student_msg.created_at >= ?`,
		teacherID, thirtyDaysAgo, thirtyDaysAgo,
	).Scan(&avgResponseTimeMinutes)

	if err == nil && avgResponseTimeMinutes > 0 {
		switch {
		case avgResponseTimeMinutes < 5:
			profile.AvgResponseTime = "即时响应（<5分钟）"
		case avgResponseTimeMinutes < 30:
			profile.AvgResponseTime = "快速响应（5-30分钟）"
		case avgResponseTimeMinutes < 120:
			profile.AvgResponseTime = "正常响应（30分钟-2小时）"
		default:
			profile.AvgResponseTime = fmt.Sprintf("延迟响应（平均%.0f分钟）", avgResponseTimeMinutes)
		}
	}

	// 2. 统计活跃学生数
	err = s.db.QueryRow(`
		SELECT COUNT(DISTINCT student_persona_id)
		FROM conversations
		WHERE teacher_id = ? AND created_at >= ? AND student_persona_id > 0`,
		teacherID, thirtyDaysAgo,
	).Scan(&profile.ActiveStudentCount)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("[MemorySummarizer] 查询活跃学生数失败 (teacher_id=%d): %v", teacherID, err)
		profile.ActiveStudentCount = 0
	}

	// 3. 查询课程发布频率和内容标签
	var courseCount int
	var firstCourseTime, lastCourseTime *time.Time

	err = s.db.QueryRow(`
		SELECT COUNT(*), MIN(created_at), MAX(created_at)
		FROM knowledge_items
		WHERE teacher_id = ? AND item_type = 'course' AND status = 'active'`,
		teacherID,
	).Scan(&courseCount, &firstCourseTime, &lastCourseTime)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("[MemorySummarizer] 查询课程数据失败 (teacher_id=%d): %v", teacherID, err)
	} else if courseCount > 0 && firstCourseTime != nil && lastCourseTime != nil {
		// 计算课程发布频率
		days := lastCourseTime.Sub(*firstCourseTime).Hours() / 24
		if days > 0 {
			weeklyAvg := float64(courseCount) / (days / 7.0)
			switch {
			case weeklyAvg >= 3:
				profile.CoursePublishFrequency = fmt.Sprintf("高频（每周%.1f次）", weeklyAvg)
			case weeklyAvg >= 1:
				profile.CoursePublishFrequency = fmt.Sprintf("中频（每周%.1f次）", weeklyAvg)
			default:
				profile.CoursePublishFrequency = fmt.Sprintf("低频（每月%.1f次）", weeklyAvg*4)
			}
		} else {
			profile.CoursePublishFrequency = "刚起步"
		}
	}
	profile.TotalCoursesPublished = courseCount

	// 4. 提取知识库内容标签
	rows, err := s.db.Query(`
		SELECT DISTINCT tags
		FROM knowledge_items
		WHERE teacher_id = ? AND status = 'active' AND tags IS NOT NULL AND tags != '[]'`,
		teacherID,
	)

	if err == nil {
		defer rows.Close()
		tagSet := make(map[string]bool)
		for rows.Next() {
			var tagsJSON string
			if err := rows.Scan(&tagsJSON); err != nil {
				continue
			}
			var tags []string
			if err := json.Unmarshal([]byte(tagsJSON), &tags); err == nil {
				for _, tag := range tags {
					tag = strings.TrimSpace(tag)
					if tag != "" && len(tagSet) < 10 { // 最多10个标签
						tagSet[tag] = true
					}
				}
			}
		}
		for tag := range tagSet {
			profile.KnowledgeContentTags = append(profile.KnowledgeContentTags, tag)
		}
	}

	// 5. 推断教学风格
	profile.TeachingStyle = s.inferTeachingStyle(profile)

	return profile, nil
}

// inferTeachingStyle 根据画像数据推断教学风格
func (s *MemorySummarizer) inferTeachingStyle(profile *TeacherProfile) string {
	// 基于规则推断教学风格
	// 高回复频率 + 中等长度 → 苏格拉底式引导
	// 中等频率 + 较长回复 → 详细讲解型
	// 低频率 + 简洁回复 → 点拨启发型

	if profile.TotalConversations == 0 {
		return "待观察"
	}

	// 默认风格
	style := "综合型"

	// 根据回复特征判断
	if strings.Contains(profile.ResponseFrequency, "高频") {
		if strings.Contains(profile.AvgResponseLength, "简洁") {
			style = "互动引导型"
		} else if strings.Contains(profile.AvgResponseLength, "较长") {
			style = "细致辅导型"
		} else {
			style = "苏格拉底式引导"
		}
	} else if strings.Contains(profile.ResponseFrequency, "中频") {
		if strings.Contains(profile.AvgResponseLength, "较长") {
			style = "详细讲解型"
		} else {
			style = "循序渐进型"
		}
	} else {
		style = "点拨启发型"
	}

	// 如果有LLM，尝试调用LLM分析
	if s.llmBaseURL != "" && s.llmAPIKey != "" {
		llmStyle := s.callLLMForTeachingStyle(profile)
		if llmStyle != "" {
			style = llmStyle
		}
	}

	return style
}

// callLLMForTeachingStyle 调用LLM分析教学风格
func (s *MemorySummarizer) callLLMForTeachingStyle(profile *TeacherProfile) string {
	profileJSON, _ := json.MarshalIndent(profile, "", "  ")

	prompt := fmt.Sprintf(`你是一个教学风格分析专家。请根据以下老师画像数据，用一句话描述该老师的教学风格特点。

老师画像数据：
%s

请直接输出教学风格描述，不超过20个字。例如：善于启发引导、注重互动交流、讲解细致入微等。`, string(profileJSON))

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
		"max_tokens":  50,
	}

	jsonData, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return ""
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.llmAPIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ""
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
		return ""
	}

	if len(respData.Choices) == 0 {
		return ""
	}

	return strings.TrimSpace(respData.Choices[0].Message.Content)
}

// runTeacherProfileGeneration 执行所有老师的画像生成
func (s *MemorySummarizer) runTeacherProfileGeneration() {
	log.Println("[MemorySummarizer] 开始执行老师画像生成任务...")

	if s.db == nil {
		log.Println("[MemorySummarizer] 数据库连接不可用，跳过老师画像生成")
		return
	}

	// 1. 获取所有老师用户
	rows, err := s.db.Query(`SELECT id FROM users WHERE role = 'teacher'`)
	if err != nil {
		log.Printf("[MemorySummarizer] 查询老师列表失败: %v", err)
		return
	}
	defer rows.Close()

	var teacherIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			continue
		}
		teacherIDs = append(teacherIDs, id)
	}

	if len(teacherIDs) == 0 {
		log.Println("[MemorySummarizer] 没有找到老师用户")
		return
	}

	log.Printf("[MemorySummarizer] 找到 %d 个老师用户，开始生成画像...", len(teacherIDs))

	successCount := 0
	for _, teacherID := range teacherIDs {
		// 2. 生成老师画像
		profile, err := s.generateTeacherProfile(teacherID)
		if err != nil {
			log.Printf("[MemorySummarizer] 生成老师画像失败 (teacher_id=%d): %v", teacherID, err)
			continue
		}

		// 3. 存储到 users.profile_snapshot
		profileJSON, err := json.Marshal(profile)
		if err != nil {
			log.Printf("[MemorySummarizer] 序列化老师画像失败 (teacher_id=%d): %v", teacherID, err)
			continue
		}

		userRepo := database.NewUserRepository(s.db)
		if err := userRepo.UpdateProfileSnapshot(teacherID, string(profileJSON)); err != nil {
			log.Printf("[MemorySummarizer] 更新老师画像失败 (teacher_id=%d): %v", teacherID, err)
			continue
		}

		successCount++
		log.Printf("[MemorySummarizer] 老师画像已更新 (teacher_id=%d)", teacherID)
	}

	log.Printf("[MemorySummarizer] 老师画像生成任务完成，成功 %d/%d", successCount, len(teacherIDs))
}
