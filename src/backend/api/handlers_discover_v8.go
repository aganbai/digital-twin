package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetDiscoverRequest 发现页请求
type GetDiscoverRequest struct {
	Type     string `form:"type"` // class / teacher / all
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=10"`
}

// GetDiscoverResponse 发现页响应
type GetDiscoverResponse struct {
	Items      []DiscoverItem `json:"items"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// DiscoverItem 发现页推荐项
type DiscoverItem struct {
	ID          int64    `json:"id"`
	Type        string   `json:"type"` // class / teacher
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Avatar      string   `json:"avatar,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	MemberCount int      `json:"member_count,omitempty"`
	TeacherName string   `json:"teacher_name,omitempty"`
	School      string   `json:"school,omitempty"`
	Subject     string   `json:"subject,omitempty"`
	AgeGroup    []string `json:"age_group,omitempty"`
}

// HandleGetDiscover 获取发现页推荐
func (h *Handler) HandleGetDiscover(c *gin.Context) {
	// 获取热门班级（按成员数排序，最多10个）
	hotClasses, _ := h.getDiscoverClasses(10, 0)

	// 获取推荐教师（最多10个）
	recommendedTeachers, _ := h.getDiscoverTeachers(10, 0)

	// 可用学科列表
	subjects := []string{"语文", "数学", "英语", "物理", "化学", "生物", "其他"}

	c.JSON(http.StatusOK, gin.H{
		"hot_classes":          hotClasses,
		"recommended_teachers": recommendedTeachers,
		"subjects":             subjects,
	})
}

// getDiscoverClasses 获取推荐班级
func (h *Handler) getDiscoverClasses(limit, offset int) ([]DiscoverItem, int) {
	// 查询公开班级（有分享链接的）
	rows, err := h.db.DB.Query(`
		SELECT 
			c.id, c.name, c.description, c.subject, c.age_group,
			p.nickname as teacher_name, p.avatar,
			(SELECT COUNT(*) FROM class_members WHERE class_id = c.id) as member_count
		FROM classes c
		JOIN personas p ON c.persona_id = p.id
		WHERE c.is_active = 1 AND c.share_link != ''
		ORDER BY member_count DESC
		LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, 0
	}
	defer rows.Close()

	var items []DiscoverItem
	for rows.Next() {
		var item DiscoverItem
		var memberCount int
		var ageGroupStr string
		err := rows.Scan(
			&item.ID, &item.Title, &item.Description, &item.Subject, &ageGroupStr,
			&item.TeacherName, &item.Avatar, &memberCount,
		)
		if err != nil {
			continue
		}
		item.Type = "class"
		item.MemberCount = memberCount
		// 将 JSON 字符串反序列化为 []string
		if ageGroupStr != "" {
			json.Unmarshal([]byte(ageGroupStr), &item.AgeGroup)
		}
		if item.Subject != "" {
			item.Tags = append(item.Tags, item.Subject)
		}
		if len(item.AgeGroup) > 0 {
			item.Tags = append(item.Tags, item.AgeGroup...)
		}
		items = append(items, item)
	}

	// 查询总数
	var total int
	h.db.DB.QueryRow(`
		SELECT COUNT(*) FROM classes 
		WHERE is_active = 1 AND share_link != ''`).Scan(&total)

	return items, total
}

// getDiscoverTeachers 获取推荐教师
func (h *Handler) getDiscoverTeachers(limit, offset int) ([]DiscoverItem, int) {
	// 查询公开教师分身
	rows, err := h.db.DB.Query(`
		SELECT 
			p.id, p.nickname, p.school, p.description, p.avatar,
			(SELECT COUNT(*) FROM teacher_student_relations 
			 WHERE teacher_id = p.user_id AND status = 'approved') as student_count
		FROM personas p
		WHERE p.role = 'teacher' AND p.is_public = 1 AND p.is_active = 1
		ORDER BY student_count DESC
		LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, 0
	}
	defer rows.Close()

	var items []DiscoverItem
	for rows.Next() {
		var item DiscoverItem
		var studentCount int
		err := rows.Scan(
			&item.ID, &item.Title, &item.School, &item.Description, &item.Avatar, &studentCount,
		)
		if err != nil {
			continue
		}
		item.Type = "teacher"
		item.MemberCount = studentCount
		item.TeacherName = item.Title
		if item.School != "" {
			item.Tags = append(item.Tags, item.School)
		}
		items = append(items, item)
	}

	// 查询总数
	var total int
	h.db.DB.QueryRow(`
		SELECT COUNT(*) FROM personas 
		WHERE role = 'teacher' AND is_public = 1 AND is_active = 1`).Scan(&total)

	return items, total
}

// GetDiscoverDetailRequest 发现页详情请求
type GetDiscoverDetailRequest struct {
	Type string `form:"type" binding:"required"` // class / teacher
	ID   int64  `form:"id" binding:"required"`
}

// GetDiscoverDetailResponse 发现页详情响应
type GetDiscoverDetailResponse struct {
	ID                int64    `json:"id"`
	Type              string   `json:"type"`
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	Avatar            string   `json:"avatar,omitempty"`
	TeacherName       string   `json:"teacher_name,omitempty"`
	School            string   `json:"school,omitempty"`
	Subject           string   `json:"subject,omitempty"`
	AgeGroup          []string `json:"age_group,omitempty"`
	MemberCount       int      `json:"member_count"`
	Tags              []string `json:"tags,omitempty"`
	ShareLink         string   `json:"share_link,omitempty"`
	InviteCode        string   `json:"invite_code,omitempty"`
	IsJoined          bool     `json:"is_joined"`
	ApplicationStatus string   `json:"application_status,omitempty"` // pending / approved / rejected
}

// HandleGetDiscoverDetail 获取发现页详情
func (h *Handler) HandleGetDiscoverDetail(c *gin.Context) {
	itemType := c.Query("type")
	idStr := c.Query("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "无效的ID"})
		return
	}

	var result GetDiscoverDetailResponse

	switch itemType {
	case "class":
		result = h.getClassDetail(id, c)
	case "teacher":
		result = h.getTeacherDetail(id, c)
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "无效的类型"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// getClassDetail 获取班级详情
func (h *Handler) getClassDetail(classID int64, c *gin.Context) GetDiscoverDetailResponse {
	var result GetDiscoverDetailResponse
	result.Type = "class"
	result.ID = classID

	row := h.db.DB.QueryRow(`
		SELECT 
			c.name, c.description, c.subject, c.age_group,
			c.share_link, c.invite_code,
			p.nickname, p.avatar,
			(SELECT COUNT(*) FROM class_members WHERE class_id = c.id)
		FROM classes c
		JOIN personas p ON c.persona_id = p.id
		WHERE c.id = ?`, classID)

	var memberCount int
	var ageGroupStr string
	row.Scan(
		&result.Title, &result.Description, &result.Subject, &ageGroupStr,
		&result.ShareLink, &result.InviteCode,
		&result.TeacherName, &result.Avatar, &memberCount,
	)
	result.MemberCount = memberCount

	// 将 JSON 字符串反序列化为 []string
	if ageGroupStr != "" {
		json.Unmarshal([]byte(ageGroupStr), &result.AgeGroup)
	}

	if result.Subject != "" {
		result.Tags = append(result.Tags, result.Subject)
	}
	if len(result.AgeGroup) > 0 {
		result.Tags = append(result.Tags, result.AgeGroup...)
	}

	// 检查用户是否已加入
	userID, exists := c.Get("user_id")
	if exists {
		uid := userID.(int64)
		var studentPersonaID int64
		h.db.DB.QueryRow(`
			SELECT id FROM personas WHERE user_id = ? AND role = 'student' LIMIT 1`, uid).Scan(&studentPersonaID)

		var isJoined int
		h.db.DB.QueryRow(`
			SELECT COUNT(*) FROM class_members 
			WHERE class_id = ? AND student_persona_id = ?`, classID, studentPersonaID).Scan(&isJoined)
		result.IsJoined = isJoined > 0

		// 检查申请状态
		if !result.IsJoined {
			var status string
			h.db.DB.QueryRow(`
				SELECT status FROM class_join_requests 
				WHERE class_id = ? AND student_persona_id = ?`, classID, studentPersonaID).Scan(&status)
			result.ApplicationStatus = status
		}
	}

	return result
}

// getTeacherDetail 获取教师详情
func (h *Handler) getTeacherDetail(personaID int64, c *gin.Context) GetDiscoverDetailResponse {
	var result GetDiscoverDetailResponse
	result.Type = "teacher"
	result.ID = personaID

	row := h.db.DB.QueryRow(`
		SELECT 
			p.nickname, p.school, p.description, p.avatar,
			(SELECT COUNT(*) FROM teacher_student_relations 
			 WHERE teacher_id = p.user_id AND status = 'approved')
		FROM personas p
		WHERE p.id = ? AND p.role = 'teacher'`, personaID)

	var studentCount int
	row.Scan(&result.Title, &result.School, &result.Description, &result.Avatar, &studentCount)
	result.MemberCount = studentCount
	result.TeacherName = result.Title

	if result.School != "" {
		result.Tags = append(result.Tags, result.School)
	}

	// 检查用户是否已关联此教师
	userID, exists := c.Get("user_id")
	if exists {
		uid := userID.(int64)
		var teacherUserID int64
		h.db.DB.QueryRow(`SELECT user_id FROM personas WHERE id = ?`, personaID).Scan(&teacherUserID)

		var isJoined int
		h.db.DB.QueryRow(`
			SELECT COUNT(*) FROM teacher_student_relations 
			WHERE teacher_id = ? AND student_id = ? AND status = 'approved'`,
			teacherUserID, uid).Scan(&isJoined)
		result.IsJoined = isJoined > 0

		// 检查申请状态
		if !result.IsJoined {
			var status string
			h.db.DB.QueryRow(`
				SELECT status FROM teacher_student_relations 
				WHERE teacher_id = ? AND student_id = ?`, teacherUserID, uid).Scan(&status)
			result.ApplicationStatus = status
		}
	}

	return result
}

// HandleDiscoverSearch 搜索班级/老师
func (h *Handler) HandleDiscoverSearch(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "keyword 为必填参数"})
		return
	}

	searchType := c.DefaultQuery("type", "all")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	likeKeyword := "%" + keyword + "%"

	var items []DiscoverItem
	var total int

	switch searchType {
	case "class":
		items, total = h.searchDiscoverClasses(likeKeyword, pageSize, offset)
	case "teacher":
		items, total = h.searchDiscoverTeachers(likeKeyword, pageSize, offset)
	default:
		// 混合搜索
		classItems, classTotal := h.searchDiscoverClasses(likeKeyword, pageSize/2, offset)
		teacherItems, teacherTotal := h.searchDiscoverTeachers(likeKeyword, pageSize/2, offset)
		items = append(classItems, teacherItems...)
		total = classTotal + teacherTotal
	}

	totalPages := (total + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, GetDiscoverResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// searchDiscoverClasses 按关键词搜索推荐班级
func (h *Handler) searchDiscoverClasses(likeKeyword string, limit, offset int) ([]DiscoverItem, int) {
	rows, err := h.db.DB.Query(`
		SELECT 
			c.id, c.name, c.description, c.subject, c.age_group,
			p.nickname as teacher_name, p.avatar,
			(SELECT COUNT(*) FROM class_members WHERE class_id = c.id) as member_count
		FROM classes c
		JOIN personas p ON c.persona_id = p.id
		WHERE c.is_active = 1 AND c.share_link != '' AND (c.name LIKE ? OR p.nickname LIKE ?)
		ORDER BY member_count DESC
		LIMIT ? OFFSET ?`, likeKeyword, likeKeyword, limit, offset)
	if err != nil {
		return nil, 0
	}
	defer rows.Close()

	var items []DiscoverItem
	for rows.Next() {
		var item DiscoverItem
		var memberCount int
		var ageGroupStr string
		err := rows.Scan(
			&item.ID, &item.Title, &item.Description, &item.Subject, &ageGroupStr,
			&item.TeacherName, &item.Avatar, &memberCount,
		)
		if err != nil {
			continue
		}
		item.Type = "class"
		item.MemberCount = memberCount
		// 将 JSON 字符串反序列化为 []string
		if ageGroupStr != "" {
			json.Unmarshal([]byte(ageGroupStr), &item.AgeGroup)
		}
		if item.Subject != "" {
			item.Tags = append(item.Tags, item.Subject)
		}
		if len(item.AgeGroup) > 0 {
			item.Tags = append(item.Tags, item.AgeGroup...)
		}
		items = append(items, item)
	}

	// 查询总数
	var total int
	h.db.DB.QueryRow(`
		SELECT COUNT(*) FROM classes c
		JOIN personas p ON c.persona_id = p.id
		WHERE c.is_active = 1 AND c.share_link != '' AND (c.name LIKE ? OR p.nickname LIKE ?)`,
		likeKeyword, likeKeyword).Scan(&total)

	return items, total
}

// searchDiscoverTeachers 按关键词搜索推荐教师
func (h *Handler) searchDiscoverTeachers(likeKeyword string, limit, offset int) ([]DiscoverItem, int) {
	rows, err := h.db.DB.Query(`
		SELECT 
			p.id, p.nickname, p.school, p.description, p.avatar,
			(SELECT COUNT(*) FROM teacher_student_relations 
			 WHERE teacher_id = p.user_id AND status = 'approved') as student_count
		FROM personas p
		WHERE p.role = 'teacher' AND p.is_public = 1 AND p.is_active = 1 AND (p.nickname LIKE ? OR p.school LIKE ?)
		ORDER BY student_count DESC
		LIMIT ? OFFSET ?`, likeKeyword, likeKeyword, limit, offset)
	if err != nil {
		return nil, 0
	}
	defer rows.Close()

	var items []DiscoverItem
	for rows.Next() {
		var item DiscoverItem
		var studentCount int
		err := rows.Scan(
			&item.ID, &item.Title, &item.School, &item.Description, &item.Avatar, &studentCount,
		)
		if err != nil {
			continue
		}
		item.Type = "teacher"
		item.MemberCount = studentCount
		item.TeacherName = item.Title
		if item.School != "" {
			item.Tags = append(item.Tags, item.School)
		}
		items = append(items, item)
	}

	// 查询总数
	var total int
	h.db.DB.QueryRow(`
		SELECT COUNT(*) FROM personas p
		WHERE p.role = 'teacher' AND p.is_public = 1 AND p.is_active = 1 AND (p.nickname LIKE ? OR p.school LIKE ?)`,
		likeKeyword, likeKeyword).Scan(&total)

	return items, total
}
