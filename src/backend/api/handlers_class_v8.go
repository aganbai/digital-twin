package api

import (
	"crypto/rand"
	"digital-twin/src/backend/database"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// validAgeGroups 合法的年龄段列表
var validAgeGroups = map[string]bool{
	"学前":    true,
	"小学低年级": true,
	"小学高年级": true,
	"初中":    true,
	"高中":    true,
	"成人":    true,
}

// CreateClassRequestV8 创建班级请求（迭代8扩展版）
type CreateClassRequestV8 struct {
	Name               string   `json:"name" binding:"required"`
	Description        string   `json:"description"`
	TeacherDisplayName string   `json:"teacher_display_name"`
	Subject            string   `json:"subject"`
	AgeGroup           []string `json:"age_group"`
}

// CreateClassResponseV8 创建班级响应
type CreateClassResponseV8 struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	InviteCode string `json:"invite_code"`
	ShareLink  string `json:"share_link"`
	QRCodeURL  string `json:"qr_code_url"`
}

// HandleCreateClassV8 创建班级（迭代8增强版）
func (h *Handler) HandleCreateClassV8(c *gin.Context) {
	userID, _ := c.Get("user_id")
	teacherID := userID.(int64)

	var req CreateClassRequestV8
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "请求参数错误: " + err.Error()})
		return
	}

	// 验证 AgeGroup
	if len(req.AgeGroup) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "age_group 不能为空"})
		return
	}
	for _, ag := range req.AgeGroup {
		if !validAgeGroups[ag] {
			c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "无效的 age_group 值: " + ag})
			return
		}
	}

	// 将 AgeGroup 序列化为 JSON 字符串存入数据库
	ageGroupJSON, err := json.Marshal(req.AgeGroup)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "序列化 age_group 失败"})
		return
	}

	// 获取教师当前分身
	personaID, err := h.getDefaultPersonaID(teacherID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "获取分身失败"})
		return
	}

	// 生成邀请码
	inviteCode := generateInviteCode()
	shareLink := fmt.Sprintf("/pages/share-join/index?code=%s", inviteCode)
	qrCodeURL := fmt.Sprintf("/api/qrcode?text=%s", inviteCode) // 简化处理

	// 创建班级
	result, err := h.db.DB.Exec(`
		INSERT INTO classes (
			persona_id, name, description,
			teacher_display_name, subject, age_group,
			share_link, invite_code, qr_code_url,
			is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		personaID, req.Name, req.Description,
		req.TeacherDisplayName, req.Subject, string(ageGroupJSON),
		shareLink, inviteCode, qrCodeURL,
		time.Now(), time.Now(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "创建班级失败: " + err.Error()})
		return
	}

	classID, _ := result.LastInsertId()

	c.JSON(http.StatusOK, CreateClassResponseV8{
		ID:         classID,
		Name:       req.Name,
		InviteCode: inviteCode,
		ShareLink:  shareLink,
		QRCodeURL:  qrCodeURL,
	})
}

// generateInviteCode 生成邀请码
func generateInviteCode() string {
	bytes := make([]byte, 6)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:8]
}

// GetClassShareInfoResponse 班级分享信息响应
type GetClassShareInfoResponse struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	TeacherDisplayName string `json:"teacher_display_name"`
	Subject            string `json:"subject"`
	AgeGroup           string `json:"age_group"`
	ShareLink          string `json:"share_link"`
	InviteCode         string `json:"invite_code"`
	QRCodeURL          string `json:"qr_code_url"`
	MemberCount        int    `json:"member_count"`
}

// HandleGetClassShareInfo 获取班级分享信息
func (h *Handler) HandleGetClassShareInfo(c *gin.Context) {
	userID, _ := c.Get("user_id")
	teacherID := userID.(int64)

	classIDStr := c.Param("id")
	classID, err := strconv.ParseInt(classIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "无效的班级ID"})
		return
	}

	// 验证班级所有权
	var personaID int64
	err = h.db.DB.QueryRow(`
		SELECT persona_id FROM classes WHERE id = ?`, classID).Scan(&personaID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Code: 40401, Message: "班级不存在"})
		return
	}

	// 验证教师权限
	var ownerID int64
	err = h.db.DB.QueryRow(`
		SELECT user_id FROM personas WHERE id = ?`, personaID).Scan(&ownerID)
	if err != nil || ownerID != teacherID {
		c.JSON(http.StatusForbidden, ErrorResponse{Code: 40301, Message: "无权访问此班级"})
		return
	}

	// 查询班级信息
	var info GetClassShareInfoResponse
	var memberCount int
	err = h.db.DB.QueryRow(`
		SELECT id, name, description, teacher_display_name, subject, age_group,
			share_link, invite_code, qr_code_url
		FROM classes WHERE id = ?`, classID).Scan(
		&info.ID, &info.Name, &info.Description, &info.TeacherDisplayName,
		&info.Subject, &info.AgeGroup, &info.ShareLink, &info.InviteCode, &info.QRCodeURL,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "查询失败"})
		return
	}

	// 查询成员数
	h.db.DB.QueryRow(`
		SELECT COUNT(*) FROM class_members WHERE class_id = ?`, classID).Scan(&memberCount)
	info.MemberCount = memberCount

	c.JSON(http.StatusOK, info)
}

// JoinClassRequest 加入班级请求
type JoinClassRequest struct {
	InviteCode     string `json:"invite_code" binding:"required"`
	RequestMessage string `json:"request_message"`
	Age            int    `json:"age"`
	Gender         string `json:"gender"`
	FamilyInfo     string `json:"family_info"`
}

// HandleJoinClass 申请加入班级
func (h *Handler) HandleJoinClass(c *gin.Context) {
	userID, _ := c.Get("user_id")
	studentID := userID.(int64)

	var req JoinClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "请求参数错误"})
		return
	}

	// 获取学生分身ID
	studentPersonaID, err := h.getDefaultPersonaID(studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "获取分身失败"})
		return
	}

	// 根据邀请码查找班级
	var classID int64
	err = h.db.DB.QueryRow(`
		SELECT id FROM classes WHERE invite_code = ?`, req.InviteCode).Scan(&classID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Code: 40401, Message: "无效的邀请码"})
		return
	}

	// 检查是否已在班级中
	var exists int
	h.db.DB.QueryRow(`
		SELECT COUNT(*) FROM class_members WHERE class_id = ? AND student_persona_id = ?`,
		classID, studentPersonaID).Scan(&exists)
	if exists > 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40002, Message: "已在此班级中"})
		return
	}

	// 创建加入申请
	requestRepo := database.NewClassJoinRequestRepository(h.db)
	reqObj := &database.ClassJoinRequest{
		ClassID:           classID,
		StudentPersonaID:  studentPersonaID,
		StudentID:         studentID,
		Status:            "pending",
		RequestMessage:    req.RequestMessage,
		StudentAge:        req.Age,
		StudentGender:     req.Gender,
		StudentFamilyInfo: req.FamilyInfo,
		RequestTime:       time.Now(),
	}

	requestID, err := requestRepo.CreateJoinRequest(reqObj)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "创建申请失败: " + err.Error()})
		return
	}

	// 检查基础信息完整性
	profileIncomplete := req.Age == 0 || req.Gender == ""

	response := gin.H{
		"request_id": requestID,
		"status":     "pending",
		"message":    "申请已提交，等待教师审批",
	}
	if profileIncomplete {
		response["profile_incomplete"] = true
		response["message"] = "申请已提交，等待教师审批。建议完善个人基础信息"
	}

	c.JSON(http.StatusOK, response)
}

// GetPendingRequestsResponse 待审批列表响应
type GetPendingRequestsResponse struct {
	Requests []database.ClassJoinRequestItem `json:"requests"`
	Total    int                             `json:"total"`
}

// HandleGetPendingRequests 获取待审批列表
func (h *Handler) HandleGetPendingRequests(c *gin.Context) {
	userID, _ := c.Get("user_id")
	teacherID := userID.(int64)

	requestRepo := database.NewClassJoinRequestRepository(h.db)
	requests, err := requestRepo.GetPendingRequestsByTeacher(teacherID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "查询失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, GetPendingRequestsResponse{
		Requests: requests,
		Total:    len(requests),
	})
}

// ApproveRequestRequest 审批请求
type ApproveRequestRequest struct {
	TeacherEvaluation string  `json:"teacher_evaluation"`
	Age               *int    `json:"age,omitempty"`
	Gender            *string `json:"gender,omitempty"`
	FamilyInfo        *string `json:"family_info,omitempty"`
}

// HandleApproveJoinRequest 审批通过
func (h *Handler) HandleApproveJoinRequest(c *gin.Context) {
	userID, _ := c.Get("user_id")
	teacherID := userID.(int64)

	requestIDStr := c.Param("id")
	requestID, err := strconv.ParseInt(requestIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "无效的请求ID"})
		return
	}

	var req ApproveRequestRequest
	c.ShouldBindJSON(&req)

	requestRepo := database.NewClassJoinRequestRepository(h.db)

	// 获取申请详情
	request, err := requestRepo.GetJoinRequestByID(requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "查询失败"})
		return
	}
	if request == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Code: 40401, Message: "申请不存在"})
		return
	}

	// 校验申请状态：只有 pending 状态才能审批
	if request.Status != "pending" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40002, Message: "该申请已处理，当前状态: " + request.Status})
		return
	}

	// 验证教师权限
	var classPersonaID int64
	err = h.db.DB.QueryRow(`
		SELECT persona_id FROM classes WHERE id = ?`, request.ClassID).Scan(&classPersonaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "查询班级失败"})
		return
	}

	var ownerID int64
	err = h.db.DB.QueryRow(`
		SELECT user_id FROM personas WHERE id = ?`, classPersonaID).Scan(&ownerID)
	if err != nil || ownerID != teacherID {
		c.JSON(http.StatusForbidden, ErrorResponse{Code: 40301, Message: "无权审批此申请"})
		return
	}

	// 更新申请状态
	if err := requestRepo.ApproveJoinRequest(requestID, req.TeacherEvaluation); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "审批失败: " + err.Error()})
		return
	}

	// 如果教师提供了修改字段，用教师提供的值覆盖学生申请值
	finalAge := request.StudentAge
	finalGender := request.StudentGender
	finalFamilyInfo := request.StudentFamilyInfo
	if req.Age != nil {
		finalAge = *req.Age
	}
	if req.Gender != nil {
		finalGender = *req.Gender
	}
	if req.FamilyInfo != nil {
		finalFamilyInfo = *req.FamilyInfo
	}

	// 添加到班级成员
	_, err = h.db.DB.Exec(`
		INSERT INTO class_members (
			class_id, student_persona_id, joined_at,
			approval_status, teacher_evaluation, age, gender, family_info,
			request_time, approval_time
		) VALUES (?, ?, ?, 'approved', ?, ?, ?, ?, ?, ?)
		ON CONFLICT(class_id, student_persona_id) DO UPDATE SET
			approval_status = 'approved',
			teacher_evaluation = excluded.teacher_evaluation,
			age = excluded.age,
			gender = excluded.gender,
			family_info = excluded.family_info,
			approval_time = excluded.approval_time`,
		request.ClassID, request.StudentPersonaID, time.Now(),
		req.TeacherEvaluation, finalAge, finalGender,
		finalFamilyInfo, request.RequestTime, time.Now(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "添加成员失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "已批准加入申请",
	})
}

// HandleRejectJoinRequest 审批拒绝
func (h *Handler) HandleRejectJoinRequest(c *gin.Context) {
	userID, _ := c.Get("user_id")
	teacherID := userID.(int64)

	requestIDStr := c.Param("id")
	requestID, err := strconv.ParseInt(requestIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "无效的请求ID"})
		return
	}

	var req ApproveRequestRequest
	c.ShouldBindJSON(&req)

	requestRepo := database.NewClassJoinRequestRepository(h.db)

	// 获取申请详情
	request, err := requestRepo.GetJoinRequestByID(requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "查询失败"})
		return
	}
	if request == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Code: 40401, Message: "申请不存在"})
		return
	}

	// 校验申请状态：只有 pending 状态才能拒绝
	if request.Status != "pending" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40002, Message: "该申请已处理，当前状态: " + request.Status})
		return
	}

	// 验证教师权限
	var classPersonaID int64
	err = h.db.DB.QueryRow(`
		SELECT persona_id FROM classes WHERE id = ?`, request.ClassID).Scan(&classPersonaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "查询班级失败"})
		return
	}

	var ownerID int64
	err = h.db.DB.QueryRow(`
		SELECT user_id FROM personas WHERE id = ?`, classPersonaID).Scan(&ownerID)
	if err != nil || ownerID != teacherID {
		c.JSON(http.StatusForbidden, ErrorResponse{Code: 40301, Message: "无权审批此申请"})
		return
	}

	// 更新申请状态
	if err := requestRepo.RejectJoinRequest(requestID, req.TeacherEvaluation); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "拒绝失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "已拒绝加入申请",
	})
}

// GetClassMembersV8Response 班级成员列表响应（迭代8增强版）
type GetClassMembersV8Response struct {
	Members []ClassMemberV8 `json:"members"`
	Total   int             `json:"total"`
}

// ClassMemberV8 班级成员信息（迭代8增强版）
type ClassMemberV8 struct {
	ID                int64      `json:"id"`
	StudentPersonaID  int64      `json:"student_persona_id"`
	StudentNickname   string     `json:"student_nickname"`
	StudentAvatar     string     `json:"student_avatar"`
	Age               int        `json:"age"`
	Gender            string     `json:"gender"`
	FamilyInfo        string     `json:"family_info,omitempty"`
	TeacherEvaluation string     `json:"teacher_evaluation,omitempty"`
	JoinedAt          time.Time  `json:"joined_at"`
	ApprovalTime      *time.Time `json:"approval_time,omitempty"`
}

// HandleGetClassMembersV8 获取班级成员列表（迭代8增强版）
func (h *Handler) HandleGetClassMembersV8(c *gin.Context) {
	userID, _ := c.Get("user_id")
	teacherID := userID.(int64)

	classIDStr := c.Param("id")
	classID, err := strconv.ParseInt(classIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "无效的班级ID"})
		return
	}

	// 验证班级所有权
	var personaID int64
	err = h.db.DB.QueryRow(`
		SELECT persona_id FROM classes WHERE id = ?`, classID).Scan(&personaID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Code: 40401, Message: "班级不存在"})
		return
	}

	var ownerID int64
	err = h.db.DB.QueryRow(`
		SELECT user_id FROM personas WHERE id = ?`, personaID).Scan(&ownerID)
	if err != nil || ownerID != teacherID {
		c.JSON(http.StatusForbidden, ErrorResponse{Code: 40301, Message: "无权访问此班级"})
		return
	}

	// 查询成员列表
	rows, err := h.db.DB.Query(`
		SELECT 
			cm.id, cm.student_persona_id, p.nickname, p.avatar,
			cm.age, cm.gender, cm.family_info, cm.teacher_evaluation,
			cm.joined_at, cm.approval_time
		FROM class_members cm
		JOIN personas p ON cm.student_persona_id = p.id
		WHERE cm.class_id = ?
		ORDER BY cm.joined_at DESC`, classID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "查询失败"})
		return
	}
	defer rows.Close()

	var members []ClassMemberV8
	for rows.Next() {
		var m ClassMemberV8
		err := rows.Scan(
			&m.ID, &m.StudentPersonaID, &m.StudentNickname, &m.StudentAvatar,
			&m.Age, &m.Gender, &m.FamilyInfo, &m.TeacherEvaluation,
			&m.JoinedAt, &m.ApprovalTime,
		)
		if err != nil {
			continue
		}
		members = append(members, m)
	}

	c.JSON(http.StatusOK, GetClassMembersV8Response{
		Members: members,
		Total:   len(members),
	})
}
