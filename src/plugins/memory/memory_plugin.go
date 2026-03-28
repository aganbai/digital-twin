package memory

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"digital-twin/src/backend/database"
	"digital-twin/src/harness/core"
)

// MemoryPlugin 记忆插件
type MemoryPlugin struct {
	*core.BasePlugin
	db        *sql.DB
	memRepo   *database.MemoryRepository
	store     *MemoryStore
	maxMemories int // 每个用户最大记忆数量
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

// Init 初始化记忆插件
func (p *MemoryPlugin) Init(config map[string]interface{}) error {
	if err := p.BasePlugin.Init(config); err != nil {
		return err
	}

	// 读取 retention.max_memories_per_user 配置
	if v, ok := config["retention.max_memories_per_user"]; ok {
		p.maxMemories = toInt(v, p.maxMemories)
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
	case "list":
		output, err = p.handleList(input)
	default:
		// 管道模式：action 不是 memory 插件自己的 action（包括空 action 和其他插件的 action 如 "chat"）
		output, err = p.handlePipeline(input)
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

	memories, err := p.store.RetrieveRelevant(studentID, teacherID, limit)
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

	memoryID, err := p.store.StoreMemory(studentID, teacherID, memoryType, content, importance)
	if err != nil {
		return nil, fmt.Errorf("存储记忆失败: %w", err)
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

	// merge 上游 Data
	outputData := mergeData(input.Data, nil)

	// 如果有 student_id 和 teacher_id，自动检索记忆
	if studentID > 0 && teacherID > 0 {
		limit := 10
		memories, err := p.store.RetrieveRelevant(studentID, teacherID, limit)
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

// memoriesToMapSlice 将 Memory 切片转换为 map 切片
func memoriesToMapSlice(memories []*database.Memory) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(memories))
	for _, mem := range memories {
		m := map[string]interface{}{
			"id":          mem.ID,
			"student_id":  mem.StudentID,
			"teacher_id":  mem.TeacherID,
			"memory_type": mem.MemoryType,
			"content":     mem.Content,
			"importance":  mem.Importance,
			"created_at":  mem.CreatedAt.Format(time.RFC3339),
			"updated_at":  mem.UpdatedAt.Format(time.RFC3339),
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
