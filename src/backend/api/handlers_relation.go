package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"digital-twin/src/backend/database"

	"github.com/gin-gonic/gin"
)

// ======================== 师生关系接口 ========================

// HandleInviteStudent 教师邀请学生
// POST /api/relations/invite
func (h *Handler) HandleInviteStudent(c *gin.Context) {
	var req struct {
		StudentID        int64 `json:"student_id" binding:"required"`
		StudentPersonaID int64 `json:"student_persona_id"` // 可选，学生分身ID
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	// 从 JWT 获取当前教师 user_id
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
	personaRepo := database.NewPersonaRepository(db)

	// 校验学生存在且角色为 student
	student, err := userRepo.GetByID(req.StudentID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询学生信息失败: "+err.Error())
		return
	}
	if student == nil || student.Role != "student" {
		Error(c, http.StatusNotFound, 40005, "学生不存在")
		return
	}

	// 如果使用分身维度
	if teacherPersonaID > 0 {
		// 如果未传 student_persona_id，使用学生的默认分身
		studentPersonaID := req.StudentPersonaID
		if studentPersonaID == 0 {
			if student.DefaultPersonaID > 0 {
				studentPersonaID = student.DefaultPersonaID
			}
		}

		// 校验 student_persona_id 有效性
		if studentPersonaID > 0 {
			studentPersona, err := personaRepo.GetByID(studentPersonaID)
			if err != nil {
				Error(c, http.StatusInternalServerError, 50001, "查询学生分身失败: "+err.Error())
				return
			}
			if studentPersona == nil || studentPersona.UserID != req.StudentID {
				Error(c, http.StatusBadRequest, 40004, "无效的学生分身ID")
				return
			}
		}

		// 校验分身维度关系不存在
		if studentPersonaID > 0 {
			existing, err := relationRepo.GetByPersonas(teacherPersonaID, studentPersonaID)
			if err != nil {
				Error(c, http.StatusInternalServerError, 50001, "查询关系失败: "+err.Error())
				return
			}
			if existing != nil {
				Error(c, http.StatusConflict, 40009, "师生关系已存在")
				return
			}
		}

		// 创建带分身维度的关系
		id, err := relationRepo.CreateWithPersonas(teacherID, req.StudentID, teacherPersonaID, studentPersonaID, "approved", "teacher")
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "创建关系失败: "+err.Error())
			return
		}

		created, err := relationRepo.GetByID(id)
		if err != nil || created == nil {
			Error(c, http.StatusInternalServerError, 50001, "查询关系详情失败")
			return
		}

		Success(c, created)
		return
	}

	// 向后兼容：user_id 维度
	existing, err := relationRepo.GetByTeacherAndStudent(teacherID, req.StudentID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询关系失败: "+err.Error())
		return
	}
	if existing != nil {
		Error(c, http.StatusConflict, 40009, "师生关系已存在")
		return
	}

	rel := &database.TeacherStudentRelation{
		TeacherID:   teacherID,
		StudentID:   req.StudentID,
		Status:      "approved",
		InitiatedBy: "teacher",
	}
	id, err := relationRepo.Create(rel)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "创建关系失败: "+err.Error())
		return
	}

	created, err := relationRepo.GetByID(id)
	if err != nil || created == nil {
		Error(c, http.StatusInternalServerError, 50001, "查询关系详情失败")
		return
	}

	Success(c, created)
}

// HandleApplyTeacher 学生申请使用分身
// POST /api/relations/apply
func (h *Handler) HandleApplyTeacher(c *gin.Context) {
	var req struct {
		TeacherID        int64 `json:"teacher_id" binding:"required"`
		TeacherPersonaID int64 `json:"teacher_persona_id"` // 可选，教师分身ID
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	// 从 JWT 获取当前学生 user_id
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

	userRepo := database.NewUserRepository(db)
	relationRepo := database.NewRelationRepository(db)
	personaRepo := database.NewPersonaRepository(db)

	// 校验教师存在且角色为 teacher
	teacher, err := userRepo.GetByID(req.TeacherID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询教师信息失败: "+err.Error())
		return
	}
	if teacher == nil || teacher.Role != "teacher" {
		Error(c, http.StatusNotFound, 40005, "教师不存在")
		return
	}

	// 如果使用分身维度
	if studentPersonaID > 0 {
		// 如果未传 teacher_persona_id，使用教师的默认分身
		teacherPersonaID := req.TeacherPersonaID
		if teacherPersonaID == 0 {
			if teacher.DefaultPersonaID > 0 {
				teacherPersonaID = teacher.DefaultPersonaID
			}
		}

		// 校验 teacher_persona_id 有效性
		if teacherPersonaID > 0 {
			teacherPersona, err := personaRepo.GetByID(teacherPersonaID)
			if err != nil {
				Error(c, http.StatusInternalServerError, 50001, "查询教师分身失败: "+err.Error())
				return
			}
			if teacherPersona == nil || teacherPersona.UserID != req.TeacherID {
				Error(c, http.StatusBadRequest, 40004, "无效的教师分身ID")
				return
			}
		}

		// 校验分身维度关系不存在
		if teacherPersonaID > 0 {
			existing, err := relationRepo.GetByPersonas(teacherPersonaID, studentPersonaID)
			if err != nil {
				Error(c, http.StatusInternalServerError, 50001, "查询关系失败: "+err.Error())
				return
			}
			if existing != nil {
				Error(c, http.StatusConflict, 40009, "师生关系已存在")
				return
			}
		}

		// 创建带分身维度的关系
		id, err := relationRepo.CreateWithPersonas(req.TeacherID, studentID, teacherPersonaID, studentPersonaID, "pending", "student")
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "创建关系失败: "+err.Error())
			return
		}

		created, err := relationRepo.GetByID(id)
		if err != nil || created == nil {
			Error(c, http.StatusInternalServerError, 50001, "查询关系详情失败")
			return
		}

		Success(c, created)
		return
	}

	// 向后兼容：user_id 维度
	existing, err := relationRepo.GetByTeacherAndStudent(req.TeacherID, studentID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询关系失败: "+err.Error())
		return
	}
	if existing != nil {
		Error(c, http.StatusConflict, 40009, "师生关系已存在")
		return
	}

	rel := &database.TeacherStudentRelation{
		TeacherID:   req.TeacherID,
		StudentID:   studentID,
		Status:      "pending",
		InitiatedBy: "student",
	}
	id, err := relationRepo.Create(rel)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "创建关系失败: "+err.Error())
		return
	}

	created, err := relationRepo.GetByID(id)
	if err != nil || created == nil {
		Error(c, http.StatusInternalServerError, 50001, "查询关系详情失败")
		return
	}

	Success(c, created)
}

// HandleApproveRelation 教师审批同意
// PUT /api/relations/:id/approve
func (h *Handler) HandleApproveRelation(c *gin.Context) {
	// 从 JWT 获取当前教师 user_id
	userID, _ := c.Get("user_id")
	teacherID, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	// 从路径获取关系 id
	idStr := c.Param("id")
	relID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || relID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的关系 ID")
		return
	}

	// 解析请求体（可选参数：评语和班级）
	var req struct {
		Comment string `json:"comment"`  // 教师评语
		ClassID *int64 `json:"class_id"` // 分配的班级ID
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// 允许空 body，保持向后兼容
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	relationRepo := database.NewRelationRepository(db)
	classRepo := database.NewClassRepository(db)

	// 查询关系记录
	rel, err := relationRepo.GetByID(relID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询关系失败: "+err.Error())
		return
	}
	if rel == nil {
		Error(c, http.StatusNotFound, 40005, "关系不存在")
		return
	}

	// 校验 teacher_id == 当前教师
	if rel.TeacherID != teacherID {
		Error(c, http.StatusForbidden, 40003, "无权操作此关系")
		return
	}

	// 如果指定了 class_id，校验班级存在且属于当前教师分身
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
		if class.PersonaID != rel.TeacherPersonaID {
			Error(c, http.StatusForbidden, 40018, "无权操作该班级")
			return
		}
	}

	// 更新状态为 approved，同时保存评语和班级
	if err := relationRepo.ApproveWithDetails(relID, req.Comment, req.ClassID); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "更新关系状态失败: "+err.Error())
		return
	}

	// 如果指定了班级，将学生加入班级
	var joinedClass bool
	if req.ClassID != nil && rel.StudentPersonaID > 0 {
		isMember, _ := classRepo.IsMember(*req.ClassID, rel.StudentPersonaID)
		if !isMember {
			_, err = classRepo.AddMember(*req.ClassID, rel.StudentPersonaID)
			if err == nil {
				joinedClass = true
			}
		}
	}

	Success(c, gin.H{
		"id":                 relID,
		"status":             "approved",
		"comment":            req.Comment,
		"class_id":           req.ClassID,
		"joined_class":       joinedClass,
		"teacher_persona_id": rel.TeacherPersonaID,
		"student_persona_id": rel.StudentPersonaID,
		"updated_at":         time.Now(),
	})
}

// HandleRejectRelation 教师审批拒绝
// PUT /api/relations/:id/reject
func (h *Handler) HandleRejectRelation(c *gin.Context) {
	// 从 JWT 获取当前教师 user_id
	userID, _ := c.Get("user_id")
	teacherID, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	// 从路径获取关系 id
	idStr := c.Param("id")
	relID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || relID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的关系 ID")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	relationRepo := database.NewRelationRepository(db)

	// 查询关系记录
	rel, err := relationRepo.GetByID(relID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询关系失败: "+err.Error())
		return
	}
	if rel == nil {
		Error(c, http.StatusNotFound, 40005, "关系不存在")
		return
	}

	// 校验 teacher_id == 当前教师
	if rel.TeacherID != teacherID {
		Error(c, http.StatusForbidden, 40003, "无权操作此关系")
		return
	}

	// 更新状态为 rejected
	if err := relationRepo.UpdateStatus(relID, "rejected"); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "更新关系状态失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"id":                 relID,
		"status":             "rejected",
		"teacher_persona_id": rel.TeacherPersonaID,
		"student_persona_id": rel.StudentPersonaID,
		"updated_at":         time.Now(),
	})
}

// HandleGetRelations 获取师生关系列表
// GET /api/relations
func (h *Handler) HandleGetRelations(c *gin.Context) {
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

	relationRepo := database.NewRelationRepository(db)

	// 分身维度查询（persona_id > 0 时优先使用）
	if personaIDInt64 > 0 {
		if roleStr == "teacher" {
			rels, total, err := relationRepo.ListByTeacherPersona(personaIDInt64, status, offset, pageSize)
			if err != nil {
				Error(c, http.StatusInternalServerError, 50001, "查询关系列表失败: "+err.Error())
				return
			}
			SuccessPage(c, rels, total, page, pageSize)
		} else {
			rels, total, err := relationRepo.ListByStudentPersona(personaIDInt64, status, offset, pageSize)
			if err != nil {
				Error(c, http.StatusInternalServerError, 50001, "查询关系列表失败: "+err.Error())
				return
			}
			SuccessPage(c, rels, total, page, pageSize)
		}
		return
	}

	// 向后兼容：persona_id = 0 时回退到 user_id 维度查询
	if roleStr == "teacher" {
		rels, total, err := relationRepo.ListByTeacherWithStudent(userIDInt64, status, offset, pageSize)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "查询关系列表失败: "+err.Error())
			return
		}
		SuccessPage(c, rels, total, page, pageSize)
	} else {
		rels, total, err := relationRepo.ListByStudentWithTeacher(userIDInt64, status, offset, pageSize)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "查询关系列表失败: "+err.Error())
			return
		}
		SuccessPage(c, rels, total, page, pageSize)
	}
}

// ======================== V2.0 迭代3 新增方法 ========================

// HandleToggleRelation 启用/停用学生访问权限
// PUT /api/relations/:id/toggle
func (h *Handler) HandleToggleRelation(c *gin.Context) {
	idStr := c.Param("id")
	relID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || relID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的关系ID")
		return
	}

	userID, _ := c.Get("user_id")
	teacherID, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	var req struct {
		IsActive *bool `json:"is_active" binding:"required"`
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

	relationRepo := database.NewRelationRepository(db)

	// 查询关系记录
	rel, err := relationRepo.GetByID(relID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询关系失败: "+err.Error())
		return
	}
	if rel == nil {
		Error(c, http.StatusNotFound, 40007, "师生关系不存在")
		return
	}

	// 校验 teacher_id == 当前教师
	if rel.TeacherID != teacherID {
		Error(c, http.StatusForbidden, 40014, "无权操作该师生关系")
		return
	}

	// 执行启停
	isActiveInt := 0
	if *req.IsActive {
		isActiveInt = 1
	}

	result, err := relationRepo.ToggleRelation(relID, isActiveInt)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "更新师生关系状态失败: "+err.Error())
		return
	}

	Success(c, result)
}
