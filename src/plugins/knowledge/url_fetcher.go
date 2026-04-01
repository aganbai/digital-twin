package knowledge

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

// URLFetcher URL 抓取器
type URLFetcher struct{}

// NewURLFetcher 创建 URL 抓取器
func NewURLFetcher() *URLFetcher {
	return &URLFetcher{}
}

// maxContentLength 内容最大字符数
const maxContentLength = 100000

// Fetch 抓取 URL 内容，返回标题和正文
func (f *URLFetcher) Fetch(url string) (title string, content string, err error) {
	// Mock 模式：直接返回模拟数据，不进行真实 HTTP 请求
	if os.Getenv("WX_MODE") == "mock" {
		mockTitle := "Mock 文档 - " + url
		mockContent := "这是通过 URL 导入的模拟文档内容。\n\n" +
			"URL: " + url + "\n\n" +
			"Python 是一种高级编程语言，广泛用于数据科学、人工智能和 Web 开发。\n" +
			"Go 是 Google 开发的编程语言，以其并发性能和简洁语法著称。\n" +
			"JavaScript 是 Web 开发的核心语言，支持前端和后端开发。"
		return mockTitle, mockContent, nil
	}

	// 创建带超时的 HTTP 客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 创建请求并设置 User-Agent
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "DigitalTwin/2.0")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("请求 URL 失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("HTTP 状态码异常: %d", resp.StatusCode)
	}

	// 编码检测：使用 charset 包根据 Content-Type 自动转换为 UTF-8
	contentType := resp.Header.Get("Content-Type")
	reader, err := charset.NewReader(resp.Body, contentType)
	if err != nil {
		return "", "", fmt.Errorf("编码转换失败: %w", err)
	}

	// 读取全部内容
	bodyBytes, err := io.ReadAll(reader)
	if err != nil {
		return "", "", fmt.Errorf("读取响应内容失败: %w", err)
	}

	// 解析 HTML
	doc, err := html.Parse(strings.NewReader(string(bodyBytes)))
	if err != nil {
		return "", "", fmt.Errorf("解析 HTML 失败: %w", err)
	}

	// 提取标题
	title = extractTitle(doc)

	// 提取正文内容
	content = extractBodyText(doc)

	// 检查内容是否为空
	if strings.TrimSpace(content) == "" {
		return "", "", fmt.Errorf("页面内容为空")
	}

	// 内容截断
	runes := []rune(content)
	if len(runes) > maxContentLength {
		content = string(runes[:maxContentLength])
	}

	return title, content, nil
}

// skipTags 需要跳过的非正文标签
var skipTags = map[string]bool{
	"script":   true,
	"style":    true,
	"nav":      true,
	"footer":   true,
	"header":   true,
	"noscript": true,
	"iframe":   true,
	"svg":      true,
}

// getTextContent 获取节点及其所有子节点的文本内容
func getTextContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(getTextContent(c))
	}
	return sb.String()
}

// extractTitle 从 DOM 树中提取 <title> 标签内容
func extractTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" {
		return getTextContent(n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if t := extractTitle(c); t != "" {
			return t
		}
	}
	return ""
}

// extractBodyText 从 DOM 树中提取 <body> 内的文本内容
func extractBodyText(n *html.Node) string {
	// 先找到 body 节点
	body := findBody(n)
	if body == nil {
		// 如果没有 body 标签，直接从根节点提取
		body = n
	}

	var sb strings.Builder
	collectText(body, &sb)
	return strings.TrimSpace(sb.String())
}

// findBody 在 DOM 树中查找 <body> 节点
func findBody(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "body" {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findBody(c); found != nil {
			return found
		}
	}
	return nil
}

// collectText 递归收集节点中的文本内容，跳过非正文标签
func collectText(n *html.Node, sb *strings.Builder) {
	// 跳过非正文标签
	if n.Type == html.ElementNode && skipTags[n.Data] {
		return
	}

	// 文本节点：收集内容
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			if sb.Len() > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(text)
		}
		return
	}

	// 块级元素前后添加换行
	isBlock := isBlockElement(n)
	if isBlock && sb.Len() > 0 {
		sb.WriteString("\n")
	}

	// 递归处理子节点
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectText(c, sb)
	}

	if isBlock && sb.Len() > 0 {
		sb.WriteString("\n")
	}
}

// blockElements 块级元素集合
var blockElements = map[string]bool{
	"div": true, "p": true, "h1": true, "h2": true, "h3": true,
	"h4": true, "h5": true, "h6": true, "article": true, "section": true,
	"aside": true, "main": true, "blockquote": true, "pre": true,
	"ul": true, "ol": true, "li": true, "table": true, "tr": true,
	"br": true, "hr": true, "dd": true, "dt": true, "dl": true,
	"figcaption": true, "figure": true, "address": true,
}

// isBlockElement 判断是否是块级元素
func isBlockElement(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	return blockElements[n.Data]
}
