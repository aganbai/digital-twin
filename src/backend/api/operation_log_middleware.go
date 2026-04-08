package api

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"digital-twin/src/backend/database"

	"github.com/gin-gonic/gin"
)

// OperationLogMiddleware 操作日志中间件
func OperationLogMiddleware(logDB *database.LogDatabase) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只记录需要的方法
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		start := time.Now()

		// 读取请求体（需要恢复）
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 使用 ResponseWriter 包装器捕获状态码
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		// 计算耗时
		duration := time.Since(start)

		// 提取用户信息
		userID := int64(0)
		userRole := ""
		personaID := int64(0)
		if userIDVal, exists := c.Get("user_id"); exists {
			switch v := userIDVal.(type) {
			case int64:
				userID = v
			case float64:
				userID = int64(v)
			}
		}
		if roleVal, exists := c.Get("role"); exists {
			userRole, _ = roleVal.(string)
		}
		if personaIDVal, exists := c.Get("persona_id"); exists {
			switch v := personaIDVal.(type) {
			case int64:
				personaID = v
			case float64:
				personaID = int64(v)
			}
		}

		// 映射 action
		action := mapAction(c.Request.Method, c.Request.URL.Path)

		// 提取资源信息
		resource, resourceID := extractResource(c.Request.URL.Path)

		// 异步写入日志
		go func() {
			logDB.DB.Exec(`
				INSERT INTO operation_logs (
					user_id, user_role, persona_id, action, resource, resource_id,
					detail, ip, user_agent, platform, status_code, duration_ms
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`,
				userID, userRole, personaID, action, resource, resourceID,
				string(requestBody), c.ClientIP(), c.Request.UserAgent(),
				c.GetHeader("X-Platform"), blw.status, duration.Milliseconds(),
			)
		}()
	}
}

// bodyLogWriter 响应写入器包装
type bodyLogWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyLogWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// mapAction 根据请求方法和路径映射 action
func mapAction(method, path string) string {
	// 参考需求文档中的 action 映射表
	switch {
	case strings.HasPrefix(path, "/api/auth/wx-login"):
		return "user.login"
	case strings.HasPrefix(path, "/api/auth/wx-h5-callback"):
		return "user.login"
	case strings.HasPrefix(path, "/api/auth/register"):
		return "user.register"
	case strings.HasPrefix(path, "/api/auth/complete-profile"):
		return "user.profile_update"
	case strings.HasPrefix(path, "/api/chat"):
		return "chat.send_message"
	case strings.HasPrefix(path, "/api/classes"):
		if method == http.MethodPost {
			return "class.create"
		} else if method == http.MethodPut {
			return "class.update"
		} else if method == http.MethodDelete {
			return "class.delete"
		}
		return "class.read"
	case strings.HasPrefix(path, "/api/personas"):
		if method == http.MethodPost {
			return "persona.create"
		} else if method == http.MethodPut {
			return "persona.update"
		} else if method == http.MethodDelete {
			return "persona.delete"
		}
		return "persona.read"
	case strings.HasPrefix(path, "/api/knowledge"):
		if method == http.MethodPost {
			return "knowledge.create"
		} else if method == http.MethodPut {
			return "knowledge.update"
		} else if method == http.MethodDelete {
			return "knowledge.delete"
		}
		return "knowledge.read"
	case strings.HasPrefix(path, "/api/admin"):
		return "admin.operation"
	case strings.HasPrefix(path, "/api/upload"):
		return "file.upload"
	case strings.HasPrefix(path, "/api/feedback"):
		if method == http.MethodPost {
			return "feedback.submit"
		}
		return "feedback.read"
	default:
		if method == http.MethodGet {
			return "api.read"
		}
		return "api.write"
	}
}

// extractResource 提取资源类型和ID
func extractResource(path string) (string, string) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 2 {
		// 例如 /api/users/123 -> resource: users, resource_id: 123
		if len(parts) >= 3 {
			return parts[1], parts[2]
		}
		return parts[1], ""
	}
	return "", ""
}
