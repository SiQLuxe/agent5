# Agent-TUI

一个功能强大的 AI 助手终端用户界面，支持代码编辑器、多 AI 模型集成、项目管理和插件系统。

## 功能特性

- 🤖 **多 AI 模型支持**: OpenAI、DeepSeek、Anthropic Claude 及本地模型
- 📝 **代码编辑器**: 内置代码编辑器，支持语法高亮
- 📁 **文件树**: 项目文件浏览和管理
- 💬 **聊天面板**: 与 AI 模型进行对话
- 🛠️ **调试工具**: 执行命令和调试代码
- 🔌 **插件系统**: 支持扩展功能

## 技术栈

- Go 1.22+
- Bubble Tea (TUI 框架)
- Lipgloss (样式)
- Chroma (代码高亮)

## 快速开始

### 安装

```bash
go install github.com/example/agent-tui/cmd/agent@latest
```

### 构建

```bash
git clone https://github.com/example/agent-tui.git
cd agent-tui
go build -o agent-tui ./cmd/agent
```

### 运行

```bash
./agent-tui
```

## 快捷键

| 快捷键 | 功能 |
|--------|------|
| `Ctrl+C` | 退出程序 |
| `Ctrl+E` | 切换聊天/编辑模式 |
| `Ctrl+B` | 显示/隐藏侧边栏 |

## 配置

配置文件位于 `~/.agent-tui/config.toml`:

```toml
[api_keys]
openai = "your-api-key"
deepseek = "your-api-key"
anthropic = "your-api-key"

default_model = "gpt-4"
theme = "dark"
approval_mode = "auto"
max_subagents = 5
```

## 支持的 AI 模型

- OpenAI: gpt-4, gpt-4-turbo, gpt-3.5-turbo
- DeepSeek: deepseek-chat, deepseek-code, deepseek-r1
- Anthropic: claude-3-opus, claude-3-sonnet, claude-3-haiku
- 本地模型 (OpenAI 兼容 API)

## 项目结构

```
agent-tui/
├── cmd/
│   └── agent/
│       └── main.go
├── internal/
│   ├── ai/           # AI 客户端
│   ├── data/         # 数据存储
│   ├── service/      # 业务服务
│   └── ui/           # UI 组件
├── docs/             # 文档
└── go.mod
```

## 许可证

MIT License