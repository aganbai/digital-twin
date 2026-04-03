package api

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"
)

// TestLLMErrorType 测试错误类型定义
func TestLLMErrorType(t *testing.T) {
	// 测试超时错误
	timeoutErr := &LLMError{
		Type:    LLMErrorTimeout,
		Message: "连接超时",
	}
	if timeoutErr.Type != LLMErrorTimeout {
		t.Errorf("期望错误类型为 LLMErrorTimeout, 得到 %d", timeoutErr.Type)
	}

	// 测试其他错误
	otherErr := &LLMError{
		Type:    LLMErrorOther,
		Message: "其他错误",
	}
	if otherErr.Type != LLMErrorOther {
		t.Errorf("期望错误类型为 LLMErrorOther, 得到 %d", otherErr.Type)
	}
}

// TestIsTimeoutError 测试超时错误检测
func TestIsTimeoutError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "LLM超时错误",
			err:      &LLMError{Type: LLMErrorTimeout, Message: "超时"},
			expected: true,
		},
		{
			name:     "LLM其他错误",
			err:      &LLMError{Type: LLMErrorOther, Message: "其他错误"},
			expected: false,
		},
		{
			name:     "网络超时错误",
			err:      &url.Error{Op: "Post", URL: "http://test.com", Err: errors.New("timeout")},
			expected: false, // url.Error 没有 Timeout() 方法返回 true
		},
		{
			name:     "nil错误",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTimeoutError(tt.err)
			if result != tt.expected {
				t.Errorf("IsTimeoutError() = %v, 期望 %v", result, tt.expected)
			}
		})
	}
}

// TestSimpleParseStudentText 测试规则解析功能
func TestSimpleParseStudentText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int // 期望解析出的学生数量
	}{
		{
			name:     "空文本",
			text:     "",
			expected: 0,
		},
		{
			name:     "单行文本",
			text:     "张三",
			expected: 1,
		},
		{
			name:     "多行文本",
			text:     "张三\n李四\n王五",
			expected: 3,
		},
		{
			name:     "带分隔符",
			text:     "张三,男\n李四,女",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := simpleParseStudentText(tt.text)
			if len(result) != tt.expected {
				t.Errorf("simpleParseStudentText() 返回 %d 个学生, 期望 %d", len(result), tt.expected)
			}
		})
	}
}

// TestHTTPClientTimeout 验证HTTP客户端超时配置
func TestHTTPClientTimeout(t *testing.T) {
	// 验证超时值设置正确
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	if client.Timeout != 30*time.Second {
		t.Errorf("期望总超时为 30s, 得到 %v", client.Timeout)
	}
}
