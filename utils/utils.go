package utils

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// Logger 日志记录器
type Logger struct {
	EnableLogs bool
}

// NewLogger 创建新的日志记录器
func NewLogger(enableLogs bool) *Logger {
	return &Logger{EnableLogs: enableLogs}
}

// Info 记录信息日志
func (l *Logger) Info(format string, args ...interface{}) {
	if l.EnableLogs {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		message := fmt.Sprintf(format, args...)
		fmt.Printf("[INFO] %s %s\n", timestamp, message)
	}
}

// Error 记录错误日志
func (l *Logger) Error(format string, args ...interface{}) {
	if l.EnableLogs {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		message := fmt.Sprintf(format, args...)
		fmt.Printf("[ERROR] %s %s\n", timestamp, message)
	}
}

// Debug 记录调试日志
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.EnableLogs {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		message := fmt.Sprintf(format, args...)
		fmt.Printf("[DEBUG] %s %s\n", timestamp, message)
	}
}

// Warn 记录警告日志
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.EnableLogs {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		message := fmt.Sprintf(format, args...)
		fmt.Printf("[WARN] %s %s\n", timestamp, message)
	}
}

// GenerateUserID 生成用户ID
func GenerateUserID(conn net.Conn) string {
	addr := conn.RemoteAddr().String()
	// 移除端口号，只保留IP地址
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		addr = addr[:idx]
	}
	return fmt.Sprintf("user_%s_%d", addr, time.Now().Unix())
}

// GenerateUsername 生成默认用户名
func GenerateUsername(conn net.Conn) string {
	addr := conn.RemoteAddr().String()
	// 移除端口号，只保留IP地址
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		addr = addr[:idx]
	}
	return fmt.Sprintf("用户_%s", addr)
}

// ValidateUsername 验证用户名
func ValidateUsername(username string) error {
	if len(username) == 0 {
		return fmt.Errorf("用户名不能为空")
	}
	if len(username) > 20 {
		return fmt.Errorf("用户名长度不能超过20个字符")
	}
	// 检查是否包含特殊字符
	for _, char := range username {
		if char < 32 || char > 126 {
			return fmt.Errorf("用户名包含无效字符")
		}
	}
	return nil
}

// FormatDuration 格式化持续时间
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0f秒", d.Seconds())
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%d分%d秒", minutes, seconds)
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%d小时%d分", hours, minutes)
}

// IsValidIP 检查IP地址是否有效
func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// GetLocalIP 获取本地IP地址
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

// TruncateString 截断字符串
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// SanitizeInput 清理用户输入
func SanitizeInput(input string) string {
	// 移除控制字符
	var result strings.Builder
	for _, char := range input {
		if char >= 32 && char != 127 {
			result.WriteRune(char)
		}
	}
	return strings.TrimSpace(result.String())
}

// RetryWithBackoff 带退避的重试机制
func RetryWithBackoff(maxRetries int, baseDelay time.Duration, fn func() error) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		if i < maxRetries-1 {
			delay := baseDelay * time.Duration(1<<uint(i))
			time.Sleep(delay)
		}
	}
	return lastErr
}
