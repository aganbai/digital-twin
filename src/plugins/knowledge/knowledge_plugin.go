package knowledge

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"digital-twin/src/backend/database"
	"digital-twin/src/harness/core"
)

// KnowledgePlugin 知识库插件
type KnowledgePlugin struct {
	*core.BasePlugin
	db           *sql.DB
	docRepo      *database.DocumentRepository
	vectorStore  *InMemoryVectorStore
	vectorClient *VectorClient
	chunker      *TextChunker
}

// NewKnowledgePlugin 创建知识库插件
func NewKnowledgePlugin(name string, db *sql.DB) *KnowledgePlugin {
	return &KnowledgePlugin{
		BasePlugin:   core.NewBasePlugin(name, "1.0.0", core.PluginTypeKnowledge),
		db:           db,
		docRepo:      database.NewDocumentRepository(db),
		vectorStore:  nil, // 不再初始化，所有调用已改为 vectorClient；保留字段以备未来降级使用
		vectorClient: NewVectorClient(),
	}
}

// Init 初始化知识库插件
// 读取 retrieval.max_results, similarity_threshold, document_processing.chunk_size/chunk_overlap
func (p *KnowledgePlugin) Init(config map[string]interface{}) error {
	// 调用基类 Init
	if err := p.BasePlugin.Init(config); err != nil {
		return err
	}

	// 读取分块配置
	chunkSize := 500   // 默认值
	chunkOverlap := 50 // 默认值

	if v, ok := config["document_processing.chunk_size"]; ok {
		chunkSize = toInt(v, chunkSize)
	}
	if v, ok := config["document_processing.chunk_overlap"]; ok {
		chunkOverlap = toInt(v, chunkOverlap)
	}

	p.chunker = NewTextChunker(chunkSize, chunkOverlap)
	return nil
}

// isKnowledgeAction 判断是否是知识库插件自己的 action
func (p *KnowledgePlugin) isKnowledgeAction(action string) bool {
	switch action {
	case "add", "search", "list", "delete", "preview":
		return true
	default:
		return false
	}
}

// Execute 执行知识库操作
// 根据 input.Data["action"] 分发到不同的处理逻辑
func (p *KnowledgePlugin) Execute(ctx context.Context, input *core.PluginInput) (*core.PluginOutput, error) {
	start := time.Now()

	action, _ := input.Data["action"].(string)

	var output *core.PluginOutput
	var err error

	switch action {
	case "add":
		output, err = p.handleAdd(input)
	case "search":
		output, err = p.handleSearch(input)
	case "list":
		output, err = p.handleList(input)
	case "delete":
		output, err = p.handleDelete(input)
	case "preview":
		output, err = p.handlePreview(input)
	default:
		// 判断是否为管道模式：需要有 message 字段
		_, hasMessage := input.Data["message"]
		if !hasMessage {
			// 非管道模式，action 无效或缺失
			errMsg := "缺少 action 参数"
			if action != "" {
				errMsg = fmt.Sprintf("不支持的 action: %s", action)
			}
			output = &core.PluginOutput{
				Success: false,
				Data:    map[string]interface{}{"error_code": 40004},
				Error:   errMsg,
			}
		} else {
			// 管道模式：自动检索知识并注入到 Data
			output, err = p.handlePipeline(input)
			if err != nil {
				return &core.PluginOutput{
					Success:  false,
					Data:     map[string]interface{}{"error_code": 50001},
					Error:    err.Error(),
					Duration: time.Since(start),
				}, nil
			}
			output.Duration = time.Since(start)
			return output, nil
		}
	}

	if err != nil {
		return &core.PluginOutput{
			Success:  false,
			Data:     map[string]interface{}{"error_code": 50001},
			Error:    err.Error(),
			Duration: time.Since(start),
		}, nil
	}

	output.Duration = time.Since(start)
	return output, nil
}

// handlePipeline 管道模式：自动检索知识并注入到 Data
func (p *KnowledgePlugin) handlePipeline(input *core.PluginInput) (*core.PluginOutput, error) {
	// 从 Data 获取 message 作为检索 query
	query, _ := input.Data["message"].(string)
	var teacherID int64
	if v, ok := input.Data["teacher_id"]; ok {
		teacherID = toInt64(v, 0)
	}

	// 获取分身ID（用于 scope 过滤）
	var teacherPersonaID int64
	if v, ok := input.Data["teacher_persona_id"]; ok {
		teacherPersonaID = toInt64(v, 0)
	}
	var studentPersonaID int64
	if v, ok := input.Data["student_persona_id"]; ok {
		studentPersonaID = toInt64(v, 0)
	}

	// merge 上游 Data
	outputData := mergeData(input.Data, nil)

	// 如果有 query 和 teacher_id，执行检索
	if query != "" && teacherID > 0 {
		collectionName := fmt.Sprintf("teacher_%d", teacherID)
		// 先检索较多结果，再按 scope 过滤（使用 VectorClient 调用 Python 服务）
		results := p.vectorClient.Search(collectionName, query, 20)

		// 如果有 student_persona_id，需要按 scope 过滤
		if studentPersonaID > 0 && teacherPersonaID > 0 {
			results = p.filterByScope(results, teacherPersonaID, studentPersonaID)
		}

		// 限制最终返回数量
		if len(results) > 5 {
			results = results[:5]
		}

		var chunksOutput []map[string]interface{}
		for _, r := range results {
			chunksOutput = append(chunksOutput, map[string]interface{}{
				"content":     r.Content,
				"score":       r.Score,
				"document_id": r.DocumentID,
				"title":       r.Title,
			})
		}
		if chunksOutput == nil {
			chunksOutput = []map[string]interface{}{}
		}
		outputData["chunks"] = chunksOutput
	} else {
		outputData["chunks"] = []map[string]interface{}{}
	}

	return &core.PluginOutput{
		Success:  true,
		Data:     outputData,
		Metadata: map[string]interface{}{"plugin": "knowledge", "mode": "pipeline"},
	}, nil
}

// filterByScope 根据学生分身的 scope 过滤检索结果
// 只保留 scope=global 或 scope=class（学生所在班级）或 scope=student（指定给该学生）的文档
func (p *KnowledgePlugin) filterByScope(results []SearchResult, teacherPersonaID, studentPersonaID int64) []SearchResult {
	// 查询学生所在班级
	classRepo := database.NewClassRepository(p.db)
	classIDs, err := classRepo.GetClassIDsByStudentPersona(teacherPersonaID, studentPersonaID)
	if err != nil {
		// 查询失败时返回所有结果（降级处理）
		return results
	}

	// 构建班级ID集合
	classIDSet := make(map[int64]bool)
	for _, cid := range classIDs {
		classIDSet[cid] = true
	}

	// 查询向量存储中每个结果对应的 chunk 的 scope 信息
	// 由于 SearchResult 不包含 scope 信息，需要通过 document_id 从数据库查询
	docRepo := database.NewDocumentRepository(p.db)
	docCache := make(map[int64]*database.Document)

	var filtered []SearchResult
	for _, r := range results {
		// 查询文档的 scope 信息（带缓存）
		doc, exists := docCache[r.DocumentID]
		if !exists {
			doc, err = docRepo.GetByID(r.DocumentID)
			if err != nil || doc == nil {
				// 查询失败或文档不存在，跳过
				continue
			}
			docCache[r.DocumentID] = doc
		}

		// 判断 scope 是否可见
		switch doc.Scope {
		case "global", "":
			filtered = append(filtered, r)
		case "class":
			if classIDSet[doc.ScopeID] {
				filtered = append(filtered, r)
			}
		case "student":
			if doc.ScopeID == studentPersonaID {
				filtered = append(filtered, r)
			}
		}
	}

	return filtered
}

// handleAdd 添加文档
func (p *KnowledgePlugin) handleAdd(input *core.PluginInput) (*core.PluginOutput, error) {
	title, _ := input.Data["title"].(string)
	content, _ := input.Data["content"].(string)
	tags, _ := input.Data["tags"].(string)
	summary, _ := input.Data["summary"].(string)

	// 从 UserContext 获取 teacher_id
	var teacherID int64
	if input.UserContext != nil && input.UserContext.UserID != "" {
		teacherID = toInt64(input.UserContext.UserID, 0)
	}
	// 也允许从 Data 中直接传入
	if v, ok := input.Data["teacher_id"]; ok {
		teacherID = toInt64(v, teacherID)
	}

	// 获取 scope 相关字段
	scope, _ := input.Data["scope"].(string)
	if scope == "" {
		scope = "global"
	}
	var scopeID int64
	if v, ok := input.Data["scope_id"]; ok {
		scopeID = toInt64(v, 0)
	}
	var personaID int64
	if v, ok := input.Data["persona_id"]; ok {
		personaID = toInt64(v, 0)
	}

	// 参数校验
	if title == "" || content == "" {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "标题和内容不能为空",
		}, nil
	}

	if teacherID <= 0 {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "缺少 teacher_id",
		}, nil
	}

	// 创建文档记录
	doc := &database.Document{
		TeacherID: teacherID,
		Title:     title,
		Content:   content,
		DocType:   "text",
		Tags:      tags,
		Status:    "active",
		Scope:     scope,
		ScopeID:   scopeID,
		PersonaID: personaID,
		Summary:   summary,
	}

	docID, err := p.docRepo.Create(doc)
	if err != nil {
		return nil, fmt.Errorf("创建文档失败: %w", err)
	}

	// 文本分块
	chunks := p.chunker.Chunk(content)

	// 存入向量存储
	collectionName := fmt.Sprintf("teacher_%d", teacherID)
	var vectorChunks []VectorChunk
	for i, chunk := range chunks {
		vectorChunks = append(vectorChunks, VectorChunk{
			ID:         fmt.Sprintf("doc_%d_chunk_%d", docID, i),
			Content:    chunk,
			DocumentID: docID,
			TeacherID:  teacherID,
			Title:      title,
			ChunkIndex: i,
			Scope:      scope,
			ScopeID:    scopeID,
		})
	}

	if err := p.vectorClient.AddDocuments(collectionName, docID, title, vectorChunks); err != nil {
		return nil, fmt.Errorf("存储文档块失败: %w", err)
	}

	// 构建输出数据，merge 上游 Data
	outputData := mergeData(input.Data, map[string]interface{}{
		"document_id":  docID,
		"chunks_count": len(chunks),
	})

	return &core.PluginOutput{
		Success:  true,
		Data:     outputData,
		Metadata: map[string]interface{}{"plugin": "knowledge", "action": "add"},
	}, nil
}

// handleSearch 语义检索
func (p *KnowledgePlugin) handleSearch(input *core.PluginInput) (*core.PluginOutput, error) {
	query, _ := input.Data["query"].(string)
	var teacherID int64
	if v, ok := input.Data["teacher_id"]; ok {
		teacherID = toInt64(v, 0)
	}
	limit := 5 // 默认返回 5 条
	if v, ok := input.Data["limit"]; ok {
		limit = toInt(v, 5)
	}

	// 参数校验
	if query == "" {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "查询内容不能为空",
		}, nil
	}

	if teacherID <= 0 {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "缺少 teacher_id",
		}, nil
	}

	// 从向量存储中检索（使用 VectorClient 调用 Python 服务）
	collectionName := fmt.Sprintf("teacher_%d", teacherID)
	results := p.vectorClient.Search(collectionName, query, limit)

	// 转换为输出格式
	var chunksOutput []map[string]interface{}
	for _, r := range results {
		chunksOutput = append(chunksOutput, map[string]interface{}{
			"content":     r.Content,
			"score":       r.Score,
			"document_id": r.DocumentID,
			"title":       r.Title,
		})
	}

	if chunksOutput == nil {
		chunksOutput = []map[string]interface{}{}
	}

	// 构建输出数据，merge 上游 Data
	outputData := mergeData(input.Data, map[string]interface{}{
		"chunks": chunksOutput,
	})

	return &core.PluginOutput{
		Success:  true,
		Data:     outputData,
		Metadata: map[string]interface{}{"plugin": "knowledge", "action": "search"},
	}, nil
}

// handleList 文档列表
func (p *KnowledgePlugin) handleList(input *core.PluginInput) (*core.PluginOutput, error) {
	var teacherID int64
	if v, ok := input.Data["teacher_id"]; ok {
		teacherID = toInt64(v, 0)
	}
	page := 1
	if v, ok := input.Data["page"]; ok {
		page = toInt(v, 1)
	}
	pageSize := 10
	if v, ok := input.Data["page_size"]; ok {
		pageSize = toInt(v, 10)
	}

	if teacherID <= 0 {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "缺少 teacher_id",
		}, nil
	}

	offset := (page - 1) * pageSize
	docs, total, err := p.docRepo.GetByTeacherID(teacherID, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询文档列表失败: %w", err)
	}

	// 转换为输出格式
	var docsOutput []map[string]interface{}
	for _, doc := range docs {
		docsOutput = append(docsOutput, map[string]interface{}{
			"id":         doc.ID,
			"title":      doc.Title,
			"doc_type":   doc.DocType,
			"tags":       doc.Tags,
			"status":     doc.Status,
			"created_at": doc.CreatedAt.Format(time.RFC3339),
			"updated_at": doc.UpdatedAt.Format(time.RFC3339),
		})
	}

	if docsOutput == nil {
		docsOutput = []map[string]interface{}{}
	}

	// 构建输出数据，merge 上游 Data
	outputData := mergeData(input.Data, map[string]interface{}{
		"documents": docsOutput,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})

	return &core.PluginOutput{
		Success:  true,
		Data:     outputData,
		Metadata: map[string]interface{}{"plugin": "knowledge", "action": "list"},
	}, nil
}

// handleDelete 删除文档
func (p *KnowledgePlugin) handleDelete(input *core.PluginInput) (*core.PluginOutput, error) {
	var documentID int64
	if v, ok := input.Data["document_id"]; ok {
		documentID = toInt64(v, 0)
	}

	if documentID <= 0 {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "缺少 document_id",
		}, nil
	}

	// 查询文档获取 teacher_id
	doc, err := p.docRepo.GetByID(documentID)
	if err != nil {
		return nil, fmt.Errorf("查询文档失败: %w", err)
	}
	if doc == nil {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40005},
			Error:   "文档不存在",
		}, nil
	}

	// 删除 SQLite 记录（软删除）
	if err := p.docRepo.Delete(documentID); err != nil {
		return nil, fmt.Errorf("删除文档失败: %w", err)
	}

	// 删除向量存储中的数据（使用 VectorClient 调用 Python 服务）
	collectionName := fmt.Sprintf("teacher_%d", doc.TeacherID)
	if err := p.vectorClient.DeleteByDocumentID(collectionName, documentID); err != nil {
		return nil, fmt.Errorf("删除向量数据失败: %w", err)
	}

	// 构建输出数据，merge 上游 Data
	outputData := mergeData(input.Data, map[string]interface{}{
		"document_id": documentID,
		"deleted":     true,
	})

	return &core.PluginOutput{
		Success:  true,
		Data:     outputData,
		Metadata: map[string]interface{}{"plugin": "knowledge", "action": "delete"},
	}, nil
}

// handlePreview 预览文档（只执行切片，不入库）
func (p *KnowledgePlugin) handlePreview(input *core.PluginInput) (*core.PluginOutput, error) {
	content, _ := input.Data["content"].(string)
	title, _ := input.Data["title"].(string)
	tags, _ := input.Data["tags"].(string)

	if content == "" {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "内容不能为空",
		}, nil
	}

	// 只执行切片，不入库
	chunks := p.chunker.Chunk(content)

	var chunksOutput []map[string]interface{}
	for i, chunk := range chunks {
		chunkRunes := []rune(chunk)
		chunksOutput = append(chunksOutput, map[string]interface{}{
			"index":   i,
			"content": chunk,
			"length":  len(chunkRunes),
		})
	}
	if chunksOutput == nil {
		chunksOutput = []map[string]interface{}{}
	}

	outputData := mergeData(input.Data, map[string]interface{}{
		"title":        title,
		"tags":         tags,
		"chunks":       chunksOutput,
		"chunks_count": len(chunks),
		"content":      content,
	})

	return &core.PluginOutput{
		Success:  true,
		Data:     outputData,
		Metadata: map[string]interface{}{"plugin": "knowledge", "action": "preview"},
	}, nil
}

// mergeData 合并上游 Data 和本插件输出字段
func mergeData(upstream map[string]interface{}, pluginData map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range upstream {
		result[k] = v
	}
	for k, v := range pluginData {
		result[k] = v
	}
	return result
}

// toInt 将 interface{} 转换为 int
func toInt(v interface{}, defaultVal int) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case string:
		// 不做字符串转换，返回默认值
		return defaultVal
	default:
		return defaultVal
	}
}

// toInt64 将 interface{} 转换为 int64
func toInt64(v interface{}, defaultVal int64) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case float64:
		return int64(val)
	case string:
		var result int64
		fmt.Sscanf(val, "%d", &result)
		if result > 0 {
			return result
		}
		return defaultVal
	default:
		return defaultVal
	}
}
