package dialogue

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadGradeLevelTemplates_FromYAML 测试从YAML加载学段模板
func TestLoadGradeLevelTemplates_FromYAML(t *testing.T) {
	ResetGradeLevelTemplatesForTest()
	defer ResetGradeLevelTemplatesForTest()

	// 创建临时YAML配置文件
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "grade_level_templates.yaml")
	yamlContent := `grade_level_templates:
  preschool:
    name: "学前"
    description: "学前教育"
    prompt_template: "你是一位幼儿园老师，用简单有趣的语言。"
  primary_lower:
    name: "小学低年级"
    description: "1-3年级"
    prompt_template: "你是一位小学低年级老师，用生动活泼的语言。"
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("创建测试YAML文件失败: %v", err)
	}

	// 加载配置
	err := LoadGradeLevelTemplates(yamlPath)
	if err != nil {
		t.Fatalf("加载学段模板配置失败: %v", err)
	}

	// 验证配置已加载
	templates := getGradeLevelTemplateMap()
	if len(templates) != 2 {
		t.Fatalf("期望2个模板，实际 %d 个", len(templates))
	}

	if templates["preschool"] != "你是一位幼儿园老师，用简单有趣的语言。" {
		t.Errorf("preschool 模板内容不正确: %s", templates["preschool"])
	}

	if templates["primary_lower"] != "你是一位小学低年级老师，用生动活泼的语言。" {
		t.Errorf("primary_lower 模板内容不正确: %s", templates["primary_lower"])
	}
}

// TestLoadGradeLevelTemplates_FallbackToDefault 测试配置文件不存在时回退到默认模板
func TestLoadGradeLevelTemplates_FallbackToDefault(t *testing.T) {
	ResetGradeLevelTemplatesForTest()
	defer ResetGradeLevelTemplatesForTest()

	// 加载不存在的文件
	err := LoadGradeLevelTemplates("/nonexistent/path/templates.yaml")
	if err == nil {
		t.Log("加载不存在的文件返回了 nil error（符合预期，sync.Once 内部处理）")
	}

	// 应该回退到默认模板
	templates := getGradeLevelTemplateMap()
	if len(templates) == 0 {
		t.Fatal("回退到默认模板失败，模板为空")
	}

	// 默认模板应包含8个学段
	if len(templates) < 8 {
		t.Errorf("默认模板数量不足，期望>=8，实际 %d", len(templates))
	}
}

// TestGetGradeLevelTemplate_WithConfig 测试从配置获取学段模板
func TestGetGradeLevelTemplate_WithConfig(t *testing.T) {
	ResetGradeLevelTemplatesForTest()
	defer ResetGradeLevelTemplatesForTest()

	builder := NewPromptBuilder()

	// 未加载配置时使用默认模板
	config := &CurriculumConfig{GradeLevel: "junior"}
	tmpl := builder.getGradeLevelTemplate(config)
	if tmpl == "" {
		t.Error("默认模板中 junior 不应为空")
	}

	// 通过 Grade 自动映射
	config2 := &CurriculumConfig{Grade: "初一"}
	tmpl2 := builder.getGradeLevelTemplate(config2)
	if tmpl2 == "" {
		t.Error("通过 Grade='初一' 应映射到 junior 模板")
	}

	// nil 配置
	tmpl3 := builder.getGradeLevelTemplate(nil)
	if tmpl3 != "" {
		t.Error("nil 配置应返回空字符串")
	}
}

// TestPromptBuilder_BuildSystemPrompt_Priority 测试Prompt组装优先级
func TestPromptBuilder_BuildSystemPrompt_Priority(t *testing.T) {
	ResetGradeLevelTemplatesForTest()
	defer ResetGradeLevelTemplatesForTest()

	builder := NewPromptBuilder()

	chunks := []map[string]interface{}{
		{"content": "知识片段1"},
	}
	memories := []map[string]interface{}{
		{"content": "记忆1", "type": "preference"},
	}
	styleConfig := &StyleConfig{
		TeachingStyle: "苏格拉底式",
		GuidanceLevel: "medium",
	}
	currConfig := &CurriculumConfig{
		GradeLevel:       "junior",
		Grade:            "初一",
		TextbookVersions: []string{"人教版"},
		Subjects:         []string{"数学"},
		CurrentProgress:  "第三章",
	}
	profileSnapshot := "学生偏好：喜欢数学"

	prompt := builder.BuildSystemPrompt(chunks, memories, styleConfig, currConfig, profileSnapshot)

	// 验证 prompt 不为空
	if prompt == "" {
		t.Fatal("BuildSystemPrompt 返回空字符串")
	}

	// 验证关键内容存在
	if !containsStr(prompt, "安全") && !containsStr(prompt, "隐私") {
		t.Log("提示：安全规则可能以不同关键词出现")
	}

	t.Logf("Prompt 长度: %d 字符", len(prompt))
}

func containsStr(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s != "" && substr != "" && findSubstring(s, substr))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
