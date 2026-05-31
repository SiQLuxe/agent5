git add configs/config.toml
git commit -m "chore: document prefixed default_model format"
git add internal/data/config/config.go internal/data/config/config_test.go
git commit -m "feat(config): parse default_model client:model prefix"
git add internal/ai/factory.go internal/ai/factory_test.go
git commit -m "feat(ai): add NewClientFromConfig factory to pick client by prefix"
git add cmd/agent/main.go
git commit -m "feat(cmd): use NewClientFromConfig factory for AI client selection"
# Dynamic Model Selection Implementation Plan

> **对于自动化执行者：** 必需子技能：使用 `superpowers:subagent-driven-development`（推荐）或 `superpowers:executing-plans` 按步骤实现。步骤使用复选框（`- [ ]`）语法跟踪。

**目标：** 通过配置项 `default_model` 使用前缀格式 `<client>:<model>` 实现动态选择 AI 客户端（支持 `openai`, `deepseek`, `anthropic`, `local`），并在远程客户端缺少 API Key 时立即失败（快速报错）。

**架构：** 在配置加载后解析 `default_model` 的前缀，前缀映射到配置字段（`openai` → `api_keys.openai`，`deepseek` → `api_keys.deepseek`，`anthropic` → `api_keys.anthropic`，`local` → `auth_token` + `base_url`）。新增 `internal/ai/factory.go` 提供 `NewClientFromConfig(cfg *config.Config) (ai.Client, error)`，负责解析、校验并构造对应客户端。`cmd/agent/main.go` 使用该工厂函数创建 AI 客户端。

**技术栈：** Go 标准库（`net/http`, `fmt`, `errors`, `testing`），使用 `net/http/httptest` 做测试；可选 `github.com/stretchr/testify` 简化断言。

---

## 需要创建/修改的文件

- 修改：`configs/config.toml`（示例更新）
- 修改：`internal/data/config/config.go`（添加前缀解析辅助函数）
- 新建：`internal/ai/factory.go`（工厂：解析前缀、校验 API Key、构造客户端）
- 修改：`cmd/agent/main.go`（使用工厂替换直接创建 LocalClient 的逻辑）
- 新建：`internal/ai/factory_test.go`（单元测试：解析、映射、错误场景）

---

## 实施步骤（TDD 风格，细粒度）

### 任务 1：更新示例配置

**文件：** 修改 `configs/config.toml`

- [ ] 步骤 1：编辑 `configs/config.toml`，示例写明前缀格式

```toml
[api_keys]
openai = ""
deepseek = ""
anthropic = ""

# 格式：<client>:<model>
default_model = "openai:gpt-4"

theme = "dark"
approval_mode = "manual"
max_subagents = 3
auth_token = "test-api-key-123"
base_url = "http://localhost:8080"
```

- [ ] 步骤 2：提交更改

```bash
git add configs/config.toml
git commit -m "chore: document prefixed default_model format"
```

---

### 任务 2：在配置包中添加解析函数

**文件：** 修改 `internal/data/config/config.go`

- [ ] 步骤 1：编写单元测试 `internal/data/config/config_test.go` 验证前缀解析（表格驱动测试）

测试示例：

```go
func TestParseModelPrefix(t *testing.T) {
    cases := []struct{ in, wantClient, wantModel string }{
        {"openai:gpt-4", "openai", "gpt-4"},
        {"local:kimi-k2.6", "local", "kimi-k2.6"},
    }
    for _, c := range cases {
        cl, md, err := ParseModelPrefix(c.in)
        if err != nil { t.Fatalf("unexpected err: %v", err) }
        if cl != c.wantClient || md != c.wantModel { t.Fatalf("mismatch: got %s:%s", cl, md) }
    }
}
```

- [ ] 步骤 2：实现 `ParseModelPrefix(s string) (client, model string, err error)`：
  - 使用 `strings.SplitN(s, ":", 2)`，必须得到两部分且非空
  - 允许前缀：`openai`, `deepseek`, `anthropic`, `local`
  - 不满足格式或未知前缀返回错误

- [ ] 步骤 3：运行该包的测试

```bash
cd internal/data && go test -v ./...
```

- [ ] 步骤 4：提交

```bash
git add internal/data/config/config.go internal/data/config/config_test.go
git commit -m "feat(config): parse default_model client:model prefix"
```

---

### 任务 3：实现 AI 客户端工厂

**文件：** 新建 `internal/ai/factory.go` 与 `internal/ai/factory_test.go`

- [ ] 步骤 1：编写 `NewClientFromConfig(cfg *config.Config) (ai.Client, error)` 的表格驱动单元测试：
  - 场景：`openai:gpt-4` 且 `cfg.APIKeys.OpenAI` 存在 → 返回 `OpenAIClient`
  - 场景：`openai:gpt-4` 且 `cfg.APIKeys.OpenAI` 为空 → 返回错误（立即失败）
  - 场景：`local:kimi-k2.6` → 返回 `LocalClient`（使用 `cfg.AuthToken` 与 `cfg.BaseURL`）
  - 场景：未知前缀 → 返回错误

测试样例：

```go
func TestNewClientFromConfig(t *testing.T) {
    cfg := &config.Config{ APIKeys: config.APIKeys{OpenAI: "key"}, DefaultModel: "openai:gpt-4" }
    c, err := NewClientFromConfig(cfg)
    if err != nil || c == nil { t.Fatalf("expected client, got err=%v", err) }
}
```

- [ ] 步骤 2：在 `factory.go` 中实现逻辑：

```go
func NewClientFromConfig(cfg *config.Config) (Client, error) {
    prefix, model, err := config.ParseModelPrefix(cfg.DefaultModel)
    if err != nil { return nil, err }
    switch prefix {
    case "openai":
        if cfg.APIKeys.OpenAI == "" { return nil, fmt.Errorf("openai api key not set") }
        return NewOpenAIClient(cfg.APIKeys.OpenAI, "", model)
    case "deepseek":
        if cfg.APIKeys.DeepSeek == "" { return nil, fmt.Errorf("deepseek api key not set") }
        return NewDeepSeekClient(cfg.APIKeys.DeepSeek, "", model)
    case "anthropic":
        if cfg.APIKeys.Anthropic == "" { return nil, fmt.Errorf("anthropic api key not set") }
        return NewAnthropicClient(cfg.APIKeys.Anthropic, "", model)
    case "local":
        return NewLocalClient(cfg.AuthToken, cfg.BaseURL, model)
    default:
        return nil, fmt.Errorf("unsupported client: %s", prefix)
    }
}
```

- [ ] 步骤 3：运行 `internal/ai` 的测试

```bash
cd internal/ai && go test -v ./...
```

- [ ] 步骤 4：提交

```bash
git add internal/ai/factory.go internal/ai/factory_test.go
git commit -m "feat(ai): add NewClientFromConfig factory to pick client by prefix"
```

---

### 任务 4：更新应用启动流程

**文件：** 修改 `cmd/agent/main.go`

- [ ] 步骤 1：将原先写死的 `NewLocalClient` 替换为 `ai.NewClientFromConfig(cfg)`：

替换前：

```go
aiClient, _ := ai.NewLocalClient(cfg.AuthToken, cfg.BaseURL, cfg.DefaultModel)
```

替换后：

```go
aiClient, err := ai.NewClientFromConfig(cfg)
if err != nil {
    log.Fatalf("failed to create AI client: %v", err)
}
```

- [ ] 步骤 2：构建并做快速 smoke test

```bash
go build ./cmd/agent
./agent-tui  # 或使用相应的可执行文件名
```

预期：当配置为 `openai:...` 且没有 `api_keys.openai` 时，程序在启动阶段以清晰错误退出（快速失败）。

- [ ] 步骤 3：提交

```bash
git add cmd/agent/main.go
git commit -m "feat(cmd): use NewClientFromConfig factory for AI client selection"
```

---

### 任务 5：补充单元测试（错误与映射）

**文件：** `internal/ai/factory_test.go`（或补充现有测试）

- [ ] 步骤 1：添加丢失 API Key 的测试用例

示例：

```go
func TestFactory_MissingAPIKey(t *testing.T) {
    cfg := &config.Config{ DefaultModel: "openai:gpt-4", APIKeys: config.APIKeys{OpenAI: ""} }
    _, err := NewClientFromConfig(cfg)
    if err == nil { t.Fatal("expected error when openai key missing") }
}
```

- [ ] 步骤 2：运行所有 AI 包测试

```bash
cd internal/ai && go test -v ./...
```

- [ ] 步骤 3：提交测试更改

```bash
git add internal/ai/factory_test.go
git commit -m "test(ai): add missing-key and mapping tests for NewClientFromConfig"
```

---

### 任务 6：更新文档示例

- [ ] 步骤 1：在 README 或 `configs/config.toml` 示例中加入 `openai:gpt-4` 的说明
- [ ] 步骤 2：提交

---

## 验收标准

- `default_model` 正确解析 `<client>:<model>` 格式
- 当使用远程客户端前缀且对应 API Key 缺失时，应用启动阶段立即失败并输出明确错误
- `local:...` 前缀仍然可用，使用 `auth_token` 与 `base_url`
- 单元测试覆盖解析、映射和缺失 Key 的行为
- 代码改动最小且有测试保障

---

保存路径：`docs/superpowers/plans/2026-05-22-dynamic-model-selection-implementation.md`
