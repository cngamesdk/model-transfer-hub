# AI Model Transfer Hub - Makefile

# 项目信息
PROJECT_NAME := model-transfer-hub
BINARY_NAME := server
VERSION := 1.1.1
BUILD_TIME := $(shell date +%Y%m%d%H%M%S)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 路径配置
CMD_PATH := cmd/server
BIN_DIR := bin
LOG_DIR := logs
CONFIG_FILE := config.yaml

# Go 配置
GO := go
GOFLAGS := -v
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# 颜色输出
COLOR_RESET := \033[0m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_BLUE := \033[34m
COLOR_RED := \033[31m

.PHONY: all build clean run test help install deps tidy fmt vet lint docker db-init db-migrate dev prod

# 默认目标
all: clean fmt vet build

## help: 显示帮助信息
help:
	@echo ""
	@echo "$(COLOR_BLUE)AI Model Transfer Hub - Makefile 帮助$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_GREEN)可用命令:$(COLOR_RESET)"
	@echo ""
	@echo "  $(COLOR_YELLOW)make build$(COLOR_RESET)        - 编译项目"
	@echo "  $(COLOR_YELLOW)make run$(COLOR_RESET)          - 运行服务（默认配置）"
	@echo "  $(COLOR_YELLOW)make dev$(COLOR_RESET)          - 运行开发环境"
	@echo "  $(COLOR_YELLOW)make prod$(COLOR_RESET)         - 运行生产环境"
	@echo "  $(COLOR_YELLOW)make test$(COLOR_RESET)         - 运行测试"
	@echo "  $(COLOR_YELLOW)make clean$(COLOR_RESET)        - 清理编译产物"
	@echo "  $(COLOR_YELLOW)make deps$(COLOR_RESET)         - 下载依赖"
	@echo "  $(COLOR_YELLOW)make tidy$(COLOR_RESET)         - 整理依赖"
	@echo "  $(COLOR_YELLOW)make fmt$(COLOR_RESET)          - 格式化代码"
	@echo "  $(COLOR_YELLOW)make vet$(COLOR_RESET)          - 代码检查"
	@echo "  $(COLOR_YELLOW)make lint$(COLOR_RESET)         - 代码规范检查"
	@echo "  $(COLOR_YELLOW)make install$(COLOR_RESET)      - 安装到系统"
	@echo "  $(COLOR_YELLOW)make docker$(COLOR_RESET)       - 构建Docker镜像"
	@echo "  $(COLOR_YELLOW)make db-init$(COLOR_RESET)      - 初始化数据库"
	@echo "  $(COLOR_YELLOW)make db-migrate$(COLOR_RESET)   - 数据库迁移"
	@echo ""
	@echo "$(COLOR_GREEN)示例:$(COLOR_RESET)"
	@echo "  make build              # 编译项目"
	@echo "  make run CONFIG=dev     # 使用开发配置运行"
	@echo "  make test               # 运行测试"
	@echo ""

## build: 编译项目
build:
	@echo "$(COLOR_BLUE)>>> 编译项目...$(COLOR_RESET)"
	@mkdir -p $(BIN_DIR)
	@$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_PATH)/main.go
	@echo "$(COLOR_GREEN)✓ 编译完成: $(BIN_DIR)/$(BINARY_NAME)$(COLOR_RESET)"
	@ls -lh $(BIN_DIR)/$(BINARY_NAME)

## build-all: 编译所有平台
build-all:
	@echo "$(COLOR_BLUE)>>> 编译多平台版本...$(COLOR_RESET)"
	@mkdir -p $(BIN_DIR)
	@echo "编译 Linux (amd64)..."
	@GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_PATH)/main.go
	@echo "编译 Linux (arm64)..."
	@GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_PATH)/main.go
	@echo "编译 macOS (amd64)..."
	@GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_PATH)/main.go
	@echo "编译 macOS (arm64)..."
	@GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_PATH)/main.go
	@echo "$(COLOR_GREEN)✓ 多平台编译完成$(COLOR_RESET)"
	@ls -lh $(BIN_DIR)/

## clean: 清理编译产物
clean:
	@echo "$(COLOR_BLUE)>>> 清理编译产物...$(COLOR_RESET)"
	@rm -rf $(BIN_DIR)
	@echo "$(COLOR_GREEN)✓ 清理完成$(COLOR_RESET)"

## run: 运行服务（默认配置）
run: build
	@echo "$(COLOR_BLUE)>>> 启动服务...$(COLOR_RESET)"
	@./$(BIN_DIR)/$(BINARY_NAME) -c $(CONFIG_FILE)

## dev: 运行开发环境
dev: build
	@echo "$(COLOR_BLUE)>>> 启动开发环境...$(COLOR_RESET)"
	@./$(BIN_DIR)/$(BINARY_NAME) -c config.dev.yaml

## prod: 运行生产环境
prod: build
	@echo "$(COLOR_BLUE)>>> 启动生产环境...$(COLOR_RESET)"
	@./$(BIN_DIR)/$(BINARY_NAME) -c config.prod.yaml

## test: 运行测试
test:
	@echo "$(COLOR_BLUE)>>> 运行测试...$(COLOR_RESET)"
	@$(GO) test -v -race -coverprofile=coverage.out ./...
	@echo "$(COLOR_GREEN)✓ 测试完成$(COLOR_RESET)"

## test-coverage: 运行测试并查看覆盖率
test-coverage: test
	@echo "$(COLOR_BLUE)>>> 生成覆盖率报告...$(COLOR_RESET)"
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)✓ 覆盖率报告: coverage.html$(COLOR_RESET)"

## bench: 运行性能测试
bench:
	@echo "$(COLOR_BLUE)>>> 运行性能测试...$(COLOR_RESET)"
	@$(GO) test -bench=. -benchmem ./...

## deps: 下载依赖
deps:
	@echo "$(COLOR_BLUE)>>> 下载依赖...$(COLOR_RESET)"
	@$(GO) mod download
	@echo "$(COLOR_GREEN)✓ 依赖下载完成$(COLOR_RESET)"

## tidy: 整理依赖
tidy:
	@echo "$(COLOR_BLUE)>>> 整理依赖...$(COLOR_RESET)"
	@$(GO) mod tidy
	@echo "$(COLOR_GREEN)✓ 依赖整理完成$(COLOR_RESET)"

## fmt: 格式化代码
fmt:
	@echo "$(COLOR_BLUE)>>> 格式化代码...$(COLOR_RESET)"
	@$(GO) fmt ./...
	@echo "$(COLOR_GREEN)✓ 代码格式化完成$(COLOR_RESET)"

## vet: 代码检查
vet:
	@echo "$(COLOR_BLUE)>>> 代码检查...$(COLOR_RESET)"
	@$(GO) vet ./...
	@echo "$(COLOR_GREEN)✓ 代码检查完成$(COLOR_RESET)"

## lint: 代码规范检查（需要golangci-lint）
lint:
	@echo "$(COLOR_BLUE)>>> 代码规范检查...$(COLOR_RESET)"
	@which golangci-lint > /dev/null || (echo "$(COLOR_RED)错误: 请先安装 golangci-lint$(COLOR_RESET)" && exit 1)
	@golangci-lint run ./...
	@echo "$(COLOR_GREEN)✓ 代码规范检查完成$(COLOR_RESET)"

## install: 安装到系统
install: build
	@echo "$(COLOR_BLUE)>>> 安装到系统...$(COLOR_RESET)"
	@install -m 755 $(BIN_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "$(COLOR_GREEN)✓ 安装完成: /usr/local/bin/$(BINARY_NAME)$(COLOR_RESET)"

## uninstall: 从系统卸载
uninstall:
	@echo "$(COLOR_BLUE)>>> 从系统卸载...$(COLOR_RESET)"
	@rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "$(COLOR_GREEN)✓ 卸载完成$(COLOR_RESET)"

## docker: 构建Docker镜像
docker:
	@echo "$(COLOR_BLUE)>>> 构建Docker镜像...$(COLOR_RESET)"
	@docker build -t $(PROJECT_NAME):$(VERSION) .
	@docker tag $(PROJECT_NAME):$(VERSION) $(PROJECT_NAME):latest
	@echo "$(COLOR_GREEN)✓ Docker镜像构建完成$(COLOR_RESET)"

## docker-run: 运行Docker容器
docker-run:
	@echo "$(COLOR_BLUE)>>> 运行Docker容器...$(COLOR_RESET)"
	@docker run -d \
		--name $(PROJECT_NAME) \
		-p 8080:8080 \
		-v $(PWD)/config.yaml:/app/config.yaml \
		-v $(PWD)/logs:/app/logs \
		$(PROJECT_NAME):latest

## docker-stop: 停止Docker容器
docker-stop:
	@echo "$(COLOR_BLUE)>>> 停止Docker容器...$(COLOR_RESET)"
	@docker stop $(PROJECT_NAME)
	@docker rm $(PROJECT_NAME)

## db-init: 初始化数据库
db-init:
	@echo "$(COLOR_BLUE)>>> 初始化数据库...$(COLOR_RESET)"
	@mysql -u jishu -p'7nDU0Mn#Osq3ka1J!' -h 192.168.60.219 < init.sql
	@echo "$(COLOR_GREEN)✓ 数据库初始化完成$(COLOR_RESET)"

## db-migrate: 数据库迁移（自动迁移）
db-migrate:
	@echo "$(COLOR_BLUE)>>> 数据库迁移...$(COLOR_RESET)"
	@echo "服务启动时会自动迁移，无需手动执行"

## logs: 查看日志
logs:
	@tail -f $(LOG_DIR)/$$(date +%Y-%m-%d)/info.log

## logs-error: 查看错误日志
logs-error:
	@tail -f $(LOG_DIR)/$$(date +%Y-%m-%d)/error.log

## version: 显示版本信息
version:
	@echo "$(COLOR_BLUE)AI Model Transfer Hub$(COLOR_RESET)"
	@echo "版本: $(VERSION)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "Git提交: $(GIT_COMMIT)"

## release: 创建发布版本
release: clean test build-all
	@echo "$(COLOR_BLUE)>>> 创建发布版本...$(COLOR_RESET)"
	@mkdir -p release
	@tar -czf release/$(PROJECT_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(BIN_DIR) $(BINARY_NAME)-linux-amd64
	@tar -czf release/$(PROJECT_NAME)-$(VERSION)-linux-arm64.tar.gz -C $(BIN_DIR) $(BINARY_NAME)-linux-arm64
	@tar -czf release/$(PROJECT_NAME)-$(VERSION)-darwin-amd64.tar.gz -C $(BIN_DIR) $(BINARY_NAME)-darwin-amd64
	@tar -czf release/$(PROJECT_NAME)-$(VERSION)-darwin-arm64.tar.gz -C $(BIN_DIR) $(BINARY_NAME)-darwin-arm64
	@echo "$(COLOR_GREEN)✓ 发布版本创建完成: release/$(COLOR_RESET)"
	@ls -lh release/

## check: 运行所有检查
check: fmt vet test
	@echo "$(COLOR_GREEN)✓ 所有检查通过$(COLOR_RESET)"

## watch: 监听文件变化并自动重新编译运行（需要安装 air）
watch:
	@which air > /dev/null || (echo "$(COLOR_RED)错误: 请先安装 air (go install github.com/cosmtrek/air@latest)$(COLOR_RESET)" && exit 1)
	@air

## info: 显示项目信息
info:
	@echo "$(COLOR_BLUE)项目信息$(COLOR_RESET)"
	@echo "项目名称: $(PROJECT_NAME)"
	@echo "版本: $(VERSION)"
	@echo "Go版本: $(shell $(GO) version)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "Git提交: $(GIT_COMMIT)"
	@echo ""
	@echo "$(COLOR_BLUE)目录信息$(COLOR_RESET)"
	@echo "二进制目录: $(BIN_DIR)"
	@echo "日志目录: $(LOG_DIR)"
	@echo "配置文件: $(CONFIG_FILE)"
