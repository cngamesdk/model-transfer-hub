# AI Model Transfer Hub

AI 模型统一代理网关，提供 OpenAI 兼容的 API 接口，同时支持 Anthropic 原生 Messages API。后端可接入 OpenAI、Anthropic、Google 等多个 AI 提供商。

## 功能特性

- OpenAI 格式 `/v1/chat/completions` 接口（自动转换 Anthropic 响应格式）
- Anthropic 原生 `/v1/messages` 接口（纯透传，支持 tool use / thinking 等全部特性）
- 流式响应（SSE）— 双格式支持：OpenAI SSE 和 Anthropic SSE
- Token 认证与权限管理
- 请求限流（RPM / RPH）
- 请求链路追踪（X-Trace-Id）
- Token 用量统计（input_tokens / output_tokens 分别记录）
- 日志按天分割、自动滚动
- 开发模式（`dev_mode`）旁路认证

## 快速开始

### 1. 前置条件

- Go 1.25+
- MySQL 5.6+

### 2. 配置文件

```bash
cp config.example.yaml config.dev.yaml
```

编辑 `config.dev.yaml`，填写数据库连接和 AI 提供商 API Key：

```yaml
database:
  host: 127.0.0.1
  port: 3306
  db-name: model_transfer
  username: root
  password: your-password

providers:
  - name: openai
    enabled: true
    base_url: https://api.openai.com/v1
    api_key: sk-your-key
    models: [gpt-4, gpt-3.5-turbo]

  - name: anthropic
    enabled: true
    base_url: https://api.anthropic.com/v1
    api_key: sk-ant-your-key
    models: [claude-opus-4-6, claude-3-sonnet]
```

### 3. 启动服务

```bash
go run cmd/server/main.go -c config.dev.yaml
```

服务默认监听 `:8080`，启动后自动迁移数据库表。

## API 接口

### 健康检查

```
GET /health
```

响应：
```json
{"status": "ok", "message": "服务运行正常"}
```

### OpenAI 格式 — 聊天完成

```
POST /v1/chat/completions
Authorization: Bearer {token}
Content-Type: application/json
```

**非流式请求：**
```json
{
  "model": "claude-opus-4-6",
  "messages": [{"role": "user", "content": "Hello"}]
}
```

响应为标准 OpenAI `chat.completion` 格式：
```json
{
  "id": "msg_xxx",
  "object": "chat.completion",
  "model": "claude-opus-4-6",
  "choices": [{
    "index": 0,
    "message": {"role": "assistant", "content": "Hi!"},
    "finish_reason": "end_turn"
  }],
  "usage": {"prompt_tokens": 8, "completion_tokens": 2, "total_tokens": 10}
}
```

**流式请求：**
```json
{
  "model": "claude-opus-4-6",
  "messages": [{"role": "user", "content": "Hello"}],
  "stream": true
}
```

响应为标准 OpenAI SSE 格式：
```
data: {"id":"...","object":"chat.completion.chunk","choices":[{"delta":{"role":"assistant"},"finish_reason":null}]}

data: {"id":"...","object":"chat.completion.chunk","choices":[{"delta":{"content":"Hi!"},"finish_reason":null}]}

data: [DONE]
```

### Anthropic 原生格式 — Messages

```
POST /v1/messages
Authorization: Bearer {token}
Content-Type: application/json
```

**非流式请求：**
```json
{
  "model": "claude-opus-4-6",
  "messages": [{"role": "user", "content": "Hello"}],
  "max_tokens": 1024
}
```

响应为 Anthropic 原生格式（包含 `tool_use`、`cache_*` tokens 等全部字段）。

**流式请求：**
```json
{
  "model": "claude-opus-4-6",
  "messages": [{"role": "user", "content": "Hello"}],
  "max_tokens": 1024,
  "stream": true
}
```

响应为 Anthropic 原生 SSE 格式：
```
event: message_start
data: {"type":"message_start","message":{...}}

event: content_block_delta
data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"Hi!"}}

event: message_delta
data: {"type":"message_delta","usage":{"output_tokens":2}}

event: message_stop
data: {"type":"message_stop"}
```

### 文本完成

```
POST /v1/completions
Authorization: Bearer {token}
Content-Type: application/json

{"model": "gpt-3.5-turbo", "prompt": "Hello", "max_tokens": 10}
```

## Claude Code 接入

将模型中转服务作为 Claude Code 的后端：

```bash
export ANTHROPIC_AUTH_TOKEN="sk-your-token"
export ANTHROPIC_BASE_URL="https://ai.cngamesdk.com"
```

Claude Code 通过 `/v1/messages` 接口与服务通信，保留 Anthropic 全部原生特性（含 tool use、thinking、cache 等）。

## 配置说明

### 系统配置

| 字段 | 说明 | 默认值 |
|------|------|--------|
| `system.addr` | 监听端口 | `8080` |
| `system.mode` | 运行模式 | `release` |
| `dev_mode` | 开发模式（跳过认证） | `false` |

### 数据库

| 字段 | 说明 |
|------|------|
| `database.type` | 数据库类型（mysql） |
| `database.host` | 主机地址 |
| `database.port` | 端口 |
| `database.db-name` | 数据库名 |
| `database.username` | 用户名 |
| `database.password` | 密码 |
| `database.max-idle-conns` | 最大空闲连接数 |
| `database.max-open-conns` | 最大打开连接数 |

启动时自动创建 `dim_ai_token`（Token 管理）和 `ods_ai_usage_log`（用量日志）表。

### 日志

| 字段 | 说明 |
|------|------|
| `zap.level` | 日志级别（debug/info/warn/error） |
| `zap.format` | 输出格式（json/console） |
| `zap.director` | 日志目录 |
| `zap.retention-day` | 保留天数 |
| `zap.max-size` | 单文件最大大小（MB） |

### 提供商

```yaml
providers:
  - name: openai          # 提供商标识
    enabled: true         # 是否启用
    base_url: https://... # API 基础 URL
    api_key: sk-xxx       # API Key
    timeout: 300          # 请求超时（秒）
    models:               # 支持的模型列表
      - gpt-4
      - gpt-3.5-turbo
```

### 限流

| 字段 | 说明 |
|------|------|
| `rate_limit.enabled` | 是否启用 |
| `rate_limit.default_rpm` | 每分钟请求数 |
| `rate_limit.default_rph` | 每小时请求数 |

### 链路追踪

| 字段 | 说明 |
|------|------|
| `trace.header_name` | Trace ID 请求头名称 |
| `trace.generate_if_missing` | 缺失时自动生成 |

## 数据库表

### dim_ai_token — Token 管理

| 字段 | 说明 |
|------|------|
| `id` | 主键 |
| `token` | Token 字符串（唯一索引） |
| `name` | Token 名称 |
| `type` | 类型（1=企业, 2=个人） |
| `token_limit` | Token 配额上限 |
| `used_tokens` | 已使用 Token 数 |
| `expire_at` | 过期时间 |
| `status` | 状态（1=启用, 2=禁用） |

### ods_ai_usage_log — 用量日志

| 字段 | 说明 |
|------|------|
| `trace_id` | 链路追踪 ID |
| `token_id` | 关联 Token ID |
| `token_name` | Token 名称 |
| `provider` | AI 提供商 |
| `model` | 模型名称 |
| `request_tokens` | 输入 Token 数 |
| `response_tokens` | 输出 Token 数 |
| `total_tokens` | 总 Token 数 |
| `duration_ms` | 请求耗时（毫秒） |
| `status_code` | HTTP 状态码 |

## 项目结构

```
model-transfer-hub/
├── cmd/server/main.go          # 启动入口
├── config/                     # 配置结构定义
│   ├── config.go               # Server / System / Database 等
│   ├── provider.go             # Provider 模型
│   └── zap.go                  # 日志编码
├── core/                       # 核心组件
│   ├── server.go               # HTTP 服务生命周期
│   ├── viper.go                # Viper 配置加载
│   ├── zap.go                  # Zap 初始化
│   └── internal/               # 日志切割、zap core
├── global/global.go            # 全局变量
├── initialize/
│   ├── gorm.go                 # 数据库初始化
│   └── router.go               # 路由注册
├── middleware/
│   ├── trace.go                # Trace ID 注入
│   ├── logger.go               # 请求日志
│   ├── token_auth.go           # Bearer Token 验证
│   └── rate_limit.go           # 限流
├── handler/
│   ├── health.go               # 健康检查
│   ├── chat_completions.go     # OpenAI /v1/chat/completions
│   ├── messages_handler.go     # Anthropic /v1/messages（透传）
│   └── completions.go          # OpenAI /v1/completions
├── service/
│   ├── proxy_service.go        # 代理编排 + 用量记录
│   └── adapter/
│       ├── adapter.go          # Adapter 接口
│       ├── factory.go          # 工厂（按模型路由）
│       ├── openai.go           # OpenAI 适配器
│       ├── anthropic.go        # Anthropic 适配器
│       ├── google.go           # Google 适配器（未实现）
│       ├── http_client.go      # HTTP 请求辅助
│       └── sse_converter.go    # Anthropic SSE → OpenAI SSE 转换
├── model/
│   ├── request.go              # 请求模型
│   ├── response.go             # 响应模型（含 StreamResponse）
│   ├── token.go                # AiToken 数据库模型
│   └── usage.go                # AiUsageLog 数据库模型
├── pkg/
│   ├── logger/context_logger.go # Context 日志（trace_id / token_id）
│   └── trace/trace_id.go       # UUID 生成
├── config.example.yaml         # 配置示例
├── config.dev.yaml             # 开发环境配置
└── Makefile                    # 构建脚本
```

## 开发模式

配置 `dev_mode: true` 后，无 `Authorization` header 的请求自动使用 `dev` 身份通过认证，方便本地调试。

```yaml
system:
  mode: debug
dev_mode: true
```

## License

MIT
