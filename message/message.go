package message

import (
	"fmt"
	"strings"
	"time"
)

// MessageType 消息类型
type MessageType int

const (
	TypeChat      MessageType = iota // 聊天消息
	TypeSystem                       // 系统消息
	TypeCommand                      // 命令消息
	TypePrivate                      // 私聊消息
	TypeBroadcast                    // 广播消息
)

// Message 消息结构体
type Message struct {
	Type      MessageType // 消息类型
	From      string      // 发送者
	To        string      // 目标用户
	Content   string      // 消息内容
	Timestamp time.Time   // 时间戳
}

// NewMessage 创建新消息
func NewMessage(msgType MessageType, from, content string) *Message {
	return &Message{
		Type:      msgType,
		From:      from,
		Content:   content,
		Timestamp: time.Now(),
	}
}

// NewPrivateMessage 创建私聊消息
func NewPrivateMessage(from, to, content string) *Message {
	return &Message{
		Type:      TypePrivate,
		From:      from,
		To:        to,
		Content:   content,
		Timestamp: time.Now(),
	}
}

// NewSystemMessage 创建系统消息
func NewSystemMessage(content string) *Message {
	return &Message{
		Type:      TypeSystem,
		Content:   content,
		Timestamp: time.Now(),
	}
}

// FormatMessage 格式化消息
func (m *Message) FormatMessage() string {
	switch m.Type {
	case TypeSystem:
		return fmt.Sprintf("[系统] %s\n", m.Content)
	case TypeChat:
		return fmt.Sprintf("[%s] %s\n", m.From, m.Content)
	case TypePrivate:
		return fmt.Sprintf("[私聊] %s -> %s: %s\n", m.From, m.To, m.Content)
	case TypeBroadcast:
		return fmt.Sprintf("[广播] %s: %s\n", m.From, m.Content)
	default:
		return fmt.Sprintf("[%s] %s\n", m.From, m.Content)
	}
}

// CommandParser 命令解析器
type CommandParser struct{}

// NewCommandParser 创建新的命令解析器
func NewCommandParser() *CommandParser {
	return &CommandParser{}
}

// ParseCommand 解析用户输入的命令
func (cp *CommandParser) ParseCommand(input string) (Command, error) {
	input = strings.TrimSpace(input) // 去除空格

	if !strings.HasPrefix(input, "\\") { // 判断是否以\开头
		return Command{Type: CmdChat, Content: input}, nil
	}

	parts := strings.Fields(input) // 分割字符串
	if len(parts) == 0 {
		return Command{}, fmt.Errorf("无效命令")
	}

	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "\\who":
		return Command{Type: CmdWho}, nil
	case "\\rename":
		if len(parts) < 2 {
			return Command{}, fmt.Errorf("重命名命令格式: \\rename <新用户名>")
		}
		return Command{Type: CmdRename, Content: parts[1]}, nil
	case "\\help":
		return Command{Type: CmdHelp}, nil
	case "\\quit", "\\exit":
		return Command{Type: CmdQuit}, nil
	case "\\time":
		return Command{Type: CmdTime}, nil
	case "\\stats":
		return Command{Type: CmdStats}, nil
	case "\\whisper", "\\w":
		if len(parts) < 3 {
			return Command{}, fmt.Errorf("私聊命令格式: \\whisper <用户名> <消息>")
		}
		target := parts[1]
		content := strings.Join(parts[2:], " ")
		return Command{Type: CmdWhisper, Target: target, Content: content}, nil
	default:
		return Command{}, fmt.Errorf("未知命令: %s", cmd)
	}
}

// CommandType 命令类型
type CommandType int

const (
	CmdChat CommandType = iota
	CmdWho
	CmdRename
	CmdHelp
	CmdQuit
	CmdTime
	CmdStats
	CmdWhisper
)

// Command 命令结构体
type Command struct {
	Type    CommandType // 命令类型
	Content string      // 命令内容
	Target  string      // 目标用户
}

// GetHelpMessage 获取帮助信息
func GetHelpMessage() string {
	return `可用命令:
  \\who          - 查看在线用户列表
  \\rename <name> - 重命名
  \\whisper <user> <msg> - 私聊消息
  \\time          - 显示当前时间
  \\stats         - 显示聊天室统计信息
  \\help          - 显示此帮助信息
  \\quit          - 退出聊天室
  \\exit          - 退出聊天室`
}

// GetWelcomeMessage 获取欢迎消息
func GetWelcomeMessage() string {
	return `欢迎来到Go聊天室!
输入 \\help 查看可用命令。
输入消息开始聊天吧！`
}

// FormatUserJoinMessage 格式化用户加入消息
func FormatUserJoinMessage(username string) string {
	return fmt.Sprintf("用户 [%s] 加入了聊天室", username)
}

// FormatUserLeaveMessage 格式化用户离开消息
func FormatUserLeaveMessage(username string) string {
	return fmt.Sprintf("用户 [%s] 离开了聊天室", username)
}

// FormatUserRenameMessage 格式化用户重命名消息
func FormatUserRenameMessage(oldName, newName string) string {
	return fmt.Sprintf("用户 [%s] 将昵称改为 [%s]", oldName, newName)
}

// FormatTimeMessage 格式化时间消息
func FormatTimeMessage() string {
	return fmt.Sprintf("当前时间: %s", time.Now().Format("2006-01-02 15:04:05"))
}

// FormatStatsMessage 格式化统计信息
func FormatStatsMessage(userCount, maxUsers int) string {
	return fmt.Sprintf("聊天室统计: 当前用户 %d/%d", userCount, maxUsers)
}
