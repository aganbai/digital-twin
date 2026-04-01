package manager

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	"digital-twin/src/backend/database"
	"digital-twin/src/harness/config"
	"digital-twin/src/harness/core"
)

// HarnessManager Harness模式管理器
type HarnessManager struct {
	mu sync.RWMutex

	// 插件注册表
	plugins map[string]core.Plugin
	
	// 管道映射
	pipelines map[string]*Pipeline
	
	// 配置路径
	configPath string
	
	// 配置对象
	harnessConfig *config.HarnessConfig
	
	// 数据库
	db *database.Database
	
	// 事件总线
	eventBus core.EventBus
	
	// 状态
	status ManagerStatus

	// 启动时间
	startTime time.Time
	
	// 指标收集器
	metrics *MetricsCollector
}

// ManagerStatus 管理器状态
type ManagerStatus string

const (
	ManagerStatusStopped  ManagerStatus = "stopped"
	ManagerStatusStarting ManagerStatus = "starting"
	ManagerStatusRunning  ManagerStatus = "running"
	ManagerStatusStopping ManagerStatus = "stopping"
)

// NewHarnessManager 创建新的Harness管理器
func NewHarnessManager(configPath string) (*HarnessManager, error) {
	manager := &HarnessManager{
		configPath:    configPath,
		plugins:       make(map[string]core.Plugin),
		pipelines:     make(map[string]*Pipeline),
		status:        ManagerStatusStopped,
		metrics:       NewMetricsCollector(),
	}

	return manager, nil
}

// Start 启动Harness管理器
func (hm *HarnessManager) Start() error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hm.status != ManagerStatusStopped {
		return fmt.Errorf("manager is already %s", hm.status)
	}

	hm.status = ManagerStatusStarting

	// 加载配置
	if err := hm.loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// 初始化事件总线
	if err := hm.initEventBus(); err != nil {
		return fmt.Errorf("failed to init event bus: %v", err)
	}

	// 加载并初始化插件
	if err := hm.loadPlugins(); err != nil {
		return fmt.Errorf("failed to load plugins: %v", err)
	}

	// 构建管道
	if err := hm.buildPipelines(); err != nil {
		return fmt.Errorf("failed to build pipelines: %v", err)
	}

	// 启动健康检查
	go hm.startHealthCheck()

	hm.status = ManagerStatusRunning
	hm.startTime = time.Now()
	
	hm.metrics.RecordManagerStart()
	
	return nil
}

// Stop 停止Harness管理器
func (hm *HarnessManager) Stop() error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hm.status != ManagerStatusRunning {
		return fmt.Errorf("manager is not running, current status: %s", hm.status)
	}

	hm.status = ManagerStatusStopping

	// 停止所有插件
	for name, plugin := range hm.plugins {
		if err := plugin.Destroy(); err != nil {
			fmt.Printf("Warning: failed to destroy plugin %s: %v\n", name, err)
		}
	}

	// 清理管道
	hm.pipelines = make(map[string]*Pipeline)

	// 清理插件
	hm.plugins = make(map[string]core.Plugin)

	hm.status = ManagerStatusStopped
	
	hm.metrics.RecordManagerStop()

	return nil
}

// RegisterPlugin 注册插件
func (hm *HarnessManager) RegisterPlugin(plugin core.Plugin) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	name := plugin.Name()
	
	if _, exists := hm.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	hm.plugins[name] = plugin
	
	hm.metrics.RecordPluginRegistration(name)

	return nil
}

// UnregisterPlugin 注销插件
func (hm *HarnessManager) UnregisterPlugin(name string) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	plugin, exists := hm.plugins[name]
	if !exists {
		return core.ErrPluginNotFound
	}

	// 从所有管道中移除该插件
	for _, pipeline := range hm.pipelines {
		pipeline.RemovePlugin(name)
	}

	// 销毁插件
	if err := plugin.Destroy(); err != nil {
		return fmt.Errorf("failed to destroy plugin %s: %v", name, err)
	}

	delete(hm.plugins, name)
	
	hm.metrics.RecordPluginUnregistration(name)

	return nil
}

// GetPlugin 获取插件
func (hm *HarnessManager) GetPlugin(name string) (core.Plugin, error) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	plugin, exists := hm.plugins[name]
	if !exists {
		return nil, core.ErrPluginNotFound
	}

	return plugin, nil
}

// ListPlugins 列出所有插件
func (hm *HarnessManager) ListPlugins() []core.PluginInfo {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	var plugins []core.PluginInfo
	for _, plugin := range hm.plugins {
		plugins = append(plugins, core.PluginInfo{
			Name:    plugin.Name(),
			Version: plugin.Version(),
			Type:    plugin.Type(),
			Enabled: true, // TODO: 从配置中获取
		})
	}

	return plugins
}

// ExecutePipeline 执行管道
func (hm *HarnessManager) ExecutePipeline(pipelineName string, input *core.PluginInput) (*core.PluginOutput, error) {
	hm.mu.RLock()
	pipeline, exists := hm.pipelines[pipelineName]
	hm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("pipeline %s not found", pipelineName)
	}

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(input.Context, pipeline.Timeout())
	defer cancel()

	input.Context = ctx

	startTime := time.Now()
	
	// 执行管道
	output, err := pipeline.Execute(ctx, input)
	
	duration := time.Since(startTime)
	
	// 记录指标
	hm.metrics.RecordPipelineExecution(pipelineName, duration, err == nil)

	if err != nil {
		return nil, fmt.Errorf("pipeline execution failed: %v", err)
	}

	output.Duration = duration

	return output, nil
}

// CreatePipeline 创建新管道
func (hm *HarnessManager) CreatePipeline(config *core.PipelineConfig) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if _, exists := hm.pipelines[config.Name]; exists {
		return fmt.Errorf("pipeline %s already exists", config.Name)
	}

	pipeline := NewPipeline(config.Name)
	
	// 添加插件到管道
	for _, pluginName := range config.Plugins {
		plugin, exists := hm.plugins[pluginName]
		if !exists {
			return fmt.Errorf("plugin %s not found for pipeline %s", pluginName, config.Name)
		}
		
		if err := pipeline.AddPlugin(plugin); err != nil {
			return fmt.Errorf("failed to add plugin %s to pipeline: %v", pluginName, err)
		}
	}

	// 设置超时
	if config.Timeout != "" {
		timeout, err := time.ParseDuration(config.Timeout)
		if err != nil {
			return fmt.Errorf("invalid timeout format: %v", err)
		}
		pipeline.SetTimeout(timeout)
	}

	hm.pipelines[config.Name] = pipeline
	
	hm.metrics.RecordPipelineCreation(config.Name)

	return nil
}

// RemovePipeline 移除管道
func (hm *HarnessManager) RemovePipeline(name string) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if _, exists := hm.pipelines[name]; !exists {
		return fmt.Errorf("pipeline %s not found", name)
	}

	delete(hm.pipelines, name)
	
	hm.metrics.RecordPipelineRemoval(name)

	return nil
}

// GetPipeline 获取管道
func (hm *HarnessManager) GetPipeline(name string) (*Pipeline, error) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	pipeline, exists := hm.pipelines[name]
	if !exists {
		return nil, fmt.Errorf("pipeline %s not found", name)
	}

	return pipeline, nil
}

// ListPipelines 列出所有管道
func (hm *HarnessManager) ListPipelines() []string {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	var names []string
	for name := range hm.pipelines {
		names = append(names, name)
	}

	return names
}

// HealthCheck 健康检查
func (hm *HarnessManager) HealthCheck() map[string]interface{} {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	// 检查插件健康状态（去重：只统计配置文件中定义的插件名，跳过类型别名）
	pluginDetails := make(map[string]string)
	healthyCount := 0
	configPluginNames := make(map[string]bool)
	if hm.harnessConfig != nil {
		for name := range hm.harnessConfig.Plugins {
			configPluginNames[name] = true
		}
	}
	for name, plugin := range hm.plugins {
		// 如果有配置信息，只统计配置中定义的插件名
		if len(configPluginNames) > 0 && !configPluginNames[name] {
			continue
		}
		if err := plugin.HealthCheck(); err != nil {
			pluginDetails[name] = err.Error()
		} else {
			pluginDetails[name] = "healthy"
			healthyCount++
		}
	}

	// 管道名称列表
	pipelineNames := make([]string, 0, len(hm.pipelines))
	for name := range hm.pipelines {
		pipelineNames = append(pipelineNames, name)
	}

	// 数据库连接状态
	dbStatus := "disconnected"
	if hm.db != nil && hm.db.DB != nil {
		if err := hm.db.DB.Ping(); err == nil {
			dbStatus = "connected"
		}
	}

	// 计算运行时长
	var uptimeSeconds int64
	if !hm.startTime.IsZero() {
		uptimeSeconds = int64(time.Since(hm.startTime).Seconds())
	}

	// 版本号
	version := ""
	if hm.harnessConfig != nil {
		version = hm.harnessConfig.System.Version
	}

	health := map[string]interface{}{
		"status":          string(hm.status),
		"timestamp":       time.Now().Format(time.RFC3339),
		"uptime_seconds":  uptimeSeconds,
		"plugins": map[string]interface{}{
			"total":   len(pluginDetails),
			"healthy": healthyCount,
			"details": pluginDetails,
		},
		"pipelines": map[string]interface{}{
			"total": len(hm.pipelines),
			"names": pipelineNames,
		},
		"database": dbStatus,
		"version":  version,
	}

	return health
}

// GetConfig 获取配置对象（供 API 层使用）
func (hm *HarnessManager) GetConfig() *config.HarnessConfig {
	return hm.harnessConfig
}

// GetDB 获取数据库连接（供 API 层使用）
func (hm *HarnessManager) GetDB() *sql.DB {
	if hm.db != nil {
		return hm.db.DB
	}
	return nil
}

// 私有方法

func (hm *HarnessManager) loadConfig() error {
	// 加载配置文件
	cfg, err := config.LoadConfig(hm.configPath)
	if err != nil {
		return fmt.Errorf("加载配置文件失败: %w", err)
	}
	hm.harnessConfig = cfg

	// 初始化数据库连接
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/digital-twin.db"
	}

	db, err := database.NewDatabase(dbPath)
	if err != nil {
		return fmt.Errorf("初始化数据库失败: %w", err)
	}
	hm.db = db

	return nil
}

func (hm *HarnessManager) initEventBus() error {
	// 初始化事件总线
	// 可以使用内存事件总线或Redis等分布式事件总线
	hm.eventBus = NewMemoryEventBus()
	return nil
}

func (hm *HarnessManager) loadPlugins() error {
	if hm.harnessConfig == nil {
		return fmt.Errorf("配置未加载")
	}

	for name, pluginCfg := range hm.harnessConfig.Plugins {
		if !pluginCfg.Enabled {
			continue
		}

		// 使用工厂函数创建插件
		plugin, err := CreatePlugin(name, pluginCfg.Type, hm.db.DB)
		if err != nil {
			return fmt.Errorf("创建插件 %s 失败: %w", name, err)
		}

		// 将嵌套配置展平为 "key.subkey" 格式
		flatConfig := flattenConfig(pluginCfg.Config, "")

		// 初始化插件
		if err := plugin.Init(flatConfig); err != nil {
			return fmt.Errorf("初始化插件 %s 失败: %w", name, err)
		}

		// 直接注册插件（Start 已持有锁，避免死锁）
		if _, exists := hm.plugins[name]; exists {
			return fmt.Errorf("插件 %s 已注册", name)
		}
		hm.plugins[name] = plugin
		hm.metrics.RecordPluginRegistration(name)

		// 同时按 type 注册别名（如 "dialogue" → socratic-dialogue），
		// 方便 handler 按类型查找插件
		if _, exists := hm.plugins[pluginCfg.Type]; !exists {
			hm.plugins[pluginCfg.Type] = plugin
		}
	}

	return nil
}

func (hm *HarnessManager) buildPipelines() error {
	if hm.harnessConfig == nil {
		return fmt.Errorf("配置未加载")
	}

	for name, pipelineCfg := range hm.harnessConfig.Pipelines {
		if _, exists := hm.pipelines[name]; exists {
			return fmt.Errorf("管道 %s 已存在", name)
		}

		pipeline := NewPipeline(name)

		// 添加插件到管道
		for _, pluginName := range pipelineCfg.Plugins {
			plugin, exists := hm.plugins[pluginName]
			if !exists {
				return fmt.Errorf("管道 %s 引用的插件 %s 不存在", name, pluginName)
			}
			if err := pipeline.AddPlugin(plugin); err != nil {
				return fmt.Errorf("添加插件 %s 到管道 %s 失败: %w", pluginName, name, err)
			}
		}

		// 设置超时
		if pipelineCfg.Timeout != "" {
			timeout, err := time.ParseDuration(pipelineCfg.Timeout)
			if err != nil {
				return fmt.Errorf("管道 %s 超时格式无效: %w", name, err)
			}
			pipeline.SetTimeout(timeout)
		}

		hm.pipelines[name] = pipeline
		hm.metrics.RecordPipelineCreation(name)
	}

	return nil
}

func (hm *HarnessManager) startHealthCheck() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 执行健康检查
			health := hm.HealthCheck()
			
			// 发布健康检查事件
			hm.eventBus.Publish(NewHealthCheckEvent(health))
			
			// 记录健康检查指标
			hm.metrics.RecordHealthCheck()
		}
	}
}

// MetricsCollector 指标收集器
type MetricsCollector struct {
	mu sync.RWMutex
	
	// 管理器指标
	managerStarts   int64
	managerStops    int64
	
	// 插件指标
	pluginRegistrations int64
	pluginUnregistrations int64
	
	// 管道指标
	pipelineCreations int64
	pipelineRemovals  int64
	pipelineExecutions int64
	pipelineFailures  int64
	
	// 健康检查
	healthChecks int64
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
}

func (mc *MetricsCollector) RecordManagerStart() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.managerStarts++
}

func (mc *MetricsCollector) RecordManagerStop() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.managerStops++
}

func (mc *MetricsCollector) RecordPluginRegistration(name string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.pluginRegistrations++
}

func (mc *MetricsCollector) RecordPluginUnregistration(name string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.pluginUnregistrations++
}

func (mc *MetricsCollector) RecordPipelineCreation(name string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.pipelineCreations++
}

func (mc *MetricsCollector) RecordPipelineRemoval(name string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.pipelineRemovals++
}

func (mc *MetricsCollector) RecordPipelineExecution(name string, duration time.Duration, success bool) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.pipelineExecutions++
	if !success {
		mc.pipelineFailures++
	}
}

func (mc *MetricsCollector) RecordHealthCheck() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.healthChecks++
}

func (mc *MetricsCollector) GetMetrics() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return map[string]interface{}{
		"manager_starts":           mc.managerStarts,
		"manager_stops":            mc.managerStops,
		"plugin_registrations":     mc.pluginRegistrations,
		"plugin_unregistrations":   mc.pluginUnregistrations,
		"pipeline_creations":       mc.pipelineCreations,
		"pipeline_removals":        mc.pipelineRemovals,
		"pipeline_executions":      mc.pipelineExecutions,
		"pipeline_failures":         mc.pipelineFailures,
		"health_checks":            mc.healthChecks,
	}
}

// flattenConfig 将嵌套的 map 展平为 "key.subkey" 格式
func flattenConfig(m map[string]interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch val := v.(type) {
		case map[string]interface{}:
			for fk, fv := range flattenConfig(val, key) {
				result[fk] = fv
			}
		case map[interface{}]interface{}:
			// YAML 解析可能产生 map[interface{}]interface{} 类型
			converted := make(map[string]interface{})
			for mk, mv := range val {
				converted[fmt.Sprintf("%v", mk)] = mv
			}
			for fk, fv := range flattenConfig(converted, key) {
				result[fk] = fv
			}
		default:
			result[key] = v
		}
	}
	return result
}