package api

import (
	"digital-twin/src/harness/manager"
	"digital-twin/src/plugins/auth"

	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter(mgr *manager.HarnessManager) *gin.Engine {
	r := gin.New()

	// 获取 JWT 管理器
	jwtManager := GetJWTManager(mgr)

	// 获取 CORS 配置
	cfg := mgr.GetConfig()
	var corsConfig = cfg.Security.CORS

	// 全局中间件
	r.Use(RecoveryMiddleware(), RequestLogMiddleware(), CORSMiddleware(corsConfig))

	handler := NewHandler(mgr)

	api := r.Group("/api")
	{
		// 认证接口（无需鉴权）
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", handler.HandleRegister)
			authGroup.POST("/login", handler.HandleLogin)
			authGroup.POST("/wx-login", handler.HandleWxLogin)
			authGroup.POST("/refresh", handler.HandleRefresh) // refresh 自行处理 token 验证（支持过期宽限期）
		}

		// 需要鉴权的路由
		authorized := api.Group("")
		if jwtManager != nil {
			authorized.Use(auth.JWTAuthMiddleware(jwtManager))
		}
		{
			authorized.POST("/auth/complete-profile", handler.HandleCompleteProfile)
			authorized.POST("/chat", handler.HandleChat)
			authorized.GET("/conversations/sessions", handler.HandleGetSessions)
			authorized.GET("/conversations", handler.HandleGetConversations)
			authorized.GET("/teachers", handler.HandleGetTeachers)
			authorized.GET("/user/profile", handler.HandleGetUserProfile)
			authorized.POST("/documents", auth.RoleRequired("teacher", "admin"), handler.HandleAddDocument)
			authorized.GET("/documents", auth.RoleRequired("teacher", "admin"), handler.HandleGetDocuments)
			authorized.DELETE("/documents/:id", auth.RoleRequired("teacher", "admin"), handler.HandleDeleteDocument)
			authorized.GET("/memories", handler.HandleGetMemories)
		}

		// 系统接口
		system := api.Group("/system")
		{
			system.GET("/health", handler.HandleHealthCheck)
			if jwtManager != nil {
				system.GET("/plugins", auth.JWTAuthMiddleware(jwtManager), auth.RoleRequired("admin"), handler.HandleGetPlugins)
				system.GET("/pipelines", auth.JWTAuthMiddleware(jwtManager), auth.RoleRequired("admin"), handler.HandleGetPipelines)
			}
		}
	}

	return r
}
