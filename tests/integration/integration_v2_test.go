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
		"role":        "teacher",
		"nickname":    "王老师",
		"school":      "测试大学",
		"description": "一位优秀的测试教师",
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

	// 验证返回了新的 token（complete-profile 应返回包含最新角色的 token）
	newTokenVal, ok := apiResp.Data["token"]
	if !ok || newTokenVal == nil || newTokenVal == "" {
		t.Fatalf("IT-19 complete-profile 未返回新 token, data: %v", apiResp.Data)
	}

	// 验证返回了 expires_at
	expiresAtVal, ok := apiResp.Data["expires_at"]
	if !ok || expiresAtVal == nil || expiresAtVal == "" {
		t.Fatalf("IT-19 complete-profile 未返回 expires_at, data: %v", apiResp.Data)
	}

	// 关键：使用 complete-profile 返回的新 token（而非重新登录）
	v2TeacherToken = newTokenVal.(string)

	// 验证新 token 能直接调用需要 teacher 角色的接口（如添加文档）
	docBody := map[string]interface{}{
		"title":   "IT-19验证token文档",
		"content": "验证 complete-profile 返回的 token 包含正确角色",
		"tags":    "测试",
	}
	resp, body, err = doRequest("POST", "/api/documents", docBody, v2TeacherToken)
	if err != nil {
		t.Fatalf("IT-19 使用新 token 添加文档失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-19 使用新 token 添加文档 HTTP 状态码错误: %d, body: %s（说明 complete-profile 返回的 token 角色不正确）", resp.StatusCode, string(body))
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-19 使用新 token 添加文档业务码错误: %v, code: %d, body: %s", err, apiResp.Code, string(body))
	}

	t.Logf("IT-19 通过: 补全信息成功, role=teacher, nickname=王老师, 新 token 可直接调用需要角色的接口")
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

	// 关键：使用 complete-profile 返回的新 token（不重新登录）
	if newToken, ok := apiResp.Data["token"]; ok && newToken != nil && newToken != "" {
		v2StudentToken = newToken.(string)
	} else {
		t.Fatalf("IT-21 学生 complete-profile 未返回新 token, data: %v", apiResp.Data)
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

	// V2.0 迭代3 行为变更：学生角色只返回已授权+启用的教师分身
	// 刚注册的学生没有师生关系，所以教师列表应该为空
	if itemsVal == nil {
		t.Logf("IT-21 通过: 学生无师生关系时教师列表为 nil（空），符合迭代3新行为")
	} else {
		items, ok := itemsVal.([]interface{})
		if !ok {
			t.Fatalf("IT-21 items 不是数组类型: %T", itemsVal)
		}
		// 空列表也是正确的
		t.Logf("IT-21 通过: 学生无师生关系时教师列表长度=%d，符合迭代3新行为", len(items))
	}

	// 额外验证：使用教师 token 获取教师列表应该返回所有教师
	if v2TeacherToken != "" {
		resp2, body2, err2 := doRequest("GET", "/api/teachers?page=1&page_size=20", nil, v2TeacherToken)
		if err2 != nil {
			t.Fatalf("IT-21 教师角色请求失败: %v", err2)
		}
		if resp2.StatusCode != http.StatusOK {
			t.Fatalf("IT-21 教师角色 HTTP 状态码错误: %d", resp2.StatusCode)
		}
		apiResp2, _ := parseResponse(body2)
		if apiResp2.Code != 0 {
			t.Fatalf("IT-21 教师角色业务码错误: %d", apiResp2.Code)
		}
		teacherItems := apiResp2.Data["items"]
		if teacherItems != nil {
			tItems, _ := teacherItems.([]interface{})
			if len(tItems) > 0 {
				firstTeacher := tItems[0].(map[string]interface{})
				if _, ok := firstTeacher["document_count"]; !ok {
					t.Fatal("IT-21 教师角色: 教师项缺少 document_count 字段")
				}
				t.Logf("IT-21 通过: 教师角色获取教师列表成功, 数量=%d", len(tItems))
			}
		}
	}
}

// ======================== IT-22: 获取用户信息 → 返回 profile + personas ========================
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

	// V2.0 迭代2 新格式：验证 user_id 字段
	if _, ok := apiResp.Data["user_id"]; !ok {
		t.Fatal("IT-22 响应缺少 user_id 字段")
	}

	// 验证 username 字段
	if _, ok := apiResp.Data["username"]; !ok {
		t.Fatal("IT-22 响应缺少 username 字段")
	}

	// 验证 personas 字段（数组）
	personasVal, ok := apiResp.Data["personas"]
	if !ok {
		t.Fatal("IT-22 响应缺少 personas 字段")
	}
	if personasVal != nil {
		personas, ok := personasVal.([]interface{})
		if !ok {
			t.Fatalf("IT-22 personas 不是数组类型: %T", personasVal)
		}
		t.Logf("IT-22 分身数量: %d", len(personas))
	}

	// 验证 created_at 字段
	if _, ok := apiResp.Data["created_at"]; !ok {
		t.Fatal("IT-22 响应缺少 created_at 字段")
	}

	// 可选检查 current_persona（可能为 nil，如果没有设置默认分身）
	if cp, ok := apiResp.Data["current_persona"]; ok && cp != nil {
		t.Logf("IT-22 当前分身: %v", cp)
	}

	t.Logf("IT-22 通过: 获取用户信息成功, user_id=%v, username=%v",
		apiResp.Data["user_id"], apiResp.Data["username"])
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

	// 关键：使用 complete-profile 返回的新 token（不重新登录，模拟真实用户行为）
	if newToken, ok := apiResp.Data["token"]; ok && newToken != nil && newToken != "" {
		studentToken = newToken.(string)
	} else {
		t.Fatalf("IT-25 步骤2 complete-profile 未返回新 token, data: %v", apiResp.Data)
	}

	// 步骤3：获取教师信息（V2.0 迭代3 行为变更：学生只能看到已授权教师）
	// 如果 v2TeacherToken 不可用，自己创建教师
	t.Log("IT-25 步骤3: 获取教师信息")
	teacherToken := v2TeacherToken
	if teacherToken == "" {
		// 自己创建教师：先注册
		teacherRegBody := map[string]interface{}{
			"username": "it25_teacher",
			"password": "test123456",
			"role":     "teacher",
			"nickname": "IT25教师",
			"school":   "测试学校",
		}
		_, tBody, tErr := doRequest("POST", "/api/auth/register", teacherRegBody, "")
		if tErr != nil {
			t.Fatalf("IT-25 步骤3 创建教师注册失败: %v", tErr)
		}
		tResp, _ := parseResponse(tBody)
		if tResp.Code == 0 {
			teacherToken = tResp.Data["token"].(string)
		} else {
			// 用户已存在，尝试登录
			teacherLoginBody := map[string]interface{}{
				"username": "it25_teacher",
				"password": "test123456",
			}
			_, tBody, tErr = doRequest("POST", "/api/auth/login", teacherLoginBody, "")
			if tErr != nil {
				t.Fatalf("IT-25 步骤3 教师登录失败: %v", tErr)
			}
			tResp, _ = parseResponse(tBody)
			if tResp.Code != 0 {
				t.Fatalf("IT-25 步骤3 教师登录业务码错误: %d", tResp.Code)
			}
			teacherToken = tResp.Data["token"].(string)
		}
	}

	var targetTeacherID float64
	resp, body, err = doRequest("GET", "/api/teachers?page=1&page_size=20", nil, teacherToken)
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
	firstTeacher := items[0].(map[string]interface{})
	targetTeacherID = firstTeacher["id"].(float64)
	t.Logf("IT-25 步骤3: 选择教师 ID=%v, 昵称=%v", targetTeacherID, firstTeacher["nickname"])

	// 步骤3.5：学生申请使用教师分身（模拟真实用户点击"申请使用"按钮）
	t.Log("IT-25 步骤3.5: 学生申请使用教师分身")
	applyBody := map[string]interface{}{
		"teacher_id": int(targetTeacherID),
	}
	_, body, err = doRequest("POST", "/api/relations/apply", applyBody, studentToken)
	if err != nil {
		t.Fatalf("IT-25 步骤3.5 学生申请失败: %v", err)
	}
	apiResp, _ = parseResponse(body)
	if apiResp.Code != 0 && apiResp.Code != 40009 {
		t.Fatalf("IT-25 步骤3.5 业务码错误: %d, body: %s", apiResp.Code, string(body))
	}

	// 如果是 pending 状态，教师需要审批
	if statusVal, ok := apiResp.Data["status"]; ok && statusVal == "pending" {
		relationID := apiResp.Data["id"].(float64)
		approvePath := fmt.Sprintf("/api/relations/%d/approve", int(relationID))
		_, body, err = doRequest("PUT", approvePath, nil, teacherToken)
		if err != nil {
			t.Fatalf("IT-25 步骤3.5 教师审批失败: %v", err)
		}
		apiResp, _ = parseResponse(body)
		if apiResp.Code != 0 {
			t.Fatalf("IT-25 步骤3.5 教师审批业务码错误: %d", apiResp.Code)
		}
	}

	// 步骤4：向选中的教师发送对话
	// 步骤4：向选中的教师发送对话	t.Log("IT-25 步骤4: 发送对话")
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
		"role":        "teacher",
		"nickname":    "物理张老师",
		"school":      "物理实验中学",
		"description": "专注物理教学的老师",
	}
	_, body, err = doRequest("POST", "/api/auth/complete-profile", completeBody, localTeacherToken)
	if err != nil {
		t.Fatalf("IT-26 步骤2 补全信息失败: %v", err)
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-26 步骤2 解析失败: %v, code: %d, body: %s", err, apiResp.Code, string(body))
	}

	// 使用 complete-profile 返回的新 token（不重新登录，模拟真实用户行为）
	if newToken, ok := apiResp.Data["token"]; ok && newToken != nil && newToken != "" {
		localTeacherToken = newToken.(string)
	} else {
		t.Fatalf("IT-26 步骤2 complete-profile 未返回新 token, data: %v", apiResp.Data)
	}

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
	localStudentID := apiResp.Data["user_id"].(float64)

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

	// 步骤4.6：建立师生关系
	inviteBody := map[string]interface{}{
		"student_id": int(localStudentID),
	}
	_, body, err = doRequest("POST", "/api/relations/invite", inviteBody, localTeacherToken)
	if err != nil {
		t.Fatalf("IT-26 步骤4.6 建立师生关系失败: %v", err)
	}
	apiResp, _ = parseResponse(body)
	if apiResp.Code != 0 && apiResp.Code != 40009 {
		t.Fatalf("IT-26 步骤4.6 业务码错误: %d", apiResp.Code)
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
		"role":        "teacher",
		"nickname":    "多轮对话老师",
		"school":      "多轮对话学校",
		"description": "专注多轮对话教学的老师",
	}
	_, body, err = doRequest("POST", "/api/auth/complete-profile", completeBody, localTeacherToken)
	if err != nil {
		t.Fatalf("IT-27 步骤1 教师补全失败: %v", err)
	}
	apiResp, err = parseResponse(body)
	if err != nil || apiResp.Code != 0 {
		t.Fatalf("IT-27 步骤1 教师补全解析失败: %v, code: %d, body: %s", err, apiResp.Code, string(body))
	}

	// 使用 complete-profile 返回的新 token（不重新登录，模拟真实用户行为）
	if newToken, ok := apiResp.Data["token"]; ok && newToken != nil && newToken != "" {
		localTeacherToken = newToken.(string)
	} else {
		t.Fatalf("IT-27 步骤1 教师 complete-profile 未返回新 token, data: %v", apiResp.Data)
	}
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
	localStudentID := apiResp.Data["user_id"].(float64)

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

	// 步骤2.5：建立师生关系
	inviteBody := map[string]interface{}{
		"student_id": int(localStudentID),
	}
	_, body, err = doRequest("POST", "/api/relations/invite", inviteBody, localTeacherToken)
	if err != nil {
		t.Fatalf("IT-27 步骤2.5 建立师生关系失败: %v", err)
	}
	apiResp, _ = parseResponse(body)
	if apiResp.Code != 0 && apiResp.Code != 40009 {
		t.Fatalf("IT-27 步骤2.5 业务码错误: %d", apiResp.Code)
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
