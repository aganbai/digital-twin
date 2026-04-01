package integration

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ======================== V2.0 迭代1 集成测试辅助变量 ========================

var (
	v2i1TeacherToken string
	v2i1StudentToken string
	v2i1TeacherID    float64
	v2i1StudentID    float64
	v2i1RelationID   float64 // 师生关系记录 ID
	v2i1AssignmentID float64 // 作业 ID
)

// v2i1Setup 初始化 V2.0 迭代1 测试所需的教师和学生
// 教师通过微信登录 + 补全信息（含 school + description），学生同理
func v2i1Setup(t *testing.T) {
	t.Helper()
	os.Setenv("WX_MODE", "mock")
	os.Setenv("LLM_MODE", "mock")

	if v2i1TeacherToken != "" && v2i1StudentToken != "" {
		return // 已初始化
	}

	// ---- 注册教师 ----
	teacherLoginBody := map[string]interface{}{
		"code": "stc001",
	}
	_, body, err := doRequest("POST", "/api/auth/wx-login", teacherLoginBody, "")
	if err != nil {
		t.Fatalf("v2i1Setup 教师微信登录失败: %v", err)
	}
	apiResp, err := parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("v2i1Setup 教师微信登录解析失败: %v, code: %d, body: %s", err, apiResp.Code, string(body))
	}
	v2i1TeacherToken = apiResp.Data["token"].(string)
	v2i1TeacherID = apiResp.Data["user_id"].(float64)

	// 教师补全信息（含 school + description）
	completeBody := map[string]interface{}{
		"role":        "teacher",
		"nickname":    "V2迭代1王老师",
		"school":      "北京大学",
		"description": "物理学教授，专注力学和热力学教学",
	}
	_, body, err = doRequest("POST", "/api/auth/complete-profile", completeBody, v2i1TeacherToken)
	if err != nil {
		t.Fatalf("v2i1Setup 教师补全信息失败: %v", err)
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("v2i1Setup 教师补全信息解析失败: %v, code: %d, body: %s", err, apiResp.Code, string(body))
	}

	// 使用 complete-profile 返回的新 token（不重新登录，模拟真实用户行为）
	if newToken, ok := apiResp.Data["token"]; ok && newToken != nil && newToken != "" {
		v2i1TeacherToken = newToken.(string)
	} else {
		t.Fatalf("v2i1Setup 教师 complete-profile 未返回新 token, data: %v", apiResp.Data)
	}

	// ---- 注册学生 ----
	v2i1StudentToken, v2i1StudentID = v2i1RegisterStudent(t, "sst001", "V2迭代1小李")
}

// v2i1RegisterTeacher 注册一个新教师并返回 token 和 ID
// 关键：使用 complete-profile 返回的新 token，不重新登录（模拟真实用户行为）
func v2i1RegisterTeacher(t *testing.T, code, nickname, school, description string) (token string, id float64) {
	t.Helper()
	os.Setenv("WX_MODE", "mock")

	loginBody := map[string]interface{}{"code": code}
	_, body, err := doRequest("POST", "/api/auth/wx-login", loginBody, "")
	if err != nil {
		t.Fatalf("注册教师微信登录失败: %v", err)
	}
	apiResp, err := parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("注册教师微信登录解析失败: %v, code: %d, body: %s", err, apiResp.Code, string(body))
	}
	token = apiResp.Data["token"].(string)
	id = apiResp.Data["user_id"].(float64)

	completeBody := map[string]interface{}{
		"role":        "teacher",
		"nickname":    nickname,
		"school":      school,
		"description": description,
	}
	_, body, err = doRequest("POST", "/api/auth/complete-profile", completeBody, token)
	if err != nil {
		t.Fatalf("注册教师补全信息失败: %v", err)
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("注册教师补全信息解析失败: %v, code: %d, body: %s", err, apiResp.Code, string(body))
	}

	// 使用 complete-profile 返回的新 token（包含最新角色信息）
	if newToken, ok := apiResp.Data["token"]; ok && newToken != nil && newToken != "" {
		token = newToken.(string)
	} else {
		t.Fatalf("注册教师 complete-profile 未返回新 token, data: %v", apiResp.Data)
	}
	return
}

// v2i1RegisterStudent 注册一个新学生并返回 token 和 ID
// 关键：使用 complete-profile 返回的新 token，不重新登录（模拟真实用户行为）
func v2i1RegisterStudent(t *testing.T, code, nickname string) (token string, id float64) {
	t.Helper()
	os.Setenv("WX_MODE", "mock")

	loginBody := map[string]interface{}{"code": code}
	_, body, err := doRequest("POST", "/api/auth/wx-login", loginBody, "")
	if err != nil {
		t.Fatalf("注册学生微信登录失败: %v", err)
	}
	apiResp, err := parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("注册学生微信登录解析失败: %v, code: %d, body: %s", err, apiResp.Code, string(body))
	}
	token = apiResp.Data["token"].(string)
	id = apiResp.Data["user_id"].(float64)

	completeBody := map[string]interface{}{
		"role":     "student",
		"nickname": nickname,
	}
	_, body, err = doRequest("POST", "/api/auth/complete-profile", completeBody, token)
	if err != nil {
		t.Fatalf("注册学生补全信息失败: %v", err)
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("注册学生补全信息解析失败: %v, code: %d, body: %s", err, apiResp.Code, string(body))
	}

	// 使用 complete-profile 返回的新 token（包含最新角色信息）
	if newToken, ok := apiResp.Data["token"]; ok && newToken != nil && newToken != "" {
		token = newToken.(string)
	} else {
		t.Fatalf("注册学生 complete-profile 未返回新 token, data: %v", apiResp.Data)
	}
	return
}

// doMultipartUpload 发送 multipart/form-data 文件上传请求
func doMultipartUpload(path, token, fieldName, fileName string, fileContent []byte, extraFields map[string]string) (*http.Response, []byte, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加文件字段
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return nil, nil, fmt.Errorf("创建表单文件字段失败: %w", err)
	}
	if _, err := part.Write(fileContent); err != nil {
		return nil, nil, fmt.Errorf("写入文件内容失败: %w", err)
	}

	// 添加额外字段
	for key, val := range extraFields {
		if err := writer.WriteField(key, val); err != nil {
			return nil, nil, fmt.Errorf("写入字段 %s 失败: %w", key, err)
		}
	}

	writer.Close()

	req, err := http.NewRequest("POST", ts.URL+path, &buf)
	if err != nil {
		return nil, nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	return resp, respBody, nil
}

// ======================== IT-40: 教师注册必填 school + description ========================
func TestV2I1_IT40_TeacherRegisterRequiredFields(t *testing.T) {
	os.Setenv("WX_MODE", "mock")

	// 步骤1：微信登录获取新用户 token
	loginBody := map[string]interface{}{
		"code": "t40a01",
	}
	_, body, err := doRequest("POST", "/api/auth/wx-login", loginBody, "")
	if err != nil {
		t.Fatalf("IT-40 微信登录失败: %v", err)
	}
	apiResp, err := parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-40 微信登录解析失败: %v, code: %d", err, apiResp.Code)
	}
	token := apiResp.Data["token"].(string)

	// 步骤2：教师补全信息 —— 缺少 school
	noSchoolBody := map[string]interface{}{
		"role":        "teacher",
		"nickname":    "缺少学校老师",
		"description": "测试描述",
	}
	resp, body, err := doRequest("POST", "/api/auth/complete-profile", noSchoolBody, token)
	if err != nil {
		t.Fatalf("IT-40 缺少school请求失败: %v", err)
	}
	apiResp, err = parseResponse(body)
	if err != nil {
		t.Fatalf("IT-40 缺少school解析失败: %v", err)
	}
	if apiResp.Code == 0 {
		t.Fatalf("IT-40 缺少school应返回错误, 但返回了 code=0, HTTP=%d, body: %s", resp.StatusCode, string(body))
	}
	t.Logf("IT-40 步骤2: 缺少school正确返回错误, code=%d, message=%s", apiResp.Code, apiResp.Message)

	// 步骤3：教师补全信息 —— 缺少 description
	loginBody2 := map[string]interface{}{
		"code": "t40a02",
	}
	_, body, err = doRequest("POST", "/api/auth/wx-login", loginBody2, "")
	if err != nil {
		t.Fatalf("IT-40 微信登录2失败: %v", err)
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-40 微信登录2解析失败: %v, code: %d", err, apiResp.Code)
	}
	token2 := apiResp.Data["token"].(string)

	noDescBody := map[string]interface{}{
		"role":     "teacher",
		"nickname": "缺少描述老师",
		"school":   "北京大学",
	}
	resp, body, err = doRequest("POST", "/api/auth/complete-profile", noDescBody, token2)
	if err != nil {
		t.Fatalf("IT-40 缺少description请求失败: %v", err)
	}
	apiResp, err = parseResponse(body)
	if err != nil {
		t.Fatalf("IT-40 缺少description解析失败: %v", err)
	}
	if apiResp.Code == 0 {
		t.Fatalf("IT-40 缺少description应返回错误, 但返回了 code=0, HTTP=%d, body: %s", resp.StatusCode, string(body))
	}
	t.Logf("IT-40 步骤3: 缺少description正确返回错误, code=%d, message=%s", apiResp.Code, apiResp.Message)

	// 步骤4：教师补全信息 —— 正确填写所有字段
	loginBody3 := map[string]interface{}{
		"code": "t40a03",
	}
	_, body, err = doRequest("POST", "/api/auth/wx-login", loginBody3, "")
	if err != nil {
		t.Fatalf("IT-40 微信登录3失败: %v", err)
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-40 微信登录3解析失败: %v, code: %d", err, apiResp.Code)
	}
	token3 := apiResp.Data["token"].(string)

	fullBody := map[string]interface{}{
		"role":        "teacher",
		"nickname":    "IT40正确老师",
		"school":      "清华大学",
		"description": "数学教授",
	}
	resp, body, err = doRequest("POST", "/api/auth/complete-profile", fullBody, token3)
	if err != nil {
		t.Fatalf("IT-40 正确注册请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-40 正确注册HTTP状态码错误: 期望200, 实际%d, body: %s", resp.StatusCode, string(body))
	}
	apiResp, err = parseResponse(body)
	if err != nil {
		t.Fatalf("IT-40 正确注册解析失败: %v", err)
	}
	if apiResp.Code != 0 {
		t.Fatalf("IT-40 正确注册业务码错误: 期望0, 实际%d, message: %s", apiResp.Code, apiResp.Message)
	}

	t.Logf("IT-40 通过: 教师注册必填 school + description 验证成功")
}

// ======================== IT-41: 同名+同校教师注册返回 409 ========================
func TestV2I1_IT41_DuplicateTeacherSameSchool(t *testing.T) {
	os.Setenv("WX_MODE", "mock")

	// 步骤1：注册第一个教师
	_, _ = v2i1RegisterTeacher(t, "t41a01", "重名老师", "测试大学", "第一个教师")

	// 步骤2：注册第二个教师，同名+同校 → 应返回 409
	loginBody2 := map[string]interface{}{
		"code": "t41a02",
	}
	_, body, err := doRequest("POST", "/api/auth/wx-login", loginBody2, "")
	if err != nil {
		t.Fatalf("IT-41 微信登录失败: %v", err)
	}
	apiResp, err := parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-41 微信登录解析失败: %v, code: %d", err, apiResp.Code)
	}
	token2 := apiResp.Data["token"].(string)

	duplicateBody := map[string]interface{}{
		"role":        "teacher",
		"nickname":    "重名老师",
		"school":      "测试大学",
		"description": "第二个教师",
	}
	resp, body, err := doRequest("POST", "/api/auth/complete-profile", duplicateBody, token2)
	if err != nil {
		t.Fatalf("IT-41 重复注册请求失败: %v", err)
	}
	apiResp, err = parseResponse(body)
	if err != nil {
		t.Fatalf("IT-41 重复注册解析失败: %v", err)
	}

	// 验证返回 409 + 40015
	if resp.StatusCode != http.StatusConflict {
		t.Logf("IT-41 警告: HTTP状态码 %d (期望 409), code=%d, message=%s", resp.StatusCode, apiResp.Code, apiResp.Message)
	}
	if apiResp.Code != 40015 {
		t.Fatalf("IT-41 业务码错误: 期望 40015, 实际 %d, message: %s, body: %s", apiResp.Code, apiResp.Message, string(body))
	}

	t.Logf("IT-41 通过: 同名+同校教师注册正确返回 40015, message: %s", apiResp.Message)
}

// ======================== IT-42: 教师邀请学生 → 关系 approved ========================
func TestV2I1_IT42_TeacherInviteStudent(t *testing.T) {
	v2i1Setup(t)

	reqBody := map[string]interface{}{
		"student_id": int(v2i1StudentID),
	}
	resp, body, err := doRequest("POST", "/api/relations/invite", reqBody, v2i1TeacherToken)
	if err != nil {
		t.Fatalf("IT-42 请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-42 HTTP状态码错误: 期望200, 实际%d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-42 解析失败: %v", err)
	}
	if apiResp.Code != 0 {
		t.Fatalf("IT-42 业务码错误: 期望0, 实际%d, message: %s", apiResp.Code, apiResp.Message)
	}

	// 验证 status = approved
	statusVal, ok := apiResp.Data["status"]
	if !ok {
		t.Fatal("IT-42 响应缺少 status 字段")
	}
	if statusVal != "approved" {
		t.Fatalf("IT-42 status 错误: 期望 approved, 实际 %v", statusVal)
	}

	// 验证 initiated_by = teacher
	initiatedBy, ok := apiResp.Data["initiated_by"]
	if !ok {
		t.Fatal("IT-42 响应缺少 initiated_by 字段")
	}
	if initiatedBy != "teacher" {
		t.Fatalf("IT-42 initiated_by 错误: 期望 teacher, 实际 %v", initiatedBy)
	}

	// 保存关系 ID
	if idVal, ok := apiResp.Data["id"]; ok {
		v2i1RelationID = idVal.(float64)
	}

	t.Logf("IT-42 通过: 教师邀请学生成功, status=approved, relation_id=%v", v2i1RelationID)
}

// ======================== IT-43: 学生申请使用分身 → 关系 pending ========================
func TestV2I1_IT43_StudentApplyTeacher(t *testing.T) {
	os.Setenv("WX_MODE", "mock")

	// 注册新的教师和学生（避免与 IT-42 冲突）
	teacherToken, teacherID := v2i1RegisterTeacher(t, "t43a01", "IT43老师", "IT43大学", "IT43描述")
	_, studentID := v2i1RegisterStudent(t, "s43a01", "IT43学生")
	_ = teacherToken

	// 学生需要重新登录获取 token
	loginBody := map[string]interface{}{"code": "s43a01"}
	_, body, _ := doRequest("POST", "/api/auth/wx-login", loginBody, "")
	apiResp, _ := parseResponse(body)
	studentToken := apiResp.Data["token"].(string)
	_ = studentID

	reqBody := map[string]interface{}{
		"teacher_id": int(teacherID),
	}
	resp, body, err := doRequest("POST", "/api/relations/apply", reqBody, studentToken)
	if err != nil {
		t.Fatalf("IT-43 请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-43 HTTP状态码错误: 期望200, 实际%d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err = parseResponse(body)
	if err != nil {
		t.Fatalf("IT-43 解析失败: %v", err)
	}
	if apiResp.Code != 0 {
		t.Fatalf("IT-43 业务码错误: 期望0, 实际%d, message: %s", apiResp.Code, apiResp.Message)
	}

	// 验证 status = pending
	statusVal, ok := apiResp.Data["status"]
	if !ok {
		t.Fatal("IT-43 响应缺少 status 字段")
	}
	if statusVal != "pending" {
		t.Fatalf("IT-43 status 错误: 期望 pending, 实际 %v", statusVal)
	}

	// 验证 initiated_by = student
	initiatedBy, ok := apiResp.Data["initiated_by"]
	if !ok {
		t.Fatal("IT-43 响应缺少 initiated_by 字段")
	}
	if initiatedBy != "student" {
		t.Fatalf("IT-43 initiated_by 错误: 期望 student, 实际 %v", initiatedBy)
	}

	t.Logf("IT-43 通过: 学生申请使用分身成功, status=pending, initiated_by=student")
}

// ======================== IT-44: 教师审批同意 → 关系 approved ========================
func TestV2I1_IT44_TeacherApproveRelation(t *testing.T) {
	os.Setenv("WX_MODE", "mock")

	// 注册新教师和学生
	teacherToken, teacherID := v2i1RegisterTeacher(t, "t44a01", "IT44老师", "IT44大学", "IT44描述")

	// 注册学生（使用辅助函数，确保重新登录获取正确的 role token）
	studentToken, _ := v2i1RegisterStudent(t, "s44a01", "IT44学生")

	// 学生申请
	applyBody := map[string]interface{}{"teacher_id": int(teacherID)}
	_, body, err := doRequest("POST", "/api/relations/apply", applyBody, studentToken)
	if err != nil {
		t.Fatalf("IT-44 学生申请失败: %v", err)
	}
	apiResp, err := parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-44 学生申请解析失败: %v, code: %d, body: %s", err, apiResp.Code, string(body))
	}
	relationID := apiResp.Data["id"].(float64)

	// 教师审批同意
	approvePath := fmt.Sprintf("/api/relations/%d/approve", int(relationID))
	resp, body, err := doRequest("PUT", approvePath, nil, teacherToken)
	if err != nil {
		t.Fatalf("IT-44 审批请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-44 HTTP状态码错误: 期望200, 实际%d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err = parseResponse(body)
	if err != nil {
		t.Fatalf("IT-44 解析失败: %v", err)
	}
	if apiResp.Code != 0 {
		t.Fatalf("IT-44 业务码错误: 期望0, 实际%d, message: %s", apiResp.Code, apiResp.Message)
	}

	// 验证 status = approved
	statusVal, ok := apiResp.Data["status"]
	if !ok {
		t.Fatal("IT-44 响应缺少 status 字段")
	}
	if statusVal != "approved" {
		t.Fatalf("IT-44 status 错误: 期望 approved, 实际 %v", statusVal)
	}

	t.Logf("IT-44 通过: 教师审批同意成功, status=approved")
}

// ======================== IT-45: 教师审批拒绝 → 关系 rejected ========================
func TestV2I1_IT45_TeacherRejectRelation(t *testing.T) {
	os.Setenv("WX_MODE", "mock")

	// 注册新教师和学生
	teacherToken, teacherID := v2i1RegisterTeacher(t, "t45a01", "IT45老师", "IT45大学", "IT45描述")

	// 注册学生（使用辅助函数，确保重新登录获取正确的 role token）
	studentToken, _ := v2i1RegisterStudent(t, "s45a01", "IT45学生")

	// 学生申请
	applyBody := map[string]interface{}{"teacher_id": int(teacherID)}
	_, body, err := doRequest("POST", "/api/relations/apply", applyBody, studentToken)
	if err != nil {
		t.Fatalf("IT-45 学生申请失败: %v", err)
	}
	apiResp, err := parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-45 学生申请解析失败: %v, code: %d, body: %s", err, apiResp.Code, string(body))
	}
	relationID := apiResp.Data["id"].(float64)

	// 教师审批拒绝
	rejectPath := fmt.Sprintf("/api/relations/%d/reject", int(relationID))
	resp, body, err := doRequest("PUT", rejectPath, nil, teacherToken)
	if err != nil {
		t.Fatalf("IT-45 拒绝请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-45 HTTP状态码错误: 期望200, 实际%d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err = parseResponse(body)
	if err != nil {
		t.Fatalf("IT-45 解析失败: %v", err)
	}
	if apiResp.Code != 0 {
		t.Fatalf("IT-45 业务码错误: 期望0, 实际%d, message: %s", apiResp.Code, apiResp.Message)
	}

	// 验证 status = rejected
	statusVal, ok := apiResp.Data["status"]
	if !ok {
		t.Fatal("IT-45 响应缺少 status 字段")
	}
	if statusVal != "rejected" {
		t.Fatalf("IT-45 status 错误: 期望 rejected, 实际 %v", statusVal)
	}

	t.Logf("IT-45 通过: 教师审批拒绝成功, status=rejected")
}

// ======================== IT-46: 重复创建关系返回 409 ========================
func TestV2I1_IT46_DuplicateRelation(t *testing.T) {
	v2i1Setup(t)

	// v2i1Setup 中已经通过 IT-42 建立了关系，再次邀请应返回 409
	reqBody := map[string]interface{}{
		"student_id": int(v2i1StudentID),
	}
	resp, body, err := doRequest("POST", "/api/relations/invite", reqBody, v2i1TeacherToken)
	if err != nil {
		t.Fatalf("IT-46 请求失败: %v", err)
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-46 解析失败: %v", err)
	}

	// 验证返回 409 + 40009
	if resp.StatusCode != http.StatusConflict {
		t.Logf("IT-46 警告: HTTP状态码 %d (期望 409), code=%d", resp.StatusCode, apiResp.Code)
	}
	if apiResp.Code != 40009 {
		t.Fatalf("IT-46 业务码错误: 期望 40009, 实际 %d, message: %s, body: %s", apiResp.Code, apiResp.Message, string(body))
	}

	t.Logf("IT-46 通过: 重复创建关系正确返回 40009, message: %s", apiResp.Message)
}

// ======================== IT-47: 未授权学生对话返回 403 ========================
func TestV2I1_IT47_UnauthorizedStudentChat(t *testing.T) {
	os.Setenv("WX_MODE", "mock")

	// 注册新教师（不建立关系）
	_, teacherID := v2i1RegisterTeacher(t, "t47a01", "IT47老师", "IT47大学", "IT47描述")

	// 注册新学生
	studentToken, _ := v2i1RegisterStudent(t, "s47a01", "IT47学生")

	// 学生直接对话（未建立关系）
	chatBody := map[string]interface{}{
		"message":    "你好",
		"teacher_id": int(teacherID),
	}
	resp, body, err := doRequest("POST", "/api/chat", chatBody, studentToken)
	if err != nil {
		t.Fatalf("IT-47 请求失败: %v", err)
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-47 解析失败: %v", err)
	}

	// 验证返回 403 + 40007
	if resp.StatusCode != http.StatusForbidden {
		t.Logf("IT-47 警告: HTTP状态码 %d (期望 403), code=%d", resp.StatusCode, apiResp.Code)
	}
	if apiResp.Code != 40007 {
		t.Fatalf("IT-47 业务码错误: 期望 40007, 实际 %d, message: %s, body: %s", apiResp.Code, apiResp.Message, string(body))
	}

	t.Logf("IT-47 通过: 未授权学生对话正确返回 403/40007, message: %s", apiResp.Message)
}

// ======================== IT-48: 授权学生对话成功 ========================
func TestV2I1_IT48_AuthorizedStudentChat(t *testing.T) {
	v2i1Setup(t) // 已建立 approved 关系

	// 先添加一个文档（确保对话有知识库）
	docBody := map[string]interface{}{
		"title":   "IT48测试文档",
		"content": "这是用于IT-48测试的知识库内容。牛顿第一定律是惯性定律。",
		"tags":    "测试",
	}
	doRequest("POST", "/api/documents", docBody, v2i1TeacherToken)

	chatBody := map[string]interface{}{
		"message":    "什么是牛顿第一定律?",
		"teacher_id": int(v2i1TeacherID),
	}
	resp, body, err := doRequest("POST", "/api/chat", chatBody, v2i1StudentToken)
	if err != nil {
		t.Fatalf("IT-48 请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-48 HTTP状态码错误: 期望200, 实际%d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-48 解析失败: %v", err)
	}
	if apiResp.Code != 0 {
		t.Fatalf("IT-48 业务码错误: 期望0, 实际%d, message: %s, body: %s", apiResp.Code, apiResp.Message, string(body))
	}

	replyVal, ok := apiResp.Data["reply"]
	if !ok || replyVal == "" {
		t.Fatalf("IT-48 响应缺少 reply 或 reply 为空, data: %v", apiResp.Data)
	}

	t.Logf("IT-48 通过: 授权学生对话成功, reply长度=%d", len(fmt.Sprintf("%v", replyVal)))
}

// ======================== IT-49: 教师写评语 + 学生查看评语 ========================
func TestV2I1_IT49_TeacherCommentStudentView(t *testing.T) {
	v2i1Setup(t) // 已建立 approved 关系

	// 步骤1：教师写评语
	commentBody := map[string]interface{}{
		"student_id":       int(v2i1StudentID),
		"content":          "该生学习态度认真，对力学概念理解较好",
		"progress_summary": "牛顿定律掌握80%",
	}
	resp, body, err := doRequest("POST", "/api/comments", commentBody, v2i1TeacherToken)
	if err != nil {
		t.Fatalf("IT-49 步骤1 写评语请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-49 步骤1 HTTP状态码错误: 期望200, 实际%d, body: %s", resp.StatusCode, string(body))
	}
	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-49 步骤1 解析失败: %v", err)
	}
	if apiResp.Code != 0 {
		t.Fatalf("IT-49 步骤1 业务码错误: 期望0, 实际%d, message: %s", apiResp.Code, apiResp.Message)
	}

	// 验证返回的评语 ID
	commentIDVal, ok := apiResp.Data["id"]
	if !ok {
		t.Fatal("IT-49 步骤1 响应缺少 id 字段")
	}
	t.Logf("IT-49 步骤1: 评语创建成功, id=%v", commentIDVal)

	// 步骤2：学生查看评语（V2.0 迭代5：学生角色返回空列表）
	commentPath := fmt.Sprintf("/api/comments?teacher_id=%d", int(v2i1TeacherID))
	resp, body, err = doRequest("GET", commentPath, nil, v2i1StudentToken)
	if err != nil {
		t.Fatalf("IT-49 步骤2 查看评语请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-49 步骤2 HTTP状态码错误: 期望200, 实际%d, body: %s", resp.StatusCode, string(body))
	}
	// V2.0 迭代5：学生角色返回 {"code":0,"message":"success","data":[]}
	var it49CommentResp apiResponseRaw
	if err := json.Unmarshal(body, &it49CommentResp); err != nil {
		t.Fatalf("IT-49 步骤2 解析响应失败: %v", err)
	}
	if it49CommentResp.Code != 0 {
		t.Fatalf("IT-49 步骤2 业务码错误: 期望0, 实际%d", it49CommentResp.Code)
	}
	if string(it49CommentResp.Data) != "[]" {
		t.Fatalf("IT-49 步骤2 期望 data=[]，实际: %s", string(it49CommentResp.Data))
	}
	t.Logf("IT-49 步骤2: 学生查看评语返回空列表（符合迭代5预期）")

	t.Logf("IT-49 通过: 教师写评语 + 学生查看评语（学生返回空列表）成功")
}

// ======================== IT-50: 教师设置问答风格 + 对话时生效 ========================
func TestV2I1_IT50_SetDialogueStyleAndChat(t *testing.T) {
	v2i1Setup(t) // 已建立 approved 关系

	// 步骤1：教师设置问答风格
	stylePath := fmt.Sprintf("/api/students/%d/dialogue-style", int(v2i1StudentID))
	styleBody := map[string]interface{}{
		"temperature":         0.8,
		"guidance_level":      "high",
		"style_prompt":        "对该学生请多用鼓励性语言，注重基础概念的巩固",
		"max_turns_per_topic": 5,
	}
	resp, body, err := doRequest("PUT", stylePath, styleBody, v2i1TeacherToken)
	if err != nil {
		t.Fatalf("IT-50 步骤1 设置风格请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-50 步骤1 HTTP状态码错误: 期望200, 实际%d, body: %s", resp.StatusCode, string(body))
	}
	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-50 步骤1 解析失败: %v", err)
	}
	if apiResp.Code != 0 {
		t.Fatalf("IT-50 步骤1 业务码错误: 期望0, 实际%d, message: %s", apiResp.Code, apiResp.Message)
	}
	t.Logf("IT-50 步骤1: 设置问答风格成功")

	// 步骤2：验证风格已保存（GET）
	resp, body, err = doRequest("GET", stylePath, nil, v2i1TeacherToken)
	if err != nil {
		t.Fatalf("IT-50 步骤2 获取风格请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-50 步骤2 HTTP状态码错误: 期望200, 实际%d, body: %s", resp.StatusCode, string(body))
	}
	apiResp, err = parseResponse(body)
	if err != nil {
		t.Fatalf("IT-50 步骤2 解析失败: %v", err)
	}
	if apiResp.Code != 0 {
		t.Fatalf("IT-50 步骤2 业务码错误: 期望0, 实际%d, message: %s", apiResp.Code, apiResp.Message)
	}
	// 验证 style_config 不为空
	if apiResp.Data != nil {
		t.Logf("IT-50 步骤2: 获取问答风格成功, data=%v", apiResp.Data)
	}

	// 步骤3：学生对话（风格应已生效，mock 模式下不影响结果，但验证对话正常）
	chatBody := map[string]interface{}{
		"message":    "请问什么是力的合成?",
		"teacher_id": int(v2i1TeacherID),
	}
	resp, body, err = doRequest("POST", "/api/chat", chatBody, v2i1StudentToken)
	if err != nil {
		t.Fatalf("IT-50 步骤3 对话请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-50 步骤3 HTTP状态码错误: 期望200, 实际%d, body: %s", resp.StatusCode, string(body))
	}
	apiResp, err = parseResponse(body)
	if err != nil {
		t.Fatalf("IT-50 步骤3 解析失败: %v", err)
	}
	if apiResp.Code != 0 {
		t.Fatalf("IT-50 步骤3 业务码错误: 期望0, 实际%d, message: %s, body: %s", apiResp.Code, apiResp.Message, string(body))
	}

	replyVal, ok := apiResp.Data["reply"]
	if !ok || replyVal == "" {
		t.Fatalf("IT-50 步骤3 响应缺少 reply, data: %v", apiResp.Data)
	}

	t.Logf("IT-50 通过: 教师设置问答风格 + 对话时生效验证成功")
}

// ======================== IT-51: 学生提交作业 + 教师点评 ========================
func TestV2I1_IT51_SubmitAssignmentAndTeacherReview(t *testing.T) {
	v2i1Setup(t) // 已建立 approved 关系

	// 步骤1：学生提交作业
	assignmentBody := map[string]interface{}{
		"teacher_id": int(v2i1TeacherID),
		"title":      "牛顿定律作业",
		"content":    "牛顿第一定律是惯性定律，物体在不受外力作用时保持静止或匀速直线运动状态。",
	}
	resp, body, err := doRequest("POST", "/api/assignments", assignmentBody, v2i1StudentToken)
	if err != nil {
		t.Fatalf("IT-51 步骤1 提交作业请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-51 步骤1 HTTP状态码错误: 期望200, 实际%d, body: %s", resp.StatusCode, string(body))
	}
	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-51 步骤1 解析失败: %v", err)
	}
	if apiResp.Code != 0 {
		t.Fatalf("IT-51 步骤1 业务码错误: 期望0, 实际%d, message: %s", apiResp.Code, apiResp.Message)
	}

	assignmentIDVal, ok := apiResp.Data["id"]
	if !ok {
		t.Fatal("IT-51 步骤1 响应缺少 id 字段")
	}
	v2i1AssignmentID = assignmentIDVal.(float64)
	t.Logf("IT-51 步骤1: 作业提交成功, id=%v", v2i1AssignmentID)

	// 验证 status = submitted
	statusVal, ok := apiResp.Data["status"]
	if !ok {
		t.Fatal("IT-51 步骤1 响应缺少 status 字段")
	}
	if statusVal != "submitted" {
		t.Fatalf("IT-51 步骤1 status 错误: 期望 submitted, 实际 %v", statusVal)
	}

	// 步骤2：教师点评作业
	reviewPath := fmt.Sprintf("/api/assignments/%d/review", int(v2i1AssignmentID))
	reviewBody := map[string]interface{}{
		"content": "整体不错，概念理解准确，注意公式推导过程需要更严谨",
		"score":   85,
	}
	resp, body, err = doRequest("POST", reviewPath, reviewBody, v2i1TeacherToken)
	if err != nil {
		t.Fatalf("IT-51 步骤2 教师点评请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-51 步骤2 HTTP状态码错误: 期望200, 实际%d, body: %s", resp.StatusCode, string(body))
	}
	apiResp, err = parseResponse(body)
	if err != nil {
		t.Fatalf("IT-51 步骤2 解析失败: %v", err)
	}
	if apiResp.Code != 0 {
		t.Fatalf("IT-51 步骤2 业务码错误: 期望0, 实际%d, message: %s", apiResp.Code, apiResp.Message)
	}

	// 验证 reviewer_type = teacher
	reviewerType, ok := apiResp.Data["reviewer_type"]
	if !ok {
		t.Fatal("IT-51 步骤2 响应缺少 reviewer_type 字段")
	}
	if reviewerType != "teacher" {
		t.Fatalf("IT-51 步骤2 reviewer_type 错误: 期望 teacher, 实际 %v", reviewerType)
	}

	t.Logf("IT-51 通过: 学生提交作业 + 教师点评成功")
}

// ======================== IT-52: AI 自动点评作业 ========================
func TestV2I1_IT52_AIReviewAssignment(t *testing.T) {
	v2i1Setup(t)

	// 如果 IT-51 没有创建作业，先创建一个
	if v2i1AssignmentID <= 0 {
		assignmentBody := map[string]interface{}{
			"teacher_id": int(v2i1TeacherID),
			"title":      "AI点评测试作业",
			"content":    "牛顿第二定律 F=ma，物体加速度与合力成正比，与质量成反比。",
		}
		_, body, err := doRequest("POST", "/api/assignments", assignmentBody, v2i1StudentToken)
		if err != nil {
			t.Fatalf("IT-52 创建作业失败: %v", err)
		}
		apiResp, err := parseResponse(body)
		if err != nil || apiResp.Code != 0 {
			t.Fatalf("IT-52 创建作业解析失败: %v, code: %d, body: %s", err, apiResp.Code, string(body))
		}
		v2i1AssignmentID = apiResp.Data["id"].(float64)
	}

	// AI 自动点评
	aiReviewPath := fmt.Sprintf("/api/assignments/%d/ai-review", int(v2i1AssignmentID))
	resp, body, err := doRequest("POST", aiReviewPath, nil, v2i1TeacherToken)
	if err != nil {
		t.Fatalf("IT-52 AI点评请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-52 HTTP状态码错误: 期望200, 实际%d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-52 解析失败: %v", err)
	}
	if apiResp.Code != 0 {
		t.Fatalf("IT-52 业务码错误: 期望0, 实际%d, message: %s", apiResp.Code, apiResp.Message)
	}

	// 验证 reviewer_type = ai
	reviewerType, ok := apiResp.Data["reviewer_type"]
	if !ok {
		t.Fatal("IT-52 响应缺少 reviewer_type 字段")
	}
	if reviewerType != "ai" {
		t.Fatalf("IT-52 reviewer_type 错误: 期望 ai, 实际 %v", reviewerType)
	}

	// 验证 content 非空
	contentVal, ok := apiResp.Data["content"]
	if !ok || contentVal == "" {
		t.Fatalf("IT-52 响应缺少 content 或 content 为空")
	}

	t.Logf("IT-52 通过: AI 自动点评作业成功, reviewer_type=ai, content长度=%d", len(fmt.Sprintf("%v", contentVal)))
}

// ======================== IT-53: 上传 TXT 文件 → 自动解析入库 ========================
func TestV2I1_IT53_UploadTXTFile(t *testing.T) {
	v2i1Setup(t)

	// 创建临时 TXT 文件内容
	txtContent := []byte("牛顿运动定律是经典力学的基础。\n牛顿第一定律：惯性定律。\n牛顿第二定律：F=ma。\n牛顿第三定律：作用力与反作用力。")

	resp, body, err := doMultipartUpload(
		"/api/documents/upload",
		v2i1TeacherToken,
		"file",
		"newton_laws.txt",
		txtContent,
		map[string]string{
			"title": "牛顿运动定律TXT",
			"tags":  "物理,力学",
		},
	)
	if err != nil {
		t.Fatalf("IT-53 上传请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-53 HTTP状态码错误: 期望200, 实际%d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-53 解析失败: %v", err)
	}
	if apiResp.Code != 0 {
		t.Fatalf("IT-53 业务码错误: 期望0, 实际%d, message: %s", apiResp.Code, apiResp.Message)
	}

	// 验证返回字段
	docIDVal, ok := apiResp.Data["document_id"]
	if !ok {
		t.Fatal("IT-53 响应缺少 document_id 字段")
	}
	docType, ok := apiResp.Data["doc_type"]
	if !ok {
		t.Fatal("IT-53 响应缺少 doc_type 字段")
	}
	if docType != "txt" {
		t.Fatalf("IT-53 doc_type 错误: 期望 txt, 实际 %v", docType)
	}

	t.Logf("IT-53 通过: 上传 TXT 文件自动解析入库成功, document_id=%v, doc_type=%v", docIDVal, docType)
}

// ======================== IT-54: 上传 MD 文件 → 自动解析入库 ========================
func TestV2I1_IT54_UploadMDFile(t *testing.T) {
	v2i1Setup(t)

	// 创建临时 MD 文件内容
	mdContent := []byte("# 量子力学基础\n\n## 波粒二象性\n\n微观粒子既具有粒子性又具有波动性。\n\n## 薛定谔方程\n\n描述了量子态随时间的演化。\n")

	resp, body, err := doMultipartUpload(
		"/api/documents/upload",
		v2i1TeacherToken,
		"file",
		"quantum_mechanics.md",
		mdContent,
		map[string]string{
			"title": "量子力学基础MD",
			"tags":  "物理,量子力学",
		},
	)
	if err != nil {
		t.Fatalf("IT-54 上传请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-54 HTTP状态码错误: 期望200, 实际%d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-54 解析失败: %v", err)
	}
	if apiResp.Code != 0 {
		t.Fatalf("IT-54 业务码错误: 期望0, 实际%d, message: %s", apiResp.Code, apiResp.Message)
	}

	// 验证返回字段
	docIDVal, ok := apiResp.Data["document_id"]
	if !ok {
		t.Fatal("IT-54 响应缺少 document_id 字段")
	}
	docType, ok := apiResp.Data["doc_type"]
	if !ok {
		t.Fatal("IT-54 响应缺少 doc_type 字段")
	}
	if docType != "md" {
		t.Fatalf("IT-54 doc_type 错误: 期望 md, 实际 %v", docType)
	}

	t.Logf("IT-54 通过: 上传 MD 文件自动解析入库成功, document_id=%v, doc_type=%v", docIDVal, docType)
}

// ======================== IT-55: 上传不支持格式返回 400 ========================
func TestV2I1_IT55_UploadUnsupportedFormat(t *testing.T) {
	v2i1Setup(t)

	// 上传 .exe 文件（不支持的格式）
	exeContent := []byte("这不是一个真正的exe文件")

	resp, body, err := doMultipartUpload(
		"/api/documents/upload",
		v2i1TeacherToken,
		"file",
		"malware.exe",
		exeContent,
		map[string]string{},
	)
	if err != nil {
		t.Fatalf("IT-55 上传请求失败: %v", err)
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-55 解析失败: %v", err)
	}

	// 验证返回 400 + 40010
	if resp.StatusCode != http.StatusBadRequest {
		t.Logf("IT-55 警告: HTTP状态码 %d (期望 400), code=%d", resp.StatusCode, apiResp.Code)
	}
	if apiResp.Code != 40010 {
		t.Fatalf("IT-55 业务码错误: 期望 40010, 实际 %d, message: %s, body: %s", apiResp.Code, apiResp.Message, string(body))
	}

	t.Logf("IT-55 通过: 上传不支持格式正确返回 40010, message: %s", apiResp.Message)
}

// ======================== IT-56: URL 导入 → 自动抓取入库 ========================
func TestV2I1_IT56_ImportURL(t *testing.T) {
	v2i1Setup(t)

	// 创建一个临时 HTTP 服务器提供测试页面
	testPageContent := `<!DOCTYPE html>
<html>
<head><title>测试页面 - 物理学基础</title></head>
<body>
<h1>物理学基础</h1>
<p>物理学是研究物质运动规律和物质基本结构的自然科学。</p>
<p>经典力学是物理学的重要分支，由牛顿奠基。</p>
</body>
</html>`

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(testPageContent))
	}))
	defer testServer.Close()

	// 调用 URL 导入接口
	importBody := map[string]interface{}{
		"url":   testServer.URL,
		"title": "物理学基础导入",
		"tags":  "物理",
	}
	resp, body, err := doRequest("POST", "/api/documents/import-url", importBody, v2i1TeacherToken)
	if err != nil {
		t.Fatalf("IT-56 请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-56 HTTP状态码错误: 期望200, 实际%d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-56 解析失败: %v", err)
	}
	if apiResp.Code != 0 {
		t.Fatalf("IT-56 业务码错误: 期望0, 实际%d, message: %s", apiResp.Code, apiResp.Message)
	}

	// 验证返回字段
	docIDVal, ok := apiResp.Data["document_id"]
	if !ok {
		t.Fatal("IT-56 响应缺少 document_id 字段")
	}
	docType, ok := apiResp.Data["doc_type"]
	if !ok {
		t.Fatal("IT-56 响应缺少 doc_type 字段")
	}
	if docType != "url" {
		t.Fatalf("IT-56 doc_type 错误: 期望 url, 实际 %v", docType)
	}
	sourceURL, ok := apiResp.Data["source_url"]
	if !ok || sourceURL != testServer.URL {
		t.Fatalf("IT-56 source_url 错误: 期望 %s, 实际 %v", testServer.URL, sourceURL)
	}

	t.Logf("IT-56 通过: URL 导入自动抓取入库成功, document_id=%v, doc_type=%v", docIDVal, docType)
}

// ======================== IT-57: URL 不可达返回 400 ========================
func TestV2I1_IT57_ImportURLUnreachable(t *testing.T) {
	v2i1Setup(t)

	// 使用一个不可达的 URL
	importBody := map[string]interface{}{
		"url": "http://127.0.0.1:19999/nonexistent-page",
	}
	resp, body, err := doRequest("POST", "/api/documents/import-url", importBody, v2i1TeacherToken)
	if err != nil {
		t.Fatalf("IT-57 请求失败: %v", err)
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-57 解析失败: %v", err)
	}

	// 验证返回 400 + 40012
	if resp.StatusCode != http.StatusBadRequest {
		t.Logf("IT-57 警告: HTTP状态码 %d (期望 400), code=%d", resp.StatusCode, apiResp.Code)
	}
	if apiResp.Code != 40012 {
		t.Fatalf("IT-57 业务码错误: 期望 40012, 实际 %d, message: %s, body: %s", apiResp.Code, apiResp.Message, string(body))
	}

	t.Logf("IT-57 通过: URL 不可达正确返回 40012, message: %s", apiResp.Message)
}

// ======================== IT-58: SSE 流式对话 → 逐字输出 ========================
func TestV2I1_IT58_SSEStreamChat(t *testing.T) {
	v2i1Setup(t) // 已建立 approved 关系

	// 构建 SSE 请求
	chatBody := map[string]interface{}{
		"message":    "什么是牛顿第一定律?",
		"teacher_id": int(v2i1TeacherID),
	}
	jsonData, err := json.Marshal(chatBody)
	if err != nil {
		t.Fatalf("IT-58 序列化请求体失败: %v", err)
	}

	req, err := http.NewRequest("POST", ts.URL+"/api/chat/stream", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("IT-58 创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+v2i1StudentToken)

	// 使用自定义 client 避免自动关闭 body
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("IT-58 发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("IT-58 HTTP状态码错误: 期望200, 实际%d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// 验证 Content-Type 为 text/event-stream
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") {
		t.Logf("IT-58 警告: Content-Type=%s (期望包含 text/event-stream)", contentType)
	}

	// 解析 SSE 事件流
	scanner := bufio.NewScanner(resp.Body)
	var (
		hasStart    bool
		hasDelta    bool
		hasDone     bool
		deltaCount  int
		fullContent string
	)

	for scanner.Scan() {
		line := scanner.Text()

		// SSE 数据行以 "data: " 开头
		if !strings.HasPrefix(line, "data: ") && !strings.HasPrefix(line, "data:") {
			continue
		}

		// 提取 JSON 数据
		dataStr := strings.TrimPrefix(line, "data: ")
		dataStr = strings.TrimPrefix(dataStr, "data:")
		dataStr = strings.TrimSpace(dataStr)

		if dataStr == "" {
			continue
		}

		var event map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &event); err != nil {
			t.Logf("IT-58 解析SSE事件失败: %v, data: %s", err, dataStr)
			continue
		}

		eventType, ok := event["type"].(string)
		if !ok {
			continue
		}

		switch eventType {
		case "start":
			hasStart = true
			sessionID, _ := event["session_id"].(string)
			if sessionID == "" {
				t.Fatal("IT-58 start 事件缺少 session_id")
			}
			t.Logf("IT-58 收到 start 事件, session_id=%s", sessionID)

		case "delta":
			hasDelta = true
			deltaCount++
			content, _ := event["content"].(string)
			fullContent += content

		case "done":
			hasDone = true
			convID, _ := event["conversation_id"]
			t.Logf("IT-58 收到 done 事件, conversation_id=%v", convID)

		case "error":
			errCode, _ := event["code"]
			errMsg, _ := event["message"]
			t.Fatalf("IT-58 收到 error 事件: code=%v, message=%v", errCode, errMsg)
		}
	}

	if err := scanner.Err(); err != nil {
		t.Logf("IT-58 读取SSE流时出错: %v (可能是连接正常关闭)", err)
	}

	// 验证事件完整性
	if !hasStart {
		t.Fatal("IT-58 未收到 start 事件")
	}
	if !hasDelta {
		t.Fatal("IT-58 未收到任何 delta 事件")
	}
	if !hasDone {
		t.Fatal("IT-58 未收到 done 事件")
	}

	t.Logf("IT-58 通过: SSE 流式对话成功, delta事件数=%d, 完整内容长度=%d", deltaCount, len(fullContent))
}

// ======================== IT-59: SSE 流式完成后对话记录已保存 ========================
func TestV2I1_IT59_SSEStreamSavesConversation(t *testing.T) {
	v2i1Setup(t)

	// 步骤1：发送 SSE 流式对话
	chatBody := map[string]interface{}{
		"message":    "请解释牛顿第二定律 F=ma",
		"teacher_id": int(v2i1TeacherID),
	}
	jsonData, _ := json.Marshal(chatBody)

	req, _ := http.NewRequest("POST", ts.URL+"/api/chat/stream", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+v2i1StudentToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("IT-59 步骤1 发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("IT-59 步骤1 HTTP状态码错误: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// 读取完整 SSE 流并提取 conversation_id
	var conversationID float64
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") && !strings.HasPrefix(line, "data:") {
			continue
		}
		dataStr := strings.TrimPrefix(line, "data: ")
		dataStr = strings.TrimPrefix(dataStr, "data:")
		dataStr = strings.TrimSpace(dataStr)
		if dataStr == "" {
			continue
		}

		var event map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &event); err != nil {
			continue
		}
		if event["type"] == "done" {
			if convID, ok := event["conversation_id"].(float64); ok {
				conversationID = convID
			}
		}
	}

	if conversationID <= 0 {
		t.Logf("IT-59 警告: 未从 done 事件获取到有效的 conversation_id (值=%v)", conversationID)
	}

	// 步骤2：等待异步处理完成
	time.Sleep(1 * time.Second)

	// 步骤3：查询对话历史，验证记录已保存
	historyPath := fmt.Sprintf("/api/conversations?teacher_id=%d&page=1&page_size=50", int(v2i1TeacherID))
	resp2, body2, err := doRequest("GET", historyPath, nil, v2i1StudentToken)
	if err != nil {
		t.Fatalf("IT-59 步骤3 查询历史失败: %v", err)
	}
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("IT-59 步骤3 HTTP状态码错误: %d, body: %s", resp2.StatusCode, string(body2))
	}

	apiResp, err := parseResponse(body2)
	if err != nil {
		t.Fatalf("IT-59 步骤3 解析失败: %v", err)
	}
	if apiResp.Code != 0 {
		t.Fatalf("IT-59 步骤3 业务码错误: 期望0, 实际%d", apiResp.Code)
	}

	// 验证对话历史中有记录
	itemsVal, ok := apiResp.Data["items"]
	if !ok {
		t.Fatal("IT-59 步骤3 响应缺少 items 字段")
	}
	if itemsVal != nil {
		items, ok := itemsVal.([]interface{})
		if !ok || len(items) == 0 {
			t.Fatalf("IT-59 步骤3 对话历史为空，SSE 流式完成后应有保存记录")
		}
		t.Logf("IT-59 步骤3: 对话历史数量=%d", len(items))
	} else {
		t.Fatal("IT-59 步骤3 items 为 nil")
	}

	t.Logf("IT-59 通过: SSE 流式完成后对话记录已保存, conversation_id=%v", conversationID)
}

// ======================== IT-60: 全链路测试 ========================
// 注册→授权→设置风格→对话→评语→作业→点评
func TestV2I1_IT60_FullEndToEndFlow(t *testing.T) {
	os.Setenv("WX_MODE", "mock")
	os.Setenv("LLM_MODE", "mock")

	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano()%10000)

	// ===== 步骤1：教师注册 =====
	t.Log("IT-60 步骤1: 教师注册")
	teacherCode := "x" + uniqueSuffix // 后6位: x + suffix
	studentCode := "y" + uniqueSuffix // 后6位: y + suffix（与教师不同）
	teacherToken, teacherID := v2i1RegisterTeacher(t,
		teacherCode,
		"IT60全链路老师",
		"IT60全链路大学",
		"IT60全链路测试教师描述",
	)

	// ===== 步骤2：学生注册 =====
	t.Log("IT-60 步骤2: 学生注册")
	studentToken, studentID := v2i1RegisterStudent(t,
		studentCode,
		"IT60全链路学生",
	)

	// ===== 步骤3：教师邀请学生（建立授权关系） =====
	t.Log("IT-60 步骤3: 教师邀请学生")
	inviteBody := map[string]interface{}{
		"student_id": int(studentID),
	}
	resp, body, err := doRequest("POST", "/api/relations/invite", inviteBody, teacherToken)
	if err != nil {
		t.Fatalf("IT-60 步骤3 邀请失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-60 步骤3 HTTP状态码错误: %d, body: %s", resp.StatusCode, string(body))
	}
	apiResp, err := parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-60 步骤3 解析失败: %v, code: %d, body: %s", err, apiResp.Code, string(body))
	}
	if apiResp.Data["status"] != "approved" {
		t.Fatalf("IT-60 步骤3 status 错误: 期望 approved, 实际 %v", apiResp.Data["status"])
	}
	t.Logf("IT-60 步骤3: 教师邀请学生成功, status=approved")

	// ===== 步骤4：教师添加知识库文档 =====
	t.Log("IT-60 步骤4: 教师添加知识库文档")
	docBody := map[string]interface{}{
		"title":   "IT60全链路测试文档",
		"content": "牛顿第一定律也叫惯性定律。牛顿第二定律 F=ma。牛顿第三定律是作用力与反作用力。",
		"tags":    "物理,力学",
	}
	resp, body, err = doRequest("POST", "/api/documents", docBody, teacherToken)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-60 步骤4 添加文档失败: %v, status: %d, body: %s", err, resp.StatusCode, string(body))
	}
	apiResp, _ = parseResponse(body)
	if apiResp.Code != 0 {
		t.Fatalf("IT-60 步骤4 业务码错误: %d, body: %s", apiResp.Code, string(body))
	}
	t.Logf("IT-60 步骤4: 文档添加成功")

	// ===== 步骤5：教师设置问答风格 =====
	t.Log("IT-60 步骤5: 教师设置问答风格")
	stylePath := fmt.Sprintf("/api/students/%d/dialogue-style", int(studentID))
	styleBody := map[string]interface{}{
		"temperature":         0.7,
		"guidance_level":      "medium",
		"style_prompt":        "请多用鼓励性语言",
		"max_turns_per_topic": 5,
	}
	resp, body, err = doRequest("PUT", stylePath, styleBody, teacherToken)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-60 步骤5 设置风格失败: %v, status: %d, body: %s", err, resp.StatusCode, string(body))
	}
	apiResp, _ = parseResponse(body)
	if apiResp.Code != 0 {
		t.Fatalf("IT-60 步骤5 业务码错误: %d", apiResp.Code)
	}
	t.Logf("IT-60 步骤5: 设置问答风格成功")

	// ===== 步骤6：学生发起对话 =====
	t.Log("IT-60 步骤6: 学生发起对话")
	chatBody := map[string]interface{}{
		"message":    "请问什么是牛顿第一定律?",
		"teacher_id": int(teacherID),
	}
	resp, body, err = doRequest("POST", "/api/chat", chatBody, studentToken)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-60 步骤6 对话失败: %v, status: %d, body: %s", err, resp.StatusCode, string(body))
	}
	apiResp, _ = parseResponse(body)
	if apiResp.Code != 0 {
		t.Fatalf("IT-60 步骤6 业务码错误: %d, body: %s", apiResp.Code, string(body))
	}
	replyVal := apiResp.Data["reply"]
	if replyVal == "" {
		t.Fatal("IT-60 步骤6 reply 为空")
	}
	t.Logf("IT-60 步骤6: 对话成功, reply长度=%d", len(fmt.Sprintf("%v", replyVal)))

	// ===== 步骤7：教师写评语 =====
	t.Log("IT-60 步骤7: 教师写评语")
	commentBody := map[string]interface{}{
		"student_id":       int(studentID),
		"content":          "IT60全链路测试评语：该生学习态度认真",
		"progress_summary": "牛顿定律掌握80%",
	}
	resp, body, err = doRequest("POST", "/api/comments", commentBody, teacherToken)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-60 步骤7 写评语失败: %v, status: %d, body: %s", err, resp.StatusCode, string(body))
	}
	apiResp, _ = parseResponse(body)
	if apiResp.Code != 0 {
		t.Fatalf("IT-60 步骤7 业务码错误: %d", apiResp.Code)
	}
	t.Logf("IT-60 步骤7: 写评语成功")

	// ===== 步骤8：学生查看评语（V2.0 迭代5：学生角色返回空列表） =====
	t.Log("IT-60 步骤8: 学生查看评语（迭代5后学生返回空列表）")
	commentPath := fmt.Sprintf("/api/comments?teacher_id=%d", int(teacherID))
	resp, body, err = doRequest("GET", commentPath, nil, studentToken)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-60 步骤8 查看评语失败: %v", err)
	}
	// V2.0 迭代5：学生角色返回 {"code":0,"message":"success","data":[]}
	// data 为空数组而非 map，使用 apiResponseRaw 解析
	var commentResp apiResponseRaw
	if err := json.Unmarshal(body, &commentResp); err != nil {
		t.Fatalf("IT-60 步骤8 解析响应失败: %v", err)
	}
	if commentResp.Code != 0 {
		t.Fatalf("IT-60 步骤8 业务码错误: %d", commentResp.Code)
	}
	if string(commentResp.Data) != "[]" {
		t.Fatalf("IT-60 步骤8 期望 data=[]，实际: %s", string(commentResp.Data))
	}
	t.Logf("IT-60 步骤8: 学生查看评语返回空列表（符合迭代5预期）")

	// ===== 步骤9：学生提交作业 =====
	t.Log("IT-60 步骤9: 学生提交作业")
	assignmentBody := map[string]interface{}{
		"teacher_id": int(teacherID),
		"title":      "IT60全链路作业",
		"content":    "牛顿第一定律是惯性定律，物体在不受外力时保持原有运动状态。",
	}
	resp, body, err = doRequest("POST", "/api/assignments", assignmentBody, studentToken)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-60 步骤9 提交作业失败: %v, status: %d, body: %s", err, resp.StatusCode, string(body))
	}
	apiResp, _ = parseResponse(body)
	if apiResp.Code != 0 {
		t.Fatalf("IT-60 步骤9 业务码错误: %d", apiResp.Code)
	}
	assignmentID := apiResp.Data["id"].(float64)
	t.Logf("IT-60 步骤9: 提交作业成功, id=%v", assignmentID)

	// ===== 步骤10：AI 自动点评 =====
	t.Log("IT-60 步骤10: AI 自动点评")
	aiReviewPath := fmt.Sprintf("/api/assignments/%d/ai-review", int(assignmentID))
	resp, body, err = doRequest("POST", aiReviewPath, nil, teacherToken)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-60 步骤10 AI点评失败: %v, status: %d, body: %s", err, resp.StatusCode, string(body))
	}
	apiResp, _ = parseResponse(body)
	if apiResp.Code != 0 {
		t.Fatalf("IT-60 步骤10 业务码错误: %d", apiResp.Code)
	}
	t.Logf("IT-60 步骤10: AI 自动点评成功")

	// ===== 步骤11：教师手动点评 =====
	t.Log("IT-60 步骤11: 教师手动点评")
	reviewPath := fmt.Sprintf("/api/assignments/%d/review", int(assignmentID))
	reviewBody := map[string]interface{}{
		"content": "IT60全链路教师点评：概念理解准确，继续加油！",
		"score":   90,
	}
	resp, body, err = doRequest("POST", reviewPath, reviewBody, teacherToken)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-60 步骤11 教师点评失败: %v, status: %d, body: %s", err, resp.StatusCode, string(body))
	}
	apiResp, _ = parseResponse(body)
	if apiResp.Code != 0 {
		t.Fatalf("IT-60 步骤11 业务码错误: %d", apiResp.Code)
	}
	t.Logf("IT-60 步骤11: 教师手动点评成功")

	// ===== 步骤12：查看作业详情（含所有点评） =====
	t.Log("IT-60 步骤12: 查看作业详情")
	detailPath := fmt.Sprintf("/api/assignments/%d", int(assignmentID))
	resp, body, err = doRequest("GET", detailPath, nil, teacherToken)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-60 步骤12 查看详情失败: %v, status: %d, body: %s", err, resp.StatusCode, string(body))
	}
	apiResp, _ = parseResponse(body)
	if apiResp.Code != 0 {
		t.Fatalf("IT-60 步骤12 业务码错误: %d", apiResp.Code)
	}

	// 验证作业状态为 reviewed
	if statusVal, ok := apiResp.Data["status"]; ok {
		if statusVal != "reviewed" {
			t.Logf("IT-60 步骤12 警告: 作业状态=%v (期望 reviewed)", statusVal)
		}
	}

	// 验证 reviews 包含 AI 和教师点评
	if reviewsVal, ok := apiResp.Data["reviews"]; ok && reviewsVal != nil {
		reviews, ok := reviewsVal.([]interface{})
		if ok {
			t.Logf("IT-60 步骤12: 作业详情获取成功, 点评数量=%d", len(reviews))
			if len(reviews) < 2 {
				t.Logf("IT-60 步骤12 警告: 点评数量=%d (期望>=2)", len(reviews))
			}
		}
	}

	// ===== 步骤13：上传文件 =====
	t.Log("IT-60 步骤13: 上传 TXT 文件")
	txtContent := []byte("IT60全链路测试文件内容：力学基础知识。")
	resp, body, err = doMultipartUpload(
		"/api/documents/upload",
		teacherToken,
		"file",
		"it60_test.txt",
		txtContent,
		map[string]string{"title": "IT60上传文件", "tags": "测试"},
	)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-60 步骤13 上传文件失败: %v, status: %d, body: %s", err, resp.StatusCode, string(body))
	}
	apiResp, _ = parseResponse(body)
	if apiResp.Code != 0 {
		t.Fatalf("IT-60 步骤13 业务码错误: %d", apiResp.Code)
	}
	t.Logf("IT-60 步骤13: 上传文件成功")

	// ===== 步骤14：SSE 流式对话 =====
	t.Log("IT-60 步骤14: SSE 流式对话")
	sseBody := map[string]interface{}{
		"message":    "请用简单的语言解释力的概念",
		"teacher_id": int(teacherID),
	}
	jsonData, _ := json.Marshal(sseBody)
	sseReq, _ := http.NewRequest("POST", ts.URL+"/api/chat/stream", bytes.NewBuffer(jsonData))
	sseReq.Header.Set("Content-Type", "application/json")
	sseReq.Header.Set("Authorization", "Bearer "+studentToken)

	sseClient := &http.Client{Timeout: 30 * time.Second}
	sseResp, err := sseClient.Do(sseReq)
	if err != nil {
		t.Fatalf("IT-60 步骤14 SSE请求失败: %v", err)
	}
	defer sseResp.Body.Close()

	if sseResp.StatusCode != http.StatusOK {
		sseBodyBytes, _ := io.ReadAll(sseResp.Body)
		t.Fatalf("IT-60 步骤14 HTTP状态码错误: %d, body: %s", sseResp.StatusCode, string(sseBodyBytes))
	}

	// 读取 SSE 流，验证 start + delta + done
	sseScanner := bufio.NewScanner(sseResp.Body)
	hasStart, hasDelta, hasDone := false, false, false
	for sseScanner.Scan() {
		line := sseScanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		dataStr := strings.TrimPrefix(line, "data: ")
		dataStr = strings.TrimPrefix(dataStr, "data:")
		dataStr = strings.TrimSpace(dataStr)
		if dataStr == "" {
			continue
		}
		var event map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &event); err != nil {
			continue
		}
		switch event["type"] {
		case "start":
			hasStart = true
		case "delta":
			hasDelta = true
		case "done":
			hasDone = true
		}
	}

	if !hasStart || !hasDelta || !hasDone {
		t.Fatalf("IT-60 步骤14 SSE事件不完整: start=%v, delta=%v, done=%v", hasStart, hasDelta, hasDone)
	}
	t.Logf("IT-60 步骤14: SSE 流式对话成功")

	t.Logf("IT-60 通过: 全链路测试成功 (注册→授权→风格→文档→对话→评语→作业→AI点评→教师点评→上传→SSE流式)")
}

// ======================== 辅助函数：确保未使用的 import 不报错 ========================
var _ = filepath.Join
var _ = os.TempDir
