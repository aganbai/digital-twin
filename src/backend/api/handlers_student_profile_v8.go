package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// UpdateStudentProfileRequest 更新学生基础信息请求
type UpdateStudentProfileRequest struct {
	Age        *int    `json:"age"`         // 年龄，可选
	Gender     *string `json:"gender"`      // 性别，可选，"male"/"female"
	FamilyInfo *string `json:"family_info"` // 家庭情况，可选
}

// HandleUpdateStudentProfile 更新学生基础信息
// PUT /api/user/student-profile
// 将信息保存到 class_members 表中该学生的所有记录
func (h *Handler) HandleUpdateStudentProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	studentID := userID.(int64)

	var req UpdateStudentProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "请求参数错误: " + err.Error()})
		return
	}

	// 至少提供一个字段
	if req.Age == nil && req.Gender == nil && req.FamilyInfo == nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "至少需要提供一个更新字段"})
		return
	}

	// 验证性别字段
	if req.Gender != nil && *req.Gender != "male" && *req.Gender != "female" && *req.Gender != "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "gender 只能为 'male' 或 'female'"})
		return
	}

	// 验证年龄字段
	if req.Age != nil && (*req.Age < 0 || *req.Age > 150) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: 40001, Message: "age 必须在 0-150 之间"})
		return
	}

	// 获取该学生的所有分身ID
	rows, err := h.db.DB.Query(`SELECT id FROM personas WHERE user_id = ? AND role = 'student'`, studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "查询分身失败"})
		return
	}
	defer rows.Close()

	var personaIDs []int64
	for rows.Next() {
		var pid int64
		if err := rows.Scan(&pid); err != nil {
			continue
		}
		personaIDs = append(personaIDs, pid)
	}

	if len(personaIDs) == 0 {
		// 没有分身，直接返回成功（学生可能还未创建分身或加入班级）
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
		return
	}

	// 动态构建 UPDATE 语句，只更新提供的字段
	setClauses := []string{}
	args := []interface{}{}

	if req.Age != nil {
		setClauses = append(setClauses, "age = ?")
		args = append(args, *req.Age)
	}
	if req.Gender != nil {
		setClauses = append(setClauses, "gender = ?")
		args = append(args, *req.Gender)
	}
	if req.FamilyInfo != nil {
		setClauses = append(setClauses, "family_info = ?")
		args = append(args, *req.FamilyInfo)
	}

	// 构建 IN 子句
	placeholders := ""
	for i, pid := range personaIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args = append(args, pid)
	}

	query := "UPDATE class_members SET "
	for i, clause := range setClauses {
		if i > 0 {
			query += ", "
		}
		query += clause
	}
	query += " WHERE student_persona_id IN (" + placeholders + ")"

	_, err = h.db.DB.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: 50001, Message: "更新学生信息失败: " + err.Error()})
		return
	}

	// 同时更新 class_join_requests 表中该学生的待处理申请
	reqArgs := []interface{}{}
	reqSetClauses := []string{}
	if req.Age != nil {
		reqSetClauses = append(reqSetClauses, "student_age = ?")
		reqArgs = append(reqArgs, *req.Age)
	}
	if req.Gender != nil {
		reqSetClauses = append(reqSetClauses, "student_gender = ?")
		reqArgs = append(reqArgs, *req.Gender)
	}
	if req.FamilyInfo != nil {
		reqSetClauses = append(reqSetClauses, "student_family_info = ?")
		reqArgs = append(reqArgs, *req.FamilyInfo)
	}

	if len(reqSetClauses) > 0 {
		reqQuery := "UPDATE class_join_requests SET "
		for i, clause := range reqSetClauses {
			if i > 0 {
				reqQuery += ", "
			}
			reqQuery += clause
		}
		reqQuery += " WHERE student_id = ? AND status = 'pending'"
		reqArgs = append(reqArgs, studentID)
		// 忽略错误，这是辅助更新
		_, _ = h.db.DB.Exec(reqQuery, reqArgs...)
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
}
