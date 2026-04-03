package api

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"digital-twin/src/harness/config"

	"github.com/gin-gonic/gin"
)

// TokenBucket 令牌桶限流器
type TokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64 // 每秒补充的令牌数
	lastRefill time.Time
	nowFunc    func() time.Time // 可注入的时间函数，便于测试
}

// NewTokenBucket 创建令牌桶
func NewTokenBucket(maxTokens float64, refillRate float64) *TokenBucket {
	return &TokenBucket{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
		nowFunc:    time.Now,
	}
}

// Allow 尝试消耗一个令牌，返回是否允许
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := tb.nowFunc()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * tb.refillRate
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}
	tb.lastRefill = now

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

// RateLimitStore 存储每个IP/用户的限流器
type RateLimitStore struct {
	mu         sync.RWMutex
	limiters   map[string]*TokenBucket
	maxTokens  float64
	refillRate float64
}

// NewRateLimitStore 创建限流存储
func NewRateLimitStore(maxTokens float64, refillRate float64) *RateLimitStore {
	store := &RateLimitStore{
		limiters:   make(map[string]*TokenBucket),
		maxTokens:  maxTokens,
		refillRate: refillRate,
	}
	// 定期清理过期的限流器
	go store.cleanup()
	return store
}

// getLimiter 获取或创建限流器
func (s *RateLimitStore) getLimiter(key string) *TokenBucket {
	s.mu.RLock()
	limiter, exists := s.limiters[key]
	s.mu.RUnlock()

	if exists {
		return limiter
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 双重检查
	if limiter, exists = s.limiters[key]; exists {
		return limiter
	}

	limiter = NewTokenBucket(s.maxTokens, s.refillRate)
	s.limiters[key] = limiter
	return limiter
}

// cleanup 定期清理不活跃的限流器（每10分钟）
func (s *RateLimitStore) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for key, limiter := range s.limiters {
			limiter.mu.Lock()
			if now.Sub(limiter.lastRefill) > 30*time.Minute {
				delete(s.limiters, key)
			}
			limiter.mu.Unlock()
		}
		s.mu.Unlock()
	}
}

// Reset 清空所有限流器状态（仅供测试使用）
func (s *RateLimitStore) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.limiters = make(map[string]*TokenBucket)
}

// GlobalRateLimitMiddleware 全局请求限流中间件
// 基于客户端IP进行限流，默认 100 req/s per IP
func GlobalRateLimitMiddleware(cfg config.RateLimitingConfig) gin.HandlerFunc {
	// 支持环境变量禁用限流（测试环境使用）
	if os.Getenv("RATE_LIMIT_DISABLED") == "true" {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	if !cfg.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// 全局限流参数
	burstSize := float64(cfg.BurstSize)
	if burstSize <= 0 {
		burstSize = 20
	}
	refillRate := float64(cfg.RequestsPerMinute) / 60.0
	if refillRate <= 0 {
		refillRate = 100.0 / 1.0 // 默认 100 req/s
	}

	store := NewRateLimitStore(burstSize, refillRate)

	return func(c *gin.Context) {
		// 认证相关接口和系统接口不限流
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/auth/") || strings.HasPrefix(path, "/api/system/") {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		key := "global:" + clientIP

		limiter := store.getLimiter(key)
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    40051,
				"message": "请求频率超限，请稍后重试",
				"data":    nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ChatRateLimitMiddleware 对话接口单独限流中间件
// 基于用户ID进行更严格的限流，默认 10 req/min per user
func ChatRateLimitMiddleware(cfg config.RateLimitingConfig) gin.HandlerFunc {
	// 支持环境变量禁用限流（测试环境使用）
	if os.Getenv("RATE_LIMIT_DISABLED") == "true" {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	if !cfg.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// 对话限流参数
	chatBurst := float64(cfg.ChatBurstSize)
	if chatBurst <= 0 {
		chatBurst = 5
	}
	chatRefillRate := float64(cfg.ChatRequestsPerMinute) / 60.0
	if chatRefillRate <= 0 {
		chatRefillRate = 10.0 / 60.0 // 默认 10 req/min
	}

	store := NewRateLimitStore(chatBurst, chatRefillRate)

	return func(c *gin.Context) {
		// 优先使用用户ID，回退到IP
		key := "chat:"
		if userID, exists := c.Get("user_id"); exists {
			key += formatUserKey(userID)
		} else {
			key += c.ClientIP()
		}

		limiter := store.getLimiter(key)
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    40051,
				"message": "对话请求频率超限，请稍后重试",
				"data":    nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// formatUserKey 将用户ID格式化为字符串key
func formatUserKey(userID interface{}) string {
	switch v := userID.(type) {
	case int64:
		return fmt.Sprintf("%d", v)
	case int:
		return fmt.Sprintf("%d", v)
	case string:
		return v
	default:
		return "unknown"
	}
}
