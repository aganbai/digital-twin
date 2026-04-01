package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"digital-twin/src/backend/database"

	"github.com/gin-gonic/gin"
)

// ======================== 评语接口 ========================

// HandleCreateComment 教师写评语
// POST /api/comments
func (h *Handler) HandleCreateComment(c *gin.Context) {
	var req struct {
		StudentID        int64  `json:"student_id" binding:"required"`
		StudentPersonaID int64  `json:"student_persona_id"` // 可选，学生分身ID
		Content          string `json:"content" binding:"required"`
		ProgressSummary  string `json:"progress_summary"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	// 从 JWT 获取教师 user_id
	userID, _ := c.Get("user_id")
	teacherID, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	// 从 JWT 获取 persona_id（教师分身ID）
	personaID, _ := c.Get("persona_id")
	teacherPersonaID, _ := personaID.(int64)

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	userRepo := database.NewUserRepository(db)
	relationRepo := database.NewRelationRepository(db)
	commentRepo := database.NewCommentRepository(db)

	// 校验 student_id 存在且角色为 student
	student, err := userRepo.GetByID(req.StudentID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询学生信息失败: "+err.Error())
		return
	}
	if student == nil || student.Role != "student" {
		Error(c, http.StatusNotFound, 40005, "学生不存在")
		return
	}

	// 分身维度
	if teacherPersonaID > 0 {
		studentPersonaID := req.StudentPersonaID
		if studentPersonaID == 0 && student.DefaultPersonaID > 0 {
			studentPersonaID = student.DefaultPersonaID
		}

		// 校验分身维度授权关系
		if studentPersonaID > 0 {
			approved, err := relationRepo.IsApprovedByPersonas(teacherPersonaID, studentPersonaID)
			if err != nil {
				Error(c, http.StatusInternalServerError, 50001, "查询授权关系失败: "+err.Error())
				return
			}
			if !approved {
				// 回退到 user_id 维度校验
				approved, err = relationRepo.IsApproved(teacherID, req.StudentID)
				if err != nil {
					Error(c, http.StatusInternalServerError, 50001, "查询授权关系失败: "+err.Error())
					return
				}
				if !approved {
					Error(c, http.StatusForbidden, 40007, "未获得该学生的授权关系")
					return
				}
			}
		}

		// 创建带分身维度的评语
		comment := &database.TeacherComment{
			TeacherID:        teacherID,
			StudentID:        req.StudentID,
			TeacherPersonaID: teacherPersonaID,
			StudentPersonaID: studentPersonaID,
			Content:          req.Content,
			ProgressSummary:  req.ProgressSummary,
		}
		id, err := commentRepo.CreateWithPersonas(comment)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "创建评语失败: "+err.Error())
			return
		}

		Success(c, gin.H{
			"id":                 id,
			"teacher_id":         teacherID,
			"student_id":         req.StudentID,
			"teacher_persona_id": teacherPersonaID,
			"student_persona_id": studentPersonaID,
			"content":            req.Content,
			"progress_summary":   req.ProgressSummary,
			"created_at":         time.Now(),
		})
		return
	}

	// 向后兼容：user_id 维度
	approved, err := relationRepo.IsApproved(teacherID, req.StudentID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询授权关系失败: "+err.Error())
		return
	}
	if !approved {
		Error(c, http.StatusForbidden, 40007, "未获得该学生的授权关系")
		return
	}

	comment := &database.TeacherComment{
		TeacherID:       teacherID,
		StudentID:       req.StudentID,
		Content:         req.Content,
		ProgressSummary: req.ProgressSummary,
	}
	id, err := commentRepo.Create(comment)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "创建评语失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"id":               id,
		"teacher_id":       teacherID,
		"student_id":       req.StudentID,
		"content":          req.Content,
		"progress_summary": req.ProgressSummary,
		"created_at":       time.Now(),
	})
}

// HandleGetComments 获取评语列表
// GET /api/comments
// V2.0 迭代5：学生角色调用时直接返回空列表（评语改为学生备注，学生不可见）
func (h *Handler) HandleGetComments(c *gin.Context) {
	// 从 JWT 获取 user_id 和 role
	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}
	role, _ := c.Get("role")
	roleStr := fmt.Sprintf("%v", role)

	// V2.0 迭代5：学生角色直接返回空列表
	if roleStr == "student" {
		Success(c, []interface{}{})
		return
	}

	// 从 JWT 获取 persona_id
	personaID, _ := c.Get("persona_id")
	personaIDInt64, _ := personaID.(int64)

	// 解析 query 参数
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

	commentRepo := database.NewCommentRepository(db)

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
			items, total, err := commentRepo.ListByTeacherPersona(personaIDInt64, studentPersonaIDPtr, offset, pageSize)
			if err != nil {
				Error(c, http.StatusInternalServerError, 50001, "查询评语列表失败: "+err.Error())
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
			items, total, err := commentRepo.ListByStudentPersona(personaIDInt64, teacherPersonaIDPtr, offset, pageSize)
			if err != nil {
				Error(c, http.StatusInternalServerError, 50001, "查询评语列表失败: "+err.Error())
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

		items, total, err := commentRepo.ListByTeacherWithNames(userIDInt64, studentIDPtr, offset, pageSize)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "查询评语列表失败: "+err.Error())
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

		items, total, err := commentRepo.ListByStudentWithNames(userIDInt64, teacherIDPtr, offset, pageSize)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "查询评语列表失败: "+err.Error())
			return
		}
		SuccessPage(c, items, total, page, pageSize)
	}
}
