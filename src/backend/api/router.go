package api

import (
	"digital-twin/src/backend/database"
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

	// V2.0 迭代7 M6: 全局API限流中间件
	if cfg.Performance.RateLimiting.Enabled {
		r.Use(GlobalRateLimitMiddleware(cfg.Performance.RateLimiting))
	}

	handler := NewHandler(mgr)

	// V2.0 迭代10: 初始化日志数据库
	var logDB *database.LogDatabase
	if sqlDB := mgr.GetDB(); sqlDB != nil {
		logDB, _ = database.NewLogDatabase("data/operation_logs.db")
		// 添加操作日志中间件
		if logDB != nil {
			r.Use(OperationLogMiddleware(logDB))
		}
	}

	// V2.0 迭代10: 初始化管理员处理器和H5处理器
	var adminHandler *AdminHandler
	var h5Handler *H5Handler
	if logDB != nil && mgr.GetDB() != nil {
		db := &database.Database{DB: mgr.GetDB()}
		adminHandler = NewAdminHandler(db, logDB)
	}
	
	// H5处理器（H5登录是公开接口，不依赖日志数据库）
	authPlugin := getAuthPlugin(mgr)
	if authPlugin != nil {
		h5Handler = NewH5Handler(authPlugin)
	}

	api := r.Group("/api")
	{
		// 认证接口（无需鉴权）
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", handler.HandleRegister)
			authGroup.POST("/login", handler.HandleLogin)
			authGroup.POST("/wx-login", handler.HandleWxLogin)
			authGroup.POST("/refresh", handler.HandleRefresh) // refresh 自行处理 token 验证（支持过期宽限期）
			// V2.0 迭代10: H5认证接口
			if h5Handler != nil {
				authGroup.GET("/wx-h5-login-url", h5Handler.HandleWxH5LoginURL)
				authGroup.POST("/wx-h5-callback", h5Handler.HandleWxH5Callback)
			}
		}

		// V2.0 迭代10: 平台配置（无需鉴权）
		api.GET("/platform/config", HandleGetPlatformConfig)

		// 需要鉴权的路由
		authorized := api.Group("")
		if jwtManager != nil {
			authorized.Use(auth.JWTAuthMiddleware(jwtManager))
		}

		// V2.0 迭代10: 用户状态检查中间件（检查用户是否被禁用）
		if jwtManager != nil && mgr.GetDB() != nil {
			userRepo := database.NewUserRepository(mgr.GetDB())
			authorized.Use(UserStatusChecker(userRepo))
		}
		{
			authorized.POST("/auth/complete-profile", handler.HandleCompleteProfile)
			// 对话接口（额外注册对话限流中间件）
			chatRateLimit := ChatRateLimitMiddleware(cfg.Performance.RateLimiting)
			authorized.POST("/chat", chatRateLimit, handler.HandleChat)
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

			// V2.0 迭代6 风格配置
			authorized.PUT("/styles", auth.RoleRequired("teacher"), handler.HandleSetStyleConfig)
			authorized.GET("/styles", handler.HandleGetStyleConfig)

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

			// 知识库增强
			authorized.POST("/documents/upload", auth.RoleRequired("teacher", "admin"), handler.HandleUploadDocument)
			authorized.POST("/documents/import-url", auth.RoleRequired("teacher", "admin"), handler.HandleImportURL)

			// SSE 流式对话（额外注册对话限流中间件）
			authorized.POST("/chat/stream", chatRateLimit, handler.HandleChatStream)

			// V2.0 迭代2 新增路由

			// 分身管理
			personas := authorized.Group("/personas")
			{
				personas.POST("", handler.HandleCreatePersona)
				personas.GET("", handler.HandleGetPersonas)
				personas.PUT("/:id", handler.HandleEditPersona)
				// V2.0 迭代11 M3：以下接口已废弃，已从路由中移除
				// personas.PUT("/:id/activate", handler.HandleActivatePersona)
				// personas.PUT("/:id/deactivate", handler.HandleDeactivatePersona)
				// personas.PUT("/:id/switch", handler.HandleSwitchPersona)
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

			// V2.0 迭代8 新增路由

			// 智能知识库上传
			authorized.POST("/knowledge/upload", auth.RoleRequired("teacher", "admin"), handler.HandleSmartUpload)
			authorized.GET("/knowledge", auth.RoleRequired("teacher", "admin"), handler.HandleSearchKnowledge)
			authorized.GET("/knowledge/:id", auth.RoleRequired("teacher", "admin"), handler.HandleGetKnowledgeDetail)
			authorized.PUT("/knowledge/:id", auth.RoleRequired("teacher", "admin"), handler.HandleUpdateKnowledge)
			authorized.DELETE("/knowledge/:id", auth.RoleRequired("teacher", "admin"), handler.HandleDeleteKnowledge)

			// 班级管理增强（V8版本）
			classes.POST("/v8", auth.RoleRequired("teacher"), handler.HandleCreateClassV8)
			classes.GET("/:id/share-info", auth.RoleRequired("teacher"), handler.HandleGetClassShareInfo)
			classes.GET("/:id/members/v8", auth.RoleRequired("teacher"), handler.HandleGetClassMembersV8)

			// 学生加入班级申请
			authorized.POST("/classes/join", auth.RoleRequired("student"), handler.HandleJoinClass)
			authorized.GET("/join-requests/pending", auth.RoleRequired("teacher"), handler.HandleGetPendingRequests)
			authorized.PUT("/join-requests/:id/approve", auth.RoleRequired("teacher"), handler.HandleApproveJoinRequest)
			authorized.PUT("/join-requests/:id/reject", auth.RoleRequired("teacher"), handler.HandleRejectJoinRequest)

			// 聊天列表重构
			authorized.GET("/chat-list/teacher", auth.RoleRequired("teacher"), handler.HandleGetTeacherChatList)
			authorized.GET("/chat-list/student", auth.RoleRequired("student"), handler.HandleGetStudentTeacherList)

			// 置顶功能
			authorized.POST("/chat-pins", handler.HandlePinChat)
			authorized.DELETE("/chat-pins/:type/:id", handler.HandleUnpinChat)
			authorized.GET("/chat-pins", handler.HandleGetPinnedChats)

			// 发现页（P1优先级）
			authorized.GET("/discover", handler.HandleGetDiscover)
			authorized.GET("/discover/detail", handler.HandleGetDiscoverDetail)
			authorized.GET("/discover/search", handler.HandleDiscoverSearch)

			// 学生基础信息
			authorized.PUT("/user/student-profile", auth.RoleRequired("student"), handler.HandleUpdateStudentProfile)

			// 会话管理
			authorized.POST("/chat/new-session", handler.HandleNewSession)
			authorized.GET("/chat/quick-actions", handler.HandleGetQuickActions)

			// V2.0 迭代9 M5：会话标题生成
			authorized.POST("/conversations/sessions/:session_id/title", handler.HandleGenerateSessionTitle)

			// V2.0 迭代9 M7：课程信息发布
			authorized.POST("/courses", auth.RoleRequired("teacher"), handler.HandleCreateCourse)
			authorized.GET("/courses", auth.RoleRequired("teacher"), handler.HandleGetCourses)
			authorized.PUT("/courses/:id", auth.RoleRequired("teacher"), handler.HandleUpdateCourse)
			authorized.DELETE("/courses/:id", auth.RoleRequired("teacher"), handler.HandleDeleteCourse)
			authorized.POST("/courses/:id/push", auth.RoleRequired("teacher"), handler.HandlePushCourseNotification)

			// V2.0 迭代9 M6：头像点击API
			// 注意：classes.GET("/:id") 需要在 classes 组内注册，但要放在其他带参数的路由之前
			// 这里改为在 classes 组外单独注册，以避免路由冲突
			authorized.GET("/classes/:id", handler.HandleGetClassForStudent)
			// 教师查看学生详情（需教师角色和师生关系校验）
			authorized.GET("/students/:id/profile", auth.RoleRequired("teacher"), handler.HandleGetStudentProfile)
			// 教师更新学生评语（需教师角色和师生关系校验）
			authorized.PUT("/students/:id/evaluation", auth.RoleRequired("teacher"), handler.HandleUpdateStudentEvaluation)

			// V2.0 迭代7 新增路由

			// 教师消息推送
			authorized.POST("/teacher-messages", auth.RoleRequired("teacher"), handler.HandlePushTeacherMessage)
			authorized.GET("/teacher-messages/history", auth.RoleRequired("teacher"), handler.HandleGetTeacherMessageHistory)

			// 教材配置（R1）
			authorized.POST("/curriculum-configs", auth.RoleRequired("teacher"), handler.HandleCreateCurriculumConfig)
			authorized.GET("/curriculum-configs", auth.RoleRequired("teacher"), handler.HandleGetCurriculumConfigs)
			authorized.DELETE("/curriculum-configs/:id", auth.RoleRequired("teacher"), handler.HandleDeleteCurriculumConfig)

			// 用户反馈（R3）
			authorized.POST("/feedbacks", handler.HandleCreateFeedback)
			authorized.GET("/feedbacks", auth.RoleRequired("teacher", "admin"), handler.HandleGetFeedbacks)
			authorized.PUT("/feedbacks/:id/status", auth.RoleRequired("teacher", "admin"), handler.HandleUpdateFeedbackStatus)

			// 批量添加学生（R8）
			authorized.POST("/students/parse-text", auth.RoleRequired("teacher"), handler.HandleParseStudentText)
			authorized.POST("/students/batch-create", auth.RoleRequired("teacher"), handler.HandleBatchCreateStudents)

			// 教材配置更新（R1补全 - M3）
			authorized.PUT("/curriculum-configs/:id", auth.RoleRequired("teacher"), handler.HandleUpdateCurriculumConfig)
			authorized.GET("/curriculum-versions", handler.HandleGetCurriculumVersions)

			// 批量文档上传（R5 - M5）
			authorized.POST("/documents/batch-upload", auth.RoleRequired("teacher", "admin"), handler.HandleBatchUpload)
			authorized.GET("/batch-tasks/:task_id", auth.RoleRequired("teacher", "admin"), handler.HandleGetBatchTask)

			// V2.0 迭代10 新增路由

			// H5文件上传
			if h5Handler != nil {
				authorized.POST("/upload/h5", h5Handler.HandleH5Upload)
			}

			// V2.0 迭代11 M4：自测学生管理
			testStudent := authorized.Group("/test-student")
			{
				testStudent.GET("", auth.RoleRequired("teacher"), handler.HandleGetTestStudent)
				testStudent.POST("/reset", auth.RoleRequired("teacher"), handler.HandleResetTestStudent)
				testStudent.POST("/login", auth.RoleRequired("teacher"), handler.HandleTestStudentLogin)
			}

			// 管理员路由组
			if adminHandler != nil {
				admin := api.Group("/admin")
				if jwtManager != nil {
					admin.Use(auth.JWTAuthMiddleware(jwtManager))
					admin.Use(auth.RoleRequired("admin"))
				}
				{
					// 仪表盘
					admin.GET("/dashboard/overview", adminHandler.HandleAdminDashboardOverview)
					admin.GET("/dashboard/user-stats", adminHandler.HandleAdminUserStats)
					admin.GET("/dashboard/chat-stats", adminHandler.HandleAdminChatStats)
					admin.GET("/dashboard/knowledge-stats", adminHandler.HandleAdminKnowledgeStats)
					admin.GET("/dashboard/active-users", adminHandler.HandleAdminActiveUsers)

					// 用户管理
					admin.GET("/users", adminHandler.HandleAdminGetUsers)
					admin.PUT("/users/:id/role", adminHandler.HandleAdminUpdateUserRole)
					admin.PUT("/users/:id/status", adminHandler.HandleAdminUpdateUserStatus)

					// 反馈管理
					admin.GET("/feedbacks", adminHandler.HandleAdminGetFeedbacks)

					// 操作日志
					admin.GET("/logs", adminHandler.HandleAdminGetLogs)
					admin.GET("/logs/stats", adminHandler.HandleAdminGetLogStats)
					admin.GET("/logs/export", adminHandler.HandleAdminExportLogs)
				}
			}
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

// getAuthPlugin 获取认证插件
func getAuthPlugin(mgr *manager.HarnessManager) *auth.AuthPlugin {
	plugin, _ := mgr.GetPlugin("auth")
	if plugin == nil {
		return nil
	}
	authPlugin, ok := plugin.(*auth.AuthPlugin)
	if !ok {
		return nil
	}
	return authPlugin
}
