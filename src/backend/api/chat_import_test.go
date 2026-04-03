package api

import (
	"strings"
	"testing"
)

// ======================== parseChatJSON 测试 ========================

// TestParseChatJSON_Format1_OpenAI OpenAI messages 格式
func TestParseChatJSON_Format1_OpenAI(t *testing.T) {
	data := []byte(`{
		"messages": [
			{"role": "user", "content": "什么是二叉树？"},
			{"role": "assistant", "content": "二叉树是一种树形数据结构"}
		]
	}`)

	messages, err := parseChatJSON(data)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("期望2条消息, 实际=%d", len(messages))
	}
	if messages[0].Role != "user" {
		t.Fatalf("第1条消息角色应为 user, 实际=%s", messages[0].Role)
	}
	if messages[1].Role != "assistant" {
		t.Fatalf("第2条消息角色应为 assistant, 实际=%s", messages[1].Role)
	}
}

// TestParseChatJSON_Format2_Conversations conversations 格式
func TestParseChatJSON_Format2_Conversations(t *testing.T) {
	data := []byte(`{
		"conversations": [
			{"sender": "学生", "text": "什么是递归？"},
			{"sender": "AI", "text": "递归是函数调用自身的过程"}
		]
	}`)

	messages, err := parseChatJSON(data)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("期望2条消息, 实际=%d", len(messages))
	}
	if messages[0].Role != "user" {
		t.Fatalf("'学生'应映射为 user, 实际=%s", messages[0].Role)
	}
	if messages[1].Role != "assistant" {
		t.Fatalf("'AI'应映射为 assistant, 实际=%s", messages[1].Role)
	}
}

// TestParseChatJSON_Format3_TopLevelArray 顶层数组格式
func TestParseChatJSON_Format3_TopLevelArray(t *testing.T) {
	data := []byte(`[
		{"role": "user", "content": "你好"},
		{"role": "assistant", "content": "你好！有什么可以帮助你的？"}
	]`)

	messages, err := parseChatJSON(data)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("期望2条消息, 实际=%d", len(messages))
	}
}

// TestParseChatJSON_Format3_TopLevelArray_Mixed 顶层数组混合格式（sender/text）
func TestParseChatJSON_Format3_TopLevelArray_Mixed(t *testing.T) {
	data := []byte(`[
		{"sender": "学生", "text": "什么是链表？"},
		{"sender": "AI", "text": "链表是一种线性数据结构"}
	]`)

	messages, err := parseChatJSON(data)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("期望2条消息, 实际=%d", len(messages))
	}
	if messages[0].Content != "什么是链表？" {
		t.Fatalf("第1条消息内容不匹配, 实际=%s", messages[0].Content)
	}
}

// TestParseChatJSON_InvalidJSON 无效 JSON
func TestParseChatJSON_InvalidJSON(t *testing.T) {
	data := []byte(`not a json`)
	_, err := parseChatJSON(data)
	if err == nil {
		t.Fatal("无效 JSON 应返回错误")
	}
}

// TestParseChatJSON_EmptyMessages 空消息列表
func TestParseChatJSON_EmptyMessages(t *testing.T) {
	data := []byte(`{"messages": []}`)
	messages, err := parseChatJSON(data)
	if err == nil && len(messages) > 0 {
		t.Fatal("空消息列表应返回错误或空结果")
	}
}

// TestParseChatJSON_EmptyConversations 空 conversations 列表
func TestParseChatJSON_EmptyConversations(t *testing.T) {
	data := []byte(`{"conversations": []}`)
	messages, err := parseChatJSON(data)
	if err == nil && len(messages) > 0 {
		t.Fatal("空 conversations 列表应返回错误或空结果")
	}
}

// TestParseChatJSON_EmptyArray 空顶层数组
func TestParseChatJSON_EmptyArray(t *testing.T) {
	data := []byte(`[]`)
	messages, err := parseChatJSON(data)
	if err == nil && len(messages) > 0 {
		t.Fatal("空数组应返回错误或空结果")
	}
}

// TestParseChatJSON_EmptyContent 消息内容为空时应被过滤
func TestParseChatJSON_EmptyContent(t *testing.T) {
	data := []byte(`{
		"messages": [
			{"role": "user", "content": ""},
			{"role": "assistant", "content": "有效回复"},
			{"role": "user", "content": "有效问题"},
			{"role": "assistant", "content": ""}
		]
	}`)

	messages, err := parseChatJSON(data)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	// 空内容的消息应被过滤
	if len(messages) != 2 {
		t.Fatalf("期望2条有效消息, 实际=%d", len(messages))
	}
}

// TestParseChatJSON_UnrecognizedFormat 无法识别的格式
func TestParseChatJSON_UnrecognizedFormat(t *testing.T) {
	data := []byte(`{"foo": "bar"}`)
	_, err := parseChatJSON(data)
	if err == nil {
		t.Fatal("无法识别的格式应返回错误")
	}
}

// TestParseChatJSON_MultipleConversations 多轮对话
func TestParseChatJSON_MultipleConversations(t *testing.T) {
	data := []byte(`{
		"messages": [
			{"role": "user", "content": "问题1"},
			{"role": "assistant", "content": "回答1"},
			{"role": "user", "content": "问题2"},
			{"role": "assistant", "content": "回答2"},
			{"role": "user", "content": "问题3"},
			{"role": "assistant", "content": "回答3"}
		]
	}`)

	messages, err := parseChatJSON(data)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(messages) != 6 {
		t.Fatalf("期望6条消息, 实际=%d", len(messages))
	}
}

// ======================== buildQAContent 测试 ========================

// TestBuildQAContent 测试 Q&A 内容构建
func TestBuildQAContent(t *testing.T) {
	messages := []chatMessage{
		{Role: "user", Content: "什么是二叉树？"},
		{Role: "assistant", Content: "二叉树是一种树形数据结构"},
		{Role: "user", Content: "有哪些遍历方式？"},
		{Role: "assistant", Content: "前序、中序、后序遍历"},
	}

	content := buildQAContent(messages)

	if content == "" {
		t.Fatal("内容不应为空")
	}

	// 验证包含 Q: 和 A: 格式
	if !strings.Contains(content, "Q: 什么是二叉树？") {
		t.Error("内容应包含第一个问题")
	}
	if !strings.Contains(content, "A: 二叉树是一种树形数据结构") {
		t.Error("内容应包含第一个回答")
	}
	if !strings.Contains(content, "Q: 有哪些遍历方式？") {
		t.Error("内容应包含第二个问题")
	}
	if !strings.Contains(content, "A: 前序、中序、后序遍历") {
		t.Error("内容应包含第二个回答")
	}
}

// TestBuildQAContent_ConsecutiveUserMessages 连续用户消息
func TestBuildQAContent_ConsecutiveUserMessages(t *testing.T) {
	messages := []chatMessage{
		{Role: "user", Content: "第一部分"},
		{Role: "user", Content: "第二部分"},
		{Role: "assistant", Content: "回答"},
	}

	content := buildQAContent(messages)

	if !strings.Contains(content, "Q: 第一部分") {
		t.Error("应包含第一部分")
	}
	if !strings.Contains(content, "第二部分") {
		t.Error("应包含第二部分")
	}
	if !strings.Contains(content, "A: 回答") {
		t.Error("应包含回答")
	}
}

// TestBuildQAContent_ConsecutiveAssistantMessages 连续助手消息
func TestBuildQAContent_ConsecutiveAssistantMessages(t *testing.T) {
	messages := []chatMessage{
		{Role: "user", Content: "问题"},
		{Role: "assistant", Content: "回答第一段"},
		{Role: "assistant", Content: "回答第二段"},
	}

	content := buildQAContent(messages)

	if !strings.Contains(content, "Q: 问题") {
		t.Error("应包含问题")
	}
	if !strings.Contains(content, "A: 回答第一段") {
		t.Error("应包含回答第一段")
	}
	if !strings.Contains(content, "回答第二段") {
		t.Error("应包含回答第二段")
	}
}

// TestBuildQAContent_Empty 空消息列表
func TestBuildQAContent_Empty(t *testing.T) {
	content := buildQAContent([]chatMessage{})
	if content != "" {
		t.Errorf("空消息列表应返回空字符串, 实际=%q", content)
	}
}

// ======================== extractQAPairs 测试 ========================

// TestExtractQAPairs 测试问答对提取
func TestExtractQAPairs(t *testing.T) {
	messages := []chatMessage{
		{Role: "user", Content: "什么是二叉树？"},
		{Role: "assistant", Content: "二叉树是一种树形数据结构"},
		{Role: "user", Content: "有哪些遍历方式？"},
		{Role: "assistant", Content: "前序、中序、后序遍历"},
	}

	pairs := extractQAPairs(messages)
	if len(pairs) != 2 {
		t.Fatalf("期望2个问答对, 实际=%d", len(pairs))
	}
	if pairs[0].Question != "什么是二叉树？" {
		t.Errorf("第1个问答对问题不匹配: %s", pairs[0].Question)
	}
	if pairs[0].Answer != "二叉树是一种树形数据结构" {
		t.Errorf("第1个问答对回答不匹配: %s", pairs[0].Answer)
	}
	if pairs[1].Question != "有哪些遍历方式？" {
		t.Errorf("第2个问答对问题不匹配: %s", pairs[1].Question)
	}
}

// TestExtractQAPairs_NoPairs 无法配对的消息
func TestExtractQAPairs_NoPairs(t *testing.T) {
	messages := []chatMessage{
		{Role: "user", Content: "问题1"},
		{Role: "user", Content: "问题2"},
	}

	pairs := extractQAPairs(messages)
	if len(pairs) != 0 {
		t.Fatalf("连续用户消息不应产生问答对, 实际=%d", len(pairs))
	}
}

// TestExtractQAPairs_OnlyAssistant 只有助手消息
func TestExtractQAPairs_OnlyAssistant(t *testing.T) {
	messages := []chatMessage{
		{Role: "assistant", Content: "回答1"},
		{Role: "assistant", Content: "回答2"},
	}

	pairs := extractQAPairs(messages)
	if len(pairs) != 0 {
		t.Fatalf("只有助手消息不应产生问答对, 实际=%d", len(pairs))
	}
}

// ======================== normalizeRole / normalizeSender 测试 ========================

// TestNormalizeRole 测试角色标准化
func TestNormalizeRole(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user", "user"},
		{"User", "user"},
		{"human", "user"},
		{"student", "user"},
		{"学生", "user"},
		{"assistant", "assistant"},
		{"AI", "assistant"},
		{"bot", "assistant"},
		{"teacher", "assistant"},
		{"老师", "assistant"},
	}

	for _, tt := range tests {
		result := normalizeRole(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeRole(%q) = %q, 期望 %q", tt.input, result, tt.expected)
		}
	}
}

// TestNormalizeSender 测试发送者标准化
func TestNormalizeSender(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"学生", "user"},
		{"我", "user"},
		{"AI", "assistant"},
		{"老师", "assistant"},
		{"unknown", "user"}, // 默认 user
	}

	for _, tt := range tests {
		result := normalizeSender(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeSender(%q) = %q, 期望 %q", tt.input, result, tt.expected)
		}
	}
}

// ======================== ParseChatJSON 导出函数测试 ========================

// TestParseChatJSON_Exported 测试导出的 ParseChatJSON 函数
func TestParseChatJSON_Exported(t *testing.T) {
	data := []byte(`{
		"messages": [
			{"role": "user", "content": "什么是二叉树？"},
			{"role": "assistant", "content": "二叉树是一种树形数据结构"},
			{"role": "user", "content": "有哪些遍历方式？"},
			{"role": "assistant", "content": "前序、中序、后序遍历"}
		]
	}`)

	pairs, err := ParseChatJSON(data)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(pairs) != 2 {
		t.Fatalf("期望2个问答对, 实际=%d", len(pairs))
	}
	if pairs[0].Question != "什么是二叉树？" {
		t.Errorf("第1个问答对问题不匹配: %s", pairs[0].Question)
	}
	if pairs[0].Answer != "二叉树是一种树形数据结构" {
		t.Errorf("第1个问答对回答不匹配: %s", pairs[0].Answer)
	}
}

// TestParseChatJSON_Exported_InvalidJSON 导出函数处理无效 JSON
func TestParseChatJSON_Exported_InvalidJSON(t *testing.T) {
	_, err := ParseChatJSON([]byte(`invalid`))
	if err == nil {
		t.Fatal("无效 JSON 应返回错误")
	}
}
