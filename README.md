# AI Model Transfer Hub

AI模型中转服务，支持OpenAI、Anthropic、Google等多个AI提供商的统一接口访问。

## 功能特性

- ✅ 统一的OpenAI格式API接口
- ✅ 支持多个AI提供商（OpenAI、Anthropic、Google）
- ✅ Token认证和权限管理
- ✅ 请求限流控制
- ✅ 完整的链路追踪
- ✅ 日志按天分割、按级别分离、自动滚动
- ✅ 流式响应支持（SSE）
- ✅ 使用量统计和记录

## 快速开始

### 1. 配置文件

复制配置示例文件：

```bash
cp config.example.yaml config.yaml
```

编辑`config.yaml`，配置数据库和AI提供商信息。

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 运行服务

```bash
go run cmd/server/main.go
```

服务将在配置的端口启动（默认8080）。

## API接口

### 健康检查

```bash
GET /health
```

### 聊天完成（OpenAI格式）

```bash
POST /v1/chat/completions
Authorization: Bearer {your-token}
Content-Type: application/json

{
  "model": "gpt-4",
  "messages": [
    {"role": "user", "content": "Hello"}
  ],
  "stream": false
}
```

### 流式响应

```bash
POST /v1/chat/completions
Authorization: Bearer {your-token}
Content-Type: application/json

{
  "model": "gpt-4",
  "messages": [
    {"role": "user", "content": "Hello"}
  ],
  "stream": true
}
```

## 配置说明

### 系统配置

- `system.addr`: 监听端口
- `system.mode`: 运行模式（debug/release）

### 数据库配置

- `database.type`: 数据库类型（目前仅支持mysql）
- `database.host`: 数据库地址
- `database.port`: 数据库端口
- `database.db-name`: 数据库名称
- `database.username`: 用户名
- `database.password`: 密码

### 日志配置

- `zap.level`: 日志级别（debug/info/warn/error）
- `zap.director`: 日志目录
- `zap.retention-day`: 日志保留天数
- `zap.max-size`: 单文件最大大小（MB）

### AI提供商配置

```yaml
providers:
  - name: openai
    enabled: true
    base_url: https://api.openai.com/v1
    api_key: your-api-key
    timeout: 300
    models:
      - gpt-4
      - gpt-3.5-turbo
```

### 限流配置

- `rate_limit.enabled`: 是否启用限流
- `rate_limit.default_rpm`: 默认每分钟请求数
- `rate_limit.default_rph`: 默认每小时请求数

## 项目结构

```
model-transfer-hub/
├── cmd/server/          # 启动入口
├── config/              # 配置定义
├── core/                # 核心组件
├── global/              # 全局变量
├── initialize/          # 初始化逻辑
├── middleware/          # 中间件
├── handler/             # 请求处理
├── service/             # 业务服务
├── model/               # 数据模型
├── pkg/                 # 工具包
└── logs/                # 日志目录
```

## 使用方式
### 接入端点：
- https://ai.cngamesdk.com
### 令牌
- 加微信：doudoualvin 申请
### claude code 配置
```shell
export ANTHROPIC_AUTH_TOKEN="sk-xxx" # token
export ANTHROPIC_BASE_URL="https://ai.cngamesdk.com" # endpoint
```

## 开发计划

- [x] 项目初始化
- [x] 配置管理
- [x] 日志系统
- [x] 数据库模型
- [x] 中间件系统
- [x] HTTP服务器
- [x] 适配器实现
- [x] 代理服务
- [x] 后台管理模块

## License

MIT
