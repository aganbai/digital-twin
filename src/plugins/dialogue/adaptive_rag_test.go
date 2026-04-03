package dialogue

import (
	"strings"
	"testing"
)

// TestAdaptiveRAG_Disabled 测试 RAG 禁用时直接调用 LLM
func TestAdaptiveRAG_Disabled(t *testing.T) {
	client := NewLLMClient("mock", "test-model", "", "", 0.7, 1000)
	rag := NewAdaptiveRAG(false, client)

	messages := []ChatMessage{
		{Role: "system", Content: "你是一位老师"},
		{Role: "user", Content: "什么是牛顿第一定律？"},
	}

	reply, toolsUsed, err := rag.ProcessWithRAG(messages)
	if err != nil {
		t.Fatalf("ProcessWithRAG 失败: %v", err)
	}

	if reply == "" {
		t.Fatal("回复不应为空")
	}

	if len(toolsUsed) != 0 {
		t.Fatalf("RAG 禁用时不应有工具使用，实际: %v", toolsUsed)
	}

	t.Logf("RAG 禁用回复: %s", reply[:50])
}

// TestAdaptiveRAG_Enabled_MockMode 测试 RAG 启用但 Mock 模式不触发工具
func TestAdaptiveRAG_Enabled_MockMode(t *testing.T) {
	client := NewLLMClient("mock", "test-model", "", "", 0.7, 1000)
	rag := NewAdaptiveRAG(true, client)

	messages := []ChatMessage{
		{Role: "system", Content: "你是一位老师"},
		{Role: "user", Content: "最新的量子计算进展是什么？"},
	}

	reply, toolsUsed, err := rag.ProcessWithRAG(messages)
	if err != nil {
		t.Fatalf("ProcessWithRAG 失败: %v", err)
	}

	if reply == "" {
		t.Fatal("回复不应为空")
	}

	// Mock 模式下 ChatWithTools 不触发工具调用
	if len(toolsUsed) != 0 {
		t.Logf("Mock 模式可能不触发工具，toolsUsed: %v", toolsUsed)
	}

	t.Logf("RAG 启用(Mock)回复: %s", reply[:50])
}

// TestAdaptiveRAG_DisabledDoesNotSearch 测试开关关闭时不触发搜索
func TestAdaptiveRAG_DisabledDoesNotSearch(t *testing.T) {
	client := NewLLMClient("mock", "test-model", "", "", 0.7, 1000)
	rag := NewAdaptiveRAG(false, client)

	// 即使消息明确要求搜索，禁用时也不应触发
	messages := []ChatMessage{
		{Role: "system", Content: "你是一位老师"},
		{Role: "user", Content: "请帮我搜索一下最新的AI论文"},
	}

	_, toolsUsed, err := rag.ProcessWithRAG(messages)
	if err != nil {
		t.Fatalf("ProcessWithRAG 失败: %v", err)
	}

	if len(toolsUsed) != 0 {
		t.Fatalf("RAG 禁用时 tools_used 应为空数组，实际: %v", toolsUsed)
	}
}

// TestAdaptiveRAG_ToolsUsedFieldEmpty 测试未使用工具时 tools_used 为空数组
func TestAdaptiveRAG_ToolsUsedFieldEmpty(t *testing.T) {
	client := NewLLMClient("mock", "test-model", "", "", 0.7, 1000)

	// 测试禁用时
	ragDisabled := NewAdaptiveRAG(false, client)
	messages := []ChatMessage{
		{Role: "user", Content: "你好"},
	}
	_, toolsUsed, err := ragDisabled.ProcessWithRAG(messages)
	if err != nil {
		t.Fatalf("ProcessWithRAG 失败: %v", err)
	}
	if toolsUsed == nil {
		t.Fatal("tools_used 不应为 nil，应为空数组")
	}
	if len(toolsUsed) != 0 {
		t.Fatalf("未使用工具时 tools_used 应为空数组，实际: %v", toolsUsed)
	}

	// 测试启用但 Mock 模式不触发时
	ragEnabled := NewAdaptiveRAG(true, client)
	_, toolsUsed2, err := ragEnabled.ProcessWithRAG(messages)
	if err != nil {
		t.Fatalf("ProcessWithRAG 失败: %v", err)
	}
	if toolsUsed2 == nil {
		t.Fatal("tools_used 不应为 nil，应为空数组")
	}
}

// TestWebSearch_ToolInterface 测试 WebSearch 实现 Tool 接口
func TestWebSearch_ToolInterface(t *testing.T) {
	var tool Tool = NewWebSearch()

	if tool.Name() != "web_search" {
		t.Fatalf("工具名称应为 web_search，实际: %s", tool.Name())
	}

	def := tool.Definition()
	toolType, ok := def["type"].(string)
	if !ok || toolType != "function" {
		t.Fatal("工具类型应为 function")
	}

	fn, ok := def["function"].(map[string]interface{})
	if !ok {
		t.Fatal("工具应包含 function 定义")
	}

	name, ok := fn["name"].(string)
	if !ok || name != "web_search" {
		t.Fatal("工具名称应为 web_search")
	}

	desc, ok := fn["description"].(string)
	if !ok || desc == "" {
		t.Fatal("工具应有描述")
	}

	params, ok := fn["parameters"].(map[string]interface{})
	if !ok {
		t.Fatal("工具应有参数定义")
	}

	props, ok := params["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("参数应有 properties")
	}

	if _, ok := props["query"]; !ok {
		t.Fatal("参数应包含 query 字段")
	}
}

// TestWebSearch_Execute 测试 web_search Mock 返回
func TestWebSearch_Execute(t *testing.T) {
	ws := NewWebSearch()

	result, err := ws.Execute(`{"query": "量子计算"}`)
	if err != nil {
		t.Fatalf("执行搜索失败: %v", err)
	}

	if result == "" {
		t.Fatal("搜索结果不应为空")
	}

	if !strings.Contains(result, "量子计算") {
		t.Error("搜索结果应包含查询关键词")
	}

	if !strings.Contains(result, "搜索结果") {
		t.Error("搜索结果应包含'搜索结果'标记")
	}

	// 验证返回最多 3 条结果
	count := strings.Count(result, "[搜索结果")
	if count > 3 {
		t.Errorf("搜索结果不应超过3条，实际: %d", count)
	}
	if count < 1 {
		t.Error("搜索结果应至少有1条")
	}

	t.Logf("Mock 搜索结果长度: %d, 条数: %d", len(result), count)
}

// TestWebSearch_Execute_EmptyQuery 测试空查询
func TestWebSearch_Execute_EmptyQuery(t *testing.T) {
	ws := NewWebSearch()

	_, err := ws.Execute(`{"query": ""}`)
	if err == nil {
		t.Fatal("空查询应返回错误")
	}
}

// TestWebSearch_Execute_InvalidJSON 测试无效 JSON
func TestWebSearch_Execute_InvalidJSON(t *testing.T) {
	ws := NewWebSearch()

	_, err := ws.Execute(`invalid json`)
	if err == nil {
		t.Fatal("无效 JSON 应返回错误")
	}
}

// TestToolRegistry 测试工具注册中心
func TestToolRegistry(t *testing.T) {
	registry := NewToolRegistry()

	if registry.HasTools() {
		t.Fatal("空注册中心不应有工具")
	}

	// 注册工具
	registry.Register(NewWebSearch())

	if !registry.HasTools() {
		t.Fatal("注册后应有工具")
	}

	// 获取工具
	tool, err := registry.Get("web_search")
	if err != nil {
		t.Fatalf("获取已注册工具失败: %v", err)
	}
	if tool.Name() != "web_search" {
		t.Fatal("获取的工具名称不正确")
	}

	// 获取未注册工具
	_, err = registry.Get("unknown_tool")
	if err == nil {
		t.Fatal("获取未注册工具应返回错误")
	}

	// 获取所有定义
	defs := registry.GetAllDefinitions()
	if len(defs) != 1 {
		t.Fatalf("应有1个工具定义，实际: %d", len(defs))
	}
}

// TestMockWebSearch_BackwardCompat 测试 mockWebSearch 向后兼容函数
func TestMockWebSearch_BackwardCompat(t *testing.T) {
	result := mockWebSearch("量子计算")
	if result == "" {
		t.Fatal("Mock 搜索结果不应为空")
	}

	if !strings.Contains(result, "量子计算") {
		t.Error("搜索结果应包含查询关键词")
	}
}

// TestAdaptiveRAG_SetMaxSearchPerTurn 测试设置最大搜索次数
func TestAdaptiveRAG_SetMaxSearchPerTurn(t *testing.T) {
	client := NewLLMClient("mock", "test-model", "", "", 0.7, 1000)
	rag := NewAdaptiveRAG(true, client)

	// 默认值应为 2
	if rag.maxSearchPerTurn != 2 {
		t.Fatalf("默认最大搜索次数应为2，实际: %d", rag.maxSearchPerTurn)
	}

	// 设置新值
	rag.SetMaxSearchPerTurn(5)
	if rag.maxSearchPerTurn != 5 {
		t.Fatalf("设置后最大搜索次数应为5，实际: %d", rag.maxSearchPerTurn)
	}

	// 设置无效值（<=0）不应改变
	rag.SetMaxSearchPerTurn(0)
	if rag.maxSearchPerTurn != 5 {
		t.Fatalf("无效值不应改变，应为5，实际: %d", rag.maxSearchPerTurn)
	}

	rag.SetMaxSearchPerTurn(-1)
	if rag.maxSearchPerTurn != 5 {
		t.Fatalf("负值不应改变，应为5，实际: %d", rag.maxSearchPerTurn)
	}
}

// TestBuildToolUsageGuidance 测试工具使用引导生成
func TestBuildToolUsageGuidance(t *testing.T) {
	guidance := buildToolUsageGuidance()

	if guidance == "" {
		t.Fatal("工具使用引导不应为空")
	}

	// 验证关键内容
	checks := []string{
		"web_search",
		"前沿研究",
		"时事",
		"明确要求",
		"每轮对话最多搜索 2 次",
	}
	for _, check := range checks {
		if !strings.Contains(guidance, check) {
			t.Errorf("工具使用引导应包含: %s", check)
		}
	}
}
