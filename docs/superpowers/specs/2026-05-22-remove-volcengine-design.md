# 移除火山引擎客户端设计

**文档版本**: v1.0
**创建日期**: 2026-05-22
**作者**: Agent Designer

---

## 1. 需求概述

移除火山引擎客户端，简化架构。国产模型通过 OpenAI 客户端连接（修改 base_url）。

### 1.1 背景

火山引擎等国产模型 API 兼容 OpenAI 协议，无需单独实现客户端。当前 `volcengine.go` 与 `openai.go` 功能重复。

### 1.2 目标

- 移除冗余的火山引擎客户端代码
- 简化配置结构
- 减少维护成本

---

## 2. 改动范围

### 2.1 删除的文件

| 文件 | 说明 |
|------|------|
| `internal/ai/volcengine.go` | 火山引擎客户端实现 |
| `internal/ai/volcengine_test.go` | 火山引擎客户端测试 |

### 2.2 修改的文件

| 文件 | 改动内容 |
|------|---------|
| `internal/ai/factory.go` | 移除 volcengine case |
| `internal/ai/factory_test.go` | 移除 volcengine 相关测试 |
| `internal/data/config/config.go` | 移除 VolcEngine 字段，更新默认值 |
| `internal/data/config/config_test.go` | 移除 VolcEngine 相关测试 |
| `internal/data/config/env.go` | 移除 VOLCENGINE_* 环境变量 |
| `internal/data/config/env_test.go` | 移除 VolcEngine 环境变量测试 |
| `configs/config.toml` | 移除 volcengine 配置 |
| `configs/config.example.toml` | 移除 volcengine 配置 |

---

## 3. 配置结构（简化后）

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

default_client = "openai"
theme = "dark"
approval_mode = "manual"
max_subagents = 3
```

---

## 4. 代码改动

### 4.1 ModelsConfig 结构

```go
// internal/data/config/config.go

type ModelsConfig struct {
    OpenAI   ModelConfig `toml:"openai"`
    DeepSeek ModelConfig `toml:"deepseek"`
    Local    ModelConfig `toml:"local"`
}
```

### 4.2 工厂函数

```go
// internal/ai/factory.go

func NewClientFromConfig(c *cfg.Config) (Client, error) {
    switch c.DefaultClient {
    case "openai":
        return NewOpenAIClient(...)
    case "deepseek":
        return NewDeepSeekClient(...)
    case "local":
        return NewLocalClient(...)
    default:
        return nil, fmt.Errorf("unsupported client: %s", c.DefaultClient)
    }
}
```

### 4.3 环境变量

移除以下环境变量：
- `VOLCENGINE_API_KEY`
- `VOLCENGINE_BASE_URL`
- `VOLCENGINE_DEFAULT_MODEL`

---

## 5. 使用火山引擎的方式

通过 OpenAI 客户端连接火山引擎：

```toml
[models.openai]
api_key = "ark-xxx"
base_url = "https://ark.cn-beijing.volces.com/api/coding"
default_model = "glm-5.1"
default_client = "openai"
```

或通过环境变量：

```bash
export OPENAI_API_KEY="ark-xxx"
export OPENAI_BASE_URL="https://ark.cn-beijing.volces.com/api/coding"
export OPENAI_DEFAULT_MODEL="glm-5.1"
export DEFAULT_CLIENT="openai"
```

---

## 6. 实施步骤

1. 删除 `internal/ai/volcengine.go` 和 `volcengine_test.go`
2. 修改 `internal/ai/factory.go` 移除 volcengine case
3. 修改 `internal/data/config/config.go` 移除 VolcEngine 字段
4. 修改 `internal/data/config/env.go` 移除 VOLCENGINE 环境变量
5. 更新配置文件
6. 运行测试验证
