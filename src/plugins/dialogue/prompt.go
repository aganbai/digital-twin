package dialogue

import (
	"fmt"
	"strings"

	"digital-twin/src/backend/database"
)

// PromptBuilder 提示词构建器
type PromptBuilder struct{}

// NewPromptBuilder 创建提示词构建器
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{}
}

// systemPromptTemplate 系统提示词模板
const systemPromptTemplate = `你是一位采用苏格拉底式教学法的AI教师助手。你的教学原则：
1. 不直接给出答案，而是通过提问引导学生思考
2. 根据学生的回答，逐步深入，层层递进
3. 当学生遇到困难时，给予适当的提示而非直接解答
4. 鼓励学生表达自己的理解，即使不完全正确
5. 在学生得出正确结论时给予肯定和总结

【相关知识】
%s

【学生记忆】
%s`

// BuildSystemPrompt 构建 system prompt
// chunks: 知识库检索到的相关知识片段
// memories: 学生的历史记忆
func (b *PromptBuilder) BuildSystemPrompt(chunks []map[string]interface{}, memories []map[string]interface{}) string {
	// 构建知识片段文本
	knowledgeText := b.buildKnowledgeText(chunks)

	// 构建学生记忆文本
	memoryText := b.buildMemoryText(memories)

	return fmt.Sprintf(systemPromptTemplate, knowledgeText, memoryText)
}

// BuildConversationMessages 构建完整的消息列表
// systemPrompt: 系统提示词
// history: 历史对话记录
// userMessage: 当前用户消息
func (b *PromptBuilder) BuildConversationMessages(systemPrompt string, history []*database.Conversation, userMessage string) []ChatMessage {
	messages := make([]ChatMessage, 0, len(history)+2)

	// 1. 添加 system prompt
	messages = append(messages, ChatMessage{
		Role:    "system",
		Content: systemPrompt,
	})

	// 2. 添加历史对话
	for _, conv := range history {
		messages = append(messages, ChatMessage{
			Role:    conv.Role,
			Content: conv.Content,
		})
	}

	// 3. 添加当前用户消息
	messages = append(messages, ChatMessage{
		Role:    "user",
		Content: userMessage,
	})

	return messages
}

// buildKnowledgeText 构建知识片段文本
func (b *PromptBuilder) buildKnowledgeText(chunks []map[string]interface{}) string {
	if len(chunks) == 0 {
		return "暂无相关知识"
	}

	var parts []string
	for i, chunk := range chunks {
		content, _ := chunk["content"].(string)
		if content == "" {
			continue
		}
		title, _ := chunk["title"].(string)
		if title != "" {
			parts = append(parts, fmt.Sprintf("%d. [%s] %s", i+1, title, content))
		} else {
			parts = append(parts, fmt.Sprintf("%d. %s", i+1, content))
		}
	}

	if len(parts) == 0 {
		return "暂无相关知识"
	}

	return strings.Join(parts, "\n")
}

// buildMemoryText 构建学生记忆文本
func (b *PromptBuilder) buildMemoryText(memories []map[string]interface{}) string {
	if len(memories) == 0 {
		return "暂无学生记忆"
	}

	var parts []string
	for i, mem := range memories {
		content, _ := mem["content"].(string)
		if content == "" {
			continue
		}
		memType, _ := mem["memory_type"].(string)
		if memType != "" {
			parts = append(parts, fmt.Sprintf("%d. [%s] %s", i+1, memType, content))
		} else {
			parts = append(parts, fmt.Sprintf("%d. %s", i+1, content))
		}
	}

	if len(parts) == 0 {
		return "暂无学生记忆"
	}

	return strings.Join(parts, "\n")
}

// memoryExtractionPromptTemplate 记忆提取提示词模板
const memoryExtractionPromptTemplate = `请从以下对话中提取关键学习记忆，以 JSON 数组格式返回。每条记忆包含：
- type: 记忆类型（conversation/learning_progress/personality_traits）
- content: 记忆内容（简洁描述，不超过100字）
- importance: 重要性（0.0-1.0）

对话内容：
学生: %s
教师: %s

要求：
1. 只提取有价值的信息，不要重复对话原文
2. 如果对话没有有价值的记忆点，返回空数组 []
3. 每次最多提取 3 条记忆

返回格式示例：
[{"type": "learning_progress", "content": "学生对牛顿第一定律有基本了解", "importance": 0.8}]`

// BuildMemoryExtractionPrompt 构建记忆提取 prompt
func (b *PromptBuilder) BuildMemoryExtractionPrompt(userMessage, aiReply string) []ChatMessage {
	prompt := fmt.Sprintf(memoryExtractionPromptTemplate, userMessage, aiReply)
	return []ChatMessage{
		{Role: "system", Content: "你是一个记忆提取助手，负责从对话中提取关键学习记忆。只返回 JSON 数组，不要返回其他内容。"},
		{Role: "user", Content: prompt},
	}
}
