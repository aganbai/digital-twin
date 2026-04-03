package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestV2I6_IT306_DialogueStyle(t *testing.T) {
	v2i2Setup(t)

	if v2i2TeacherToken == "" || v2i2StudentPersonaID <= 0 {
		t.Skip("跳过: 前置条件不满足")
	}

	// 1. PUT /api/styles - 设置
	styleBody := map[string]interface{}{
		"teacher_persona_id": int(v2i2TeacherPersonaID),
		"student_persona_id": int(v2i2StudentPersonaID),
		"style_config": map[string]interface{}{
			"temperature":         0.7,
			"guidance_level":      "high",
			"teaching_style":      "explanatory",
			"style_prompt":        "测试风格",
			"max_turns_per_topic": 5,
		},
	}

	resp, body, err := doRequest("PUT", "/api/styles", styleBody, v2i2TeacherToken)
	if err != nil {
		t.Fatalf("PUT /api/styles 请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("PUT /api/styles 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}
	apiResp, _ := parseResponse(body)
	if apiResp.Code != 0 {
		t.Fatalf("PUT /api/styles 业务码错误: %d", apiResp.Code)
	}

	// 2. GET /api/styles - 获取
	getPath := fmt.Sprintf("/api/styles?teacher_persona_id=%d&student_persona_id=%d", int(v2i2TeacherPersonaID), int(v2i2StudentPersonaID))
	resp, body, err = doRequest("GET", getPath, nil, v2i2TeacherToken)
	if err != nil {
		t.Fatalf("GET /api/styles 请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/styles 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	var getResp struct {
		Code int `json:"code"`
		Data struct {
			StyleConfig struct {
				TeachingStyle string `json:"teaching_style"`
				StylePrompt   string `json:"style_prompt"`
			} `json:"style_config"`
		} `json:"data"`
	}
	json.Unmarshal(body, &getResp)
	if getResp.Code != 0 {
		t.Fatalf("GET /api/styles 业务码错误: %d", getResp.Code)
	}
	if getResp.Data.StyleConfig.TeachingStyle != "explanatory" {
		t.Errorf("teaching_style 期望 explanatory, 实际 %s", getResp.Data.StyleConfig.TeachingStyle)
	}
	if getResp.Data.StyleConfig.StylePrompt != "测试风格" {
		t.Errorf("style_prompt 期望 测试风格, 实际 %s", getResp.Data.StyleConfig.StylePrompt)
	}

	// 3. PUT /api/styles - 无效风格测试
	invalidStyleBody := map[string]interface{}{
		"teacher_persona_id": int(v2i2TeacherPersonaID),
		"student_persona_id": int(v2i2StudentPersonaID),
		"style_config": map[string]interface{}{
			"teaching_style": "invalid_style",
		},
	}
	resp, body, err = doRequest("PUT", "/api/styles", invalidStyleBody, v2i2TeacherToken)
	if err != nil {
		t.Fatalf("PUT /api/styles 无效风格 请求失败: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("PUT /api/styles 无效风格 期望 400, 实际 %d, body: %s", resp.StatusCode, string(body))
	}
	apiResp, _ = parseResponse(body)
	if apiResp.Code != 40040 {
		t.Errorf("PUT /api/styles 无效风格业务码 期望 40040, 实际 %d", apiResp.Code)
	}
}
