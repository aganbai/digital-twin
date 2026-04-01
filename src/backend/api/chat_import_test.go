package api

import (
	"testing"
)

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
	if !contains(content, "Q: 什么是二叉树？") {
		t.Error("内容应包含第一个问题")
	}
	if !contains(content, "A: 二叉树是一种树形数据结构") {
		t.Error("内容应包含第一个回答")
	}
	if !contains(content, "Q: 有哪些遍历方式？") {
		t.Error("内容应包含第二个问题")
	}
}

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

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
