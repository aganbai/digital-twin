package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"digital-twin/src/backend/database"

	"github.com/gin-gonic/gin"
)

// ======================== 课程信息发布接口 (V2.0 迭代9 M7) ========================

// CourseListItem 课程列表项（含班级名称和推送状态）
type CourseListItem struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	ClassID   int64     `json:"class_id"`
	ClassName string    `json:"class_name"`
	CreatedAt time.Time `json:"created_at"`
	Pushed    bool      `json:"pushed"`
}

// HandleCreateCourse 发布课程信息
// POST /api/courses
// API-109
func (h *Handler) HandleCreateCourse(c *gin.Context) {
	var req struct {
		Title          string `json:"title" binding:"required"`
		Content        string `json:"content" binding:"required"`
		ClassID        int64  `json:"class_id" binding:"required"`
		PushToStudents bool   `json:"push_to_students"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	// 获取当前用户信息
	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok || userIDInt64 <= 0 {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	personaID, _ := c.Get("persona_id")
	personaIDInt64, ok := personaID.(int64)
	if !ok || personaIDInt64 <= 0 {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	personaRepo := database.NewPersonaRepository(db)
	classRepo := database.NewClassRepository(db)
	knowledgeRepo := database.NewKnowledgeRepository(&database.Database{DB: db})

	// 校验当前分身是教师分身
	persona, err := personaRepo.GetByID(personaIDInt64)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询分身失败: "+err.Error())
		return
	}
	if persona == nil {
		Error(c, http.StatusNotFound, 40013, "分身不存在")
		return
	}
	if persona.Role != "teacher" {
		Error(c, http.StatusForbidden, 40018, "只有教师才能发布课程")
		return
	}

	// 校验班级存在且属于当前教师分身
	class, err := classRepo.GetByID(req.ClassID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询班级失败: "+err.Error())
		return
	}
	if class == nil {
		Error(c, http.StatusNotFound, 40017, "班级不存在")
		return
	}
	if class.PersonaID != personaIDInt64 {
		Error(c, http.StatusForbidden, 40018, "无权操作此班级")
		return
	}

	// 创建课程记录（存入 knowledge_items 表）
	title := strings.TrimSpace(req.Title)
	content := strings.TrimSpace(req.Content)

	courseItem := &database.KnowledgeItem{
		TeacherID: userIDInt64,
		PersonaID: personaIDInt64,
		Title:     title,
		Content:   content,
		ItemType:  "course",
		Status:    "active",
		Scope:     "class",
		ScopeID:   req.ClassID,
	}

	courseID, err := knowledgeRepo.CreateKnowledgeItem(courseItem)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "创建课程失败: "+err.Error())
		return
	}

	// 如果需要推送给学生
	if req.PushToStudents {
		notificationRepo := database.NewCourseNotificationRepository(db)
		notification := &database.CourseNotification{
			CourseItemID: courseID,
			ClassID:      req.ClassID,
			TeacherID:    userIDInt64,
			PersonaID:    personaIDInt64,
			PushType:     "in_app",
			Status:       "pending",
		}
		if _, err := notificationRepo.Create(notification); err != nil {
			// 推送失败不影响课程创建，记录日志即可
			// 实际生产环境应该记录日志
		}
	}

	Success(c, gin.H{
		"id":         courseID,
		"title":      title,
		"created_at": time.Now(),
	})
}

// HandleGetCourses 获取课程列表
// GET /api/courses?class_id=1&page=1&page_size=20
// API-110
func (h *Handler) HandleGetCourses(c *gin.Context) {
	classIDStr := c.Query("class_id")
	classID, err := strconv.ParseInt(classIDStr, 10, 64)
	if err != nil || classID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "缺少或无效的 class_id 参数")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	personaID, _ := c.Get("persona_id")
	personaIDInt64, ok := personaID.(int64)
	if !ok || personaIDInt64 <= 0 {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	personaRepo := database.NewPersonaRepository(db)
	classRepo := database.NewClassRepository(db)

	// 校验当前分身是教师分身
	persona, err := personaRepo.GetByID(personaIDInt64)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询分身失败: "+err.Error())
		return
	}
	if persona == nil {
		Error(c, http.StatusNotFound, 40013, "分身不存在")
		return
	}
	if persona.Role != "teacher" {
		Error(c, http.StatusForbidden, 40018, "只有教师才能查看课程列表")
		return
	}

	// 校验班级存在且属于当前教师分身
	class, err := classRepo.GetByID(classID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询班级失败: "+err.Error())
		return
	}
	if class == nil {
		Error(c, http.StatusNotFound, 40017, "班级不存在")
		return
	}
	if class.PersonaID != personaIDInt64 {
		Error(c, http.StatusForbidden, 40018, "无权查看此班级的课程")
		return
	}

	// 查询课程列表
	courses, total, err := h.getCoursesByClassID(classID, offset, pageSize)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询课程列表失败: "+err.Error())
		return
	}

	SuccessPage(c, courses, total, page, pageSize)
}

// getCoursesByClassID 根据班级ID查询课程列表
func (h *Handler) getCoursesByClassID(classID int64, offset, limit int) ([]CourseListItem, int, error) {
	db := h.manager.GetDB()

	// 查询总数
	var total int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM knowledge_items 
		WHERE item_type = 'course' AND scope = 'class' AND scope_id = ?`,
		classID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询列表（关联班级表获取班级名称，关联推送通知表判断是否已推送）
	query := `
		SELECT ki.id, ki.title, ki.content, ki.scope_id as class_id, 
		       COALESCE(c.name, '') as class_name, ki.created_at,
		       CASE WHEN cn.id IS NOT NULL THEN 1 ELSE 0 END as pushed
		FROM knowledge_items ki
		LEFT JOIN classes c ON ki.scope_id = c.id
		LEFT JOIN course_notifications cn ON ki.id = cn.course_item_id
		WHERE ki.item_type = 'course' AND ki.scope = 'class' AND ki.scope_id = ?
		ORDER BY ki.created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := db.Query(query, classID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var courses []CourseListItem
	for rows.Next() {
		var course CourseListItem
		var pushedInt int
		if err := rows.Scan(
			&course.ID, &course.Title, &course.Content, &course.ClassID,
			&course.ClassName, &course.CreatedAt, &pushedInt,
		); err != nil {
			return nil, 0, err
		}
		course.Pushed = pushedInt == 1
		courses = append(courses, course)
	}

	return courses, total, nil
}

// HandleUpdateCourse 更新课程信息
// PUT /api/courses/:id
// API-111
func (h *Handler) HandleUpdateCourse(c *gin.Context) {
	courseIDStr := c.Param("id")
	courseID, err := strconv.ParseInt(courseIDStr, 10, 64)
	if err != nil || courseID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的课程ID")
		return
	}

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok || userIDInt64 <= 0 {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	personaID, _ := c.Get("persona_id")
	personaIDInt64, ok := personaID.(int64)
	if !ok || personaIDInt64 <= 0 {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	knowledgeRepo := database.NewKnowledgeRepository(&database.Database{DB: db})

	// 查询课程是否存在
	course, err := knowledgeRepo.GetKnowledgeItemByID(courseID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询课程失败: "+err.Error())
		return
	}
	if course == nil {
		Error(c, http.StatusNotFound, 40017, "课程不存在")
		return
	}

	// 校验课程类型
	if course.ItemType != "course" {
		Error(c, http.StatusBadRequest, 40004, "该记录不是课程类型")
		return
	}

	// 校验权限：只有创建者才能更新
	if course.TeacherID != userIDInt64 {
		Error(c, http.StatusForbidden, 40018, "无权修改此课程")
		return
	}

	// 使用原值填充未提供的字段
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = course.Title
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		content = course.Content
	}

	// 更新课程
	course.Title = title
	course.Content = content
	if err := knowledgeRepo.UpdateKnowledgeItem(course); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "更新课程失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"id":      courseID,
		"title":   title,
		"content": content,
	})
}

// HandleDeleteCourse 删除课程信息
// DELETE /api/courses/:id
// API-112
func (h *Handler) HandleDeleteCourse(c *gin.Context) {
	courseIDStr := c.Param("id")
	courseID, err := strconv.ParseInt(courseIDStr, 10, 64)
	if err != nil || courseID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的课程ID")
		return
	}

	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok || userIDInt64 <= 0 {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	knowledgeRepo := database.NewKnowledgeRepository(&database.Database{DB: db})

	// 查询课程是否存在
	course, err := knowledgeRepo.GetKnowledgeItemByID(courseID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询课程失败: "+err.Error())
		return
	}
	if course == nil {
		Error(c, http.StatusNotFound, 40017, "课程不存在")
		return
	}

	// 校验课程类型
	if course.ItemType != "course" {
		Error(c, http.StatusBadRequest, 40004, "该记录不是课程类型")
		return
	}

	// 校验权限：只有创建者才能删除
	if course.TeacherID != userIDInt64 {
		Error(c, http.StatusForbidden, 40018, "无权删除此课程")
		return
	}

	// 删除课程
	if err := knowledgeRepo.DeleteKnowledgeItem(courseID, userIDInt64); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "删除课程失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"message": "课程已删除",
	})
}

// HandlePushCourseNotification 推送课程通知
// POST /api/courses/:id/push
// API-113
func (h *Handler) HandlePushCourseNotification(c *gin.Context) {
	courseIDStr := c.Param("id")
	courseID, err := strconv.ParseInt(courseIDStr, 10, 64)
	if err != nil || courseID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的课程ID")
		return
	}

	var req struct {
		PushType string `json:"push_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	// 校验推送类型
	if req.PushType != "in_app" && req.PushType != "wechat" {
		Error(c, http.StatusBadRequest, 40004, "无效的推送类型，只支持 in_app 或 wechat")
		return
	}

	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok || userIDInt64 <= 0 {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	personaID, _ := c.Get("persona_id")
	personaIDInt64, ok := personaID.(int64)
	if !ok || personaIDInt64 <= 0 {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	knowledgeRepo := database.NewKnowledgeRepository(&database.Database{DB: db})

	// 查询课程是否存在
	course, err := knowledgeRepo.GetKnowledgeItemByID(courseID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询课程失败: "+err.Error())
		return
	}
	if course == nil {
		Error(c, http.StatusNotFound, 40017, "课程不存在")
		return
	}

	// 校验课程类型
	if course.ItemType != "course" {
		Error(c, http.StatusBadRequest, 40004, "该记录不是课程类型")
		return
	}

	// 校验权限：只有创建者才能推送
	if course.TeacherID != userIDInt64 {
		Error(c, http.StatusForbidden, 40018, "无权推送此课程")
		return
	}

	// 微信订阅消息暂不实现
	if req.PushType == "wechat" {
		Success(c, gin.H{
			"message":      "微信订阅消息推送暂未开放，敬请期待",
			"pushed_count": 0,
			"failed_count": 0,
		})
		return
	}

	// 应用内推送：创建推送通知记录
	classID := course.ScopeID
	notificationRepo := database.NewCourseNotificationRepository(db)

	// 查询班级成员数
	classRepo := database.NewClassRepository(db)
	memberCount, err := classRepo.GetMemberCount(classID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询班级成员数失败: "+err.Error())
		return
	}

	// 创建推送记录
	notification := &database.CourseNotification{
		CourseItemID: courseID,
		ClassID:      classID,
		TeacherID:    userIDInt64,
		PersonaID:    personaIDInt64,
		PushType:     "in_app",
		Status:       "pending",
	}

	if _, err := notificationRepo.Create(notification); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "创建推送记录失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"pushed_count": memberCount,
		"failed_count": 0,
	})
}
