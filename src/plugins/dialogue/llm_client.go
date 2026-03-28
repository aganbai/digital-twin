package dialogue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ChatMessage 聊天消息
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
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
	mode        string
	model       string
	apiKey      string
	baseURL     string
	temperature float64
	maxTokens   int
	httpClient  *http.Client
}

// NewLLMClient 创建大模型客户端
func NewLLMClient(mode, model, apiKey, baseURL string, temperature float64, maxTokens int) *LLMClient {
	return &LLMClient{
		mode:        mode,
		model:       model,
		apiKey:      apiKey,
		baseURL:     baseURL,
		temperature: temperature,
		maxTokens:   maxTokens,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
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
