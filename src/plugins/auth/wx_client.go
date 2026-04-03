package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WxClient 微信 API 客户端接口
type WxClient interface {
	Code2Session(code string) (*WxSessionResult, error)
}

// WxSessionResult 微信 jscode2session 返回结果
type WxSessionResult struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// RealWxClient 真实微信 API 客户端
type RealWxClient struct {
	AppID  string
	Secret string
}

// Code2Session 调用微信 jscode2session 接口，用 code 换取 openid
func (c *RealWxClient) Code2Session(code string) (*WxSessionResult, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		c.AppID, c.Secret, code,
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求微信 API 失败: %w", err)
	}
	defer resp.Body.Close()

	var result WxSessionResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析微信 API 响应失败: %w", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("微信 API 返回错误: errcode=%d, errmsg=%s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}

// MockWxClient 测试用 mock 客户端
type MockWxClient struct{}

// Code2Session 返回固定的 mock openid
// 微信小程序每次调用 wx.login() 获取的 code 都不同，
// 但同一设备/用户应该对应同一个 openid。
// Mock 模式下使用固定 openid，确保再次登录能识别为同一用户。
// 如果 code 以 "test_user_" 开头，则使用 code 作为区分（方便 E2E 测试模拟多用户）。
func (m *MockWxClient) Code2Session(code string) (*WxSessionResult, error) {
	openid := "mock_openid_default_user"
	if len(code) > 10 && code[:10] == "test_user_" {
		openid = "mock_openid_" + code
	}
	return &WxSessionResult{
		OpenID: openid,
	}, nil
}
