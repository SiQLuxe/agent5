# AI 模型调用测试设计

**日期**: 2026-05-22  
**作者**: AI Agent  
**状态**: 待审查

## 1. 概述

本文档定义了 agent-tui 项目中 AI 模型调用成功验证的完整测试策略。通过单元测试和 HTTP mock 技术，确保 OpenAI、DeepSeek、Anthropic 和本地模型的客户端正确处理正常流程、错误情况和边界条件。

### 目标

1. **验证模型调用成功** - 确保每个 AI 客户端能够成功与 API 通信并正确解析响应
2. **完整的错误覆盖** - 处理网络错误、API 错误、超时等异常情况
3. **边界情况测试** - 验证空响应、特殊字符等极端情况
4. **代码复用** - 建立测试工具库，减少重复代码

## 2. 架构设计

### 2.1 目录结构

```
internal/ai/
├── client.go                        # 现有：客户端接口定义
├── openai.go                        # 现有：OpenAI 实现
├── deepseek.go                      # 现有：DeepSeek 实现
├── anthropic.go                     # 现有：Anthropic 实现
├── local.go                         # 现有：本地模型实现
├── testutil/                        # 新增：测试工具库
│   ├── mock.go                      # Mock 响应生成器
│   ├── httptest_helpers.go          # HTTP 测试服务器助手
│   └── fixtures.go                  # 测试数据集
├── openai_test.go                   # 扩展：OpenAI 测试
├── deepseek_test.go                 # 扩展：DeepSeek 测试
├── anthropic_test.go                # 扩展：Anthropic 测试
└── local_test.go                    # 扩展：本地模型测试
```

### 2.2 测试层级

```
层级 1: 单元测试（testutil/mock.go）
  └─ 生成标准 API 响应
  └─ 验证请求格式

层级 2: 集成测试（*_test.go）
  └─ HTTP mock 服务器（httptest）
  └─ 客户端实际调用
  └─ 响应解析验证

层级 3: 场景测试
  └─ 正常路径
  └─ 错误处理
  └─ 边界情况
```

## 3. 测试工具库详设

### 3.1 mock.go - Mock 响应生成器

**目的**：生成标准的 API 响应模板，避免重复代码

**核心接口**：

```go
// 生成成功的 ChatCompletion 响应
func NewMockChatResponse(model, content string) *ChatCompletionResponse

// 生成 API 错误响应
func NewMockErrorResponse(statusCode int, message string) []byte

// 生成流式响应数据
func NewMockStreamResponse(content string) string

// 创建标准的测试 HTTP 客户端
func NewMockHTTPClient() *http.Client
```

**使用示例**：
```go
resp := testutil.NewMockChatResponse("gpt-4", "Hello, world!")
// resp 包含完整的 ChatCompletionResponse 结构，可直接在 httptest 中使用
```

### 3.2 httptest_helpers.go - HTTP 测试助手

**目的**：简化 httptest 的使用，提供通用的验证函数

**核心函数**：

```go
// 创建测试 HTTP 服务器，返回服务器和 URL
func NewTestServer(handler http.HandlerFunc) (*httptest.Server, string)

// 验证 HTTP 请求内容
func AssertHTTPRequest(t *testing.T, req *http.Request, expectedMethod, expectedPath string)

// 验证请求头
func AssertHTTPHeaders(t *testing.T, req *http.Request, expectedHeaders map[string]string)

// 验证请求体（JSON）
func AssertRequestBody(t *testing.T, req *http.Request, expected interface{})
```

### 3.3 fixtures.go - 测试数据集

**目的**：定义标准的测试数据，保持数据一致性

**常量**：

```go
const (
    ValidAPIKey     = "test-api-key-12345"
    InvalidAPIKey   = ""
    ValidBaseURL    = "http://localhost:8080"
    InvalidBaseURL  = "http://invalid-url-that-does-not-exist:9999"
)

var (
    ValidMessages = []Message{
        {Role: "user", Content: "Hello, AI!"},
    }
    EmptyMessages = []Message{}
    LongMessage   = Message{
        Role: "user",
        Content: strings.Repeat("x", 10000),
    }
)
```

## 4. 测试场景设计

### 4.1 正常路径测试

**测试场景**：客户端成功发送请求并获得正确响应

**验证点**：
- ✓ 请求能够成功发送
- ✓ 响应正确解析
- ✓ 返回值符合预期
- ✓ 错误字段为 nil

**示例**：
```
TestOpenAIClient_ChatCompletion_Success
  1. 创建 httptest 服务器，模拟 /v1/chat/completions 返回成功响应
  2. 创建 OpenAI 客户端，指向测试服务器
  3. 调用 ChatCompletion()
  4. 验证返回内容、模型名称、tokens 使用情况
```

### 4.2 错误处理测试

**测试场景**：验证客户端正确处理各种错误情况

| 错误类型 | HTTP 状态码 | 测试内容 | 验证 |
|---------|-----------|---------|------|
| 无效 API Key | 401 | 发送无效密钥 | 返回错误，不进行重试 |
| 限流 | 429 | 发送过多请求 | 返回限流错误 |
| 服务不可用 | 503 | 服务器离线 | 返回错误，建议重试 |
| 格式错误 | 400 | 请求格式不对 | 返回验证错误 |
| 响应格式错误 | 200 + 无效JSON | API 返回无效JSON | 返回解析错误 |
| 网络连接错误 | N/A | 连接拒绝 | 返回网络错误 |

### 4.3 边界情况测试

**测试场景**：处理极端或特殊输入

| 情况 | 输入 | 预期行为 |
|-----|------|--------|
| 空消息列表 | Messages: [] | 拒绝或返回验证错误 |
| 超长消息 | 10000+ 字符 | 正常发送或返回长度错误 |
| 特殊字符 | 中文、emoji、\n | 正确编码和解析 |
| 空响应 | Choices 为空 | 返回空内容但不崩溃 |
| 空 API Key | "" | 拒绝请求 |
| 空 Base URL | "" | 使用默认 URL 或返回错误 |

## 5. 每个客户端的测试清单

### 5.1 OpenAI 客户端测试（openai_test.go）

```
✓ TestOpenAIClient_ChatCompletion_Success
  - 验证成功调用和响应解析

✓ TestOpenAIClient_ChatCompletion_APIError
  - 验证 401 Unauthorized 处理
  - 验证 429 Rate Limited 处理
  - 验证 500 Server Error 处理

✓ TestOpenAIClient_ChatCompletion_NetworkError
  - 验证网络连接失败处理

✓ TestOpenAIClient_ChatCompletion_InvalidResponse
  - 验证响应 JSON 格式错误处理

✓ TestOpenAIClient_ChatCompletion_EdgeCases
  - 空消息列表
  - 超长消息
  - 特殊字符

✓ TestOpenAIClient_SetAPIKey
  - 验证 API Key 设置

✓ TestOpenAIClient_SetBaseURL
  - 验证 Base URL 设置

✓ TestOpenAIClient_ListModels
  - 验证模型列表返回
```

### 5.2 DeepSeek 客户端测试（deepseek_test.go）

类似 OpenAI，考虑 DeepSeek 特定的 API 响应格式

### 5.3 Anthropic 客户端测试（anthropic_test.go）

类似 OpenAI，考虑 Anthropic 特定的 API 响应格式和错误类型

### 5.4 本地模型客户端测试（local_test.go）

```
✓ TestLocalClient_ChatCompletion_Success
  - 验证本地 API 调用

✓ TestLocalClient_ChatCompletion_ConnectionError
  - 验证本地服务不可用

✓ TestLocalClient_DefaultValues
  - 验证默认模型和 URL
```

## 6. 测试执行策略

### 6.1 运行方式

```bash
# 运行所有 AI 客户端测试
go test -v ./internal/ai/...

# 运行单个客户端测试
go test -v ./internal/ai -run TestOpenAIClient

# 运行特定测试
go test -v ./internal/ai -run TestOpenAIClient_ChatCompletion_Success

# 显示覆盖率
go test -v ./internal/ai -cover
```

### 6.2 CI/CD 集成

- 每个 PR 必须通过所有单元测试
- 测试覆盖率不低于 80%
- 使用 `httptest` 保证快速执行（无真实网络调用）

## 7. 数据流与依赖

### 7.1 数据流

```
测试代码
  ↓
testutil/fixtures.go (准备数据)
  ↓
testutil/httptest_helpers.go (创建 mock 服务器)
  ↓
httptest.Server (模拟 API)
  ↓
客户端 (OpenAI/DeepSeek/Anthropic/Local)
  ↓
testutil/mock.go (验证响应)
  ↓
断言 (t.Errorf, assert.Equal 等)
```

### 7.2 依赖项

- `net/http/httptest` - Go 标准库，无外部依赖
- `testing` - Go 标准库
- 可选：`github.com/stretchr/testify` - 简化断言（建议但非必须）

## 8. 成功标准

1. **覆盖率** - 所有客户端函数覆盖率 ≥ 80%
2. **执行时间** - 所有测试执行时间 < 5 秒
3. **隔离性** - 测试相互独立，不依赖执行顺序
4. **可读性** - 测试名称清晰，一个测试一个场景
5. **维护性** - 使用 testutil 库减少重复代码

## 9. 扩展点

- **并发测试** - 若未来需要测试并发调用，添加 goroutine 相关的测试
- **性能测试** - 基准测试 (benchmark) 来检测性能回归
- **流式响应** - ChatCompletionStream 的完整测试
- **集成测试** - 真实 API 集成验证（可选，需要 API 密钥）

## 10. 审批清单

- [ ] 架构合理性
- [ ] 测试场景完整性
- [ ] 代码复用充分性
- [ ] 文档清晰性
- [ ] 是否遗漏重要场景
