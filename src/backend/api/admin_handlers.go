package api

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"digital-twin/src/backend/database"

	"github.com/gin-gonic/gin"
)

// AdminHandler 管理员API处理器
type AdminHandler struct {
	db       *database.Database
	logDB    *database.LogDatabase
	userRepo *database.UserRepository
}

// NewAdminHandler 创建管理员处理器
func NewAdminHandler(db *database.Database, logDB *database.LogDatabase) *AdminHandler {
	return &AdminHandler{
		db:       db,
		logDB:    logDB,
		userRepo: database.NewUserRepository(db.DB),
	}
}

// ======================== M4: 管理员仪表盘API ========================

// HandleAdminDashboardOverview 系统总览
func (h *AdminHandler) HandleAdminDashboardOverview(c *gin.Context) {
	// 用户统计
	totalUsers, _ := h.userRepo.CountUsers("")
	activeUsers, _ := h.userRepo.CountUsersByStatus("active")
	disabledUsers, _ := h.userRepo.CountUsersByStatus("disabled")
	studentCount, _ := h.userRepo.CountUsers("student")
	teacherCount, _ := h.userRepo.CountUsers("teacher")

	// 班级统计
	var classCount int
	h.db.DB.QueryRow(`SELECT COUNT(*) FROM classes`).Scan(&classCount)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"users": gin.H{
				"total":    totalUsers,
				"active":   activeUsers,
				"disabled": disabledUsers,
				"students": studentCount,
				"teachers": teacherCount,
			},
			"classes": gin.H{
				"total": classCount,
			},
		},
	})
}

// HandleAdminUserStats 用户统计
func (h *AdminHandler) HandleAdminUserStats(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	totalUsers, _ := h.userRepo.CountUsers("")
	activeUsers, _ := h.userRepo.CountUsersByStatus("active")
	studentCount, _ := h.userRepo.CountUsers("student")
	teacherCount, _ := h.userRepo.CountUsers("teacher")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total":      totalUsers,
			"active":     activeUsers,
			"students":   studentCount,
			"teachers":   teacherCount,
			"start_date": startDate,
			"end_date":   endDate,
		},
	})
}

// HandleAdminChatStats 对话统计
func (h *AdminHandler) HandleAdminChatStats(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// 查询对话总数
	var totalConversations int
	h.db.DB.QueryRow(`SELECT COUNT(*) FROM conversations`).Scan(&totalConversations)

	// 查询消息总数
	var totalMessages int
	h.db.DB.QueryRow(`SELECT COUNT(*) FROM memories`).Scan(&totalMessages)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"conversations": totalConversations,
			"messages":      totalMessages,
			"start_date":    startDate,
			"end_date":      endDate,
		},
	})
}

// HandleAdminKnowledgeStats 知识库统计
func (h *AdminHandler) HandleAdminKnowledgeStats(c *gin.Context) {
	// 查询知识库文档总数
	var totalDocs int
	h.db.DB.QueryRow(`SELECT COUNT(*) FROM knowledge_items`).Scan(&totalDocs)

	// 查询活跃文档数
	var activeDocs int
	h.db.DB.QueryRow(`SELECT COUNT(*) FROM knowledge_items WHERE status = 'active'`).Scan(&activeDocs)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_documents":  totalDocs,
			"active_documents": activeDocs,
		},
	})
}

// HandleAdminActiveUsers 活跃用户排行
func (h *AdminHandler) HandleAdminActiveUsers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	users, err := h.userRepo.GetActiveUsers(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "查询活跃用户失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    users,
	})
}

// ======================== M5: 用户管理API ========================

// HandleAdminGetUsers 用户列表
func (h *AdminHandler) HandleAdminGetUsers(c *gin.Context) {
	role := c.Query("role")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	users, total, err := h.userRepo.ListUsers(role, status, offset, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("查询用户列表失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"list":      users,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// HandleAdminUpdateUserRole 修改用户角色
func (h *AdminHandler) HandleAdminUpdateUserRole(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的用户ID",
		})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "缺少角色参数",
		})
		return
	}

	// 验证角色值
	validRoles := map[string]bool{
		"student": true,
		"teacher": true,
		"admin":   true,
	}
	if !validRoles[req.Role] {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的角色值",
		})
		return
	}

	if err := h.userRepo.UpdateUserRole(userID, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "更新用户角色失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "角色更新成功",
	})
}

// HandleAdminUpdateUserStatus 启用/禁用用户
func (h *AdminHandler) HandleAdminUpdateUserStatus(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的用户ID",
		})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "缺少状态参数",
		})
		return
	}

	// 验证状态值
	if req.Status != "active" && req.Status != "disabled" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的状态值",
		})
		return
	}

	if err := h.userRepo.UpdateUserStatus(userID, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "更新用户状态失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "状态更新成功",
	})
}

// ======================== M6: 日志管理API ========================

// HandleAdminGetLogs 查询操作日志
func (h *AdminHandler) HandleAdminGetLogs(c *gin.Context) {
	userIDStr := c.Query("user_id")
	var userID int64
	if userIDStr != "" {
		userID, _ = strconv.ParseInt(userIDStr, 10, 64)
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	params := database.LogQueryParams{
		UserID:    userID,
		UserRole:  c.Query("user_role"),
		Action:    c.Query("action"),
		Resource:  c.Query("resource"),
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
		Page:      page,
		PageSize:  pageSize,
	}

	logs, total, err := h.logDB.QueryLogs(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "查询日志失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"list":      logs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// HandleAdminGetLogStats 日志统计
func (h *AdminHandler) HandleAdminGetLogStats(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	stats, err := h.logDB.GetLogStats(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "查询日志统计失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// HandleAdminExportLogs 导出日志CSV
func (h *AdminHandler) HandleAdminExportLogs(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	params := database.LogQueryParams{
		StartDate: startDate,
		EndDate:   endDate,
		Page:      1,
		PageSize:  10000, // 导出最多10000条
	}

	logs, _, err := h.logDB.QueryLogs(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "查询日志失败",
		})
		return
	}

	// 生成CSV
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=logs_%s.csv", time.Now().Format("20060102150405")))

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// 写入表头
	writer.Write([]string{"ID", "用户ID", "用户角色", "分身ID", "操作", "资源", "资源ID", "详情", "IP", "状态码", "耗时(ms)", "时间"})

	// 写入数据
	for _, log := range logs {
		writer.Write([]string{
			strconv.FormatInt(log.ID, 10),
			strconv.FormatInt(log.UserID, 10),
			log.UserRole,
			strconv.FormatInt(log.PersonaID, 10),
			log.Action,
			log.Resource,
			log.ResourceID,
			log.Detail,
			log.IP,
			strconv.Itoa(log.StatusCode),
			strconv.Itoa(log.DurationMs),
			log.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
}

// ======================== 反馈管理API ========================

// HandleAdminGetFeedbacks 反馈列表
func (h *AdminHandler) HandleAdminGetFeedbacks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// 构建查询条件
	whereClause := "1=1"
	args := []interface{}{}
	if status != "" {
		whereClause += " AND status = ?"
		args = append(args, status)
	}

	// 查询总数
	var total int
	h.db.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM feedbacks WHERE %s", whereClause), args...).Scan(&total)

	// 查询列表
	query := fmt.Sprintf(`
		SELECT id, user_id, persona_id, type, content, contact, status, created_at
		FROM feedbacks
		WHERE %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)
	args = append(args, pageSize, offset)

	rows, err := h.db.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "查询反馈列表失败",
		})
		return
	}
	defer rows.Close()

	var feedbacks []map[string]interface{}
	for rows.Next() {
		var id, userID, personaID int64
		var fbType, content, contact, status string
		var createdAt time.Time
		if err := rows.Scan(&id, &userID, &personaID, &fbType, &content, &contact, &status, &createdAt); err != nil {
			continue
		}
		feedbacks = append(feedbacks, map[string]interface{}{
			"id":         id,
			"user_id":    userID,
			"persona_id": personaID,
			"type":       fbType,
			"content":    content,
			"contact":    contact,
			"status":     status,
			"created_at": createdAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"list":      feedbacks,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}
