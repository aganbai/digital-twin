package config

import (
	"os"
	"path/filepath"
	"testing"
)

// 测试用的最小 YAML 配置
const testYAMLContent = `
system:
  name: "test-harness"
  version: "1.0.0"
  environment: "test"
  debug: true
  logging:
    level: "debug"
    format: "json"
    output: "stdout"
  monitoring:
    enabled: false
    metrics_port: 9090
    health_check_interval: 30s

plugins:
  test-plugin:
    enabled: true
    type: "auth"
    priority: 5
    config:
      jwt:
        secret: "test-secret"

pipelines:
  test_pipeline:
    description: "测试管道"
    plugins:
      - "test-plugin"
    timeout: "30s"

feature_flags:
  enable_analytics: false

performance:
  max_concurrent_plugins: 10
  plugin_timeout: 30s
  cache_ttl: 5m
  rate_limiting:
    enabled: true
    requests_per_minute: 100
    burst_size: 20

security:
  cors:
    allowed_origins:
      - "http://localhost:3000"
    allowed_methods: ["GET", "POST"]
    allowed_headers: ["Content-Type"]
  encryption:
    enabled: false
    algorithm: "aes-256-gcm"
  audit:
    enabled: false
    log_level: "info"
    retention_days: 30
`

// 带环境变量占位符的 YAML 配置
const testYAMLWithEnvVars = `
system:
  name: "test-harness"
  version: "1.0.0"
  environment: "${TEST_ENV:-development}"
  debug: true
  logging:
    level: "${LOG_LEVEL:-info}"
    format: "json"
    output: "stdout"
  monitoring:
    enabled: false
    metrics_port: 9090
    health_check_interval: 30s

plugins:
  auth-plugin:
    enabled: true
    type: "auth"
    priority: 5
    config:
      jwt:
        secret: "${TEST_JWT_SECRET:-default-secret}"

pipelines:
  test_pipeline:
    description: "测试管道"
    plugins:
      - "auth-plugin"
    timeout: "30s"
`

func TestLoadConfig_BasicParsing(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(cfgPath, []byte(testYAMLContent), 0644); err != nil {
		t.Fatalf("创建测试配置文件失败: %v", err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 验证系统配置
	if cfg.System.Name != "test-harness" {
		t.Errorf("系统名称不匹配: got %q, want %q", cfg.System.Name, "test-harness")
	}
	if cfg.System.Version != "1.0.0" {
		t.Errorf("版本不匹配: got %q, want %q", cfg.System.Version, "1.0.0")
	}
	if cfg.System.Environment != "test" {
		t.Errorf("环境不匹配: got %q, want %q", cfg.System.Environment, "test")
	}
	if !cfg.System.Debug {
		t.Error("Debug 应为 true")
	}

	// 验证日志配置
	if cfg.System.Logging.Level != "debug" {
		t.Errorf("日志级别不匹配: got %q, want %q", cfg.System.Logging.Level, "debug")
	}

	// 验证插件配置
	plugin, ok := cfg.Plugins["test-plugin"]
	if !ok {
		t.Fatal("未找到 test-plugin 插件配置")
	}
	if plugin.Type != "auth" {
		t.Errorf("插件类型不匹配: got %q, want %q", plugin.Type, "auth")
	}
	if plugin.Priority != 5 {
		t.Errorf("插件优先级不匹配: got %d, want %d", plugin.Priority, 5)
	}
	if !plugin.Enabled {
		t.Error("插件应为启用状态")
	}

	// 验证管道配置
	pipeline, ok := cfg.Pipelines["test_pipeline"]
	if !ok {
		t.Fatal("未找到 test_pipeline 管道配置")
	}
	if len(pipeline.Plugins) != 1 || pipeline.Plugins[0] != "test-plugin" {
		t.Errorf("管道插件列表不匹配: got %v", pipeline.Plugins)
	}
	if pipeline.Timeout != "30s" {
		t.Errorf("管道超时不匹配: got %q, want %q", pipeline.Timeout, "30s")
	}
}

func TestLoadConfig_EnvVarReplacement(t *testing.T) {
	// 设置环境变量
	os.Setenv("TEST_ENV", "production")
	os.Setenv("LOG_LEVEL", "warn")
	os.Setenv("TEST_JWT_SECRET", "my-super-secret")
	defer func() {
		os.Unsetenv("TEST_ENV")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("TEST_JWT_SECRET")
	}()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(cfgPath, []byte(testYAMLWithEnvVars), 0644); err != nil {
		t.Fatalf("创建测试配置文件失败: %v", err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 验证环境变量替换
	if cfg.System.Environment != "production" {
		t.Errorf("环境变量替换失败: got %q, want %q", cfg.System.Environment, "production")
	}
	if cfg.System.Logging.Level != "warn" {
		t.Errorf("环境变量替换失败: got %q, want %q", cfg.System.Logging.Level, "warn")
	}
}

func TestLoadConfig_EnvVarDefaultValues(t *testing.T) {
	// 确保环境变量未设置
	os.Unsetenv("TEST_ENV")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("TEST_JWT_SECRET")

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(cfgPath, []byte(testYAMLWithEnvVars), 0644); err != nil {
		t.Fatalf("创建测试配置文件失败: %v", err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 验证默认值
	if cfg.System.Environment != "development" {
		t.Errorf("默认值不正确: got %q, want %q", cfg.System.Environment, "development")
	}
	if cfg.System.Logging.Level != "info" {
		t.Errorf("默认值不正确: got %q, want %q", cfg.System.Logging.Level, "info")
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("应返回文件不存在错误")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "invalid.yaml")
	if err := os.WriteFile(cfgPath, []byte("{{invalid yaml"), 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	_, err := LoadConfig(cfgPath)
	if err == nil {
		t.Fatal("应返回 YAML 解析错误")
	}
}

func TestLoadConfig_ValidationMissingName(t *testing.T) {
	yamlContent := `
system:
  name: ""
  version: "1.0.0"
plugins: {}
pipelines: {}
`
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(cfgPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	_, err := LoadConfig(cfgPath)
	if err == nil {
		t.Fatal("应返回校验错误：system.name 为空")
	}
}

func TestLoadConfig_ValidationMissingVersion(t *testing.T) {
	yamlContent := `
system:
  name: "test"
  version: ""
plugins: {}
pipelines: {}
`
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(cfgPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	_, err := LoadConfig(cfgPath)
	if err == nil {
		t.Fatal("应返回校验错误：system.version 为空")
	}
}

func TestLoadConfig_ValidationEmptyPluginType(t *testing.T) {
	yamlContent := `
system:
  name: "test"
  version: "1.0.0"
plugins:
  bad-plugin:
    enabled: true
    type: ""
pipelines: {}
`
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(cfgPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	_, err := LoadConfig(cfgPath)
	if err == nil {
		t.Fatal("应返回校验错误：插件 type 为空")
	}
}

func TestLoadConfig_ValidationEmptyPipelinePlugins(t *testing.T) {
	yamlContent := `
system:
  name: "test"
  version: "1.0.0"
plugins: {}
pipelines:
  bad_pipeline:
    description: "空管道"
    plugins: []
`
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(cfgPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	_, err := LoadConfig(cfgPath)
	if err == nil {
		t.Fatal("应返回校验错误：管道 plugins 为空")
	}
}

func TestReplaceEnvVars(t *testing.T) {
	os.Setenv("MY_VAR", "hello")
	defer os.Unsetenv("MY_VAR")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "有环境变量-无默认值",
			input:    "${MY_VAR}",
			expected: "hello",
		},
		{
			name:     "有环境变量-有默认值",
			input:    "${MY_VAR:-world}",
			expected: "hello",
		},
		{
			name:     "无环境变量-有默认值",
			input:    "${NONEXISTENT:-fallback}",
			expected: "fallback",
		},
		{
			name:     "无环境变量-无默认值",
			input:    "${NONEXISTENT}",
			expected: "",
		},
		{
			name:     "混合文本",
			input:    "prefix-${MY_VAR}-suffix",
			expected: "prefix-hello-suffix",
		},
		{
			name:     "多个变量",
			input:    "${MY_VAR} and ${NONEXISTENT:-default}",
			expected: "hello and default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replaceEnvVars(tt.input)
			if result != tt.expected {
				t.Errorf("replaceEnvVars(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLoadConfig_RealHarnessYAML(t *testing.T) {
	// 测试加载真实的 harness.yaml 配置文件
	cfgPath := filepath.Join("..", "..", "..", "configs", "harness.yaml")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		t.Skip("跳过：harness.yaml 文件不存在")
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("加载 harness.yaml 失败: %v", err)
	}

	// 验证基本结构
	if cfg.System.Name != "digital-twin-harness" {
		t.Errorf("系统名称不匹配: got %q", cfg.System.Name)
	}

	// 验证插件数量
	if len(cfg.Plugins) < 4 {
		t.Errorf("插件数量不足: got %d, want >= 4", len(cfg.Plugins))
	}

	// 验证管道数量
	if len(cfg.Pipelines) < 2 {
		t.Errorf("管道数量不足: got %d, want >= 2", len(cfg.Pipelines))
	}
}
