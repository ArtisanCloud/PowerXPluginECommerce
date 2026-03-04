package config

import (
	"os"
	"strings"
)

const (
	RuntimeDriverAuto      = "auto"
	RuntimeDriverLocal     = "local"
	RuntimeDriverFramework = "framework"
)

// RuntimeDriversConfig controls runtime subsystem driver selection.
// Values: auto | local | framework
type RuntimeDriversConfig struct {
	WebSocket  string `yaml:"websocket" json:"websocket"`
	EventTopic string `yaml:"event_topic" json:"event_topic"`
	Task       string `yaml:"task" json:"task"`
	Cache      string `yaml:"cache" json:"cache"`
}

func (c *Config) ensureRuntimeDriversConfig() *RuntimeDriversConfig {
	if c == nil {
		return &RuntimeDriversConfig{}
	}
	if c.Runtime == nil {
		c.Runtime = &RuntimeConfig{}
	}
	if c.Runtime.Drivers == nil {
		c.Runtime.Drivers = &RuntimeDriversConfig{}
	}
	return c.Runtime.Drivers
}

func (c *Config) ResolveWebSocketDriver() string {
	return c.resolveRuntimeDriver(c.runtimeDriverValue("websocket"), "POWERX_RUNTIME_DRIVER_WEBSOCKET")
}

func (c *Config) ResolveEventTopicDriver() string {
	return c.resolveRuntimeDriver(c.runtimeDriverValue("event_topic"), "POWERX_RUNTIME_DRIVER_EVENT_TOPIC")
}

func (c *Config) ResolveTaskDriver() string {
	return c.resolveRuntimeDriver(c.runtimeDriverValue("task"), "POWERX_RUNTIME_DRIVER_TASK")
}

func (c *Config) ResolveCacheDriver() string {
	return c.resolveRuntimeDriver(c.runtimeDriverValue("cache"), "POWERX_RUNTIME_DRIVER_CACHE")
}

func (c *Config) runtimeDriverValue(kind string) string {
	drivers := c.ensureRuntimeDriversConfig()
	switch kind {
	case "websocket":
		return drivers.WebSocket
	case "event_topic":
		return drivers.EventTopic
	case "task":
		return drivers.Task
	case "cache":
		return drivers.Cache
	default:
		return ""
	}
}

func (c *Config) resolveRuntimeDriver(value string, envKey string) string {
	if envValue := strings.TrimSpace(os.Getenv(envKey)); envValue != "" {
		return normalizeRuntimeDriver(envValue)
	}
	driver := normalizeRuntimeDriver(value)
	if driver != RuntimeDriverAuto {
		return driver
	}
	if strings.TrimSpace(os.Getenv("POWERX_PROXY")) == "1" {
		return RuntimeDriverFramework
	}
	return RuntimeDriverLocal
}

func normalizeRuntimeDriver(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", RuntimeDriverAuto:
		return RuntimeDriverAuto
	case "host", "powerx", RuntimeDriverFramework:
		return RuntimeDriverFramework
	case RuntimeDriverLocal:
		return RuntimeDriverLocal
	default:
		return RuntimeDriverAuto
	}
}
