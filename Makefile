# Go聊天室项目 Makefile

# 变量定义
BINARY_NAME=chatroom
CLIENT_NAME=client
TEST_NAME=test
BUILD_DIR=build

# 默认目标
.PHONY: all
all: clean build

# 清理构建文件
.PHONY: clean
clean:
	@echo "清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f $(CLIENT_NAME)
	@rm -f $(TEST_NAME)

# 创建构建目录
$(BUILD_DIR):
	@mkdir -p $(BUILD_DIR)

# 构建服务器
.PHONY: build
build: $(BUILD_DIR)
	@echo "构建服务器..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "构建客户端..."
	@go build -o $(BUILD_DIR)/$(CLIENT_NAME) client/client.go
	@echo "构建测试程序..."
	@go build -o $(BUILD_DIR)/$(TEST_NAME) test/test.go
	@echo "构建完成!"

# 运行服务器
.PHONY: run
run: build
	@echo "启动服务器..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

# 运行服务器（后台）
.PHONY: run-daemon
run-daemon: build
	@echo "后台启动服务器..."
	@nohup ./$(BUILD_DIR)/$(BINARY_NAME) > server.log 2>&1 &
	@echo "服务器已在后台启动，日志文件: server.log"

# 停止服务器
.PHONY: stop
stop:
	@echo "停止服务器..."
	@pkill -f $(BINARY_NAME) || true

# 运行测试
.PHONY: test
test: build
	@echo "运行功能测试..."
	@./$(BUILD_DIR)/$(TEST_NAME)

# 运行单元测试
.PHONY: test-unit
test-unit:
	@echo "运行单元测试..."
	@go test ./...

# 格式化代码
.PHONY: fmt
fmt:
	@echo "格式化代码..."
	@go fmt ./...

# 代码检查
.PHONY: lint
lint:
	@echo "代码检查..."
	@golangci-lint run

# 安装依赖
.PHONY: deps
deps:
	@echo "安装依赖..."
	@go mod tidy
	@go mod download

# 显示帮助
.PHONY: help
help:
	@echo "Go聊天室项目 Makefile"
	@echo ""
	@echo "可用命令:"
	@echo "  make all        - 清理并构建所有程序"
	@echo "  make build      - 构建服务器、客户端和测试程序"
	@echo "  make run        - 构建并运行服务器"
	@echo "  make run-daemon - 后台运行服务器"
	@echo "  make stop       - 停止服务器"
	@echo "  make test       - 运行功能测试"
	@echo "  make test-unit  - 运行单元测试"
	@echo "  make fmt        - 格式化代码"
	@echo "  make lint       - 代码检查"
	@echo "  make deps       - 安装依赖"
	@echo "  make clean      - 清理构建文件"
	@echo "  make help       - 显示此帮助信息"
	@echo ""
	@echo "示例:"
	@echo "  make run        # 启动服务器"
	@echo "  make test       # 运行测试"

# 开发模式：自动重启服务器
.PHONY: dev
dev:
	@echo "开发模式：监听文件变化并自动重启服务器..."
	@which air > /dev/null || (echo "请先安装 air: go install github.com/cosmtrek/air@latest" && exit 1)
	@air

# 性能测试
.PHONY: bench
bench:
	@echo "运行性能测试..."
	@go test -bench=. ./...

# 生成文档
.PHONY: docs
docs:
	@echo "生成文档..."
	@go doc -all ./...

# 检查代码覆盖率
.PHONY: coverage
coverage:
	@echo "检查代码覆盖率..."
	@go test -cover ./...
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"
