package user

import (
	"fmt"
	"sync"
	"time"
)

// User 用户结构体
type User struct {
	ID       string      // 用户ID
	Name     string      // 用户名
	MsgChan  chan string // 消息通道
	DoneChan chan bool   // 退出信号
	JoinTime time.Time   // 加入时间
	LastSeen time.Time   // 最后活跃时间
	IsActive bool        // 是否活跃
}

// UserManager 用户管理器
type UserManager struct {
	users    map[string]*User // 用户列表
	mutex    sync.RWMutex     // 互斥锁
	maxUsers int              // 最大用户数
}

// NewUserManager 创建新的用户管理器
func NewUserManager(maxUsers int) *UserManager {
	return &UserManager{
		users:    make(map[string]*User),
		maxUsers: maxUsers,
	}
}

// CreateUser 创建新用户
func (um *UserManager) CreateUser(id, name string) (*User, error) {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	// 检查用户数量限制
	if len(um.users) >= um.maxUsers {
		return nil, fmt.Errorf("聊天室已满，无法加入新用户")
	}

	// 检查用户ID是否已存在
	if _, exists := um.users[id]; exists {
		return nil, fmt.Errorf("用户ID已存在")
	}

	user := &User{
		ID:       id,
		Name:     name,
		MsgChan:  make(chan string, 100),
		DoneChan: make(chan bool, 1),
		JoinTime: time.Now(),
		LastSeen: time.Now(),
		IsActive: true,
	}

	um.users[id] = user
	return user, nil
}

// GetUser 获取用户
func (um *UserManager) GetUser(id string) (*User, bool) {
	um.mutex.RLock()
	defer um.mutex.RUnlock()
	user, exists := um.users[id]
	return user, exists
}

// RemoveUser 移除用户
func (um *UserManager) RemoveUser(id string) (*User, bool) {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	user, exists := um.users[id]
	if exists {
		delete(um.users, id)
		user.IsActive = false
		close(user.MsgChan)
		close(user.DoneChan)
	}
	return user, exists
}

// GetAllUsers 获取所有用户
func (um *UserManager) GetAllUsers() []*User {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	users := make([]*User, 0, len(um.users))
	for _, user := range um.users {
		users = append(users, user)
	}
	return users
}

// GetUserCount 获取用户数量
func (um *UserManager) GetUserCount() int {
	um.mutex.RLock()
	defer um.mutex.RUnlock()
	return len(um.users)
}

// UpdateUserLastSeen 更新用户最后活跃时间
func (um *UserManager) UpdateUserLastSeen(id string) {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	if user, exists := um.users[id]; exists {
		user.LastSeen = time.Now()
	}
}

// RenameUser 重命名用户
func (um *UserManager) RenameUser(id, newName string) error {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	user, exists := um.users[id]
	if !exists {
		return fmt.Errorf("用户不存在")
	}

	// 检查新用户名是否已被使用
	for _, u := range um.users {
		if u.ID != id && u.Name == newName {
			return fmt.Errorf("用户名已被使用")
		}
	}

	user.Name = newName
	return nil
}

// GetUserList 获取用户列表字符串
func (um *UserManager) GetUserList() string {
	users := um.GetAllUsers()
	if len(users) == 0 {
		return "当前没有在线用户\n"
	}

	result := fmt.Sprintf("当前在线用户 (%d人):\n", len(users))
	for _, user := range users {
		onlineTime := time.Since(user.JoinTime).Round(time.Second)
		result += fmt.Sprintf("- %s (ID: %s, 在线时长: %s)\n",
			user.Name, user.ID, onlineTime)
	}
	return result
}

// BroadcastToAll 向所有用户广播消息
func (um *UserManager) BroadcastToAll(message string) {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	for _, user := range um.users {
		select {
		case user.MsgChan <- message:
		default:
			// 如果用户的消息通道已满，跳过该用户
		}
	}
}

// BroadcastToOthers 向除指定用户外的所有用户广播消息
func (um *UserManager) BroadcastToOthers(excludeID, message string) {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	for _, user := range um.users {
		if user.ID != excludeID {
			select {
			case user.MsgChan <- message:
			default:
				// 如果用户的消息通道已满，跳过该用户
			}
		}
	}
}

// SendToUser 向指定用户发送消息
func (um *UserManager) SendToUser(userID, message string) error {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	user, exists := um.users[userID]
	if !exists {
		return fmt.Errorf("用户不存在")
	}

	select {
	case user.MsgChan <- message:
		return nil
	default:
		return fmt.Errorf("用户消息通道已满")
	}
}
