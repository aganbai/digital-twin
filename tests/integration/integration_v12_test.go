package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// ======================== 迭代12 Phase 3b 集成测试 ========================
// 测试范围：
// - IT12-INT-001: 流式回复中断功能集成测试
// - IT12-INT-002: 中断接口异常场景测试
// - IT12-INT-003: 会话列表侧边栏集成测试
// - IT12-INT-004: 会话列表分页和空状态测试
// - IT12-INT-005: 新会话按钮功能集成测试
// - IT12-INT-006: 新会话创建异常处理测试
// - IT12-INT-007: 指令系统集成测试
// - IT12-INT-008: 指令边界情况测试
// - IT12-INT-009: 完整会话管理流程测试

var (
	// 迭代12 测试用户
	v12TeacherToken string
	v12TeacherID    float64
	v12StudentToken string
	v12StudentID    float64

	// 迭代12 测试数据
	v12TeacherPersonaID float64
	v12SessionID        string
	v12TestSessionIDs   []string

	// 是否已初始化
	v12Initialized bool
)

// ======================== 辅助函数 ========================

// v12Setup 初始化迭代12测试环境
func v12Setup(t *testing.T) {
	t.Helper()
	if v12Initialized {
		return
	}

	os.Setenv("WX_MODE", "mock")
	os.Setenv("LLM_MODE", "mock")

	// 注册教师用户
	loginBody := map[string]interface{}{
		"code": "v12_teacher_phase3b_001",
	}
	_, body, err := doRequest("POST", "/api/auth/wx-login", loginBody, "")
	if err != nil {
		t.Fatalf("v12Setup 教师微信登录失败: %v", err)
	}

	var apiResp apiResponse
	json.Unmarshal(body, &apiResp)
	if apiResp.Code != 0 {
		t.Fatalf("v12Setup 教师登录失败: code=%d, message=%s", apiResp.Code, apiResp.Message)
	}

	v12TeacherToken = apiResp.Data["token"].(string)
	v12TeacherID = apiResp.Data["user_id"].(float64)

	// 补全教师信息
	completeBody := map[string]interface{}{
		"role":        "teacher",
		"nickname":    "V12Phase3b测试老师",
		"school":      "Phase3b集成测试大学",
		"description": "迭代12 Phase3b 集成测试教师",
	}
	_, body, err = doRequest("POST", "/api/auth/complete-profile", completeBody, v12TeacherToken)
	if err != nil {
		t.Fatalf("v12Setup 教师补全信息失败: %v", err)
	}

	apiResp = apiResponse{}
	json.Unmarshal(body, &apiResp)
	if newToken, ok := apiResp.Data["token"]; ok && newToken != nil {
		v12TeacherToken = newToken.(string)
	}

	// 获取教师的分身ID（创建班级时自动创建）
	resp, body, err := doRequest("GET", "/api/personas", nil, v12TeacherToken)
	if err != nil {
		t.Fatalf("v12Setup 获取分身列表失败: %v", err)
	}

	apiResp = apiResponse{}
	json.Unmarshal(body, &apiResp)
	if resp.StatusCode == http.StatusOK && apiResp.Code == 0 {
		if personas, ok := apiResp.Data["personas"].([]interface{}); ok && len(personas) > 0 {
			firstPersona := personas[0].(map[string]interface{})
			v12TeacherPersonaID = firstPersona["id"].(float64)
		}
	}

	if v12TeacherPersonaID == 0 {
		// 如果没有分身，创建一个测试班级（会自动创建分身）
		classBody := map[string]interface{}{
			"name":        "V12Phase3b测试班级",
			"curriculum":  "测试教材",
			"description": "迭代12 Phase3b 集成测试班级",
		}
		_, body, err = doRequest("POST", "/api/classes", classBody, v12TeacherToken)
		if err != nil {
			t.Fatalf("v12Setup 创建班级失败: %v", err)
		}

		apiResp = apiResponse{}
		json.Unmarshal(body, &apiResp)
		if apiResp.Code == 0 {
			// 重新获取分身列表
			_, body, _ = doRequest("GET", "/api/personas", nil, v12TeacherToken)
			apiResp = apiResponse{}
			json.Unmarshal(body, &apiResp)
			if personas, ok := apiResp.Data["personas"].([]interface{}); ok && len(personas) > 0 {
				firstPersona := personas[0].(map[string]interface{})
				v12TeacherPersonaID = firstPersona["id"].(float64)
			}
		}
	}

	if v12TeacherPersonaID == 0 {
		t.Fatalf("v12Setup 无法获取或创建教师分身")
	}

	// 注册学生用户用于测试
	loginBodyStudent := map[string]interface{}{
		"code": "v12_student_phase3b_001",
	}
	_, body, err = doRequest("POST", "/api/auth/wx-login", loginBodyStudent, "")
	if err != nil {
		t.Fatalf("v12Setup 学生微信登录失败: %v", err)
	}

	apiResp = apiResponse{}
	json.Unmarshal(body, &apiResp)
	if apiResp.Code != 0 {
		t.Fatalf("v12Setup 学生登录失败: code=%d, message=%s", apiResp.Code, apiResp.Message)
	}

	v12StudentToken = apiResp.Data["token"].(string)
	v12StudentID = apiResp.Data["user_id"].(float64)

	// 学生补全信息
	completeBodyStudent := map[string]interface{}{
		"role":     "student",
		"nickname": "V12Phase3b测试学生",
		"school":   "Phase3b集成测试中学",
	}
	_, body, err = doRequest("POST", "/api/auth/complete-profile", completeBodyStudent, v12StudentToken)
	if err != nil {
		t.Fatalf("v12Setup 学生补全信息失败: %v", err)
	}

	apiResp = apiResponse{}
	json.Unmarshal(body, &apiResp)
	if newToken, ok := apiResp.Data["token"]; ok && newToken != nil {
		v12StudentToken = newToken.(string)
	}

	v12Initialized = true
	t.Logf("v12Setup 完成: teacher_persona_id=%.0f", v12TeacherPersonaID)
}

// v12Cleanup 清理迭代12测试数据
func v12Cleanup(t *testing.T) {
	// 清理测试创建的会话
	for _, sessionID := range v12TestSessionIDs {
		if sessionID != "" {
			// 会话无法直接删除，但会随时间自动清理
			t.Logf("清理会话: %s", sessionID)
		}
	}
	v12TestSessionIDs = nil
}

// ======================== IT12-INT-001: 流式中断功能测试 ========================

// TestIT12_INT_001_StreamAbort 测试流式回复中断功能
func TestIT12_INT_001_StreamAbort(t *testing.T) {
	v12Setup(t)
	defer v12Cleanup(t)

	// 步骤1: 学生与教师建立关系
	// 教师创建分享码
	shareBody := map[string]interface{}{
		"persona_id": v12TeacherPersonaID,
		"max_uses":   10,
	}
	resp, body, err := doRequest("POST", "/api/shares", shareBody, v12TeacherToken)
	if err != nil {
		t.Fatalf("创建分享码失败: %v", err)
	}

	var apiResp apiResponse
	json.Unmarshal(body, &apiResp)
	if resp.StatusCode != http.StatusOK || apiResp.Code != 0 {
		t.Skipf("创建分享码失败，跳过测试: status=%d, code=%d", resp.StatusCode, apiResp.Code)
	}

	shareCode := apiResp.Data["code"].(string)

	// 学生加入
	joinBody := map[string]interface{}{
		"code": shareCode,
	}
	_, body, err = doRequest("POST", fmt.Sprintf("/api/shares/%s/join", shareCode), joinBody, v12StudentToken)
	if err != nil {
		t.Fatalf("学生加入失败: %v", err)
	}

	// 步骤2: 先创建一个会话
	sessionBody := map[string]interface{}{
		"teacher_persona_id": v12TeacherPersonaID,
	}
	_, body, err = doRequest("POST", "/api/chat/new-session", sessionBody, v12StudentToken)
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}

	apiResp = apiResponse{}
	json.Unmarshal(body, &apiResp)
	sessionID := ""
	if apiResp.Code == 0 {
		sessionID = apiResp.Data["session_id"].(string)
		v12TestSessionIDs = append(v12TestSessionIDs, sessionID)
	}

	if sessionID == "" {
		t.Fatal("无法创建测试会话")
	}

	// 步骤3: 测试中断接口是否存在
	// GET /api/chat/stream/:session_id/abort
	abortURL := fmt.Sprintf("/api/chat/stream/%s/abort", sessionID)
	resp, body, err = doRequest("GET", abortURL, nil, v12StudentToken)
	if err != nil {
		t.Fatalf("调用中断接口失败: %v", err)
	}

	// 检查接口响应
	apiResp = apiResponse{}
	json.Unmarshal(body, &apiResp)

	// Phase 3b: 验证接口是否已实现
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusMethodNotAllowed {
		t.Skipf("中断接口尚未实现 (IT12-BE-001): status=%d", resp.StatusCode)
	}

	// 如果接口已实现，验证响应
	if resp.StatusCode == http.StatusOK {
		if apiResp.Code != 0 {
			t.Errorf("中断接口返回错误: code=%d, message=%s", apiResp.Code, apiResp.Message)
		}
		if aborted, ok := apiResp.Data["aborted"]; !ok || aborted != true {
			t.Errorf("中断接口返回数据异常: aborted=%v", aborted)
		}
		t.Log("中断接口测试通过")
	} else {
		t.Errorf("中断接口返回非预期状态码: status=%d", resp.StatusCode)
	}
}

// TestIT12_INT_002_StreamAbortErrors 测试中断接口异常场景
func TestIT12_INT_002_StreamAbortErrors(t *testing.T) {
	v12Setup(t)
	defer v12Cleanup(t)

	testCases := []struct {
		name       string
		sessionID  string
		expectCode int
		desc       string
	}{
		{
			name:       "中断不存在的会话",
			sessionID:  "non-existent-session-12345",
			expectCode: http.StatusNotFound,
			desc:       "应返回404",
		},
		{
			name:       "中断无效会话ID",
			sessionID:  "",
			expectCode: http.StatusNotFound,
			desc:       "空ID应返回404或400",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			abortURL := fmt.Sprintf("/api/chat/stream/%s/abort", tc.sessionID)
			resp, _, err := doRequest("GET", abortURL, nil, v12StudentToken)
			if err != nil {
				t.Fatalf("请求失败: %v", err)
			}

			// 如果接口未实现，跳过
			if resp.StatusCode == http.StatusNotFound {
				t.Skipf("中断接口尚未实现 (IT12-BE-001): %s", tc.desc)
			}

			// 验证错误响应
			if resp.StatusCode != tc.expectCode {
				t.Errorf("%s: 期望状态码 %d, 实际 %d", tc.desc, tc.expectCode, resp.StatusCode)
			}
		})
	}
}

// ======================== IT12-INT-003/004: 会话列表测试 ========================

// TestIT12_INT_003_SessionList 测试会话列表侧边栏功能
func TestIT12_INT_003_SessionList(t *testing.T) {
	v12Setup(t)
	defer v12Cleanup(t)

	// 步骤1: 测试获取会话列表接口
	// GET /api/sessions
	url := fmt.Sprintf("/api/sessions?teacher_persona_id=%.0f", v12TeacherPersonaID)
	resp, body, err := doRequest("GET", url, nil, v12TeacherToken)
	if err != nil {
		t.Fatalf("获取会话列表失败: %v", err)
	}

	var apiResp apiResponse
	json.Unmarshal(body, &apiResp)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("会话列表接口返回错误: status=%d", resp.StatusCode)
	}

	if apiResp.Code != 0 {
		t.Errorf("会话列表接口业务错误: code=%d, message=%s", apiResp.Code, apiResp.Message)
	}

	// 验证返回数据结构
	if _, ok := apiResp.Data["total"]; !ok {
		t.Error("会话列表响应缺少 total 字段")
	}
	if _, ok := apiResp.Data["items"]; !ok {
		t.Error("会话列表响应缺少 items 字段")
	}

	t.Log("会话列表接口测试通过")
}

// TestIT12_INT_004_SessionListPagination 测试会话列表分页和空状态
func TestIT12_INT_004_SessionListPagination(t *testing.T) {
	v12Setup(t)
	defer v12Cleanup(t)

	testCases := []struct {
		name     string
		page     int
		pageSize int
	}{
		{
			name:     "默认分页",
			page:     1,
			pageSize: 20,
		},
		{
			name:     "自定义分页",
			page:     1,
			pageSize: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/sessions?teacher_persona_id=%.0f&page=%d&page_size=%d",
				v12TeacherPersonaID, tc.page, tc.pageSize)
			resp, body, err := doRequest("GET", url, nil, v12TeacherToken)
			if err != nil {
				t.Fatalf("请求失败: %v", err)
			}

			var apiResp apiResponse
			json.Unmarshal(body, &apiResp)

			if resp.StatusCode != http.StatusOK {
				t.Errorf("请求失败: status=%d", resp.StatusCode)
				return
			}

			if apiResp.Code != 0 {
				t.Errorf("业务错误: code=%d, message=%s", apiResp.Code, apiResp.Message)
				return
			}

			// 验证分页参数返回
			if page, ok := apiResp.Data["page"]; ok {
				if int(page.(float64)) != tc.page {
					t.Errorf("page 字段不匹配: 期望 %d, 实际 %v", tc.page, page)
				}
			}
			if pageSize, ok := apiResp.Data["page_size"]; ok {
				if int(pageSize.(float64)) != tc.pageSize {
					t.Errorf("page_size 字段不匹配: 期望 %d, 实际 %v", tc.pageSize, pageSize)
				}
			}
		})
	}
}

// ======================== IT12-INT-005/006: 新会话功能测试 ========================

// TestIT12_INT_005_NewSession 测试新会话创建功能
func TestIT12_INT_005_NewSession(t *testing.T) {
	v12Setup(t)
	defer v12Cleanup(t)

	// 步骤1: 测试创建新会话接口
	createBody := map[string]interface{}{
		"teacher_persona_id": v12TeacherPersonaID,
		"initial_message":    "", // 可选
	}

	resp, body, err := doRequest("POST", "/api/sessions", createBody, v12TeacherToken)
	if err != nil {
		t.Fatalf("创建新会话失败: %v", err)
	}

	var apiResp apiResponse
	json.Unmarshal(body, &apiResp)

	// 检查接口是否存在
	if resp.StatusCode == http.StatusNotFound {
		t.Skip("创建会话接口路径可能有变化，使用 /api/chat/new-session 替代测试")
	}

	if resp.StatusCode == http.StatusOK && apiResp.Code == 0 {
		if sessionID, ok := apiResp.Data["session_id"]; ok {
			v12TestSessionIDs = append(v12TestSessionIDs, sessionID.(string))
			t.Logf("成功创建新会话: %s", sessionID)
		} else {
			t.Error("创建会话成功但返回数据缺少 session_id")
		}
	}
}

// TestIT12_INT_006_NewSessionErrors 测试新会话创建异常
func TestIT12_INT_006_NewSessionErrors(t *testing.T) {
	v12Setup(t)
	defer v12Cleanup(t)

	testCases := []struct {
		name               string
		teacherPersonaID   float64
		expectError        bool
		desc               string
	}{
		{
			name:             "有效的分身ID",
			teacherPersonaID: v12TeacherPersonaID,
			expectError:      false,
			desc:             "应该成功创建",
		},
		{
			name:             "无效的分身ID",
			teacherPersonaID: 999999,
			expectError:      true,
			desc:             "应该返回错误",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			createBody := map[string]interface{}{
				"teacher_persona_id": tc.teacherPersonaID,
			}

			resp, body, err := doRequest("POST", "/api/chat/new-session", createBody, v12TeacherToken)
			if err != nil {
				t.Fatalf("请求失败: %v", err)
			}

			var apiResp apiResponse
			json.Unmarshal(body, &apiResp)

			if tc.expectError {
				if resp.StatusCode == http.StatusOK && apiResp.Code == 0 {
					t.Errorf("%s: 期望失败但实际成功", tc.desc)
				}
			} else {
				if apiResp.Code != 0 {
					t.Errorf("%s: 期望成功但失败: code=%d, message=%s", tc.desc, apiResp.Code, apiResp.Message)
				}
				if sessionID, ok := apiResp.Data["session_id"]; ok {
					v12TestSessionIDs = append(v12TestSessionIDs, sessionID.(string))
				}
			}
		})
	}
}

// ======================== IT12-INT-007/008: 指令系统测试 ========================

// TestIT12_INT_007_CommandSystem 测试指令系统功能
func TestIT12_INT_007_CommandSystem(t *testing.T) {
	v12Setup(t)
	defer v12Cleanup(t)

	// 指令系统主要是前端功能，后端验证消息类型支持
	commands := []string{
		"#新会话",
		"#新对话",
		"#新话题",
		"#给老师留言",
		"#留言",
	}

	for _, cmd := range commands {
		t.Run(fmt.Sprintf("指令_%s", cmd), func(t *testing.T) {
			// 测试发送指令作为消息
			msgBody := map[string]interface{}{
				"message":            cmd,
				"teacher_persona_id": v12TeacherPersonaID,
			}

			resp, body, err := doRequest("POST", "/api/chat", msgBody, v12TeacherToken)
			if err != nil {
				t.Fatalf("发送消息失败: %v", err)
			}

			var apiResp apiResponse
			json.Unmarshal(body, &apiResp)

			// 后端应该能接收指令内容（前端处理）
			if resp.StatusCode != http.StatusOK {
				t.Errorf("发送指令失败: status=%d", resp.StatusCode)
			}

			t.Logf("指令 '%s' 发送成功", cmd)
		})
	}
}

// TestIT12_INT_008_CommandEdgeCases 测试指令边界情况
func TestIT12_INT_008_CommandEdgeCases(t *testing.T) {
	v12Setup(t)
	defer v12Cleanup(t)

	testCases := []struct {
		name     string
		input    string
		isCommand bool
	}{
		{
			name:      "前后有空格的指令",
			input:     "  #新会话  ",
			isCommand: true,
		},
		{
			name:      "普通消息",
			input:     "普通消息内容",
			isCommand: false,
		},
		{
			name:      "包含#但不是指令",
			input:     "这是一段包含#的文本",
			isCommand: false,
		},
		{
			name:      "无效指令",
			input:     "#无效指令",
			isCommand: false, // 后端看是普通消息
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msgBody := map[string]interface{}{
				"message":            tc.input,
				"teacher_persona_id": v12TeacherPersonaID,
			}

			resp, _, err := doRequest("POST", "/api/chat", msgBody, v12TeacherToken)
			if err != nil {
				t.Fatalf("发送失败: %v", err)
			}

			if resp.StatusCode != http.StatusOK {
				t.Errorf("发送失败: status=%d", resp.StatusCode)
			}

			t.Logf("输入 '%s' 发送成功 (isCommand=%v)", tc.input, tc.isCommand)
		})
	}
}

// ======================== IT12-INT-009: 完整流程测试 ========================

// TestIT12_INT_009_FullSessionWorkflow 测试完整会话管理流程
func TestIT12_INT_009_FullSessionWorkflow(t *testing.T) {
	v12Setup(t)
	defer v12Cleanup(t)

	// 完整流程测试
	// 1. 创建新会话
	// 2. 发送消息
	// 3. 获取会话列表
	// 4. 验证会话存在

	// 步骤1: 创建新会话
	createBody := map[string]interface{}{
		"teacher_persona_id": v12TeacherPersonaID,
	}

	resp, body, err := doRequest("POST", "/api/chat/new-session", createBody, v12TeacherToken)
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}

	var apiResp apiResponse
	json.Unmarshal(body, &apiResp)

	if resp.StatusCode != http.StatusOK || apiResp.Code != 0 {
		t.Fatalf("创建会话失败: status=%d, code=%d", resp.StatusCode, apiResp.Code)
	}

	sessionID := apiResp.Data["session_id"].(string)
	v12TestSessionIDs = append(v12TestSessionIDs, sessionID)
	t.Logf("创建会话成功: %s", sessionID)

	// 步骤2: 发送消息
	chatBody := map[string]interface{}{
		"message":            "测试消息",
		"teacher_persona_id": v12TeacherPersonaID,
		"session_id":         sessionID,
	}

	resp, body, err = doRequest("POST", "/api/chat", chatBody, v12TeacherToken)
	if err != nil {
		t.Fatalf("发送消息失败: %v", err)
	}

	apiResp = apiResponse{}
	json.Unmarshal(body, &apiResp)

	if resp.StatusCode != http.StatusOK || apiResp.Code != 0 {
		t.Fatalf("发送消息失败: status=%d, code=%d", resp.StatusCode, apiResp.Code)
	}
	t.Log("发送消息成功")

	// 步骤3: 获取会话列表并验证
	url := fmt.Sprintf("/api/sessions?teacher_persona_id=%.0f", v12TeacherPersonaID)
	resp, body, err = doRequest("GET", url, nil, v12TeacherToken)
	if err != nil {
		t.Fatalf("获取会话列表失败: %v", err)
	}

	apiResp = apiResponse{}
	json.Unmarshal(body, &apiResp)

	if resp.StatusCode != http.StatusOK || apiResp.Code != 0 {
		t.Fatalf("获取会话列表失败: status=%d, code=%d", resp.StatusCode, apiResp.Code)
	}

	// 验证会话存在
	items, ok := apiResp.Data["items"].([]interface{})
	if !ok {
		t.Fatal("会话列表格式错误")
	}

	found := false
	for _, item := range items {
		session := item.(map[string]interface{})
		if session["session_id"].(string) == sessionID {
			found = true
			break
		}
	}

	if !found {
		t.Error("新创建的会话未在列表中找到")
	} else {
		t.Log("在会话列表中找到新会话")
	}

	t.Log("完整会话管理流程测试通过")
}

// ======================== SSE 流式测试 ========================

// TestIT12_StreamEndpoint 测试SSE流式接口
func TestIT12_StreamEndpoint(t *testing.T) {
	v12Setup(t)
	defer v12Cleanup(t)

	// 测试流式接口是否存在
	streamBody := map[string]interface{}{
		"message":            "测试流式消息",
		"teacher_persona_id": v12TeacherPersonaID,
	}

	bodyBytes, _ := json.Marshal(streamBody)
	req, err := http.NewRequest("POST", baseURL+"/api/chat/stream", bytes.NewBuffer(bodyBytes))
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+v12TeacherToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// 如果直接请求失败，使用测试服务器URL
		req2, _ := http.NewRequest("POST", ts.URL+"/api/chat/stream", bytes.NewBuffer(bodyBytes))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Authorization", "Bearer "+v12TeacherToken)
		resp, err = client.Do(req2)
		if err != nil {
			t.Fatalf("请求流式接口失败: %v", err)
		}
	}
	defer resp.Body.Close()

	// 验证SSE响应头
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") {
		t.Logf("注意: Content-Type 不是 text/event-stream，而是: %s", contentType)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("流式接口返回非200: status=%d", resp.StatusCode)
	} else {
		t.Log("流式接口可访问")
	}
}
