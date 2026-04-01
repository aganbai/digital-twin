package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"digital-twin/src/backend/database"
	"digital-twin/src/harness/core"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ======================== V2.0 迭代6 聊天记录导入 ========================

// HandleImportChat 聊天记录 JSON 导入知识库
// POST /api/documents/import-chat
func (h *Handler) HandleImportChat(c *gin.Context) {
	userID, _ := c.Get("user_id")
	teacherID, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	// 获取文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		Error(c, http.StatusBadRequest, 40004, "请上传聊天记录 JSON 文件")
		return
	}
	defer file.Close()

	// 校验文件大小（最大 5MB）
	if header.Size > 5*1024*1024 {
		Error(c, http.StatusBadRequest, 40036, "文件大小超出限制（最大 5MB）")
		return
	}

	// 校验文件类型
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".json" {
		Error(c, http.StatusBadRequest, 40037, "仅支持 JSON 格式文件")
		return
	}

	// 读取文件内容
	fileContent, err := io.ReadAll(file)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "读取文件失败: "+err.Error())
		return
	}

	// 解析聊天记录
	conversations, err := parseChatJSON(fileContent)
	if err != nil {
		Error(c, http.StatusBadRequest, 40037, "无效的聊天记录 JSON 格式: "+err.Error())
		return
	}

	if len(conversations) == 0 {
		Error(c, http.StatusBadRequest, 40038, "聊天记录为空（解析后无有效对话）")
		return
	}

	// 获取表单参数
	title := c.PostForm("title")
	if title == "" {
		title = fmt.Sprintf("聊天记录导入_%s", time.Now().Format("20060102_150405"))
	}
	tags := c.PostForm("tags")
	if tags == "" {
		tags = "[]"
	}
	scope := c.PostForm("scope")
	if scope == "" {
		scope = "global"
	}
	personaIDStr := c.PostForm("persona_id")
	personaID, _ := strconv.ParseInt(personaIDStr, 10, 64)

	// 拼接为结构化 Q&A 文本
	content := buildQAContent(conversations)

	// 通过 knowledge 插件入库
	plugin, err := h.manager.GetPlugin("knowledge-management")
	if err != nil {
		// 如果 knowledge 插件不可用，直接写入 documents 表
		db := h.manager.GetDB()
		if db == nil {
			Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
			return
		}

		docRepo := database.NewDocumentRepository(db)
		doc := &database.Document{
			TeacherID:       teacherID,
			Title:           title,
			Content:         content,
			DocType:         "chat",
			Tags:            tags,
			Status:          "active",
			Scope:           scope,
			PersonaID:       personaID,
			SourceSessionID: uuid.New().String(),
		}
		docID, err := docRepo.Create(doc)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "保存文档失败: "+err.Error())
			return
		}

		Success(c, gin.H{
			"document_id":        docID,
			"title":              title,
			"doc_type":           "chat",
			"conversation_count": len(conversations),
			"chunks_count":       0,
			"status":             "active",
		})
		return
	}

	// 使用 knowledge 插件入库
	scopeIDs := c.PostForm("scope_ids")
	input := &core.PluginInput{
		RequestID: uuid.New().String(),
		Data: map[string]interface{}{
			"action":            "add",
			"teacher_id":        teacherID,
			"title":             title,
			"content":           content,
			"doc_type":          "chat",
			"tags":              tags,
			"scope":             scope,
			"scope_ids":         scopeIDs,
			"persona_id":        personaID,
			"source_session_id": uuid.New().String(),
		},
		Context: c.Request.Context(),
	}

	output, err := plugin.Execute(c.Request.Context(), input)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "导入聊天记录失败: "+err.Error())
		return
	}

	if !output.Success {
		Error(c, http.StatusInternalServerError, 50001, "导入聊天记录失败: "+output.Error)
		return
	}

	// 从 output 中获取 document_id
	docID := int64(0)
	if v, ok := output.Data["document_id"]; ok {
		switch val := v.(type) {
		case int64:
			docID = val
		case float64:
			docID = int64(val)
		}
	}
	chunksCount := 0
	if v, ok := output.Data["chunks_count"]; ok {
		switch val := v.(type) {
		case int:
			chunksCount = val
		case float64:
			chunksCount = int(val)
		}
	}

	Success(c, gin.H{
		"document_id":        docID,
		"title":              title,
		"doc_type":           "chat",
		"conversation_count": len(conversations),
		"chunks_count":       chunksCount,
		"status":             "active",
	})
}

// ======================== 聊天记录解析器 ========================

// chatMessage 统一的聊天消息结构
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// parseChatJSON 解析聊天记录 JSON（支持3种格式）
func parseChatJSON(data []byte) ([]chatMessage, error) {
	// 尝试格式1: OpenAI 风格（messages 数组）
	var format1 struct {
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(data, &format1); err == nil && len(format1.Messages) > 0 {
		var result []chatMessage
		for _, msg := range format1.Messages {
			if msg.Content != "" {
				result = append(result, chatMessage{
					Role:    normalizeRole(msg.Role),
					Content: msg.Content,
				})
			}
		}
		if len(result) > 0 {
			return result, nil
		}
	}

	// 尝试格式2: conversations 数组（带 sender/text）
	var format2 struct {
		Conversations []struct {
			Sender string `json:"sender"`
			Text   string `json:"text"`
		} `json:"conversations"`
	}
	if err := json.Unmarshal(data, &format2); err == nil && len(format2.Conversations) > 0 {
		var result []chatMessage
		for _, conv := range format2.Conversations {
			if conv.Text != "" {
				result = append(result, chatMessage{
					Role:    normalizeSender(conv.Sender),
					Content: conv.Text,
				})
			}
		}
		if len(result) > 0 {
			return result, nil
		}
	}

	// 尝试格式3: 顶层数组
	var format3 []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
		Sender  string `json:"sender"`
		Text    string `json:"text"`
	}
	if err := json.Unmarshal(data, &format3); err == nil && len(format3) > 0 {
		var result []chatMessage
		for _, msg := range format3 {
			content := msg.Content
			if content == "" {
				content = msg.Text
			}
			role := msg.Role
			if role == "" {
				role = msg.Sender
			}
			if content != "" {
				result = append(result, chatMessage{
					Role:    normalizeRole(role),
					Content: content,
				})
			}
		}
		if len(result) > 0 {
			return result, nil
		}
	}

	return nil, fmt.Errorf("无法识别的 JSON 格式，支持 messages/conversations/顶层数组")
}

// normalizeRole 标准化角色名
func normalizeRole(role string) string {
	role = strings.ToLower(strings.TrimSpace(role))
	switch role {
	case "user", "human", "student", "学生":
		return "user"
	case "assistant", "ai", "bot", "teacher", "老师", "系统":
		return "assistant"
	default:
		return role
	}
}

// normalizeSender 标准化发送者名
func normalizeSender(sender string) string {
	sender = strings.ToLower(strings.TrimSpace(sender))
	switch sender {
	case "user", "human", "student", "学生", "我":
		return "user"
	case "assistant", "ai", "bot", "teacher", "老师", "系统":
		return "assistant"
	default:
		return "user" // 默认为用户
	}
}

// buildQAContent 将聊天消息列表拼接为结构化 Q&A 文本
func buildQAContent(messages []chatMessage) string {
	var sb strings.Builder
	var lastRole string

	for _, msg := range messages {
		if msg.Role == "user" {
			if lastRole == "user" {
				// 连续的用户消息，追加到上一个 Q
				sb.WriteString("\n" + msg.Content)
			} else {
				if lastRole != "" {
					sb.WriteString("\n\n")
				}
				sb.WriteString("Q: " + msg.Content)
			}
		} else if msg.Role == "assistant" {
			if lastRole == "assistant" {
				// 连续的 AI 消息，追加到上一个 A
				sb.WriteString("\n" + msg.Content)
			} else {
				sb.WriteString("\nA: " + msg.Content)
			}
		}
		lastRole = msg.Role
	}

	return sb.String()
}
