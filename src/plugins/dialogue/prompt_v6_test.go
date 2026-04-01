package dialogue

import (
	"strings"
	"testing"
)

// TestTeachingStyleTemplates_AllStylesExist 验证所有风格模板存在
func TestTeachingStyleTemplates_AllStylesExist(t *testing.T) {
	expectedStyles := []string{"socratic", "explanatory", "encouraging", "strict", "companion"}
	for _, style := range expectedStyles {
		if _, ok := teachingStyleTemplates[style]; !ok {
			t.Errorf("缺少风格模板: %s", style)
		}
	}
}

// TestValidTeachingStyles 验证有效风格列表
func TestValidTeachingStyles(t *testing.T) {
	expectedValid := []string{"socratic", "explanatory", "encouraging", "strict", "companion", "custom"}
	for _, style := range expectedValid {
		if !ValidTeachingStyles[style] {
			t.Errorf("风格 %s 应在有效列表中", style)
		}
	}

	// 无效风格
	if ValidTeachingStyles["invalid_style"] {
		t.Error("invalid_style 不应在有效列表中")
	}
}

// TestGetTeachingStyleTemplate_Default 默认使用 socratic
func TestGetTeachingStyleTemplate_Default(t *testing.T) {
	pb := NewPromptBuilder()

	// nil styleConfig → socratic
	result := pb.getTeachingStyleTemplate(nil)
	if !strings.Contains(result, "苏格拉底") {
		t.Error("nil styleConfig 应使用苏格拉底式模板")
	}

	// 空 teaching_style → socratic
	result = pb.getTeachingStyleTemplate(&StyleConfig{})
	if !strings.Contains(result, "苏格拉底") {
		t.Error("空 teaching_style 应使用苏格拉底式模板")
	}
}

// TestGetTeachingStyleTemplate_EachStyle 测试每种风格
func TestGetTeachingStyleTemplate_EachStyle(t *testing.T) {
	pb := NewPromptBuilder()

	tests := []struct {
		style    string
		contains string
	}{
		{"socratic", "苏格拉底"},
		{"explanatory", "讲解式"},
		{"encouraging", "鼓励式"},
		{"strict", "严格式"},
		{"companion", "陪伴式"},
	}

	for _, tt := range tests {
		result := pb.getTeachingStyleTemplate(&StyleConfig{TeachingStyle: tt.style})
		if !strings.Contains(result, tt.contains) {
			t.Errorf("风格 %s 的模板应包含 %q", tt.style, tt.contains)
		}
	}
}

// TestGetTeachingStyleTemplate_CustomWithPrompt custom 风格使用 style_prompt
func TestGetTeachingStyleTemplate_CustomWithPrompt(t *testing.T) {
	pb := NewPromptBuilder()

	customPrompt := "你是一个专注于编程教学的AI，用代码示例来教学"
	result := pb.getTeachingStyleTemplate(&StyleConfig{
		TeachingStyle: "custom",
		StylePrompt:   customPrompt,
	})
	if result != customPrompt {
		t.Errorf("custom 风格应使用 style_prompt, 实际=%s", result)
	}
}

// TestGetTeachingStyleTemplate_CustomWithoutPrompt custom 风格无 style_prompt 回退 socratic
func TestGetTeachingStyleTemplate_CustomWithoutPrompt(t *testing.T) {
	pb := NewPromptBuilder()

	result := pb.getTeachingStyleTemplate(&StyleConfig{
		TeachingStyle: "custom",
		StylePrompt:   "",
	})
	if !strings.Contains(result, "苏格拉底") {
		t.Error("custom 风格无 style_prompt 应回退到苏格拉底式")
	}
}

// TestGetTeachingStyleTemplate_UnknownFallback 未知风格回退 socratic
func TestGetTeachingStyleTemplate_UnknownFallback(t *testing.T) {
	pb := NewPromptBuilder()

	result := pb.getTeachingStyleTemplate(&StyleConfig{TeachingStyle: "unknown_style"})
	if !strings.Contains(result, "苏格拉底") {
		t.Error("未知风格应回退到苏格拉底式")
	}
}

// TestBuildSystemPrompt_WithTeachingStyle 测试完整 system prompt 构建
func TestBuildSystemPrompt_WithTeachingStyle(t *testing.T) {
	pb := NewPromptBuilder()

	chunks := []map[string]interface{}{
		{"title": "数学", "content": "二次方程的解法"},
	}
	memories := []map[string]interface{}{
		{"memory_type": "preference", "content": "喜欢数学"},
	}

	// encouraging 风格
	prompt := pb.BuildSystemPrompt(chunks, memories, &StyleConfig{
		TeachingStyle: "encouraging",
		GuidanceLevel: "low",
	})

	if !strings.Contains(prompt, "鼓励式") {
		t.Error("prompt 应包含鼓励式教学模板")
	}
	if !strings.Contains(prompt, "二次方程的解法") {
		t.Error("prompt 应包含知识内容")
	}
	if !strings.Contains(prompt, "喜欢数学") {
		t.Error("prompt 应包含学生记忆")
	}
}

// TestBuildStyleText_GuidanceLevel 测试引导程度说明
func TestBuildStyleText_GuidanceLevel(t *testing.T) {
	pb := NewPromptBuilder()

	// low → 少提问
	lowText := pb.buildStyleText(&StyleConfig{GuidanceLevel: "low"})
	if !strings.Contains(lowText, "少提问") {
		t.Error("low guidance_level 应包含'少提问'")
	}

	// high → 加强引导
	highText := pb.buildStyleText(&StyleConfig{GuidanceLevel: "high"})
	if !strings.Contains(highText, "加强引导") {
		t.Error("high guidance_level 应包含'加强引导'")
	}

	// medium → 无额外说明
	mediumText := pb.buildStyleText(&StyleConfig{GuidanceLevel: "medium"})
	if mediumText != "" {
		t.Error("medium guidance_level 不应有额外说明")
	}
}
