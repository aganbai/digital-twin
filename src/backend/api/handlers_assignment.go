package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"digital-twin/src/backend/database"
	"digital-twin/src/harness/core"
	"digital-twin/src/plugins/dialogue"
	"digital-twin/src/plugins/knowledge"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ======================== 作业接口 ========================

// HandleSubmitAssignment 学生提交作业
// POST /api/assignments
func (h *Handler) HandleSubmitAssignment(c *gin.Context) {
	// 从 JWT 获取学生 user_id
	userID, _ := c.Get("user_id")
	studentID, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	// 从 JWT 获取 persona_id（学生分身ID）
	personaID, _ := c.Get("persona_id")
	studentPersonaID, _ := personaID.(int64)

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	var teacherID int64
	var teacherPersonaID int64
	var title, content, filePath, fileType string

	contentType := c.ContentType()
	if strings.HasPrefix(contentType, "multipart/form-data") {
		// multipart 模式
		tidStr := c.PostForm("teacher_id")
		var parseErr error
		teacherID, parseErr = strconv.ParseInt(tidStr, 10, 64)
		if parseErr != nil || teacherID <= 0 {
			Error(c, http.StatusBadRequest, 40004, "无效的 teacher_id 参数")
			return
		}
		// 解析 teacher_persona_id（可选）
		if tpidStr := c.PostForm("teacher_persona_id"); tpidStr != "" {
			teacherPersonaID, _ = strconv.ParseInt(tpidStr, 10, 64)
		}
		title = c.PostForm("title")
		content = c.PostForm("content")

		// 处理文件上传
		file, header, err := c.Request.FormFile("file")
		if err == nil {
			defer file.Close()
			// 保存文件到 {UPLOAD_DIR}/assignments/{student_id}/{uuid}_{filename}
			baseUploadDir := os.Getenv("UPLOAD_DIR")
			if baseUploadDir == "" {
				baseUploadDir = "./uploads"
			}
			uploadDir := filepath.Join(baseUploadDir, "assignments", fmt.Sprintf("%d", studentID))
			if err := os.MkdirAll(uploadDir, 0755); err != nil {
				Error(c, http.StatusInternalServerError, 50001, "创建上传目录失败: "+err.Error())
				return
			}

			fileUUID := uuid.New().String()
			savedFilename := fileUUID + "_" + header.Filename
			filePath = filepath.Join(uploadDir, savedFilename)

			// 保存文件
			out, err := os.Create(filePath)
			if err != nil {
				Error(c, http.StatusInternalServerError, 50001, "保存文件失败: "+err.Error())
				return
			}
			defer out.Close()
			if _, err := io.Copy(out, file); err != nil {
				Error(c, http.StatusInternalServerError, 50001, "保存文件失败: "+err.Error())
				return
			}

			ext := strings.ToLower(filepath.Ext(header.Filename))
			fileType = strings.TrimPrefix(ext, ".")
		}
	} else {
		// JSON 模式
		var req struct {
			TeacherID        int64  `json:"teacher_id"`
			TeacherPersonaID int64  `json:"teacher_persona_id"` // 可选，教师分身ID
			Title            string `json:"title"`
			Content          string `json:"content"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
			return
		}
		teacherID = req.TeacherID
		teacherPersonaID = req.TeacherPersonaID
		title = req.Title
		content = req.Content
	}

	// 校验 title 非空
	if title == "" {
		Error(c, http.StatusBadRequest, 40004, "作业标题不能为空")
		return
	}

	// 校验 content 和 file 至少一个非空
	if content == "" && filePath == "" {
		Error(c, http.StatusBadRequest, 40004, "作业内容和文件至少需要提供一个")
		return
	}

	// 校验 teacher_id
	if teacherID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "缺少 teacher_id 参数")
		return
	}

	// 校验师生关系已授权
	relationRepo := database.NewRelationRepository(db)
	asgRepo := database.NewAssignmentRepository(db)

	// 分身维度
	if studentPersonaID > 0 && teacherPersonaID > 0 {
		approved, err := relationRepo.IsApprovedByPersonas(teacherPersonaID, studentPersonaID)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "查询授权关系失败: "+err.Error())
			return
		}
		if !approved {
			// 回退到 user_id 维度校验
			approved, err = relationRepo.IsApproved(teacherID, studentID)
			if err != nil {
				Error(c, http.StatusInternalServerError, 50001, "查询授权关系失败: "+err.Error())
				return
			}
			if !approved {
				Error(c, http.StatusForbidden, 40007, "未获得该教师授权，请先申请")
				return
			}
		}

		asg := &database.Assignment{
			StudentID:        studentID,
			TeacherID:        teacherID,
			TeacherPersonaID: teacherPersonaID,
			StudentPersonaID: studentPersonaID,
			Title:            title,
			Content:          content,
			FilePath:         filePath,
			FileType:         fileType,
			Status:           "submitted",
		}
		id, err := asgRepo.CreateWithPersonas(asg)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "创建作业失败: "+err.Error())
			return
		}

		Success(c, gin.H{
			"id":                 id,
			"student_id":         studentID,
			"teacher_id":         teacherID,
			"teacher_persona_id": teacherPersonaID,
			"student_persona_id": studentPersonaID,
			"title":              title,
			"status":             "submitted",
			"has_file":           filePath != "",
			"created_at":         time.Now(),
		})
		return
	}

	// 向后兼容：user_id 维度
	approved, err := relationRepo.IsApproved(teacherID, studentID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询授权关系失败: "+err.Error())
		return
	}
	if !approved {
		Error(c, http.StatusForbidden, 40007, "未获得该教师授权，请先申请")
		return
	}

	asg := &database.Assignment{
		StudentID: studentID,
		TeacherID: teacherID,
		Title:     title,
		Content:   content,
		FilePath:  filePath,
		FileType:  fileType,
		Status:    "submitted",
	}
	id, err := asgRepo.Create(asg)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "创建作业失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"id":         id,
		"student_id": studentID,
		"teacher_id": teacherID,
		"title":      title,
		"status":     "submitted",
		"has_file":   filePath != "",
		"created_at": time.Now(),
	})
}

// HandleGetAssignments 获取作业列表
// GET /api/assignments
func (h *Handler) HandleGetAssignments(c *gin.Context) {
	// 从 JWT 获取 user_id 和 role
	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}
	role, _ := c.Get("role")
	roleStr := fmt.Sprintf("%v", role)

	// 从 JWT 获取 persona_id
	personaID, _ := c.Get("persona_id")
	personaIDInt64, _ := personaID.(int64)

	// 解析 query 参数
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

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	asgRepo := database.NewAssignmentRepository(db)

	// 分身维度查询
	if personaIDInt64 > 0 {
		if roleStr == "teacher" {
			var studentPersonaIDPtr *int64
			if spidStr := c.Query("student_persona_id"); spidStr != "" {
				spid, err := strconv.ParseInt(spidStr, 10, 64)
				if err != nil || spid <= 0 {
					Error(c, http.StatusBadRequest, 40004, "无效的 student_persona_id 参数")
					return
				}
				studentPersonaIDPtr = &spid
			}
			items, total, err := asgRepo.ListByTeacherPersona(personaIDInt64, studentPersonaIDPtr, status, offset, pageSize)
			if err != nil {
				Error(c, http.StatusInternalServerError, 50001, "查询作业列表失败: "+err.Error())
				return
			}
			SuccessPage(c, items, total, page, pageSize)
		} else {
			var teacherPersonaIDPtr *int64
			if tpidStr := c.Query("teacher_persona_id"); tpidStr != "" {
				tpid, err := strconv.ParseInt(tpidStr, 10, 64)
				if err != nil || tpid <= 0 {
					Error(c, http.StatusBadRequest, 40004, "无效的 teacher_persona_id 参数")
					return
				}
				teacherPersonaIDPtr = &tpid
			}
			items, total, err := asgRepo.ListByStudentPersona(personaIDInt64, teacherPersonaIDPtr, status, offset, pageSize)
			if err != nil {
				Error(c, http.StatusInternalServerError, 50001, "查询作业列表失败: "+err.Error())
				return
			}
			SuccessPage(c, items, total, page, pageSize)
		}
		return
	}

	// 向后兼容：user_id 维度
	if roleStr == "teacher" {
		var studentIDPtr *int64
		if sidStr := c.Query("student_id"); sidStr != "" {
			sid, err := strconv.ParseInt(sidStr, 10, 64)
			if err != nil || sid <= 0 {
				Error(c, http.StatusBadRequest, 40004, "无效的 student_id 参数")
				return
			}
			studentIDPtr = &sid
		}

		items, total, err := asgRepo.ListByTeacherWithDetails(userIDInt64, studentIDPtr, status, offset, pageSize)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "查询作业列表失败: "+err.Error())
			return
		}
		SuccessPage(c, items, total, page, pageSize)
	} else {
		var teacherIDPtr *int64
		if tidStr := c.Query("teacher_id"); tidStr != "" {
			tid, err := strconv.ParseInt(tidStr, 10, 64)
			if err != nil || tid <= 0 {
				Error(c, http.StatusBadRequest, 40004, "无效的 teacher_id 参数")
				return
			}
			teacherIDPtr = &tid
		}

		items, total, err := asgRepo.ListByStudentWithDetails(userIDInt64, teacherIDPtr, status, offset, pageSize)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "查询作业列表失败: "+err.Error())
			return
		}
		SuccessPage(c, items, total, page, pageSize)
	}
}

// HandleGetAssignmentDetail 获取作业详情
// GET /api/assignments/:id
func (h *Handler) HandleGetAssignmentDetail(c *gin.Context) {
	// 从路径获取作业 id
	idStr := c.Param("id")
	asgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || asgID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的作业 ID")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	asgRepo := database.NewAssignmentRepository(db)
	reviewRepo := database.NewReviewRepository(db)
	userRepo := database.NewUserRepository(db)

	// 查询作业详情
	asg, err := asgRepo.GetByID(asgID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询作业失败: "+err.Error())
		return
	}
	if asg == nil {
		Error(c, http.StatusNotFound, 40005, "作业不存在")
		return
	}

	// 查询所有点评
	reviews, err := reviewRepo.ListByAssignment(asgID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询点评列表失败: "+err.Error())
		return
	}

	// JOIN 获取 student_nickname 和 teacher_nickname
	var studentNickname, teacherNickname string
	student, err := userRepo.GetByID(asg.StudentID)
	if err == nil && student != nil {
		studentNickname = student.Nickname
	}
	teacher, err := userRepo.GetByID(asg.TeacherID)
	if err == nil && teacher != nil {
		teacherNickname = teacher.Nickname
	}

	Success(c, gin.H{
		"id":               asg.ID,
		"student_id":       asg.StudentID,
		"student_nickname": studentNickname,
		"teacher_id":       asg.TeacherID,
		"teacher_nickname": teacherNickname,
		"title":            asg.Title,
		"content":          asg.Content,
		"file_path":        asg.FilePath,
		"file_type":        asg.FileType,
		"status":           asg.Status,
		"has_file":         asg.FilePath != "",
		"reviews":          reviews,
		"created_at":       asg.CreatedAt,
		"updated_at":       asg.UpdatedAt,
	})
}

// HandleReviewAssignment 教师点评作业
// POST /api/assignments/:id/review
func (h *Handler) HandleReviewAssignment(c *gin.Context) {
	// 从 JWT 获取教师 user_id
	userID, _ := c.Get("user_id")
	teacherID, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	// 从路径获取作业 id
	idStr := c.Param("id")
	asgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || asgID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的作业 ID")
		return
	}

	// 解析请求体
	var req struct {
		Content string   `json:"content" binding:"required"`
		Score   *float64 `json:"score"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	asgRepo := database.NewAssignmentRepository(db)
	reviewRepo := database.NewReviewRepository(db)

	// 查询作业
	asg, err := asgRepo.GetByID(asgID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询作业失败: "+err.Error())
		return
	}
	if asg == nil {
		Error(c, http.StatusNotFound, 40005, "作业不存在")
		return
	}

	// 校验作业的 teacher_id == 当前教师
	if asg.TeacherID != teacherID {
		Error(c, http.StatusForbidden, 40003, "无权点评此作业")
		return
	}

	// 创建点评记录
	review := &database.AssignmentReview{
		AssignmentID: asgID,
		ReviewerType: "teacher",
		ReviewerID:   &teacherID,
		Content:      req.Content,
		Score:        req.Score,
	}
	reviewID, err := reviewRepo.Create(review)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "创建点评失败: "+err.Error())
		return
	}

	// 更新作业状态为 reviewed
	if err := asgRepo.UpdateStatus(asgID, "reviewed"); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "更新作业状态失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"id":            reviewID,
		"assignment_id": asgID,
		"reviewer_type": "teacher",
		"reviewer_id":   teacherID,
		"content":       req.Content,
		"score":         req.Score,
		"created_at":    time.Now(),
	})
}

// HandleAIReviewAssignment AI 自动点评
// POST /api/assignments/:id/ai-review
func (h *Handler) HandleAIReviewAssignment(c *gin.Context) {
	// 从路径获取作业 id
	idStr := c.Param("id")
	asgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || asgID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的作业 ID")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	asgRepo := database.NewAssignmentRepository(db)
	reviewRepo := database.NewReviewRepository(db)
	userRepo := database.NewUserRepository(db)

	// 查询作业详情
	asg, err := asgRepo.GetByID(asgID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询作业失败: "+err.Error())
		return
	}
	if asg == nil {
		Error(c, http.StatusNotFound, 40005, "作业不存在")
		return
	}

	// 获取作业内容（如果有文件，解析文件内容）
	assignmentContent := asg.Content
	if asg.FilePath != "" {
		parser := knowledge.NewFileParser()
		fileContent, err := parser.Parse(asg.FilePath)
		if err != nil {
			// 文件解析失败不阻塞，使用已有内容
			fmt.Printf("[AI点评] 解析文件失败: %v\n", err)
		} else {
			if assignmentContent != "" {
				assignmentContent = assignmentContent + "\n\n【附件内容】\n" + fileContent
			} else {
				assignmentContent = fileContent
			}
		}
	}

	// 查询教师信息获取 nickname
	teacher, err := userRepo.GetByID(asg.TeacherID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询教师信息失败: "+err.Error())
		return
	}
	teacherNickname := "教师"
	if teacher != nil && teacher.Nickname != "" {
		teacherNickname = teacher.Nickname
	}

	// 获取 dialogue 插件的 LLMClient 和 PromptBuilder
	plugin, err := h.manager.GetPlugin("dialogue")
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "对话服务不可用")
		return
	}
	dialoguePlugin, ok := plugin.(*dialogue.DialoguePlugin)
	if !ok {
		Error(c, http.StatusInternalServerError, 50001, "对话插件类型错误")
		return
	}

	llmClient := dialoguePlugin.GetLLMClient()
	promptBuilder := dialoguePlugin.GetPromptBuilder()

	// 检索知识库获取相关知识片段
	var knowledgeChunks string
	knowledgePlugin, err := h.manager.GetPlugin("knowledge-retrieval")
	if err == nil {
		searchInput := &core.PluginInput{
			RequestID: uuid.New().String(),
			Data: map[string]interface{}{
				"action":     "search",
				"query":      asg.Title + " " + assignmentContent,
				"teacher_id": asg.TeacherID,
				"limit":      5,
			},
			Context: c.Request.Context(),
		}
		searchOutput, err := knowledgePlugin.Execute(c.Request.Context(), searchInput)
		if err == nil && searchOutput.Success {
			if chunks, ok := searchOutput.Data["chunks"].([]map[string]interface{}); ok {
				var parts []string
				for _, chunk := range chunks {
					if content, ok := chunk["content"].(string); ok && content != "" {
						parts = append(parts, content)
					}
				}
				knowledgeChunks = strings.Join(parts, "\n")
			}
		}
	}

	// 构建 AI 点评 prompt
	promptContent := promptBuilder.BuildAssignmentReviewPrompt(teacherNickname, asg.Title, assignmentContent, knowledgeChunks)

	// 调用 LLM 生成点评
	messages := []dialogue.ChatMessage{
		{Role: "system", Content: "你是一位专业的教师助手，负责对学生作业进行点评。"},
		{Role: "user", Content: promptContent},
	}

	chatResp, err := llmClient.Chat(messages)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50002, "AI 点评生成失败: "+err.Error())
		return
	}

	// 从回复中提取评分（正则匹配 数字/100 格式）
	var score *float64
	scoreRegex := regexp.MustCompile(`(\d+)\s*/\s*100`)
	matches := scoreRegex.FindStringSubmatch(chatResp.Content)
	if len(matches) >= 2 {
		if s, err := strconv.ParseFloat(matches[1], 64); err == nil {
			score = &s
		}
	}

	// 创建点评记录（reviewer_type=ai, reviewer_id=nil）
	review := &database.AssignmentReview{
		AssignmentID: asgID,
		ReviewerType: "ai",
		ReviewerID:   nil,
		Content:      chatResp.Content,
		Score:        score,
	}
	reviewID, err := reviewRepo.Create(review)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "创建点评记录失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"id":            reviewID,
		"assignment_id": asgID,
		"reviewer_type": "ai",
		"content":       chatResp.Content,
		"score":         score,
		"token_usage": gin.H{
			"prompt_tokens":     chatResp.PromptTokens,
			"completion_tokens": chatResp.CompletionTokens,
			"total_tokens":      chatResp.TotalTokens,
		},
		"created_at": time.Now(),
	})
}
