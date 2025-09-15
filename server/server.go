package server

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"chatroom/config"
	"chatroom/handler"
	"chatroom/user"
	"chatroom/utils"
)

// ChatServer 聊天服务器
type ChatServer struct {
	config            *config.Config
	userManager       *user.UserManager
	connectionHandler *handler.ConnectionHandler
	logger            *utils.Logger
	listener          net.Listener
	isRunning         bool
}

// NewChatServer 创建新的聊天服务器
func NewChatServer(cfg *config.Config) *ChatServer {
	logger := utils.NewLogger(cfg.EnableLogs)
	userManager := user.NewUserManager(cfg.MaxUsers)
	connectionHandler := handler.NewConnectionHandler(userManager, logger, cfg)

	return &ChatServer{
		config:            cfg,
		userManager:       userManager,
		connectionHandler: connectionHandler,
		logger:            logger,
		isRunning:         false,
	}
}

// Start 启动服务器
func (s *ChatServer) Start() error {
	// 验证配置
	if err := s.config.Validate(); err != nil {
		return fmt.Errorf("配置验证失败: %v", err)
	}

	// 创建监听器
	listener, err := net.Listen("tcp", s.config.GetAddress())
	if err != nil {
		return fmt.Errorf("启动服务器失败: %v", err)
	}
	s.listener = listener
	s.isRunning = true

	s.logger.Info("聊天服务器启动成功，监听地址: %s", s.config.GetAddress())
	s.logger.Info("最大用户数: %d, 超时时间: %d秒", s.config.MaxUsers, s.config.Timeout)

	// 启动信号处理
	go s.handleSignals()

	// 启动连接处理循环
	return s.acceptConnections()
}

// acceptConnections 接受连接
func (s *ChatServer) acceptConnections() error {
	for s.isRunning {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.isRunning {
				s.logger.Error("接受连接失败: %v", err)
			}
			continue
		}

		// 检查用户数量限制
		if s.userManager.GetUserCount() >= s.config.MaxUsers {
			s.logger.Warn("聊天室已满，拒绝新连接")
			conn.Write([]byte("聊天室已满，请稍后再试\n"))
			conn.Close()
			continue
		}

		// 启动新的goroutine处理连接
		go s.connectionHandler.HandleConnection(conn)
	}

	return nil
}

// Stop 停止服务器
func (s *ChatServer) Stop() {
	s.logger.Info("正在停止服务器...")
	s.isRunning = false

	if s.listener != nil {
		s.listener.Close()
	}

	// 断开所有用户连接
	users := s.userManager.GetAllUsers()
	for _, user := range users {
		close(user.DoneChan)
	}

	s.logger.Info("服务器已停止")
}

// handleSignals 处理系统信号
func (s *ChatServer) handleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	s.logger.Info("收到停止信号")
	s.Stop()
}

// GetStats 获取服务器统计信息
func (s *ChatServer) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"isRunning":    s.isRunning,
		"currentUsers": s.userManager.GetUserCount(),
		"maxUsers":     s.config.MaxUsers,
		"address":      s.config.GetAddress(),
		"timeout":      s.config.Timeout,
	}
}

// BroadcastMessage 广播消息
func (s *ChatServer) BroadcastMessage(message string) {
	s.userManager.BroadcastToAll(message)
	s.logger.Info("系统广播消息: %s", message)
}
