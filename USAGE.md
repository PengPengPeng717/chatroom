# Go聊天室使用指南

## 快速开始

### 1. 启动服务器

```bash
# 使用默认配置启动
./chatroom

# 自定义配置启动
./chatroom -host 0.0.0.0 -port 9000 -max-users 50 -timeout 60

# 查看帮助
./chatroom -help
```

### 2. 连接客户端

#### 方法1：使用内置客户端
```bash
./client 127.0.0.1 8080
```

#### 方法2：使用telnet
```bash
telnet 127.0.0.1 8080
```

#### 方法3：使用netcat
```bash
nc 127.0.0.1 8080
```

### 1.首先启动服务器
```bash

# 编译并启动服务器
go build -o chatroom.exe main.go
.\chatroom.exe
```
您应该看到类似这样的输出：

```text
=== Go聊天室服务器 ===
监听地址: 127.0.0.1:8080
最大用户数: 100
超时时间: 40秒
按 Ctrl+C 停止服务器
=====================
[INFO] 2024-XX-XX XX:XX:XX 聊天服务器启动成功，监听地址: 127.0.0.1:8080
[INFO] 2024-XX-XX XX:XX:XX 最大用户数: 100, 超时时间: 40秒
```


### 在第二个PowerShell窗口中执行
```bash
.\client.exe 127.0.0.1 8080
```
您应该看到类似这样的输出：

```text
=== Go聊天室服务器 ===
监听地址: 127.0.0.1:8080
最大用户数: 100
超时时间: 40秒
按 Ctrl+C 停止服务器
=====================
[INFO] 2024-XX-XX XX:XX:XX 聊天服务器启动成功，监听地址: 127.0.0.1:8080
[INFO] 2024-XX-XX XX:XX:XX 最大用户数: 100, 超时时间: 40秒
```


## 用户命令示例

### 基础聊天
```
Hello, everyone!          # 发送消息给所有人
```

### 查看在线用户
```
\who                      # 显示当前在线用户列表
```

### 重命名
```
\rename 新用户名          # 更改你的用户名
```

### 私聊消息
```
\whisper 用户名 消息内容   # 发送私聊消息
\w 用户名 消息内容        # 简写形式
```

### 系统信息
```
\time                     # 显示当前时间
\stats                    # 显示聊天室统计信息
\help                     # 显示帮助信息
```

### 退出聊天室
```
\quit                     # 退出聊天室
\exit                     # 退出聊天室
```

## 实际使用示例

### 场景1：多人聊天

1. **启动服务器**
   ```bash
   ./chatroom -host 0.0.0.0 -port 8080
   ```

2. **用户A连接**
   ```bash
   ./client 127.0.0.1 8080
   # 输入: \rename Alice
   # 输入: Hello, I'm Alice!
   ```

3. **用户B连接**
   ```bash
   ./client 127.0.0.1 8080
   # 输入: \rename Bob
   # 输入: Hi Alice, I'm Bob!
   ```

4. **用户C连接**
   ```bash
   ./client 127.0.0.1 8080
   # 输入: \rename Charlie
   # 输入: Hello everyone!
   ```

### 场景2：私聊对话

```
# 用户A发送私聊消息给用户B
\whisper Bob 你好，我想和你私聊

# 用户B回复私聊消息给用户A
\whisper Alice 好的，我们私聊吧
```

### 场景3：查看聊天室状态

```
# 查看在线用户
\who

# 查看统计信息
\stats

# 查看当前时间
\time
```

## 高级配置

### 环境变量配置

```bash
# 设置环境变量
export CHATROOM_HOST=0.0.0.0
export CHATROOM_PORT=9000
export CHATROOM_MAX_USERS=200
export CHATROOM_TIMEOUT=60

# 启动服务器
./chatroom
```

### Docker部署

```bash
# 构建并启动
docker-compose up -d

# 查看日志
docker-compose logs -f chatroom

# 停止服务
docker-compose down
```

### 生产环境部署

```bash
# 后台运行
nohup ./chatroom -host 0.0.0.0 -port 8080 > server.log 2>&1 &

# 查看进程
ps aux | grep chatroom

# 停止服务
pkill -f chatroom
```

## 故障排除

### 常见问题

1. **连接被拒绝**
   - 检查服务器是否已启动
   - 检查端口是否正确
   - 检查防火墙设置

2. **用户名已存在**
   - 使用 `\rename` 命令更改用户名
   - 选择其他用户名

3. **聊天室已满**
   - 等待其他用户退出
   - 联系管理员增加用户限制

4. **消息发送失败**
   - 检查网络连接
   - 重新连接服务器

### 日志查看

```bash
# 查看服务器日志
tail -f server.log

# 查看Docker容器日志
docker-compose logs -f chatroom
```

## 性能优化

### 服务器配置建议

- **小规模使用** (< 50用户)
  ```bash
  ./chatroom -max-users 50 -timeout 30
  ```

- **中等规模** (50-200用户)
  ```bash
  ./chatroom -max-users 200 -timeout 40
  ```

- **大规模使用** (> 200用户)
  ```bash
  ./chatroom -max-users 500 -timeout 60
  ```

### 系统资源监控

```bash
# 监控CPU和内存使用
top -p $(pgrep chatroom)

# 监控网络连接
netstat -an | grep 8080

# 监控文件描述符
lsof -p $(pgrep chatroom)
```

## 安全建议

1. **网络安全**
   - 使用防火墙限制访问
   - 考虑使用SSL/TLS加密
   - 定期更新系统

2. **用户管理**
   - 设置合理的用户数量限制
   - 监控异常用户行为
   - 定期清理不活跃用户

3. **日志管理**
   - 定期轮转日志文件
   - 监控错误日志
   - 备份重要数据

## 扩展功能

### 自定义命令

可以通过修改 `message/message.go` 文件添加新的命令：

```go
case "\\custom":
    return Command{Type: CmdCustom}, nil
```

### 消息持久化

可以集成Redis或数据库来持久化消息：

```go
// 在handler中添加消息存储逻辑
messageStore.Save(msg)
```

### Web界面

可以添加HTTP API和WebSocket支持来提供Web界面。

## 技术支持

如果遇到问题，请：

1. 查看日志文件
2. 检查配置参数
3. 参考README.md文档
4. 提交Issue到项目仓库
