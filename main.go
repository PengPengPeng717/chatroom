package main

import (
	"flag"
	"fmt"
	"os"

	"chatroom/config"
	"chatroom/server"
)

func main() {
	// 解析命令行参数
	var (
		host     = flag.String("host", "127.0.0.1", "服务器监听地址")
		port     = flag.Int("port", 8080, "服务器监听端口")
		maxUsers = flag.Int("max-users", 100, "最大用户数")
		timeout  = flag.Int("timeout", 40, "用户超时时间(秒)")
		help     = flag.Bool("help", false, "显示帮助信息")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// 创建配置
	cfg := config.DefaultConfig()
	cfg.Host = *host
	cfg.Port = *port
	cfg.MaxUsers = *maxUsers
	cfg.Timeout = *timeout

	// 从环境变量加载配置
	cfg.LoadFromEnv()

	// 验证配置
	if err := cfg.Validate(); err != nil {
		fmt.Printf("配置错误: %v\n", err)
		os.Exit(1)
	}

	// 创建并启动服务器
	chatServer := server.NewChatServer(cfg)

	fmt.Println("=== Go聊天室服务器 ===")
	fmt.Printf("监听地址: %s\n", cfg.GetAddress())
	fmt.Printf("最大用户数: %d\n", cfg.MaxUsers)
	fmt.Printf("超时时间: %d秒\n", cfg.Timeout)
	fmt.Println("按 Ctrl+C 停止服务器")
	fmt.Println("=====================")

	if err := chatServer.Start(); err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
		os.Exit(1)
	}
}

// showHelp 显示帮助信息
func showHelp() {
	fmt.Println("Go聊天室服务器")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  chatroom [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -host string")
	fmt.Println("        服务器监听地址 (默认: 127.0.0.1)")
	fmt.Println("  -port int")
	fmt.Println("        服务器监听端口 (默认: 8080)")
	fmt.Println("  -max-users int")
	fmt.Println("        最大用户数 (默认: 100)")
	fmt.Println("  -timeout int")
	fmt.Println("        用户超时时间，单位秒 (默认: 40)")
	fmt.Println("  -help")
	fmt.Println("        显示此帮助信息")
	fmt.Println()
	fmt.Println("环境变量:")
	fmt.Println("  CHATROOM_HOST      服务器监听地址")
	fmt.Println("  CHATROOM_PORT      服务器监听端口")
	fmt.Println("  CHATROOM_MAX_USERS 最大用户数")
	fmt.Println("  CHATROOM_TIMEOUT   用户超时时间")
	fmt.Println("  CHATROOM_LOG_LEVEL 日志级别")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  chatroom -host 0.0.0.0 -port 9000 -max-users 50")
	fmt.Println("  CHATROOM_PORT=9000 chatroom")
}
