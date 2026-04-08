package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware JWT 认证中间件
// 从 Authorization: Bearer <token> 提取 token，验证后将 user_id, username, role 存入 gin.Context
func JWTAuthMiddleware(jwtManager *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40001,
				"message": "未提供认证令牌",
			})
			c.Abort()
			return
		}

		// 提取 Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40001,
				"message": "认证令牌格式无效",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 验证 token
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			// 判断是否为过期错误
			if strings.Contains(err.Error(), "expired") {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    40002,
					"message": "令牌已过期",
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    40001,
					"message": "令牌无效",
				})
			}
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("persona_id", claims.PersonaID) // V2.0 迭代2：分身ID
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		// 设置用户角色（如果存在），用于admin权限检查
		if claims.UserRole != "" {
			c.Set("user_role", claims.UserRole)
		} else {
			c.Set("user_role", claims.Role) // 兼容旧token
		}

		c.Next()
	}
}

// OptionalJWTAuthMiddleware 可选 JWT 认证中间件
// 有 Token 时验证并注入用户信息，无 Token 时不中断请求
func OptionalJWTAuthMiddleware(jwtManager *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 无 Token，继续处理
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			// Token 格式无效，继续处理（不中断）
			c.Next()
			return
		}

		tokenString := parts[1]
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			// Token 无效，继续处理（不中断）
			c.Next()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("persona_id", claims.PersonaID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("authenticated", true)

		c.Next()
	}
}

// RoleRequired 角色权限校验中间件
func RoleRequired(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先检查用户角色（UserRole），用于admin等用户级别权限
		userRole, hasUserRole := c.Get("user_role")
		if hasUserRole {
			userRoleStr, ok := userRole.(string)
			if ok {
				for _, allowedRole := range roles {
					if userRoleStr == allowedRole {
						c.Next()
						return
					}
				}
			}
		}

		// 其次检查分身角色（Role）
		personaRole, hasRole := c.Get("role")
		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    40003,
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		roleStr, ok := personaRole.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    40003,
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		// 检查用户角色是否在允许的角色列表中
		for _, allowedRole := range roles {
			if roleStr == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"code":    40003,
			"message": "权限不足",
		})
		c.Abort()
	}
}
