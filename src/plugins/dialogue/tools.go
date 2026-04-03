package dialogue

import "fmt"

// Tool 工具接口，所有可被 LLM 调用的工具需实现此接口
type Tool interface {
	// Name 返回工具名称（唯一标识）
	Name() string
	// Definition 返回 OpenAI Function Calling 格式的工具定义
	Definition() map[string]interface{}
	// Execute 执行工具调用，接收参数 JSON 字符串，返回结果文本
	Execute(arguments string) (string, error)
}

// ToolDefinition 工具定义结构体（OpenAI Function Calling 格式）
type ToolDefinition struct {
	Type     string          `json:"type"`     // 固定为 "function"
	Function ToolFunctionDef `json:"function"` // 函数定义
}

// ToolFunctionDef 工具函数定义
type ToolFunctionDef struct {
	Name        string                 `json:"name"`        // 函数名称
	Description string                 `json:"description"` // 函数描述
	Parameters  map[string]interface{} `json:"parameters"`  // 参数 JSON Schema
}

// ToolRegistry 工具注册中心，管理所有可用工具
type ToolRegistry struct {
	tools map[string]Tool
}

// NewToolRegistry 创建工具注册中心
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register 注册工具
func (r *ToolRegistry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

// Get 获取指定名称的工具
func (r *ToolRegistry) Get(name string) (Tool, error) {
	tool, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("未注册的工具: %s", name)
	}
	return tool, nil
}

// GetAllDefinitions 获取所有工具的定义列表（用于发送给 LLM）
func (r *ToolRegistry) GetAllDefinitions() []map[string]interface{} {
	defs := make([]map[string]interface{}, 0, len(r.tools))
	for _, tool := range r.tools {
		defs = append(defs, tool.Definition())
	}
	return defs
}

// HasTools 是否有注册的工具
func (r *ToolRegistry) HasTools() bool {
	return len(r.tools) > 0
}
