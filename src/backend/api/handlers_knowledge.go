package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"digital-twin/src/backend/database"
	"digital-twin/src/harness/core"
	"digital-twin/src/plugins/knowledge"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// validScopes 合法的 scope 值
var validScopes = map[string]bool{
	"global":  true,
	"class":   true,
	"student": true,
}

// ======================== 知识库接口 ========================

// HandleAddDocument 添加文档
// POST /api/documents
// V2.0 迭代3: 支持 scope_ids 多选
func (h *Handler) HandleAddDocument(c *gin.Context) {
	var req struct {
		Title    string  `json:"title" binding:"required"`
		Content  string  `json:"content" binding:"required"`
		Tags     string  `json:"tags"`
		Scope    string  `json:"scope"`     // global / class / student，默认 global
		ScopeID  int64   `json:"scope_id"`  // 向后兼容
		ScopeIDs []int64 `json:"scope_ids"` // V2.0 迭代3 新增，优先于 scope_id
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
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
					"title":      req.Title,
					"content":    req.Content,
					"tags":       req.Tags,
					"teacher_id": userIDInt64,
					"scope":      req.Scope,
					"scope_id":   sid,
					"persona_id": personaIDInt64,
				},
				Context: c.Request.Context(),
			}

			output, err := plugin.Execute(c.Request.Context(), input)
			if err != nil {
				Error(c, http.StatusInternalServerError, 50001, "添加文档失败: "+err.Error())
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
				"title":        req.Title,
				"scope":        req.Scope,
				"scope_id":     sid,
				"status":       "active",
			})
		}

		Success(c, gin.H{
			"documents": documents,
		})
		return
	}

	// 单个 scope_id，保持旧格式
	input := &core.PluginInput{
		RequestID:   uuid.New().String(),
		UserContext: userContext,
		Data: map[string]interface{}{
			"action":     "add",
			"title":      req.Title,
			"content":    req.Content,
			"tags":       req.Tags,
			"teacher_id": userIDInt64,
			"scope":      req.Scope,
			"scope_id":   scopeIDs[0],
			"persona_id": personaIDInt64,
		},
		Context: c.Request.Context(),
	}

	output, err := plugin.Execute(c.Request.Context(), input)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "添加文档失败: "+err.Error())
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

	Success(c, gin.H{
		"document_id":  output.Data["document_id"],
		"chunks_count": output.Data["chunks_count"],
		"title":        req.Title,
		"status":       "active",
	})
}

// HandleGetDocuments 获取文档列表
// GET /api/documents
func (h *Handler) HandleGetDocuments(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	// 从 JWT 获取 persona_id
	personaID, _ := c.Get("persona_id")
	personaIDInt64, _ := personaID.(int64)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 新增 scope 筛选参数
	scopeFilter := c.Query("scope")

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	personaRepo := database.NewPersonaRepository(db)
	docRepo := database.NewDocumentRepository(db)
	classRepo := database.NewClassRepository(db)

	// 判断当前分身角色
	if personaIDInt64 > 0 {
		persona, err := personaRepo.GetByID(personaIDInt64)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "查询分身失败: "+err.Error())
			return
		}

		if persona != nil && persona.Role == "student" {
			// 学生分身：返回与该学生相关的所有文档（需要找到关联的教师分身）
			relationRepo := database.NewRelationRepository(db)
			// 查询该学生分身的 approved 关系，获取教师分身ID
			rels, _, err := relationRepo.ListByStudent(persona.UserID, "approved", 0, 100)
			if err != nil {
				Error(c, http.StatusInternalServerError, 50001, "查询师生关系失败: "+err.Error())
				return
			}

			var allDocs []database.Document
			seen := make(map[int64]bool) // 去重

			for _, rel := range rels {
				teacherPersonaID := rel.TeacherPersonaID
				if teacherPersonaID <= 0 {
					continue
				}

				// 获取学生在该教师分身下所在的班级
				classIDs, err := classRepo.GetClassIDsByStudentPersona(teacherPersonaID, personaIDInt64)
				if err != nil {
					continue
				}

				// 获取学生可见的文档
				docs, err := docRepo.GetByStudentScope(teacherPersonaID, personaIDInt64, classIDs)
				if err != nil {
					continue
				}

				for _, doc := range docs {
					if !seen[doc.ID] {
						seen[doc.ID] = true
						allDocs = append(allDocs, doc)
					}
				}
			}

			// 分页
			total := len(allDocs)
			offset := (page - 1) * pageSize
			end := offset + pageSize
			if offset > total {
				offset = total
			}
			if end > total {
				end = total
			}
			pageDocs := allDocs[offset:end]

			var docsOutput []map[string]interface{}
			for _, doc := range pageDocs {
				docsOutput = append(docsOutput, map[string]interface{}{
					"id":         doc.ID,
					"title":      doc.Title,
					"doc_type":   doc.DocType,
					"tags":       doc.Tags,
					"status":     doc.Status,
					"scope":      doc.Scope,
					"scope_id":   doc.ScopeID,
					"persona_id": doc.PersonaID,
					"created_at": doc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
					"updated_at": doc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
				})
			}
			if docsOutput == nil {
				docsOutput = []map[string]interface{}{}
			}

			SuccessPage(c, docsOutput, total, page, pageSize)
			return
		}
	}

	// 教师分身或兼容旧逻辑：使用 persona_id 查询
	if personaIDInt64 > 0 {
		docs, err := docRepo.GetByPersonaID(personaIDInt64, scopeFilter)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "查询文档列表失败: "+err.Error())
			return
		}

		// 分页
		total := len(docs)
		offset := (page - 1) * pageSize
		end := offset + pageSize
		if offset > total {
			offset = total
		}
		if end > total {
			end = total
		}
		pageDocs := docs[offset:end]

		var docsOutput []map[string]interface{}
		for _, doc := range pageDocs {
			docsOutput = append(docsOutput, map[string]interface{}{
				"id":         doc.ID,
				"title":      doc.Title,
				"doc_type":   doc.DocType,
				"tags":       doc.Tags,
				"status":     doc.Status,
				"scope":      doc.Scope,
				"scope_id":   doc.ScopeID,
				"persona_id": doc.PersonaID,
				"created_at": doc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				"updated_at": doc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			})
		}
		if docsOutput == nil {
			docsOutput = []map[string]interface{}{}
		}

		SuccessPage(c, docsOutput, total, page, pageSize)
		return
	}

	// 兼容旧逻辑：没有 persona_id 时使用 teacher_id
	plugin, err := h.manager.GetPlugin("knowledge-retrieval")
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "知识库服务不可用")
		return
	}

	input := &core.PluginInput{
		RequestID: uuid.New().String(),
		Data: map[string]interface{}{
			"action":     "list",
			"teacher_id": userIDInt64,
			"page":       page,
			"page_size":  pageSize,
		},
		Context: c.Request.Context(),
	}

	output, err := plugin.Execute(c.Request.Context(), input)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询文档列表失败: "+err.Error())
		return
	}

	if !output.Success {
		errorCode := 50001
		if code, ok := output.Data["error_code"]; ok {
			errorCode = toInt(code, 50001)
		}
		Error(c, http.StatusInternalServerError, errorCode, output.Error)
		return
	}

	total := toInt(output.Data["total"], 0)
	SuccessPage(c, output.Data["documents"], total, page, pageSize)
}

// HandleDeleteDocument 删除文档
// DELETE /api/documents/:id
func (h *Handler) HandleDeleteDocument(c *gin.Context) {
	idStr := c.Param("id")
	docID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || docID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的文档 ID")
		return
	}

	plugin, err := h.manager.GetPlugin("knowledge-retrieval")
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "知识库服务不可用")
		return
	}

	input := &core.PluginInput{
		RequestID: uuid.New().String(),
		Data: map[string]interface{}{
			"action":      "delete",
			"document_id": docID,
		},
		Context: c.Request.Context(),
	}

	output, err := plugin.Execute(c.Request.Context(), input)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "删除文档失败: "+err.Error())
		return
	}

	if !output.Success {
		errorCode := 50001
		httpStatus := http.StatusInternalServerError
		if code, ok := output.Data["error_code"]; ok {
			errorCode = toInt(code, 50001)
			if errorCode == 40005 {
				httpStatus = http.StatusNotFound
			}
		}
		Error(c, httpStatus, errorCode, output.Error)
		return
	}

	Success(c, gin.H{
		"document_id": docID,
		"deleted":     true,
	})
}

// HandleUploadDocument 文件上传
// POST /api/documents/upload
// V2.0 迭代3: 支持 scope_ids 多选
func (h *Handler) HandleUploadDocument(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	personaID, _ := c.Get("persona_id")
	personaIDInt64, _ := personaID.(int64)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		Error(c, http.StatusBadRequest, 40010, "缺少文件参数")
		return
	}
	defer file.Close()

	title := c.PostForm("title")
	tags := c.PostForm("tags")

	scope := c.PostForm("scope")
	if scope == "" {
		scope = "global"
	}
	if !validScopes[scope] {
		Error(c, http.StatusBadRequest, 40004, "无效的 scope，仅支持 global/class/student")
		return
	}

	// V2.0 迭代3: 解析 scope_ids（逗号分隔字符串）
	var scopeIDs []int64
	scopeIDsStr := c.PostForm("scope_ids")
	if scopeIDsStr != "" {
		parts := strings.Split(scopeIDsStr, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			id, err := strconv.ParseInt(p, 10, 64)
			if err != nil {
				Error(c, http.StatusBadRequest, 40004, "scope_ids 参数格式无效")
				return
			}
			scopeIDs = append(scopeIDs, id)
		}
	}
	// 向后兼容 scope_id
	if len(scopeIDs) == 0 {
		var scopeID int64
		if scopeIDStr := c.PostForm("scope_id"); scopeIDStr != "" {
			scopeID, _ = strconv.ParseInt(scopeIDStr, 10, 64)
		}
		if scopeID > 0 {
			scopeIDs = []int64{scopeID}
		} else {
			scopeIDs = []int64{0}
		}
	}

	// 校验每个 scope_id
	for _, sid := range scopeIDs {
		if err := h.validateScope(c, personaIDInt64, scope, sid); err != nil {
			return
		}
	}

	// 3. 校验文件格式（后缀：.pdf/.docx/.txt/.md）
	filename := header.Filename
	ext := strings.ToLower(filepath.Ext(filename))
	supportedFormats := map[string]bool{
		".pdf": true, ".docx": true, ".txt": true, ".md": true,
	}
	if !supportedFormats[ext] {
		Error(c, http.StatusBadRequest, 40010, "不支持的文件格式，仅支持 PDF/DOCX/TXT/MD")
		return
	}

	// 4. 校验文件大小（≤ 50MB = 52428800 bytes）
	if header.Size > 52428800 {
		Error(c, http.StatusBadRequest, 40011, "文件大小超过限制（最大 50MB）")
		return
	}

	// 5. 保存文件到 {UPLOAD_DIR}/documents/{teacher_id}/{uuid}_{filename}
	baseUploadDir := os.Getenv("UPLOAD_DIR")
	if baseUploadDir == "" {
		baseUploadDir = "./uploads"
	}
	teacherDir := filepath.Join(baseUploadDir, "documents", fmt.Sprintf("%d", userIDInt64))
	if err := os.MkdirAll(teacherDir, 0755); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "创建上传目录失败: "+err.Error())
		return
	}

	fileUUID := uuid.New().String()
	savedFilename := fileUUID + "_" + filename
	filePath := filepath.Join(teacherDir, savedFilename)

	if err := c.SaveUploadedFile(header, filePath); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "保存文件失败: "+err.Error())
		return
	}

	// 6. 调用 FileParser.Parse(filePath) 提取文本内容
	parser := knowledge.NewFileParser()
	content, err := parser.Parse(filePath)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "解析文件内容失败: "+err.Error())
		return
	}

	// 7. 自动填充 title（如果未传，使用文件名去掉后缀）
	if title == "" {
		title = strings.TrimSuffix(filename, ext)
	}

	// 文档类型（去掉前缀点号）
	docType := strings.TrimPrefix(ext, ".")

	// 8. 调用 knowledge 插件 add action（分块 + 向量化）
	plugin, err := h.manager.GetPlugin("knowledge-retrieval")
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "知识库服务不可用")
		return
	}

	userContext := &core.UserContext{
		UserID: fmt.Sprintf("%d", userIDInt64),
	}

	// V2.0 迭代3: 多个 scope_id 时为每个创建一条文档
	if len(scopeIDs) > 1 {
		var documents []gin.H
		for _, sid := range scopeIDs {
			input := &core.PluginInput{
				RequestID:   uuid.New().String(),
				UserContext: userContext,
				Data: map[string]interface{}{
					"action":     "add",
					"title":      title,
					"content":    content,
					"tags":       tags,
					"teacher_id": userIDInt64,
					"doc_type":   docType,
					"scope":      scope,
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
				Error(c, http.StatusInternalServerError, errorCode, output.Error)
				return
			}

			documents = append(documents, gin.H{
				"document_id":  output.Data["document_id"],
				"chunks_count": output.Data["chunks_count"],
				"title":        title,
				"doc_type":     docType,
				"scope":        scope,
				"scope_id":     sid,
				"file_size":    header.Size,
				"status":       "active",
			})
		}

		Success(c, gin.H{"documents": documents})
		return
	}

	input := &core.PluginInput{
		RequestID:   uuid.New().String(),
		UserContext: userContext,
		Data: map[string]interface{}{
			"action":     "add",
			"title":      title,
			"content":    content,
			"tags":       tags,
			"teacher_id": userIDInt64,
			"doc_type":   docType,
			"scope":      scope,
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
		Error(c, http.StatusInternalServerError, errorCode, output.Error)
		return
	}

	// 9. 返回结果
	Success(c, gin.H{
		"document_id":  output.Data["document_id"],
		"chunks_count": output.Data["chunks_count"],
		"title":        title,
		"doc_type":     docType,
		"file_size":    header.Size,
		"status":       "active",
	})
}

// HandleImportURL URL 导入
// POST /api/documents/import-url
// V2.0 迭代3: 支持 scope_ids 多选
func (h *Handler) HandleImportURL(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	personaID, _ := c.Get("persona_id")
	personaIDInt64, _ := personaID.(int64)

	var req struct {
		URL      string  `json:"url" binding:"required"`
		Title    string  `json:"title"`
		Tags     string  `json:"tags"`
		Scope    string  `json:"scope"`
		ScopeID  int64   `json:"scope_id"`
		ScopeIDs []int64 `json:"scope_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

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
		scopeIDs = []int64{0}
	}

	for _, sid := range scopeIDs {
		if err := h.validateScope(c, personaIDInt64, req.Scope, sid); err != nil {
			return
		}
	}

	// 3. 校验 URL 格式（必须以 http:// 或 https:// 开头）
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		Error(c, http.StatusBadRequest, 40004, "无效的 URL 格式，必须以 http:// 或 https:// 开头")
		return
	}

	// 4. 调用 URLFetcher.Fetch(url) 抓取内容
	fetcher := knowledge.NewURLFetcher()
	fetchedTitle, content, err := fetcher.Fetch(req.URL)
	if err != nil {
		Error(c, http.StatusBadRequest, 40012, "URL 内容抓取失败: "+err.Error())
		return
	}

	// 5. 如果 title 未传，使用抓取到的 title
	title := req.Title
	if title == "" {
		title = fetchedTitle
	}

	// 6. 如果 title 仍为空，使用 URL 作为 title
	if title == "" {
		title = req.URL
	}

	// 7. 调用 knowledge 插件 add action
	plugin, err := h.manager.GetPlugin("knowledge-retrieval")
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "知识库服务不可用")
		return
	}

	userContext := &core.UserContext{
		UserID: fmt.Sprintf("%d", userIDInt64),
	}

	// V2.0 迭代3: 多个 scope_id 时为每个创建一条文档
	if len(scopeIDs) > 1 {
		var documents []gin.H
		for _, sid := range scopeIDs {
			input := &core.PluginInput{
				RequestID:   uuid.New().String(),
				UserContext: userContext,
				Data: map[string]interface{}{
					"action":     "add",
					"title":      title,
					"content":    content,
					"tags":       req.Tags,
					"teacher_id": userIDInt64,
					"doc_type":   "url",
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
				Error(c, http.StatusInternalServerError, errorCode, output.Error)
				return
			}

			contentRunes := []rune(content)
			documents = append(documents, gin.H{
				"document_id":    output.Data["document_id"],
				"chunks_count":   output.Data["chunks_count"],
				"title":          title,
				"doc_type":       "url",
				"scope":          req.Scope,
				"scope_id":       sid,
				"content_length": len(contentRunes),
				"source_url":     req.URL,
				"status":         "active",
			})
		}

		Success(c, gin.H{"documents": documents})
		return
	}

	input := &core.PluginInput{
		RequestID:   uuid.New().String(),
		UserContext: userContext,
		Data: map[string]interface{}{
			"action":     "add",
			"title":      title,
			"content":    content,
			"tags":       req.Tags,
			"teacher_id": userIDInt64,
			"doc_type":   "url",
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
		Error(c, http.StatusInternalServerError, errorCode, output.Error)
		return
	}

	// 8. 返回结果
	contentRunes := []rune(content)
	Success(c, gin.H{
		"document_id":    output.Data["document_id"],
		"chunks_count":   output.Data["chunks_count"],
		"title":          title,
		"doc_type":       "url",
		"content_length": len(contentRunes),
		"source_url":     req.URL,
		"status":         "active",
	})
}

// validateScope 校验 scope 和 scope_id 的合法性
// 返回 error 非 nil 时表示已向客户端返回错误响应
func (h *Handler) validateScope(c *gin.Context, personaIDInt64 int64, scope string, scopeID int64) error {
	if scope == "global" {
		return nil
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return fmt.Errorf("数据库不可用")
	}

	if scope == "class" {
		if scopeID <= 0 {
			Error(c, http.StatusBadRequest, 40004, "scope=class 时 scope_id（班级ID）不能为空")
			return fmt.Errorf("scope_id 无效")
		}
		// 校验班级存在且属于当前教师分身
		classRepo := database.NewClassRepository(db)
		class, err := classRepo.GetByID(scopeID)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "查询班级失败: "+err.Error())
			return fmt.Errorf("查询班级失败")
		}
		if class == nil {
			Error(c, http.StatusNotFound, 40005, "班级不存在")
			return fmt.Errorf("班级不存在")
		}
		if class.PersonaID != personaIDInt64 {
			Error(c, http.StatusForbidden, 40003, "该班级不属于当前教师分身")
			return fmt.Errorf("班级权限不足")
		}
		return nil
	}

	if scope == "student" {
		if scopeID <= 0 {
			Error(c, http.StatusBadRequest, 40004, "scope=student 时 scope_id（学生分身ID）不能为空")
			return fmt.Errorf("scope_id 无效")
		}
		// 校验学生分身存在且与当前教师分身有 approved 关系
		personaRepo := database.NewPersonaRepository(db)
		studentPersona, err := personaRepo.GetByID(scopeID)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "查询学生分身失败: "+err.Error())
			return fmt.Errorf("查询学生分身失败")
		}
		if studentPersona == nil || studentPersona.Role != "student" {
			Error(c, http.StatusNotFound, 40005, "学生分身不存在")
			return fmt.Errorf("学生分身不存在")
		}
		// 检查 approved 关系
		relationRepo := database.NewRelationRepository(db)
		approved, err := relationRepo.IsApprovedByPersonas(personaIDInt64, scopeID)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "查询师生关系失败: "+err.Error())
			return fmt.Errorf("查询师生关系失败")
		}
		if !approved {
			// 回退到 user_id 维度校验
			userID, _ := c.Get("user_id")
			userIDInt64, _ := userID.(int64)
			approved, err = relationRepo.IsApproved(userIDInt64, studentPersona.UserID)
			if err != nil {
				Error(c, http.StatusInternalServerError, 50001, "查询师生关系失败: "+err.Error())
				return fmt.Errorf("查询师生关系失败")
			}
			if !approved {
				Error(c, http.StatusForbidden, 40003, "与该学生分身没有授权关系")
				return fmt.Errorf("师生关系未授权")
			}
		}
		return nil
	}

	return nil
}
