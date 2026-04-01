package memory

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"digital-twin/src/backend/database"
	"digital-twin/src/harness/core"
)

// MemoryPlugin 记忆插件
type MemoryPlugin struct {
	*core.BasePlugin
	db          *sql.DB
	memRepo     *database.MemoryRepository
	store       *MemoryStore
	maxMemories int // 每个用户最大记忆数量

	// V2.0 迭代6: LLM 层级判断配置
	llmBaseURL string
	llmAPIKey  string
	llmModel   string
}

// NewMemoryPlugin 创建记忆插件
func NewMemoryPlugin(name string, db *sql.DB) *MemoryPlugin {
	memRepo := database.NewMemoryRepository(db)
	return &MemoryPlugin{
		BasePlugin:  core.NewBasePlugin(name, "1.0.0", core.PluginTypeMemory),
		db:          db,
		memRepo:     memRepo,
		store:       NewMemoryStore(memRepo),
		maxMemories: 100, // 默认值
	}
}

// GetStore 获取 MemoryStore（供外部模块使用，如定时任务）
func (p *MemoryPlugin) GetStore() *MemoryStore {
	return p.store
}

// GetRepo 获取 MemoryRepository（供外部模块使用，如定时任务）
func (p *MemoryPlugin) GetRepo() *database.MemoryRepository {
	return p.memRepo
}

// Init 初始化记忆插件
func (p *MemoryPlugin) Init(config map[string]interface{}) error {
	if err := p.BasePlugin.Init(config); err != nil {
		return err
	}

	// 读取 retention.max_memories_per_user 配置
	if v, ok := config["retention.max_memories_per_user"]; ok {
		p.maxMemories = toInt(v, p.maxMemories)
	}

	// V2.0 迭代6: 读取 LLM 配置（用于记忆层级判断）
	if v, ok := config["llm_base_url"]; ok {
		p.llmBaseURL, _ = v.(string)
	}
	if v, ok := config["llm_api_key"]; ok {
		p.llmAPIKey, _ = v.(string)
	}
	if v, ok := config["llm_model"]; ok {
		p.llmModel, _ = v.(string)
	}

	return nil
}

// Execute 执行记忆操作
// 根据 input.Data["action"] 分发到不同的处理逻辑
// 无 action 时（在管道中执行）：自动检索该学生的相关记忆，注入到 Data["memories"]
func (p *MemoryPlugin) Execute(ctx context.Context, input *core.PluginInput) (*core.PluginOutput, error) {
	start := time.Now()

	action, _ := input.Data["action"].(string)

	var output *core.PluginOutput
	var err error

	switch action {
	case "retrieve":
		output, err = p.handleRetrieve(input)
	case "store":
		output, err = p.handleStore(input)
	case "store_with_layer":
		output, err = p.handleStoreWithLayer(input)
	case "list":
		output, err = p.handleList(input)
	case "list_with_filter":
		output, err = p.handleListWithFilter(input)
	default:
		// 判断是否为管道模式：需要有 message 字段（管道中上游会传入 message）
		_, hasMessage := input.Data["message"]
		if !hasMessage {
			// 非管道模式，action 无效或缺失
			errMsg := "缺少 action 参数"
			if action != "" {
				errMsg = fmt.Sprintf("不支持的 action: %s", action)
			}
			output = &core.PluginOutput{
				Success: false,
				Data:    map[string]interface{}{"error_code": 40004},
				Error:   errMsg,
			}
		} else {
			// 管道模式：action 不是 memory 插件自己的 action（包括空 action 和其他插件的 action 如 "chat"）
			output, err = p.handlePipeline(input)
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

// handleRetrieve 检索记忆
func (p *MemoryPlugin) handleRetrieve(input *core.PluginInput) (*core.PluginOutput, error) {
	var studentID, teacherID int64
	if v, ok := input.Data["student_id"]; ok {
		studentID = toInt64(v, 0)
	}
	if v, ok := input.Data["teacher_id"]; ok {
		teacherID = toInt64(v, 0)
	}
	// 获取分身 ID
	var teacherPersonaID, studentPersonaID int64
	if v, ok := input.Data["teacher_persona_id"]; ok {
		teacherPersonaID = toInt64(v, 0)
	}
	if v, ok := input.Data["student_persona_id"]; ok {
		studentPersonaID = toInt64(v, 0)
	}
	limit := 10
	if v, ok := input.Data["limit"]; ok {
		limit = toInt(v, 10)
	}

	if studentID <= 0 {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "缺少 student_id",
		}, nil
	}
	if teacherID <= 0 {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "缺少 teacher_id",
		}, nil
	}

	// 当 persona_id > 0 时，优先使用分身维度检索
	var memories []*database.Memory
	var err error
	if teacherPersonaID > 0 && studentPersonaID > 0 {
		memories, err = p.store.RetrieveRelevantByPersonas(teacherPersonaID, studentPersonaID, limit)
	}
	if (memories == nil || len(memories) == 0) && err == nil {
		// 分身维度未找到，回退到 user_id 维度
		memories, err = p.store.RetrieveRelevant(studentID, teacherID, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("检索记忆失败: %w", err)
	}

	memoriesOutput := memoriesToMapSlice(memories)

	// merge 上游 Data
	outputData := mergeData(input.Data, map[string]interface{}{
		"memories": memoriesOutput,
	})

	return &core.PluginOutput{
		Success:  true,
		Data:     outputData,
		Metadata: map[string]interface{}{"plugin": "memory", "action": "retrieve"},
	}, nil
}

// handleStore 存储记忆
func (p *MemoryPlugin) handleStore(input *core.PluginInput) (*core.PluginOutput, error) {
	var studentID, teacherID int64
	if v, ok := input.Data["student_id"]; ok {
		studentID = toInt64(v, 0)
	}
	if v, ok := input.Data["teacher_id"]; ok {
		teacherID = toInt64(v, 0)
	}
	// 获取分身 ID
	var teacherPersonaID, studentPersonaID int64
	if v, ok := input.Data["teacher_persona_id"]; ok {
		teacherPersonaID = toInt64(v, 0)
	}
	if v, ok := input.Data["student_persona_id"]; ok {
		studentPersonaID = toInt64(v, 0)
	}
	memoryType, _ := input.Data["memory_type"].(string)
	content, _ := input.Data["content"].(string)
	importance := 0.5
	if v, ok := input.Data["importance"]; ok {
		importance = toFloat64(v, 0.5)
	}

	if studentID <= 0 {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "缺少 student_id",
		}, nil
	}
	if teacherID <= 0 {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "缺少 teacher_id",
		}, nil
	}
	if content == "" {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "记忆内容不能为空",
		}, nil
	}

	// V2.0 迭代6: 获取记忆层级（如果指定了 memory_layer 则使用，否则默认 episodic）
	memoryLayer, _ := input.Data["memory_layer"].(string)
	if memoryLayer == "" {
		memoryLayer = p.callLLMForMemoryLayer(content, memoryType)
	}

	// 限制 episodic 上限
	if memoryLayer == "episodic" && teacherPersonaID > 0 && studentPersonaID > 0 {
		count, err := p.memRepo.CountEpisodicMemories(teacherPersonaID, studentPersonaID)
		if err == nil && count >= 50 {
			// 超过上限，返回错误，后续由定时任务合并
			return &core.PluginOutput{
				Success: false,
				Data:    map[string]interface{}{"error_code": 40005},
				Error:   "episodic 记忆超过上限 50 条",
			}, nil
		}
	}

	// 当 persona_id > 0 时，使用分身维度存储（带层级）
	var memoryID int64
	var storeErr error
	if teacherPersonaID > 0 && studentPersonaID > 0 {
		memoryID, storeErr = p.store.StoreMemoryWithLayer(studentID, teacherID, teacherPersonaID, studentPersonaID, memoryType, memoryLayer, content, importance)
	} else {
		memoryID, storeErr = p.store.StoreMemory(studentID, teacherID, memoryType, content, importance)
	}
	if storeErr != nil {
		return nil, fmt.Errorf("存储记忆失败: %w", storeErr)
	}

	// merge 上游 Data
	outputData := mergeData(input.Data, map[string]interface{}{
		"memory_id": memoryID,
	})

	return &core.PluginOutput{
		Success:  true,
		Data:     outputData,
		Metadata: map[string]interface{}{"plugin": "memory", "action": "store"},
	}, nil
}

// handleList 列表查询
func (p *MemoryPlugin) handleList(input *core.PluginInput) (*core.PluginOutput, error) {
	var studentID, teacherID int64
	if v, ok := input.Data["student_id"]; ok {
		studentID = toInt64(v, 0)
	}
	if v, ok := input.Data["teacher_id"]; ok {
		teacherID = toInt64(v, 0)
	}
	memoryType, _ := input.Data["memory_type"].(string)
	page := 1
	if v, ok := input.Data["page"]; ok {
		page = toInt(v, 1)
	}
	pageSize := 10
	if v, ok := input.Data["page_size"]; ok {
		pageSize = toInt(v, 10)
	}

	if studentID <= 0 {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "缺少 student_id",
		}, nil
	}
	if teacherID <= 0 {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "缺少 teacher_id",
		}, nil
	}

	memories, total, err := p.store.ListMemories(studentID, teacherID, memoryType, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询记忆列表失败: %w", err)
	}

	memoriesOutput := memoriesToMapSlice(memories)

	// merge 上游 Data
	outputData := mergeData(input.Data, map[string]interface{}{
		"memories":  memoriesOutput,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})

	return &core.PluginOutput{
		Success:  true,
		Data:     outputData,
		Metadata: map[string]interface{}{"plugin": "memory", "action": "list"},
	}, nil
}

// handlePipeline 管道模式：自动检索记忆并注入到 Data
func (p *MemoryPlugin) handlePipeline(input *core.PluginInput) (*core.PluginOutput, error) {
	// 从 UserContext 获取 student_id
	var studentID int64
	if input.UserContext != nil && input.UserContext.UserID != "" {
		studentID = toInt64(input.UserContext.UserID, 0)
	}

	// 从 Data 获取 teacher_id
	var teacherID int64
	if v, ok := input.Data["teacher_id"]; ok {
		teacherID = toInt64(v, 0)
	}

	// 从 Data 获取分身 ID
	var teacherPersonaID, studentPersonaID int64
	if v, ok := input.Data["teacher_persona_id"]; ok {
		teacherPersonaID = toInt64(v, 0)
	}
	if v, ok := input.Data["student_persona_id"]; ok {
		studentPersonaID = toInt64(v, 0)
	}

	// merge 上游 Data
	outputData := mergeData(input.Data, nil)

	// 如果有 student_id 和 teacher_id，自动检索记忆
	if studentID > 0 && teacherID > 0 {
		limit := 10
		var memories []*database.Memory
		var err error

		// 当 persona_id > 0 时，优先使用分身维度检索
		if teacherPersonaID > 0 && studentPersonaID > 0 {
			memories, err = p.store.RetrieveRelevantByPersonas(teacherPersonaID, studentPersonaID, limit)
		}
		if (memories == nil || len(memories) == 0) && err == nil {
			// 分身维度未找到，回退到 user_id 维度
			memories, err = p.store.RetrieveRelevant(studentID, teacherID, limit)
		}

		if err != nil {
			// 管道模式下记忆检索失败不阻断流程，只记录错误
			outputData["memories"] = []map[string]interface{}{}
			outputData["memory_error"] = err.Error()
		} else {
			outputData["memories"] = memoriesToMapSlice(memories)
		}
	} else {
		outputData["memories"] = []map[string]interface{}{}
	}

	return &core.PluginOutput{
		Success:  true,
		Data:     outputData,
		Metadata: map[string]interface{}{"plugin": "memory", "mode": "pipeline"},
	}, nil
}

// handleStoreWithLayer 带层级判断的存储（V2.0 迭代6）
// 支持通过 LLM 判断记忆层级，或直接指定 memory_layer
func (p *MemoryPlugin) handleStoreWithLayer(input *core.PluginInput) (*core.PluginOutput, error) {
	var studentID, teacherID int64
	if v, ok := input.Data["student_id"]; ok {
		studentID = toInt64(v, 0)
	}
	if v, ok := input.Data["teacher_id"]; ok {
		teacherID = toInt64(v, 0)
	}
	var teacherPersonaID, studentPersonaID int64
	if v, ok := input.Data["teacher_persona_id"]; ok {
		teacherPersonaID = toInt64(v, 0)
	}
	if v, ok := input.Data["student_persona_id"]; ok {
		studentPersonaID = toInt64(v, 0)
	}
	memoryType, _ := input.Data["memory_type"].(string)
	content, _ := input.Data["content"].(string)
	importance := 0.5
	if v, ok := input.Data["importance"]; ok {
		importance = toFloat64(v, 0.5)
	}

	if studentID <= 0 || teacherID <= 0 {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "缺少 student_id 或 teacher_id",
		}, nil
	}
	if content == "" {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "记忆内容不能为空",
		}, nil
	}

	// 获取或判断记忆层级
	memoryLayer, _ := input.Data["memory_layer"].(string)
	if memoryLayer == "" {
		// 使用 LLM 判断层级
		memoryLayer = p.classifyMemoryLayer(content, memoryType)
	}

	// 存储
	var memoryID int64
	var storeErr error
	if teacherPersonaID > 0 && studentPersonaID > 0 {
		memoryID, storeErr = p.store.StoreMemoryWithLayer(studentID, teacherID, teacherPersonaID, studentPersonaID, memoryType, memoryLayer, content, importance)
	} else {
		memoryID, storeErr = p.store.StoreMemory(studentID, teacherID, memoryType, content, importance)
	}
	if storeErr != nil {
		return nil, fmt.Errorf("存储记忆失败: %w", storeErr)
	}

	outputData := mergeData(input.Data, map[string]interface{}{
		"memory_id":    memoryID,
		"memory_layer": memoryLayer,
	})

	return &core.PluginOutput{
		Success:  true,
		Data:     outputData,
		Metadata: map[string]interface{}{"plugin": "memory", "action": "store_with_layer"},
	}, nil
}

// handleListWithFilter SQL层分页+层级筛选（V2.0 迭代6）
func (p *MemoryPlugin) handleListWithFilter(input *core.PluginInput) (*core.PluginOutput, error) {
	var teacherPersonaID, studentPersonaID int64
	if v, ok := input.Data["teacher_persona_id"]; ok {
		teacherPersonaID = toInt64(v, 0)
	}
	if v, ok := input.Data["student_persona_id"]; ok {
		studentPersonaID = toInt64(v, 0)
	}
	layer, _ := input.Data["layer"].(string)
	page := 1
	if v, ok := input.Data["page"]; ok {
		page = toInt(v, 1)
	}
	pageSize := 20
	if v, ok := input.Data["page_size"]; ok {
		pageSize = toInt(v, 20)
	}

	if teacherPersonaID <= 0 || studentPersonaID <= 0 {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "缺少 teacher_persona_id 或 student_persona_id",
		}, nil
	}

	memories, total, err := p.store.ListMemoriesWithFilter(teacherPersonaID, studentPersonaID, layer, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询记忆列表失败: %w", err)
	}

	memoriesOutput := memoriesToMapSlice(memories)

	outputData := mergeData(input.Data, map[string]interface{}{
		"memories":  memoriesOutput,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})

	return &core.PluginOutput{
		Success:  true,
		Data:     outputData,
		Metadata: map[string]interface{}{"plugin": "memory", "action": "list_with_filter"},
	}, nil
}

// classifyMemoryLayer 使用规则判断记忆层级（LLM 辅助）
// 规则：importance >= 0.8 或 memory_type 为 preference/ability 时为 core，否则 episodic
func (p *MemoryPlugin) classifyMemoryLayer(content, memoryType string) string {
	// 基于 memory_type 的规则判断
	switch memoryType {
	case "preference", "ability", "learning_goal":
		return "core"
	case "interaction", "question", "feedback":
		return "episodic"
	default:
		return "episodic"
	}
}

// callLLMForMemoryLayer 调用 LLM 判断记忆层级
func (p *MemoryPlugin) callLLMForMemoryLayer(content, memoryType string) string {
	if p.llmBaseURL == "" || p.llmAPIKey == "" {
		return p.classifyMemoryLayer(content, memoryType)
	}

	url := strings.TrimRight(p.llmBaseURL, "/") + "/chat/completions"
	model := p.llmModel
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	prompt := fmt.Sprintf("你是一个记忆分析助手。请判断以下记忆内容是核心记忆(core)还是片段记忆(episodic)。"+
		"核心记忆通常包含用户的长期偏好、能力特征、人生目标等；"+
		"片段记忆通常包含具体的交互对话、单次提问、反馈等。\n"+
		"记忆类型：%s\n记忆内容：%s\n\n请直接输出'core'或'episodic'，不要输出其他任何内容。", memoryType, content)

	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]interface{}{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.0,
	}

	importBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", url, bytes.NewReader(importBytes))
	if err != nil {
		return p.classifyMemoryLayer(content, memoryType)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.llmAPIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return p.classifyMemoryLayer(content, memoryType)
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
		return p.classifyMemoryLayer(content, memoryType)
	}

	if len(respData.Choices) > 0 {
		result := strings.TrimSpace(strings.ToLower(respData.Choices[0].Message.Content))
		if strings.Contains(result, "core") {
			return "core"
		} else if strings.Contains(result, "episodic") {
			return "episodic"
		}
	}
	return p.classifyMemoryLayer(content, memoryType)
}

// memoriesToMapSlice 将 Memory 切片转换为 map 切片
func memoriesToMapSlice(memories []*database.Memory) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(memories))
	for _, mem := range memories {
		layer := mem.MemoryLayer
		if layer == "" {
			layer = "episodic"
		}
		m := map[string]interface{}{
			"id":           mem.ID,
			"student_id":   mem.StudentID,
			"teacher_id":   mem.TeacherID,
			"memory_type":  mem.MemoryType,
			"memory_layer": layer,
			"content":      mem.Content,
			"importance":   mem.Importance,
			"created_at":   mem.CreatedAt.Format(time.RFC3339),
			"updated_at":   mem.UpdatedAt.Format(time.RFC3339),
		}
		if mem.TeacherPersonaID > 0 {
			m["teacher_persona_id"] = mem.TeacherPersonaID
		}
		if mem.StudentPersonaID > 0 {
			m["student_persona_id"] = mem.StudentPersonaID
		}
		if mem.LastAccessed != nil {
			m["last_accessed"] = mem.LastAccessed.Format(time.RFC3339)
		}
		result = append(result, m)
	}
	return result
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
