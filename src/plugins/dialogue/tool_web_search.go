package dialogue

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// WebSearch web_search 工具实现
type WebSearch struct{}

// NewWebSearch 创建 web_search 工具
func NewWebSearch() *WebSearch {
	return &WebSearch{}
}

// Name 返回工具名称
func (w *WebSearch) Name() string {
	return "web_search"
}

// Definition 返回 OpenAI Function Calling 格式的工具定义
func (w *WebSearch) Definition() map[string]interface{} {
	return map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "web_search",
			"description": "当用户询问的内容涉及最新时事、前沿研究、实时数据或知识库中没有的信息时，使用此工具进行网络搜索。不要对基础教材知识使用此工具。",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "搜索查询关键词",
					},
				},
				"required": []string{"query"},
			},
		},
	}
}

// WebSearchArgs web_search 工具参数
type WebSearchArgs struct {
	Query string `json:"query"`
}

// WebSearchResult 单条搜索结果
type WebSearchResult struct {
	Title   string `json:"title"`   // 标题
	Snippet string `json:"snippet"` // 摘要
	URL     string `json:"url"`     // 链接
}

// Execute 执行搜索（当前为 Mock 实现）
func (w *WebSearch) Execute(arguments string) (string, error) {
	var args WebSearchArgs
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("解析搜索参数失败: %w", err)
	}

	if args.Query == "" {
		return "", fmt.Errorf("搜索查询不能为空")
	}

	log.Printf("[WebSearch] Mock 搜索: %s", args.Query)

	// Mock 实现：返回模拟搜索结果，最多 3 条
	results := []WebSearchResult{
		{
			Title:   fmt.Sprintf("关于「%s」的最新研究", args.Query),
			Snippet: fmt.Sprintf("根据最新的研究资料，关于%s的相关信息如下：这是一个模拟的搜索结果，实际部署时将接入真实搜索API。", args.Query),
			URL:     "https://example.com/search?q=" + args.Query,
		},
		{
			Title:   fmt.Sprintf("「%s」百科知识", args.Query),
			Snippet: fmt.Sprintf("%s是一个重要的知识领域。这里提供了基本的背景信息和最新进展。（模拟结果）", args.Query),
			URL:     "https://example.com/wiki/" + args.Query,
		},
		{
			Title:   fmt.Sprintf("「%s」学术论文精选", args.Query),
			Snippet: fmt.Sprintf("以下是与%s相关的最新学术论文摘要和研究方向概述。（模拟结果）", args.Query),
			URL:     "https://example.com/papers/" + args.Query,
		},
	}

	return FormatSearchResults(results), nil
}

// FormatSearchResults 格式化搜索结果为文本
func FormatSearchResults(results []WebSearchResult) string {
	var parts []string
	for i, r := range results {
		parts = append(parts, fmt.Sprintf("[搜索结果%d] %s\n%s\n来源: %s", i+1, r.Title, r.Snippet, r.URL))
	}
	return strings.Join(parts, "\n\n")
}
