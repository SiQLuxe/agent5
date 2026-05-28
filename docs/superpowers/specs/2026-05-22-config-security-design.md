# 配置文件密钥安全管理设计

**文档版本**: v1.0
**创建日期**: 2026-05-22
**作者**: Agent Designer

---

## 1. 需求概述

为 Agent TUI 项目实现安全的密钥管理机制，支持多模型配置，并通过环境变量实现密钥外部化。

### 1.1 目标

- 敏感信息（API Key、Base URL）不硬编码在代码或配置模板中
- 支持多模型独立配置（OpenAI、DeepSeek、本地模型等）
- 环境变量优先级高于配置文件，便于开发和部署
- 防止敏感配置被提交到 Git

---

## 2. 配置结构设计

### 2.1 配置文件格式 (TOML)

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
api_key = ""
base_url = "https://ark.cn-beijing.volces.com/api/coding"
default_model = "glm-5.1"

# 全局默认设置
default_client = "volcengine"
theme = "dark"
approval_mode = "manual"
max_subagents = 3
```

### 2.2 环境变量

环境变量会覆盖配置文件中的对应值，命名规范采用简单直接风格：

| 环境变量 | 覆盖字段 | 说明 |
|---------|---------|------|
| `OPENAI_API_KEY` | `models.openai.api_key` | OpenAI API 密钥 |
| `OPENAI_BASE_URL` | `models.openai.base_url` | OpenAI API 端点 |
| `OPENAI_DEFAULT_MODEL` | `models.openai.default_model` | OpenAI 默认模型 |
| `DEEPSEEK_API_KEY` | `models.deepseek.api_key` | DeepSeek API 密钥 |
| `DEEPSEEK_BASE_URL` | `models.deepseek.base_url` | DeepSeek API 端点 |
| `DEEPSEEK_DEFAULT_MODEL` | `models.deepseek.default_model` | DeepSeek 默认模型 |
| `LOCAL_BASE_URL` | `models.local.base_url` | 本地模型端点 |
| `LOCAL_DEFAULT_MODEL` | `models.local.default_model` | 本地默认模型 |
| `VOLCENGINE_API_KEY` | `models.volcengine.api_key` | 火山引擎 API 密钥 |
| `VOLCENGINE_BASE_URL` | `models.volcengine.base_url` | 火山引擎端点 |
| `VOLCENGINE_DEFAULT_MODEL` | `models.volcengine.default_model` | 火山引擎默认模型 |
| `DEFAULT_CLIENT` | `default_client` | 默认使用的客户端 |

### 2.3 配置加载优先级

```
环境变量 > configs/config.toml > 默认值
```

---

## 3. 代码结构设计

### 3.1 配置文件加载逻辑

```go
// internal/data/config/config.go

type Config struct {
    Models      ModelsConfig `toml:"models"`
    DefaultClient string     `toml:"default_client"`
    Theme        string     `toml:"theme"`
    ApprovalMode string     `toml:"approval_mode"`
    MaxSubagents int        `toml:"max_subagents"`
}

type ModelsConfig struct {
    OpenAI    ModelConfig `toml:"openai"`
    DeepSeek  ModelConfig `toml:"deepseek"`
    Local     ModelConfig `toml:"local"`
    VolcEngine ModelConfig `toml:"volcengine"`
}

type ModelConfig struct {
    APIKey       string `toml:"api_key"`
    BaseURL      string `toml:"base_url"`
    DefaultModel string `toml:"default_model"`
}

// LoadConfig 从文件加载配置，然后应用环境变量覆盖
func LoadConfig(path string) (*Config, error) {
    // 1. 从 TOML 文件加载基础配置
    // 2. 调用 applyEnvOverrides() 应用环境变量覆盖
    // 3. 返回合并后的配置
}

func applyEnvOverrides(cfg *Config) {
    // 检查并应用每个环境变量
    if val := os.Getenv("OPENAI_API_KEY"); val != "" {
        cfg.Models.OpenAI.APIKey = val
    }
    // ... 其他环境变量
}
```

### 3.2 客户端工厂函数

```go
// internal/ai/factory.go

// NewClientFromConfig 根据配置创建 AI 客户端
func NewClientFromConfig(cfg *config.Config) (ai.AIClient, error) {
    defaultClient := cfg.DefaultClient
    if envClient := os.Getenv("DEFAULT_CLIENT"); envClient != "" {
        defaultClient = envClient
    }

    switch defaultClient {
    case "openai":
        return openai.NewClient(cfg.Models.OpenAI.APIKey, cfg.Models.OpenAI.BaseURL)
    case "deepseek":
        return deepseek.NewClient(cfg.Models.DeepSeek.APIKey, cfg.Models.DeepSeek.BaseURL)
    case "local":
        return local.NewClient(cfg.Models.Local.BaseURL)
    case "volcengine":
        return volcengine.NewClient(cfg.Models.VolcEngine.APIKey, cfg.Models.VolcEngine.BaseURL)
    default:
        return nil, fmt.Errorf("unknown client: %s", defaultClient)
    }
}
```

---

## 4. .gitignore 规则

```gitignore
# 敏感配置文件 - 只提交模板
configs/config.toml

# 本地数据库
*.db
.history/*.db

# 编译产物
agent-tui
*.exe
```

### 4.1 配置模板文件

`configs/config.example.toml` 作为模板提交到 Git，不包含真实密钥：

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

[models.volcengine]
api_key = "your-volcengine-key-here"
base_url = "https://ark.cn-beijing.volces.com/api/coding"
default_model = "glm-5.1"

default_client = "volcengine"
theme = "dark"
approval_mode = "manual"
max_subagents = 3
```

---

## 5. 实施步骤

1. 修改 `internal/data/config/config.go` 添加多模型配置结构
2. 实现 `applyEnvOverrides()` 函数支持环境变量覆盖
3. 创建 `internal/ai/volcengine/` 模块（如果不存在）
4. 更新客户端工厂函数支持新配置格式
5. 创建 `.gitignore` 文件
6. 更新配置文件和模板

---

## 6. 测试验证

### 6.1 环境变量覆盖测试

```bash
# 设置环境变量
export OPENAI_API_KEY="sk-test-key"
export DEFAULT_CLIENT="openai"

# 运行程序，验证使用的是环境变量中的值
./agent-tui
```

### 6.2 配置文件测试

```bash
# 不设置环境变量，使用配置文件中的值
./agent-tui
```

---

## 7. 安全性考虑

- API Key 不会在日志中打印
- 错误信息中不会包含完整的 Key 值
- 支持通过环境变量注入密钥，适合容器化部署
- 配置文件不提交到 Git，防止密钥泄露
