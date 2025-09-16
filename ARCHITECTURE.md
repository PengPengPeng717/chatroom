# Go聊天室项目架构文档

## 项目概述

这是一个基于Go语言开发的实时聊天室系统，采用TCP协议进行客户端-服务器通信。项目采用模块化设计，包含服务器端、客户端和测试程序。

## 项目结构

```
Chat_room_project/
├── main.go                 # 服务器主程序入口
├── go.mod                  # Go模块定义
├── config/                 # 配置管理模块
│   └── config.go
├── message/                # 消息处理模块
│   └── message.go
├── user/                   # 用户管理模块
│   └── user.go
├── server/                 # 服务器核心模块
│   └── server.go
├── handler/                # 连接处理模块
│   └── handler.go
├── utils/                  # 工具函数模块
│   └── utils.go
├── client/                 # 客户端程序
│   └── client.go
├── test/                   # 测试程序
│   └── test.go
├── source/                 # 原始实现（旧版本）
│   └── chatroom.go
├── build/                  # 构建输出目录
├── Makefile               # 构建脚本
├── Dockerfile             # Docker构建文件
├── docker-compose.yml     # Docker编排文件
└── README.md              # 项目说明
```

## 核心架构设计

### 1. 整体架构图

```
┌─────────────────┐    TCP连接    ┌─────────────────┐
│   客户端程序     │◄─────────────►│   聊天室服务器   │
│   (client)      │               │   (chatroom)    │
└─────────────────┘               └─────────────────┘
                                          │
                                          ▼
                                  ┌─────────────────┐
                                  │   用户管理器     │
                                  │ (UserManager)   │
                                  └─────────────────┘
                                          │
                                          ▼
                                  ┌─────────────────┐
                                  │   消息处理器     │
                                  │(MessageHandler) │
                                  └─────────────────┘
```

### 2. 模块依赖关系

```
main.go
├── config.Config
├── server.ChatServer
│   ├── user.UserManager
│   ├── handler.ConnectionHandler
│   └── utils.Logger
└── message.CommandParser
```

## 详细模块分析

### 1. 配置管理模块 (config)

#### 结构体定义

```go
type Config struct {
    Host       string  // 服务器监听地址
    Port       int     // 服务器监听端口
    MaxUsers   int     // 最大用户数
    Timeout    int     // 用户超时时间(秒)
    BufferSize int     // 缓冲区大小
    LogLevel   string  // 日志级别
    EnableLogs bool    // 是否启用日志
}
```

#### 主要接口

- `DefaultConfig() *Config` - 创建默认配置
- `LoadFromEnv()` - 从环境变量加载配置
- `GetAddress() string` - 获取服务器地址
- `Validate() error` - 验证配置有效性

#### 环境变量支持

- `CHATROOM_HOST` - 服务器监听地址
- `CHATROOM_PORT` - 服务器监听端口
- `CHATROOM_MAX_USERS` - 最大用户数
- `CHATROOM_TIMEOUT` - 用户超时时间
- `CHATROOM_LOG_LEVEL` - 日志级别

### 2. 消息处理模块 (message)

#### 类型定义

```go
// 消息类型枚举
type MessageType int
const (
    TypeChat      MessageType = iota // 聊天消息
    TypeSystem                       // 系统消息
    TypeCommand                      // 命令消息
    TypePrivate                      // 私聊消息
    TypeBroadcast                    // 广播消息
)

// 命令类型枚举
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
```

#### 结构体定义

```go
// 消息结构体
type Message struct {
    Type      MessageType // 消息类型
    From      string      // 发送者
    To        string      // 目标用户
    Content   string      // 消息内容
    Timestamp time.Time   // 时间戳
}

// 命令结构体
type Command struct {
    Type    CommandType // 命令类型
    Content string      // 命令内容
    Target  string      // 目标用户
}

// 命令解析器
type CommandParser struct{}
```

#### 主要接口

**消息创建函数:**
- `NewMessage(msgType MessageType, from, content string) *Message`
- `NewPrivateMessage(from, to, content string) *Message`
- `NewSystemMessage(content string) *Message`

**消息格式化:**
- `FormatMessage() string` - 格式化消息显示

**命令解析:**
- `ParseCommand(input string) (Command, error)` - 解析用户输入命令

**工具函数:**
- `GetHelpMessage() string` - 获取帮助信息
- `GetWelcomeMessage() string` - 获取欢迎消息
- `FormatUserJoinMessage(username string) string` - 格式化用户加入消息
- `FormatUserLeaveMessage(username string) string` - 格式化用户离开消息
- `FormatUserRenameMessage(oldName, newName string) string` - 格式化用户重命名消息
- `FormatTimeMessage() string` - 格式化时间消息
- `FormatStatsMessage(userCount, maxUsers int) string` - 格式化统计信息

#### 支持的命令

| 命令 | 功能 | 格式 |
|------|------|------|
| `\who` | 查看在线用户列表 | `\who` |
| `\rename` | 重命名 | `\rename <新用户名>` |
| `\whisper` | 私聊消息 | `\whisper <用户名> <消息>` |
| `\time` | 显示当前时间 | `\time` |
| `\stats` | 显示统计信息 | `\stats` |
| `\help` | 显示帮助信息 | `\help` |
| `\quit` | 退出聊天室 | `\quit` |
| `\exit` | 退出聊天室 | `\exit` |

### 3. 用户管理模块 (user)

#### 结构体定义

```go
// 用户结构体
type User struct {
    ID       string      // 用户ID
    Name     string      // 用户名
    MsgChan  chan string // 消息通道
    DoneChan chan bool   // 退出信号
    JoinTime time.Time   // 加入时间
    LastSeen time.Time   // 最后活跃时间
    IsActive bool        // 是否活跃
}

// 用户管理器
type UserManager struct {
    users    map[string]*User // 用户列表
    mutex    sync.RWMutex     // 互斥锁
    maxUsers int              // 最大用户数
}
```

#### 主要接口

**用户管理:**
- `CreateUser(id, name string) (*User, error)` - 创建新用户
- `GetUser(id string) (*User, bool)` - 获取用户
- `RemoveUser(id string) (*User, bool)` - 移除用户
- `GetAllUsers() []*User` - 获取所有用户
- `GetUserCount() int` - 获取用户数量

**用户操作:**
- `UpdateUserLastSeen(id string)` - 更新用户最后活跃时间
- `RenameUser(id, newName string) error` - 重命名用户
- `GetUserList() string` - 获取用户列表字符串

**消息广播:**
- `BroadcastToAll(message string)` - 向所有用户广播消息
- `BroadcastToOthers(excludeID, message string)` - 向除指定用户外的所有用户广播消息
- `SendToUser(userID, message string) error` - 向指定用户发送消息

#### 线程安全

用户管理器使用 `sync.RWMutex` 确保并发安全：
- 读操作使用 `RLock()`
- 写操作使用 `Lock()`
- 支持多个并发读操作

### 4. 服务器核心模块 (server)

#### 结构体定义

```go
// 聊天服务器
type ChatServer struct {
    config            *config.Config             // 配置
    userManager       *user.UserManager          // 用户管理器
    connectionHandler *handler.ConnectionHandler // 连接处理器
    logger            *utils.Logger              // 日志记录器
    listener          net.Listener               // 监听器
    isRunning         bool                       // 是否运行
}
```

#### 主要接口

- `NewChatServer(cfg *config.Config) *ChatServer` - 创建新的聊天服务器
- `Start() error` - 启动服务器
- `Stop()` - 停止服务器
- `GetStats() map[string]interface{}` - 获取服务器统计信息
- `BroadcastMessage(message string)` - 广播消息

#### 服务器生命周期

1. **初始化阶段:**
   - 验证配置
   - 创建监听器
   - 初始化各个组件

2. **运行阶段:**
   - 接受客户端连接
   - 为每个连接启动独立的goroutine
   - 监控系统信号

3. **关闭阶段:**
   - 停止接受新连接
   - 断开所有现有连接
   - 清理资源

### 5. 连接处理模块 (handler)

#### 结构体定义

```go
// 连接处理器
type ConnectionHandler struct {
    userManager   *user.UserManager      // 用户管理器
    commandParser *message.CommandParser // 命令解析器
    logger        *utils.Logger          // 日志记录器
    config        *config.Config         // 配置
}
```

#### 主要接口

- `NewConnectionHandler(userManager *user.UserManager, logger *utils.Logger, cfg *config.Config) *ConnectionHandler`
- `HandleConnection(conn net.Conn)` - 处理客户端连接
- `CleanupUser(currentUser *user.User)` - 清理用户资源

#### 连接处理流程

1. **连接建立:**
   - 生成用户ID和默认用户名
   - 创建用户对象
   - 发送欢迎消息
   - 广播用户加入消息

2. **消息处理:**
   - 启动消息写入goroutine
   - 启动超时监控goroutine
   - 处理客户端消息输入

3. **连接清理:**
   - 移除用户
   - 广播用户离开消息
   - 关闭连接和通道

#### 超时机制

- 使用 `time.Ticker` 定期检查用户活跃状态
- 超时时间由配置决定
- 超时后自动断开连接

### 6. 工具函数模块 (utils)

#### 结构体定义

```go
// 日志记录器
type Logger struct {
    EnableLogs bool
}
```

#### 主要接口

**日志功能:**
- `NewLogger(enableLogs bool) *Logger`
- `Info(format string, args ...interface{})`
- `Error(format string, args ...interface{})`
- `Debug(format string, args ...interface{})`
- `Warn(format string, args ...interface{})`

**用户相关:**
- `GenerateUserID(conn net.Conn) string` - 生成用户ID
- `GenerateUsername(conn net.Conn) string` - 生成默认用户名
- `ValidateUsername(username string) error` - 验证用户名

**网络相关:**
- `IsValidIP(ip string) bool` - 检查IP地址有效性
- `GetLocalIP() string` - 获取本地IP地址

**字符串处理:**
- `TruncateString(s string, maxLen int) string` - 截断字符串
- `SanitizeInput(input string) string` - 清理用户输入
- `FormatDuration(d time.Duration) string` - 格式化持续时间

**重试机制:**
- `RetryWithBackoff(maxRetries int, baseDelay time.Duration, fn func() error) error`

### 7. 客户端程序 (client)

#### 主要功能

- 连接到聊天室服务器
- 发送用户输入消息
- 接收并显示服务器消息
- 处理退出信号

#### 程序流程

1. 解析命令行参数（服务器地址和端口）
2. 建立TCP连接
3. 启动消息接收goroutine
4. 处理用户输入循环
5. 处理退出信号

### 8. 测试程序 (test)

#### 结构体定义

```go
// 测试客户端
type TestClient struct {
    conn   net.Conn
    reader *bufio.Reader
}
```

#### 测试功能

- 创建多个测试客户端
- 测试消息发送和接收
- 测试用户列表查询
- 测试用户重命名
- 测试时间查询
- 测试统计信息
- 测试私聊功能

## 数据流图

### 1. 用户连接流程

```
客户端 → TCP连接 → 服务器接受连接 → 创建用户 → 发送欢迎消息 → 广播用户加入
```

### 2. 消息处理流程

```
用户输入 → 命令解析 → 业务处理 → 消息格式化 → 广播/发送 → 客户端显示
```

### 3. 用户退出流程

```
退出信号 → 清理用户资源 → 广播用户离开 → 关闭连接 → 清理内存
```

## 并发模型

### 1. Goroutine使用

- **主服务器goroutine:** 接受新连接
- **连接处理goroutine:** 每个客户端连接一个
- **消息写入goroutine:** 每个用户一个，负责向客户端发送消息
- **超时监控goroutine:** 每个用户一个，监控用户活跃状态
- **信号处理goroutine:** 处理系统信号

### 2. 通道通信

- **用户消息通道:** `chan string` - 向用户发送消息
- **退出信号通道:** `chan bool` - 通知用户退出
- **全局消息通道:** 用于广播消息（在旧版本中使用）

### 3. 同步机制

- **读写锁:** 保护用户管理器的并发访问
- **通道同步:** 使用通道进行goroutine间通信
- **超时控制:** 使用 `time.Ticker` 和 `time.After` 实现超时

## 错误处理

### 1. 连接错误

- 网络连接失败
- 客户端异常断开
- 服务器资源不足

### 2. 用户错误

- 用户名重复
- 用户数量超限
- 无效命令输入

### 3. 系统错误

- 配置验证失败
- 文件操作失败
- 内存不足

## 性能考虑

### 1. 内存管理

- 及时清理断开的用户资源
- 使用缓冲通道避免阻塞
- 限制消息通道大小

### 2. 并发优化

- 使用读写锁提高读操作性能
- 每个连接独立的goroutine
- 异步消息处理

### 3. 网络优化

- TCP连接复用
- 设置合理的超时时间
- 缓冲区大小优化

## 扩展性设计

### 1. 模块化架构

- 清晰的模块边界
- 接口抽象
- 依赖注入

### 2. 配置灵活性

- 环境变量支持
- 命令行参数
- 配置文件支持（可扩展）

### 3. 功能扩展点

- 新的消息类型
- 新的命令支持
- 新的用户管理功能

## 部署和运维

### 1. 构建系统

- Makefile自动化构建
- Docker容器化支持
- 多阶段构建优化

### 2. 监控和日志

- 结构化日志输出
- 服务器统计信息
- 健康检查支持

### 3. 配置管理

- 环境变量配置
- 默认配置
- 配置验证

## 安全考虑

### 1. 输入验证

- 用户名格式验证
- 消息内容清理
- 命令参数验证

### 2. 资源限制

- 最大用户数限制
- 消息通道大小限制
- 超时机制

### 3. 错误处理

- 不暴露敏感信息
- 优雅的错误处理
- 资源清理

## 测试策略

### 1. 单元测试

- 各模块独立测试
- 边界条件测试
- 错误情况测试

### 2. 集成测试

- 端到端功能测试
- 多客户端并发测试
- 性能压力测试

### 3. 自动化测试

- 持续集成支持
- 自动化测试脚本
- 测试覆盖率检查

## 总结

这个Go聊天室项目采用了现代化的软件架构设计，具有以下特点：

1. **模块化设计:** 清晰的模块划分和职责分离
2. **并发安全:** 使用Go的并发原语确保线程安全
3. **可扩展性:** 良好的接口设计和扩展点
4. **可维护性:** 清晰的代码结构和文档
5. **可部署性:** 支持容器化部署和自动化构建

项目展示了Go语言在网络编程、并发处理和系统设计方面的优势，是一个很好的学习和实践项目。
