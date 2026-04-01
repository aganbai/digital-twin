package knowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"digital-twin/src/plugins/dialogue"
)

// LLMSummaryResult LLM 摘要结果
type LLMSummaryResult struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

// LLMSummarizer LLM 摘要服务
type LLMSummarizer struct {
	maxChars int
	enabled  bool
}

// NewLLMSummarizer 创建 LLM 摘要服务
func NewLLMSummarizer() *LLMSummarizer {
	maxChars := 3000
	if v := os.Getenv("LLM_SUMMARY_MAX_CHARS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxChars = n
		}
	}

	enabled := true
	if v := os.Getenv("LLM_SUMMARY_ENABLED"); v == "false" {
		enabled = false
	}

	return &LLMSummarizer{
		maxChars: maxChars,
		enabled:  enabled,
	}
}

// Summarize 调用 LLM 生成标题和摘要
func (s *LLMSummarizer) Summarize(ctx context.Context, content string) (*LLMSummaryResult, error) {
	if !s.enabled {
		return &LLMSummaryResult{}, nil
	}

	// 截取前 maxChars 个字符
	runes := []rune(content)
	if len(runes) > s.maxChars {
		runes = runes[:s.maxChars]
	}
	truncatedContent := string(runes)

	// 构建 prompt
	prompt := `你是一个文档分析助手。请根据以下文档内容，生成：
1. 一个简洁准确的标题（不超过50字）
2. 一段内容摘要（不超过200字，概括文档的核心内容和要点）

请以 JSON 格式返回：
{"title": "...", "summary": "..."}

文档内容：
` + truncatedContent

	// 从环境变量读取 LLM 配置
	mode := os.Getenv("LLM_MODE")
	if mode == "" {
		mode = "mock"
	}
	model := os.Getenv("LLM_MODEL")
	if model == "" {
		model = "qwen-turbo"
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	}

	// 创建 LLM 客户端
	client := dialogue.NewLLMClient(mode, model, apiKey, baseURL, 0.7, 1000)

	// 设置超时
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 使用 channel 处理超时
	type result struct {
		resp *dialogue.ChatResponse
		err  error
	}
	ch := make(chan result, 1)

	go func() {
		messages := []dialogue.ChatMessage{
			{Role: "user", Content: prompt},
		}
		resp, err := client.Chat(messages)
		ch <- result{resp, err}
	}()

	select {
	case <-timeoutCtx.Done():
		return &LLMSummaryResult{}, fmt.Errorf("LLM 摘要生成超时")
	case r := <-ch:
		if r.err != nil {
			return &LLMSummaryResult{}, fmt.Errorf("LLM 调用失败: %w", r.err)
		}

		// 解析 JSON 响应
		var summaryResult LLMSummaryResult
		if err := json.Unmarshal([]byte(r.resp.Content), &summaryResult); err != nil {
			// 尝试从响应中提取 JSON
			jsonStr := extractJSON(r.resp.Content)
			if jsonStr != "" {
				if err2 := json.Unmarshal([]byte(jsonStr), &summaryResult); err2 == nil {
					return &summaryResult, nil
				}
			}
			return &LLMSummaryResult{}, fmt.Errorf("解析 LLM 响应失败: %w", err)
		}

		return &summaryResult, nil
	}
}

// extractJSON 从文本中提取 JSON 字符串
func extractJSON(text string) string {
	// 查找第一个 { 和最后一个 }
	start := -1
	end := -1
	for i, r := range text {
		if r == '{' && start == -1 {
			start = i
		}
		if r == '}' {
			end = i
		}
	}
	if start >= 0 && end > start {
		return text[start : end+1]
	}
	return ""
}
