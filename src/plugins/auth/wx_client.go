package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// WxClient 微信 API 客户端接口
type WxClient interface {
	Code2Session(code string) (*WxSessionResult, error)
	// H5 网页授权接口
	GetAuthURL(redirectURI, state string) string
	GetAccessToken(code string) (*WxAccessTokenResult, error)
	GetUserInfo(accessToken, openid string) (*WxUserInfo, error)
}

// WxSessionResult 微信 jscode2session 返回结果
type WxSessionResult struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// WxAccessTokenResult 微信网页授权 access_token 返回结果
type WxAccessTokenResult struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
}

// WxUserInfo 微信用户信息
type WxUserInfo struct {
	OpenID     string   `json:"openid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	HeadImgURL string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	UnionID    string   `json:"unionid"`
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

// GetAuthURL 生成微信网页授权URL
func (c *RealWxClient) GetAuthURL(redirectURI, state string) string {
	// 微信网页授权 URL（scope=snsapi_userinfo 可获取用户信息）
	encodedRedirectURI := url.QueryEscape(redirectURI)
	return fmt.Sprintf(
		"https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_userinfo&state=%s#wechat_redirect",
		c.AppID, encodedRedirectURI, state,
	)
}

// GetAccessToken 通过 code 获取网页授权 access_token
func (c *RealWxClient) GetAccessToken(code string) (*WxAccessTokenResult, error) {
	apiURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		c.AppID, c.Secret, code,
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("请求微信 API 失败: %w", err)
	}
	defer resp.Body.Close()

	var result WxAccessTokenResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析微信 API 响应失败: %w", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("微信 API 返回错误: errcode=%d, errmsg=%s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}

// GetUserInfo 获取微信用户信息
func (c *RealWxClient) GetUserInfo(accessToken, openid string) (*WxUserInfo, error) {
	apiURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN",
		accessToken, openid,
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("请求微信 API 失败: %w", err)
	}
	defer resp.Body.Close()

	var result WxUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析微信 API 响应失败: %w", err)
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
// 如果 code 以 "v9iter_" 开头（迭代9测试专用），也使用 code 作为区分。
func (m *MockWxClient) Code2Session(code string) (*WxSessionResult, error) {
	openid := "mock_openid_default_user"
	// 支持 "test_user_" 前缀
	if len(code) > 10 && code[:10] == "test_user_" {
		openid = "mock_openid_" + code
	}
	// 支持 "v9iter_" 前缀（迭代9 E2E测试）
	if len(code) > 7 && code[:7] == "v9iter_" {
		openid = "mock_openid_" + code
	}
	// 支持 "mock_" 前缀（通用测试模式）
	if len(code) > 5 && code[:5] == "mock_" {
		openid = "mock_openid_" + code
	}
	return &WxSessionResult{
		OpenID: openid,
	}, nil
}

// GetAuthURL 生成 mock 授权URL
func (m *MockWxClient) GetAuthURL(redirectURI, state string) string {
	return fmt.Sprintf("https://mock.weixin.com/oauth2?redirect_uri=%s&state=%s", redirectURI, state)
}

// GetAccessToken 返回 mock access_token
func (m *MockWxClient) GetAccessToken(code string) (*WxAccessTokenResult, error) {
	// Mock 模式下，使用 code 生成 openid
	openid := "mock_h5_openid_default"
	if len(code) > 8 && code[:8] == "h5_user_" {
		openid = "mock_h5_openid_" + code
	}
	return &WxAccessTokenResult{
		AccessToken: "mock_access_token_" + code,
		OpenID:      openid,
		UnionID:     "mock_unionid_" + openid,
	}, nil
}

// GetUserInfo 返回 mock 用户信息
func (m *MockWxClient) GetUserInfo(accessToken, openid string) (*WxUserInfo, error) {
	return &WxUserInfo{
		OpenID:     openid,
		Nickname:   "Mock用户",
		Sex:        1,
		HeadImgURL: "https://mock.avatar.com/default.png",
	}, nil
}
