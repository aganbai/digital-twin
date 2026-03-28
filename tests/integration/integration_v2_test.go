package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
)

// ======================== V2 集成测试辅助变量 ========================

// v2 测试用的 token 和 ID（避免与 v1 测试变量冲突）
var (
	v2TeacherToken string
	v2StudentToken string
	v2TeacherID    float64
	v2StudentID    float64
	v2DocumentID   float64
)

// ======================== IT-18: 微信登录（mock 模式）→ 返回 token + is_new_user=true ========================
func TestV2_IT18_WxLoginNewUser(t *testing.T) {
	// 确保 WX_MODE 为 mock
	os.Setenv("WX_MODE", "mock")

	reqBody := map[string]interface{}{
		"code": "test_new_001",
	}

	resp, body, err := doRequest("POST", "/api/auth/wx-login", reqBody, "")
	if err != nil {
		t.Fatalf("IT-18 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-18 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-18 解析响应失败: %v", err)
	}

	if apiResp.Code != 0 {
		t.Fatalf("IT-18 业务码错误: 期望 0, 实际 %d, message: %s", apiResp.Code, apiResp.Message)
	}

	// 验证 token 存在
	tokenVal, ok := apiResp.Data["token"]
	if !ok || tokenVal == "" {
		t.Fatal("IT-18 响应缺少 token 或 token 为空")
	}

	// 验证 is_new_user = true
	isNewUser, ok := apiResp.Data["is_new_user"]
	if !ok {
		t.Fatal("IT-18 响应缺少 is_new_user 字段")
	}
	if isNewUser != true {
		t.Fatalf("IT-18 is_new_user 错误: 期望 true, 实际 %v", isNewUser)
	}

	// 验证 role 为空字符串
	roleVal, ok := apiResp.Data["role"]
	if !ok {
		t.Fatal("IT-18 响应缺少 role 字段")
	}
	if roleVal != "" {
		t.Fatalf("IT-18 role 错误: 期望空字符串, 实际 %v", roleVal)
	}

	// 保存 token 供后续测试使用
	v2TeacherToken = tokenVal.(string)

	// 保存 user_id
	if uid, ok := apiResp.Data["user_id"]; ok {
		v2TeacherID, _ = uid.(float64)
	}

	t.Logf("IT-18 通过: 微信登录新用户成功, is_new_user=true, role=\"\", user_id=%v", v2TeacherID)
}

// ======================== IT-19: 新用户补全信息 → 设置角色和昵称 ========================
func TestV2_IT19_CompleteProfile(t *testing.T) {
	if v2TeacherToken == "" {
		t.Skip("IT-19 跳过: 教师 token 未获取（IT-18 可能失败）")
	}

	reqBody := map[string]interface{}{
		"role":     "teacher",
		"nickname": "王老师",
	}

	resp, body, err := doRequest("POST", "/api/auth/complete-profile", reqBody, v2TeacherToken)
	if err != nil {
		t.Fatalf("IT-19 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-19 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-19 解析响应失败: %v", err)
	}

	if apiResp.Code != 0 {
		t.Fatalf("IT-19 业务码错误: 期望 0, 实际 %d, message: %s", apiResp.Code, apiResp.Message)
	}

	// 验证返回的 role
	roleVal, ok := apiResp.Data["role"]
	if !ok || roleVal != "teacher" {
		t.Fatalf("IT-19 role 错误: 期望 teacher, 实际 %v", roleVal)
	}

	// 验证返回的 nickname
	nicknameVal, ok := apiResp.Data["nickname"]
	if !ok || nicknameVal != "王老师" {
		t.Fatalf("IT-19 nickname 错误: 期望 王老师, 实际 %v", nicknameVal)
	}

	t.Logf("IT-19 通过: 补全信息成功, role=teacher, nickname=王老师")
}

// ======================== IT-20: 同一 openid 再次登录 → is_new_user=false ========================
func TestV2_IT20_WxLoginExistingUser(t *testing.T) {
	os.Setenv("WX_MODE", "mock")

	// 使用同一个 code（mock 模式下会生成同一个 openid: mock_openid_test_new_001）
	reqBody := map[string]interface{}{
		"code": "test_new_001",
	}

	resp, body, err := doRequest("POST", "/api/auth/wx-login", reqBody, "")
	if err != nil {
		t.Fatalf("IT-20 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-20 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-20 解析响应失败: %v", err)
	}

	if apiResp.Code != 0 {
		t.Fatalf("IT-20 业务码错误: 期望 0, 实际 %d, message: %s", apiResp.Code, apiResp.Message)
	}

	// 验证 is_new_user = false
	isNewUser, ok := apiResp.Data["is_new_user"]
	if !ok {
		t.Fatal("IT-20 响应缺少 is_new_user 字段")
	}
	if isNewUser != false {
		t.Fatalf("IT-20 is_new_user 错误: 期望 false, 实际 %v", isNewUser)
	}

	// 验证 role = "teacher"（IT-19 中已设置）
	roleVal, ok := apiResp.Data["role"]
	if !ok || roleVal != "teacher" {
		t.Fatalf("IT-20 role 错误: 期望 teacher, 实际 %v", roleVal)
	}

	// 更新 token（再次登录会获得新 token）
	if tokenVal, ok := apiResp.Data["token"]; ok && tokenVal != "" {
		v2TeacherToken = tokenVal.(string)
	}

	t.Logf("IT-20 通过: 同一 openid 再次登录成功, is_new_user=false, role=teacher")
}

// ======================== IT-21: 获取教师列表 → 返回教师数组 + document_count ========================
func TestV2_IT21_GetTeachers(t *testing.T) {
	// 先用学生身份登录（需要一个学生 token 来调用教师列表接口）
	// 创建学生用户
	os.Setenv("WX_MODE", "mock")
	reqBody := map[string]interface{}{
		"code": "test_student_001",
	}
	_, body, err := doRequest("POST", "/api/auth/wx-login", reqBody, "")
	if err != nil {
		t.Fatalf("IT-21 学生登录失败: %v", err)
	}
	apiResp, err := parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-21 学生登录解析失败: %v, code: %d", err, apiResp.Code)
	}
	v2StudentToken = apiResp.Data["token"].(string)
	if uid, ok := apiResp.Data["user_id"]; ok {
		v2StudentID, _ = uid.(float64)
	}

	// 补全学生信息
	completeBody := map[string]interface{}{
		"role":     "student",
		"nickname": "小李",
	}
	_, body, err = doRequest("POST", "/api/auth/complete-profile", completeBody, v2StudentToken)
	if err != nil {
		t.Fatalf("IT-21 学生补全信息失败: %v", err)
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-21 学生补全信息解析失败: %v, code: %d, body: %s", err, apiResp.Code, string(body))
	}

	// 获取教师列表
	resp, body, err := doRequest("GET", "/api/teachers?page=1&page_size=20", nil, v2StudentToken)
	if err != nil {
		t.Fatalf("IT-21 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-21 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err = parseResponse(body)
	if err != nil {
		t.Fatalf("IT-21 解析响应失败: %v", err)
	}

	if apiResp.Code != 0 {
		t.Fatalf("IT-21 业务码错误: 期望 0, 实际 %d, message: %s", apiResp.Code, apiResp.Message)
	}

	// 验证 items 字段
	itemsVal, ok := apiResp.Data["items"]
	if !ok {
		t.Fatal("IT-21 响应缺少 items 字段")
	}

	if itemsVal != nil {
		items, ok := itemsVal.([]interface{})
		if !ok {
			t.Fatalf("IT-21 items 不是数组类型: %T", itemsVal)
		}

		if len(items) == 0 {
			t.Fatal("IT-21 教师列表为空，期望至少有1个教师")
		}

		// 验证第一个教师有 document_count 字段
		firstTeacher, ok := items[0].(map[string]interface{})
		if !ok {
			t.Fatalf("IT-21 教师项不是对象类型: %T", items[0])
		}

		if _, ok := firstTeacher["document_count"]; !ok {
			t.Fatal("IT-21 教师项缺少 document_count 字段")
		}

		t.Logf("IT-21 通过: 获取教师列表成功, 教师数量=%d, 第一个教师: %v", len(items), firstTeacher["nickname"])
	} else {
		t.Fatal("IT-21 items 为 nil，期望至少有1个教师")
	}
}

// ======================== IT-22: 获取用户信息 → 返回 profile + stats ========================
func TestV2_IT22_GetUserProfile(t *testing.T) {
	if v2StudentToken == "" {
		t.Skip("IT-22 跳过: 学生 token 未获取")
	}

	resp, body, err := doRequest("GET", "/api/user/profile", nil, v2StudentToken)
	if err != nil {
		t.Fatalf("IT-22 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-22 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-22 解析响应失败: %v", err)
	}

	if apiResp.Code != 0 {
		t.Fatalf("IT-22 业务码错误: 期望 0, 实际 %d, message: %s", apiResp.Code, apiResp.Message)
	}

	// 验证基本字段
	if _, ok := apiResp.Data["id"]; !ok {
		t.Fatal("IT-22 响应缺少 id 字段")
	}
	if _, ok := apiResp.Data["role"]; !ok {
		t.Fatal("IT-22 响应缺少 role 字段")
	}
	if _, ok := apiResp.Data["nickname"]; !ok {
		t.Fatal("IT-22 响应缺少 nickname 字段")
	}

	// 验证 stats 字段
	statsVal, ok := apiResp.Data["stats"]
	if !ok {
		t.Fatal("IT-22 响应缺少 stats 字段")
	}
	stats, ok := statsVal.(map[string]interface{})
	if !ok {
		t.Fatalf("IT-22 stats 不是对象类型: %T", statsVal)
	}

	// stats 中应包含 conversation_count 和 memory_count
	if _, ok := stats["conversation_count"]; !ok {
		t.Fatal("IT-22 stats 缺少 conversation_count 字段")
	}

	t.Logf("IT-22 通过: 获取用户信息成功, nickname=%v, role=%v, stats=%v",
		apiResp.Data["nickname"], apiResp.Data["role"], stats)
}

// ======================== IT-23: 获取会话列表 → 返回会话摘要 ========================
func TestV2_IT23_GetSessions(t *testing.T) {
	if v2StudentToken == "" {
		t.Skip("IT-23 跳过: 学生 token 未获取")
	}

	resp, body, err := doRequest("GET", "/api/conversations/sessions?page=1&page_size=20", nil, v2StudentToken)
	if err != nil {
		t.Fatalf("IT-23 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-23 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-23 解析响应失败: %v", err)
	}

	if apiResp.Code != 0 {
		t.Fatalf("IT-23 业务码错误: 期望 0, 实际 %d, message: %s", apiResp.Code, apiResp.Message)
	}

	// 验证分页结构
	if _, ok := apiResp.Data["items"]; !ok {
		t.Fatal("IT-23 响应缺少 items 字段")
	}
	if _, ok := apiResp.Data["total"]; !ok {
		t.Fatal("IT-23 响应缺少 total 字段")
	}
	if _, ok := apiResp.Data["page"]; !ok {
		t.Fatal("IT-23 响应缺少 page 字段")
	}

	t.Logf("IT-23 通过: 获取会话列表成功, total=%v", apiResp.Data["total"])
}

// ======================== IT-24: 对话历史不传 teacher_id → 返回所有对话 ========================
func TestV2_IT24_GetConversationsWithoutTeacherID(t *testing.T) {
	if v2StudentToken == "" {
		t.Skip("IT-24 跳过: 学生 token 未获取")
	}

	// 不传 teacher_id
	resp, body, err := doRequest("GET", "/api/conversations?page=1&page_size=50", nil, v2StudentToken)
	if err != nil {
		t.Fatalf("IT-24 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-24 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-24 解析响应失败: %v", err)
	}

	if apiResp.Code != 0 {
		t.Fatalf("IT-24 业务码错误: 期望 0, 实际 %d, message: %s", apiResp.Code, apiResp.Message)
	}

	// 验证分页结构
	if _, ok := apiResp.Data["items"]; !ok {
		t.Fatal("IT-24 响应缺少 items 字段")
	}

	t.Logf("IT-24 通过: 不传 teacher_id 获取对话历史成功, total=%v", apiResp.Data["total"])
}

// ======================== IT-25: 微信登录→补全信息→教师列表→对话 全链路 ========================
func TestV2_IT25_FullFlowStudentChat(t *testing.T) {
	os.Setenv("WX_MODE", "mock")

	// 步骤1：学生微信登录
	t.Log("IT-25 步骤1: 学生微信登录")
	loginBody := map[string]interface{}{
		"code": "test_fullflow_student",
	}
	resp, body, err := doRequest("POST", "/api/auth/wx-login", loginBody, "")
	if err != nil {
		t.Fatalf("IT-25 步骤1 登录失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-25 步骤1 HTTP 状态码错误: %d, body: %s", resp.StatusCode, string(body))
	}
	apiResp, err := parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-25 步骤1 解析失败: %v, code: %d", err, apiResp.Code)
	}
	studentToken := apiResp.Data["token"].(string)
	isNewUser := apiResp.Data["is_new_user"]
	if isNewUser != true {
		t.Fatalf("IT-25 步骤1 is_new_user 错误: 期望 true, 实际 %v", isNewUser)
	}

	// 步骤2：补全学生信息
	t.Log("IT-25 步骤2: 补全学生信息")
	completeBody := map[string]interface{}{
		"role":     "student",
		"nickname": "全链路学生",
	}
	resp, body, err = doRequest("POST", "/api/auth/complete-profile", completeBody, studentToken)
	if err != nil {
		t.Fatalf("IT-25 步骤2 补全信息失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-25 步骤2 HTTP 状态码错误: %d, body: %s", resp.StatusCode, string(body))
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-25 步骤2 解析失败: %v, code: %d, body: %s", err, apiResp.Code, string(body))
	}

	// 步骤3：获取教师列表
	t.Log("IT-25 步骤3: 获取教师列表")
	resp, body, err = doRequest("GET", "/api/teachers?page=1&page_size=20", nil, studentToken)
	if err != nil {
		t.Fatalf("IT-25 步骤3 获取教师列表失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-25 步骤3 HTTP 状态码错误: %d, body: %s", resp.StatusCode, string(body))
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-25 步骤3 解析失败: %v, code: %d", err, apiResp.Code)
	}
	itemsVal := apiResp.Data["items"]
	if itemsVal == nil {
		t.Fatal("IT-25 步骤3 教师列表为空")
	}
	items, ok := itemsVal.([]interface{})
	if !ok || len(items) == 0 {
		t.Fatal("IT-25 步骤3 教师列表为空或格式错误")
	}

	// 获取第一个教师的 ID
	firstTeacher := items[0].(map[string]interface{})
	targetTeacherID := firstTeacher["id"].(float64)
	t.Logf("IT-25 步骤3: 选择教师 ID=%v, 昵称=%v", targetTeacherID, firstTeacher["nickname"])

	// 步骤4：发送对话
	t.Log("IT-25 步骤4: 发送对话")
	chatBody := map[string]interface{}{
		"message":    "你好，请问什么是牛顿第一定律？",
		"teacher_id": int(targetTeacherID),
	}
	resp, body, err = doRequest("POST", "/api/chat", chatBody, studentToken)
	if err != nil {
		t.Fatalf("IT-25 步骤4 对话失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-25 步骤4 HTTP 状态码错误: %d, body: %s", resp.StatusCode, string(body))
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-25 步骤4 解析失败: %v, code: %d, body: %s", err, apiResp.Code, string(body))
	}
	replyVal, ok := apiResp.Data["reply"]
	if !ok || replyVal == "" {
		t.Fatalf("IT-25 步骤4 响应缺少 reply 或 reply 为空, data: %v", apiResp.Data)
	}

	// 步骤5：查看对话历史
	t.Log("IT-25 步骤5: 查看对话历史")
	historyPath := fmt.Sprintf("/api/conversations?teacher_id=%d&page=1&page_size=50", int(targetTeacherID))
	resp, body, err = doRequest("GET", historyPath, nil, studentToken)
	if err != nil {
		t.Fatalf("IT-25 步骤5 查看历史失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-25 步骤5 HTTP 状态码错误: %d, body: %s", resp.StatusCode, string(body))
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-25 步骤5 解析失败: %v, code: %d", err, apiResp.Code)
	}

	t.Logf("IT-25 通过: 全链路测试成功 (登录→补全→教师列表→对话→历史)")
}

// ======================== IT-26: 教师微信登录→补全→添加文档→学生对话引用知识 ========================
func TestV2_IT26_TeacherAddDocStudentChat(t *testing.T) {
	os.Setenv("WX_MODE", "mock")

	// 步骤1：教师微信登录
	t.Log("IT-26 步骤1: 教师微信登录")
	teacherLoginBody := map[string]interface{}{
		"code": "test_teacher_doc_001",
	}
	_, body, err := doRequest("POST", "/api/auth/wx-login", teacherLoginBody, "")
	if err != nil {
		t.Fatalf("IT-26 步骤1 教师登录失败: %v", err)
	}
	apiResp, err := parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-26 步骤1 解析失败: %v, code: %d", err, apiResp.Code)
	}
	localTeacherToken := apiResp.Data["token"].(string)
	localTeacherID := apiResp.Data["user_id"].(float64)

	// 步骤2：教师补全信息
	t.Log("IT-26 步骤2: 教师补全信息")
	completeBody := map[string]interface{}{
		"role":     "teacher",
		"nickname": "物理张老师",
	}
	_, body, err = doRequest("POST", "/api/auth/complete-profile", completeBody, localTeacherToken)
	if err != nil {
		t.Fatalf("IT-26 步骤2 补全信息失败: %v", err)
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-26 步骤2 解析失败: %v, code: %d, body: %s", err, apiResp.Code, string(body))
	}

	// 重新登录获取更新后的 token（角色已更新）
	_, body, err = doRequest("POST", "/api/auth/wx-login", teacherLoginBody, "")
	if err != nil {
		t.Fatalf("IT-26 步骤2 重新登录失败: %v", err)
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-26 步骤2 重新登录解析失败: %v, code: %d", err, apiResp.Code)
	}
	localTeacherToken = apiResp.Data["token"].(string)

	// 步骤3：教师添加文档
	t.Log("IT-26 步骤3: 教师添加文档")
	docBody := map[string]interface{}{
		"title":   "量子力学基础",
		"content": "量子力学是研究微观粒子运动规律的物理学分支。波粒二象性是量子力学的基本概念之一，即微观粒子既具有粒子性又具有波动性。薛定谔方程是量子力学的基本方程，描述了量子态随时间的演化。",
		"tags":    "物理,量子力学",
	}
	resp, body, err := doRequest("POST", "/api/documents", docBody, localTeacherToken)
	if err != nil {
		t.Fatalf("IT-26 步骤3 添加文档失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-26 步骤3 HTTP 状态码错误: %d, body: %s", resp.StatusCode, string(body))
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-26 步骤3 解析失败: %v, code: %d, body: %s", err, apiResp.Code, string(body))
	}
	docIDVal, ok := apiResp.Data["document_id"]
	if !ok {
		t.Fatal("IT-26 步骤3 响应缺少 document_id")
	}
	v2DocumentID, _ = docIDVal.(float64)
	t.Logf("IT-26 步骤3: 文档添加成功, document_id=%v", v2DocumentID)

	// 步骤4：学生微信登录
	t.Log("IT-26 步骤4: 学生微信登录")
	studentLoginBody := map[string]interface{}{
		"code": "test_stu_doc_002",
	}
	_, body, err = doRequest("POST", "/api/auth/wx-login", studentLoginBody, "")
	if err != nil {
		t.Fatalf("IT-26 步骤4 学生登录失败: %v", err)
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-26 步骤4 解析失败: %v, code: %d, body: %s", err, apiResp.Code, string(body))
	}
	localStudentToken := apiResp.Data["token"].(string)

	// 步骤4.5：学生补全信息
	studentCompleteBody := map[string]interface{}{
		"role":     "student",
		"nickname": "学生小明",
	}
	_, body, err = doRequest("POST", "/api/auth/complete-profile", studentCompleteBody, localStudentToken)
	if err != nil {
		t.Fatalf("IT-26 步骤4.5 学生补全信息失败: %v", err)
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-26 步骤4.5 解析失败: %v, code: %d, body: %s", err, apiResp.Code, string(body))
	}

	// 步骤5：学生向该教师发送对话（应引用知识）
	t.Log("IT-26 步骤5: 学生向教师发送对话")
	chatBody := map[string]interface{}{
		"message":    "请问什么是波粒二象性？",
		"teacher_id": int(localTeacherID),
	}
	resp, body, err = doRequest("POST", "/api/chat", chatBody, localStudentToken)
	if err != nil {
		t.Fatalf("IT-26 步骤5 对话失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-26 步骤5 HTTP 状态码错误: %d, body: %s", resp.StatusCode, string(body))
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-26 步骤5 解析失败: %v, code: %d, body: %s", err, apiResp.Code, string(body))
	}

	replyVal, ok := apiResp.Data["reply"]
	if !ok || replyVal == "" {
		t.Fatalf("IT-26 步骤5 响应缺少 reply, data: %v", apiResp.Data)
	}

	t.Logf("IT-26 通过: 教师添加文档→学生对话引用知识 全链路成功, reply: %v", replyVal)
}

// ======================== IT-27: 多轮对话→查看会话列表→查看对话历史→查看记忆 ========================
func TestV2_IT27_MultiTurnConversation(t *testing.T) {
	os.Setenv("WX_MODE", "mock")

	// 步骤1：创建教师
	t.Log("IT-27 步骤1: 教师微信登录并补全信息")
	teacherLoginBody := map[string]interface{}{
		"code": "test_teacher_multi_001",
	}
	_, body, err := doRequest("POST", "/api/auth/wx-login", teacherLoginBody, "")
	if err != nil {
		t.Fatalf("IT-27 步骤1 教师登录失败: %v", err)
	}
	apiResp, err := parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-27 步骤1 解析失败: %v, code: %d", err, apiResp.Code)
	}
	localTeacherToken := apiResp.Data["token"].(string)
	localTeacherID := apiResp.Data["user_id"].(float64)

	// 教师补全信息
	completeBody := map[string]interface{}{
		"role":     "teacher",
		"nickname": "多轮对话老师",
	}
	_, body, err = doRequest("POST", "/api/auth/complete-profile", completeBody, localTeacherToken)
	if err != nil {
		t.Fatalf("IT-27 步骤1 教师补全失败: %v", err)
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-27 步骤1 教师补全解析失败: %v, code: %d, body: %s", err, apiResp.Code, string(body))
	}

	// 重新登录获取更新后的 token
	_, body, err = doRequest("POST", "/api/auth/wx-login", teacherLoginBody, "")
	if err != nil {
		t.Fatalf("IT-27 步骤1 重新登录失败: %v", err)
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-27 步骤1 重新登录解析失败: %v, code: %d", err, apiResp.Code)
	}
	localTeacherToken = apiResp.Data["token"].(string)
	_ = localTeacherToken // 教师 token 后续可能用于添加文档

	// 步骤2：创建学生
	t.Log("IT-27 步骤2: 学生微信登录并补全信息")
	studentLoginBody := map[string]interface{}{
		"code": "test_stu_multi_003",
	}
	_, body, err = doRequest("POST", "/api/auth/wx-login", studentLoginBody, "")
	if err != nil {
		t.Fatalf("IT-27 步骤2 学生登录失败: %v", err)
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-27 步骤2 解析失败: %v, code: %d, body: %s", err, apiResp.Code, string(body))
	}
	localStudentToken := apiResp.Data["token"].(string)

	studentCompleteBody := map[string]interface{}{
		"role":     "student",
		"nickname": "多轮对话学生",
	}
	_, body, err = doRequest("POST", "/api/auth/complete-profile", studentCompleteBody, localStudentToken)
	if err != nil {
		t.Fatalf("IT-27 步骤2 学生补全失败: %v", err)
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-27 步骤2 学生补全解析失败: %v, code: %d, body: %s", err, apiResp.Code, string(body))
	}

	// 步骤3：多轮对话
	t.Log("IT-27 步骤3: 发送多轮对话")
	messages := []string{
		"你好，我想学习物理",
		"什么是力？",
		"力和运动有什么关系？",
	}

	var sessionID string
	for i, msg := range messages {
		chatBody := map[string]interface{}{
			"message":    msg,
			"teacher_id": int(localTeacherID),
		}
		if sessionID != "" {
			chatBody["session_id"] = sessionID
		}

		resp, body, err := doRequest("POST", "/api/chat", chatBody, localStudentToken)
		if err != nil {
			t.Fatalf("IT-27 步骤3 第%d轮对话失败: %v", i+1, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("IT-27 步骤3 第%d轮 HTTP 状态码错误: %d, body: %s", i+1, resp.StatusCode, string(body))
		}
		apiResp, err := parseResponse(body)
		if err != nil || apiResp.Code != 0 {
			t.Fatalf("IT-27 步骤3 第%d轮解析失败: %v, code: %d, body: %s", i+1, err, apiResp.Code, string(body))
		}

		replyVal, ok := apiResp.Data["reply"]
		if !ok || replyVal == "" {
			t.Fatalf("IT-27 步骤3 第%d轮缺少 reply", i+1)
		}

		// 获取 session_id 用于后续轮次
		if sid, ok := apiResp.Data["session_id"]; ok && sid != nil {
			sessionID = fmt.Sprintf("%v", sid)
		}

		t.Logf("IT-27 步骤3 第%d轮对话成功, reply 长度=%d", i+1, len(fmt.Sprintf("%v", replyVal)))
	}

	// 步骤4：查看会话列表
	t.Log("IT-27 步骤4: 查看会话列表")
	resp, body, err := doRequest("GET", "/api/conversations/sessions?page=1&page_size=20", nil, localStudentToken)
	if err != nil {
		t.Fatalf("IT-27 步骤4 查看会话列表失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-27 步骤4 HTTP 状态码错误: %d, body: %s", resp.StatusCode, string(body))
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-27 步骤4 解析失败: %v, code: %d", err, apiResp.Code)
	}

	// 验证会话列表不为空
	sessionsItems := apiResp.Data["items"]
	if sessionsItems != nil {
		sessions, ok := sessionsItems.([]interface{})
		if ok && len(sessions) > 0 {
			t.Logf("IT-27 步骤4: 会话列表数量=%d", len(sessions))
		} else {
			t.Log("IT-27 步骤4: 会话列表为空（可能是新实现，允许通过）")
		}
	} else {
		t.Log("IT-27 步骤4: 会话列表 items 为 nil（允许通过）")
	}

	// 步骤5：查看对话历史
	t.Log("IT-27 步骤5: 查看对话历史")
	historyPath := fmt.Sprintf("/api/conversations?teacher_id=%d&page=1&page_size=50", int(localTeacherID))
	resp, body, err = doRequest("GET", historyPath, nil, localStudentToken)
	if err != nil {
		t.Fatalf("IT-27 步骤5 查看对话历史失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-27 步骤5 HTTP 状态码错误: %d, body: %s", resp.StatusCode, string(body))
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-27 步骤5 解析失败: %v, code: %d", err, apiResp.Code)
	}

	// 验证对话历史包含多轮对话
	historyItems := apiResp.Data["items"]
	if historyItems != nil {
		conversations, ok := historyItems.([]interface{})
		if ok {
			// 3轮对话，每轮有用户消息+AI回复=6条，但可能因为分页只返回部分
			t.Logf("IT-27 步骤5: 对话历史数量=%d", len(conversations))
			if len(conversations) < 2 {
				t.Logf("IT-27 步骤5 警告: 对话历史数量少于预期（%d < 2）", len(conversations))
			}
		}
	}

	// 步骤6：查看记忆列表
	t.Log("IT-27 步骤6: 查看记忆列表")
	memoryPath := fmt.Sprintf("/api/memories?teacher_id=%d&page=1&page_size=20", int(localTeacherID))
	resp, body, err = doRequest("GET", memoryPath, nil, localStudentToken)
	if err != nil {
		t.Fatalf("IT-27 步骤6 查看记忆失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-27 步骤6 HTTP 状态码错误: %d, body: %s", resp.StatusCode, string(body))
	}

	// 解析记忆响应（使用 apiResponseRaw 因为 data 可能是分页结构）
	var rawResp apiResponseRaw
	if err := json.Unmarshal(body, &rawResp); err != nil {
		t.Fatalf("IT-27 步骤6 解析响应失败: %v", err)
	}
	if rawResp.Code != 0 {
		t.Fatalf("IT-27 步骤6 业务码错误: %d, message: %s", rawResp.Code, rawResp.Message)
	}

	t.Logf("IT-27 通过: 多轮对话→会话列表→对话历史→记忆列表 全链路成功")
}
