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
			authorized.GET("/memories", handler.HandleGetMemoriesV2)

			// V2.0 迭代6 记忆管理
			authorized.PUT("/memories/:id", auth.RoleRequired("teacher"), handler.HandleUpdateMemory)
			authorized.DELETE("/memories/:id", auth.RoleRequired("teacher"), handler.HandleDeleteMemory)
			authorized.POST("/memories/summarize", auth.RoleRequired("teacher"), handler.HandleSummarizeMemories)

			// V2.0 迭代1 新增路由

			// 师生关系
			relations := authorized.Group("/relations")
			{
				relations.POST("/invite", auth.RoleRequired("teacher"), handler.HandleInviteStudent)
				relations.POST("/apply", auth.RoleRequired("student"), handler.HandleApplyTeacher)
				relations.PUT("/:id/approve", auth.RoleRequired("teacher"), handler.HandleApproveRelation)
				relations.PUT("/:id/reject", auth.RoleRequired("teacher"), handler.HandleRejectRelation)
				relations.GET("", handler.HandleGetRelations)
			}

			// 评语
			authorized.POST("/comments", auth.RoleRequired("teacher"), handler.HandleCreateComment)
			authorized.GET("/comments", handler.HandleGetComments)

			// 搜索学生（静态路由必须在参数路由 /students/:id 之前注册，避免 Gin 路由冲突）
			authorized.GET("/students/search", auth.RoleRequired("teacher"), handler.HandleSearchStudents)

			// 问答风格
			authorized.PUT("/students/:id/dialogue-style", auth.RoleRequired("teacher"), handler.HandleSetDialogueStyle)
			authorized.GET("/students/:id/dialogue-style", handler.HandleGetDialogueStyle)

			// 作业
			assignments := authorized.Group("/assignments")
			{
				assignments.POST("", auth.RoleRequired("student"), handler.HandleSubmitAssignment)
				assignments.GET("", handler.HandleGetAssignments)
				assignments.GET("/:id", handler.HandleGetAssignmentDetail)
				assignments.POST("/:id/review", auth.RoleRequired("teacher"), handler.HandleReviewAssignment)
				assignments.POST("/:id/ai-review", handler.HandleAIReviewAssignment)
			}

			// 知识库增强
			authorized.POST("/documents/upload", auth.RoleRequired("teacher", "admin"), handler.HandleUploadDocument)
			authorized.POST("/documents/import-url", auth.RoleRequired("teacher", "admin"), handler.HandleImportURL)

			// SSE 流式对话
			authorized.POST("/chat/stream", handler.HandleChatStream)

			// V2.0 迭代2 新增路由

			// 分身管理
			personas := authorized.Group("/personas")
			{
				personas.POST("", handler.HandleCreatePersona)
				personas.GET("", handler.HandleGetPersonas)
				personas.PUT("/:id", handler.HandleEditPersona)
				personas.PUT("/:id/activate", handler.HandleActivatePersona)
				personas.PUT("/:id/deactivate", handler.HandleDeactivatePersona)
				personas.PUT("/:id/switch", handler.HandleSwitchPersona)
			}

			// 班级管理
			classes := authorized.Group("/classes")
			{
				classes.POST("", auth.RoleRequired("teacher"), handler.HandleCreateClass)
				classes.GET("", auth.RoleRequired("teacher"), handler.HandleGetClasses)
				classes.PUT("/:id", auth.RoleRequired("teacher"), handler.HandleUpdateClass)
				classes.DELETE("/:id", auth.RoleRequired("teacher"), handler.HandleDeleteClass)
				classes.GET("/:id/members", auth.RoleRequired("teacher"), handler.HandleGetClassMembers)
				classes.POST("/:id/members", auth.RoleRequired("teacher"), handler.HandleAddClassMember)
				classes.DELETE("/:id/members/:member_id", auth.RoleRequired("teacher"), handler.HandleRemoveClassMember)
			}

			// 分享码管理
			shares := authorized.Group("/shares")
			{
				shares.POST("", auth.RoleRequired("teacher"), handler.HandleCreateShare)
				shares.GET("", auth.RoleRequired("teacher"), handler.HandleGetShares)
				shares.POST("/:code/join", handler.HandleJoinByShare)
				shares.PUT("/:id/deactivate", auth.RoleRequired("teacher"), handler.HandleDeactivateShare)
			}

			// V2.0 迭代3 新增路由

			// 教师仪表盘
			personas.GET("/:id/dashboard", handler.HandleGetPersonaDashboard)

			// 班级启停
			classes.PUT("/:id/toggle", auth.RoleRequired("teacher"), handler.HandleToggleClass)

			// 师生关系启停
			relations.PUT("/:id/toggle", auth.RoleRequired("teacher"), handler.HandleToggleRelation)

			// 知识库预览
			authorized.POST("/documents/preview", auth.RoleRequired("teacher", "admin"), handler.HandlePreviewDocument)
			authorized.POST("/documents/preview-upload", auth.RoleRequired("teacher", "admin"), handler.HandlePreviewUpload)
			authorized.POST("/documents/preview-url", auth.RoleRequired("teacher", "admin"), handler.HandlePreviewURL)
			authorized.POST("/documents/confirm", auth.RoleRequired("teacher", "admin"), handler.HandleConfirmDocument)

			// V2.0 迭代5 新增路由

			// 通用文件上传
			authorized.POST("/upload", handler.HandleUpload)

			// V2.0 迭代4 新增路由

			// 分身广场
			personas.GET("/marketplace", handler.HandleGetMarketplace)

			// 分身公开设置
			personas.PUT("/:id/visibility", handler.HandleSetVisibility)

			// 教师真人介入对话
			authorized.POST("/chat/teacher-reply", auth.RoleRequired("teacher"), handler.HandleTeacherReply)
			authorized.GET("/chat/takeover-status", handler.HandleGetTakeoverStatus)
			authorized.POST("/chat/end-takeover", auth.RoleRequired("teacher"), handler.HandleEndTakeover)
			authorized.GET("/conversations/student/:student_persona_id", auth.RoleRequired("teacher"), handler.HandleGetStudentConversations)
			// V2.0 迭代6 新增路由

			// 聊天记录导入
			authorized.POST("/documents/import-chat", auth.RoleRequired("teacher", "admin"), handler.HandleImportChat)
		}

		// 分享码公开接口（可选鉴权）
		sharePublic := api.Group("/shares")
		if jwtManager != nil {
			sharePublic.Use(auth.OptionalJWTAuthMiddleware(jwtManager))
		}
		{
			sharePublic.GET("/:code/info", handler.HandleGetShareInfoV2)
		}
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
