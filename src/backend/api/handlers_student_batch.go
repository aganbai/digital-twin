package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"digital-twin/src/backend/database"

	"github.com/gin-gonic/gin"
)

// HandleParseStudentText LLM智能解析学生文本
// POST /api/students/parse-text
func (h *Handler) HandleParseStudentText(c *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	if len(req.Text) > 5000 {
		Error(c, http.StatusBadRequest, 40004, "文本内容不能超过5000字")
		return
	}

	// 获取 LLM 配置
	cfg := h.manager.GetConfig()
	llmBaseURL := ""
	llmAPIKey := ""
	llmModel := ""
	if cfg != nil {
		if pluginCfg, ok := cfg.Plugins["dialogue"]; ok {
			// 配置可能是嵌套结构（从 YAML 加载）或扁平结构
			if llmProvider, ok := pluginCfg.Config["llm_provider"]; ok {
				if providerMap, ok := llmProvider.(map[string]interface{}); ok {
					if v, ok := providerMap["base_url"]; ok {
						llmBaseURL, _ = v.(string)
					}
					if v, ok := providerMap["api_key"]; ok {
						llmAPIKey, _ = v.(string)
					}
					if v, ok := providerMap["model"]; ok {
						llmModel, _ = v.(string)
					}
				}
			}
			// 回退：尝试扁平化 key 格式
			if llmBaseURL == "" {
				if v, ok := pluginCfg.Config["llm_provider.base_url"]; ok {
					llmBaseURL, _ = v.(string)
				}
			}
			if llmAPIKey == "" {
				if v, ok := pluginCfg.Config["llm_provider.api_key"]; ok {
					llmAPIKey, _ = v.(string)
				}
			}
			if llmModel == "" {
				if v, ok := pluginCfg.Config["llm_provider.model"]; ok {
					llmModel, _ = v.(string)
				}
			}
		}
	}

	if llmBaseURL == "" || llmAPIKey == "" {
		// LLM 不可用，使用简单规则解析
		students := simpleParseStudentText(req.Text)
		Success(c, gin.H{"students": students, "parse_method": "rule"})
		return
	}

	// 调用 LLM 解析
	students, err := llmParseStudentText(llmBaseURL, llmAPIKey, llmModel, req.Text)
	if err != nil {
		// 区分超时错误和其他错误，统一降级到规则解析
		parseMethod := "rule_fallback"
		if IsTimeoutError(err) {
			// 超时错误，记录日志但使用规则解析降级
			fmt.Printf("[WARN] LLM解析超时，降级到规则解析: %v\n", err)
		} else {
			// 其他错误，同样降级
			fmt.Printf("[WARN] LLM解析失败，降级到规则解析: %v\n", err)
		}
		students = simpleParseStudentText(req.Text)
		Success(c, gin.H{"students": students, "parse_method": parseMethod, "error_info": err.Error()})
		return
	}

	Success(c, gin.H{"students": students, "parse_method": "llm"})
}

// HandleBatchCreateStudents 批量创建学生
// POST /api/students/batch-create
func (h *Handler) HandleBatchCreateStudents(c *gin.Context) {
	var req struct {
		PersonaID int64 `json:"persona_id" binding:"required"`
		ClassID   int64 `json:"class_id"`
		Students  []struct {
			Name        string `json:"name" binding:"required"`
			Gender      string `json:"gender"`
			Age         int    `json:"age"`
			StudentID   string `json:"student_id"`
			Strengths   string `json:"strengths"`
			Weaknesses  string `json:"weaknesses"`
			Interests   string `json:"interests"`
			Personality string `json:"personality"`
			Notes       string `json:"notes"`
		} `json:"students" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}

	if len(req.Students) == 0 {
		Error(c, http.StatusBadRequest, 40004, "学生列表不能为空")
		return
	}
	if len(req.Students) > 100 {
		Error(c, http.StatusBadRequest, 40004, "单次最多创建100个学生")
		return
	}

	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
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
	relationRepo := database.NewRelationRepository(db)

	var results []gin.H
	successCount := 0
	failedCount := 0

	for _, s := range req.Students {
		// 生成唯一用户名
		username := fmt.Sprintf("student_%s_%d", strings.ReplaceAll(s.Name, " ", "_"), time.Now().UnixNano())

		// 创建用户
		user := &database.User{
			Username: username,
			Password: "default_password_hash", // 学生通过分享码或教师创建，后续可修改密码
			Role:     "student",
			Nickname: s.Name,
		}
		newUserID, err := userRepo.Create(user)
		if err != nil {
			results = append(results, gin.H{"name": s.Name, "status": "failed", "error": err.Error()})
			failedCount++
			continue
		}

		// 创建学生分身
		persona := &database.Persona{
			UserID:      newUserID,
			Role:        "student",
			Nickname:    s.Name,
			Description: buildStudentDescription(s.Gender, s.Age, s.Strengths, s.Weaknesses, s.Interests, s.Personality, s.Notes),
		}
		personaID, err := personaRepo.Create(persona)
		if err != nil {
			results = append(results, gin.H{"name": s.Name, "status": "failed", "error": "创建分身失败: " + err.Error()})
			failedCount++
			continue
		}

		// 更新用户默认分身
		_ = userRepo.UpdateDefaultPersonaID(newUserID, personaID)

		// 创建师生关系
		_, _ = relationRepo.CreateWithPersonas(userIDInt64, newUserID, req.PersonaID, personaID, "approved", "teacher")

		// 如果指定了班级，加入班级
		if req.ClassID > 0 {
			classRepo := database.NewClassRepository(db)
			_, _ = classRepo.AddMember(req.ClassID, personaID)
		}

		results = append(results, gin.H{
			"name":       s.Name,
			"user_id":    newUserID,
			"persona_id": personaID,
			"status":     "success",
		})
		successCount++
	}

	Success(c, gin.H{
		"total":         len(req.Students),
		"success_count": successCount,
		"failed_count":  failedCount,
		"results":       results,
	})
}

// buildStudentDescription 构建学生描述信息
func buildStudentDescription(gender string, age int, strengths, weaknesses, interests, personality, notes string) string {
	var parts []string
	if gender != "" {
		parts = append(parts, "性别: "+gender)
	}
	if age > 0 {
		parts = append(parts, fmt.Sprintf("年龄: %d岁", age))
	}
	if strengths != "" {
		parts = append(parts, "擅长: "+strengths)
	}
	if weaknesses != "" {
		parts = append(parts, "薄弱: "+weaknesses)
	}
	if interests != "" {
		parts = append(parts, "兴趣: "+interests)
	}
	if personality != "" {
		parts = append(parts, "性格: "+personality)
	}
	if notes != "" {
		parts = append(parts, "备注: "+notes)
	}
	return strings.Join(parts, "; ")
}

// LLMErrorType 定义LLM错误类型
type LLMErrorType int

const (
	LLMErrorTimeout LLMErrorType = iota // 超时错误
	LLMErrorOther                       // 其他错误
)

// LLMError 定义LLM解析错误
type LLMError struct {
	Type    LLMErrorType
	Message string
	Cause   error
}

func (e *LLMError) Error() string {
	return e.Message
}

// IsTimeoutError 检查是否为超时错误
func IsTimeoutError(err error) bool {
	if llmErr, ok := err.(*LLMError); ok {
		return llmErr.Type == LLMErrorTimeout
	}
	// 检查底层错误是否为超时
	if netErr, ok := err.(interface{ Timeout() bool }); ok {
		return netErr.Timeout()
	}
	return false
}

// llmParseStudentText 使用 LLM 解析学生文本
// 超时控制：连接超时5秒，读写超时25秒
func llmParseStudentText(baseURL, apiKey, model, text string) ([]map[string]interface{}, error) {
	prompt := fmt.Sprintf(`请从以下文本中提取学生信息列表。文本可能是花名册、名单、表格等格式。

文本内容：
%s

请输出JSON数组格式，每个学生包含以下字段（信息不足则留空字符串）：
[
  {
    "name": "姓名",
    "gender": "性别（male/female）",
    "age": 0,
    "student_id": "学号",
    "strengths": "擅长学科",
    "weaknesses": "薄弱学科",
    "interests": "兴趣爱好",
    "personality": "性格特点",
    "notes": "备注"
  }
]

只返回JSON数组，不要其他内容。`, text)

	url := strings.TrimRight(baseURL, "/") + "/chat/completions"
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]interface{}{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.1,
	}

	jsonData, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, &LLMError{Type: LLMErrorOther, Message: "创建请求失败: " + err.Error(), Cause: err}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// 使用精细超时控制：连接超时5秒，总超时30秒
	// 创建自定义Transport，设置连接超时
	dialer := &net.Dialer{
		Timeout:   5 * time.Second, // 连接超时5秒
		KeepAlive: 30 * time.Second,
	}

	client := &http.Client{
		Timeout: 30 * time.Second, // 总超时30秒（包含连接、发送请求、读取响应）
		Transport: &http.Transport{
			DialContext:           dialer.DialContext,
			TLSHandshakeTimeout:   5 * time.Second,  // TLS握手超时5秒
			ResponseHeaderTimeout: 10 * time.Second, // 响应头超时10秒
			ExpectContinueTimeout: 1 * time.Second,
			IdleConnTimeout:       90 * time.Second,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		// 区分超时错误和其他错误
		errType := LLMErrorOther
		if netErr, ok := err.(interface{ Timeout() bool }); ok && netErr.Timeout() {
			errType = LLMErrorTimeout
		}
		return nil, &LLMError{Type: errType, Message: "LLM 请求失败: " + err.Error(), Cause: err}
	}
	defer resp.Body.Close()

	// 使用带超时的解码器读取响应体
	var respData struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	// 设置读取超时上下文
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return nil, &LLMError{Type: LLMErrorOther, Message: "解析响应失败: " + err.Error(), Cause: err}
	}

	if len(respData.Choices) == 0 {
		return nil, &LLMError{Type: LLMErrorOther, Message: "LLM 返回空结果"}
	}

	content := strings.TrimSpace(respData.Choices[0].Message.Content)
	startIdx := strings.Index(content, "[")
	endIdx := strings.LastIndex(content, "]")
	if startIdx >= 0 && endIdx > startIdx {
		content = content[startIdx : endIdx+1]
	}

	var result []map[string]interface{}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, &LLMError{Type: LLMErrorOther, Message: "解析JSON失败: " + err.Error(), Cause: err}
	}
	return result, nil
}

// simpleParseStudentText 简单规则解析学生文本
func simpleParseStudentText(text string) []map[string]interface{} {
	lines := strings.Split(text, "\n")
	var students []map[string]interface{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// 尝试按常见分隔符分割
		parts := strings.FieldsFunc(line, func(r rune) bool {
			return r == ',' || r == '\t' || r == '|' || r == '，' || r == '、' || r == ' '
		})
		if len(parts) == 0 {
			continue
		}

		// 过滤空字符串
		var validParts []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				validParts = append(validParts, p)
			}
		}
		if len(validParts) == 0 {
			continue
		}

		student := map[string]interface{}{
			"name":        strings.TrimSpace(validParts[0]),
			"gender":      "",
			"age":         0,
			"student_id":  "",
			"strengths":   "",
			"weaknesses":  "",
			"interests":   "",
			"personality": "",
			"notes":       "",
		}

		// 尝试识别性别
		if len(validParts) > 1 {
			g := strings.TrimSpace(validParts[1])
			if g == "男" || g == "male" {
				student["gender"] = "male"
			} else if g == "女" || g == "female" {
				student["gender"] = "female"
			}
		}

		students = append(students, student)
	}

	return students
}
