package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"digital-twin/src/backend/database"
	"digital-twin/src/harness/core"

	"golang.org/x/crypto/bcrypt"
)

// AuthPlugin 认证插件
type AuthPlugin struct {
	*core.BasePlugin
	db          *sql.DB
	userRepo    *database.UserRepository
	personaRepo *database.PersonaRepository // V2.0 迭代2：分身仓库
	jwtManager  *JWTManager
	wxClient    WxClient
}

// NewAuthPlugin 创建认证插件
func NewAuthPlugin(name string, db *sql.DB) *AuthPlugin {
	return &AuthPlugin{
		BasePlugin:  core.NewBasePlugin(name, "1.0.0", core.PluginTypeAuth),
		db:          db,
		userRepo:    database.NewUserRepository(db),
		personaRepo: database.NewPersonaRepository(db),
	}
}

// Init 初始化认证插件
// 从 config 中读取 jwt.secret, jwt.expiry, jwt.issuer
func (p *AuthPlugin) Init(config map[string]interface{}) error {
	// 调用基类 Init
	if err := p.BasePlugin.Init(config); err != nil {
		return err
	}

	// 读取 JWT 配置
	secret, _ := config["jwt.secret"].(string)
	if secret == "" {
		secret = "default-secret-key"
	}

	issuer, _ := config["jwt.issuer"].(string)
	if issuer == "" {
		issuer = "digital-twin"
	}

	expiryStr, _ := config["jwt.expiry"].(string)
	expiry := 24 * time.Hour // 默认 24 小时
	if expiryStr != "" {
		if d, err := time.ParseDuration(expiryStr); err == nil {
			expiry = d
		}
	}

	p.jwtManager = NewJWTManager(secret, issuer, expiry)

	// 根据 WX_MODE 环境变量初始化微信客户端
	wxMode := os.Getenv("WX_MODE")
	if wxMode == "mock" {
		p.wxClient = &MockWxClient{}
	} else {
		p.wxClient = &RealWxClient{
			AppID:  os.Getenv("WX_APPID"),
			Secret: os.Getenv("WX_SECRET"),
		}
	}

	return nil
}

// isAuthAction 判断是否是认证插件自己的 action
func (p *AuthPlugin) isAuthAction(action string) bool {
	switch action {
	case "register", "login", "verify", "refresh", "wx-login", "complete-profile", "wx-h5-login-url", "wx-h5-callback":
		return true
	default:
		return false
	}
}

// Execute 执行认证操作
// 根据 input.Data["action"] 分发到不同的处理逻辑
func (p *AuthPlugin) Execute(ctx context.Context, input *core.PluginInput) (*core.PluginOutput, error) {
	start := time.Now()

	action, _ := input.Data["action"].(string)

	// 管道透传模式：action 不是 auth 插件的 action 且 UserContext 已填充
	if !p.isAuthAction(action) {
		if input.UserContext != nil && input.UserContext.UserID != "" {
			outputData := mergeData(input.Data, nil)
			return &core.PluginOutput{
				Success:  true,
				Data:     outputData,
				Duration: time.Since(start),
				Metadata: map[string]interface{}{"plugin": "auth", "mode": "passthrough"},
			}, nil
		}
		if action == "" {
			return &core.PluginOutput{
				Success:  false,
				Data:     map[string]interface{}{"error_code": 40004},
				Error:    "缺少 action 参数",
				Duration: time.Since(start),
			}, nil
		}
	}

	var output *core.PluginOutput
	var err error

	switch action {
	case "register":
		output, err = p.handleRegister(input)
	case "login":
		output, err = p.handleLogin(input)
	case "verify":
		output, err = p.handleVerify(input)
	case "refresh":
		output, err = p.handleRefresh(input)
	case "wx-login":
		output, err = p.handleWxLogin(input)
	case "complete-profile":
		output, err = p.handleCompleteProfile(input)
	case "wx-h5-login-url":
		output, err = p.handleWxH5LoginURL(input)
	case "wx-h5-callback":
		output, err = p.handleWxH5Callback(input)
	default:
		output = &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   fmt.Sprintf("不支持的 action: %s", action),
		}
	}

	if err != nil {
		return &core.PluginOutput{
			Success:  false,
			Data:     map[string]interface{}{"error_code": 50001},
			Error:    err.Error(),
			Duration: time.Since(start),
		}, nil
	}

	output.Duration = time.Since(start)
	return output, nil
}

// GetJWTManager 获取 JWT 管理器（供外部使用，如中间件）
func (p *AuthPlugin) GetJWTManager() *JWTManager {
	return p.jwtManager
}

// handleRegister 处理用户注册
func (p *AuthPlugin) handleRegister(input *core.PluginInput) (*core.PluginOutput, error) {
	// 提取参数
	username, _ := input.Data["username"].(string)
	password, _ := input.Data["password"].(string)
	role, _ := input.Data["role"].(string)
	nickname, _ := input.Data["nickname"].(string)
	email, _ := input.Data["email"].(string)

	// 参数校验
	if username == "" || password == "" {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "用户名和密码不能为空",
		}, nil
	}

	if role == "" {
		role = "student"
	}

	// 检查用户名是否已存在
	existingUser, err := p.userRepo.GetByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if existingUser != nil {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40006},
			Error:   "用户名已存在",
		}, nil
	}

	// 密码 bcrypt 加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建用户
	user := &database.User{
		Username: username,
		Password: string(hashedPassword),
		Role:     role,
		Nickname: nickname,
		Email:    email,
	}

	userID, err := p.userRepo.Create(user)
	if err != nil {
		// 检查是否为唯一约束冲突
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "unique") {
			return &core.PluginOutput{
				Success: false,
				Data:    map[string]interface{}{"error_code": 40006},
				Error:   "用户名已存在",
			}, nil
		}
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 生成 token（用户角色和分身角色相同）
	token, expiresAt, err := p.jwtManager.GenerateTokenWithUserRole(userID, username, role, role)
	if err != nil {
		return nil, fmt.Errorf("生成 token 失败: %w", err)
	}

	// 构建输出数据，先 merge 上游 Data
	outputData := mergeData(input.Data, map[string]interface{}{
		"user_id":    userID,
		"token":      token,
		"role":       role,
		"nickname":   nickname,
		"expires_at": expiresAt.Format(time.RFC3339),
	})

	return &core.PluginOutput{
		Success:  true,
		Data:     outputData,
		Metadata: map[string]interface{}{"plugin": "auth", "action": "register"},
	}, nil
}

// handleLogin 处理用户登录
func (p *AuthPlugin) handleLogin(input *core.PluginInput) (*core.PluginOutput, error) {
	username, _ := input.Data["username"].(string)
	password, _ := input.Data["password"].(string)

	// 参数校验
	if username == "" || password == "" {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "用户名和密码不能为空",
		}, nil
	}

	// 查询用户
	user, err := p.userRepo.GetByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40001},
			Error:   "用户名或密码错误",
		}, nil
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40001},
			Error:   "用户名或密码错误",
		}, nil
	}

	// 生成 token（用户角色和分身角色相同）
	token, expiresAt, err := p.jwtManager.GenerateTokenWithUserRole(user.ID, user.Username, user.Role, user.Role)
	if err != nil {
		return nil, fmt.Errorf("生成 token 失败: %w", err)
	}

	// 构建输出数据，先 merge 上游 Data
	outputData := mergeData(input.Data, map[string]interface{}{
		"user_id":    user.ID,
		"token":      token,
		"role":       user.Role,
		"nickname":   user.Nickname,
		"expires_at": expiresAt.Format(time.RFC3339),
	})

	return &core.PluginOutput{
		Success:  true,
		Data:     outputData,
		Metadata: map[string]interface{}{"plugin": "auth", "action": "login"},
	}, nil
}

// handleVerify 验证 JWT token
func (p *AuthPlugin) handleVerify(input *core.PluginInput) (*core.PluginOutput, error) {
	tokenStr, _ := input.Data["token"].(string)
	if tokenStr == "" {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40001},
			Error:   "缺少 token",
		}, nil
	}

	claims, err := p.jwtManager.ValidateToken(tokenStr)
	if err != nil {
		if strings.Contains(err.Error(), "expired") {
			return &core.PluginOutput{
				Success: false,
				Data:    map[string]interface{}{"error_code": 40002},
				Error:   "令牌已过期",
			}, nil
		}
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40001},
			Error:   "令牌无效",
		}, nil
	}

	// 填充 UserContext
	if input.UserContext == nil {
		input.UserContext = &core.UserContext{}
	}
	input.UserContext.UserID = fmt.Sprintf("%d", claims.UserID)
	input.UserContext.Role = claims.Role

	// 构建输出数据，先 merge 上游 Data
	outputData := mergeData(input.Data, map[string]interface{}{
		"user_id":  claims.UserID,
		"username": claims.Username,
		"role":     claims.Role,
	})

	return &core.PluginOutput{
		Success:  true,
		Data:     outputData,
		Metadata: map[string]interface{}{"plugin": "auth", "action": "verify"},
	}, nil
}

// handleRefresh 刷新 token
func (p *AuthPlugin) handleRefresh(input *core.PluginInput) (*core.PluginOutput, error) {
	tokenStr, _ := input.Data["token"].(string)
	if tokenStr == "" {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40001},
			Error:   "缺少 token",
		}, nil
	}

	// 先尝试正常验证
	claims, err := p.jwtManager.ValidateToken(tokenStr)
	if err != nil {
		// 正常验证失败，尝试忽略过期来解析
		parsedClaims, isExpired, parseErr := p.jwtManager.ParseTokenIgnoreExpiry(tokenStr)
		if parseErr != nil {
			// 签名无效等根本性错误
			return &core.PluginOutput{
				Success: false,
				Data:    map[string]interface{}{"error_code": 40001},
				Error:   "令牌无效，无法刷新",
			}, nil
		}

		if isExpired {
			// 检查是否在宽限期内
			if parsedClaims.ExpiresAt != nil {
				expiredDuration := time.Since(parsedClaims.ExpiresAt.Time)
				if expiredDuration > RefreshGracePeriod {
					return &core.PluginOutput{
						Success: false,
						Data:    map[string]interface{}{"error_code": 40002},
						Error:   "令牌已过期且超过刷新宽限期",
					}, nil
				}
			}
		}

		// 宽限期内，使用解析出的 claims
		claims = parsedClaims
	}

	// 生成新 token（含 persona_id 和 user_role）
	// 兼容旧token：如果没有UserRole则使用Role
	userRole := claims.UserRole
	if userRole == "" {
		userRole = claims.Role
	}
	newToken, expiresAt, err := p.jwtManager.GenerateTokenWithUserRole(claims.UserID, claims.Username, claims.Role, userRole, claims.PersonaID)
	if err != nil {
		return nil, fmt.Errorf("生成新 token 失败: %w", err)
	}

	outputData := mergeData(input.Data, map[string]interface{}{
		"user_id":    claims.UserID,
		"token":      newToken,
		"role":       claims.Role,
		"expires_at": expiresAt.Format(time.RFC3339),
	})

	return &core.PluginOutput{
		Success:  true,
		Data:     outputData,
		Metadata: map[string]interface{}{"plugin": "auth", "action": "refresh"},
	}, nil
}

// handleWxLogin 处理微信登录
// V2.0 迭代2 改造：返回分身列表和当前分身
func (p *AuthPlugin) handleWxLogin(input *core.PluginInput) (*core.PluginOutput, error) {
	code, _ := input.Data["code"].(string)
	if code == "" {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "缺少 code 参数",
		}, nil
	}

	// 调用微信 API 获取 openid
	session, err := p.wxClient.Code2Session(code)
	if err != nil {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 50001},
			Error:   "微信登录失败",
		}, nil
	}

	if session.OpenID == "" {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "无效的登录凭证",
		}, nil
	}

	// 根据 openid 查询用户
	user, err := p.userRepo.GetByOpenID(session.OpenID)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	isNewUser := false
	if user == nil {
		// 新用户：自动创建
		isNewUser = true
		openidSuffix := session.OpenID
		if len(openidSuffix) > 6 {
			openidSuffix = openidSuffix[len(openidSuffix)-6:]
		}

		// 生成随机密码
		randomBytes := make([]byte, 16)
		rand.Read(randomBytes)
		randomPassword := hex.EncodeToString(randomBytes)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(randomPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("密码加密失败: %w", err)
		}

		user = &database.User{
			Username: "wx_用户_" + openidSuffix,
			Password: string(hashedPassword),
			Role:     "",
			Nickname: "",
			OpenID:   session.OpenID,
		}

		userID, err := p.userRepo.CreateWithOpenID(user)
		if err != nil {
			return nil, fmt.Errorf("创建微信用户失败: %w", err)
		}
		user.ID = userID
	}

	// V2.0 迭代2：查询用户的分身列表
	personas, err := p.personaRepo.ListByUserID(user.ID, "")
	if err != nil {
		return nil, fmt.Errorf("查询分身列表失败: %w", err)
	}

	// 确定当前分身和角色
	var currentPersona interface{}
	personaID := int64(0)
	role := user.Role

	if len(personas) > 0 {
		// 有分身，找默认分身
		if user.DefaultPersonaID > 0 {
			for _, persona := range personas {
				if persona.ID == user.DefaultPersonaID {
					currentPersona = persona
					personaID = persona.ID
					role = persona.Role
					break
				}
			}
		}
		// 如果默认分身未找到，使用第一个
		if currentPersona == nil {
			currentPersona = personas[0]
			personaID = personas[0].ID
			role = personas[0].Role
		}
	}

	// 生成 JWT token（含 persona_id 和 user_role）
	// user.Role 是用户级别角色（可能为admin），role 是当前分身角色
	token, expiresAt, err := p.jwtManager.GenerateTokenWithUserRole(user.ID, user.Username, role, user.Role, personaID)
	if err != nil {
		return nil, fmt.Errorf("生成 token 失败: %w", err)
	}

	outputData := mergeData(input.Data, map[string]interface{}{
		"user_id":         user.ID,
		"token":           token,
		"role":            role,
		"nickname":        user.Nickname,
		"is_new_user":     isNewUser,
		"expires_at":      expiresAt.Format(time.RFC3339),
		"personas":        personas,
		"current_persona": currentPersona,
	})

	return &core.PluginOutput{
		Success:  true,
		Data:     outputData,
		Metadata: map[string]interface{}{"plugin": "auth", "action": "wx-login"},
	}, nil
}

// handleCompleteProfile 处理新用户补全信息（角色+昵称）
// V2.0 迭代2 改造：内部转为创建分身
// V2.0 迭代11 M4：教师角色创建自测学生
func (p *AuthPlugin) handleCompleteProfile(input *core.PluginInput) (*core.PluginOutput, error) {
	// 从 input.Data 获取 user_id（由 JWT 中间件解析后传入）
	userID := int64(0)
	switch v := input.Data["user_id"].(type) {
	case int64:
		userID = v
	case float64:
		userID = int64(v)
	case int:
		userID = int64(v)
	}
	if userID == 0 {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40001},
			Error:   "用户信息无效",
		}, nil
	}

	role, _ := input.Data["role"].(string)
	nickname, _ := input.Data["nickname"].(string)
	school, _ := input.Data["school"].(string)
	description, _ := input.Data["description"].(string)

	// 校验 role
	if role != "teacher" && role != "student" {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "角色只能是 teacher 或 student",
		}, nil
	}

	// 校验 nickname
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "昵称不能为空",
		}, nil
	}
	nicknameRunes := []rune(nickname)
	if len(nicknameRunes) > 20 {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "昵称长度不能超过20个字符",
		}, nil
	}

	user, err := p.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40005},
			Error:   "用户不存在",
		}, nil
	}

	// V2.0 迭代2 改造：不再检查 user.Role != ""，改为创建分身
	if role == "teacher" {
		school = strings.TrimSpace(school)
		if school == "" {
			return &core.PluginOutput{
				Success: false,
				Data:    map[string]interface{}{"error_code": 40004},
				Error:   "教师角色必须填写学校名称",
			}, nil
		}
		description = strings.TrimSpace(description)
		if description == "" {
			return &core.PluginOutput{
				Success: false,
				Data:    map[string]interface{}{"error_code": 40004},
				Error:   "教师角色必须填写分身描述",
			}, nil
		}
		// 检查同名+同校唯一性（在 personas 表中）
		exists, err := p.personaRepo.CheckTeacherPersonaExists(nickname, school, 0)
		if err != nil {
			return nil, fmt.Errorf("检查教师分身唯一性失败: %w", err)
		}
		if exists {
			return &core.PluginOutput{
				Success: false,
				Data:    map[string]interface{}{"error_code": 40015},
				Error:   "该学校已有同名教师分身，请修改名称",
			}, nil
		}
	}

	// 创建分身
	persona := &database.Persona{
		UserID:      userID,
		Role:        role,
		Nickname:    nickname,
		School:      school,
		Description: description,
	}
	personaID, err := p.personaRepo.Create(persona)
	if err != nil {
		return nil, fmt.Errorf("创建分身失败: %w", err)
	}

	// 如果是第一个分身，设为默认分身
	count, _ := p.personaRepo.CountByUserID(userID)
	if count <= 1 {
		p.userRepo.UpdateDefaultPersonaID(userID, personaID)
	}

	// 同时更新 users 表的 role（向后兼容）
	if user.Role == "" {
		p.userRepo.UpdateProfile(userID, role, nickname, school, description)
	}

	// 生成新的 JWT token（含 persona_id 和 user_role）
	// user_role 使用分身角色，因为用户完成资料补全时选择了该角色
	token, expiresAt, err := p.jwtManager.GenerateTokenWithUserRole(userID, user.Username, role, role, personaID)
	if err != nil {
		return nil, fmt.Errorf("生成新 token 失败: %w", err)
	}

	outputData := mergeData(input.Data, map[string]interface{}{
		"user_id":     userID,
		"role":        role,
		"nickname":    nickname,
		"school":      school,
		"description": description,
		"persona_id":  personaID,
		"token":       token,
		"expires_at":  expiresAt.Format(time.RFC3339),
	})

	// V2.0 迭代11 M4：教师角色创建自测学生
	if role == "teacher" {
		testStudent, password, err := p.createTestStudent(userID, user.Username)
		if err != nil {
			// 创建自测学生失败不影响主流程，仅记录日志
			fmt.Printf("创建自测学生失败: %v\n", err)
		} else if testStudent != nil {
			// 查询自测学生的分身
			testPersona, _ := p.personaRepo.GetStudentPersonaByUserID(testStudent.ID)
			testStudentInfo := map[string]interface{}{
				"user_id":       testStudent.ID,
				"username":      testStudent.Username,
				"nickname":      testStudent.Nickname,
				"password_hint": "初始密码为6位数字，请在首次登录后修改",
			}
			if testPersona != nil {
				testStudentInfo["persona_id"] = testPersona.ID
			}
			// 只有创建时才返回密码
			if password != "" {
				testStudentInfo["initial_password"] = password
			}
			outputData["test_student"] = testStudentInfo
		}
	}

	return &core.PluginOutput{
		Success:  true,
		Data:     outputData,
		Metadata: map[string]interface{}{"plugin": "auth", "action": "complete-profile"},
	}, nil
}

// createTestStudent 创建自测学生账号
func (p *AuthPlugin) createTestStudent(teacherID int64, teacherUsername string) (*database.User, string, error) {
	// 检查是否已存在自测学生
	existing, err := p.userRepo.FindByTestTeacherID(teacherID)
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

	userID, err := p.userRepo.CreateTestStudent(testStudent)
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

	personaID, err := p.personaRepo.Create(persona)
	if err != nil {
		return nil, "", fmt.Errorf("创建自测学生分身失败: %w", err)
	}

	// 更新默认分身ID
	p.userRepo.UpdateDefaultPersonaID(userID, personaID)

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

// handleWxH5LoginURL 生成微信H5网页授权URL
func (p *AuthPlugin) handleWxH5LoginURL(input *core.PluginInput) (*core.PluginOutput, error) {
	redirectURI, _ := input.Data["redirect_uri"].(string)
	state, _ := input.Data["state"].(string)

	if redirectURI == "" {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "缺少 redirect_uri 参数",
		}, nil
	}

	// 生成授权URL
	authURL := p.wxClient.GetAuthURL(redirectURI, state)

	outputData := mergeData(input.Data, map[string]interface{}{
		"auth_url": authURL,
	})

	return &core.PluginOutput{
		Success:  true,
		Data:     outputData,
		Metadata: map[string]interface{}{"plugin": "auth", "action": "wx-h5-login-url"},
	}, nil
}

// handleWxH5Callback 处理微信H5授权回调
func (p *AuthPlugin) handleWxH5Callback(input *core.PluginInput) (*core.PluginOutput, error) {
	code, _ := input.Data["code"].(string)
	if code == "" {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "缺少 code 参数",
		}, nil
	}

	// 调用微信 API 获取 access_token 和 openid
	tokenResp, err := p.wxClient.GetAccessToken(code)
	if err != nil {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 50001},
			Error:   "微信授权失败",
		}, nil
	}

	if tokenResp.OpenID == "" {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40004},
			Error:   "无效的授权凭证",
		}, nil
	}

	// 获取用户信息（需 scope=snsapi_userinfo）
	userInfo, err := p.wxClient.GetUserInfo(tokenResp.AccessToken, tokenResp.OpenID)
	var nickname string
	if err == nil && userInfo != nil {
		nickname = userInfo.Nickname
	}

	// 根据 openid 查询用户
	user, err := p.userRepo.GetByOpenID(tokenResp.OpenID)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	isNewUser := false
	if user == nil {
		// 新用户：自动创建
		isNewUser = true
		openidSuffix := tokenResp.OpenID
		if len(openidSuffix) > 6 {
			openidSuffix = openidSuffix[len(openidSuffix)-6:]
		}

		// 生成随机密码
		randomBytes := make([]byte, 16)
		rand.Read(randomBytes)
		randomPassword := hex.EncodeToString(randomBytes)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(randomPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("密码加密失败: %w", err)
		}

		// 使用微信昵称作为默认昵称
		if nickname == "" {
			nickname = "微信用户"
		}

		user = &database.User{
			Username:  "wx_h5_" + openidSuffix,
			Password:  string(hashedPassword),
			Role:      "",
			Nickname:  nickname,
			OpenID:    tokenResp.OpenID,
			WxUnionID: tokenResp.UnionID,
		}

		userID, err := p.userRepo.CreateWithOpenID(user)
		if err != nil {
			return nil, fmt.Errorf("创建微信H5用户失败: %w", err)
		}
		user.ID = userID
	} else {
		// 已有用户，检查是否需要更新 UnionID
		if user.WxUnionID == "" && tokenResp.UnionID != "" {
			p.userRepo.UpdateWxUnionID(user.ID, tokenResp.UnionID)
		}
	}

	// 检查用户状态
	if user.Status == "disabled" {
		return &core.PluginOutput{
			Success: false,
			Data:    map[string]interface{}{"error_code": 40003},
			Error:   "用户已被禁用",
		}, nil
	}

	// 查询用户的分身列表
	personas, err := p.personaRepo.ListByUserID(user.ID, "")
	if err != nil {
		return nil, fmt.Errorf("查询分身列表失败: %w", err)
	}

	// 确定当前分身和角色
	var currentPersona interface{}
	personaID := int64(0)
	role := user.Role

	if len(personas) > 0 {
		if user.DefaultPersonaID > 0 {
			for _, persona := range personas {
				if persona.ID == user.DefaultPersonaID {
					currentPersona = persona
					personaID = persona.ID
					role = persona.Role
					break
				}
			}
		}
		if currentPersona == nil {
			currentPersona = personas[0]
			personaID = personas[0].ID
			role = personas[0].Role
		}
	}

	// 生成 JWT token（含 user_role）
	token, expiresAt, err := p.jwtManager.GenerateTokenWithUserRole(user.ID, user.Username, role, user.Role, personaID)
	if err != nil {
		return nil, fmt.Errorf("生成 token 失败: %w", err)
	}

	outputData := mergeData(input.Data, map[string]interface{}{
		"user_id":         user.ID,
		"token":           token,
		"role":            role,
		"nickname":        user.Nickname,
		"is_new_user":     isNewUser,
		"expires_at":      expiresAt.Format(time.RFC3339),
		"personas":        personas,
		"current_persona": currentPersona,
	})

	return &core.PluginOutput{
		Success:  true,
		Data:     outputData,
		Metadata: map[string]interface{}{"plugin": "auth", "action": "wx-h5-callback"},
	}, nil
}

// mergeData 合并上游 Data 和本插件输出字段
// 先复制 input.Data 所有字段到 outputData，再添加/覆盖本插件的输出字段
func mergeData(upstream map[string]interface{}, pluginData map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	// 先复制上游数据
	for k, v := range upstream {
		result[k] = v
	}
	// 再覆盖/添加本插件数据
	for k, v := range pluginData {
		result[k] = v
	}
	return result
}
