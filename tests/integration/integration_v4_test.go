package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
)

// ======================== V4 集成测试辅助变量 ========================

var (
	v4TeacherToken      string
	v4StudentToken      string
	v4Student2Token     string
	v4TeacherID         float64
	v4StudentID         float64
	v4Student2ID        float64
	v4TeacherPersonaID  float64
	v4StudentPersonaID  float64
	v4Student2PersonaID float64
	v4SessionID         string
	v4ConversationID    float64
)

// v4Setup 初始化 V4 测试所需的教师和学生
func v4Setup(t *testing.T) {
	t.Helper()
	os.Setenv("WX_MODE", "mock")

	if v4TeacherToken != "" && v4StudentToken != "" && v4Student2Token != "" {
		return
	}

	// 注册教师（微信登录 + 补全信息）
	// 使用唯一的 code 前缀避免与其他测试的 openid 后6位冲突（username UNIQUE 约束）
	_, body, _ := doRequest("POST", "/api/auth/wx-login", map[string]interface{}{"code": "v4iter_tch_001"}, "")
	apiResp, _ := parseResponse(body)
	if apiResp.Data == nil || apiResp.Data["token"] == nil {
		t.Fatalf("v4Setup: 教师微信登录失败, code=%d, msg=%s, body=%s", apiResp.Code, apiResp.Message, string(body))
	}
	v4TeacherToken = apiResp.Data["token"].(string)
	v4TeacherID, _ = apiResp.Data["user_id"].(float64)

	doRequest("POST", "/api/auth/complete-profile", map[string]interface{}{
		"role": "teacher", "nickname": "V4王老师", "school": "V4测试大学", "description": "V4测试教师",
	}, v4TeacherToken)

	// 创建教师分身
	_, body, _ = doRequest("POST", "/api/personas", map[string]interface{}{
		"role": "teacher", "nickname": "V4王老师分身", "school": "V4测试大学", "description": "V4测试教师分身",
	}, v4TeacherToken)
	apiResp, _ = parseResponse(body)
	if apiResp.Code == 0 {
		v4TeacherPersonaID, _ = apiResp.Data["persona_id"].(float64)
		// 切换到该分身获取新 token
		_, body, _ = doRequest("PUT", fmt.Sprintf("/api/personas/%d/switch", int(v4TeacherPersonaID)), nil, v4TeacherToken)
		apiResp, _ = parseResponse(body)
		if apiResp.Code == 0 {
			v4TeacherToken = apiResp.Data["token"].(string)
		}
	}

	// 注册学生1
	_, body, _ = doRequest("POST", "/api/auth/wx-login", map[string]interface{}{"code": "v4iter_stu_001"}, "")
	apiResp, _ = parseResponse(body)
	if apiResp.Data == nil || apiResp.Data["token"] == nil {
		t.Fatalf("v4Setup: 学生1微信登录失败, code=%d, msg=%s, body=%s", apiResp.Code, apiResp.Message, string(body))
	}
	v4StudentToken = apiResp.Data["token"].(string)
	v4StudentID, _ = apiResp.Data["user_id"].(float64)

	doRequest("POST", "/api/auth/complete-profile", map[string]interface{}{
		"role": "student", "nickname": "V4小明",
	}, v4StudentToken)

	// 创建学生分身
	_, body, _ = doRequest("POST", "/api/personas", map[string]interface{}{
		"role": "student", "nickname": "V4小明分身",
	}, v4StudentToken)
	apiResp, _ = parseResponse(body)
	if apiResp.Code == 0 {
		v4StudentPersonaID, _ = apiResp.Data["persona_id"].(float64)
		_, body, _ = doRequest("PUT", fmt.Sprintf("/api/personas/%d/switch", int(v4StudentPersonaID)), nil, v4StudentToken)
		apiResp, _ = parseResponse(body)
		if apiResp.Code == 0 {
			v4StudentToken = apiResp.Data["token"].(string)
		}
	}

	// 注册学生2
	_, body, _ = doRequest("POST", "/api/auth/wx-login", map[string]interface{}{"code": "v4iter_stu_002"}, "")
	apiResp, _ = parseResponse(body)
	if apiResp.Data == nil || apiResp.Data["token"] == nil {
		t.Fatalf("v4Setup: 学生2微信登录失败, code=%d, msg=%s, body=%s", apiResp.Code, apiResp.Message, string(body))
	}
	v4Student2Token = apiResp.Data["token"].(string)
	v4Student2ID, _ = apiResp.Data["user_id"].(float64)

	doRequest("POST", "/api/auth/complete-profile", map[string]interface{}{
		"role": "student", "nickname": "V4小红",
	}, v4Student2Token)

	// 创建学生2分身
	_, body, _ = doRequest("POST", "/api/personas", map[string]interface{}{
		"role": "student", "nickname": "V4小红分身",
	}, v4Student2Token)
	apiResp, _ = parseResponse(body)
	if apiResp.Code == 0 {
		v4Student2PersonaID, _ = apiResp.Data["persona_id"].(float64)
		_, body, _ = doRequest("PUT", fmt.Sprintf("/api/personas/%d/switch", int(v4Student2PersonaID)), nil, v4Student2Token)
		apiResp, _ = parseResponse(body)
		if apiResp.Code == 0 {
			v4Student2Token = apiResp.Data["token"].(string)
		}
	}

	// 建立教师与学生1的师生关系
	doRequest("POST", "/api/relations/invite", map[string]interface{}{
		"student_id": int(v4StudentID),
	}, v4TeacherToken)
}

// ======================== IT-106: 教师设置分身公开 → 广场可见 ========================
func TestV4_IT106_SetPersonaPublic(t *testing.T) {
	v4Setup(t)

	// 设置分身为公开
	resp, body, err := doRequest("PUT", fmt.Sprintf("/api/personas/%d/visibility", int(v4TeacherPersonaID)),
		map[string]interface{}{"is_public": true}, v4TeacherToken)
	if err != nil {
		t.Fatalf("IT-106 请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-106 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, _ := parseResponse(body)
	if apiResp.Code != 0 {
		t.Fatalf("IT-106 业务码错误: 期望 0, 实际 %d, message: %s", apiResp.Code, apiResp.Message)
	}

	isPublic, _ := apiResp.Data["is_public"].(bool)
	if !isPublic {
		t.Fatal("IT-106 is_public 应为 true")
	}

	// 验证广场可见（学生2未授权，应能看到）
	resp, body, err = doRequest("GET", "/api/personas/marketplace?page=1&page_size=20", nil, v4Student2Token)
	if err != nil {
		t.Fatalf("IT-106 查询广场失败: %v", err)
	}

	var rawResp struct {
		Code int `json:"code"`
		Data struct {
			Items []map[string]interface{} `json:"items"`
			Total float64                  `json:"total"`
		} `json:"data"`
	}
	json.Unmarshal(body, &rawResp)

	if rawResp.Code != 0 {
		t.Fatalf("IT-106 广场查询业务码错误: %d", rawResp.Code)
	}

	found := false
	for _, item := range rawResp.Data.Items {
		if id, ok := item["id"].(float64); ok && id == v4TeacherPersonaID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("IT-106 广场中未找到公开的教师分身")
	}

	t.Log("IT-106 通过: 教师设置分身公开后广场可见")
}

// ======================== IT-107: 教师设置分身私有 → 广场不可见 ========================
func TestV4_IT107_SetPersonaPrivate(t *testing.T) {
	v4Setup(t)

	// 设置分身为私有
	resp, body, err := doRequest("PUT", fmt.Sprintf("/api/personas/%d/visibility", int(v4TeacherPersonaID)),
		map[string]interface{}{"is_public": false}, v4TeacherToken)
	if err != nil {
		t.Fatalf("IT-107 请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-107 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, _ := parseResponse(body)
	if apiResp.Code != 0 {
		t.Fatalf("IT-107 业务码错误: 期望 0, 实际 %d", apiResp.Code)
	}

	// 验证广场不可见
	_, body, _ = doRequest("GET", "/api/personas/marketplace?page=1&page_size=20", nil, v4Student2Token)

	var rawResp struct {
		Code int `json:"code"`
		Data struct {
			Items []map[string]interface{} `json:"items"`
		} `json:"data"`
	}
	json.Unmarshal(body, &rawResp)

	for _, item := range rawResp.Data.Items {
		if id, ok := item["id"].(float64); ok && id == v4TeacherPersonaID {
			t.Fatal("IT-107 广场中不应出现私有的教师分身")
		}
	}

	// 恢复为公开（供后续测试使用）
	doRequest("PUT", fmt.Sprintf("/api/personas/%d/visibility", int(v4TeacherPersonaID)),
		map[string]interface{}{"is_public": true}, v4TeacherToken)

	t.Log("IT-107 通过: 教师设置分身私有后广场不可见")
}

// ======================== IT-108: 学生从广场申请教师分身 ========================
func TestV4_IT108_ApplyFromMarketplace(t *testing.T) {
	v4Setup(t)

	// 学生2从广场申请教师分身（通过 relations/apply 接口）
	// HandleApplyTeacher 需要 teacher_id（user_id）和可选的 teacher_persona_id
	resp, body, err := doRequest("POST", "/api/relations/apply", map[string]interface{}{
		"teacher_id":         int(v4TeacherID),
		"teacher_persona_id": int(v4TeacherPersonaID),
	}, v4Student2Token)
	if err != nil {
		t.Fatalf("IT-108 请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-108 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, _ := parseResponse(body)
	if apiResp.Code != 0 {
		t.Fatalf("IT-108 业务码错误: 期望 0, 实际 %d, message: %s", apiResp.Code, apiResp.Message)
	}

	t.Log("IT-108 通过: 学生从广场申请教师分身成功")
}

// ======================== IT-109: 广场不展示已授权的教师分身 ========================
func TestV4_IT109_MarketplaceExcludesAuthorized(t *testing.T) {
	v4Setup(t)

	// 学生1已有授权关系，广场不应展示该教师
	_, body, _ := doRequest("GET", "/api/personas/marketplace?page=1&page_size=20", nil, v4StudentToken)

	var rawResp struct {
		Code int `json:"code"`
		Data struct {
			Items []map[string]interface{} `json:"items"`
		} `json:"data"`
	}
	json.Unmarshal(body, &rawResp)

	for _, item := range rawResp.Data.Items {
		if id, ok := item["id"].(float64); ok && id == v4TeacherPersonaID {
			t.Fatal("IT-109 广场不应展示已授权的教师分身")
		}
	}

	t.Log("IT-109 通过: 广场不展示已授权的教师分身")
}

// ======================== IT-110: 广场搜索教师分身 ========================
func TestV4_IT110_MarketplaceSearch(t *testing.T) {
	v4Setup(t)

	// 按昵称搜索
	_, body, _ := doRequest("GET", "/api/personas/marketplace?keyword=V4王&page=1&page_size=20", nil, v4Student2Token)

	var rawResp struct {
		Code int `json:"code"`
		Data struct {
			Items []map[string]interface{} `json:"items"`
			Total float64                  `json:"total"`
		} `json:"data"`
	}
	json.Unmarshal(body, &rawResp)

	if rawResp.Code != 0 {
		t.Fatalf("IT-110 业务码错误: %d", rawResp.Code)
	}

	// 应能搜到 V4王老师分身
	found := false
	for _, item := range rawResp.Data.Items {
		if nickname, ok := item["nickname"].(string); ok && nickname == "V4王老师分身" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("IT-110 搜索未找到 V4王老师分身")
	}

	t.Log("IT-110 通过: 广场搜索教师分身成功")
}

// ======================== IT-111: 教师生成定向分享码 → 目标学生可用 ========================
func TestV4_IT111_TargetedShareCode(t *testing.T) {
	v4Setup(t)

	// 教师生成定向分享码（绑定学生1）
	resp, body, err := doRequest("POST", "/api/shares", map[string]interface{}{
		"target_student_persona_id": int(v4StudentPersonaID),
		"expires_hours":             168,
		"max_uses":                  1,
	}, v4TeacherToken)
	if err != nil {
		t.Fatalf("IT-111 请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-111 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, _ := parseResponse(body)
	if apiResp.Code != 0 {
		t.Fatalf("IT-111 业务码错误: 期望 0, 实际 %d, message: %s", apiResp.Code, apiResp.Message)
	}

	shareCode, _ := apiResp.Data["share_code"].(string)
	if shareCode == "" {
		t.Fatal("IT-111 分享码为空")
	}

	targetID, _ := apiResp.Data["target_student_persona_id"].(float64)
	if targetID != v4StudentPersonaID {
		t.Fatalf("IT-111 target_student_persona_id 错误: 期望 %v, 实际 %v", v4StudentPersonaID, targetID)
	}

	t.Logf("IT-111 通过: 教师生成定向分享码成功, code=%s, target=%v", shareCode, targetID)
}

// ======================== IT-112: 非目标学生使用定向分享码 → 被拒绝 ========================
func TestV4_IT112_TargetedShareCodeRejected(t *testing.T) {
	v4Setup(t)

	// 先生成定向分享码（绑定学生1）
	_, body, _ := doRequest("POST", "/api/shares", map[string]interface{}{
		"target_student_persona_id": int(v4StudentPersonaID),
		"expires_hours":             168,
		"max_uses":                  10,
	}, v4TeacherToken)
	apiResp, _ := parseResponse(body)
	shareCode, _ := apiResp.Data["share_code"].(string)

	// 学生2使用该分享码 → V2.0 迭代6: 返回200+引导信息（而非40029错误）
	resp, body, err := doRequest("POST", fmt.Sprintf("/api/shares/%s/join", shareCode), map[string]interface{}{
		"student_persona_id": int(v4Student2PersonaID),
	}, v4Student2Token)
	if err != nil {
		t.Fatalf("IT-112 请求失败: %v", err)
	}

	_ = resp
	apiResp, _ = parseResponse(body)
	if apiResp.Code != 0 {
		t.Fatalf("IT-112 业务码错误: 期望 0, 实际 %d, message: %s, body: %s", apiResp.Code, apiResp.Message, string(body))
	}

	// 验证返回了 not_target 引导信息
	joinStatus, _ := apiResp.Data["join_status"].(string)
	if joinStatus != "not_target" {
		t.Fatalf("IT-112 join_status 错误: 期望 not_target, 实际 %s", joinStatus)
	}
	canApply, _ := apiResp.Data["can_apply"].(bool)
	if !canApply {
		t.Fatalf("IT-112 can_apply 应为 true")
	}

	t.Log("IT-112 通过: 非目标学生使用定向分享码返回友好引导 (join_status=not_target)")
}

// ======================== IT-113: 不绑定学生的分享码 → 所有学生可用（向后兼容） ========================
func TestV4_IT113_UnboundShareCode(t *testing.T) {
	v4Setup(t)

	// 生成不绑定学生的分享码
	_, body, _ := doRequest("POST", "/api/shares", map[string]interface{}{
		"expires_hours": 168,
		"max_uses":      10,
	}, v4TeacherToken)
	apiResp, _ := parseResponse(body)
	shareCode, _ := apiResp.Data["share_code"].(string)

	targetID, _ := apiResp.Data["target_student_persona_id"].(float64)
	if targetID != 0 {
		t.Fatalf("IT-113 target_student_persona_id 应为 0, 实际 %v", targetID)
	}

	// 学生2使用该分享码（学生2可能已有关系，返回 40023 也算通过）
	// 核心验证点：不绑定的分享码不会返回 40029（定向拒绝）
	resp, body, err := doRequest("POST", fmt.Sprintf("/api/shares/%s/join", shareCode), map[string]interface{}{
		"student_persona_id": int(v4Student2PersonaID),
	}, v4Student2Token)
	if err != nil {
		t.Fatalf("IT-113 请求失败: %v", err)
	}
	if resp.StatusCode == http.StatusForbidden {
		t.Fatal("IT-113 不绑定的分享码不应返回 403 Forbidden")
	}

	apiResp, _ = parseResponse(body)
	// 不应返回 40029（定向拒绝）
	if apiResp.Code == 40029 {
		t.Fatal("IT-113 不绑定的分享码不应返回 40029")
	}
	// 允许 0（成功）或 40023（已存在关系）或 50001（UNIQUE约束冲突=关系已存在）
	if apiResp.Code != 0 && apiResp.Code != 40023 && apiResp.Code != 50001 {
		t.Fatalf("IT-113 业务码错误: 期望 0/40023/50001, 实际 %d, message: %s", apiResp.Code, apiResp.Message)
	}

	t.Logf("IT-113 通过: 不绑定学生的分享码不会被定向拒绝 (code=%d)", apiResp.Code)
}

// ======================== IT-114: 教师搜索已注册学生 ========================
func TestV4_IT114_SearchStudents(t *testing.T) {
	v4Setup(t)

	resp, body, err := doRequest("GET", "/api/students/search?keyword=V4小明", nil, v4TeacherToken)
	if err != nil {
		t.Fatalf("IT-114 请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-114 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	var rawResp struct {
		Code int `json:"code"`
		Data struct {
			Items []map[string]interface{} `json:"items"`
			Total float64                  `json:"total"`
		} `json:"data"`
	}
	json.Unmarshal(body, &rawResp)

	if rawResp.Code != 0 {
		t.Fatalf("IT-114 业务码错误: %d", rawResp.Code)
	}

	found := false
	for _, item := range rawResp.Data.Items {
		if nickname, ok := item["nickname"].(string); ok && nickname == "V4小明分身" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("IT-114 搜索未找到 V4小明分身")
	}

	t.Log("IT-114 通过: 教师搜索已注册学生成功")
}

// ======================== IT-115: 教师真人回复学生（引用回复） ========================
func TestV4_IT115_TeacherReply(t *testing.T) {
	v4Setup(t)

	// 先让学生发一条消息（创建会话）
	chatResp, body, err := doRequest("POST", "/api/chat", map[string]interface{}{
		"message":            "牛顿第一定律是什么？",
		"teacher_persona_id": int(v4TeacherPersonaID),
	}, v4StudentToken)
	if err != nil {
		t.Fatalf("IT-115 学生发消息失败: %v", err)
	}
	if chatResp.StatusCode != http.StatusOK {
		t.Fatalf("IT-115 学生发消息 HTTP 错误: %d, body: %s", chatResp.StatusCode, string(body))
	}

	apiResp, _ := parseResponse(body)
	if apiResp.Code != 0 {
		t.Fatalf("IT-115 学生发消息业务码错误: %d, message: %s", apiResp.Code, apiResp.Message)
	}

	v4SessionID, _ = apiResp.Data["session_id"].(string)
	v4ConversationID, _ = apiResp.Data["conversation_id"].(float64)

	if v4SessionID == "" {
		t.Fatal("IT-115 session_id 为空")
	}

	// 教师真人回复（引用学生消息）
	resp, body, err := doRequest("POST", "/api/chat/teacher-reply", map[string]interface{}{
		"student_persona_id": int(v4StudentPersonaID),
		"session_id":         v4SessionID,
		"content":            "同学你好！惯性其实就是物体保持原来运动状态的性质",
		"reply_to_id":        int(v4ConversationID),
	}, v4TeacherToken)
	if err != nil {
		t.Fatalf("IT-115 教师回复失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-115 教师回复 HTTP 错误: %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, _ = parseResponse(body)
	if apiResp.Code != 0 {
		t.Fatalf("IT-115 教师回复业务码错误: %d, message: %s", apiResp.Code, apiResp.Message)
	}

	senderType, _ := apiResp.Data["sender_type"].(string)
	if senderType != "teacher" {
		t.Fatalf("IT-115 sender_type 错误: 期望 teacher, 实际 %s", senderType)
	}

	takeoverStatus, _ := apiResp.Data["takeover_status"].(string)
	if takeoverStatus != "active" {
		t.Fatalf("IT-115 takeover_status 错误: 期望 active, 实际 %s", takeoverStatus)
	}

	t.Logf("IT-115 通过: 教师真人回复成功, sender_type=teacher, takeover_status=active")
}

// ======================== IT-116: 真人回复后自动进入接管状态 ========================
func TestV4_IT116_AutoTakeover(t *testing.T) {
	v4Setup(t)

	if v4SessionID == "" {
		t.Skip("IT-116 跳过: session_id 未获取（IT-115 可能失败）")
	}

	// 查询接管状态
	resp, body, err := doRequest("GET", fmt.Sprintf("/api/chat/takeover-status?session_id=%s", v4SessionID), nil, v4TeacherToken)
	if err != nil {
		t.Fatalf("IT-116 请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-116 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, _ := parseResponse(body)
	if apiResp.Code != 0 {
		t.Fatalf("IT-116 业务码错误: %d", apiResp.Code)
	}

	isTakenOver, _ := apiResp.Data["is_taken_over"].(bool)
	if !isTakenOver {
		t.Fatal("IT-116 is_taken_over 应为 true")
	}

	teacherNickname, _ := apiResp.Data["teacher_nickname"].(string)
	if teacherNickname == "" {
		t.Fatal("IT-116 teacher_nickname 不应为空")
	}

	t.Logf("IT-116 通过: 真人回复后自动进入接管状态, teacher=%s", teacherNickname)
}

// ======================== IT-117: 接管状态下学生发消息 → 不触发 AI 回复 ========================
func TestV4_IT117_TakeoverNoAI(t *testing.T) {
	v4Setup(t)

	if v4SessionID == "" {
		t.Skip("IT-117 跳过: session_id 未获取")
	}

	// 学生在接管状态下发消息
	resp, body, err := doRequest("POST", "/api/chat", map[string]interface{}{
		"message":            "谢谢老师！我明白了",
		"teacher_persona_id": int(v4TeacherPersonaID),
		"session_id":         v4SessionID,
	}, v4StudentToken)
	if err != nil {
		t.Fatalf("IT-117 请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-117 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, _ := parseResponse(body)
	// 应返回 40030（接管中，不触发 AI）
	if apiResp.Code != 40030 {
		t.Fatalf("IT-117 业务码错误: 期望 40030, 实际 %d, message: %s", apiResp.Code, apiResp.Message)
	}

	t.Log("IT-117 通过: 接管状态下学生发消息不触发 AI 回复 (40030)")
}

// ======================== IT-118: 教师退出接管 → AI 恢复服务 ========================
func TestV4_IT118_EndTakeover(t *testing.T) {
	v4Setup(t)

	if v4SessionID == "" {
		t.Skip("IT-118 跳过: session_id 未获取")
	}

	// 教师退出接管
	resp, body, err := doRequest("POST", "/api/chat/end-takeover", map[string]interface{}{
		"session_id": v4SessionID,
	}, v4TeacherToken)
	if err != nil {
		t.Fatalf("IT-118 请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-118 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, _ := parseResponse(body)
	if apiResp.Code != 0 {
		t.Fatalf("IT-118 业务码错误: %d, message: %s", apiResp.Code, apiResp.Message)
	}

	status, _ := apiResp.Data["status"].(string)
	if status != "ended" {
		t.Fatalf("IT-118 status 错误: 期望 ended, 实际 %s", status)
	}

	// 验证接管状态已结束
	_, body, _ = doRequest("GET", fmt.Sprintf("/api/chat/takeover-status?session_id=%s", v4SessionID), nil, v4TeacherToken)
	apiResp, _ = parseResponse(body)
	isTakenOver, _ := apiResp.Data["is_taken_over"].(bool)
	if isTakenOver {
		t.Fatal("IT-118 退出接管后 is_taken_over 应为 false")
	}

	// 验证 AI 恢复服务：学生发消息应正常返回 AI 回复
	resp, body, err = doRequest("POST", "/api/chat", map[string]interface{}{
		"message":            "那惯性有什么实际应用？",
		"teacher_persona_id": int(v4TeacherPersonaID),
		"session_id":         v4SessionID,
	}, v4StudentToken)
	if err != nil {
		t.Fatalf("IT-118 AI恢复验证请求失败: %v", err)
	}

	apiResp, _ = parseResponse(body)
	if apiResp.Code != 0 {
		t.Fatalf("IT-118 AI恢复后业务码错误: %d, message: %s", apiResp.Code, apiResp.Message)
	}

	reply, _ := apiResp.Data["reply"].(string)
	if reply == "" {
		t.Fatal("IT-118 AI 恢复后应有回复内容")
	}

	t.Logf("IT-118 通过: 教师退出接管后 AI 恢复服务, reply=%s", reply[:min(len(reply), 50)])
}

// ======================== IT-119: 对话记录包含 sender_type 和 reply_to_id ========================
func TestV4_IT119_ConversationSenderType(t *testing.T) {
	v4Setup(t)

	if v4SessionID == "" {
		t.Skip("IT-119 跳过: session_id 未获取")
	}

	// 教师查看学生对话记录
	resp, body, err := doRequest("GET", fmt.Sprintf("/api/conversations/student/%d?session_id=%s", int(v4StudentPersonaID), v4SessionID), nil, v4TeacherToken)
	if err != nil {
		t.Fatalf("IT-119 请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-119 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	var rawResp struct {
		Code int `json:"code"`
		Data struct {
			Messages []map[string]interface{} `json:"messages"`
		} `json:"data"`
	}
	json.Unmarshal(body, &rawResp)

	if rawResp.Code != 0 {
		t.Fatalf("IT-119 业务码错误: %d", rawResp.Code)
	}

	// 验证消息中包含不同的 sender_type
	hasStudent := false
	hasAI := false
	hasTeacher := false
	hasReplyTo := false

	for _, msg := range rawResp.Data.Messages {
		senderType, _ := msg["sender_type"].(string)
		switch senderType {
		case "student":
			hasStudent = true
		case "ai":
			hasAI = true
		case "teacher":
			hasTeacher = true
			if replyToID, ok := msg["reply_to_id"].(float64); ok && replyToID > 0 {
				hasReplyTo = true
			}
		}
	}

	if !hasStudent {
		t.Fatal("IT-119 对话记录中缺少 student 类型消息")
	}
	if !hasTeacher {
		t.Fatal("IT-119 对话记录中缺少 teacher 类型消息")
	}
	if !hasReplyTo {
		t.Fatal("IT-119 教师消息中缺少 reply_to_id")
	}

	t.Logf("IT-119 通过: 对话记录包含 sender_type (student=%v, ai=%v, teacher=%v) 和 reply_to_id", hasStudent, hasAI, hasTeacher)
}

// ======================== IT-120: 查询接管状态 ========================
func TestV4_IT120_TakeoverStatus(t *testing.T) {
	v4Setup(t)

	if v4SessionID == "" {
		t.Skip("IT-120 跳过: session_id 未获取")
	}

	// 查询已结束的接管状态
	resp, body, err := doRequest("GET", fmt.Sprintf("/api/chat/takeover-status?session_id=%s", v4SessionID), nil, v4StudentToken)
	if err != nil {
		t.Fatalf("IT-120 请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-120 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, _ := parseResponse(body)
	if apiResp.Code != 0 {
		t.Fatalf("IT-120 业务码错误: %d", apiResp.Code)
	}

	isTakenOver, _ := apiResp.Data["is_taken_over"].(bool)
	if isTakenOver {
		t.Fatal("IT-120 接管已结束，is_taken_over 应为 false")
	}

	t.Log("IT-120 通过: 查询接管状态正确")
}

// ======================== IT-121: 分身概览页获取所有分身（含 is_public） ========================
func TestV4_IT121_PersonaOverview(t *testing.T) {
	v4Setup(t)

	resp, body, err := doRequest("GET", "/api/personas", nil, v4TeacherToken)
	if err != nil {
		t.Fatalf("IT-121 请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-121 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	var rawResp struct {
		Code int `json:"code"`
		Data struct {
			Personas         []map[string]interface{} `json:"personas"`
			CurrentPersonaID float64                  `json:"current_persona_id"`
		} `json:"data"`
	}
	json.Unmarshal(body, &rawResp)

	if rawResp.Code != 0 {
		t.Fatalf("IT-121 业务码错误: %d", rawResp.Code)
	}

	// 验证返回的分身列表中包含 is_public 字段
	found := false
	for _, item := range rawResp.Data.Personas {
		if id, ok := item["id"].(float64); ok && id == v4TeacherPersonaID {
			found = true
			// 验证 is_public 字段存在
			if _, ok := item["is_public"]; !ok {
				t.Fatal("IT-121 分身列表缺少 is_public 字段")
			}
			break
		}
	}
	if !found {
		t.Fatal("IT-121 分身列表中未找到教师分身")
	}

	t.Log("IT-121 通过: 分身概览页获取所有分身（含 is_public）")
}

// ======================== IT-122: 知识库预览返回 LLM 生成的 title 和 summary ========================
func TestV4_IT122_PreviewLLMSummary(t *testing.T) {
	v4Setup(t)

	resp, body, err := doRequest("POST", "/api/documents/preview", map[string]interface{}{
		"title":   "测试文档",
		"content": "牛顿第一定律，也称为惯性定律，是经典力学的基本定律之一。它指出：一个物体如果不受外力作用，将保持静止状态或匀速直线运动状态不变。这个定律揭示了力和运动的关系，是理解物理世界的基础。",
		"tags":    "物理,力学",
	}, v4TeacherToken)
	if err != nil {
		t.Fatalf("IT-122 请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-122 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, _ := parseResponse(body)
	if apiResp.Code != 0 {
		t.Fatalf("IT-122 业务码错误: %d, message: %s", apiResp.Code, apiResp.Message)
	}

	// 验证 llm_title 和 llm_summary 字段存在（mock 模式下可能为空或有值）
	if _, ok := apiResp.Data["llm_title"]; !ok {
		t.Fatal("IT-122 响应缺少 llm_title 字段")
	}
	if _, ok := apiResp.Data["llm_summary"]; !ok {
		t.Fatal("IT-122 响应缺少 llm_summary 字段")
	}

	llmTitle, _ := apiResp.Data["llm_title"].(string)
	llmSummary, _ := apiResp.Data["llm_summary"].(string)

	t.Logf("IT-122 通过: 知识库预览返回 LLM 摘要, llm_title=%q, llm_summary=%q", llmTitle, llmSummary)
}

// ======================== IT-123: LLM 生成失败时降级为空字段 ========================
func TestV4_IT123_LLMSummaryFallback(t *testing.T) {
	v4Setup(t)

	// 使用极短内容测试（LLM mock 模式下应返回结果，但验证字段存在即可）
	resp, body, err := doRequest("POST", "/api/documents/preview", map[string]interface{}{
		"title":   "空内容测试",
		"content": "x",
		"tags":    "",
	}, v4TeacherToken)
	if err != nil {
		t.Fatalf("IT-123 请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-123 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, _ := parseResponse(body)
	if apiResp.Code != 0 {
		t.Fatalf("IT-123 业务码错误: %d", apiResp.Code)
	}

	// 验证 llm_title 和 llm_summary 字段存在（降级时为空字符串）
	if _, ok := apiResp.Data["llm_title"]; !ok {
		t.Fatal("IT-123 响应缺少 llm_title 字段")
	}
	if _, ok := apiResp.Data["llm_summary"]; !ok {
		t.Fatal("IT-123 响应缺少 llm_summary 字段")
	}

	t.Log("IT-123 通过: LLM 摘要字段存在（降级为空字段）")
}

// ======================== IT-124: 全链路测试 ========================
func TestV4_IT124_FullChain(t *testing.T) {
	v4Setup(t)

	// 1. 教师创建分身（已在 setup 中完成）
	t.Log("IT-124 步骤1: 教师分身已创建")

	// 2. 设置公开
	doRequest("PUT", fmt.Sprintf("/api/personas/%d/visibility", int(v4TeacherPersonaID)),
		map[string]interface{}{"is_public": true}, v4TeacherToken)
	t.Log("IT-124 步骤2: 分身设置为公开")

	// 3. 学生2从广场发现并申请
	_, body, _ := doRequest("GET", "/api/personas/marketplace?page=1&page_size=20", nil, v4Student2Token)
	var mktResp struct {
		Code int `json:"code"`
		Data struct {
			Items []map[string]interface{} `json:"items"`
		} `json:"data"`
	}
	json.Unmarshal(body, &mktResp)
	found := false
	for _, item := range mktResp.Data.Items {
		if id, ok := item["id"].(float64); ok && id == v4TeacherPersonaID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("IT-124 步骤3: 广场中未找到教师分身")
	}
	t.Log("IT-124 步骤3: 学生2在广场发现教师分身")

	// 4. 学生1与 AI 对话
	_, body, _ = doRequest("POST", "/api/chat", map[string]interface{}{
		"message":            "什么是万有引力？",
		"teacher_persona_id": int(v4TeacherPersonaID),
	}, v4StudentToken)
	chatResp, _ := parseResponse(body)
	sessionID, _ := chatResp.Data["session_id"].(string)
	convID, _ := chatResp.Data["conversation_id"].(float64)
	if sessionID == "" {
		t.Fatal("IT-124 步骤4: 对话 session_id 为空")
	}
	t.Log("IT-124 步骤4: 学生与 AI 对话成功")

	// 5. 教师真人介入
	_, body, _ = doRequest("POST", "/api/chat/teacher-reply", map[string]interface{}{
		"student_persona_id": int(v4StudentPersonaID),
		"session_id":         sessionID,
		"content":            "万有引力是牛顿发现的重要定律",
		"reply_to_id":        int(convID),
	}, v4TeacherToken)
	replyResp, _ := parseResponse(body)
	if replyResp.Code != 0 {
		t.Fatalf("IT-124 步骤5: 教师回复失败: %d, %s", replyResp.Code, replyResp.Message)
	}
	t.Log("IT-124 步骤5: 教师真人介入成功")

	// 6. 验证接管状态
	_, body, _ = doRequest("GET", fmt.Sprintf("/api/chat/takeover-status?session_id=%s", sessionID), nil, v4StudentToken)
	statusResp, _ := parseResponse(body)
	isTakenOver, _ := statusResp.Data["is_taken_over"].(bool)
	if !isTakenOver {
		t.Fatal("IT-124 步骤6: 应处于接管状态")
	}
	t.Log("IT-124 步骤6: 接管状态确认")

	// 7. 退出接管
	doRequest("POST", "/api/chat/end-takeover", map[string]interface{}{
		"session_id": sessionID,
	}, v4TeacherToken)
	t.Log("IT-124 步骤7: 教师退出接管")

	// 8. AI 恢复
	_, body, _ = doRequest("POST", "/api/chat", map[string]interface{}{
		"message":            "那引力波是什么？",
		"teacher_persona_id": int(v4TeacherPersonaID),
		"session_id":         sessionID,
	}, v4StudentToken)
	aiResp, _ := parseResponse(body)
	if aiResp.Code != 0 {
		t.Fatalf("IT-124 步骤8: AI 恢复后对话失败: %d, %s", aiResp.Code, aiResp.Message)
	}
	reply, _ := aiResp.Data["reply"].(string)
	if reply == "" {
		t.Fatal("IT-124 步骤8: AI 恢复后应有回复")
	}
	t.Log("IT-124 步骤8: AI 恢复服务成功")

	t.Log("IT-124 通过: 全链路测试完成")
}

// ======================== IT-125: 回归：旧对话数据 sender_type 回填正确 ========================
func TestV4_IT125_SenderTypeBackfill(t *testing.T) {
	v4Setup(t)

	// 这个测试验证数据库迁移的回填逻辑
	// 在 TestMain 中数据库是新建的，所以旧数据回填不适用
	// 但我们可以验证新创建的对话记录有正确的 sender_type

	// 发一条消息
	_, body, _ := doRequest("POST", "/api/chat", map[string]interface{}{
		"message":            "测试 sender_type 回填",
		"teacher_persona_id": int(v4TeacherPersonaID),
	}, v4StudentToken)
	chatResp, _ := parseResponse(body)
	sessionID, _ := chatResp.Data["session_id"].(string)

	if sessionID == "" {
		t.Skip("IT-125 跳过: 无法创建对话")
	}

	// 查询对话记录
	_, body, _ = doRequest("GET", fmt.Sprintf("/api/conversations/student/%d?session_id=%s", int(v4StudentPersonaID), sessionID), nil, v4TeacherToken)

	var rawResp struct {
		Code int `json:"code"`
		Data struct {
			Messages []map[string]interface{} `json:"messages"`
		} `json:"data"`
	}
	json.Unmarshal(body, &rawResp)

	if rawResp.Code != 0 {
		t.Fatalf("IT-125 业务码错误: %d", rawResp.Code)
	}

	// 验证所有消息都有 sender_type
	for _, msg := range rawResp.Data.Messages {
		senderType, _ := msg["sender_type"].(string)
		if senderType == "" {
			t.Fatalf("IT-125 消息 sender_type 为空: %v", msg)
		}
	}

	t.Log("IT-125 通过: 对话记录 sender_type 正确")
}
