# 配置文件密钥安全管理实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现安全的密钥管理，支持多模型配置和环境变量覆盖

**Architecture:** 修改配置模块支持多模型结构，实现环境变量优先级覆盖机制，添加 .gitignore 防止敏感文件泄露

**Tech Stack:** Go, TOML, os.Getenv

---

## 文件结构

```
internal/data/config/
├── config.go          # 核心配置结构和加载逻辑（修改）
├── config_test.go     # 配置测试（修改）
└── env.go             # 环境变量覆盖逻辑（新建）

internal/ai/
├── factory.go         # 客户端工厂（修改）
└── volcengine.go      # 火山引擎客户端（新建，如不存在）

configs/
├── config.toml        # 实际配置文件
└── config.example.toml # 配置模板

.gitignore             # Git 忽略规则（新建）
```

---

## Task 1: 修改配置结构支持多模型

**Files:**
- Modify: `internal/data/config/config.go`
- Modify: `internal/data/config/config_test.go`

- [ ] **Step 1: 读取现有配置代码**

```go
// internal/data/config/config.go 当前结构
type Config struct {
    APIKeys      APIKeys `toml:"api_keys"`
    DefaultModel string  `toml:"default_model"`
    Theme        string  `toml:"theme"`
    ApprovalMode string  `toml:"approval_mode"`
    MaxSubagents int     `toml:"max_subagents"`
    AuthToken    string  `toml:"auth_token"`
    BaseURL      string  `toml:"base_url"`
}

type APIKeys struct {
    OpenAI    string `toml:"openai"`
    DeepSeek  string `toml:"deepseek"`
    Anthropic string `toml:"anthropic"`
}
```

- [ ] **Step 2: 更新配置结构**

```go
type Config struct {
    Models        ModelsConfig `toml:"models"`
    DefaultClient string       `toml:"default_client"`
    Theme         string       `toml:"theme"`
    ApprovalMode  string       `toml:"approval_mode"`
    MaxSubagents  int          `toml:"max_subagents"`
}

type ModelsConfig struct {
    OpenAI     ModelConfig `toml:"openai"`
    DeepSeek   ModelConfig `toml:"deepseek"`
    Local      ModelConfig `toml:"local"`
    VolcEngine ModelConfig `toml:"volcengine"`
}

type ModelConfig struct {
    APIKey       string `toml:"api_key"`
    BaseURL      string `toml:"base_url"`
    DefaultModel string `toml:"default_model"`
}
```

- [ ] **Step 3: 更新测试验证新结构**

```go
// internal/data/config/config_test.go
func TestConfigStructure(t *testing.T) {
    cfg := GetDefaultConfig()
    if cfg.DefaultClient != "volcengine" {
        t.Errorf("expected default client 'volcengine', got '%s'", cfg.DefaultClient)
    }
    if cfg.Models.VolcEngine.BaseURL == "" {
        t.Error("volcengine base url should have default value")
    }
}
```

- [ ] **Step 4: 运行测试验证**

Run: `go test -v ./internal/data/config/...`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/data/config/config.go internal/data/config/config_test.go
git commit -m "refactor: update config structure for multi-model support"
```

---

## Task 2: 创建环境变量覆盖逻辑

**Files:**
- Create: `internal/data/config/env.go`
- Modify: `internal/data/config/config.go`

- [ ] **Step 1: 创建 env.go 文件**

```go
// internal/data/config/env.go
package config

import "os"

func applyEnvOverrides(cfg *Config) {
    // OpenAI
    if val := os.Getenv("OPENAI_API_KEY"); val != "" {
        cfg.Models.OpenAI.APIKey = val
    }
    if val := os.Getenv("OPENAI_BASE_URL"); val != "" {
        cfg.Models.OpenAI.BaseURL = val
    }
    if val := os.Getenv("OPENAI_DEFAULT_MODEL"); val != "" {
        cfg.Models.OpenAI.DefaultModel = val
    }

    // DeepSeek
    if val := os.Getenv("DEEPSEEK_API_KEY"); val != "" {
        cfg.Models.DeepSeek.APIKey = val
    }
    if val := os.Getenv("DEEPSEEK_BASE_URL"); val != "" {
        cfg.Models.DeepSeek.BaseURL = val
    }
    if val := os.Getenv("DEEPSEEK_DEFAULT_MODEL"); val != "" {
        cfg.Models.DeepSeek.DefaultModel = val
    }

    // Local
    if val := os.Getenv("LOCAL_BASE_URL"); val != "" {
        cfg.Models.Local.BaseURL = val
    }
    if val := os.Getenv("LOCAL_DEFAULT_MODEL"); val != "" {
        cfg.Models.Local.DefaultModel = val
    }

    // VolcEngine
    if val := os.Getenv("VOLCENGINE_API_KEY"); val != "" {
        cfg.Models.VolcEngine.APIKey = val
    }
    if val := os.Getenv("VOLCENGINE_BASE_URL"); val != "" {
        cfg.Models.VolcEngine.BaseURL = val
    }
    if val := os.Getenv("VOLCENGINE_DEFAULT_MODEL"); val != "" {
        cfg.Models.VolcEngine.DefaultModel = val
    }

    // Default client
    if val := os.Getenv("DEFAULT_CLIENT"); val != "" {
        cfg.DefaultClient = val
    }
}
```

- [ ] **Step 2: 修改 LoadConfig 应用环境变量覆盖**

```go
// internal/data/config/config.go
func LoadConfig(path string) (*Config, error) {
    var cfg Config
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    if err := toml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }
    applyEnvOverrides(&cfg)
    return &cfg, nil
}
```

- [ ] **Step 3: 编写环境变量覆盖测试**

```go
// internal/data/config/env_test.go
func TestEnvOverrides(t *testing.T) {
    os.Setenv("OPENAI_API_KEY", "test-key")
    defer os.Unsetenv("OPENAI_API_KEY")

    cfg := &Config{
        Models: ModelsConfig{
            OpenAI: ModelConfig{APIKey: "original"},
        },
    }
    applyEnvOverrides(cfg)

    if cfg.Models.OpenAI.APIKey != "test-key" {
        t.Errorf("expected 'test-key', got '%s'", cfg.Models.OpenAI.APIKey)
    }
}
```

- [ ] **Step 4: 运行测试验证**

Run: `go test -v ./internal/data/config/...`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/data/config/env.go internal/data/config/config.go
git commit -m "feat: add environment variable override support"
```

---

## Task 3: 创建火山引擎客户端

**Files:**
- Create: `internal/ai/volcengine.go`
- Create: `internal/ai/volcengine_test.go`

- [ ] **Step 1: 创建火山引擎客户端**

```go
// internal/ai/volcengine.go
package ai

import (
    "bytes"
    "context"
    "encoding/json"
    "io"
    "net/http"
    "strings"
)

type VolcEngineClient struct {
    apiKey  string
    baseURL string
}

func NewVolcEngineClient(apiKey, baseURL string) *VolcEngineClient {
    if baseURL == "" {
        baseURL = "https://ark.cn-beijing.volces.com/api/coding"
    }
    return &VolcEngineClient{
        apiKey:  apiKey,
        baseURL: baseURL,
    }
}

func (c *VolcEngineClient) GetModels() []ModelInfo {
    return []ModelInfo{
        {ID: "glm-4", Name: "GLM-4", MaxTokens: 128000},
        {ID: "glm-4v", Name: "GLM-4V", MaxTokens: 4096},
        {ID: "glm-5", Name: "GLM-5", MaxTokens: 128000},
    }
}

func (c *VolcEngineClient) ValidateAPIKey() error {
    return nil
}

func (c *VolcEngineClient) Completion(ctx context.Context, req CompletionRequest) (<-chan CompletionResponse, error) {
    ch := make(chan CompletionResponse)

    go func() {
        defer close(ch)

        messages := make([]map[string]string, len(req.Messages))
        for i, msg := range req.Messages {
            messages[i] = map[string]string{"role": msg.Role, "content": msg.Content}
        }

        payload := map[string]interface{}{
            "model":    req.Model,
            "messages": messages,
            "stream":   req.Stream,
        }

        jsonData, _ := json.Marshal(payload)
        httpReq, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", strings.NewReader(string(jsonData)))
        httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
        httpReq.Header.Set("Content-Type", "application/json")

        client := &http.Client{}
        resp, err := client.Do(httpReq)
        if err != nil {
            ch <- CompletionResponse{Error: err}
            return
        }
        defer resp.Body.Close()

        if req.Stream {
            c.handleStream(resp.Body, ch)
        } else {
            c.handleNonStream(resp.Body, ch)
        }
    }()

    return ch, nil
}

func (c *VolcEngineClient) handleStream(body io.ReadCloser, ch chan CompletionResponse) {
    buf := make([]byte, 1024)
    for {
        n, err := body.Read(buf)
        if err != nil {
            break
        }
        ch <- CompletionResponse{Content: string(buf[:n]), Done: false}
    }
    ch <- CompletionResponse{Done: true}
}

func (c *VolcEngineClient) handleNonStream(body io.ReadCloser, ch chan CompletionResponse) {
    data, _ := io.ReadAll(body)
    ch <- CompletionResponse{Content: string(data), Done: true}
}
```

- [ ] **Step 2: 编写测试**

```go
// internal/ai/volcengine_test.go
package ai

import "testing"

func TestNewVolcEngineClient(t *testing.T) {
    c := NewVolcEngineClient("test-key", "")
    if c.apiKey != "test-key" {
        t.Errorf("expected api key 'test-key', got '%s'", c.apiKey)
    }
    if c.baseURL != "https://ark.cn-beijing.volces.com/api/coding" {
        t.Errorf("unexpected default base url: %s", c.baseURL)
    }
}

func TestVolcEngineGetModels(t *testing.T) {
    c := NewVolcEngineClient("", "")
    models := c.GetModels()
    if len(models) == 0 {
        t.Error("should return default models")
    }
}
```

- [ ] **Step 3: 运行测试验证**

Run: `go test -v ./internal/ai/... -run VolcEngine`
Expected: PASS

- [ ] **Step 4: 提交**

```bash
git add internal/ai/volcengine.go internal/ai/volcengine_test.go
git commit -m "feat: add VolcEngine client support"
```

---

## Task 4: 更新客户端工厂函数

**Files:**
- Modify: `internal/ai/factory.go`

- [ ] **Step 1: 读取现有工厂代码**

```go
// internal/ai/factory.go 当前内容
func NewClientFromConfig(cfg *config.Config) (AIClient, error) {
    client, model, err := config.ParseModelPrefix(cfg.DefaultModel)
    // ... 现有逻辑
}
```

- [ ] **Step 2: 更新工厂函数**

```go
// internal/ai/factory.go
func NewClientFromConfig(cfg *config.Config) (AIClient, error) {
    defaultClient := cfg.DefaultClient
    if defaultClient == "" {
        defaultClient = "volcengine"
    }

    switch defaultClient {
    case "openai":
        return NewOpenAIClient(cfg.Models.OpenAI.APIKey, cfg.Models.OpenAI.BaseURL), nil
    case "deepseek":
        return NewDeepSeekClient(cfg.Models.DeepSeek.APIKey, cfg.Models.DeepSeek.BaseURL), nil
    case "local":
        return NewLocalClient(cfg.Models.Local.BaseURL), nil
    case "volcengine":
        return NewVolcEngineClient(cfg.Models.VolcEngine.APIKey, cfg.Models.VolcEngine.BaseURL), nil
    default:
        return nil, fmt.Errorf("unknown client: %s", defaultClient)
    }
}
```

- [ ] **Step 3: 编写测试**

```go
// internal/ai/factory_test.go 添加测试
func TestNewClientFromConfigVolcEngine(t *testing.T) {
    cfg := &config.Config{
        DefaultClient: "volcengine",
        Models: config.ModelsConfig{
            VolcEngine: config.ModelConfig{
                APIKey:  "test-key",
                BaseURL: "https://ark.cn-beijing.volces.com/api/coding",
            },
        },
    }

    client, err := NewClientFromConfig(cfg)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if client == nil {
        t.Error("client should not be nil")
    }
}
```

- [ ] **Step 4: 运行测试验证**

Run: `go test -v ./internal/ai/... -run Factory`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/ai/factory.go internal/ai/factory_test.go
git commit -m "feat: update factory to support multi-model config"
```

---

## Task 5: 创建 .gitignore 和更新配置文件

**Files:**
- Create: `.gitignore`
- Modify: `configs/config.toml`
- Modify: `configs/config.example.toml`

- [ ] **Step 1: 创建 .gitignore**

```
# 敏感配置文件
configs/config.toml

# 数据库
*.db
.history/*.db

# 编译产物
agent-tui
*.exe
*.exe~
```

- [ ] **Step 2: 更新 config.toml**

```toml
# configs/config.toml

[models.openai]
api_key = ""
base_url = "https://api.openai.com/v1"
default_model = "gpt-4"

[models.deepseek]
api_key = ""
base_url = "https://api.deepseek.com"
default_model = "deepseek-chat"

[models.local]
base_url = "http://localhost:11434"
default_model = "llama3"

[models.volcengine]
api_key = "ark-6447fabd-7d92-47e4-b502-1955a7a3fa2d-fdefa"
base_url = "https://ark.cn-beijing.volces.com/api/coding"
default_model = "glm-5.1"

default_client = "volcengine"
theme = "dark"
approval_mode = "manual"
max_subagents = 3
```

- [ ] **Step 3: 更新 config.example.toml**

```toml
# configs/config.example.toml

[models.openai]
api_key = "your-openai-key-here"
base_url = "https://api.openai.com/v1"
default_model = "gpt-4"

[models.deepseek]
api_key = "your-deepseek-key-here"
base_url = "https://api.deepseek.com"
default_model = "deepseek-chat"

[models.local]
base_url = "http://localhost:11434"
default_model = "llama3"

[models.volcengine]
api_key = "your-volcengine-key-here"
base_url = "https://ark.cn-beijing.volces.com/api/coding"
default_model = "glm-5.1"

default_client = "volcengine"
theme = "dark"
approval_mode = "manual"
max_subagents = 3
```

- [ ] **Step 4: 提交**

```bash
git add .gitignore configs/config.toml configs/config.example.toml
git commit -m "feat: add gitignore and update config files for security"
```

---

## Task 6: 集成测试

**Files:**
- Modify: `cmd/agent/main.go`

- [ ] **Step 1: 验证主程序集成**

```bash
go build -o agent-tui ./cmd/agent
```

- [ ] **Step 2: 测试环境变量覆盖**

```bash
# 设置环境变量测试
export VOLCENGINE_API_KEY="test-env-key"
export DEFAULT_CLIENT="openai"
./agent-tui
```

- [ ] **Step 3: 测试无环境变量回退**

```bash
# 不设置环境变量
unset VOLCENGINE_API_KEY
unset DEFAULT_CLIENT
./agent-tui
```

- [ ] **Step 4: 提交**

```bash
git add cmd/agent/main.go
git commit -m "test: verify config integration works correctly"
```

---

## 实施总结

| Task | 描述 | 预计改动文件 |
|------|------|-------------|
| 1 | 修改配置结构支持多模型 | 2 files |
| 2 | 创建环境变量覆盖逻辑 | 2 files |
| 3 | 创建火山引擎客户端 | 2 files |
| 4 | 更新客户端工厂函数 | 2 files |
| 5 | 创建 .gitignore 和配置文件 | 3 files |
| 6 | 集成测试 | 1 file |

**总计**: 约 12 个文件改动

---

## 自检清单

- [x] Spec 覆盖：所有设计需求都有对应任务
- [x] 占位符扫描：无 TBD/TODO
- [x] 类型一致性：ModelConfig、ModelsConfig 等结构在所有任务中保持一致
