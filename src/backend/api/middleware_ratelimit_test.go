package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"digital-twin/src/harness/config"

	"github.com/gin-gonic/gin"
)

func TestGlobalRateLimitMiddleware_Disabled(t *testing.T) {
	cfg := config.RateLimitingConfig{
		Enabled: false,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(GlobalRateLimitMiddleware(cfg))
	r.GET("/api/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 限流关闭时所有请求应全部通过
	for i := 0; i < 20; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/test", nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("限流未启用时请求应通过，第 %d 次请求返回 %d", i+1, w.Code)
		}
	}
}

func TestGlobalRateLimitMiddleware_Triggered(t *testing.T) {
	cfg := config.RateLimitingConfig{
		Enabled:           true,
		RequestsPerMinute: 180, // 3 req/s
		BurstSize:         3,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(GlobalRateLimitMiddleware(cfg))
	r.GET("/api/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 前3次应通过（突发容量为3）
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/test", nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("前3次请求应通过，第 %d 次请求返回 %d", i+1, w.Code)
		}
	}

	// 第4次应被限流
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/test", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("第4次请求应返回 429，实际返回 %d", w.Code)
	}
}

func TestGlobalRateLimitMiddleware_AuthPathExcluded(t *testing.T) {
	cfg := config.RateLimitingConfig{
		Enabled:           true,
		RequestsPerMinute: 60,
		BurstSize:         1,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(GlobalRateLimitMiddleware(cfg))
	r.POST("/api/auth/login", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 认证接口不受限流影响
	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/auth/login", nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("认证接口不应被限流，第 %d 次请求返回 %d", i+1, w.Code)
		}
	}
}

func TestChatRateLimitMiddleware_Disabled(t *testing.T) {
	cfg := config.RateLimitingConfig{
		Enabled: false,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ChatRateLimitMiddleware(cfg))
	r.POST("/api/chat", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 限流关闭时所有请求应全部通过
	for i := 0; i < 20; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/chat", nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("限流未启用时请求应通过")
		}
	}
}

func TestChatRateLimitMiddleware_Triggered(t *testing.T) {
	cfg := config.RateLimitingConfig{
		Enabled:               true,
		ChatRequestsPerMinute: 60, // 1 req/s
		ChatBurstSize:         2,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ChatRateLimitMiddleware(cfg))
	r.POST("/api/chat", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 前2次应通过
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/chat", nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("前2次请求应通过，第 %d 次请求返回 %d", i+1, w.Code)
		}
	}

	// 第3次应被限流
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/chat", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("第3次请求应返回 429，实际返回 %d", w.Code)
	}
}

func TestTokenBucket_Allow(t *testing.T) {
	tb := NewTokenBucket(3, 1.0) // 容量3，每秒补充1个

	// 消耗3个令牌
	for i := 0; i < 3; i++ {
		if !tb.Allow() {
			t.Fatalf("前3次应该允许")
		}
	}

	// 第4次应该被拒绝
	if tb.Allow() {
		t.Fatal("令牌用完后应该拒绝")
	}

	// 等待1秒后应该恢复1个令牌
	time.Sleep(1100 * time.Millisecond)
	if !tb.Allow() {
		t.Fatal("等待1秒后应该有1个令牌")
	}
}

func TestTokenBucket_Refill(t *testing.T) {
	tb := NewTokenBucket(5, 2.0) // 容量5，每秒补充2个

	// 消耗全部令牌
	for i := 0; i < 5; i++ {
		tb.Allow()
	}

	// 应该被拒绝
	if tb.Allow() {
		t.Fatal("令牌用完后应该拒绝")
	}

	// 等待1秒后应该恢复约2个令牌
	time.Sleep(1100 * time.Millisecond)
	if !tb.Allow() {
		t.Fatal("等待1秒后应该有令牌可用")
	}
	if !tb.Allow() {
		t.Fatal("等待1秒后应该有2个令牌可用")
	}
}

func TestFormatUserKey(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		{"user123", "user123"},
		{int64(42), "42"},
		{int(99), "99"},
		{nil, "unknown"},
		{3.14, "unknown"},
	}

	for _, tt := range tests {
		result := formatUserKey(tt.input)
		if result != tt.expected {
			t.Errorf("formatUserKey(%v) = %s, 期望 %s", tt.input, result, tt.expected)
		}
	}
}

func TestChatRateLimitMiddleware_WithUserID(t *testing.T) {
	cfg := config.RateLimitingConfig{
		Enabled:               true,
		ChatRequestsPerMinute: 60,
		ChatBurstSize:         2,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	// 模拟设置 user_id 的中间件
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "test_user_1")
		c.Next()
	})
	r.Use(ChatRateLimitMiddleware(cfg))
	r.POST("/api/chat", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 用户1的前2次应通过
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/chat", nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("用户1前2次请求应通过，第 %d 次返回 %d", i+1, w.Code)
		}
	}

	// 用户1的第3次应被限流
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/chat", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("用户1第3次请求应返回 429，实际返回 %d", w.Code)
	}
}

func TestRateLimitStore_DifferentKeys(t *testing.T) {
	store := NewRateLimitStore(2, 1.0)

	// 不同 key 应有独立的限流器
	limiter1 := store.getLimiter("user:1")
	limiter2 := store.getLimiter("user:2")

	if limiter1 == limiter2 {
		t.Fatal("不同 key 应返回不同的限流器")
	}

	// 同一 key 应返回相同的限流器
	limiter1Again := store.getLimiter("user:1")
	if limiter1 != limiter1Again {
		t.Fatal("相同 key 应返回相同的限流器")
	}
}

func TestGlobalRateLimit_429ResponseFormat(t *testing.T) {
	cfg := config.RateLimitingConfig{
		Enabled:           true,
		RequestsPerMinute: 60,
		BurstSize:         1,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(GlobalRateLimitMiddleware(cfg))
	r.GET("/api/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 第1次通过
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/test", nil)
	r.ServeHTTP(w, req)

	// 第2次触发限流，验证响应格式
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("应返回 429，实际返回 %d", w.Code)
	}

	body := w.Body.String()
	if !contains(body, "40051") {
		t.Fatalf("响应应包含错误码 40051，实际: %s", body)
	}
	if !contains(body, "请求频率超限") {
		t.Fatalf("响应应包含限流提示信息，实际: %s", body)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
