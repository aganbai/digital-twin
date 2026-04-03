package api

import (
	"digital-twin/src/backend/database"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SmartUploadRequest 智能上传请求
type SmartUploadRequest struct {
	Type      string   `json:"type" binding:"required"` // url / text / file
	URL       string   `json:"url,omitempty"`
	Content   string   `json:"content,omitempty"`
	FileURLs  []string `json:"file_urls,omitempty"`
	Title     string   `json:"title,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Scope     string   `json:"scope,omitempty"`      // global / class / student
	ScopeID   int64    `json:"scope_id,omitempty"`   // 班级ID或学生ID
	PersonaID int64    `json:"persona_id,omitempty"` // 分身ID
}

// SmartUploadResponse 智能上传响应
type SmartUploadResponse struct {
	Items []KnowledgeItemResult `json:"items"`
}

// KnowledgeItemResult 知识库条目结果
type KnowledgeItemResult struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Type    string `json:"type"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// HandleSmartUpload 智能知识库上传（URL/文字/文件统一入口）
func (h *Handler) HandleSmartUpload(c *gin.Context) {
	userID, _ := c.Get("user_id")
	teacherID := userID.(int64)

	var req SmartUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "请求参数错误: " + err.Error()})
		return
	}

	// 如果没有指定persona_id，获取默认分身
	if req.PersonaID == 0 {
		defaultPersonaID, err := h.getDefaultPersonaID(teacherID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "获取默认分身失败"})
			return
		}
		req.PersonaID = defaultPersonaID
	}

	knowledgeRepo := database.NewKnowledgeRepository(h.db)
	var results []KnowledgeItemResult

	switch req.Type {
	case "url":
		result, err := h.processURLUpload(knowledgeRepo, teacherID, req)
		if err != nil {
			result.Message = err.Error()
		}
		results = append(results, result)

	case "text":
		result, err := h.processTextUpload(knowledgeRepo, teacherID, req)
		if err != nil {
			result.Message = err.Error()
		}
		results = append(results, result)

	case "file":
		for _, fileURL := range req.FileURLs {
			result, err := h.processFileUpload(knowledgeRepo, teacherID, fileURL, req)
			if err != nil {
				result.Message = err.Error()
			}
			results = append(results, result)
		}

	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "不支持的类型: " + req.Type})
		return
	}

	c.JSON(http.StatusOK, SmartUploadResponse{Items: results})
}

// processURLUpload 处理URL上传
func (h *Handler) processURLUpload(repo *database.KnowledgeRepository, teacherID int64, req SmartUploadRequest) (KnowledgeItemResult, error) {
	result := KnowledgeItemResult{Type: "url", Status: "processing"}

	// 解析URL
	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		result.Status = "failed"
		return result, fmt.Errorf("URL解析失败")
	}

	// 获取URL内容（简化实现，实际可能需要爬虫服务）
	title := req.Title
	if title == "" {
		title = parsedURL.Host + parsedURL.Path
	}

	content := fmt.Sprintf("URL来源: %s\n待解析内容...", req.URL)

	item := &database.KnowledgeItem{
		TeacherID: teacherID,
		PersonaID: req.PersonaID,
		Title:     title,
		Content:   content,
		ItemType:  "url",
		SourceURL: req.URL,
		Tags:      database.TagsToJSON(req.Tags),
		Status:    "active",
		Scope:     req.Scope,
		ScopeID:   req.ScopeID,
	}

	id, err := repo.CreateKnowledgeItem(item)
	if err != nil {
		result.Status = "failed"
		return result, err
	}

	result.ID = id
	result.Title = title
	result.Status = "success"
	return result, nil
}

// processTextUpload 处理文字上传
func (h *Handler) processTextUpload(repo *database.KnowledgeRepository, teacherID int64, req SmartUploadRequest) (KnowledgeItemResult, error) {
	result := KnowledgeItemResult{Type: "text", Status: "processing"}

	if req.Content == "" {
		result.Status = "failed"
		return result, fmt.Errorf("内容不能为空")
	}

	title := req.Title
	if title == "" {
		// 超过100字取前20字+…，否则显示"教学笔记"
		runeContent := []rune(req.Content)
		if len(runeContent) > 100 {
			title = string(runeContent[:20]) + "…"
		} else {
			title = "教学笔记"
		}
	}

	item := &database.KnowledgeItem{
		TeacherID: teacherID,
		PersonaID: req.PersonaID,
		Title:     title,
		Content:   req.Content,
		ItemType:  "text",
		Tags:      database.TagsToJSON(req.Tags),
		Status:    "active",
		Scope:     req.Scope,
		ScopeID:   req.ScopeID,
	}

	id, err := repo.CreateKnowledgeItem(item)
	if err != nil {
		result.Status = "failed"
		return result, err
	}

	result.ID = id
	result.Title = title
	result.Status = "success"
	return result, nil
}

// processFileUpload 处理文件上传
func (h *Handler) processFileUpload(repo *database.KnowledgeRepository, teacherID int64, fileURL string, req SmartUploadRequest) (KnowledgeItemResult, error) {
	result := KnowledgeItemResult{Type: "file", Status: "processing"}

	// 解析文件名
	fileName := filepath.Base(fileURL)
	title := req.Title
	if title == "" {
		title = fileName
	}

	// 获取文件大小（简化处理）
	fileSize := int64(0)

	item := &database.KnowledgeItem{
		TeacherID: teacherID,
		PersonaID: req.PersonaID,
		Title:     title,
		Content:   "文件内容待解析...",
		ItemType:  "file",
		FileURL:   fileURL,
		FileName:  fileName,
		FileSize:  fileSize,
		Tags:      database.TagsToJSON(req.Tags),
		Status:    "processing", // 文件需要异步处理
		Scope:     req.Scope,
		ScopeID:   req.ScopeID,
	}

	id, err := repo.CreateKnowledgeItem(item)
	if err != nil {
		result.Status = "failed"
		return result, err
	}

	result.ID = id
	result.Title = title
	result.Status = "processing"
	return result, nil
}

// getDefaultPersonaID 获取用户默认分身ID
func (h *Handler) getDefaultPersonaID(userID int64) (int64, error) {
	// 查询用户默认分身
	var personaID int64
	err := h.db.DB.QueryRow(`
		SELECT default_persona_id FROM users WHERE id = ?`, userID).Scan(&personaID)
	if err != nil {
		return 0, err
	}
	if personaID == 0 {
		// 如果没有默认分身，获取第一个教师分身
		err = h.db.DB.QueryRow(`
			SELECT id FROM personas WHERE user_id = ? AND role = 'teacher' LIMIT 1`, userID).Scan(&personaID)
		if err != nil {
			return 0, err
		}
	}
	return personaID, nil
}

// SearchKnowledgeRequest 搜索知识库请求
type SearchKnowledgeRequest struct {
	Keyword  string `form:"keyword"`
	ItemType string `form:"item_type"`
	Scope    string `form:"scope"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
}

// SearchKnowledgeResponse 搜索知识库响应
type SearchKnowledgeResponse struct {
	Items      []database.KnowledgeItemListItem `json:"items"`
	Total      int64                            `json:"total"`
	Page       int                              `json:"page"`
	PageSize   int                              `json:"page_size"`
	TotalPages int                              `json:"total_pages"`
}

// HandleSearchKnowledge 搜索知识库
func (h *Handler) HandleSearchKnowledge(c *gin.Context) {
	userID, _ := c.Get("user_id")
	teacherID := userID.(int64)

	var req SearchKnowledgeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "请求参数错误"})
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	knowledgeRepo := database.NewKnowledgeRepository(h.db)
	offset := (req.Page - 1) * req.PageSize

	items, total, err := knowledgeRepo.SearchKnowledgeItems(
		teacherID, req.Keyword, req.ItemType, req.Scope, req.PageSize, offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "搜索失败: " + err.Error()})
		return
	}

	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	c.JSON(http.StatusOK, SearchKnowledgeResponse{
		Items:      items,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	})
}

// UpdateKnowledgeRequest 更新知识库请求
type UpdateKnowledgeRequest struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
	Scope   string   `json:"scope"`
	ScopeID int64    `json:"scope_id"`
}

// HandleUpdateKnowledge 更新知识库条目
func (h *Handler) HandleUpdateKnowledge(c *gin.Context) {
	userID, _ := c.Get("user_id")
	teacherID := userID.(int64)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "无效的ID"})
		return
	}

	var req UpdateKnowledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "请求参数错误"})
		return
	}

	knowledgeRepo := database.NewKnowledgeRepository(h.db)

	// 验证所有权
	item, err := knowledgeRepo.GetKnowledgeItemByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "查询失败"})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Code: 40401, Message: "知识库条目不存在"})
		return
	}
	if item.TeacherID != teacherID {
		c.JSON(http.StatusForbidden, ErrorResponse{Code: 40301, Message: "无权操作此条目"})
		return
	}

	// 更新字段
	if req.Title != "" {
		item.Title = req.Title
	}
	if req.Content != "" {
		item.Content = req.Content
	}
	if len(req.Tags) > 0 {
		item.Tags = database.TagsToJSON(req.Tags)
	}
	if req.Scope != "" {
		item.Scope = req.Scope
	}
	item.ScopeID = req.ScopeID

	if err := knowledgeRepo.UpdateKnowledgeItem(item); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "更新失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// HandleDeleteKnowledge 删除知识库条目
func (h *Handler) HandleDeleteKnowledge(c *gin.Context) {
	userID, _ := c.Get("user_id")
	teacherID := userID.(int64)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "无效的ID"})
		return
	}

	knowledgeRepo := database.NewKnowledgeRepository(h.db)

	// 验证所有权
	item, err := knowledgeRepo.GetKnowledgeItemByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "查询失败"})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Code: 40401, Message: "知识库条目不存在"})
		return
	}
	if item.TeacherID != teacherID {
		c.JSON(http.StatusForbidden, ErrorResponse{Code: 40301, Message: "无权操作此条目"})
		return
	}

	if err := knowledgeRepo.DeleteKnowledgeItem(id, teacherID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "删除失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// HandleGetKnowledgeDetail 获取知识库详情
func (h *Handler) HandleGetKnowledgeDetail(c *gin.Context) {
	userID, _ := c.Get("user_id")
	teacherID := userID.(int64)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "无效的ID"})
		return
	}

	knowledgeRepo := database.NewKnowledgeRepository(h.db)
	item, err := knowledgeRepo.GetKnowledgeItemByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "查询失败"})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Code: 40401, Message: "知识库条目不存在"})
		return
	}
	if item.TeacherID != teacherID {
		c.JSON(http.StatusForbidden, ErrorResponse{Code: 40301, Message: "无权查看此条目"})
		return
	}

	c.JSON(http.StatusOK, item)
}
