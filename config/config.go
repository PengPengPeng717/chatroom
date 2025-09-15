package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config 服务器配置
type Config struct {
	Host       string
	Port       int
	MaxUsers   int
	Timeout    int
	BufferSize int
	LogLevel   string
	EnableLogs bool
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Host:       "127.0.0.1",
		Port:       8080,
		MaxUsers:   100,
		Timeout:    40,
		BufferSize: 1024,
		LogLevel:   "INFO",
		EnableLogs: true,
	}
}

// LoadFromEnv 从环境变量加载配置
func (c *Config) LoadFromEnv() {
	if host := os.Getenv("CHATROOM_HOST"); host != "" {
		c.Host = host
	}

	if portStr := os.Getenv("CHATROOM_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			c.Port = port
		}
	}

	if maxUsersStr := os.Getenv("CHATROOM_MAX_USERS"); maxUsersStr != "" {
		if maxUsers, err := strconv.Atoi(maxUsersStr); err == nil {
			c.MaxUsers = maxUsers
		}
	}

	if timeoutStr := os.Getenv("CHATROOM_TIMEOUT"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil {
			c.Timeout = timeout
		}
	}

	if logLevel := os.Getenv("CHATROOM_LOG_LEVEL"); logLevel != "" {
		c.LogLevel = logLevel
	}
}

// GetAddress 获取服务器地址
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("端口号必须在1-65535之间")
	}
	if c.MaxUsers < 1 {
		return fmt.Errorf("最大用户数必须大于0")
	}
	if c.Timeout < 1 {
		return fmt.Errorf("超时时间必须大于0")
	}
	return nil
}
