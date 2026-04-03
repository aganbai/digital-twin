package dialogue

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ChatMessage 聊天消息
type ChatMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`   // assistant 消息中的工具调用
	ToolCallID string     `json:"tool_call_id,omitempty"` // tool 消息中关联的工具调用ID
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Content          string `json:"content"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	TotalTokens      int    `json:"total_tokens"`
}

// LLMClient 大模型客户端
type LLMClient struct {
	mode               string
	model              string
	apiKey             string
	baseURL            string
	temperature        float64
	defaultTemperature float64
	maxTokens          int
	httpClient         *http.Client
	mu                 sync.Mutex
}

// NewLLMClient 创建大模型客户端
func NewLLMClient(mode, model, apiKey, baseURL string, temperature float64, maxTokens int) *LLMClient {
	return &LLMClient{
		mode:               mode,
		model:              model,
		apiKey:             apiKey,
		baseURL:            baseURL,
		temperature:        temperature,
		defaultTemperature: temperature,
		maxTokens:          maxTokens,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// SetTemperature 临时设置 temperature（用于个性化风格），并发安全
func (c *LLMClient) SetTemperature(temp float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.temperature = temp
}

// ResetTemperature 恢复默认 temperature，并发安全
func (c *LLMClient) ResetTemperature() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.temperature = c.defaultTemperature
}

// Chat 调用大模型进行对话
func (c *LLMClient) Chat(messages []ChatMessage) (*ChatResponse, error) {
	switch c.mode {
	case "api":
		return c.chatAPI(messages)
	case "mock":
		return c.chatMock(messages)
	default:
		return c.chatMock(messages)
	}
}

// ChatStream 流式调用大模型
// onDelta 回调函数，每次收到一个文本片段时调用
// 返回完整的 ChatResponse（包含完整回复和 token 统计）
func (c *LLMClient) ChatStream(messages []ChatMessage, onDelta func(content string)) (*ChatResponse, error) {
	switch c.mode {
	case "api":
		return c.chatStreamAPI(messages, onDelta)
	case "mock":
		return c.chatStreamMock(messages, onDelta)
	default:
		return c.chatStreamMock(messages, onDelta)
	}
}

// ChatWithTools 带工具定义的对话调用（支持 Function Calling）
func (c *LLMClient) ChatWithTools(messages []ChatMessage, tools []map[string]interface{}) (*ChatResponseWithTools, error) {
	switch c.mode {
	case "api":
		return c.chatWithToolsAPI(messages, tools)
	case "mock":
		return c.chatWithToolsMock(messages, tools)
	default:
		return c.chatWithToolsMock(messages, tools)
	}
}

// chatWithToolsAPI 调用 API（带 tools 参数）
func (c *LLMClient) chatWithToolsAPI(messages []ChatMessage, tools []map[string]interface{}) (*ChatResponseWithTools, error) {
	reqBody := map[string]interface{}{
		"model":       c.model,
		"messages":    messages,
		"temperature": c.temperature,
		"max_tokens":  c.maxTokens,
		"tools":       tools,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("调用 LLM API 失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LLM API 返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	// 解析包含 tool_calls 的响应
	var apiResp struct {
		Choices []struct {
			Message struct {
				Content   string     `json:"content"`
				ToolCalls []ToolCall `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("LLM API 返回空的 choices")
	}

	return &ChatResponseWithTools{
		Content:          apiResp.Choices[0].Message.Content,
		ToolCalls:        apiResp.Choices[0].Message.ToolCalls,
		PromptTokens:     apiResp.Usage.PromptTokens,
		CompletionTokens: apiResp.Usage.CompletionTokens,
		TotalTokens:      apiResp.Usage.TotalTokens,
	}, nil
}

// chatWithToolsMock Mock 模式的工具调用（不触发工具调用，直接返回）
func (c *LLMClient) chatWithToolsMock(messages []ChatMessage, tools []map[string]interface{}) (*ChatResponseWithTools, error) {
	resp, err := c.chatMock(messages)
	if err != nil {
		return nil, err
	}
	return &ChatResponseWithTools{
		Content:          resp.Content,
		ToolCalls:        nil, // Mock 模式不触发工具调用
		PromptTokens:     resp.PromptTokens,
		CompletionTokens: resp.CompletionTokens,
		TotalTokens:      resp.TotalTokens,
	}, nil
}

// apiRequest API 请求体
type apiRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

// apiResponse API 响应体
type apiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// chatAPI 调用 OpenAI 兼容 API
func (c *LLMClient) chatAPI(messages []ChatMessage) (*ChatResponse, error) {
	reqBody := apiRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: c.temperature,
		MaxTokens:   c.maxTokens,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("调用 LLM API 失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LLM API 返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("LLM API 返回空的 choices")
	}

	return &ChatResponse{
		Content:          apiResp.Choices[0].Message.Content,
		PromptTokens:     apiResp.Usage.PromptTokens,
		CompletionTokens: apiResp.Usage.CompletionTokens,
		TotalTokens:      apiResp.Usage.TotalTokens,
	}, nil
}

// chatMock 模拟回复（用于测试）
func (c *LLMClient) chatMock(messages []ChatMessage) (*ChatResponse, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("消息列表不能为空")
	}

	// 获取最后一条用户消息
	var userMessage string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			userMessage = messages[i].Content
			break
		}
	}

	// 生成苏格拉底式回复
	reply := fmt.Sprintf("这是一个很好的问题！关于「%s」，让我反过来问你：你觉得这个问题的关键点在哪里？你能尝试从不同角度来思考一下吗？", userMessage)

	// 模拟 token 计数
	promptTokens := 0
	for _, msg := range messages {
		promptTokens += len([]rune(msg.Content)) / 2
	}
	completionTokens := len([]rune(reply)) / 2

	return &ChatResponse{
		Content:          reply,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
	}, nil
}

// MemoryExtraction 记忆提取结果
type MemoryExtraction struct {
	Type       string  `json:"type"`
	Content    string  `json:"content"`
	Importance float64 `json:"importance"`
}

// ExtractMemories 调用 LLM 提取记忆
func (c *LLMClient) ExtractMemories(messages []ChatMessage) ([]MemoryExtraction, error) {
	switch c.mode {
	case "api":
		return c.extractMemoriesAPI(messages)
	case "mock":
		return c.extractMemoriesMock(messages)
	default:
		return c.extractMemoriesMock(messages)
	}
}

// extractMemoriesAPI 调用 API 提取记忆
func (c *LLMClient) extractMemoriesAPI(messages []ChatMessage) ([]MemoryExtraction, error) {
	resp, err := c.chatAPI(messages)
	if err != nil {
		return nil, fmt.Errorf("调用 LLM 提取记忆失败: %w", err)
	}

	// 解析 JSON 响应
	var memories []MemoryExtraction
	if err := json.Unmarshal([]byte(resp.Content), &memories); err != nil {
		// 如果解析失败，尝试提取 JSON 部分
		content := resp.Content
		startIdx := strings.Index(content, "[")
		endIdx := strings.LastIndex(content, "]")
		if startIdx >= 0 && endIdx > startIdx {
			jsonStr := content[startIdx : endIdx+1]
			if err2 := json.Unmarshal([]byte(jsonStr), &memories); err2 != nil {
				return nil, fmt.Errorf("解析记忆提取结果失败: %w", err2)
			}
		} else {
			return nil, fmt.Errorf("LLM 返回的内容不包含有效的 JSON 数组: %s", content)
		}
	}

	return memories, nil
}

// extractMemoriesMock Mock 模式提取记忆
func (c *LLMClient) extractMemoriesMock(messages []ChatMessage) ([]MemoryExtraction, error) {
	return []MemoryExtraction{
		{
			Type:       "conversation",
			Content:    "学生进行了一次对话学习",
			Importance: 0.5,
		},
	}, nil
}

// streamAPIRequest 流式 API 请求体
type streamAPIRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
	Stream      bool          `json:"stream"`
}

// streamDeltaResponse SSE 流式响应中每个 data 行的结构
type streamDeltaResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		Index        int     `json:"index"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// chatStreamAPI 流式调用 OpenAI 兼容 API
func (c *LLMClient) chatStreamAPI(messages []ChatMessage, onDelta func(content string)) (*ChatResponse, error) {
	reqBody := streamAPIRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: c.temperature,
		MaxTokens:   c.maxTokens,
		Stream:      true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	// 流式请求使用更长的超时时间
	streamClient := &http.Client{
		Timeout: 120 * time.Second,
	}

	resp, err := streamClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("调用 LLM API 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LLM API 返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	// 逐行读取 SSE 流
	scanner := bufio.NewScanner(resp.Body)
	var fullContent strings.Builder
	var lastUsage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	}

	for scanner.Scan() {
		line := scanner.Text()

		// 跳过空行
		if line == "" {
			continue
		}

		// 检查 data: 前缀
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		// 检查流结束标记
		if data == "[DONE]" {
			break
		}

		// 解析 JSON
		var deltaResp streamDeltaResponse
		if err := json.Unmarshal([]byte(data), &deltaResp); err != nil {
			// 跳过无法解析的行
			continue
		}

		// 提取 delta content
		if len(deltaResp.Choices) > 0 {
			content := deltaResp.Choices[0].Delta.Content
			if content != "" {
				fullContent.WriteString(content)
				onDelta(content)
			}
		}

		// 记录最后的 usage 信息（部分 API 在最后一个 chunk 返回 usage）
		if deltaResp.Usage != nil {
			lastUsage = deltaResp.Usage
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取 SSE 流失败: %w", err)
	}

	// 构造响应
	chatResp := &ChatResponse{
		Content: fullContent.String(),
	}

	if lastUsage != nil {
		chatResp.PromptTokens = lastUsage.PromptTokens
		chatResp.CompletionTokens = lastUsage.CompletionTokens
		chatResp.TotalTokens = lastUsage.TotalTokens
	} else {
		// 如果 API 没有返回 usage，估算 token 数
		promptTokens := 0
		for _, msg := range messages {
			promptTokens += len([]rune(msg.Content)) / 2
		}
		completionTokens := len([]rune(chatResp.Content)) / 2
		chatResp.PromptTokens = promptTokens
		chatResp.CompletionTokens = completionTokens
		chatResp.TotalTokens = promptTokens + completionTokens
	}

	return chatResp, nil
}

// chatStreamMock 模拟流式回复（用于测试）
func (c *LLMClient) chatStreamMock(messages []ChatMessage, onDelta func(content string)) (*ChatResponse, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("消息列表不能为空")
	}

	// 获取最后一条用户消息
	var userMessage string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			userMessage = messages[i].Content
			break
		}
	}

	// 生成苏格拉底式回复
	reply := fmt.Sprintf("这是一个很好的问题！关于「%s」，让我反过来问你：你觉得这个问题的关键点在哪里？你能尝试从不同角度来思考一下吗？", userMessage)

	// 按字符逐个输出
	for _, char := range reply {
		onDelta(string(char))
		time.Sleep(50 * time.Millisecond)
	}

	// 模拟 token 计数
	promptTokens := 0
	for _, msg := range messages {
		promptTokens += len([]rune(msg.Content)) / 2
	}
	completionTokens := len([]rune(reply)) / 2

	return &ChatResponse{
		Content:          reply,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
	}, nil
}
