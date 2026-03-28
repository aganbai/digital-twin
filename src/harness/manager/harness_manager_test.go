package manager

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"digital-twin/src/harness/core"
)

// getProjectRoot 获取项目根目录（向上查找 go.mod）
func getProjectRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// TestNewHarnessManager 测试创建新的Harness管理器
func TestNewHarnessManager(t *testing.T) {
	manager, err := NewHarnessManager("configs/harness.yaml")
	if err != nil {
		t.Fatalf("Failed to create harness manager: %v", err)
	}

	if manager == nil {
		t.Error("Expected non-nil manager")
	}
}

// TestHarnessManagerLifecycle 测试管理器的生命周期
func TestHarnessManagerLifecycle(t *testing.T) {
	root := getProjectRoot()
	if root == "" {
		t.Skip("无法找到项目根目录")
	}

	// 切换到项目根目录以便加载配置
	origDir, _ := os.Getwd()
	os.Chdir(root)
	defer os.Chdir(origDir)

	// 设置 JWT_SECRET 环境变量（配置文件中引用了）
	os.Setenv("JWT_SECRET", "test-secret-key-for-unit-testing-32chars")
	defer os.Unsetenv("JWT_SECRET")

	// 使用临时数据库
	tmpDB := filepath.Join(t.TempDir(), "test.db")
	os.Setenv("DB_PATH", tmpDB)
	defer os.Unsetenv("DB_PATH")

	manager, err := NewHarnessManager("configs/harness.yaml")
	if err != nil {
		t.Fatalf("Failed to create harness manager: %v", err)
	}

	// 测试启动
	err = manager.Start()
	if err != nil {
		t.Errorf("Failed to start manager: %v", err)
		return
	}

	// 测试健康检查
	health := manager.HealthCheck()
	if health["status"] != ManagerStatusRunning {
		t.Errorf("Expected status 'running', got '%v'", health["status"])
	}

	// 验证插件已加载
	plugins := manager.ListPlugins()
	if len(plugins) == 0 {
		t.Error("Expected plugins to be loaded")
	}

	// 验证管道已构建
	pipelines := manager.ListPipelines()
	if len(pipelines) == 0 {
		t.Error("Expected pipelines to be built")
	}

	// 测试停止
	err = manager.Stop()
	if err != nil {
		t.Errorf("Failed to stop manager: %v", err)
	}
}

// TestPipelineCreation 测试管道创建
func TestPipelineCreation(t *testing.T) {
	root := getProjectRoot()
	if root == "" {
		t.Skip("无法找到项目根目录")
	}

	origDir, _ := os.Getwd()
	os.Chdir(root)
	defer os.Chdir(origDir)

	os.Setenv("JWT_SECRET", "test-secret-key-for-unit-testing-32chars")
	defer os.Unsetenv("JWT_SECRET")

	tmpDB := filepath.Join(t.TempDir(), "test.db")
	os.Setenv("DB_PATH", tmpDB)
	defer os.Unsetenv("DB_PATH")

	manager, err := NewHarnessManager("configs/harness.yaml")
	if err != nil {
		t.Fatalf("Failed to create harness manager: %v", err)
	}

	// 启动管理器
	err = manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	defer manager.Stop()

	// 创建测试管道配置
	config := &core.PipelineConfig{
		Name:        "test_pipeline",
		Description: "Test pipeline",
		Plugins:     []string{},
		Timeout:     "30s",
	}

	// 测试管道创建
	err = manager.CreatePipeline(config)
	if err != nil {
		t.Errorf("Failed to create pipeline: %v", err)
	}

	// 验证管道存在
	pipelines := manager.ListPipelines()
	found := false
	for _, name := range pipelines {
		if name == "test_pipeline" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Pipeline not found in list")
	}
}

// TestPluginRegistration 测试插件注册
func TestPluginRegistration(t *testing.T) {
	root := getProjectRoot()
	if root == "" {
		t.Skip("无法找到项目根目录")
	}

	origDir, _ := os.Getwd()
	os.Chdir(root)
	defer os.Chdir(origDir)

	os.Setenv("JWT_SECRET", "test-secret-key-for-unit-testing-32chars")
	defer os.Unsetenv("JWT_SECRET")

	tmpDB := filepath.Join(t.TempDir(), "test.db")
	os.Setenv("DB_PATH", tmpDB)
	defer os.Unsetenv("DB_PATH")

	manager, err := NewHarnessManager("configs/harness.yaml")
	if err != nil {
		t.Fatalf("Failed to create harness manager: %v", err)
	}

	// 启动管理器
	err = manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	defer manager.Stop()

	// 创建测试插件
	testPlugin := &TestPlugin{
		name:    "test_plugin",
		version: "1.0.0",
		ptype:   core.PluginTypeKnowledge,
	}

	// 注册插件
	err = manager.RegisterPlugin(testPlugin)
	if err != nil {
		t.Errorf("Failed to register plugin: %v", err)
	}

	// 验证插件存在
	plugins := manager.ListPlugins()
	found := false
	for _, plugin := range plugins {
		if plugin.Name == "test_plugin" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Plugin not found in list")
	}
}

// TestPlugin 测试插件实现
type TestPlugin struct {
	name    string
	version string
	ptype   core.PluginType
}

func (p *TestPlugin) Name() string {
	return p.name
}

func (p *TestPlugin) Version() string {
	return p.version
}

func (p *TestPlugin) Type() core.PluginType {
	return p.ptype
}

func (p *TestPlugin) Init(config map[string]interface{}) error {
	return nil
}

func (p *TestPlugin) Execute(ctx context.Context, input *core.PluginInput) (*core.PluginOutput, error) {
	return &core.PluginOutput{
		Success: true,
		Data:    map[string]interface{}{"message": "test response"},
	}, nil
}

func (p *TestPlugin) HealthCheck() error {
	return nil
}

func (p *TestPlugin) Destroy() error {
	return nil
}

func (p *TestPlugin) GetConfig() map[string]interface{} {
	return make(map[string]interface{})
}

func (p *TestPlugin) UpdateConfig(config map[string]interface{}) error {
	return nil
}