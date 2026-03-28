package core

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Plugin 插件基础接口
type Plugin interface {
	// Name 返回插件名称
	Name() string
	
	// Version 返回插件版本
	Version() string
	
	// Type 返回插件类型
	Type() PluginType
	
	// Init 初始化插件
	Init(config map[string]interface{}) error
	
	// Execute 执行插件功能
	Execute(ctx context.Context, input *PluginInput) (*PluginOutput, error)
	
	// HealthCheck 健康检查
	HealthCheck() error
	
	// Destroy 销毁插件，释放资源
	Destroy() error
	
	// GetConfig 获取当前配置
	GetConfig() map[string]interface{}
	
	// UpdateConfig 更新配置
	UpdateConfig(config map[string]interface{}) error
}

// PluginType 插件类型枚举
type PluginType string

const (
	PluginTypeKnowledge PluginType = "knowledge"
	PluginTypeDialogue  PluginType = "dialogue"
	PluginTypeMemory    PluginType = "memory"
	PluginTypeAuth      PluginType = "auth"
	PluginTypeAnalytics PluginType = "analytics"
	PluginTypeExport    PluginType = "export"
)

// PluginInput 插件输入参数
type PluginInput struct {
	// 请求ID，用于追踪
	RequestID string `json:"request_id"`
	
	// 用户上下文
	UserContext *UserContext `json:"user_context"`
	
	// 输入数据
	Data map[string]interface{} `json:"data"`
	
	// 插件特定参数
	Params map[string]interface{} `json:"params"`
	
	// 执行上下文（包含超时等）
	Context context.Context `json:"-"`
}

// PluginOutput 插件输出结果
type PluginOutput struct {
	// 是否成功
	Success bool `json:"success"`
	
	// 输出数据
	Data map[string]interface{} `json:"data"`
	
	// 错误信息（如果Success为false）
	Error string `json:"error,omitempty"`
	
	// 执行耗时
	Duration time.Duration `json:"duration"`
	
	// 插件元数据
	Metadata map[string]interface{} `json:"metadata"`
}

// UserContext 用户上下文
type UserContext struct {
	// 用户ID
	UserID string `json:"user_id"`
	
	// 用户角色
	Role string `json:"role"`
	
	// 会话ID
	SessionID string `json:"session_id"`
	
	// 租户ID（多租户支持）
	TenantID string `json:"tenant_id"`
	
	// 用户属性
	Attributes map[string]interface{} `json:"attributes"`
}

// PluginInfo 插件信息
type PluginInfo struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Type        PluginType             `json:"type"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	Enabled     bool                   `json:"enabled"`
	Priority    int                    `json:"priority"`
	Config      map[string]interface{} `json:"config"`
	Status      PluginStatus           `json:"status"`
	LastError   string                 `json:"last_error,omitempty"`
}

// PluginStatus 插件状态
type PluginStatus string

const (
	PluginStatusStopped   PluginStatus = "stopped"
	PluginStatusStarting  PluginStatus = "starting"
	PluginStatusRunning   PluginStatus = "running"
	PluginStatusStopping  PluginStatus = "stopping"
	PluginStatusError     PluginStatus = "error"
)

// BasePlugin 插件基础实现
type BasePlugin struct {
	name    string
	version string
	ptype   PluginType
	config  map[string]interface{}
	status  PluginStatus
}

// NewBasePlugin 创建基础插件
func NewBasePlugin(name, version string, ptype PluginType) *BasePlugin {
	return &BasePlugin{
		name:    name,
		version: version,
		ptype:   ptype,
		config:  make(map[string]interface{}),
		status:  PluginStatusStopped,
	}
}

func (p *BasePlugin) Name() string {
	return p.name
}

func (p *BasePlugin) Version() string {
	return p.version
}

func (p *BasePlugin) Type() PluginType {
	return p.ptype
}

func (p *BasePlugin) Init(config map[string]interface{}) error {
	p.config = config
	p.status = PluginStatusRunning
	return nil
}

func (p *BasePlugin) HealthCheck() error {
	if p.status != PluginStatusRunning {
		return fmt.Errorf("plugin %s is not running", p.name)
	}
	return nil
}

func (p *BasePlugin) GetConfig() map[string]interface{} {
	return p.config
}

func (p *BasePlugin) UpdateConfig(config map[string]interface{}) error {
	p.config = config
	return nil
}

func (p *BasePlugin) Destroy() error {
	p.status = PluginStatusStopped
	return nil
}

// PluginRegistry 插件注册表接口
type PluginRegistry interface {
	// Register 注册插件
	Register(plugin Plugin) error
	
	// Unregister 注销插件
	Unregister(name string) error
	
	// GetPlugin 获取插件
	GetPlugin(name string) (Plugin, error)
	
	// ListPlugins 列出所有插件
	ListPlugins() []PluginInfo
	
	// GetPluginsByType 按类型获取插件
	GetPluginsByType(ptype PluginType) []Plugin
}

// Pipeline 管道接口
type Pipeline interface {
	// Name 管道名称
	Name() string
	
	// Execute 执行管道
	Execute(ctx context.Context, input *PluginInput) (*PluginOutput, error)
	
	// AddPlugin 添加插件到管道
	AddPlugin(plugin Plugin) error
	
	// RemovePlugin 从管道移除插件
	RemovePlugin(pluginName string) error
	
	// GetPlugins 获取管道中的插件
	GetPlugins() []Plugin
}

// PipelineConfig 管道配置
type PipelineConfig struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Plugins     []string `json:"plugins"`
	Timeout     string   `json:"timeout"`
}

// Event 事件接口（用于插件间通信）
type Event interface {
	Name() string
	Data() interface{}
	Timestamp() time.Time
}

// EventBus 事件总线接口
type EventBus interface {
	// Publish 发布事件
	Publish(event Event) error
	
	// Subscribe 订阅事件
	Subscribe(eventName string, handler func(Event)) error
	
	// Unsubscribe 取消订阅
	Unsubscribe(eventName string, handler func(Event)) error
}

// 错误定义
var (
	ErrPluginNotFound    = errors.New("plugin not found")
	ErrPluginNotEnabled  = errors.New("plugin not enabled")
	ErrPluginInitFailed  = errors.New("plugin initialization failed")
	ErrPluginExecuteFailed = errors.New("plugin execution failed")
	ErrPipelineNotFound  = errors.New("pipeline not found")
	ErrInvalidPluginType = errors.New("invalid plugin type")
)

// 工具函数

// NewPluginInput 创建插件输入
func NewPluginInput(requestID string, userContext *UserContext) *PluginInput {
	return &PluginInput{
		RequestID:   requestID,
		UserContext: userContext,
		Data:        make(map[string]interface{}),
		Params:      make(map[string]interface{}),
		Context:     context.Background(),
	}
}

// NewPluginOutput 创建插件输出
func NewPluginOutput(success bool, data map[string]interface{}) *PluginOutput {
	return &PluginOutput{
		Success:  success,
		Data:     data,
		Metadata: make(map[string]interface{}),
	}
}

// WithError 为输出添加错误信息
func (po *PluginOutput) WithError(err error) *PluginOutput {
	po.Error = err.Error()
	po.Success = false
	return po
}

// WithMetadata 添加元数据
func (po *PluginOutput) WithMetadata(key string, value interface{}) *PluginOutput {
	if po.Metadata == nil {
		po.Metadata = make(map[string]interface{})
	}
	po.Metadata[key] = value
	return po
}