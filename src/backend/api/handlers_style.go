package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"digital-twin/src/backend/database"

	import_dialogue "digital-twin/src/plugins/dialogue"

	"github.com/gin-gonic/gin"
)

// ======================== 问答风格接口 ========================

// HandleSetDialogueStyle 设置学生问答风格
// PUT /api/students/:id/dialogue-style
func (h *Handler) HandleSetDialogueStyle(c *gin.Context) {
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

	// 从路径获取学生 id
	idStr := c.Param("id")
	studentID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || studentID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的学生 ID")
		return
	}

	// 解析请求体
	var req struct {
		StudentPersonaID *int64   `json:"student_persona_id"` // 可选，学生分身ID
		Temperature      *float64 `json:"temperature"`
		GuidanceLevel    *string  `json:"guidance_level"`
		TeachingStyle    *string  `json:"teaching_style"` // V2.0 迭代6: 教学风格
		StylePrompt      *string  `json:"style_prompt"`
		MaxTurnsPerTopic *int     `json:"max_turns_per_topic"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	// 校验 temperature 范围 0.1-1.0（如果传了的话）
	if req.Temperature != nil {
		if *req.Temperature < 0.1 || *req.Temperature > 1.0 {
			Error(c, http.StatusBadRequest, 40004, "temperature 必须在 0.1-1.0 之间")
			return
		}
	}

	// 校验 guidance_level 枚举 low/medium/high（如果传了的话）
	if req.GuidanceLevel != nil {
		gl := *req.GuidanceLevel
		if gl != "low" && gl != "medium" && gl != "high" {
			Error(c, http.StatusBadRequest, 40004, "guidance_level 必须为 low/medium/high")
			return
		}
	}

	// V2.0 迭代6: 校验 teaching_style
	if req.TeachingStyle != nil {
		ts := *req.TeachingStyle
		if ts != "" && !import_dialogue.ValidTeachingStyles[ts] {
			Error(c, http.StatusBadRequest, 40040, "无效的教学风格类型，可选值: socratic/explanatory/encouraging/strict/companion/custom")
			return
		}
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	relationRepo := database.NewRelationRepository(db)

	// 构建 StyleConfig 并序列化为 JSON 字符串
	styleConfig := database.StyleConfig{}
	if req.Temperature != nil {
		styleConfig.Temperature = *req.Temperature
	}
	if req.GuidanceLevel != nil {
		styleConfig.GuidanceLevel = *req.GuidanceLevel
	}
	if req.TeachingStyle != nil {
		styleConfig.TeachingStyle = *req.TeachingStyle
	}
	if req.StylePrompt != nil {
		styleConfig.StylePrompt = *req.StylePrompt
	}
	if req.MaxTurnsPerTopic != nil {
		styleConfig.MaxTurnsPerTopic = *req.MaxTurnsPerTopic
	}

	configJSON, err := json.Marshal(styleConfig)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "序列化风格配置失败: "+err.Error())
		return
	}

	styleRepo := database.NewStyleRepository(db)

	// 分身维度
	if teacherPersonaID > 0 {
		studentPersonaID := int64(0)
		if req.StudentPersonaID != nil {
			studentPersonaID = *req.StudentPersonaID
		} else {
			// 尝试获取学生默认分身
			userRepo := database.NewUserRepository(db)
			student, err := userRepo.GetByID(studentID)
			if err == nil && student != nil && student.DefaultPersonaID > 0 {
				studentPersonaID = student.DefaultPersonaID
			}
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
				approved, err = relationRepo.IsApproved(teacherID, studentID)
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

		style := &database.StudentDialogueStyle{
			TeacherID:        teacherID,
			StudentID:        studentID,
			TeacherPersonaID: teacherPersonaID,
			StudentPersonaID: studentPersonaID,
			StyleConfig:      string(configJSON),
		}
		id, err := styleRepo.UpsertWithPersonas(style)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "设置问答风格失败: "+err.Error())
			return
		}

		Success(c, gin.H{
			"id":                 id,
			"teacher_id":         teacherID,
			"student_id":         studentID,
			"teacher_persona_id": teacherPersonaID,
			"student_persona_id": studentPersonaID,
			"style_config":       styleConfig,
			"updated_at":         time.Now(),
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
		Error(c, http.StatusForbidden, 40007, "未获得该学生的授权关系")
		return
	}

	style := &database.StudentDialogueStyle{
		TeacherID:   teacherID,
		StudentID:   studentID,
		StyleConfig: string(configJSON),
	}
	id, err := styleRepo.Upsert(style)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "设置问答风格失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"id":           id,
		"teacher_id":   teacherID,
		"student_id":   studentID,
		"style_config": styleConfig,
		"updated_at":   time.Now(),
	})
}

// HandleGetDialogueStyle 获取学生问答风格
// GET /api/students/:id/dialogue-style
func (h *Handler) HandleGetDialogueStyle(c *gin.Context) {
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

	// 从路径获取学生 id
	idStr := c.Param("id")
	studentID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || studentID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的学生 ID")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	styleRepo := database.NewStyleRepository(db)

	// 分身维度查询
	if personaIDInt64 > 0 {
		var teacherPersonaID, studentPersonaID int64
		if roleStr == "teacher" {
			teacherPersonaID = personaIDInt64
			// 尝试从 query 参数获取 student_persona_id
			if spidStr := c.Query("student_persona_id"); spidStr != "" {
				studentPersonaID, _ = strconv.ParseInt(spidStr, 10, 64)
			}
		} else {
			studentPersonaID = personaIDInt64
			// 从 query 参数获取 teacher_persona_id
			if tpidStr := c.Query("teacher_persona_id"); tpidStr != "" {
				teacherPersonaID, _ = strconv.ParseInt(tpidStr, 10, 64)
			}
		}

		if teacherPersonaID > 0 && studentPersonaID > 0 {
			style, err := styleRepo.GetByPersonas(teacherPersonaID, studentPersonaID)
			if err != nil {
				Error(c, http.StatusInternalServerError, 50001, "查询问答风格失败: "+err.Error())
				return
			}
			if style == nil {
				Success(c, nil)
				return
			}

			var sc database.StyleConfig
			if err := json.Unmarshal([]byte(style.StyleConfig), &sc); err != nil {
				Error(c, http.StatusInternalServerError, 50001, "解析风格配置失败: "+err.Error())
				return
			}

			Success(c, gin.H{
				"id":                 style.ID,
				"teacher_id":         style.TeacherID,
				"student_id":         style.StudentID,
				"teacher_persona_id": style.TeacherPersonaID,
				"student_persona_id": style.StudentPersonaID,
				"style_config":       sc,
				"created_at":         style.CreatedAt,
				"updated_at":         style.UpdatedAt,
			})
			return
		}
	}

	// 向后兼容：user_id 维度
	var teacherID int64
	if roleStr == "student" {
		tidStr := c.Query("teacher_id")
		teacherID, err = strconv.ParseInt(tidStr, 10, 64)
		if err != nil || teacherID <= 0 {
			Error(c, http.StatusBadRequest, 40004, "缺少或无效的 teacher_id 参数")
			return
		}
	} else if roleStr == "teacher" {
		teacherID = userIDInt64
	} else {
		Error(c, http.StatusForbidden, 40003, "无权访问")
		return
	}

	style, err := styleRepo.GetByTeacherAndStudent(teacherID, studentID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询问答风格失败: "+err.Error())
		return
	}

	if style == nil {
		Success(c, nil)
		return
	}

	var styleConfig database.StyleConfig
	if err := json.Unmarshal([]byte(style.StyleConfig), &styleConfig); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "解析风格配置失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"id":           style.ID,
		"teacher_id":   style.TeacherID,
		"student_id":   style.StudentID,
		"style_config": styleConfig,
		"created_at":   style.CreatedAt,
		"updated_at":   style.UpdatedAt,
	})
}
