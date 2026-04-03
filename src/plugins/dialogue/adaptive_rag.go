package dialogue

import (
	"encoding/json"
	"fmt"
	"log"
)

// AdaptiveRAG Adaptive RAG 管理器
// 实现 Function Calling 框架，让 LLM 自主决定是否需要外部搜索
type AdaptiveRAG struct {
	enabled          bool
	maxSearchPerTurn int
	llmClient        *LLMClient
	toolRegistry     *ToolRegistry
}

// NewAdaptiveRAG 创建 Adaptive RAG 管理器
func NewAdaptiveRAG(enabled bool, llmClient *LLMClient) *AdaptiveRAG {
	rag := &AdaptiveRAG{
		enabled:          enabled,
		maxSearchPerTurn: 2, // 默认每轮最多2次搜索
		llmClient:        llmClient,
		toolRegistry:     NewToolRegistry(),
	}

	// 注册 web_search 工具
	rag.toolRegistry.Register(NewWebSearch())

	return rag
}

// SetMaxSearchPerTurn 设置每轮最大搜索次数
func (r *AdaptiveRAG) SetMaxSearchPerTurn(max int) {
	if max > 0 {
		r.maxSearchPerTurn = max
	}
}

// ToolCallResult 工具调用结果
type ToolCallResult struct {
	ToolName string `json:"tool_name"`
	Query    string `json:"query"`
	Result   string `json:"result"`
}

// ProcessWithRAG 带 Adaptive RAG 的对话处理
// 返回最终回复内容和使用的工具列表
func (r *AdaptiveRAG) ProcessWithRAG(messages []ChatMessage) (string, []string, error) {
	if !r.enabled {
		// RAG 未启用，直接调用 LLM
		resp, err := r.llmClient.Chat(messages)
		if err != nil {
			return "", nil, err
		}
		return resp.Content, []string{}, nil
	}

	// 获取所有工具定义
	toolDefs := r.toolRegistry.GetAllDefinitions()

	// 第一轮调用：带 tools 定义，让 LLM 决定是否需要搜索
	resp, err := r.llmClient.ChatWithTools(messages, toolDefs)
	if err != nil {
		// 如果 ChatWithTools 不支持（如 mock 模式），回退到普通调用
		log.Printf("[AdaptiveRAG] ChatWithTools 失败，回退到普通调用: %v", err)
		plainResp, err := r.llmClient.Chat(messages)
		if err != nil {
			return "", nil, err
		}
		return plainResp.Content, []string{}, nil
	}

	// 检查是否有工具调用
	if len(resp.ToolCalls) == 0 {
		// LLM 决定不需要搜索，直接返回
		return resp.Content, []string{}, nil
	}

	// 执行工具调用（最多 maxSearchPerTurn 次）
	var toolsUsed []string
	searchCount := 0

	// 将 assistant 的 tool_calls 消息加入上下文（包含 tool_calls 字段）
	messages = append(messages, ChatMessage{
		Role:      "assistant",
		Content:   resp.Content,
		ToolCalls: resp.ToolCalls,
	})

	for _, toolCall := range resp.ToolCalls {
		if searchCount >= r.maxSearchPerTurn {
			break
		}

		// 通过 ToolRegistry 查找并执行工具
		tool, err := r.toolRegistry.Get(toolCall.Function.Name)
		if err != nil {
			log.Printf("[AdaptiveRAG] 未知工具: %s", toolCall.Function.Name)
			continue
		}

		searchCount++
		toolsUsed = append(toolsUsed, toolCall.Function.Name)

		// 执行工具
		result, err := tool.Execute(toolCall.Function.Arguments)
		if err != nil {
			log.Printf("[AdaptiveRAG] 执行工具 %s 失败: %v", toolCall.Function.Name, err)
			result = fmt.Sprintf("工具调用失败: %v", err)
		}

		// 将工具结果添加到消息中（包含 tool_call_id 关联）
		messages = append(messages, ChatMessage{
			Role:       "tool",
			Content:    result,
			ToolCallID: toolCall.ID,
		})
	}

	// 如果有工具被调用，再次请求 LLM 生成最终回复
	if searchCount > 0 {
		finalResp, err := r.llmClient.Chat(messages)
		if err != nil {
			return "", toolsUsed, fmt.Errorf("生成最终回复失败: %w", err)
		}
		return finalResp.Content, toolsUsed, nil
	}

	return resp.Content, toolsUsed, nil
}

// mockWebSearch Mock 搜索实现（保留向后兼容，内部委托给 WebSearch 工具）
// Deprecated: 请使用 WebSearch.Execute
func mockWebSearch(query string) string {
	args, _ := json.Marshal(WebSearchArgs{Query: query})
	ws := NewWebSearch()
	result, err := ws.Execute(string(args))
	if err != nil {
		return fmt.Sprintf("搜索失败: %v", err)
	}
	return result
}

// ToolCall LLM 返回的工具调用
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall 函数调用详情
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ProcessToolCalls 仅执行工具调用阶段，返回增强后的消息列表和使用的工具
// 用于流式对话场景：先执行工具调用，再流式生成最终回复
func (r *AdaptiveRAG) ProcessToolCalls(messages []ChatMessage) ([]ChatMessage, []string, error) {
	if !r.enabled {
		return messages, []string{}, nil
	}

	toolDefs := r.toolRegistry.GetAllDefinitions()
	resp, err := r.llmClient.ChatWithTools(messages, toolDefs)
	if err != nil {
		log.Printf("[AdaptiveRAG] ChatWithTools 失败，跳过工具调用: %v", err)
		return messages, []string{}, nil
	}

	if len(resp.ToolCalls) == 0 {
		return messages, []string{}, nil
	}

	// 执行工具调用
	var toolsUsed []string
	searchCount := 0

	messages = append(messages, ChatMessage{
		Role:      "assistant",
		Content:   resp.Content,
		ToolCalls: resp.ToolCalls,
	})

	for _, toolCall := range resp.ToolCalls {
		if searchCount >= r.maxSearchPerTurn {
			break
		}

		tool, err := r.toolRegistry.Get(toolCall.Function.Name)
		if err != nil {
			continue
		}

		searchCount++
		toolsUsed = append(toolsUsed, toolCall.Function.Name)

		result, err := tool.Execute(toolCall.Function.Arguments)
		if err != nil {
			result = fmt.Sprintf("工具调用失败: %v", err)
		}

		messages = append(messages, ChatMessage{
			Role:       "tool",
			Content:    result,
			ToolCallID: toolCall.ID,
		})
	}

	return messages, toolsUsed, nil
}

// ChatResponseWithTools 带工具调用的 LLM 响应
type ChatResponseWithTools struct {
	Content          string     `json:"content"`
	ToolCalls        []ToolCall `json:"tool_calls"`
	PromptTokens     int        `json:"prompt_tokens"`
	CompletionTokens int        `json:"completion_tokens"`
	TotalTokens      int        `json:"total_tokens"`
}
