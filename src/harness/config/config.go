package config

// HarnessConfig Harness 模式总配置结构体
type HarnessConfig struct {
	System       SystemConfig                `yaml:"system"`
	Plugins      map[string]PluginConfig     `yaml:"plugins"`
	Pipelines    map[string]PipelineConfig   `yaml:"pipelines"`
	FeatureFlags FeatureFlagsConfig          `yaml:"feature_flags"`
	Performance  PerformanceConfig           `yaml:"performance"`
	Security     SecurityConfig              `yaml:"security"`
}

// SystemConfig 系统配置
type SystemConfig struct {
	Name        string           `yaml:"name"`
	Version     string           `yaml:"version"`
	Environment string           `yaml:"environment"`
	Debug       bool             `yaml:"debug"`
	Logging     LoggingConfig    `yaml:"logging"`
	Monitoring  MonitoringConfig `yaml:"monitoring"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	Enabled            bool   `yaml:"enabled"`
	MetricsPort        int    `yaml:"metrics_port"`
	HealthCheckInterval string `yaml:"health_check_interval"`
}

// PluginConfig 插件配置
type PluginConfig struct {
	Enabled  bool                   `yaml:"enabled"`
	Type     string                 `yaml:"type"`
	Priority int                    `yaml:"priority"`
	Config   map[string]interface{} `yaml:"config"`
}

// PipelineConfig 管道配置
type PipelineConfig struct {
	Description string   `yaml:"description"`
	Plugins     []string `yaml:"plugins"`
	Timeout     string   `yaml:"timeout"`
}

// FeatureFlagsConfig 功能开关配置
type FeatureFlagsConfig struct {
	EnableAnalytics          bool `yaml:"enable_analytics"`
	EnableExport             bool `yaml:"enable_export"`
	EnableMultiTenant        bool `yaml:"enable_multi_tenant"`
	EnablePluginMarketplace  bool `yaml:"enable_plugin_marketplace"`
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	MaxConcurrentPlugins int              `yaml:"max_concurrent_plugins"`
	PluginTimeout        string           `yaml:"plugin_timeout"`
	CacheTTL             string           `yaml:"cache_ttl"`
	RateLimiting         RateLimitingConfig `yaml:"rate_limiting"`
}

// RateLimitingConfig 限流配置
type RateLimitingConfig struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerMinute int  `yaml:"requests_per_minute"`
	BurstSize         int  `yaml:"burst_size"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	CORS       CORSConfig       `yaml:"cors"`
	Encryption EncryptionConfig `yaml:"encryption"`
	Audit      AuditConfig      `yaml:"audit"`
}

// CORSConfig CORS 配置
type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods"`
	AllowedHeaders []string `yaml:"allowed_headers"`
}

// EncryptionConfig 数据加密配置
type EncryptionConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Algorithm string `yaml:"algorithm"`
}

// AuditConfig 审计日志配置
type AuditConfig struct {
	Enabled       bool   `yaml:"enabled"`
	LogLevel      string `yaml:"log_level"`
	RetentionDays int    `yaml:"retention_days"`
}
