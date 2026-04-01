package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"digital-twin/src/backend/api"
	"digital-twin/src/harness/manager"
)

// 全局变量，保存测试服务器和 token
var (
	ts           *httptest.Server
	mgr          *manager.HarnessManager
	teacherToken string
	studentToken string
	adminToken   string
	teacherID    float64
	studentID    float64
	adminID      float64
	documentID   float64
	dbPath       string
)

// getProjectRoot 获取项目根目录
func getProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	// tests/integration/integration_test.go -> 项目根目录
	return filepath.Join(filepath.Dir(filename), "..", "..")
}

// TestMain 测试入口，设置和清理环境
func TestMain(m *testing.M) {
	projectRoot := getProjectRoot()

	// 创建临时数据库文件
	tmpDir := os.TempDir()
	dbPath = filepath.Join(tmpDir, "test_integration.db")

	// 清理可能存在的旧数据库
	os.Remove(dbPath)
	os.Remove(dbPath + "-wal")
	os.Remove(dbPath + "-shm")

	// 设置环境变量
	os.Setenv("JWT_SECRET", "test-secret-key-for-integration-testing")
	os.Setenv("LLM_MODE", "mock")
	os.Setenv("WX_MODE", "mock")
	os.Setenv("DB_PATH", dbPath)
	os.Setenv("OPENAI_API_KEY", "test-key")
	os.Setenv("OPENAI_BASE_URL", "http://localhost:11434/v1")
	os.Setenv("LLM_MODEL", "qwen-turbo")
	os.Setenv("VECTOR_DB_MODE", "memory")

	// 创建 HarnessManager
	configPath := filepath.Join(projectRoot, "configs", "harness.yaml")
	var err error
	mgr, err = manager.NewHarnessManager(configPath)
	if err != nil {
		fmt.Printf("创建 HarnessManager 失败: %v\n", err)
		os.Exit(1)
	}

	// 启动管理器
	if err := mgr.Start(); err != nil {
		fmt.Printf("启动 HarnessManager 失败: %v\n", err)
		os.Exit(1)
	}

	// 设置路由并启动测试服务器
	router := api.SetupRouter(mgr)
	ts = httptest.NewServer(router)

	// 运行测试
	code := m.Run()

	// 清理
	ts.Close()
	mgr.Stop()
	os.Remove(dbPath)
	os.Remove(dbPath + "-wal")
	os.Remove(dbPath + "-shm")

	os.Exit(code)
}

// apiResponse 通用 API 响应结构
type apiResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

// apiResponseRaw 通用 API 响应结构（Data 为 raw JSON）
type apiResponseRaw struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// doRequest 发送 HTTP 请求
func doRequest(method, path string, body interface{}, token string) (*http.Response, []byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, ts.URL+path, reqBody)
	if err != nil {
		return nil, nil, fmt.Errorf("创建请求失败: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	return resp, respBody, nil
}

// parseResponse 解析 JSON 响应
func parseResponse(body []byte) (*apiResponse, error) {
	var resp apiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, body: %s", err, string(body))
	}
	return &resp, nil
}

// ======================== IT-01: 用户注册（教师） ========================
func TestIT01_RegisterTeacher(t *testing.T) {
	reqBody := map[string]interface{}{
		"username": "teacher_wang",
		"password": "123456",
		"role":     "teacher",
		"nickname": "王老师",
	}

	resp, body, err := doRequest("POST", "/api/auth/register", reqBody, "")
	if err != nil {
		t.Fatalf("IT-01 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-01 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-01 解析响应失败: %v", err)
	}

	if apiResp.Code != 0 {
		t.Fatalf("IT-01 业务码错误: 期望 0, 实际 %d, message: %s", apiResp.Code, apiResp.Message)
	}

	// 验证返回字段
	userIDVal, ok := apiResp.Data["user_id"]
	if !ok {
		t.Fatal("IT-01 响应缺少 user_id 字段")
	}
	userIDFloat, ok := userIDVal.(float64)
	if !ok || userIDFloat <= 0 {
		t.Fatalf("IT-01 user_id 无效: %v", userIDVal)
	}
	teacherID = userIDFloat

	tokenVal, ok := apiResp.Data["token"]
	if !ok || tokenVal == "" {
		t.Fatal("IT-01 响应缺少 token 或 token 为空")
	}
	teacherToken = tokenVal.(string)

	t.Logf("IT-01 通过: 教师注册成功, user_id=%v, token长度=%d", teacherID, len(teacherToken))
}

// ======================== IT-02: 用户注册（学生） ========================
func TestIT02_RegisterStudent(t *testing.T) {
	reqBody := map[string]interface{}{
		"username": "student_li",
		"password": "123456",
		"role":     "student",
		"nickname": "小李",
	}

	resp, body, err := doRequest("POST", "/api/auth/register", reqBody, "")
	if err != nil {
		t.Fatalf("IT-02 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-02 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-02 解析响应失败: %v", err)
	}

	if apiResp.Code != 0 {
		t.Fatalf("IT-02 业务码错误: 期望 0, 实际 %d, message: %s", apiResp.Code, apiResp.Message)
	}

	userIDVal, ok := apiResp.Data["user_id"]
	if !ok {
		t.Fatal("IT-02 响应缺少 user_id 字段")
	}
	userIDFloat, ok := userIDVal.(float64)
	if !ok || userIDFloat <= 0 {
		t.Fatalf("IT-02 user_id 无效: %v", userIDVal)
	}
	studentID = userIDFloat

	tokenVal, ok := apiResp.Data["token"]
	if !ok || tokenVal == "" {
		t.Fatal("IT-02 响应缺少 token 或 token 为空")
	}
	studentToken = tokenVal.(string)

	t.Logf("IT-02 通过: 学生注册成功, user_id=%v, token长度=%d", studentID, len(studentToken))
}

// ======================== IT-03: 重复注册 ========================
func TestIT03_DuplicateRegister(t *testing.T) {
	reqBody := map[string]interface{}{
		"username": "teacher_wang",
		"password": "123456",
		"role":     "teacher",
		"nickname": "王老师",
	}

	resp, body, err := doRequest("POST", "/api/auth/register", reqBody, "")
	if err != nil {
		t.Fatalf("IT-03 请求失败: %v", err)
	}

	// 解析响应
	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-03 解析响应失败: %v", err)
	}

	if apiResp.Code != 40006 {
		t.Fatalf("IT-03 业务码错误: 期望 40006, 实际 %d, HTTP状态=%d, message: %s, body: %s",
			apiResp.Code, resp.StatusCode, apiResp.Message, string(body))
	}

	t.Logf("IT-03 通过: 重复注册正确返回 40006, message: %s", apiResp.Message)
}

// ======================== IT-04: 用户登录 ========================
func TestIT04_Login(t *testing.T) {
	reqBody := map[string]interface{}{
		"username": "teacher_wang",
		"password": "123456",
	}

	resp, body, err := doRequest("POST", "/api/auth/login", reqBody, "")
	if err != nil {
		t.Fatalf("IT-04 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-04 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-04 解析响应失败: %v", err)
	}

	if apiResp.Code != 0 {
		t.Fatalf("IT-04 业务码错误: 期望 0, 实际 %d, message: %s", apiResp.Code, apiResp.Message)
	}

	tokenVal, ok := apiResp.Data["token"]
	if !ok || tokenVal == "" {
		t.Fatal("IT-04 响应缺少 token 或 token 为空")
	}

	// 更新 teacherToken 为登录获取的最新 token
	teacherToken = tokenVal.(string)

	t.Logf("IT-04 通过: 登录成功, token长度=%d", len(teacherToken))
}

// ======================== IT-05: 登录密码错误 ========================
func TestIT05_LoginWrongPassword(t *testing.T) {
	reqBody := map[string]interface{}{
		"username": "teacher_wang",
		"password": "wrong_password",
	}

	resp, body, err := doRequest("POST", "/api/auth/login", reqBody, "")
	if err != nil {
		t.Fatalf("IT-05 请求失败: %v", err)
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-05 解析响应失败: %v", err)
	}

	if apiResp.Code != 40001 {
		t.Fatalf("IT-05 业务码错误: 期望 40001, 实际 %d, HTTP状态=%d, message: %s",
			apiResp.Code, resp.StatusCode, apiResp.Message)
	}

	t.Logf("IT-05 通过: 密码错误正确返回 40001, message: %s", apiResp.Message)
}

// ======================== IT-06: 无效 Token 访问 ========================
func TestIT06_InvalidToken(t *testing.T) {
	reqBody := map[string]interface{}{
		"message":    "你好",
		"teacher_id": 1,
	}

	resp, body, err := doRequest("POST", "/api/chat", reqBody, "invalid-token-12345")
	if err != nil {
		t.Fatalf("IT-06 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("IT-06 HTTP 状态码错误: 期望 401, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	t.Logf("IT-06 通过: 无效 token 正确返回 401")
}

// ======================== IT-07: 添加知识文档（教师） ========================
func TestIT07_AddDocument(t *testing.T) {
	if teacherToken == "" {
		t.Skip("IT-07 跳过: 教师 token 未获取")
	}

	reqBody := map[string]interface{}{
		"title":   "牛顿运动定律",
		"content": "牛顿第一定律（惯性定律）：一切物体在没有受到外力作用的时候，总保持静止状态或匀速直线运动状态。牛顿第二定律：物体加速度的大小跟作用力成正比，跟物体的质量成反比。牛顿第三定律：两个物体之间的作用力和反作用力总是大小相等，方向相反，作用在同一条直线上。",
		"tags":    "物理,力学",
	}

	resp, body, err := doRequest("POST", "/api/documents", reqBody, teacherToken)
	if err != nil {
		t.Fatalf("IT-07 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-07 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-07 解析响应失败: %v", err)
	}

	if apiResp.Code != 0 {
		t.Fatalf("IT-07 业务码错误: 期望 0, 实际 %d, message: %s", apiResp.Code, apiResp.Message)
	}

	docIDVal, ok := apiResp.Data["document_id"]
	if !ok {
		t.Fatal("IT-07 响应缺少 document_id 字段")
	}
	docIDFloat, ok := docIDVal.(float64)
	if !ok || docIDFloat <= 0 {
		t.Fatalf("IT-07 document_id 无效: %v", docIDVal)
	}
	documentID = docIDFloat

	t.Logf("IT-07 通过: 文档添加成功, document_id=%v", documentID)
}

// ======================== IT-08: 学生添加文档（权限不足） ========================
func TestIT08_StudentAddDocument(t *testing.T) {
	if studentToken == "" {
		t.Skip("IT-08 跳过: 学生 token 未获取")
	}

	reqBody := map[string]interface{}{
		"title":   "测试文档",
		"content": "测试内容",
		"tags":    "测试",
	}

	resp, body, err := doRequest("POST", "/api/documents", reqBody, studentToken)
	if err != nil {
		t.Fatalf("IT-08 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("IT-08 HTTP 状态码错误: 期望 403, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	t.Logf("IT-08 通过: 学生添加文档正确返回 403")
}

// ======================== IT-09: 获取文档列表 ========================
func TestIT09_GetDocuments(t *testing.T) {
	if teacherToken == "" {
		t.Skip("IT-09 跳过: 教师 token 未获取")
	}

	resp, body, err := doRequest("GET", "/api/documents?page=1&page_size=20", nil, teacherToken)
	if err != nil {
		t.Fatalf("IT-09 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-09 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-09 解析响应失败: %v", err)
	}

	if apiResp.Code != 0 {
		t.Fatalf("IT-09 业务码错误: 期望 0, 实际 %d, message: %s", apiResp.Code, apiResp.Message)
	}

	// 验证 items 字段存在（分页响应格式）
	itemsVal, ok := apiResp.Data["items"]
	if !ok {
		t.Fatal("IT-09 响应缺少 items 字段")
	}

	// items 可能是 nil（空列表）或数组
	if itemsVal != nil {
		items, ok := itemsVal.([]interface{})
		if !ok {
			t.Fatalf("IT-09 items 不是数组类型: %T", itemsVal)
		}
		t.Logf("IT-09 通过: 获取文档列表成功, 文档数量=%d", len(items))
	} else {
		t.Logf("IT-09 通过: 获取文档列表成功, 文档数量=0")
	}
}

// ======================== IT-10: 删除文档 ========================
func TestIT10_DeleteDocument(t *testing.T) {
	if teacherToken == "" {
		t.Skip("IT-10 跳过: 教师 token 未获取")
	}
	if documentID <= 0 {
		t.Skip("IT-10 跳过: 文档 ID 未获取")
	}

	path := fmt.Sprintf("/api/documents/%d", int(documentID))
	resp, body, err := doRequest("DELETE", path, nil, teacherToken)
	if err != nil {
		t.Fatalf("IT-10 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-10 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-10 解析响应失败: %v", err)
	}

	if apiResp.Code != 0 {
		t.Fatalf("IT-10 业务码错误: 期望 0, 实际 %d, message: %s", apiResp.Code, apiResp.Message)
	}

	deletedVal, ok := apiResp.Data["deleted"]
	if !ok {
		t.Fatal("IT-10 响应缺少 deleted 字段")
	}
	deleted, ok := deletedVal.(bool)
	if !ok || !deleted {
		t.Fatalf("IT-10 deleted 值错误: 期望 true, 实际 %v", deletedVal)
	}

	t.Logf("IT-10 通过: 文档删除成功, document_id=%v", documentID)
}

// ======================== IT-10b: 建立师生关系（教师邀请学生） ========================
func TestIT10b_EstablishRelation(t *testing.T) {
	if teacherToken == "" || studentID == 0 {
		t.Skip("IT-10b 跳过: 教师 token 或学生 ID 未获取")
	}

	reqBody := map[string]interface{}{
		"student_id": int(studentID),
	}

	resp, body, err := doRequest("POST", "/api/relations/invite", reqBody, teacherToken)
	if err != nil {
		t.Fatalf("IT-10b 请求失败: %v", err)
	}

	// 允许 200（新建）或 409（已存在）
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
		t.Fatalf("IT-10b HTTP 状态码错误: 期望 200 或 409, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	t.Logf("IT-10b 通过: 师生关系建立成功, status=%d", resp.StatusCode)
}

// ======================== IT-11: 学生对话（全链路） ========================
func TestIT11_StudentChat(t *testing.T) {
	if studentToken == "" {
		t.Skip("IT-11 跳过: 学生 token 未获取")
	}

	reqBody := map[string]interface{}{
		"message":    "什么是牛顿第一定律?",
		"teacher_id": int(teacherID),
	}

	resp, body, err := doRequest("POST", "/api/chat", reqBody, studentToken)
	if err != nil {
		t.Fatalf("IT-11 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-11 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-11 解析响应失败: %v", err)
	}

	if apiResp.Code != 0 {
		t.Fatalf("IT-11 业务码错误: 期望 0, 实际 %d, message: %s, body: %s", apiResp.Code, apiResp.Message, string(body))
	}

	replyVal, ok := apiResp.Data["reply"]
	if !ok || replyVal == "" {
		t.Fatalf("IT-11 响应缺少 reply 或 reply 为空, data: %v", apiResp.Data)
	}

	t.Logf("IT-11 通过: 对话成功, reply: %v", replyVal)
}

// ======================== IT-12: 对话历史查询 ========================
func TestIT12_GetConversations(t *testing.T) {
	if studentToken == "" {
		t.Skip("IT-12 跳过: 学生 token 未获取")
	}

	path := fmt.Sprintf("/api/conversations?teacher_id=%d&page=1&page_size=20", int(teacherID))
	resp, body, err := doRequest("GET", path, nil, studentToken)
	if err != nil {
		t.Fatalf("IT-12 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-12 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-12 解析响应失败: %v", err)
	}

	if apiResp.Code != 0 {
		t.Fatalf("IT-12 业务码错误: 期望 0, 实际 %d, message: %s", apiResp.Code, apiResp.Message)
	}

	// 验证 items 字段存在
	_, ok := apiResp.Data["items"]
	if !ok {
		t.Fatal("IT-12 响应缺少 items 字段")
	}

	t.Logf("IT-12 通过: 对话历史查询成功")
}

// ======================== IT-13: 记忆列表查询 ========================
func TestIT13_GetMemories(t *testing.T) {
	if studentToken == "" {
		t.Skip("IT-13 跳过: 学生 token 未获取")
	}

	path := fmt.Sprintf("/api/memories?teacher_id=%d&page=1&page_size=20", int(teacherID))
	resp, body, err := doRequest("GET", path, nil, studentToken)
	if err != nil {
		t.Fatalf("IT-13 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-13 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-13 解析响应失败: %v", err)
	}

	if apiResp.Code != 0 {
		t.Fatalf("IT-13 业务码错误: 期望 0, 实际 %d, message: %s", apiResp.Code, apiResp.Message)
	}

	t.Logf("IT-13 通过: 记忆列表查询成功")
}

// ======================== IT-14: 健康检查 ========================
func TestIT14_HealthCheck(t *testing.T) {
	resp, body, err := doRequest("GET", "/api/system/health", nil, "")
	if err != nil {
		t.Fatalf("IT-14 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-14 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-14 解析响应失败: %v", err)
	}

	if apiResp.Code != 0 {
		t.Fatalf("IT-14 业务码错误: 期望 0, 实际 %d, message: %s", apiResp.Code, apiResp.Message)
	}

	statusVal, ok := apiResp.Data["status"]
	if !ok || statusVal == "" {
		t.Fatal("IT-14 响应缺少 status 字段")
	}

	t.Logf("IT-14 通过: 健康检查成功, status=%v", statusVal)
}

// ======================== IT-15: 插件列表（admin） ========================
func TestIT15_GetPlugins(t *testing.T) {
	// 先注册 admin 用户
	if adminToken == "" {
		reqBody := map[string]interface{}{
			"username": "admin_user",
			"password": "admin123",
			"role":     "admin",
			"nickname": "管理员",
		}

		_, body, err := doRequest("POST", "/api/auth/register", reqBody, "")
		if err != nil {
			t.Fatalf("IT-15 注册 admin 失败: %v", err)
		}

		apiResp, err := parseResponse(body)
		if err != nil {
			t.Fatalf("IT-15 解析 admin 注册响应失败: %v", err)
		}

		if apiResp.Code != 0 {
			t.Fatalf("IT-15 admin 注册业务码错误: %d, message: %s", apiResp.Code, apiResp.Message)
		}

		tokenVal, ok := apiResp.Data["token"]
		if !ok || tokenVal == "" {
			t.Fatal("IT-15 admin 注册响应缺少 token")
		}
		adminToken = tokenVal.(string)

		if idVal, ok := apiResp.Data["user_id"]; ok {
			adminID, _ = idVal.(float64)
		}
	}

	resp, body, err := doRequest("GET", "/api/system/plugins", nil, adminToken)
	if err != nil {
		t.Fatalf("IT-15 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-15 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	// HandleGetPlugins 返回的是 []core.PluginInfo 数组，
	// 被 Success 包装后为 {"code":0, "message":"success", "data": [...]}
	var rawResp apiResponseRaw
	if err := json.Unmarshal(body, &rawResp); err != nil {
		t.Fatalf("IT-15 解析响应失败: %v", err)
	}

	if rawResp.Code != 0 {
		t.Fatalf("IT-15 业务码错误: 期望 0, 实际 %d, message: %s", rawResp.Code, rawResp.Message)
	}

	// data 可能是数组或对象
	var plugins []interface{}
	if err := json.Unmarshal(rawResp.Data, &plugins); err != nil {
		// 可能是 {"plugins": [...]} 格式
		var dataObj map[string]interface{}
		if err2 := json.Unmarshal(rawResp.Data, &dataObj); err2 != nil {
			t.Fatalf("IT-15 解析 data 失败: 数组解析=%v, 对象解析=%v, raw=%s", err, err2, string(rawResp.Data))
		}
		if pluginsVal, ok := dataObj["plugins"]; ok {
			if arr, ok := pluginsVal.([]interface{}); ok {
				plugins = arr
			}
		}
		if plugins == nil {
			t.Fatalf("IT-15 无法解析插件列表, data: %s", string(rawResp.Data))
		}
	}

	if len(plugins) == 0 {
		t.Fatal("IT-15 插件列表为空")
	}

	t.Logf("IT-15 通过: 插件列表获取成功, 插件数量=%d", len(plugins))
}

// ======================== IT-16: 管道列表（admin） ========================
func TestIT16_GetPipelines(t *testing.T) {
	if adminToken == "" {
		t.Skip("IT-16 跳过: admin token 未获取")
	}

	resp, body, err := doRequest("GET", "/api/system/pipelines", nil, adminToken)
	if err != nil {
		t.Fatalf("IT-16 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-16 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	// HandleGetPipelines 返回的是 []string，
	// 被 Success 包装后为 {"code":0, "message":"success", "data": [...]}
	var rawResp apiResponseRaw
	if err := json.Unmarshal(body, &rawResp); err != nil {
		t.Fatalf("IT-16 解析响应失败: %v", err)
	}

	if rawResp.Code != 0 {
		t.Fatalf("IT-16 业务码错误: 期望 0, 实际 %d, message: %s", rawResp.Code, rawResp.Message)
	}

	// data 可能是数组或对象
	var pipelines []interface{}
	if err := json.Unmarshal(rawResp.Data, &pipelines); err != nil {
		// 可能是 {"pipelines": [...]} 格式
		var dataObj map[string]interface{}
		if err2 := json.Unmarshal(rawResp.Data, &dataObj); err2 != nil {
			t.Fatalf("IT-16 解析 data 失败: 数组解析=%v, 对象解析=%v, raw=%s", err, err2, string(rawResp.Data))
		}
		if pipelinesVal, ok := dataObj["pipelines"]; ok {
			if arr, ok := pipelinesVal.([]interface{}); ok {
				pipelines = arr
			}
		}
		if pipelines == nil {
			t.Fatalf("IT-16 无法解析管道列表, data: %s", string(rawResp.Data))
		}
	}

	if len(pipelines) == 0 {
		t.Fatal("IT-16 管道列表为空")
	}

	t.Logf("IT-16 通过: 管道列表获取成功, 管道数量=%d", len(pipelines))
}

// ======================== IT-17: 令牌刷新 ========================
func TestIT17_RefreshToken(t *testing.T) {
	if teacherToken == "" {
		t.Skip("IT-17 跳过: 教师 token 未获取")
	}

	resp, body, err := doRequest("POST", "/api/auth/refresh", nil, teacherToken)
	if err != nil {
		t.Fatalf("IT-17 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("IT-17 HTTP 状态码错误: 期望 200, 实际 %d, body: %s", resp.StatusCode, string(body))
	}

	apiResp, err := parseResponse(body)
	if err != nil {
		t.Fatalf("IT-17 解析响应失败: %v", err)
	}

	if apiResp.Code != 0 {
		t.Fatalf("IT-17 业务码错误: 期望 0, 实际 %d, message: %s", apiResp.Code, apiResp.Message)
	}

	tokenVal, ok := apiResp.Data["token"]
	if !ok || tokenVal == "" {
		t.Fatal("IT-17 响应缺少 token 或 token 为空")
	}

	newToken := tokenVal.(string)
	if newToken == teacherToken {
		t.Log("IT-17 警告: 新 token 与旧 token 相同（可能因为时间间隔太短）")
	}

	t.Logf("IT-17 通过: 令牌刷新成功, 新 token 长度=%d", len(newToken))
}
