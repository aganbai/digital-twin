package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"digital-twin/src/backend/database"

	"github.com/gin-gonic/gin"
)

// HandleCreateFeedback 提交反馈
// POST /api/feedbacks
func (h *Handler) HandleCreateFeedback(c *gin.Context) {
	var req struct {
		FeedbackType string                 `json:"feedback_type" binding:"required"`
		Content      string                 `json:"content" binding:"required"`
		ContextInfo  map[string]interface{} `json:"context_info"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	// 验证反馈类型
	validTypes := map[string]bool{
		"suggestion":    true,
		"bug":           true,
		"content_issue": true,
		"other":         true,
	}
	if !validTypes[req.FeedbackType] {
		Error(c, http.StatusBadRequest, 40042, "无效的反馈类型，可选值: suggestion/bug/content_issue/other")
		return
	}

	// 内容长度限制
	if len(req.Content) > 2000 {
		Error(c, http.StatusBadRequest, 40004, "反馈内容不能超过2000字")
		return
	}

	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	contextInfoJSON := "{}"
	if req.ContextInfo != nil {
		if data, err := json.Marshal(req.ContextInfo); err == nil {
			contextInfoJSON = string(data)
		}
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	repo := database.NewFeedbackRepository(db)
	feedback := &database.Feedback{
		UserID:       userIDInt64,
		FeedbackType: req.FeedbackType,
		Content:      req.Content,
		ContextInfo:  contextInfoJSON,
	}

	id, err := repo.Create(feedback)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "提交反馈失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"id":      id,
		"message": "反馈提交成功，感谢您的反馈！",
	})
}

// HandleGetFeedbacks 获取反馈列表（管理员）
// GET /api/feedbacks
func (h *Handler) HandleGetFeedbacks(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	repo := database.NewFeedbackRepository(db)
	offset := (page - 1) * pageSize
	feedbacks, total, err := repo.ListAll(status, offset, pageSize)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询反馈列表失败: "+err.Error())
		return
	}

	SuccessPage(c, feedbacks, total, page, pageSize)
}

// HandleUpdateFeedbackStatus 更新反馈状态
// PUT /api/feedbacks/:id/status
func (h *Handler) HandleUpdateFeedbackStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的反馈ID")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	validStatuses := map[string]bool{
		"pending":  true,
		"reviewed": true,
		"resolved": true,
	}
	if !validStatuses[req.Status] {
		Error(c, http.StatusBadRequest, 40004, "无效的状态值，可选值: pending/reviewed/resolved")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	repo := database.NewFeedbackRepository(db)
	if err := repo.UpdateStatus(id, req.Status); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "更新反馈状态失败: "+err.Error())
		return
	}

	Success(c, gin.H{"message": "状态更新成功"})
}
