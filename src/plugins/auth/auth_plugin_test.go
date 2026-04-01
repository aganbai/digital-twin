package auth

import (
	"context"
	"os"
	"testing"
	"time"

	"digital-twin/src/backend/database"
	"digital-twin/src/harness/core"
)

// setupTestDB 创建测试用数据库
func setupTestDB(t *testing.T) *database.Database {
	t.Helper()
	tmpFile := t.TempDir() + "/test.db"
	db, err := database.NewDatabase(tmpFile)
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
		os.Remove(tmpFile)
	})
	return db
}

// setupAuthPlugin 创建测试用认证插件
func setupAuthPlugin(t *testing.T) (*AuthPlugin, *database.Database) {
	t.Helper()
	db := setupTestDB(t)
	plugin := NewAuthPlugin("test-auth", db.DB)
	err := plugin.Init(map[string]interface{}{
		"jwt.secret": "test-secret-key-12345",
		"jwt.issuer": "test-issuer",
		"jwt.expiry": "1h",
	})
	if err != nil {
		t.Fatalf("初始化插件失败: %v", err)
	}
	return plugin, db
}

// ==================== JWT 测试 ====================

func TestJWTManager_GenerateAndValidateToken(t *testing.T) {
	jm := NewJWTManager("test-secret", "test-issuer", time.Hour)

	token, expiresAt, err := jm.GenerateToken(1, "testuser", "student")
	if err != nil {
		t.Fatalf("生成 token 失败: %v", err)
	}
	if token == "" {
		t.Fatal("生成的 token 为空")
	}
	if expiresAt.Before(time.Now()) {
		t.Fatal("过期时间不应早于当前时间")
	}

	// 验证 token
	claims, err := jm.ValidateToken(token)
	if err != nil {
		t.Fatalf("验证 token 失败: %v", err)
	}
	if claims.UserID != 1 {
		t.Errorf("期望 UserID=1, 实际=%d", claims.UserID)
	}
	if claims.Username != "testuser" {
		t.Errorf("期望 Username=testuser, 实际=%s", claims.Username)
	}
	if claims.Role != "student" {
		t.Errorf("期望 Role=student, 实际=%s", claims.Role)
	}
}

func TestJWTManager_InvalidToken(t *testing.T) {
	jm := NewJWTManager("test-secret", "test-issuer", time.Hour)

	_, err := jm.ValidateToken("invalid-token-string")
	if err == nil {
		t.Fatal("应该返回错误")
	}
}

func TestJWTManager_WrongSecret(t *testing.T) {
	jm1 := NewJWTManager("secret-1", "issuer", time.Hour)
	jm2 := NewJWTManager("secret-2", "issuer", time.Hour)

	token, _, err := jm1.GenerateToken(1, "user", "student")
	if err != nil {
		t.Fatalf("生成 token 失败: %v", err)
	}

	_, err = jm2.ValidateToken(token)
	if err == nil {
		t.Fatal("使用错误的密钥验证应该失败")
	}
}

// ==================== AuthPlugin 测试 ====================

func TestAuthPlugin_NewAndInit(t *testing.T) {
	db := setupTestDB(t)
	plugin := NewAuthPlugin("auth-test", db.DB)

	if plugin.Name() != "auth-test" {
		t.Errorf("期望名称=auth-test, 实际=%s", plugin.Name())
	}
	if plugin.Type() != core.PluginTypeAuth {
		t.Errorf("期望类型=auth, 实际=%s", plugin.Type())
	}

	err := plugin.Init(map[string]interface{}{
		"jwt.secret": "my-secret",
		"jwt.issuer": "my-issuer",
		"jwt.expiry": "2h",
	})
	if err != nil {
		t.Fatalf("初始化失败: %v", err)
	}

	if plugin.GetJWTManager() == nil {
		t.Fatal("JWTManager 不应为 nil")
	}
}

func TestAuthPlugin_InitDefaultConfig(t *testing.T) {
	db := setupTestDB(t)
	plugin := NewAuthPlugin("auth-test", db.DB)

	// 使用空配置，应使用默认值
	err := plugin.Init(map[string]interface{}{})
	if err != nil {
		t.Fatalf("使用默认配置初始化失败: %v", err)
	}
	if plugin.GetJWTManager() == nil {
		t.Fatal("JWTManager 不应为 nil")
	}
}

func TestAuthPlugin_Register(t *testing.T) {
	plugin, _ := setupAuthPlugin(t)

	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action":   "register",
			"username": "newuser",
			"password": "password123",
			"role":     "student",
			"nickname": "新用户",
			"email":    "new@test.com",
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行注册失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("注册应该成功, 错误: %s", output.Error)
	}
	if output.Data["user_id"] == nil {
		t.Fatal("应返回 user_id")
	}
	if output.Data["token"] == nil {
		t.Fatal("应返回 token")
	}
	if output.Data["role"] != "student" {
		t.Errorf("期望 role=student, 实际=%v", output.Data["role"])
	}
}

func TestAuthPlugin_Register_MergeData(t *testing.T) {
	plugin, _ := setupAuthPlugin(t)

	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action":       "register",
			"username":     "mergeuser",
			"password":     "password123",
			"role":         "teacher",
			"custom_field": "should_be_preserved",
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行注册失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("注册应该成功, 错误: %s", output.Error)
	}

	// 验证上游 Data 被 merge
	if output.Data["custom_field"] != "should_be_preserved" {
		t.Error("上游数据 custom_field 应该被保留")
	}
	if output.Data["action"] != "register" {
		t.Error("上游数据 action 应该被保留")
	}
}

func TestAuthPlugin_Register_DuplicateUsername(t *testing.T) {
	plugin, _ := setupAuthPlugin(t)

	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action":   "register",
			"username": "dupuser",
			"password": "password123",
		},
		Context: context.Background(),
	}

	// 第一次注册
	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("第一次注册失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("第一次注册应该成功: %s", output.Error)
	}

	// 第二次注册相同用户名
	output, err = plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if output.Success {
		t.Fatal("重复注册应该失败")
	}
	if output.Data["error_code"] != 40006 {
		t.Errorf("期望错误码 40006, 实际=%v", output.Data["error_code"])
	}
}

func TestAuthPlugin_Register_MissingParams(t *testing.T) {
	plugin, _ := setupAuthPlugin(t)

	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action":   "register",
			"username": "",
			"password": "",
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if output.Success {
		t.Fatal("缺少参数应该失败")
	}
	if output.Data["error_code"] != 40004 {
		t.Errorf("期望错误码 40004, 实际=%v", output.Data["error_code"])
	}
}

func TestAuthPlugin_Login(t *testing.T) {
	plugin, _ := setupAuthPlugin(t)

	// 先注册
	regInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":   "register",
			"username": "loginuser",
			"password": "mypassword",
			"role":     "teacher",
			"nickname": "登录用户",
		},
		Context: context.Background(),
	}
	_, err := plugin.Execute(context.Background(), regInput)
	if err != nil {
		t.Fatalf("注册失败: %v", err)
	}

	// 登录
	loginInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":   "login",
			"username": "loginuser",
			"password": "mypassword",
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), loginInput)
	if err != nil {
		t.Fatalf("登录执行失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("登录应该成功, 错误: %s", output.Error)
	}
	if output.Data["token"] == nil {
		t.Fatal("应返回 token")
	}
	if output.Data["role"] != "teacher" {
		t.Errorf("期望 role=teacher, 实际=%v", output.Data["role"])
	}
	if output.Data["nickname"] != "登录用户" {
		t.Errorf("期望 nickname=登录用户, 实际=%v", output.Data["nickname"])
	}
	if output.Data["expires_at"] == nil {
		t.Fatal("应返回 expires_at")
	}
}

func TestAuthPlugin_Login_WrongPassword(t *testing.T) {
	plugin, _ := setupAuthPlugin(t)

	// 先注册
	regInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":   "register",
			"username": "wrongpwuser",
			"password": "correctpassword",
		},
		Context: context.Background(),
	}
	_, _ = plugin.Execute(context.Background(), regInput)

	// 使用错误密码登录
	loginInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":   "login",
			"username": "wrongpwuser",
			"password": "wrongpassword",
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), loginInput)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if output.Success {
		t.Fatal("错误密码登录应该失败")
	}
	if output.Data["error_code"] != 40001 {
		t.Errorf("期望错误码 40001, 实际=%v", output.Data["error_code"])
	}
}

func TestAuthPlugin_Login_UserNotExist(t *testing.T) {
	plugin, _ := setupAuthPlugin(t)

	loginInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":   "login",
			"username": "nonexistuser",
			"password": "password",
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), loginInput)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if output.Success {
		t.Fatal("不存在的用户登录应该失败")
	}
	if output.Data["error_code"] != 40001 {
		t.Errorf("期望错误码 40001, 实际=%v", output.Data["error_code"])
	}
}

func TestAuthPlugin_Verify(t *testing.T) {
	plugin, _ := setupAuthPlugin(t)

	// 先注册获取 token
	regInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":   "register",
			"username": "verifyuser",
			"password": "password123",
			"role":     "student",
		},
		Context: context.Background(),
	}
	regOutput, _ := plugin.Execute(context.Background(), regInput)
	token := regOutput.Data["token"].(string)

	// 验证 token
	verifyInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action": "verify",
			"token":  token,
		},
		UserContext: &core.UserContext{},
		Context:    context.Background(),
	}

	output, err := plugin.Execute(context.Background(), verifyInput)
	if err != nil {
		t.Fatalf("验证执行失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("验证应该成功, 错误: %s", output.Error)
	}
	if output.Data["username"] != "verifyuser" {
		t.Errorf("期望 username=verifyuser, 实际=%v", output.Data["username"])
	}
	if output.Data["role"] != "student" {
		t.Errorf("期望 role=student, 实际=%v", output.Data["role"])
	}
}

func TestAuthPlugin_Verify_InvalidToken(t *testing.T) {
	plugin, _ := setupAuthPlugin(t)

	verifyInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action": "verify",
			"token":  "invalid-token",
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), verifyInput)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if output.Success {
		t.Fatal("无效 token 验证应该失败")
	}
	if output.Data["error_code"] != 40001 {
		t.Errorf("期望错误码 40001, 实际=%v", output.Data["error_code"])
	}
}

func TestAuthPlugin_Verify_MissingToken(t *testing.T) {
	plugin, _ := setupAuthPlugin(t)

	verifyInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action": "verify",
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), verifyInput)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if output.Success {
		t.Fatal("缺少 token 应该失败")
	}
}

func TestAuthPlugin_Refresh(t *testing.T) {
	plugin, _ := setupAuthPlugin(t)

	// 先注册获取 token
	regInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":   "register",
			"username": "refreshuser",
			"password": "password123",
			"role":     "teacher",
		},
		Context: context.Background(),
	}
	regOutput, _ := plugin.Execute(context.Background(), regInput)
	token := regOutput.Data["token"].(string)

	// 刷新 token
	refreshInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action": "refresh",
			"token":  token,
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), refreshInput)
	if err != nil {
		t.Fatalf("刷新执行失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("刷新应该成功, 错误: %s", output.Error)
	}
	newToken := output.Data["token"].(string)
	if newToken == "" {
		t.Fatal("应返回新 token")
	}
	if newToken == token {
		t.Log("注意: 新旧 token 可能相同（在同一秒内生成）")
	}
}

func TestAuthPlugin_InvalidAction(t *testing.T) {
	plugin, _ := setupAuthPlugin(t)

	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action": "unknown",
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if output.Success {
		t.Fatal("未知 action 应该失败")
	}
}

func TestAuthPlugin_MissingAction(t *testing.T) {
	plugin, _ := setupAuthPlugin(t)

	input := &core.PluginInput{
		Data:    map[string]interface{}{},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if output.Success {
		t.Fatal("缺少 action 应该失败")
	}
	if output.Data["error_code"] != 40004 {
		t.Errorf("期望错误码 40004, 实际=%v", output.Data["error_code"])
	}
}

// ==================== 微信登录辅助函数 ====================

// setupAuthPluginWithMock 创建使用 MockWxClient 的认证插件
func setupAuthPluginWithMock(t *testing.T) (*AuthPlugin, *database.Database) {
	t.Helper()
	// 设置 WX_MODE=mock 确保使用 MockWxClient
	os.Setenv("WX_MODE", "mock")
	t.Cleanup(func() {
		os.Unsetenv("WX_MODE")
	})
	return setupAuthPlugin(t)
}

// ==================== 微信登录测试 ====================

func TestWxLogin_NewUser(t *testing.T) {
	plugin, _ := setupAuthPluginWithMock(t)

	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action": "wx-login",
			"code":   "test_code_001",
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行微信登录失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("新用户微信登录应该成功, 错误: %s", output.Error)
	}

	// 验证 is_new_user=true
	isNewUser, ok := output.Data["is_new_user"].(bool)
	if !ok {
		t.Fatal("应返回 is_new_user 字段")
	}
	if !isNewUser {
		t.Error("新用户首次登录 is_new_user 应为 true")
	}

	// 验证 role 为空（新用户未补全信息）
	role, _ := output.Data["role"].(string)
	if role != "" {
		t.Errorf("新用户 role 应为空, 实际=%q", role)
	}

	// 验证返回了 token
	if output.Data["token"] == nil || output.Data["token"] == "" {
		t.Fatal("应返回 token")
	}

	// 验证返回了 user_id
	if output.Data["user_id"] == nil {
		t.Fatal("应返回 user_id")
	}
}

func TestWxLogin_ExistingUser(t *testing.T) {
	plugin, _ := setupAuthPluginWithMock(t)

	// 第一次登录（创建新用户）
	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action": "wx-login",
			"code":   "existing_user_code",
		},
		Context: context.Background(),
	}

	firstOutput, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("第一次微信登录失败: %v", err)
	}
	if !firstOutput.Success {
		t.Fatalf("第一次微信登录应该成功, 错误: %s", firstOutput.Error)
	}
	firstUserID := firstOutput.Data["user_id"]

	// 补全用户信息
	completeInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":   "complete-profile",
			"user_id":  firstUserID,
			"role":     "student",
			"nickname": "测试学生",
		},
		Context: context.Background(),
	}
	completeOutput, err := plugin.Execute(context.Background(), completeInput)
	if err != nil {
		t.Fatalf("补全信息失败: %v", err)
	}
	if !completeOutput.Success {
		t.Fatalf("补全信息应该成功, 错误: %s", completeOutput.Error)
	}

	// 第二次登录（已有用户）
	secondOutput, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("第二次微信登录失败: %v", err)
	}
	if !secondOutput.Success {
		t.Fatalf("第二次微信登录应该成功, 错误: %s", secondOutput.Error)
	}

	// 验证 is_new_user=false
	isNewUser, ok := secondOutput.Data["is_new_user"].(bool)
	if !ok {
		t.Fatal("应返回 is_new_user 字段")
	}
	if isNewUser {
		t.Error("已有用户再次登录 is_new_user 应为 false")
	}

	// 验证 role 和 nickname 正确
	if secondOutput.Data["role"] != "student" {
		t.Errorf("期望 role=student, 实际=%v", secondOutput.Data["role"])
	}
	if secondOutput.Data["nickname"] != "测试学生" {
		t.Errorf("期望 nickname=测试学生, 实际=%v", secondOutput.Data["nickname"])
	}
}

func TestWxLogin_EmptyCode(t *testing.T) {
	plugin, _ := setupAuthPluginWithMock(t)

	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action": "wx-login",
			"code":   "",
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if output.Success {
		t.Fatal("空 code 微信登录应该失败")
	}

	// 验证错误码 40004
	errorCode, ok := output.Data["error_code"].(int)
	if !ok {
		t.Fatalf("应返回 error_code, 实际 Data=%v", output.Data)
	}
	if errorCode != 40004 {
		t.Errorf("期望错误码 40004, 实际=%d", errorCode)
	}
}

func TestWxLogin_MockClient(t *testing.T) {
	// 验证 MockWxClient 返回 mock_openid_{code}
	mockClient := &MockWxClient{}

	testCases := []struct {
		code           string
		expectedOpenID string
	}{
		{"abc123", "mock_openid_abc123"},
		{"test", "mock_openid_test"},
		{"", "mock_openid_"},
	}

	for _, tc := range testCases {
		result, err := mockClient.Code2Session(tc.code)
		if err != nil {
			t.Fatalf("MockWxClient.Code2Session(%q) 返回错误: %v", tc.code, err)
		}
		if result.OpenID != tc.expectedOpenID {
			t.Errorf("MockWxClient.Code2Session(%q) 期望 OpenID=%q, 实际=%q",
				tc.code, tc.expectedOpenID, result.OpenID)
		}
	}
}

// ==================== 补全信息测试 ====================

func TestCompleteProfile_Success(t *testing.T) {
	plugin, _ := setupAuthPluginWithMock(t)

	// 先通过微信登录创建新用户
	wxInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action": "wx-login",
			"code":   "profile_test_code",
		},
		Context: context.Background(),
	}
	wxOutput, err := plugin.Execute(context.Background(), wxInput)
	if err != nil {
		t.Fatalf("微信登录失败: %v", err)
	}
	if !wxOutput.Success {
		t.Fatalf("微信登录应该成功: %s", wxOutput.Error)
	}
	userID := wxOutput.Data["user_id"]

	// 补全角色和昵称
	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action":      "complete-profile",
			"user_id":     userID,
			"role":        "teacher",
			"nickname":    "李老师",
			"school":      "测试大学",
			"description": "一位优秀的测试教师",
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("补全信息执行失败: %v", err)
	}
	if !output.Success {
		t.Fatalf("补全信息应该成功, 错误: %s", output.Error)
	}

	// 验证返回数据
	if output.Data["role"] != "teacher" {
		t.Errorf("期望 role=teacher, 实际=%v", output.Data["role"])
	}
	if output.Data["nickname"] != "李老师" {
		t.Errorf("期望 nickname=李老师, 实际=%v", output.Data["nickname"])
	}
}

func TestCompleteProfile_InvalidRole(t *testing.T) {
	plugin, _ := setupAuthPluginWithMock(t)

	// 先创建新用户
	wxInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action": "wx-login",
			"code":   "invalid_role_code",
		},
		Context: context.Background(),
	}
	wxOutput, _ := plugin.Execute(context.Background(), wxInput)
	userID := wxOutput.Data["user_id"]

	// 使用无效角色
	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action":   "complete-profile",
			"user_id":  userID,
			"role":     "admin", // 不是 teacher/student
			"nickname": "测试用户",
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if output.Success {
		t.Fatal("无效角色应该失败")
	}
	if output.Data["error_code"] != 40004 {
		t.Errorf("期望错误码 40004, 实际=%v", output.Data["error_code"])
	}
}

func TestCompleteProfile_EmptyNickname(t *testing.T) {
	plugin, _ := setupAuthPluginWithMock(t)

	// 先创建新用户
	wxInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action": "wx-login",
			"code":   "empty_nickname_code",
		},
		Context: context.Background(),
	}
	wxOutput, _ := plugin.Execute(context.Background(), wxInput)
	userID := wxOutput.Data["user_id"]

	// nickname 为空
	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action":   "complete-profile",
			"user_id":  userID,
			"role":     "student",
			"nickname": "",
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if output.Success {
		t.Fatal("空昵称应该失败")
	}
	if output.Data["error_code"] != 40004 {
		t.Errorf("期望错误码 40004, 实际=%v", output.Data["error_code"])
	}
}

func TestCompleteProfile_AlreadyCompleted(t *testing.T) {
	plugin, _ := setupAuthPluginWithMock(t)

	// 先创建新用户
	wxInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action": "wx-login",
			"code":   "already_completed_code",
		},
		Context: context.Background(),
	}
	wxOutput, _ := plugin.Execute(context.Background(), wxInput)
	userID := wxOutput.Data["user_id"]

	// 第一次补全
	completeInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action":   "complete-profile",
			"user_id":  userID,
			"role":     "student",
			"nickname": "小明",
		},
		Context: context.Background(),
	}
	firstOutput, err := plugin.Execute(context.Background(), completeInput)
	if err != nil {
		t.Fatalf("第一次补全执行失败: %v", err)
	}
	if !firstOutput.Success {
		t.Fatalf("第一次补全应该成功, 错误: %s", firstOutput.Error)
	}

	// 第二次补全（重复调用）
	completeInput2 := &core.PluginInput{
		Data: map[string]interface{}{
			"action":   "complete-profile",
			"user_id":  userID,
			"role":     "teacher",
			"nickname": "另一个名字",
		},
		Context: context.Background(),
	}
	secondOutput, err := plugin.Execute(context.Background(), completeInput2)
	if err != nil {
		t.Fatalf("第二次补全执行失败: %v", err)
	}
	if secondOutput.Success {
		t.Fatal("已完善用户重复补全应该失败")
	}
	if secondOutput.Data["error_code"] != 40004 {
		t.Errorf("期望错误码 40004, 实际=%v", secondOutput.Data["error_code"])
	}
}

func TestCompleteProfile_NicknameTooLong(t *testing.T) {
	plugin, _ := setupAuthPluginWithMock(t)

	// 先创建新用户
	wxInput := &core.PluginInput{
		Data: map[string]interface{}{
			"action": "wx-login",
			"code":   "long_nickname_code",
		},
		Context: context.Background(),
	}
	wxOutput, _ := plugin.Execute(context.Background(), wxInput)
	userID := wxOutput.Data["user_id"]

	// nickname 超过 20 个字符（使用中文，每个中文算1个rune）
	longNickname := "这是一个超过二十个字符的昵称用来测试长度限制的功能"
	input := &core.PluginInput{
		Data: map[string]interface{}{
			"action":   "complete-profile",
			"user_id":  userID,
			"role":     "student",
			"nickname": longNickname,
		},
		Context: context.Background(),
	}

	output, err := plugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if output.Success {
		t.Fatal("超长昵称应该失败")
	}
	if output.Data["error_code"] != 40004 {
		t.Errorf("期望错误码 40004, 实际=%v", output.Data["error_code"])
	}
}
