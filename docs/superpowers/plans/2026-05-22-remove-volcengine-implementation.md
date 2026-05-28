# 移除火山引擎客户端实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除火山引擎客户端及相关配置，简化架构

**Architecture:** 删除 volcengine.go 和相关测试，修改配置结构移除 VolcEngine 字段，更新工厂函数

**Tech Stack:** Go, TOML

---

## 文件结构

```
删除:
- internal/ai/volcengine.go
- internal/ai/volcengine_test.go

修改:
- internal/ai/factory.go           # 移除 volcengine case
- internal/ai/factory_test.go     # 移除 volcengine 测试
- internal/data/config/config.go   # 移除 VolcEngine 字段
- internal/data/config/env.go      # 移除 VOLCENGINE_* 环境变量
- configs/config.toml              # 移除 volcengine 配置
- configs/config.example.toml      # 移除 volcengine 配置
```

---

## Task 1: 删除火山引擎客户端文件

**Files:**
- Delete: `internal/ai/volcengine.go`
- Delete: `internal/ai/volcengine_test.go`

- [ ] **Step 1: 删除 volcengine.go**

```bash
rm internal/ai/volcengine.go
```

- [ ] **Step 2: 删除 volcengine_test.go**

```bash
rm internal/ai/volcengine_test.go
```

- [ ] **Step 3: 验证文件已删除**

```bash
ls internal/ai/volcengine*.go 2>&1 || echo "Files deleted"
```

- [ ] **Step 4: 提交**

```bash
git add -A
git commit -m "refactor: remove volcengine client (use openai client with custom base_url)"
```

---

## Task 2: 修改工厂函数

**Files:**
- Modify: `internal/ai/factory.go`
- Modify: `internal/ai/factory_test.go`

- [ ] **Step 1: 读取并修改 factory.go**

修改 switch 语句，移除 volcengine case：

```go
switch clientType {
case "openai":
    apiKey := c.Models.OpenAI.APIKey
    if apiKey == "" {
        return nil, fmt.Errorf("openai api key not set")
    }
    baseURL := c.Models.OpenAI.BaseURL
    model := c.Models.OpenAI.DefaultModel
    return NewOpenAIClient(apiKey, baseURL, model)
case "deepseek":
    apiKey := c.Models.DeepSeek.APIKey
    if apiKey == "" {
        return nil, fmt.Errorf("deepseek api key not set")
    }
    baseURL := c.Models.DeepSeek.BaseURL
    model := c.Models.DeepSeek.DefaultModel
    return NewDeepSeekClient(apiKey, baseURL, model)
case "local":
    baseURL := c.Models.Local.BaseURL
    model := c.Models.Local.DefaultModel
    return NewLocalClient("", baseURL, model)
default:
    return nil, fmt.Errorf("unsupported client: %s", clientType)
}
```

- [ ] **Step 2: 修改 factory_test.go**

移除所有 volcengine 相关测试函数：
- `TestNewClientFromConfigVolcEngine`
- `TestVolcEngineGetModels`

- [ ] **Step 3: 运行测试验证**

```bash
go test -v ./internal/ai/...
```

- [ ] **Step 4: 提交**

```bash
git add internal/ai/factory.go internal/ai/factory_test.go
git commit -m "refactor: remove volcengine from factory"
```

---

## Task 3: 修改配置结构

**Files:**
- Modify: `internal/data/config/config.go`
- Modify: `internal/data/config/config_test.go`

- [ ] **Step 1: 修改 config.go**

移除 ModelsConfig 中的 VolcEngine 字段：

```go
type ModelsConfig struct {
    OpenAI   ModelConfig `toml:"openai"`
    DeepSeek ModelConfig `toml:"deepseek"`
    Local    ModelConfig `toml:"local"`
}
```

更新 GetDefaultConfig() 移除 volcengine 默认配置：

```go
func GetDefaultConfig() *Config {
    return &Config{
        Models: ModelsConfig{
            OpenAI: ModelConfig{
                APIKey:       "",
                BaseURL:      "https://api.openai.com/v1",
                DefaultModel: "gpt-4",
            },
            DeepSeek: ModelConfig{
                APIKey:       "",
                BaseURL:      "https://api.deepseek.com",
                DefaultModel: "deepseek-chat",
            },
            Local: ModelConfig{
                BaseURL:      "http://localhost:11434",
                DefaultModel: "llama3",
            },
        },
        DefaultClient: "openai",
        Theme:         "dark",
        ApprovalMode:  "manual",
        MaxSubagents:  3,
    }
}
```

- [ ] **Step 2: 修改 config_test.go**

移除 VolcEngine 相关测试：
- `TestConfigStructure` 中的 volcengine 测试

- [ ] **Step 3: 运行测试验证**

```bash
go test -v ./internal/data/config/...
```

- [ ] **Step 4: 提交**

```bash
git add internal/data/config/config.go internal/data/config/config_test.go
git commit -m "refactor: remove volcengine from config structure"
```

---

## Task 4: 修改环境变量覆盖逻辑

**Files:**
- Modify: `internal/data/config/env.go`
- Modify: `internal/data/config/env_test.go`

- [ ] **Step 1: 修改 env.go**

移除所有 VOLCENGINE_* 环境变量处理：

```go
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

    // Default client
    if val := os.Getenv("DEFAULT_CLIENT"); val != "" {
        cfg.DefaultClient = val
    }
}
```

- [ ] **Step 2: 修改 env_test.go**

移除 `TestEnvOverridesVolcEngine` 函数

- [ ] **Step 3: 运行测试验证**

```bash
go test -v ./internal/data/config/...
```

- [ ] **Step 4: 提交**

```bash
git add internal/data/config/env.go internal/data/config/env_test.go
git commit -m "refactor: remove volcengine environment variables"
```

---

## Task 5: 更新配置文件

**Files:**
- Modify: `configs/config.toml`
- Modify: `configs/config.example.toml`

- [ ] **Step 1: 更新 config.toml**

```toml
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

default_client = "openai"
theme = "dark"
approval_mode = "manual"
max_subagents = 3
```

- [ ] **Step 2: 更新 config.example.toml**

```toml
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

default_client = "openai"
theme = "dark"
approval_mode = "manual"
max_subagents = 3
```

- [ ] **Step 3: 提交**

```bash
git add configs/config.toml configs/config.example.toml
git commit -m "refactor: remove volcengine from config files"
```

---

## Task 6: 集成测试

**Files:**
- Modify: `cmd/agent/main.go` (如有引用)

- [ ] **Step 1: 编译测试**

```bash
go build -o agent-tui ./cmd/agent
```

- [ ] **Step 2: 运行所有测试**

```bash
go test -v ./...
```

- [ ] **Step 3: 验证编译产物**

```bash
ls -la agent-tui
```

- [ ] **Step 4: 提交**

```bash
git add -A
git commit -m "test: verify volcengine removal integration"
```

---

## 实施总结

| Task | 描述 | 预计改动 |
|------|------|---------|
| 1 | 删除火山引擎客户端文件 | 删除 2 files |
| 2 | 修改工厂函数 | 修改 2 files |
| 3 | 修改配置结构 | 修改 2 files |
| 4 | 修改环境变量逻辑 | 修改 2 files |
| 5 | 更新配置文件 | 修改 2 files |
| 6 | 集成测试 | 测试验证 |

**总计**: 删除 2 files，修改 8 files

---

## 自检清单

- [x] Spec 覆盖：所有删除和修改需求都有对应任务
- [x] 占位符扫描：无 TBD/TODO
- [x] 类型一致性：ModelsConfig 只保留 OpenAI/DeepSeek/Local
