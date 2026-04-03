package api

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"digital-twin/src/backend/database"

	"github.com/gin-gonic/gin"
)

// 分享码字符集（排除容易混淆的 I/O/0/1）
const shareCodeCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
const shareCodeLength = 8

// generateShareCode 生成随机分享码
func generateShareCode() (string, error) {
	code := make([]byte, shareCodeLength)
	for i := 0; i < shareCodeLength; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(shareCodeCharset))))
		if err != nil {
			return "", err
		}
		code[i] = shareCodeCharset[idx.Int64()]
	}
	return string(code), nil
}

// HandleCreateShare 创建分享码
// POST /api/shares
func (h *Handler) HandleCreateShare(c *gin.Context) {
	// 从 JWT 获取 user_id 和 persona_id
	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	personaID, _ := c.Get("persona_id")
	personaIDInt64, ok := personaID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	// 解析请求体
	var req struct {
		ClassID                *int64 `json:"class_id"`
		ExpiresHours           *int   `json:"expires_hours"`
		MaxUses                *int   `json:"max_uses"`
		TargetStudentPersonaID int64  `json:"target_student_persona_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// 允许空 body，使用默认值
	}

	// 默认值
	expiresHours := 168 // 7天
	if req.ExpiresHours != nil {
		expiresHours = *req.ExpiresHours
	}
	maxUses := 0 // 不限
	if req.MaxUses != nil {
		maxUses = *req.MaxUses
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	personaRepo := database.NewPersonaRepository(db)
	classRepo := database.NewClassRepository(db)
	shareRepo := database.NewShareRepository(db)

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
		Error(c, http.StatusForbidden, 40018, "仅教师分身可创建分享码")
		return
	}
	if persona.UserID != userIDInt64 {
		Error(c, http.StatusForbidden, 40018, "无权操作该分身")
		return
	}

	// 如果指定了 class_id，校验班级存在且属于当前教师分身
	var className string
	if req.ClassID != nil {
		class, err := classRepo.GetByID(*req.ClassID)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "查询班级失败: "+err.Error())
			return
		}
		if class == nil {
			Error(c, http.StatusNotFound, 40017, "班级不存在")
			return
		}
		if class.PersonaID != personaIDInt64 {
			Error(c, http.StatusForbidden, 40018, "无权操作该班级")
			return
		}
		className = class.Name
	}

	// 生成 8 位随机分享码
	shareCode, err := generateShareCode()
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "生成分享码失败: "+err.Error())
		return
	}

	// 计算过期时间
	expiresAt := time.Now().Add(time.Duration(expiresHours) * time.Hour)

	// 创建分享码记录
	share := &database.PersonaShare{
		TeacherPersonaID:       personaIDInt64,
		ShareCode:              shareCode,
		ClassID:                req.ClassID,
		TargetStudentPersonaID: req.TargetStudentPersonaID,
		ExpiresAt:              &expiresAt,
		MaxUses:                maxUses,
	}

	shareID, err := shareRepo.Create(share)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "创建分享码失败: "+err.Error())
		return
	}

	// 查询目标学生昵称
	var targetStudentNickname string
	if req.TargetStudentPersonaID > 0 {
		sp, _ := personaRepo.GetByID(req.TargetStudentPersonaID)
		if sp != nil {
			targetStudentNickname = sp.Nickname
		}
	}

	Success(c, gin.H{
		"id":                        shareID,
		"share_code":                shareCode,
		"class_id":                  req.ClassID,
		"class_name":                className,
		"target_student_persona_id": req.TargetStudentPersonaID,
		"target_student_nickname":   targetStudentNickname,
		"expires_at":                expiresAt,
		"max_uses":                  maxUses,
		"used_count":                0,
		"is_active":                 true,
		"created_at":                time.Now(),
	})
}

// HandleGetShareInfo 获取分享码预览信息
// GET /api/shares/:code/info
func (h *Handler) HandleGetShareInfo(c *gin.Context) {
	shareCode := c.Param("code")
	if shareCode == "" {
		Error(c, http.StatusBadRequest, 40004, "缺少分享码参数")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	shareRepo := database.NewShareRepository(db)

	// 获取分享码信息（含教师分身信息）
	info, err := shareRepo.GetShareInfo(shareCode)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询分享码信息失败: "+err.Error())
		return
	}

	Success(c, info)
}

// HandleJoinByShare 通过分享码加入
// POST /api/shares/:code/join
func (h *Handler) HandleJoinByShare(c *gin.Context) {
	shareCode := c.Param("code")
	if shareCode == "" {
		Error(c, http.StatusBadRequest, 40004, "缺少分享码参数")
		return
	}

	// 从 JWT 获取 user_id
	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	// 解析请求体
	var req struct {
		StudentPersonaID *int64 `json:"student_persona_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// 允许空 body
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	shareRepo := database.NewShareRepository(db)
	personaRepo := database.NewPersonaRepository(db)
	relationRepo := database.NewRelationRepository(db)
	classRepo := database.NewClassRepository(db)

	// 确定 studentPersonaID
	var studentPersonaID int64
	if req.StudentPersonaID != nil && *req.StudentPersonaID > 0 {
		// 请求指定了 student_persona_id，校验属于当前用户且 role=student
		sp, err := personaRepo.GetByID(*req.StudentPersonaID)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "查询学生分身失败: "+err.Error())
			return
		}
		if sp == nil {
			Error(c, http.StatusNotFound, 40013, "学生分身不存在")
			return
		}
		if sp.UserID != userIDInt64 {
			Error(c, http.StatusForbidden, 40018, "无权操作该分身")
			return
		}
		if sp.Role != "student" {
			Error(c, http.StatusBadRequest, 40004, "指定的分身不是学生分身")
			return
		}
		studentPersonaID = *req.StudentPersonaID
	} else {
		// 未指定 student_persona_id，查询当前用户的所有学生分身
		studentPersonas, err := personaRepo.ListStudentPersonasByUserID(userIDInt64)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "查询学生分身列表失败: "+err.Error())
			return
		}
		if len(studentPersonas) == 0 {
			// 没有学生分身，返回 40022 引导用户创建
			Error(c, http.StatusBadRequest, 40022, "需要先创建学生分身才能加入")
			return
		}
		if len(studentPersonas) == 1 {
			// 只有一个学生分身，自动使用
			studentPersonaID = studentPersonas[0].ID
		} else {
			// 有多个学生分身，返回列表让用户选择
			personaList := make([]map[string]interface{}, 0, len(studentPersonas))
			for _, p := range studentPersonas {
				personaList = append(personaList, map[string]interface{}{
					"id":       p.ID,
					"nickname": p.Nickname,
					"school":   p.School,
				})
			}
			Success(c, gin.H{
				"need_select_persona": true,
				"student_personas":    personaList,
				"message":             "请选择一个学生分身加入",
			})
			return
		}
	}

	// 校验分享码有效
	share, err := shareRepo.GetByCode(shareCode)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询分享码失败: "+err.Error())
		return
	}
	if share == nil {
		Error(c, http.StatusNotFound, 40021, "分享码不存在")
		return
	}
	if share.IsActive != 1 {
		Error(c, http.StatusBadRequest, 40022, "分享码已失效")
		return
	}
	if share.ExpiresAt != nil && share.ExpiresAt.Before(time.Now()) {
		Error(c, http.StatusBadRequest, 40022, "分享码已过期")
		return
	}
	if share.MaxUses > 0 && share.UsedCount >= share.MaxUses {
		Error(c, http.StatusBadRequest, 40024, "分享码使用次数已达上限")
		return
	}

	// 🆕 V2.0 迭代6：定向邀请校验 - 非目标学生返回友好引导而非错误
	if share.TargetStudentPersonaID > 0 && share.TargetStudentPersonaID != studentPersonaID {
		// 查询教师分身信息
		tp, _ := personaRepo.GetByID(share.TeacherPersonaID)
		teacherNickname := ""
		if tp != nil {
			teacherNickname = tp.Nickname
		}
		Success(c, gin.H{
			"join_status":        "not_target",
			"teacher_persona_id": share.TeacherPersonaID,
			"teacher_nickname":   teacherNickname,
			"message":            "该邀请码是老师专门发给特定同学的，你可以向老师发起申请",
			"can_apply":          true,
		})
		return
	}

	// 校验 student_persona_id 是有效的学生分身且属于当前用户
	studentPersona, err := personaRepo.GetByID(studentPersonaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询学生分身失败: "+err.Error())
		return
	}
	if studentPersona == nil {
		Error(c, http.StatusNotFound, 40013, "学生分身不存在")
		return
	}
	if studentPersona.Role != "student" {
		Error(c, http.StatusBadRequest, 40004, "指定的分身不是学生分身")
		return
	}
	if studentPersona.UserID != userIDInt64 {
		Error(c, http.StatusForbidden, 40018, "无权操作该分身")
		return
	}

	// 查询教师分身获取 teacher_user_id
	teacherPersona, err := personaRepo.GetByID(share.TeacherPersonaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询教师分身失败: "+err.Error())
		return
	}
	if teacherPersona == nil {
		Error(c, http.StatusNotFound, 40013, "教师分身不存在")
		return
	}

	// 校验该学生分身与教师分身之间没有已存在的关系（approved 或 pending）
	existingRel, err := relationRepo.GetByPersonas(share.TeacherPersonaID, studentPersonaID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询师生关系失败: "+err.Error())
		return
	}
	if existingRel != nil {
		if existingRel.Status == "approved" {
			Error(c, http.StatusConflict, 40023, "已存在师生关系")
		} else if existingRel.Status == "pending" {
			Error(c, http.StatusConflict, 40023, "已提交申请，请等待教师审批")
		} else {
			Error(c, http.StatusConflict, 40023, "已存在关系记录")
		}
		return
	}

	// 创建师生关系（status=pending, initiated_by=share），需教师审批后才能对话
	relationID, err := relationRepo.CreateWithPersonas(
		teacherPersona.UserID, studentPersona.UserID,
		share.TeacherPersonaID, studentPersonaID,
		"pending", "share",
	)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "创建师生关系失败: "+err.Error())
		return
	}

	// 记录分享码关联的班级信息（不再自动加入，等教师审批时分配）
	var classID *int64
	var className string
	if share.ClassID != nil {
		class, err := classRepo.GetByID(*share.ClassID)
		if err == nil && class != nil {
			classID = &class.ID
			className = class.Name
		}
	}

	// 增加分享码使用次数
	_ = shareRepo.IncrementUsedCount(share.ID)

	Success(c, gin.H{
		"relation_id":        relationID,
		"teacher_persona_id": share.TeacherPersonaID,
		"teacher_nickname":   teacherPersona.Nickname,
		"class_id":           classID,
		"class_name":         className,
		"joined_class":       false,
		"join_status":        "pending_approval",
		"message":            "申请已提交，请等待教师审批",
	})
}

// HandleGetShares 获取分享码列表
// GET /api/shares
func (h *Handler) HandleGetShares(c *gin.Context) {
	// 从 JWT 获取 persona_id
	personaID, _ := c.Get("persona_id")
	personaIDInt64, ok := personaID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	personaRepo := database.NewPersonaRepository(db)
	shareRepo := database.NewShareRepository(db)

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
		Error(c, http.StatusForbidden, 40018, "仅教师分身可查看分享码列表")
		return
	}

	// 获取分享码列表
	items, err := shareRepo.ListByPersonaID(personaIDInt64)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询分享码列表失败: "+err.Error())
		return
	}

	Success(c, items)
}

// HandleDeactivateShare 停用分享码
// PUT /api/shares/:id/deactivate
func (h *Handler) HandleDeactivateShare(c *gin.Context) {
	// 从路径获取 share_id
	shareIDStr := c.Param("id")
	shareID, err := strconv.ParseInt(shareIDStr, 10, 64)
	if err != nil || shareID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的分享码ID")
		return
	}

	// 从 JWT 获取 persona_id
	personaID, _ := c.Get("persona_id")
	personaIDInt64, ok := personaID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	shareRepo := database.NewShareRepository(db)

	// 校验分享码存在且属于当前教师分身
	share, err := shareRepo.GetByID(shareID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询分享码失败: "+err.Error())
		return
	}
	if share == nil {
		Error(c, http.StatusNotFound, 40021, "分享码不存在")
		return
	}
	if share.TeacherPersonaID != personaIDInt64 {
		Error(c, http.StatusForbidden, 40018, "无权操作该分享码")
		return
	}

	// 停用分享码
	if err := shareRepo.Deactivate(shareID); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "停用分享码失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"message": "分享码已停用",
	})
}

// HandleGetShareInfoV2 获取分享码预览信息（V2.0 迭代6：支持可选鉴权 + join_status）
// GET /api/shares/:code/info
func (h *Handler) HandleGetShareInfoV2(c *gin.Context) {
	shareCode := c.Param("code")
	if shareCode == "" {
		Error(c, http.StatusBadRequest, 40004, "缺少分享码参数")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	shareRepo := database.NewShareRepository(db)
	personaRepo := database.NewPersonaRepository(db)

	// 获取分享码信息
	info, err := shareRepo.GetShareInfo(shareCode)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询分享码信息失败: "+err.Error())
		return
	}

	if info == nil {
		Error(c, http.StatusNotFound, 40021, "分享码不存在")
		return
	}

	// 判断 join_status
	joinStatus := "need_login" // 默认未登录

	// 检查是否已认证（通过 OptionalJWTAuthMiddleware 设置）
	_, authenticated := c.Get("authenticated")
	if authenticated {
		userID, _ := c.Get("user_id")
		userIDInt64, _ := userID.(int64)
		personaID, _ := c.Get("persona_id")
		personaIDInt64, _ := personaID.(int64)

		// 获取学生分身
		var studentPersonaID int64
		if personaIDInt64 > 0 {
			// 检查当前分身是否为学生分身
			currentPersona, _ := personaRepo.GetByID(personaIDInt64)
			if currentPersona != nil && currentPersona.Role == "student" {
				studentPersonaID = personaIDInt64
			}
		}

		if studentPersonaID == 0 && userIDInt64 > 0 {
			// 查询用户的学生分身
			studentPersonas, _ := personaRepo.ListStudentPersonasByUserID(userIDInt64)
			if len(studentPersonas) > 0 {
				studentPersonaID = studentPersonas[0].ID
			}
		}

		if studentPersonaID == 0 {
			joinStatus = "need_persona"
		} else {
			// 获取分享码对应的教师分身ID
			share, _ := shareRepo.GetByCode(shareCode)
			if share != nil {
				// 检查是否已是该教师的学生
				relationRepo := database.NewRelationRepository(db)
				existingRel, _ := relationRepo.GetByPersonas(share.TeacherPersonaID, studentPersonaID)
				if existingRel != nil {
					if existingRel.Status == "approved" {
						joinStatus = "already_joined"
					} else if existingRel.Status == "pending" {
						joinStatus = "pending_approval"
					} else {
						joinStatus = "can_join"
					}
				} else if share.TargetStudentPersonaID > 0 && share.TargetStudentPersonaID != studentPersonaID {
					joinStatus = "not_target"
				} else {
					joinStatus = "can_join"
				}
			}
		}
	}

	// 将 join_status 注入到响应中
	Success(c, gin.H{
		"teacher_persona_id":        info.TeacherPersonaID,
		"teacher_nickname":          info.TeacherNickname,
		"teacher_school":            info.TeacherSchool,
		"teacher_description":       info.TeacherDescription,
		"class_name":                info.ClassName,
		"target_student_persona_id": info.TargetStudentPersonaID,
		"target_student_nickname":   info.TargetStudentNickname,
		"is_valid":                  info.IsValid,
		"join_status":               joinStatus,
	})
}
