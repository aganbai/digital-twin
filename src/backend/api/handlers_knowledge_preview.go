package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"digital-twin/src/harness/core"
	"digital-twin/src/plugins/knowledge"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ======================== V2.0 迭代3 知识库预览API ========================

// HandlePreviewDocument 文本预览
// POST /api/documents/preview
// 权限: teacher/admin
func (h *Handler) HandlePreviewDocument(c *gin.Context) {
	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
		Tags    string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	// 检查缓存大小
	if getPreviewCacheSize() >= previewCacheMax {
		Error(c, http.StatusTooManyRequests, 42900, "预览缓存已满，请稍后再试")
		return
	}

	// 调用 knowledge-retrieval 插件的 preview action 获取切片结果
	plugin, err := h.manager.GetPlugin("knowledge-retrieval")
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "知识库服务不可用")
		return
	}

	userID, _ := c.Get("user_id")
	userIDInt64, _ := userID.(int64)

	userContext := &core.UserContext{
		UserID: fmt.Sprintf("%d", userIDInt64),
	}

	input := &core.PluginInput{
		RequestID:   uuid.New().String(),
		UserContext: userContext,
		Data: map[string]interface{}{
			"action":  "preview",
			"title":   req.Title,
			"content": req.Content,
			"tags":    req.Tags,
		},
		Context: c.Request.Context(),
	}

	output, err := plugin.Execute(c.Request.Context(), input)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "预览失败: "+err.Error())
		return
	}

	if !output.Success {
		errorCode := 50001
		if code, ok := output.Data["error_code"]; ok {
			errorCode = toInt(code, 50001)
		}
		Error(c, http.StatusBadRequest, errorCode, output.Error)
		return
	}

	// 生成 preview_id
	previewID := "preview_" + uuid.New().String()

	// 获取切片结果
	chunks, _ := output.Data["chunks"].([]map[string]interface{})
	if chunks == nil {
		chunks = []map[string]interface{}{}
	}

	// 存入缓存
	entry := &previewCacheEntry{
		Content:   req.Content,
		Title:     req.Title,
		Tags:      req.Tags,
		Chunks:    chunks,
		DocType:   "text",
		CreatedAt: time.Now(),
	}
	previewCache.Store(previewID, entry)

	// 确保清理 goroutine 启动
	initPreviewCacheCleanup()

	// 🆕 V2.0 迭代4：LLM 智能摘要
	llmTitle := ""
	llmSummary := ""
	summarizer := knowledge.NewLLMSummarizer()
	summaryResult, summaryErr := summarizer.Summarize(c.Request.Context(), req.Content)
	if summaryErr == nil && summaryResult != nil {
		llmTitle = summaryResult.Title
		llmSummary = summaryResult.Summary
	}

	contentRunes := []rune(req.Content)
	Success(c, gin.H{
		"preview_id":  previewID,
		"title":       req.Title,
		"tags":        req.Tags,
		"total_chars": len(contentRunes),
		"chunks":      chunks,
		"chunk_count": len(chunks),
		"llm_title":   llmTitle,
		"llm_summary": llmSummary,
	})
}

// HandlePreviewUpload 文件上传预览
// POST /api/documents/preview-upload
// 权限: teacher/admin
func (h *Handler) HandlePreviewUpload(c *gin.Context) {
	// 检查缓存大小
	if getPreviewCacheSize() >= previewCacheMax {
		Error(c, http.StatusTooManyRequests, 42900, "预览缓存已满，请稍后再试")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		Error(c, http.StatusBadRequest, 40010, "缺少文件参数")
		return
	}
	defer file.Close()

	title := c.PostForm("title")
	tags := c.PostForm("tags")

	// 校验文件格式
	filename := header.Filename
	ext := strings.ToLower(filepath.Ext(filename))
	supportedFormats := map[string]bool{
		".pdf": true, ".docx": true, ".txt": true, ".md": true,
	}
	if !supportedFormats[ext] {
		Error(c, http.StatusBadRequest, 40010, "不支持的文件格式，仅支持 PDF/DOCX/TXT/MD")
		return
	}

	// 校验文件大小（≤ 50MB）
	if header.Size > 52428800 {
		Error(c, http.StatusBadRequest, 40011, "文件大小超过限制（最大 50MB）")
		return
	}

	// 保存文件到临时目录
	tmpDir := os.TempDir()
	fileUUID := uuid.New().String()
	filePath := filepath.Join(tmpDir, fileUUID+"_"+filename)
	if err := c.SaveUploadedFile(header, filePath); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "保存临时文件失败: "+err.Error())
		return
	}
	defer os.Remove(filePath) // 预览完成后清理临时文件

	// 使用 FileParser 提取文本
	parser := knowledge.NewFileParser()
	content, err := parser.Parse(filePath)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "解析文件内容失败: "+err.Error())
		return
	}

	// 自动填充 title
	if title == "" {
		title = strings.TrimSuffix(filename, ext)
	}

	docType := strings.TrimPrefix(ext, ".")

	// 调用 knowledge-retrieval 插件的 preview action
	plugin, err := h.manager.GetPlugin("knowledge-retrieval")
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "知识库服务不可用")
		return
	}

	userID, _ := c.Get("user_id")
	userIDInt64, _ := userID.(int64)

	userContext := &core.UserContext{
		UserID: fmt.Sprintf("%d", userIDInt64),
	}

	input := &core.PluginInput{
		RequestID:   uuid.New().String(),
		UserContext: userContext,
		Data: map[string]interface{}{
			"action":  "preview",
			"title":   title,
			"content": content,
			"tags":    tags,
		},
		Context: c.Request.Context(),
	}

	output, err := plugin.Execute(c.Request.Context(), input)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "预览失败: "+err.Error())
		return
	}

	if !output.Success {
		errorCode := 50001
		if code, ok := output.Data["error_code"]; ok {
			errorCode = toInt(code, 50001)
		}
		Error(c, http.StatusBadRequest, errorCode, output.Error)
		return
	}

	// 生成 preview_id
	previewID := "preview_" + uuid.New().String()

	// 获取切片结果
	chunks, _ := output.Data["chunks"].([]map[string]interface{})
	if chunks == nil {
		chunks = []map[string]interface{}{}
	}

	// 存入缓存
	entry := &previewCacheEntry{
		Content:   content,
		Title:     title,
		Tags:      tags,
		Chunks:    chunks,
		DocType:   docType,
		CreatedAt: time.Now(),
	}
	previewCache.Store(previewID, entry)

	// 确保清理 goroutine 启动
	initPreviewCacheCleanup()

	// 🆕 V2.0 迭代4：LLM 智能摘要
	llmTitle := ""
	llmSummary := ""
	summarizer := knowledge.NewLLMSummarizer()
	summaryResult, summaryErr := summarizer.Summarize(c.Request.Context(), content)
	if summaryErr == nil && summaryResult != nil {
		llmTitle = summaryResult.Title
		llmSummary = summaryResult.Summary
	}

	contentRunes := []rune(content)
	Success(c, gin.H{
		"preview_id":  previewID,
		"title":       title,
		"tags":        tags,
		"doc_type":    docType,
		"total_chars": len(contentRunes),
		"chunks":      chunks,
		"chunk_count": len(chunks),
		"llm_title":   llmTitle,
		"llm_summary": llmSummary,
	})
}

// HandlePreviewURL URL 导入预览
// POST /api/documents/preview-url
// 权限: teacher/admin
func (h *Handler) HandlePreviewURL(c *gin.Context) {
	var req struct {
		URL   string `json:"url" binding:"required"`
		Title string `json:"title"`
		Tags  string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	// 检查缓存大小
	if getPreviewCacheSize() >= previewCacheMax {
		Error(c, http.StatusTooManyRequests, 42900, "预览缓存已满，请稍后再试")
		return
	}

	// 校验 URL 格式
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		Error(c, http.StatusBadRequest, 40004, "无效的 URL 格式，必须以 http:// 或 https:// 开头")
		return
	}

	// 使用 URLFetcher 抓取内容
	fetcher := knowledge.NewURLFetcher()
	fetchedTitle, content, err := fetcher.Fetch(req.URL)
	if err != nil {
		Error(c, http.StatusBadRequest, 40012, "URL 内容抓取失败: "+err.Error())
		return
	}

	// 如果 title 未传，使用抓取到的 title
	title := req.Title
	if title == "" {
		title = fetchedTitle
	}
	if title == "" {
		title = req.URL
	}

	// 调用 knowledge-retrieval 插件的 preview action
	plugin, err := h.manager.GetPlugin("knowledge-retrieval")
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "知识库服务不可用")
		return
	}

	userID, _ := c.Get("user_id")
	userIDInt64, _ := userID.(int64)

	userContext := &core.UserContext{
		UserID: fmt.Sprintf("%d", userIDInt64),
	}

	input := &core.PluginInput{
		RequestID:   uuid.New().String(),
		UserContext: userContext,
		Data: map[string]interface{}{
			"action":  "preview",
			"title":   title,
			"content": content,
			"tags":    req.Tags,
		},
		Context: c.Request.Context(),
	}

	output, err := plugin.Execute(c.Request.Context(), input)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "预览失败: "+err.Error())
		return
	}

	if !output.Success {
		errorCode := 50001
		if code, ok := output.Data["error_code"]; ok {
			errorCode = toInt(code, 50001)
		}
		Error(c, http.StatusBadRequest, errorCode, output.Error)
		return
	}

	// 生成 preview_id
	previewID := "preview_" + uuid.New().String()

	// 获取切片结果
	chunks, _ := output.Data["chunks"].([]map[string]interface{})
	if chunks == nil {
		chunks = []map[string]interface{}{}
	}

	// 存入缓存
	entry := &previewCacheEntry{
		Content:   content,
		Title:     title,
		Tags:      req.Tags,
		Chunks:    chunks,
		DocType:   "url",
		SourceURL: req.URL,
		CreatedAt: time.Now(),
	}
	previewCache.Store(previewID, entry)

	// 确保清理 goroutine 启动
	initPreviewCacheCleanup()

	// 🆕 V2.0 迭代4：LLM 智能摘要
	llmTitle := ""
	llmSummary := ""
	summarizer := knowledge.NewLLMSummarizer()
	summaryResult, summaryErr := summarizer.Summarize(c.Request.Context(), content)
	if summaryErr == nil && summaryResult != nil {
		llmTitle = summaryResult.Title
		llmSummary = summaryResult.Summary
	}

	contentRunes := []rune(content)
	Success(c, gin.H{
		"preview_id":  previewID,
		"title":       title,
		"tags":        req.Tags,
		"doc_type":    "url",
		"source_url":  req.URL,
		"total_chars": len(contentRunes),
		"chunks":      chunks,
		"chunk_count": len(chunks),
		"llm_title":   llmTitle,
		"llm_summary": llmSummary,
	})
}

// HandleConfirmDocument 确认入库
// POST /api/documents/confirm
// 权限: teacher/admin
func (h *Handler) HandleConfirmDocument(c *gin.Context) {
	var req struct {
		PreviewID string  `json:"preview_id" binding:"required"`
		Title     string  `json:"title"`
		Tags      string  `json:"tags"`
		Summary   string  `json:"summary"`
		Scope     string  `json:"scope"`
		ScopeID   int64   `json:"scope_id"`
		ScopeIDs  []int64 `json:"scope_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	// 从缓存中取出预览条目
	val, ok := previewCache.Load(req.PreviewID)
	if !ok {
		Error(c, http.StatusBadRequest, 40028, "预览 ID 无效或已过期")
		return
	}

	entry, ok := val.(*previewCacheEntry)
	if !ok || time.Since(entry.CreatedAt) > previewCacheExpiry {
		previewCache.Delete(req.PreviewID)
		Error(c, http.StatusBadRequest, 40028, "预览 ID 无效或已过期")
		return
	}

	// 允许用户在确认时修改 title 和 tags
	title := entry.Title
	if req.Title != "" {
		title = req.Title
	}
	tags := entry.Tags
	if req.Tags != "" {
		tags = req.Tags
	}

	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	personaID, _ := c.Get("persona_id")
	personaIDInt64, _ := personaID.(int64)

	if req.Scope == "" {
		req.Scope = "global"
	}
	if !validScopes[req.Scope] {
		Error(c, http.StatusBadRequest, 40004, "无效的 scope，仅支持 global/class/student")
		return
	}

	// scope_ids 向后兼容处理
	scopeIDs := req.ScopeIDs
	if len(scopeIDs) == 0 && req.ScopeID > 0 {
		scopeIDs = []int64{req.ScopeID}
	}
	if len(scopeIDs) == 0 {
		scopeIDs = []int64{0} // global
	}

	// 校验每个 scope_id
	for _, sid := range scopeIDs {
		if err := h.validateScope(c, personaIDInt64, req.Scope, sid); err != nil {
			return
		}
	}

	plugin, err := h.manager.GetPlugin("knowledge-retrieval")
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "知识库服务不可用")
		return
	}

	userContext := &core.UserContext{
		UserID: fmt.Sprintf("%d", userIDInt64),
	}

	// 多个 scope_id 时，为每个创建一条文档记录
	if len(scopeIDs) > 1 {
		var documents []gin.H
		for _, sid := range scopeIDs {
			input := &core.PluginInput{
				RequestID:   uuid.New().String(),
				UserContext: userContext,
				Data: map[string]interface{}{
					"action":     "add",
					"title":      title,
					"content":    entry.Content,
					"tags":       tags,
					"summary":    req.Summary,
					"teacher_id": userIDInt64,
					"doc_type":   entry.DocType,
					"scope":      req.Scope,
					"scope_id":   sid,
					"persona_id": personaIDInt64,
				},
				Context: c.Request.Context(),
			}

			output, err := plugin.Execute(c.Request.Context(), input)
			if err != nil {
				Error(c, http.StatusInternalServerError, 50001, "文档入库失败: "+err.Error())
				return
			}
			if !output.Success {
				errorCode := 50001
				if code, ok := output.Data["error_code"]; ok {
					errorCode = toInt(code, 50001)
				}
				Error(c, http.StatusBadRequest, errorCode, output.Error)
				return
			}

			documents = append(documents, gin.H{
				"document_id":  output.Data["document_id"],
				"chunks_count": output.Data["chunks_count"],
				"title":        title,
				"scope":        req.Scope,
				"scope_id":     sid,
				"status":       "active",
			})
		}

		// 入库成功后删除缓存
		previewCache.Delete(req.PreviewID)

		Success(c, gin.H{
			"documents": documents,
		})
		return
	}

	// 单个 scope_id
	input := &core.PluginInput{
		RequestID:   uuid.New().String(),
		UserContext: userContext,
		Data: map[string]interface{}{
			"action":     "add",
			"title":      title,
			"content":    entry.Content,
			"tags":       tags,
			"summary":    req.Summary,
			"teacher_id": userIDInt64,
			"doc_type":   entry.DocType,
			"scope":      req.Scope,
			"scope_id":   scopeIDs[0],
			"persona_id": personaIDInt64,
		},
		Context: c.Request.Context(),
	}

	output, err := plugin.Execute(c.Request.Context(), input)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "文档入库失败: "+err.Error())
		return
	}

	if !output.Success {
		errorCode := 50001
		if code, ok := output.Data["error_code"]; ok {
			errorCode = toInt(code, 50001)
		}
		Error(c, http.StatusBadRequest, errorCode, output.Error)
		return
	}

	// 入库成功后删除缓存
	previewCache.Delete(req.PreviewID)

	Success(c, gin.H{
		"document_id":  output.Data["document_id"],
		"chunks_count": output.Data["chunks_count"],
		"title":        title,
		"status":       "active",
	})
}
