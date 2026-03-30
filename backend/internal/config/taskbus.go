package config

import "strings"

// TaskBusConfig 描述 TaskBus 的运行参数。
type TaskBusConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
	// Adapter supports: framework (recommended), local (standalone fallback), noop(disabled fallback)
	Adapter     string `yaml:"adapter" json:"adapter"`
	TopicPrefix string `yaml:"topic_prefix" json:"topic_prefix"`
}

// TaskBusEnabled returns whether TaskBus should be enabled.
func (c *Config) TaskBusEnabled() bool {
	if c == nil || c.TaskBus == nil {
		return false
	}
	return c.TaskBus.Enabled
}

// TaskBusTopicPrefix returns configured prefix or a default value.
func (c *Config) TaskBusTopicPrefix() string {
	if c == nil || c.TaskBus == nil || strings.TrimSpace(c.TaskBus.TopicPrefix) == "" {
		return "powerx.channel"
	}
	return strings.TrimSuffix(c.TaskBus.TopicPrefix, ".")
}

// TaskBusAdapter returns adapter name (lower-case).
func (c *Config) TaskBusAdapter() string {
	if c == nil || c.TaskBus == nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(c.TaskBus.Adapter))
}
