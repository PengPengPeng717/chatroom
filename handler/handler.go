package handler

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"chatroom/config"
	"chatroom/message"
	"chatroom/user"
	"chatroom/utils"
)

// ConnectionHandler 连接处理器
type ConnectionHandler struct {
	userManager   *user.UserManager
	commandParser *message.CommandParser
	logger        *utils.Logger
	config        *config.Config
}

// NewConnectionHandler 创建新的连接处理器
func NewConnectionHandler(userManager *user.UserManager, logger *utils.Logger, cfg *config.Config) *ConnectionHandler {
	return &ConnectionHandler{
		userManager:   userManager,
		commandParser: message.NewCommandParser(),
		logger:        logger,
		config:        cfg,
	}
}

// HandleConnection 处理客户端连接
func (ch *ConnectionHandler) HandleConnection(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	ch.logger.Info("客户端已连接: %s", clientAddr)

	// 生成用户ID和默认用户名
	userID := utils.GenerateUserID(conn)
	defaultUsername := utils.GenerateUsername(conn)

	// 创建用户
	currentUser, err := ch.userManager.CreateUser(userID, defaultUsername)
	if err != nil {
		ch.logger.Error("创建用户失败: %v", err)
		conn.Write([]byte(fmt.Sprintf("错误: %s\n", err.Error())))
		return
	}

	ch.logger.Info("用户 %s (ID: %s) 已创建", currentUser.Name, currentUser.ID)

	// 发送欢迎消息
	welcomeMsg := message.GetWelcomeMessage()
	ch.userManager.SendToUser(currentUser.ID, welcomeMsg)

	// 广播用户加入消息
	joinMsg := message.FormatUserJoinMessage(currentUser.Name)
	ch.userManager.BroadcastToOthers(currentUser.ID, joinMsg)

	// 启动消息写入协程
	go ch.writeToClient(currentUser, conn)

	// 启动超时监控协程
	timeoutChan := make(chan bool, 1)
	go ch.watchTimeout(currentUser, timeoutChan)

	// 处理客户端消息
	ch.handleClientMessages(currentUser, conn, timeoutChan)
}

// handleClientMessages 处理客户端消息
func (ch *ConnectionHandler) handleClientMessages(currentUser *user.User, conn net.Conn, timeoutChan chan bool) {
	reader := bufio.NewReader(conn)

	for {
		// 设置读取超时
		conn.SetReadDeadline(time.Now().Add(time.Duration(ch.config.Timeout) * time.Second))

		// 读取客户端数据
		data, err := reader.ReadString('\n')
		if err != nil {
			ch.logger.Info("用户 %s 断开连接: %v", currentUser.Name, err)
			break
		}

		// 更新用户最后活跃时间
		ch.userManager.UpdateUserLastSeen(currentUser.ID)

		// 清理输入数据
		input := utils.SanitizeInput(strings.TrimSpace(data))
		if len(input) == 0 {
			continue
		}

		ch.logger.Debug("收到来自 %s 的消息: %s", currentUser.Name, input)

		// 处理用户输入
		if err := ch.processUserInput(currentUser, input); err != nil {
			ch.logger.Error("处理用户输入失败: %v", err)
			ch.userManager.SendToUser(currentUser.ID, fmt.Sprintf("错误: %s\n", err.Error()))
		}

		// 检查用户是否请求退出
		select {
		case <-currentUser.DoneChan:
			ch.logger.Info("用户 %s 请求退出", currentUser.Name)
			return
		default:
		}
	}
}

// processUserInput 处理用户输入
func (ch *ConnectionHandler) processUserInput(currentUser *user.User, input string) error {
	// 解析命令
	cmd, err := ch.commandParser.ParseCommand(input)
	if err != nil {
		return err
	}

	// 处理不同类型的命令
	switch cmd.Type {
	case message.CmdChat:
		// 普通聊天消息
		chatMsg := message.NewMessage(message.TypeChat, currentUser.Name, cmd.Content)
		ch.userManager.BroadcastToAll(chatMsg.FormatMessage())
		ch.logger.Info("用户 %s 发送消息: %s", currentUser.Name, utils.TruncateString(cmd.Content, 50))

	case message.CmdWho:
		// 查询在线用户
		userList := ch.userManager.GetUserList()
		ch.userManager.SendToUser(currentUser.ID, userList)

	case message.CmdRename:
		// 重命名
		oldName := currentUser.Name
		if err := ch.userManager.RenameUser(currentUser.ID, cmd.Content); err != nil {
			return err
		}

		// 更新当前用户对象
		if updatedUser, exists := ch.userManager.GetUser(currentUser.ID); exists {
			*currentUser = *updatedUser
		}

		// 发送成功消息
		ch.userManager.SendToUser(currentUser.ID, fmt.Sprintf("用户名已更改为: %s\n", currentUser.Name))

		// 广播重命名消息
		renameMsg := message.FormatUserRenameMessage(oldName, currentUser.Name)
		ch.userManager.BroadcastToOthers(currentUser.ID, renameMsg)

	case message.CmdHelp:
		// 显示帮助信息
		helpMsg := message.GetHelpMessage()
		ch.userManager.SendToUser(currentUser.ID, helpMsg+"\n")

	case message.CmdTime:
		// 显示当前时间
		timeMsg := message.FormatTimeMessage()
		ch.userManager.SendToUser(currentUser.ID, timeMsg+"\n")

	case message.CmdStats:
		// 显示统计信息
		statsMsg := message.FormatStatsMessage(ch.userManager.GetUserCount(), ch.config.MaxUsers)
		ch.userManager.SendToUser(currentUser.ID, statsMsg+"\n")

	case message.CmdWhisper:
		// 私聊消息
		if err := ch.handleWhisper(currentUser, cmd.Target, cmd.Content); err != nil {
			return err
		}

	case message.CmdQuit:
		// 退出聊天室
		ch.userManager.SendToUser(currentUser.ID, "正在退出聊天室...\n")
		close(currentUser.DoneChan)
		return nil

	default:
		return fmt.Errorf("未知命令类型")
	}

	return nil
}

// writeToClient 向客户端写入消息
func (ch *ConnectionHandler) writeToClient(currentUser *user.User, conn net.Conn) {
	defer func() {
		ch.logger.Debug("用户 %s 的消息写入协程已退出", currentUser.Name)
	}()

	for {
		select {
		case msg, ok := <-currentUser.MsgChan:
			if !ok {
				return
			}

			// 设置写入超时
			conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

			_, err := conn.Write([]byte(msg))
			if err != nil {
				ch.logger.Error("向用户 %s 发送消息失败: %v", currentUser.Name, err)
				close(currentUser.DoneChan)
				return
			}

		case <-currentUser.DoneChan:
			return
		}
	}
}

// watchTimeout 监控用户超时
func (ch *ConnectionHandler) watchTimeout(currentUser *user.User, timeoutChan chan bool) {
	ticker := time.NewTicker(time.Duration(ch.config.Timeout) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 检查用户是否超时
			if user, exists := ch.userManager.GetUser(currentUser.ID); exists {
				timeSinceLastSeen := time.Since(user.LastSeen)
				if timeSinceLastSeen > time.Duration(ch.config.Timeout)*time.Second {
					ch.logger.Info("用户 %s 超时，自动断开连接", currentUser.Name)
					timeoutChan <- true
					close(currentUser.DoneChan)
					return
				}
			} else {
				// 用户已被移除
				return
			}

		case <-currentUser.DoneChan:
			return
		}
	}
}

// handleWhisper 处理私聊消息
func (ch *ConnectionHandler) handleWhisper(fromUser *user.User, targetName, content string) error {
	// 查找目标用户
	var targetUser *user.User
	users := ch.userManager.GetAllUsers()
	for _, user := range users {
		if user.Name == targetName {
			targetUser = user
			break
		}
	}

	if targetUser == nil {
		return fmt.Errorf("用户 %s 不在线", targetName)
	}

	// 创建私聊消息
	whisperMsg := message.NewPrivateMessage(fromUser.Name, targetUser.Name, content)

	// 发送给目标用户
	if err := ch.userManager.SendToUser(targetUser.ID, whisperMsg.FormatMessage()); err != nil {
		return fmt.Errorf("发送私聊消息失败: %v", err)
	}

	// 发送确认消息给发送者
	confirmMsg := fmt.Sprintf("[私聊] -> %s: %s\n", targetUser.Name, content)
	ch.userManager.SendToUser(fromUser.ID, confirmMsg)

	ch.logger.Info("用户 %s 向 %s 发送私聊消息", fromUser.Name, targetUser.Name)
	return nil
}

// CleanupUser 清理用户资源
func (ch *ConnectionHandler) CleanupUser(currentUser *user.User) {
	// 移除用户
	if removedUser, exists := ch.userManager.RemoveUser(currentUser.ID); exists {
		// 广播用户离开消息
		leaveMsg := message.FormatUserLeaveMessage(removedUser.Name)
		ch.userManager.BroadcastToAll(leaveMsg)
		ch.logger.Info("用户 %s 已离开聊天室", removedUser.Name)
	}
}
