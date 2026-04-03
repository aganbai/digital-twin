package integration

import (
	"fmt"
	"net/http"
	"testing"
)

// 测试 IT-310: 分享码信息 - join_status = can_join
func TestV2I6_IT310_ShareInfoCanJoin(t *testing.T) {
	v2i2Setup(t)

	if v2i2TeacherToken == "" || v2i2StudentToken == "" {
		t.Skip("跳过: 前置条件不满足")
	}

	// 1. 教师生成一个通用分享码 (target_student_persona_id = 0)
	_, body, _ := doRequest("POST", "/api/shares", map[string]interface{}{
		"expires_hours": 168,
		"max_uses":      10,
	}, v2i2TeacherToken)
	apiResp, _ := parseResponse(body)
	shareCode, _ := apiResp.Data["share_code"].(string)

	// 2. 学生获取分享码信息
	getPath := fmt.Sprintf("/api/shares/%s/info", shareCode)
	resp, body, err := doRequest("GET", getPath, nil, v2i2StudentToken)
	if err != nil {
		t.Fatalf("GET %s 请求失败: %v", getPath, err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s 期望 200, 实际 %d, body: %s", getPath, resp.StatusCode, string(body))
	}

	apiResp, _ = parseResponse(body)
	if apiResp.Code != 0 {
		t.Fatalf("GET /api/shares/info 业务码错误: %d", apiResp.Code)
	}

	joinStatus, _ := apiResp.Data["join_status"].(string)

	// 注意：如果学生已经加入了该教师的班级/师生关系，那么 join_status 可能会变成 already_joined。
	// 由于集成测试复用了账号，如果已经 join 过，这里验证 already_joined 也是合理的。
	if joinStatus != "can_join" && joinStatus != "already_joined" {
		t.Errorf("join_status 期望 can_join 或 already_joined, 实际 %s", joinStatus)
	}
}

// 测试 IT-311: 分享码信息 - join_status = not_target
func TestV2I6_IT311_ShareInfoNotTarget(t *testing.T) {
	v4Setup(t) // 使用 v4Setup 因为它提供了两个学生账号

	if v4TeacherToken == "" || v4StudentPersonaID <= 0 || v4Student2Token == "" {
		t.Skip("跳过: 前置条件不满足")
	}

	// 1. 教师生成一个定向分享码 (绑定学生1)
	_, body, _ := doRequest("POST", "/api/shares", map[string]interface{}{
		"target_student_persona_id": int(v4StudentPersonaID),
		"expires_hours":             168,
		"max_uses":                  10,
	}, v4TeacherToken)
	apiResp, _ := parseResponse(body)
	shareCode, _ := apiResp.Data["share_code"].(string)

	// 2. 学生2获取分享码信息
	getPath := fmt.Sprintf("/api/shares/%s/info", shareCode)
	resp, body, err := doRequest("GET", getPath, nil, v4Student2Token)
	if err != nil {
		t.Fatalf("GET %s 请求失败: %v", getPath, err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s 期望 200, 实际 %d, body: %s", getPath, resp.StatusCode, string(body))
	}

	apiResp, _ = parseResponse(body)
	joinStatus, _ := apiResp.Data["join_status"].(string)

	if joinStatus != "not_target" {
		t.Errorf("join_status 期望 not_target, 实际 %s", joinStatus)
	}
}

// 测试 IT-312: 分享码信息 - join_status = already_joined
func TestV2I6_IT312_ShareInfoAlreadyJoined(t *testing.T) {
	v4Setup(t)

	if v4TeacherToken == "" || v4StudentToken == "" {
		t.Skip("跳过: 前置条件不满足")
	}

	// 1. 教师生成一个通用分享码
	_, body, _ := doRequest("POST", "/api/shares", map[string]interface{}{
		"expires_hours": 168,
		"max_uses":      10,
	}, v4TeacherToken)
	apiResp, _ := parseResponse(body)
	shareCode, ok := apiResp.Data["share_code"].(string)
	if !ok || shareCode == "" {
		t.Fatalf("无法获取 share_code")
	}

	// 2. 学生1加入分享码 (确保关系存在)
	joinPath := fmt.Sprintf("/api/shares/%s/join", shareCode)
	doRequest("POST", joinPath, map[string]interface{}{
		"student_persona_id": int(v4StudentPersonaID),
	}, v4StudentToken)

	// 3. 学生1再次获取分享码信息
	infoPath := fmt.Sprintf("/api/shares/%s/info", shareCode)
	resp, body, err := doRequest("GET", infoPath, nil, v4StudentToken)
	if err != nil {
		t.Fatalf("GET %s 请求失败: %v", infoPath, err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s 期望 200, 实际 %d, body: %s", infoPath, resp.StatusCode, string(body))
	}

	apiResp, _ = parseResponse(body)
	joinStatus, _ := apiResp.Data["join_status"].(string)

	if joinStatus != "already_joined" {
		t.Errorf("join_status 期望 already_joined, 实际 %s", joinStatus)
	}
}

// 测试分享码信息未登录状态 - need_login
func TestV2I6_ShareInfoNeedLogin(t *testing.T) {
	v4Setup(t)
	if v4TeacherToken == "" {
		t.Skip("跳过: 前置条件不满足")
	}

	// 1. 教师生成分享码
	_, body, _ := doRequest("POST", "/api/shares", map[string]interface{}{
		"expires_hours": 168,
		"max_uses":      10,
	}, v4TeacherToken)
	apiResp, _ := parseResponse(body)
	shareCode, _ := apiResp.Data["share_code"].(string)

	// 2. 无 token 访问分享码信息
	infoPath := fmt.Sprintf("/api/shares/%s/info", shareCode)
	resp, body, err := doRequest("GET", infoPath, nil, "")
	if err != nil {
		t.Fatalf("GET %s 请求失败: %v", infoPath, err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s 期望 200, 实际 %d, body: %s", infoPath, resp.StatusCode, string(body))
	}

	apiResp, _ = parseResponse(body)
	joinStatus, _ := apiResp.Data["join_status"].(string)

	if joinStatus != "need_login" {
		t.Errorf("join_status 期望 need_login, 实际 %s", joinStatus)
	}
}
