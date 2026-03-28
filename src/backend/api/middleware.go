package api

import (
	"log"
	"net/http"
	"time"

	"digital-twin/src/harness/config"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware CORS 跨域中间件
func CORSMiddleware(corsConfig config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// 检查 origin 是否在允许列表中
		allowed := false
		for _, o := range corsConfig.AllowedOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed && origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		// 设置允许的方法
		methods := "GET, POST, PUT, DELETE, OPTIONS"
		if len(corsConfig.AllowedMethods) > 0 {
			methods = ""
			for i, m := range corsConfig.AllowedMethods {
				if i > 0 {
					methods += ", "
				}
				methods += m
			}
		}
		c.Header("Access-Control-Allow-Methods", methods)

		// 设置允许的头
		headers := "Content-Type, Authorization"
		if len(corsConfig.AllowedHeaders) > 0 {
			headers = ""
			for i, h := range corsConfig.AllowedHeaders {
				if i > 0 {
					headers += ", "
				}
				headers += h
			}
		}
		c.Header("Access-Control-Allow-Headers", headers)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// 处理预检请求
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestLogMiddleware 请求日志中间件
func RequestLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		log.Printf("[API] %s %s %d %v", method, path, statusCode, duration)
	}
}

// RecoveryMiddleware panic 恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC] %v", r)
				Error(c, http.StatusInternalServerError, 50000, "服务器内部错误")
				c.Abort()
			}
		}()
		c.Next()
	}
}
