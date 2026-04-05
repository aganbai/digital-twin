package api

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"digital-twin/src/backend/database"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// ======================== 自测学生接口 (V2.0 迭代11 M4) ========================

// TestStudentInfo 自测学生信息响应
type TestStudentInfo struct {
	UserID        int64       `json:"user_id"`
	Username      string      `json:"username"`
	Nickname      string      `json:"nickname"`
	PersonaID     int64       `json:"persona_id"`
	IsActive      bool        `json:"is_active"`
	PasswordHint  string      `json:"password_hint"`
	JoinedClasses []ClassInfo `json:"joined_classes"`
	CreatedAt     string      `json:"created_at"`
}

// ClassInfo 班级简要信息
type ClassInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// HandleGetTestStudent 获取自测学生信息
// GET /api/test-student
func (h *Handler) HandleGetTestStudent(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok || userIDInt64 <= 0 {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	userRepo := database.NewUserRepository(db)
	personaRepo := database.NewPersonaRepository(db)
	classRepo := database.NewClassRepository(db)

	// 查询自测学生
	testStudent, err := userRepo.FindByTestTeacherID(userIDInt64)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询自测学生失败: "+err.Error())
		return
	}
	if testStudent == nil {
		Error(c, http.StatusNotFound, 40041, "未找到自测学生账号")
		return
	}

	// 查询自测学生的学生分身
	studentPersona, err := personaRepo.GetStudentPersonaByUserID(testStudent.ID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询自测学生分身失败: "+err.Error())
		return
	}
	if studentPersona == nil {
		Error(c, http.StatusInternalServerError, 50001, "自测学生分身不存在")
		return
	}

	// 查询自测学生加入的班级
	classes, err := classRepo.ListClassesByStudentPersonaID(studentPersona.ID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询班级列表失败: "+err.Error())
		return
	}

	// 构建响应
	joinedClasses := make([]ClassInfo, 0, len(classes))
	for _, cls := range classes {
		joinedClasses = append(joinedClasses, ClassInfo{
			ID:   cls.ID,
			Name: cls.Name,
		})
	}

	Success(c, TestStudentInfo{
		UserID:        testStudent.ID,
		Username:      testStudent.Username,
		Nickname:      testStudent.Nickname,
		PersonaID:     studentPersona.ID,
		IsActive:      testStudent.Status != "disabled",
		PasswordHint:  "初始密码为6位数字，请在首次登录后修改",
		JoinedClasses: joinedClasses,
		CreatedAt:     testStudent.CreatedAt.Format(time.RFC3339),
	})
}

// HandleResetTestStudent 重置自测学生数据
// POST /api/test-student/reset
func (h *Handler) HandleResetTestStudent(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok || userIDInt64 <= 0 {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	userRepo := database.NewUserRepository(db)
	personaRepo := database.NewPersonaRepository(db)

	// 查询自测学生
	testStudent, err := userRepo.FindByTestTeacherID(userIDInt64)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询自测学生失败: "+err.Error())
		return
	}
	if testStudent == nil {
		Error(c, http.StatusNotFound, 40041, "未找到自测学生账号")
		return
	}

	// 查询自测学生的学生分身
	studentPersona, err := personaRepo.GetStudentPersonaByUserID(testStudent.ID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询自测学生分身失败: "+err.Error())
		return
	}
	if studentPersona == nil {
		Error(c, http.StatusInternalServerError, 50001, "自测学生分身不存在")
		return
	}

	// 清空对话记录
	conversationRepo := database.NewConversationRepository(db)
	clearedConversations, err := conversationRepo.DeleteByStudentPersonaID(studentPersona.ID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "清空对话记录失败: "+err.Error())
		return
	}

	// 清空记忆
	memoryRepo := database.NewMemoryRepository(db)
	clearedMemories, err := memoryRepo.DeleteByStudentPersonaID(studentPersona.ID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "清空记忆失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"cleared_conversations": clearedConversations,
		"cleared_memories":      clearedMemories,
		"message":               "自测学生数据已重置",
	})
}

// HandleTestStudentLogin 模拟自测学生登录
// POST /api/test-student/login
func (h *Handler) HandleTestStudentLogin(c *gin.Context) {
	var req struct {
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok || userIDInt64 <= 0 {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	userRepo := database.NewUserRepository(db)
	personaRepo := database.NewPersonaRepository(db)

	// 查询自测学生
	testStudent, err := userRepo.FindByTestTeacherID(userIDInt64)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询自测学生失败: "+err.Error())
		return
	}
	if testStudent == nil {
		Error(c, http.StatusNotFound, 40041, "未找到自测学生账号")
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(testStudent.Password), []byte(req.Password)); err != nil {
		Error(c, http.StatusUnauthorized, 40001, "密码错误")
		return
	}

	// 查询自测学生的学生分身
	studentPersona, err := personaRepo.GetStudentPersonaByUserID(testStudent.ID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询自测学生分身失败: "+err.Error())
		return
	}
	if studentPersona == nil {
		Error(c, http.StatusInternalServerError, 50001, "自测学生分身不存在")
		return
	}

	// 获取 JWT 管理器
	jwtManager := GetJWTManager(h.manager)
	if jwtManager == nil {
		Error(c, http.StatusInternalServerError, 50001, "认证服务不可用")
		return
	}

	// 生成 token（用户角色为 student，分身角色为 student）
	token, expiresAt, err := jwtManager.GenerateTokenWithUserRole(
		testStudent.ID,
		testStudent.Username,
		"student",
		"student",
		studentPersona.ID,
	)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "生成令牌失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"user_id":    testStudent.ID,
		"token":      token,
		"nickname":   testStudent.Nickname,
		"expires_at": expiresAt.Format(time.RFC3339),
	})
}

// ======================== 自测学生创建辅助函数 ========================

// createTestStudent 创建自测学生账号（供 auth 插件调用）
func createTestStudent(db *sql.DB, teacherID int64, teacherUsername string) (*database.User, string, error) {
	userRepo := database.NewUserRepository(db)
	personaRepo := database.NewPersonaRepository(db)

	// 检查是否已存在自测学生
	existing, err := userRepo.FindByTestTeacherID(teacherID)
	if err != nil {
		return nil, "", fmt.Errorf("检查自测学生失败: %w", err)
	}
	if existing != nil {
		return existing, "", nil // 已存在，返回
	}

	// 生成6位随机数字密码
	password := generateNumericPassword(6)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建自测学生用户
	testStudent := &database.User{
		Username:      fmt.Sprintf("teacher_%d_test", teacherID),
		Password:      string(hashedPassword),
		Role:          "student",
		Nickname:      "测试学生",
		IsTestStudent: true,
		TestTeacherID: teacherID,
	}

	userID, err := userRepo.CreateTestStudent(testStudent)
	if err != nil {
		return nil, "", fmt.Errorf("创建自测学生用户失败: %w", err)
	}
	testStudent.ID = userID

	// 创建学生分身
	persona := &database.Persona{
		UserID:      userID,
		Role:        "student",
		Nickname:    "测试学生",
		School:      "",
		Description: "教师自测学生账号",
	}

	personaID, err := personaRepo.Create(persona)
	if err != nil {
		return nil, "", fmt.Errorf("创建自测学生分身失败: %w", err)
	}

	// 更新默认分身ID
	userRepo.UpdateDefaultPersonaID(userID, personaID)

	return testStudent, password, nil
}

// generateNumericPassword 生成指定长度的数字密码
func generateNumericPassword(length int) string {
	const digits = "0123456789"
	bytes := make([]byte, length)
	rand.Read(bytes)
	for i := range bytes {
		bytes[i] = digits[int(bytes[i])%len(digits)]
	}
	return string(bytes)
}
