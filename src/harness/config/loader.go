package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// envVarPattern 匹配 ${VAR_NAME} 和 ${VAR_NAME:-default_value} 格式的环境变量
var envVarPattern = regexp.MustCompile(`\$\{([^}:]+)(?::-([^}]*))?\}`)

// LoadConfig 从指定路径加载并解析配置文件
// 支持 ${VAR_NAME:-default} 格式的环境变量替换
func LoadConfig(path string) (*HarnessConfig, error) {
	// 1. 读取 YAML 文件为字符串
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 2. 环境变量替换
	content := replaceEnvVars(string(data))

	// 3. YAML 反序列化
	var cfg HarnessConfig
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 4. 配置校验
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("配置校验失败: %w", err)
	}

	return &cfg, nil
}

// replaceEnvVars 替换字符串中的 ${VAR_NAME:-default} 为环境变量值
func replaceEnvVars(content string) string {
	return envVarPattern.ReplaceAllStringFunc(content, func(match string) string {
		submatches := envVarPattern.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}

		varName := strings.TrimSpace(submatches[1])
		defaultValue := ""
		if len(submatches) >= 3 {
			defaultValue = submatches[2]
		}

		// 优先使用环境变量值，否则使用默认值
		if envValue, ok := os.LookupEnv(varName); ok {
			return envValue
		}
		return defaultValue
	})
}

// validateConfig 校验配置必填字段
func validateConfig(cfg *HarnessConfig) error {
	// 校验系统配置
	if cfg.System.Name == "" {
		return fmt.Errorf("system.name 不能为空")
	}
	if cfg.System.Version == "" {
		return fmt.Errorf("system.version 不能为空")
	}

	// 校验插件配置
	for name, plugin := range cfg.Plugins {
		if plugin.Type == "" {
			return fmt.Errorf("插件 %s 的 type 不能为空", name)
		}

		// 认证插件校验
		if name == "authentication" && plugin.Enabled {
			if jwtConfig, ok := plugin.Config["jwt"].(map[string]interface{}); ok {
				secret, _ := jwtConfig["secret"].(string)
				if secret == "" || secret == "default-secret-key" {
					return fmt.Errorf("认证插件的 jwt.secret 未配置或使用了默认值")
				}
			}
		}

		// 对话插件校验
		if name == "socratic-dialogue" && plugin.Enabled {
			if llmConfig, ok := plugin.Config["llm_provider"].(map[string]interface{}); ok {
				mode, _ := llmConfig["mode"].(string)
				if mode == "api" {
					apiKey, _ := llmConfig["api_key"].(string)
					if apiKey == "" {
						return fmt.Errorf("对话插件 API 模式下 api_key 不能为空")
					}
				}
			}
		}
	}

	// 校验管道配置
	for name, pipeline := range cfg.Pipelines {
		if len(pipeline.Plugins) == 0 {
			return fmt.Errorf("管道 %s 的 plugins 列表不能为空", name)
		}

		// 校验管道引用的插件是否存在且启用
		for _, pluginName := range pipeline.Plugins {
			pluginCfg, exists := cfg.Plugins[pluginName]
			if !exists {
				return fmt.Errorf("管道 %s 引用的插件 %s 未定义", name, pluginName)
			}
			if !pluginCfg.Enabled {
				return fmt.Errorf("管道 %s 引用的插件 %s 未启用", name, pluginName)
			}
		}

		// 校验超时格式
		if pipeline.Timeout != "" {
			if _, err := time.ParseDuration(pipeline.Timeout); err != nil {
				return fmt.Errorf("管道 %s 的 timeout 格式无效: %s", name, pipeline.Timeout)
			}
		}
	}

	return nil
}
