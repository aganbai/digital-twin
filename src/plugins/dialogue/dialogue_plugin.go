package dialogue

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"digital-twin/src/backend/database"
	"digital-twin/src/harness/core"

	"github.com/google/uuid"
)

// DialoguePlugin 对话插件
type DialoguePlugin struct {
	*core.BasePlugin
	db        *sql.DB
	convRepo  *database.ConversationRepository
	memRepo   *database.MemoryRepository
	llmClient *LLMClient
	prompt    *PromptBuilder

	// 配置
	historyLimit int // 对话历史条数限制
}

// NewDialoguePlugin 创建对话插件
func NewDialoguePlugin(name string, db *sql.DB) *DialoguePlugin {
	return &DialoguePlugin{
		BasePlugin:   core.NewBasePlugin(name, "1.0.0", core.PluginTypeDialogue),
		db:           db,
		convRepo:     database.NewConversationRepository(db),
		memRepo:      database.NewMemoryRepository(db),
		prompt:       NewPromptBuilder(),
		historyLimit: 10,
	}
}

// Init 初始化对话插件
func (p *DialoguePlugin) Init(config map[string]interface{}) error {
	if err := p.BasePlugin.Init(config); err != nil {
		return err
	}

	// 读取 LLM 配置
	mode := "mock"
	model := "qwen-turbo"
	apiKey := ""
	baseURL := ""
	temperature := 0.7
	maxTokens := 1000

	if v, ok := config["llm_provider.mode"]; ok {
		if s, ok := v.(string); ok {
			mode = s
		}
	}
	if v, ok := config["llm_provider.model"]; ok {
		if s, ok := v.(string); ok {
			model = s
		}
	}
	if v, ok := config["llm_provider.api_key"]; ok {
		if s, ok := v.(string); ok {
			apiKey = s
		}
	}
	if v, ok := config["llm_provider.base_url"]; ok {
		if s, ok := v.(string); ok {
			baseURL = s
		}
	}

	// 读取对话策略配置
	if v, ok := config["dialogue_strategy.temperature"]; ok {
		temperature = toFloat64(v, temperature)
	}
	if v, ok := config["dialogue_strategy.max_tokens"]; ok {
		maxTokens = toInt(v, maxTokens)
	}

	// 读取上下文管理配置
	if v, ok := config["context_management.history_limit"]; ok {
		p.historyLimit = toInt(v, p.historyLimit)
	}

	// 创建 LLM 客户端
	p.llmClient = NewLLMClient(mode, model, apiKey, baseURL, temperature, maxTokens)

	return nil
}

// Execute 执行对话操作
func (p *DialoguePlugin) Execute(ctx context.Context, input *core.PluginInput) (*core.PluginOutput, error) {
	start := time.Now()

	action, _ := input.Data["action"].(string)
	if action == "" {
		return &core.PluginOutput{
			Success:  false,
			Data:     map[string]interface{}{"error_code": 40004},
			Error:    "缺少 action 参数",
			Duration: time.Since(start),
		}, nil
	}

	var output *core.PluginOutput
	var err error

	switch action {
	case "chat":
		output, err = p.handleChat(input, start)
	case "history":
		output, err = p.handleHistory(input)
	default:
		output = &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   fmt.Sprintf("不支持的 action: %s", action),
		}
	}

	if err != nil {
		return &core.PluginOutput{
			Success:  false,
			Data:     map[string]interface{}{"error_code": 50001},
			Error:    err.Error(),
			Duration: time.Since(start),
		}, nil
	}

	output.Duration = time.Since(start)
	return output, nil
}

// handleChat 对话生成
func (p *DialoguePlugin) handleChat(input *core.PluginInput, pipelineStart time.Time) (*core.PluginOutput, error) {
	// 1. 从 Data 获取参数
	message, _ := input.Data["message"].(string)
	if message == "" {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "消息内容不能为空",
		}, nil
	}

	var teacherID int64
	if v, ok := input.Data["teacher_id"]; ok {
		teacherID = toInt64(v, 0)
	}
	if teacherID <= 0 {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "缺少 teacher_id",
		}, nil
	}

	// 获取 student_id（从 UserContext）
	var studentID int64
	if input.UserContext != nil && input.UserContext.UserID != "" {
		studentID = toInt64(input.UserContext.UserID, 0)
	}
	if studentID <= 0 {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "缺少用户信息",
		}, nil
	}

	// session_id（无则自动生成 UUID）
	sessionID, _ := input.Data["session_id"].(string)
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	// 2. 从 Data 获取上游注入的 memories 和 chunks
	var memories []map[string]interface{}
	if v, ok := input.Data["memories"]; ok {
		if mems, ok := v.([]map[string]interface{}); ok {
			memories = mems
		}
	}

	var chunks []map[string]interface{}
	if v, ok := input.Data["chunks"]; ok {
		if cks, ok := v.([]map[string]interface{}); ok {
			chunks = cks
		}
	}

	// 3. 查询最近对话历史
	history, _, err := p.convRepo.GetByStudentAndTeacher(studentID, teacherID, 0, p.historyLimit)
	if err != nil {
		return nil, fmt.Errorf("查询对话历史失败: %w", err)
	}

	// 4. 用 PromptBuilder 构建完整 prompt
	systemPrompt := p.prompt.BuildSystemPrompt(chunks, memories)
	chatMessages := p.prompt.BuildConversationMessages(systemPrompt, history, message)

	// 5. 调用 LLMClient 生成回复
	chatResp, err := p.llmClient.Chat(chatMessages)
	if err != nil {
		return nil, fmt.Errorf("调用 LLM 失败: %w", err)
	}

	// 6. 保存用户消息和 AI 回复到 conversations 表
	userConv := &database.Conversation{
		StudentID:  studentID,
		TeacherID:  teacherID,
		SessionID:  sessionID,
		Role:       "user",
		Content:    message,
		TokenCount: chatResp.PromptTokens,
	}
	_, err = p.convRepo.Create(userConv)
	if err != nil {
		return nil, fmt.Errorf("保存用户消息失败: %w", err)
	}

	aiConv := &database.Conversation{
		StudentID:  studentID,
		TeacherID:  teacherID,
		SessionID:  sessionID,
		Role:       "assistant",
		Content:    chatResp.Content,
		TokenCount: chatResp.CompletionTokens,
	}
	convID, err := p.convRepo.Create(aiConv)
	if err != nil {
		return nil, fmt.Errorf("保存 AI 回复失败: %w", err)
	}

	// 7. 异步提取记忆并存储
	go p.extractAndStoreMemories(studentID, teacherID, message, chatResp.Content)

	// 8. 计算管道耗时
	pipelineDuration := time.Since(pipelineStart).Milliseconds()

	// 9. 返回结果（chat action 不需要 merge 上游 Data，是管道最后一个插件）
	outputData := map[string]interface{}{
		"reply":                chatResp.Content,
		"session_id":           sessionID,
		"conversation_id":      convID,
		"token_usage": map[string]interface{}{
			"prompt_tokens":     chatResp.PromptTokens,
			"completion_tokens": chatResp.CompletionTokens,
			"total_tokens":      chatResp.TotalTokens,
		},
		"pipeline_duration_ms": pipelineDuration,
	}

	return &core.PluginOutput{
		Success:  true,
		Data:     outputData,
		Metadata: map[string]interface{}{"plugin": "dialogue", "action": "chat"},
	}, nil
}

// extractAndStoreMemories 异步提取记忆并存储
func (p *DialoguePlugin) extractAndStoreMemories(studentID, teacherID int64, userMessage, aiReply string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[记忆提取] panic recovered: %v\n", r)
		}
	}()

	// 构建记忆提取 prompt
	messages := p.prompt.BuildMemoryExtractionPrompt(userMessage, aiReply)

	// 调用 LLM 提取记忆
	memories, err := p.llmClient.ExtractMemories(messages)
	if err != nil {
		fmt.Printf("[记忆提取] 提取失败: %v\n", err)
		return
	}

	// 存储每条记忆
	for _, mem := range memories {
		if mem.Content == "" {
			continue
		}
		_, err := p.memRepo.Create(&database.Memory{
			StudentID:  studentID,
			TeacherID:  teacherID,
			MemoryType: mem.Type,
			Content:    mem.Content,
			Importance: mem.Importance,
		})
		if err != nil {
			fmt.Printf("[记忆提取] 存储记忆失败: %v\n", err)
		}
	}
}

// handleHistory 查询对话历史
func (p *DialoguePlugin) handleHistory(input *core.PluginInput) (*core.PluginOutput, error) {
	var teacherID int64
	if v, ok := input.Data["teacher_id"]; ok {
		teacherID = toInt64(v, 0)
	}
	if teacherID <= 0 {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "缺少 teacher_id",
		}, nil
	}

	// 从 UserContext 获取 student_id
	var studentID int64
	if input.UserContext != nil && input.UserContext.UserID != "" {
		studentID = toInt64(input.UserContext.UserID, 0)
	}
	if studentID <= 0 {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "缺少用户信息",
		}, nil
	}

	page := 1
	if v, ok := input.Data["page"]; ok {
		page = toInt(v, 1)
	}
	pageSize := 10
	if v, ok := input.Data["page_size"]; ok {
		pageSize = toInt(v, 10)
	}

	offset := (page - 1) * pageSize
	convs, total, err := p.convRepo.GetByStudentAndTeacher(studentID, teacherID, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询对话历史失败: %w", err)
	}

	// 转换为输出格式
	items := make([]map[string]interface{}, 0, len(convs))
	for _, conv := range convs {
		items = append(items, map[string]interface{}{
			"id":          conv.ID,
			"session_id":  conv.SessionID,
			"role":        conv.Role,
			"content":     conv.Content,
			"token_count": conv.TokenCount,
			"created_at":  conv.CreatedAt.Format(time.RFC3339),
		})
	}

	// history action 需要 merge 上游 Data
	outputData := mergeData(input.Data, map[string]interface{}{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})

	return &core.PluginOutput{
		Success:  true,
		Data:     outputData,
		Metadata: map[string]interface{}{"plugin": "dialogue", "action": "history"},
	}, nil
}

// mergeData 合并上游 Data 和本插件输出字段
func mergeData(upstream map[string]interface{}, pluginData map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range upstream {
		result[k] = v
	}
	for k, v := range pluginData {
		result[k] = v
	}
	return result
}

// toInt 将 interface{} 转换为 int
func toInt(v interface{}, defaultVal int) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	default:
		return defaultVal
	}
}

// toInt64 将 interface{} 转换为 int64
func toInt64(v interface{}, defaultVal int64) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case float64:
		return int64(val)
	case string:
		var result int64
		fmt.Sscanf(val, "%d", &result)
		if result > 0 {
			return result
		}
		return defaultVal
	default:
		return defaultVal
	}
}

// toFloat64 将 interface{} 转换为 float64
func toFloat64(v interface{}, defaultVal float64) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	default:
		return defaultVal
	}
}
