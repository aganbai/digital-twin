package dialogue

import (
	"context"
	"fmt"
	"os"
	"testing"

	"digital-twin/src/backend/database"
	"digital-twin/src/harness/core"
)

// setupTestDB 创建测试用数据库
func setupTestDB(t *testing.T) *database.Database {
	t.Helper()
	tmpFile := t.TempDir() + "/test.db"
	db, err := database.NewDatabase(tmpFile)
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
		os.Remove(tmpFile)
	})
	return db
}

// createTestUsers 创建测试用的学生和教师
func createTestUsers(t *testing.T, db *database.Database) (studentID, teacherID int64) {
	t.Helper()
	userRepo := database.NewUserRepository(db.DB)

	teacherID, err := userRepo.Create(&database.User{
		Username: "teacher_test",
		Password: "hashed_password",
		Role:     "teacher",
		Nickname: "测试教师",
	})
	if err != nil {
		t.Fatalf("创建测试教师失败: %v", err)
	}

	studentID, err = userRepo.Create(&database.User{
		Username: "student_test",
		Password: "hashed_password",
		Role:     "student",
		Nickname: "测试学生",
	})
	if err != nil {
		t.Fatalf("创建测试学生失败: %v", err)
	}

	return studentID, teacherID
}

// setupDialoguePlugin 创建测试用对话插件（mock 模式）
func setupDialoguePlugin(t *testing.T) (*DialoguePlugin, *database.Database, int64, int64) {
	t.Helper()
	db := setupTestDB(t)
	studentID, teacherID := createTestUsers(t, db)

	plugin := NewDialoguePlugin("test-dialogue", db.DB)
	err := plugin.Init(map[string]interface{}{
		"llm_provider.mode":                "mock",
		"llm_provider.model":               "qwen-turbo",
		"dialogue_strategy.temperature":    0.7,
		"dialogue_strategy.max_tokens":     1000,
		"context_management.history_limit": 10,
	})
	if err != nil {
		t.Fatalf("初始化插件失败: %v", err)
	}

	return plugin, db, studentID, teacherID
}

// ==================== LLMClient 测试 ====================

func TestLLMClient_MockMode(t *testing.T) {
	client := NewLLMClient("mock", "test-model", "", "", 0.7, 1000)

	messages := []ChatMessage{
		{Role: "system", Content: "你是一位AI教师"},
		{Role: "user", Content: "什么是递归？"},
	}

	resp, err := client.Chat(messages)
	if err != nil {
		t.Fatalf("Mock 模式调用失败: %v", err)
	}
	if resp.Content == "" {
		t.Fatal("回复内容不应为空")
	}
	if resp.TotalTokens <= 0 {
		t.Error("token 计数应 > 0")
	}
}

func TestLLMClient_MockMode_EmptyMessages(t *testing.T) {
	client := NewLLMClient("mock", "test-model", "", "", 0.7, 1000)

	_, err := client.Chat([]ChatMessage{})
	if err == nil {
		t.Fatal("空消息列表应返回错误")
	}
}

func TestLLMClient_DefaultMode(t *testing.T) {
	// 默认模式应该回退到 mock
	client := NewLLMClient("", "test-model", "", "", 0.7, 1000)

	messages := []ChatMessage{
		{Role: "user", Content: "测试"},
	}

	resp, err := client.Chat(messages)
	if err != nil {
		t.Fatalf("默认模式调用失败: %v", err)
	}
	if resp.Content == "" {
		t.Fatal("回复内容不应为空")
	}
}

// ==================== PromptBuilder 测试 ====================

func TestPromptBuilder_BuildSystemPrompt_WithData(t *testing.T) {
	builder := NewPromptBuilder()

	chunks := []map[string]interface{}{
		{"content": "Go 是一种编程语言", "title": "Go 入门"},
		{"content": "Go 支持并发编程", "title": "Go 并发"},
	}

	memories := []map[string]interface{}{
		{"content": "学生掌握了变量定义", "memory_type": "concept"},
		{"content": "学生在循环方面有困难", "memory_type": "weakness"},
	}

	prompt := builder.BuildSystemPrompt(chunks, memories, nil, nil, "")

	// 验证包含关键内容
	if prompt == "" {
		t.Fatal("prompt 不应为空")
	}
	if !contains(prompt, "苏格拉底") {
		t.Error("prompt 应包含苏格拉底教学法描述")
	}
	if !contains(prompt, "Go 是一种编程语言") {
		t.Error("prompt 应包含知识片段")
	}
	if !contains(prompt, "学生掌握了变量定义") {
		t.Error("prompt 应包含学生记忆")
	}
}

func TestPromptBuilder_BuildSystemPrompt_Empty(t *testing.T) {
	builder := NewPromptBuilder()

	prompt := builder.BuildSystemPrompt(nil, nil, nil, nil, "")

	if !contains(prompt, "暂无相关知识") {
		t.Error("无知识时应显示默认文本")
	}
	if !contains(prompt, "暂无学生记忆") {
		t.Error("无记忆时应显示默认文本")
	}
}

func TestPromptBuilder_BuildConversationMessages(t *testing.T) {
	builder := NewPromptBuilder()

	history := []*database.Conversation{
		{Role: "user", Content: "什么是变量？"},
		{Role: "assistant", Content: "你觉得变量是什么呢？"},
	}

	messages := builder.BuildConversationMessages("system prompt", history, "我觉得变量是存储数据的容器")

	// 验证消息结构
	if len(messages) != 4 { // system + 2 history + 1 user
		t.Fatalf("期望 4 条消息, 实际=%d", len(messages))
	}
	if messages[0].Role != "system" {
		t.Error("第一条应为 system 消息")
	}
	if messages[1].Role != "user" {
		t.Error("第二条应为 user 消息（历史）")
	}
	if messages[2].Role != "assistant" {
		t.Error("第三条应为 assistant 消息（历史）")
	}
	if messages[3].Role != "user" {
		t.Error("最后一条应为 user 消息（当前）")
	}
	if messages[3].Content != "我觉得变量是存储数据的容器" {
		t.Error("最后一条消息内容不匹配")
	}
}

func TestPromptBuilder_BuildConversationMessages_NoHistory(t *testing.T) {
	builder := NewPromptBuilder()

	messages := builder.BuildConversationMessages("system prompt", nil, "你好")

	if len(messages) != 2 { // system + user
		t.Fatalf("期望 2 条消息, 实际=%d", len(messages))
	}
}

// ==================== DialoguePlugin 测试 ====================

func TestDialoguePlugin_NewAndInit(t *testing.T) {
	db := setupTestDB(t)
	plugin := NewDialoguePlugin("dialogue-test", db.DB)

	if plugin.Name() != "dialogue-test" {
		t.Errorf("期望名称=dialogue-test, 实际=%s", plugin.Name())
	}
	if plugin.Type() != core.PluginTypeDialogue {
		t.Errorf("期望类型=dialogue, 实际=%s", plugin.Type())
	}

	err := plugin.Init(map[string]interface{}{
		"llm_provider.mode":  "mock",
		"llm_provider.model": "test-model",
	})
	if err != nil {
		t.Fatalf("初始化失败: %v", err)
	}
}

func TestDialoguePlugin_ChatAction(t *testing.T) {
	plugin, _, studentID, teacherID := setupDialoguePlugin(t)

	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action":     "chat",
			"message":    "什么是递归？",
			"teacher_id": teacherID,
		},
		UserContext: &core.UserContext{
			UserID: fmt.Sprintf("%d", studentID),
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行 chat 失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("chat 应该成功, 错误: %s", output.Error)
	}

	// 验证返回字段
	if output.Data["reply"] == nil || output.Data["reply"] == "" {
		t.Fatal("应返回 reply")
	}
	if output.Data["session_id"] == nil || output.Data["session_id"] == "" {
		t.Fatal("应返回 session_id")
	}
	if output.Data["conversation_id"] == nil {
		t.Fatal("应返回 conversation_id")
	}
	if output.Data["token_usage"] == nil {
		t.Fatal("应返回 token_usage")
	}
	if output.Data["pipeline_duration_ms"] == nil {
		t.Fatal("应返回 pipeline_duration_ms")
	}
}

func TestDialoguePlugin_ChatAction_WithSessionID(t *testing.T) {
	plugin, _, studentID, teacherID := setupDialoguePlugin(t)

	customSessionID := "custom-session-123"
	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action":     "chat",
			"message":    "继续上次的话题",
			"teacher_id": teacherID,
			"session_id": customSessionID,
		},
		UserContext: &core.UserContext{
			UserID: fmt.Sprintf("%d", studentID),
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("chat 应该成功: %s", output.Error)
	}

	// 验证使用了自定义 session_id
	if output.Data["session_id"] != customSessionID {
		t.Errorf("期望 session_id=%s, 实际=%v", customSessionID, output.Data["session_id"])
	}
}

func TestDialoguePlugin_ChatAction_WithUpstreamData(t *testing.T) {
	plugin, _, studentID, teacherID := setupDialoguePlugin(t)

	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action":     "chat",
			"message":    "什么是变量？",
			"teacher_id": teacherID,
			"memories": []map[string]interface{}{
				{"content": "学生已掌握基本概念", "memory_type": "concept"},
			},
			"chunks": []map[string]interface{}{
				{"content": "变量是存储数据的容器", "title": "编程基础"},
			},
		},
		UserContext: &core.UserContext{
			UserID: fmt.Sprintf("%d", studentID),
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("chat 应该成功: %s", output.Error)
	}

	// chat action 不 merge 上游数据，所以不应有 memories/chunks
	if _, ok := output.Data["memories"]; ok {
		t.Error("chat action 不应 merge 上游 memories")
	}
}

func TestDialoguePlugin_ChatAction_MissingMessage(t *testing.T) {
	plugin, _, studentID, teacherID := setupDialoguePlugin(t)

	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action":     "chat",
			"message":    "",
			"teacher_id": teacherID,
		},
		UserContext: &core.UserContext{
			UserID: fmt.Sprintf("%d", studentID),
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if output.Success {
		t.Fatal("空消息应该失败")
	}
}

func TestDialoguePlugin_ChatAction_MissingTeacherID(t *testing.T) {
	plugin, _, studentID, _ := setupDialoguePlugin(t)

	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action":  "chat",
			"message": "测试",
		},
		UserContext: &core.UserContext{
			UserID: fmt.Sprintf("%d", studentID),
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if output.Success {
		t.Fatal("缺少 teacher_id 应该失败")
	}
}

func TestDialoguePlugin_ChatAction_MissingUserContext(t *testing.T) {
	plugin, _, _, teacherID := setupDialoguePlugin(t)

	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action":     "chat",
			"message":    "测试",
			"teacher_id": teacherID,
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if output.Success {
		t.Fatal("缺少用户信息应该失败")
	}
}

func TestDialoguePlugin_HistoryAction(t *testing.T) {
	plugin, _, studentID, teacherID := setupDialoguePlugin(t)

	// 先进行一次对话
	chatInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":     "chat",
			"message":    "什么是递归？",
			"teacher_id": teacherID,
		},
		UserContext: &core.UserContext{
			UserID: fmt.Sprintf("%d", studentID),
		},
		Context: context.Background(),
	}
	_, err := plugin.Execute(context.Background(), chatInput)
	if err != nil {
		t.Fatalf("对话失败: %v", err)
	}

	// 查询历史
	historyInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":     "history",
			"teacher_id": teacherID,
			"page":       1,
			"page_size":  10,
		},
		UserContext: &core.UserContext{
			UserID: fmt.Sprintf("%d", studentID),
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), historyInput)
	if err != nil {
		t.Fatalf("查询历史失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("查询历史应该成功, 错误: %s", output.Error)
	}

	items, ok := output.Data["items"].([]map[string]interface{})
	if !ok {
		t.Fatal("应返回 items 数组")
	}
	if len(items) != 2 { // user + assistant
		t.Errorf("期望 2 条记录, 实际=%d", len(items))
	}

	total, ok := output.Data["total"].(int)
	if !ok || total != 2 {
		t.Errorf("期望 total=2, 实际=%v", output.Data["total"])
	}
}

func TestDialoguePlugin_HistoryAction_MergeData(t *testing.T) {
	plugin, _, studentID, teacherID := setupDialoguePlugin(t)

	historyInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":       "history",
			"teacher_id":   teacherID,
			"custom_field": "should_be_preserved",
		},
		UserContext: &core.UserContext{
			UserID: fmt.Sprintf("%d", studentID),
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), historyInput)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("查询历史应该成功: %s", output.Error)
	}

	// 验证上游 Data 被 merge
	if output.Data["custom_field"] != "should_be_preserved" {
		t.Error("上游数据 custom_field 应该被保留")
	}
}

func TestDialoguePlugin_HistoryAction_MissingTeacherID(t *testing.T) {
	plugin, _, studentID, _ := setupDialoguePlugin(t)

	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action": "history",
		},
		UserContext: &core.UserContext{
			UserID: fmt.Sprintf("%d", studentID),
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if output.Success {
		t.Fatal("缺少 teacher_id 应该失败")
	}
}

func TestDialoguePlugin_MissingAction(t *testing.T) {
	plugin, _, _, _ := setupDialoguePlugin(t)

	input := &core.PluginInput{
		Data:    map[string]interface{}{},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if output.Success {
		t.Fatal("缺少 action 应该失败")
	}
}

func TestDialoguePlugin_InvalidAction(t *testing.T) {
	plugin, _, _, _ := setupDialoguePlugin(t)

	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action": "unknown",
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if output.Success {
		t.Fatal("未知 action 应该失败")
	}
}

func TestDialoguePlugin_HealthCheck(t *testing.T) {
	plugin, _, _, _ := setupDialoguePlugin(t)

	err := plugin.HealthCheck()
	if err != nil {
		t.Fatalf("健康检查失败: %v", err)
	}
}

func TestDialoguePlugin_MultiTurnChat(t *testing.T) {
	plugin, _, studentID, teacherID := setupDialoguePlugin(t)

	sessionID := "multi-turn-session"

	// 第一轮对话
	input1 := &core.PluginInput{
		Data: map[string]interface{}{
			"action":     "chat",
			"message":    "什么是递归？",
			"teacher_id": teacherID,
			"session_id": sessionID,
		},
		UserContext: &core.UserContext{
			UserID: fmt.Sprintf("%d", studentID),
		},
		Context: context.Background(),
	}

	output1, err := plugin.Execute(context.Background(), input1)
	if err != nil {
		t.Fatalf("第一轮对话失败: %v", err)
	}
	if !output1.Success {
		t.Fatalf("第一轮对话应该成功: %s", output1.Error)
	}

	// 第二轮对话
	input2 := &core.PluginInput{
		Data: map[string]interface{}{
			"action":     "chat",
			"message":    "我觉得递归就是函数调用自己",
			"teacher_id": teacherID,
			"session_id": sessionID,
		},
		UserContext: &core.UserContext{
			UserID: fmt.Sprintf("%d", studentID),
		},
		Context: context.Background(),
	}

	output2, err := plugin.Execute(context.Background(), input2)
	if err != nil {
		t.Fatalf("第二轮对话失败: %v", err)
	}
	if !output2.Success {
		t.Fatalf("第二轮对话应该成功: %s", output2.Error)
	}

	// 验证历史中有 4 条记录（2 轮 × 2 条）
	historyInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":     "history",
			"teacher_id": teacherID,
			"page":       1,
			"page_size":  20,
		},
		UserContext: &core.UserContext{
			UserID: fmt.Sprintf("%d", studentID),
		},
		Context: context.Background(),
	}

	historyOutput, err := plugin.Execute(context.Background(), historyInput)
	if err != nil {
		t.Fatalf("查询历史失败: %v", err)
	}

	total, _ := historyOutput.Data["total"].(int)
	if total != 4 {
		t.Errorf("期望 4 条历史记录, 实际=%d", total)
	}
}

// contains 辅助函数：检查字符串是否包含子串
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
