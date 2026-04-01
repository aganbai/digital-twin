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

// StyleConfig 个性化风格配置（本地定义，避免循环依赖）
type StyleConfig struct {
	Temperature      float64 `json:"temperature"`
	GuidanceLevel    string  `json:"guidance_level"`
	TeachingStyle    string  `json:"teaching_style"` // V2.0 迭代6: socratic / explanatory / encouraging / strict / companion / custom
	StylePrompt      string  `json:"style_prompt"`
	MaxTurnsPerTopic int     `json:"max_turns_per_topic"`
}

// ValidTeachingStyles 有效的教学风格列表
var ValidTeachingStyles = map[string]bool{
	"socratic":    true,
	"explanatory": true,
	"encouraging": true,
	"strict":      true,
	"companion":   true,
	"custom":      true,
}

// teachingStyleTemplates 教学风格模板映射
var teachingStyleTemplates = map[string]string{
	"socratic": `你是一位采用苏格拉底式教学法的AI教师助手。你的教学原则：
1. 不直接给出答案，而是通过提问引导学生思考
2. 根据学生的回答，逐步深入，层层递进
3. 当学生遇到困难时，给予适当的提示而非直接解答
4. 鼓励学生表达自己的理解，即使不完全正确
5. 在学生得出正确结论时给予肯定和总结`,

	"explanatory": `你是一位采用讲解式教学法的AI教师助手。你的教学原则：
1. 详细讲解知识点，配合举例说明
2. 从基础概念入手，逐步引入复杂内容
3. 用类比和比喻帮助学生理解抽象概念
4. 每个知识点都给出具体的例子
5. 在讲解后简要总结要点，确保学生理解`,

	"encouraging": `你是一位采用鼓励式教学法的AI教师助手。你的教学原则：
1. 多用肯定语言，及时表扬学生的进步
2. 当学生犯错时，先肯定努力，再温和地纠正
3. 循序渐进地引导，不给学生压力
4. 用“你做得很好”、“进步很大”等语言激励学生
5. 设置小目标，让学生体验成就感`,

	"strict": `你是一位采用严格式教学法的AI教师助手。你的教学原则：
1. 严格要求，注重准确性和规范性
2. 对错误直接指出，不含糊其辞
3. 要求学生用专业术语回答问题
4. 强调基础知识的重要性，不跳步
5. 定期检验学习成果，确保掌握牢固`,

	"companion": `你是一位采用陪伴式学习方法的AI学习伙伴。你的交流原则：
1. 像朋友一样陪伴学习，营造轻松氛围
2. 用口语化的表达，避免过于正式
3. 分享学习心得和技巧，而不是居高临下地教导
4. 适当使用表情和轻松的语气
5. 一起探讨问题，而不是单方面输出知识`,
}

// systemPromptBaseTemplate 基础系统提示词模板（知识和记忆部分）
const systemPromptBaseTemplate = `

【相关知识】
%s

【学生记忆】
%s`

// BuildSystemPrompt 构建 system prompt
// chunks: 知识库检索到的相关知识片段
// memories: 学生的历史记忆
// styleConfig: 个性化教学风格配置（可为 nil）
func (b *PromptBuilder) BuildSystemPrompt(chunks []map[string]interface{}, memories []map[string]interface{}, styleConfig *StyleConfig) string {
	// 构建知识片段文本
	knowledgeText := b.buildKnowledgeText(chunks)

	// 构建学生记忆文本
	memoryText := b.buildMemoryText(memories)

	// 获取教学风格模板
	teachingStyleText := b.getTeachingStyleTemplate(styleConfig)

	// 构建个性化教学要求段落
	styleText := b.buildStyleText(styleConfig)

	// 知识和记忆部分
	basePart := fmt.Sprintf(systemPromptBaseTemplate, knowledgeText, memoryText)

	// 组合：教学风格模板 + 个性化要求 + 知识和记忆
	if styleText != "" {
		return teachingStyleText + "\n\n" + styleText + basePart
	}
	return teachingStyleText + basePart
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

// getTeachingStyleTemplate 获取教学风格模板
func (b *PromptBuilder) getTeachingStyleTemplate(styleConfig *StyleConfig) string {
	style := "socratic" // 默认苏格拉底式

	if styleConfig != nil && styleConfig.TeachingStyle != "" {
		style = styleConfig.TeachingStyle
	}

	// custom 风格：如果 style_prompt 不为空，使用 style_prompt；否则回退到 socratic
	if style == "custom" {
		if styleConfig != nil && styleConfig.StylePrompt != "" {
			return styleConfig.StylePrompt
		}
		style = "socratic" // 回退
	}

	if tmpl, ok := teachingStyleTemplates[style]; ok {
		return tmpl
	}

	// 未知风格，回退到 socratic
	return teachingStyleTemplates["socratic"]
}

// buildStyleText 构建个性化教学要求文本
func (b *PromptBuilder) buildStyleText(styleConfig *StyleConfig) string {
	if styleConfig == nil {
		return ""
	}

	var parts []string
	parts = append(parts, "【个性化教学要求】")

	// 追加 style_prompt 内容
	if styleConfig.StylePrompt != "" {
		parts = append(parts, styleConfig.StylePrompt)
	}

	// 根据 guidance_level 追加引导程度说明
	switch styleConfig.GuidanceLevel {
	case "low":
		parts = append(parts, "请尽量少提问，多给出直接的解释和答案")
	case "high":
		parts = append(parts, "请加强引导，每次回复包含至少一个引导性问题")
	case "medium":
		// 保持默认教学风格，不追加
	}

	if len(parts) <= 1 {
		// 只有标题没有内容，不输出
		return ""
	}

	return strings.Join(parts, "\n")
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

// BuildAssignmentReviewPrompt 构建作业点评 prompt
// teacherNickname: 教师昵称
// title: 作业标题
// content: 作业内容
// knowledgeChunks: 知识库检索到的相关知识片段（可为空）
func (b *PromptBuilder) BuildAssignmentReviewPrompt(teacherNickname, title, content, knowledgeChunks string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("你是%s的数字分身，请根据以下知识库内容，对学生的作业进行点评。\n\n", teacherNickname))
	sb.WriteString("要求：\n")
	sb.WriteString("1. 指出作业中的优点\n")
	sb.WriteString("2. 指出需要改进的地方\n")
	sb.WriteString("3. 给出具体的改进建议\n")
	sb.WriteString("4. 给出评分（0-100），格式为\"评分: XX/100\"\n")

	if knowledgeChunks != "" {
		sb.WriteString("\n【知识库参考】\n")
		sb.WriteString(knowledgeChunks)
		sb.WriteString("\n")
	}

	sb.WriteString("\n【学生作业】\n")
	sb.WriteString(fmt.Sprintf("标题: %s\n", title))
	sb.WriteString(fmt.Sprintf("内容: %s", content))

	return sb.String()
}
