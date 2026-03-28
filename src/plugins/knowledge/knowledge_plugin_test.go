package knowledge

import (
	"context"
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

// setupKnowledgePlugin 创建测试用知识库插件
func setupKnowledgePlugin(t *testing.T) (*KnowledgePlugin, *database.Database) {
	t.Helper()
	db := setupTestDB(t)
	plugin := NewKnowledgePlugin("test-knowledge", db.DB)
	err := plugin.Init(map[string]interface{}{
		"document_processing.chunk_size":    100,
		"document_processing.chunk_overlap": 20,
		"retrieval.max_results":             10,
		"retrieval.similarity_threshold":    0.1,
	})
	if err != nil {
		t.Fatalf("初始化插件失败: %v", err)
	}
	return plugin, db
}

// createTestTeacher 在数据库中创建测试教师用户
func createTestTeacher(t *testing.T, db *database.Database) int64 {
	t.Helper()
	userRepo := database.NewUserRepository(db.DB)
	id, err := userRepo.Create(&database.User{
		Username: "teacher_test",
		Password: "hashed_password",
		Role:     "teacher",
		Nickname: "测试教师",
	})
	if err != nil {
		t.Fatalf("创建测试教师失败: %v", err)
	}
	return id
}

// ==================== TextChunker 测试 ====================

func TestTextChunker_BasicChunking(t *testing.T) {
	chunker := NewTextChunker(10, 0)
	text := "abcdefghijklmnopqrstuvwxyz"
	chunks := chunker.Chunk(text)

	if len(chunks) == 0 {
		t.Fatal("应该产生分块")
	}

	// 验证每个块的大小
	for i, chunk := range chunks {
		runes := []rune(chunk)
		if i < len(chunks)-1 && len(runes) != 10 {
			t.Errorf("块 %d 大小应为 10, 实际=%d", i, len(runes))
		}
	}
}

func TestTextChunker_WithOverlap(t *testing.T) {
	chunker := NewTextChunker(10, 3)
	text := "abcdefghijklmnopqrst" // 20 个字符
	chunks := chunker.Chunk(text)

	if len(chunks) < 2 {
		t.Fatalf("应该产生至少 2 个分块, 实际=%d", len(chunks))
	}

	// 验证 overlap: 第二个块应该从第 7 个字符开始（10-3=7）
	if len(chunks) >= 2 {
		runes := []rune(chunks[1])
		expected := []rune("hijklmnopq")
		if string(runes) != string(expected) {
			t.Errorf("第二个块期望=%s, 实际=%s", string(expected), string(runes))
		}
	}
}

func TestTextChunker_ShortText(t *testing.T) {
	chunker := NewTextChunker(100, 10)
	text := "short"
	chunks := chunker.Chunk(text)

	if len(chunks) != 1 {
		t.Fatalf("短文本应该只有 1 个块, 实际=%d", len(chunks))
	}
	if chunks[0] != "short" {
		t.Errorf("期望='short', 实际=%s", chunks[0])
	}
}

func TestTextChunker_EmptyText(t *testing.T) {
	chunker := NewTextChunker(100, 10)
	chunks := chunker.Chunk("")

	if len(chunks) != 0 {
		t.Fatalf("空文本不应产生分块, 实际=%d", len(chunks))
	}
}

func TestTextChunker_ChineseText(t *testing.T) {
	chunker := NewTextChunker(5, 0)
	text := "你好世界测试文本"
	chunks := chunker.Chunk(text)

	if len(chunks) == 0 {
		t.Fatal("中文文本应该产生分块")
	}
	// 第一个块应该是 5 个中文字符
	runes := []rune(chunks[0])
	if len(runes) != 5 {
		t.Errorf("第一个块应有 5 个字符, 实际=%d", len(runes))
	}
}

func TestTextChunker_InvalidParams(t *testing.T) {
	// chunkSize <= 0 应使用默认值
	chunker := NewTextChunker(0, 0)
	if chunker.chunkSize != 500 {
		t.Errorf("chunkSize 应为默认值 500, 实际=%d", chunker.chunkSize)
	}

	// overlap >= chunkSize 应调整
	chunker2 := NewTextChunker(10, 15)
	if chunker2.chunkOverlap >= chunker2.chunkSize {
		t.Errorf("chunkOverlap 不应 >= chunkSize")
	}
}

// ==================== InMemoryVectorStore 测试 ====================

func TestVectorStore_AddAndSearch(t *testing.T) {
	store := NewInMemoryVectorStore()

	chunks := []VectorChunk{
		{ID: "1", Content: "Go 语言是一种编程语言", DocumentID: 1, TeacherID: 1, Title: "Go 入门"},
		{ID: "2", Content: "Python 是一种流行的脚本语言", DocumentID: 2, TeacherID: 1, Title: "Python 入门"},
		{ID: "3", Content: "Java 是一种面向对象的编程语言", DocumentID: 3, TeacherID: 1, Title: "Java 入门"},
	}

	err := store.AddDocuments("test_collection", chunks)
	if err != nil {
		t.Fatalf("添加文档失败: %v", err)
	}

	// 搜索包含 "Go" 的内容
	results := store.Search("test_collection", "Go", 5)
	if len(results) == 0 {
		t.Fatal("应该找到结果")
	}

	// 第一个结果应该包含 "Go"
	found := false
	for _, r := range results {
		if r.Title == "Go 入门" {
			found = true
			break
		}
	}
	if !found {
		t.Error("搜索 Go 应该找到 Go 入门文档")
	}
}

func TestVectorStore_SearchEmptyCollection(t *testing.T) {
	store := NewInMemoryVectorStore()

	results := store.Search("nonexistent", "query", 5)
	if len(results) != 0 {
		t.Error("空集合搜索应返回空结果")
	}
}

func TestVectorStore_DeleteByDocumentID(t *testing.T) {
	store := NewInMemoryVectorStore()

	chunks := []VectorChunk{
		{ID: "1", Content: "文档1内容", DocumentID: 1, TeacherID: 1, Title: "文档1"},
		{ID: "2", Content: "文档1第二块", DocumentID: 1, TeacherID: 1, Title: "文档1"},
		{ID: "3", Content: "文档2内容", DocumentID: 2, TeacherID: 1, Title: "文档2"},
	}

	store.AddDocuments("test", chunks)

	// 删除文档1
	err := store.DeleteByDocumentID("test", 1)
	if err != nil {
		t.Fatalf("删除失败: %v", err)
	}

	// 搜索文档1内容应该找不到
	results := store.Search("test", "文档1", 5)
	for _, r := range results {
		if r.DocumentID == 1 {
			t.Error("文档1应该已被删除")
		}
	}
}

func TestVectorStore_SearchWithLimit(t *testing.T) {
	store := NewInMemoryVectorStore()

	var chunks []VectorChunk
	for i := 0; i < 10; i++ {
		chunks = append(chunks, VectorChunk{
			ID:         "chunk_" + string(rune('0'+i)),
			Content:    "测试内容关键词",
			DocumentID: int64(i + 1),
			TeacherID:  1,
			Title:      "测试文档",
		})
	}

	store.AddDocuments("test", chunks)

	results := store.Search("test", "关键词", 3)
	if len(results) > 3 {
		t.Errorf("结果数量应 <= 3, 实际=%d", len(results))
	}
}

// ==================== KnowledgePlugin 测试 ====================

func TestKnowledgePlugin_NewAndInit(t *testing.T) {
	db := setupTestDB(t)
	plugin := NewKnowledgePlugin("knowledge-test", db.DB)

	if plugin.Name() != "knowledge-test" {
		t.Errorf("期望名称=knowledge-test, 实际=%s", plugin.Name())
	}
	if plugin.Type() != core.PluginTypeKnowledge {
		t.Errorf("期望类型=knowledge, 实际=%s", plugin.Type())
	}

	err := plugin.Init(map[string]interface{}{
		"document_processing.chunk_size":    200,
		"document_processing.chunk_overlap": 30,
	})
	if err != nil {
		t.Fatalf("初始化失败: %v", err)
	}
}

func TestKnowledgePlugin_AddDocument(t *testing.T) {
	plugin, db := setupKnowledgePlugin(t)
	teacherID := createTestTeacher(t, db)

	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action":     "add",
			"title":      "测试文档",
			"content":    "这是一段测试内容，用于验证知识库插件的文档添加功能。内容需要足够长以便测试分块功能。这里添加更多的文本内容来确保能够产生多个分块。",
			"tags":       "test,knowledge",
			"teacher_id": teacherID,
		},
		UserContext: &core.UserContext{UserID: "1"},
		Context:     context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行添加失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("添加应该成功, 错误: %s", output.Error)
	}
	if output.Data["document_id"] == nil {
		t.Fatal("应返回 document_id")
	}
	if output.Data["chunks_count"] == nil {
		t.Fatal("应返回 chunks_count")
	}

	chunksCount, ok := output.Data["chunks_count"].(int)
	if !ok || chunksCount <= 0 {
		t.Errorf("chunks_count 应 > 0, 实际=%v", output.Data["chunks_count"])
	}
}

func TestKnowledgePlugin_AddDocument_MergeData(t *testing.T) {
	plugin, db := setupKnowledgePlugin(t)
	teacherID := createTestTeacher(t, db)

	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action":       "add",
			"title":        "Merge 测试",
			"content":      "测试 merge 功能的文档内容",
			"teacher_id":   teacherID,
			"custom_field": "preserved_value",
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("添加应该成功: %s", output.Error)
	}

	// 验证上游 Data 被 merge
	if output.Data["custom_field"] != "preserved_value" {
		t.Error("上游数据 custom_field 应该被保留")
	}
}

func TestKnowledgePlugin_AddDocument_MissingParams(t *testing.T) {
	plugin, _ := setupKnowledgePlugin(t)

	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action":     "add",
			"title":      "",
			"content":    "",
			"teacher_id": int64(1),
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if output.Success {
		t.Fatal("缺少参数应该失败")
	}
	if output.Data["error_code"] != 40004 {
		t.Errorf("期望错误码 40004, 实际=%v", output.Data["error_code"])
	}
}

func TestKnowledgePlugin_Search(t *testing.T) {
	plugin, db := setupKnowledgePlugin(t)
	teacherID := createTestTeacher(t, db)

	// 先添加文档
	addInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":     "add",
			"title":      "Go 编程入门",
			"content":    "Go 语言是 Google 开发的一种静态类型的编程语言，它具有简洁的语法和高效的并发支持",
			"teacher_id": teacherID,
		},
		Context: context.Background(),
	}
	_, err := plugin.Execute(context.Background(), addInput)
	if err != nil {
		t.Fatalf("添加文档失败: %v", err)
	}

	// 搜索
	searchInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":     "search",
			"query":      "Go 编程",
			"teacher_id": teacherID,
			"limit":      5,
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), searchInput)
	if err != nil {
		t.Fatalf("搜索执行失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("搜索应该成功, 错误: %s", output.Error)
	}

	chunks, ok := output.Data["chunks"].([]map[string]interface{})
	if !ok {
		t.Fatal("应返回 chunks 数组")
	}
	if len(chunks) == 0 {
		t.Fatal("应找到匹配结果")
	}
}

func TestKnowledgePlugin_Search_MissingQuery(t *testing.T) {
	plugin, _ := setupKnowledgePlugin(t)

	searchInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":     "search",
			"query":      "",
			"teacher_id": int64(1),
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), searchInput)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if output.Success {
		t.Fatal("缺少 query 应该失败")
	}
}

func TestKnowledgePlugin_List(t *testing.T) {
	plugin, db := setupKnowledgePlugin(t)
	teacherID := createTestTeacher(t, db)

	// 添加多个文档
	for i := 0; i < 3; i++ {
		addInput := &core.PluginInput{
			Data: map[string]interface{}{
				"action":     "add",
				"title":      "文档" + string(rune('A'+i)),
				"content":    "这是文档的内容",
				"teacher_id": teacherID,
			},
			Context: context.Background(),
		}
		_, err := plugin.Execute(context.Background(), addInput)
		if err != nil {
			t.Fatalf("添加文档失败: %v", err)
		}
	}

	// 列表查询
	listInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":     "list",
			"teacher_id": teacherID,
			"page":       1,
			"page_size":  10,
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), listInput)
	if err != nil {
		t.Fatalf("列表查询失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("列表查询应该成功, 错误: %s", output.Error)
	}

	docs, ok := output.Data["documents"].([]map[string]interface{})
	if !ok {
		t.Fatal("应返回 documents 数组")
	}
	if len(docs) != 3 {
		t.Errorf("期望 3 个文档, 实际=%d", len(docs))
	}

	total, ok := output.Data["total"].(int)
	if !ok || total != 3 {
		t.Errorf("期望 total=3, 实际=%v", output.Data["total"])
	}
}

func TestKnowledgePlugin_Delete(t *testing.T) {
	plugin, db := setupKnowledgePlugin(t)
	teacherID := createTestTeacher(t, db)

	// 先添加文档
	addInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":     "add",
			"title":      "待删除文档",
			"content":    "这个文档将被删除",
			"teacher_id": teacherID,
		},
		Context: context.Background(),
	}
	addOutput, err := plugin.Execute(context.Background(), addInput)
	if err != nil {
		t.Fatalf("添加文档失败: %v", err)
	}

	docID := addOutput.Data["document_id"]

	// 删除文档
	deleteInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":      "delete",
			"document_id": docID,
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), deleteInput)
	if err != nil {
		t.Fatalf("删除执行失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("删除应该成功, 错误: %s", output.Error)
	}
	if output.Data["deleted"] != true {
		t.Error("应返回 deleted=true")
	}

	// 验证列表中已经看不到
	listInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":     "list",
			"teacher_id": teacherID,
			"page":       1,
			"page_size":  10,
		},
		Context: context.Background(),
	}
	listOutput, _ := plugin.Execute(context.Background(), listInput)
	docs, _ := listOutput.Data["documents"].([]map[string]interface{})
	if len(docs) != 0 {
		t.Errorf("删除后文档列表应为空, 实际=%d", len(docs))
	}
}

func TestKnowledgePlugin_Delete_NotExist(t *testing.T) {
	plugin, _ := setupKnowledgePlugin(t)

	deleteInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":      "delete",
			"document_id": int64(99999),
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), deleteInput)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if output.Success {
		t.Fatal("删除不存在的文档应该失败")
	}
	if output.Data["error_code"] != 40005 {
		t.Errorf("期望错误码 40005, 实际=%v", output.Data["error_code"])
	}
}

func TestKnowledgePlugin_InvalidAction(t *testing.T) {
	plugin, _ := setupKnowledgePlugin(t)

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

func TestKnowledgePlugin_MissingAction(t *testing.T) {
	plugin, _ := setupKnowledgePlugin(t)

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
	if output.Data["error_code"] != 40004 {
		t.Errorf("期望错误码 40004, 实际=%v", output.Data["error_code"])
	}
}
