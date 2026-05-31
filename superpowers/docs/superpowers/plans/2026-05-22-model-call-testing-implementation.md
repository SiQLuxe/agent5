# AI 模型调用测试 实现计划

> **对于自动化执行者:** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 来逐任务实现该计划。步骤使用复选框（`- [ ]`）语法进行跟踪。

**目标**：为 agent-tui 的所有 AI 客户端（OpenAI、DeepSeek、Anthropic、本地模型）建立完整的单元测试框架，通过 httptest mock 验证模型调用成功、错误处理和边界情况。

**架构**：建立分层测试架构 - 测试工具库（mock 生成器、httptest 助手、测试数据）+ 客户端单元测试（使用 httptest 模拟 HTTP 服务器）。每个客户端独立测试，代码复用最大化。

**技术栈**：Go 标准库 (`net/http/httptest`, `testing`), 无外部依赖（可选 `testify` 简化断言）

---

## 文件结构映射

**新增文件**：
- `internal/ai/testutil/mock.go` - Mock 响应生成器
- `internal/ai/testutil/httptest_helpers.go` - HTTP 测试服务器助手
- `internal/ai/testutil/fixtures.go` - 测试数据集

**修改文件**：
- `internal/ai/openai_test.go` - 扩展 OpenAI 测试覆盖
- `internal/ai/deepseek_test.go` - 扩展 DeepSeek 测试覆盖
- `internal/ai/anthropic_test.go` - 扩展 Anthropic 测试覆盖
- `internal/ai/local_test.go` - 扩展本地模型测试覆盖

**现有文件**（不修改）：
- `internal/ai/client.go` - 客户端接口
- `internal/ai/openai.go`, `deepseek.go`, `anthropic.go`, `local.go` - 客户端实现

---

## 实现任务

### Task 1: 创建测试工具库 - fixtures.go

**文件**：
- Create: `internal/ai/testutil/fixtures.go`

- [ ] **Step 1: 创建 testutil 目录和 fixtures 文件**

```bash
mkdir -p internal/ai/testutil
touch internal/ai/testutil/fixtures.go
```

- [ ] **Step 2: 编写测试数据集**

```go
package testutil

import (
	"strings"
	"github.com/example/agent-tui/internal/ai"
)

// API Key 测试数据
const (
	ValidAPIKey     = "test-api-key-12345"
	InvalidAPIKey   = ""
	ExpiredAPIKey   = "expired-key"
)

// Base URL 测试数据
const (
	ValidBaseURL           = "http://localhost:8080"
	InvalidBaseURL         = "http://invalid-url-that-does-not-exist:9999"
	LocalhostURL           = "http://127.0.0.1:8000"
)

// 消息测试数据
var (
	ValidMessages = []ai.Message{
		{Role: "user", Content: "Hello, AI!"},
	}

	MultipleMessages = []ai.Message{
		{Role: "user", Content: "First message"},
		{Role: "assistant", Content: "First response"},
		{Role: "user", Content: "Second message"},
	}

	EmptyMessages = []ai.Message{}

	LongMessage = ai.Message{
		Role:    "user",
		Content: strings.Repeat("x", 10000),
	}

	SpecialCharMessage = ai.Message{
		Role:    "user",
		Content: "测试中文和emoji 🎉\n带换行符",
	}
)

// ChatCompletionRequest 测试数据
var (
	ValidChatRequest = ai.ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: ValidMessages,
		Stream:   false,
	}

	EmptyMessageRequest = ai.ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: EmptyMessages,
		Stream:   false,
	}
)

// 模型列表
var ValidModels = []string{
	"gpt-4",
	"gpt-4-turbo",
	"gpt-3.5-turbo",
}
```

- [ ] **Step 3: 验证文件编译**

```bash
cd internal/ai && go build ./testutil
```

期望：编译成功，无错误

- [ ] **Step 4: Commit**

```bash
git add internal/ai/testutil/fixtures.go
git commit -m "test: add test fixtures for AI client testing"
```

---

### Task 2: 创建测试工具库 - mock.go

**文件**：
- Create: `internal/ai/testutil/mock.go`

- [ ] **Step 1: 编写 Mock 响应生成器**

```go
package testutil

import (
	"encoding/json"
	"time"
	"github.com/example/agent-tui/internal/ai"
)

// NewMockChatResponse 生成成功的 ChatCompletion 响应
func NewMockChatResponse(model, content string) *ai.ChatCompletionResponse {
	return &ai.ChatCompletionResponse{
		ID:      "chatcmpl-test-12345",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []struct {
			Index        int           `json:"index"`
			Message      ai.Message    `json:"message"`
			FinishReason string        `json:"finish_reason"`
		}{
			{
				Index: 0,
				Message: ai.Message{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: "stop",
			},
		},
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}
}

// NewMockErrorResponse 生成 API 错误响应的 JSON
func NewMockErrorResponse(statusCode int, message string) []byte {
	errResp := map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
			"type":    "api_error",
			"code":    statusCode,
		},
	}
	data, _ := json.Marshal(errResp)
	return data
}

// NewMockStreamResponse 生成流式响应（SSE 格式）
func NewMockStreamResponse(content string) string {
	resp := map[string]interface{}{
		"id":      "chatcmpl-stream-123",
		"object":  "chat.completion.chunk",
		"created": time.Now().Unix(),
		"model":   "gpt-4",
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"delta": map[string]interface{}{
					"role":    "assistant",
					"content": content,
				},
				"finish_reason": nil,
			},
		},
	}
	data, _ := json.Marshal(resp)
	return "data: " + string(data) + "\n\n"
}

// MockResponseTemplates 返回 OpenAI 格式的 mock 响应模板
var MockResponseTemplates = map[string]string{
	"success_openai": `{
		"id": "chatcmpl-test",
		"object": "chat.completion",
		"created": 1234567890,
		"model": "gpt-4",
		"choices": [{
			"index": 0,
			"message": {"role": "assistant", "content": "Test response"},
			"finish_reason": "stop"
		}],
		"usage": {"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30}
	}`,
	"error_unauthorized": `{
		"error": {"message": "Unauthorized", "type": "invalid_request_error", "code": 401}
	}`,
	"error_ratelimit": `{
		"error": {"message": "Rate limit exceeded", "type": "rate_limit_error", "code": 429}
	}`,
	"error_server": `{
		"error": {"message": "Internal server error", "type": "server_error", "code": 500}
	}`,
}
```

- [ ] **Step 2: 验证编译和语法**

```bash
cd internal/ai && go build ./testutil
```

期望：编译成功

- [ ] **Step 3: Commit**

```bash
git add internal/ai/testutil/mock.go
git commit -m "test: add mock response generators"
```

---

### Task 3: 创建测试工具库 - httptest_helpers.go

**文件**：
- Create: `internal/ai/testutil/httptest_helpers.go`

- [ ] **Step 1: 编写 HTTP 测试服务器助手**

```go
package testutil

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// NewTestServer 创建一个测试 HTTP 服务器，返回服务器和 URL
func NewTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, string) {
	server := httptest.NewServer(handler)
	t.Cleanup(func() { server.Close() })
	return server, server.URL
}

// AssertHTTPRequest 验证 HTTP 请求的方法和路径
func AssertHTTPRequest(t *testing.T, req *http.Request, expectedMethod, expectedPath string) {
	if req.Method != expectedMethod {
		t.Errorf("expected method %s, got %s", expectedMethod, req.Method)
	}
	if req.URL.Path != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, req.URL.Path)
	}
}

// AssertHTTPHeaders 验证请求中的特定请求头
func AssertHTTPHeaders(t *testing.T, req *http.Request, expectedHeaders map[string]string) {
	for key, expectedValue := range expectedHeaders {
		actualValue := req.Header.Get(key)
		if actualValue != expectedValue {
			t.Errorf("expected header %s=%s, got %s", key, expectedValue, actualValue)
		}
	}
}

// AssertRequestBody 验证请求体是否能解析为指定的 JSON 结构
func AssertRequestBody(t *testing.T, req *http.Request, expectedData interface{}) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Errorf("failed to read request body: %v", err)
		return
	}

	var actualData interface{}
	if err := json.Unmarshal(body, &actualData); err != nil {
		t.Errorf("failed to parse request body as JSON: %v", err)
		return
	}

	expectedJSON, _ := json.Marshal(expectedData)
	var expectedParsed interface{}
	json.Unmarshal(expectedJSON, &expectedParsed)

	actualJSON, _ := json.Marshal(actualData)
	expectedJSON, _ = json.Marshal(expectedParsed)

	if string(actualJSON) != string(expectedJSON) {
		t.Errorf("request body mismatch.\nExpected: %s\nActual: %s",
			string(expectedJSON), string(actualJSON))
	}
}

// MockHandler 创建一个标准的 mock 处理函数
type MockHandler struct {
	StatusCode   int
	ResponseBody string
	RequestsReceived int
}

// ServeHTTP 实现 http.Handler 接口
func (m *MockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.RequestsReceived++
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(m.StatusCode)
	w.Write([]byte(m.ResponseBody))
}
```

- [ ] **Step 2: 验证编译**

```bash
cd internal/ai && go build ./testutil
```

期望：编译成功

- [ ] **Step 3: Commit**

```bash
git add internal/ai/testutil/httptest_helpers.go
git commit -m "test: add HTTP test server helpers"
```

---

### Task 4: 扩展 OpenAI 客户端测试

**文件**：
- Modify: `internal/ai/openai_test.go`

- [ ] **Step 1: 编写成功调用测试**

```go
package ai

import (
	"testing"
	"github.com/example/agent-tui/internal/ai/testutil"
)

func TestOpenAIClient_ChatCompletion_Success(t *testing.T) {
	// 创建 mock 响应
	mockResponse := testutil.NewMockChatResponse("gpt-4", "Hello from OpenAI")
	mockJSON := testutil.MockResponseTemplates["success_openai"]

	// 创建测试服务器
	server, url := testutil.NewTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		testutil.AssertHTTPRequest(t, r, "POST", "/v1/chat/completions")
		testutil.AssertHTTPHeaders(t, r, map[string]string{
			"Authorization": "Bearer test-api-key-12345",
		})
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockJSON))
	})

	// 创建客户端
	client, err := NewOpenAIClient(testutil.ValidAPIKey, url, "gpt-4")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// 执行测试
	resp, err := client.ChatCompletion(testutil.ValidChatRequest)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 验证响应
	if resp == nil {
		t.Fatal("response is nil")
	}
	if resp.Model != "gpt-4" {
		t.Errorf("expected model gpt-4, got %s", resp.Model)
	}
	if len(resp.Choices) == 0 {
		t.Fatal("no choices in response")
	}
	if resp.Choices[0].Message.Content != "Test response" {
		t.Errorf("unexpected response content: %s", resp.Choices[0].Message.Content)
	}
}
```

- [ ] **Step 2: 运行测试验证通过**

```bash
cd internal/ai && go test -v -run TestOpenAIClient_ChatCompletion_Success
```

期望：PASS

- [ ] **Step 3: 编写 API 错误测试（401 Unauthorized）**

```go
func TestOpenAIClient_ChatCompletion_Unauthorized(t *testing.T) {
	mockJSON := testutil.MockResponseTemplates["error_unauthorized"]

	server, url := testutil.NewTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(401)
		w.Write([]byte(mockJSON))
	})

	client, _ := NewOpenAIClient("invalid-key", url, "gpt-4")
	resp, err := client.ChatCompletion(testutil.ValidChatRequest)

	if err == nil {
		t.Error("expected error for 401, got nil")
	}
	if resp != nil {
		t.Error("expected nil response for error case")
	}
}
```

- [ ] **Step 4: 编写 API 错误测试（429 Rate Limited）**

```go
func TestOpenAIClient_ChatCompletion_RateLimited(t *testing.T) {
	mockJSON := testutil.MockResponseTemplates["error_ratelimit"]

	server, url := testutil.NewTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(429)
		w.Write([]byte(mockJSON))
	})

	client, _ := NewOpenAIClient(testutil.ValidAPIKey, url, "gpt-4")
	resp, err := client.ChatCompletion(testutil.ValidChatRequest)

	if err == nil {
		t.Error("expected error for 429, got nil")
	}
}
```

- [ ] **Step 5: 编写 API 错误测试（500 Server Error）**

```go
func TestOpenAIClient_ChatCompletion_ServerError(t *testing.T) {
	mockJSON := testutil.MockResponseTemplates["error_server"]

	server, url := testutil.NewTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(mockJSON))
	})

	client, _ := NewOpenAIClient(testutil.ValidAPIKey, url, "gpt-4")
	resp, err := client.ChatCompletion(testutil.ValidChatRequest)

	if err == nil {
		t.Error("expected error for 500, got nil")
	}
}
```

- [ ] **Step 6: 编写响应解析错误测试**

```go
func TestOpenAIClient_ChatCompletion_InvalidJSON(t *testing.T) {
	server, url := testutil.NewTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	})

	client, _ := NewOpenAIClient(testutil.ValidAPIKey, url, "gpt-4")
	resp, err := client.ChatCompletion(testutil.ValidChatRequest)

	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
	if resp != nil {
		t.Error("expected nil response")
	}
}
```

- [ ] **Step 7: 编写边界情况测试（特殊字符）**

```go
func TestOpenAIClient_ChatCompletion_SpecialCharacters(t *testing.T) {
	mockResponse := testutil.NewMockChatResponse("gpt-4", "回复: 测试🎉")
	respJSON, _ := json.Marshal(mockResponse)

	server, url := testutil.NewTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(respJSON)
	})

	client, _ := NewOpenAIClient(testutil.ValidAPIKey, url, "gpt-4")
	req := ai.ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []ai.Message{{Role: "user", Content: "测试中文和emoji"}},
	}
	resp, err := client.ChatCompletion(req)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp == nil || resp.Choices[0].Message.Content != "回复: 测试🎉" {
		t.Error("special characters not handled correctly")
	}
}
```

- [ ] **Step 8: 运行所有 OpenAI 测试**

```bash
cd internal/ai && go test -v -run TestOpenAIClient
```

期望：所有测试 PASS

- [ ] **Step 9: Commit**

```bash
git add internal/ai/openai_test.go
git commit -m "test: add comprehensive OpenAI client tests with httptest"
```

---

### Task 5: 扩展 DeepSeek 客户端测试

**文件**：
- Modify: `internal/ai/deepseek_test.go`

- [ ] **Step 1-2: 编写基础测试框架**

按照 OpenAI 测试的模式编写，调整 API 端点和响应格式（DeepSeek 格式可能略有不同）

关键测试：
- `TestDeepSeekClient_ChatCompletion_Success`
- `TestDeepSeekClient_ChatCompletion_Unauthorized`
- `TestDeepSeekClient_ChatCompletion_RateLimited`
- `TestDeepSeekClient_ChatCompletion_ServerError`
- `TestDeepSeekClient_ChatCompletion_InvalidJSON`
- `TestDeepSeekClient_ChatCompletion_EdgeCases`

- [ ] **Step 3: 运行所有 DeepSeek 测试**

```bash
cd internal/ai && go test -v -run TestDeepSeekClient
```

期望：所有测试 PASS

- [ ] **Step 4: Commit**

```bash
git add internal/ai/deepseek_test.go
git commit -m "test: add comprehensive DeepSeek client tests"
```

---

### Task 6: 扩展 Anthropic 客户端测试

**文件**：
- Modify: `internal/ai/anthropic_test.go`

- [ ] **Step 1-2: 编写基础测试框架**

按照 OpenAI/DeepSeek 测试的模式编写，调整 Anthropic 特定的 API 端点和响应格式

关键测试：
- `TestAnthropicClient_ChatCompletion_Success`
- `TestAnthropicClient_ChatCompletion_Unauthorized`
- `TestAnthropicClient_ChatCompletion_RateLimited`
- `TestAnthropicClient_ChatCompletion_ServerError`
- `TestAnthropicClient_ChatCompletion_InvalidJSON`
- `TestAnthropicClient_ChatCompletion_EdgeCases`

- [ ] **Step 3: 运行所有 Anthropic 测试**

```bash
cd internal/ai && go test -v -run TestAnthropicClient
```

期望：所有测试 PASS

- [ ] **Step 4: Commit**

```bash
git add internal/ai/anthropic_test.go
git commit -m "test: add comprehensive Anthropic client tests"
```

---

### Task 7: 扩展本地模型客户端测试

**文件**：
- Modify: `internal/ai/local_test.go`

- [ ] **Step 1-2: 编写基础测试框架**

本地模型测试关键点：
- 连接到本地 HTTP 服务器而不是远程 API
- 验证默认 URL 和模型
- 测试连接失败场景

关键测试：
- `TestLocalClient_ChatCompletion_Success`
- `TestLocalClient_ChatCompletion_ConnectionError`
- `TestLocalClient_ChatCompletion_InvalidResponse`
- `TestLocalClient_DefaultValues`
- `TestLocalClient_SetURLAndModel`

- [ ] **Step 3: 运行所有本地模型测试**

```bash
cd internal/ai && go test -v -run TestLocalClient
```

期望：所有测试 PASS

- [ ] **Step 4: Commit**

```bash
git add internal/ai/local_test.go
git commit -m "test: add comprehensive local model client tests"
```

---

### Task 8: 整体测试覆盖率验证

**文件**：
- No new files

- [ ] **Step 1: 运行所有 AI 客户端测试**

```bash
cd internal/ai && go test -v ./...
```

期望：所有测试 PASS

- [ ] **Step 2: 检查测试覆盖率**

```bash
cd internal/ai && go test -cover ./...
```

期望：每个包的覆盖率 ≥ 80%

- [ ] **Step 3: 生成详细覆盖率报告**

```bash
cd internal/ai && go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

查看 `coverage.html` 文件，验证关键代码路径已覆盖

- [ ] **Step 4: 运行基准测试（可选）**

```bash
cd internal/ai && go test -bench=. -benchtime=1s
```

验证测试性能（所有测试应在 2 秒内完成）

- [ ] **Step 5: 最终提交**

```bash
git add internal/ai/testutil/
git add internal/ai/*_test.go
git commit -m "test: complete AI client testing framework with 80%+ coverage"
```

---

## 验收标准

✓ **功能覆盖**
- OpenAI、DeepSeek、Anthropic、Local 四个客户端都有完整测试
- 每个客户端覆盖：成功、错误（401/429/500）、解析错误、边界情况

✓ **代码质量**
- 测试覆盖率 ≥ 80%
- 所有测试命名清晰（TestXXX_Scene_ExpectedResult）
- 使用 testutil 库减少代码重复

✓ **性能**
- 所有测试运行时间 < 5 秒
- 使用 httptest 无网络调用

✓ **可维护性**
- 每个测试独立、不依赖执行顺序
- 新增客户端可以轻松添加测试

---

## 注意事项

1. **导入调整** - 确保所有测试文件导入了 testutil 包
2. **JSON 处理** - 某些客户端可能有特殊的 JSON 格式，需要调整 mock 响应
3. **错误类型** - 各客户端的错误处理可能有差异，测试时验证实际错误类型
4. **并发测试** - 本计划暂不包含并发测试，若需添加可在扩展阶段进行
5. **流式响应** - ChatCompletionStream 的详细测试留作扩展项
