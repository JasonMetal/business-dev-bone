package config

import "business-dev-bone/internal/core/options"

// Config 运行期配置，仅依赖 Options，无业务字段
type Config struct {
	*options.Options
}

// CreateFromOptions 从命令行/配置文件得到的 Options 生成 Config
func CreateFromOptions(opts *options.Options) (*Config, error) {
	return &Config{Options: opts}, nil
}
