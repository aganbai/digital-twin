package manager

import (
	"context"
	"fmt"
	"time"

	"digital-twin/src/harness/core"
)

// Pipeline 管道实现
type Pipeline struct {
	name     string
	plugins  []core.Plugin
	timeout  time.Duration
}

// NewPipeline 创建新管道
func NewPipeline(name string) *Pipeline {
	return &Pipeline{
		name:    name,
		plugins: make([]core.Plugin, 0),
		timeout: 30 * time.Second, // 默认超时时间
	}
}

// Name 返回管道名称
func (p *Pipeline) Name() string {
	return p.name
}

// Execute 执行管道
func (p *Pipeline) Execute(ctx context.Context, input *core.PluginInput) (*core.PluginOutput, error) {
	startTime := time.Now()
	var currentInput = input
	
	for _, plugin := range p.plugins {
		// 检查 context 是否已超时或取消
		if ctx.Err() != nil {
			return &core.PluginOutput{
				Success: false,
				Data:    map[string]interface{}{"error_code": 50004},
				Error:   fmt.Sprintf("管道执行超时: %v", ctx.Err()),
			}, fmt.Errorf("pipeline timeout before plugin %s: %v", plugin.Name(), ctx.Err())
		}

		pluginStart := time.Now()
		output, err := plugin.Execute(ctx, currentInput)
		pluginDuration := time.Since(pluginStart)

		if err != nil {
			return nil, fmt.Errorf("plugin %s execution failed (took %v): %v", plugin.Name(), pluginDuration, err)
		}

		if !output.Success {
			return output, nil
		}
		
		// 将当前插件的输出作为下一个插件的输入
		// 使用 currentInput.UserContext 而非 input.UserContext，确保上一个插件对 UserContext 的修改能传递下去
		currentInput = &core.PluginInput{
			RequestID:   input.RequestID,
			UserContext: currentInput.UserContext,
			Data:        output.Data,
			Params:      input.Params,
			Context:     input.Context,
		}
	}
	
	return &core.PluginOutput{
		Success:  true,
		Data:     currentInput.Data,
		Duration: time.Since(startTime),
	}, nil
}

// AddPlugin 添加插件到管道
func (p *Pipeline) AddPlugin(plugin core.Plugin) error {
	p.plugins = append(p.plugins, plugin)
	return nil
}

// RemovePlugin 从管道移除插件
func (p *Pipeline) RemovePlugin(pluginName string) error {
	for i, plugin := range p.plugins {
		if plugin.Name() == pluginName {
			p.plugins = append(p.plugins[:i], p.plugins[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("plugin %s not found in pipeline", pluginName)
}

// GetPlugins 获取管道中的插件
func (p *Pipeline) GetPlugins() []core.Plugin {
	return p.plugins
}

// Timeout 获取管道超时时间
func (p *Pipeline) Timeout() time.Duration {
	return p.timeout
}

// SetTimeout 设置管道超时时间
func (p *Pipeline) SetTimeout(timeout time.Duration) {
	p.timeout = timeout
}