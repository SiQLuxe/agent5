# TUI Redesign Design Spec

**Date**: 2026-05-27
**Status**: Draft
**Scope**: `internal/ui/` 全部子模块

---

## 1. Overview

重新设计 agent-tui 界面，采用极简现代风格。移除无用模块，聚焦聊天体验，增加多会话标签支持。

## 2. Layout

```
┌─────────────────────────────────┐
│ StatusBar                       │  ← 项目名 + 连接状态
├─────────────────────────────────┤
│                                 │
│ Chat Area                       │  ← 消息列表，可滚动
│                                 │
├─────────────────────────────────┤
│ ❯                               │  ← 极简输入行，仅光标
├─────────────────────────────────┤
│ 代码审查·main.go | 部署配置 | + │  ← 底部标签栏
└─────────────────────────────────┘
```

渲染顺序（从上到下）：StatusBar → ChatArea → InputLine → TabDock

## 3. Tab Dock

### 3.1 位置
底部（bottom dock），输入行下方。

### 3.2 样式
- **激活标签**：蓝色背景 `#569cd6` + 白色粗体 + 圆角 `border-radius:3px`
- **非激活标签**：灰色文字 `#858585`，无背景
- **分隔符**：`|` 灰色 `#444`
- **新建按钮**：`+` 绿色 `#4ec9b0` 粗体

### 3.3 动态标签名
标签名按优先级生成：
1. Slash 命令 → 取命令名（如 `/fix lint errors`）
2. AI 推断 → 首条消息发送后，AI 推断主题（如 `代码审查 · main.go`）
3. 首条消息摘要 → 截取前 20 字符
4. 默认 → `New Session`

### 3.4 快捷键
| 快捷键 | 功能 |
|--------|------|
| Ctrl+Tab | 下一个会话 |
| Ctrl+Shift+Tab | 上一个会话 |
| Ctrl+N | 新建会话 |
| Ctrl+W | 关闭当前会话 |
| Ctrl+1~9 | 跳到第 N 个会话 |
| Ctrl+R | 重命名当前会话 |

## 4. Input Line

极简设计：
- 仅显示 `❯` 提示符（绿色 `#4ec9b0` 粗体）
- 无 placeholder 文字
- 无提示信息
- 高度 28px，深色背景 `#1a1a1a`，顶部 1px 分隔线

## 5. Role Display

### 5.1 用户消息
- 仅头像 `👤` + 颜色（绿色 `#4ec9b0`）
- 无名字 badge
- 右侧灰色时间戳

### 5.2 AI 助手消息
- Badge：`🤖 Luxe` 绿色背景 `#28a745` + 白色粗体
- 名字固定为 "Luxe"（可配置）
- 右侧灰色时间戳

### 5.3 系统消息
- Badge：`⚙️ System` 紫色背景 `#9370db` + 白色粗体
- 右侧灰色时间戳

## 6. Thinking Display

### 6.1 折叠态
```
▶ thinking 847 chars · 3.2s
```
- `▶` 黄色 `#dcdcaa`
- `thinking` 黄色
- 字符数 + 耗时 灰色 `#858585`

### 6.2 展开态
```
▼ thinking 1,203 chars · 5.1s
┌─────────────────────────────────┐
│ 让我分析这段代码...              │  ← 深黄背景 #2a2a1a
│ 1. main 函数缺少错误处理         │  ← 左侧黄色竖线 #dcdcaa
│ 2. 变量命名不符合规范            │
└─────────────────────────────────┘
```

### 6.3 规则
- 点击或 `Ctrl+F` 切换折叠/展开
- 字符数实时更新
- 耗时在 thinking 完成后显示
- 内容区：深黄背景 `#2a2a1a` + 左侧 2px 黄色竖线 `#dcdcaa`

## 7. StatusBar

修复现有布局 bug：
- 左侧：项目名（蓝色粗体）
- 右侧：连接状态
- 使用 `lipgloss.JoinHorizontal(lipgloss.Left, left, right)` 而非 `Center`

## 8. Code Cleanup

### 8.1 删除模块
| 模块 | 原因 |
|------|------|
| `internal/ui/windows/editor.go` | 不再需要 Editor 窗口 |
| `internal/ui/windows/terminal.go` | 不再需要 Terminal 窗口 |
| `internal/ui/windows/logs.go` | 不再需要 Logs 窗口 |
| `internal/ui/chat/chat.go` | 重复实现，与 windows/chat.go 冗余 |
| `internal/ui/layout/layout.go` | 未使用的 Layout 模块 |
| `internal/ui/rightpanel/rightpanel.go` | 未使用的 RightPanel |
| `internal/ui/filetree/filetree.go` | 未使用的 FileTree |

### 8.2 删除测试
对应测试文件一并删除：
- `internal/ui/chat/chat_test.go`
- `internal/ui/layout/layout_test.go`
- `internal/ui/rightpanel/rightpanel_test.go`
- `internal/ui/filetree/filetree_test.go`
- `internal/ui/windows/editor_test.go`（如存在）
- `internal/ui/windows/terminal_test.go`（如存在）
- `internal/ui/windows/logs_test.go`（如存在）

### 8.3 WindowManager 简化
移除 WindowManager，UI Model 直接管理 ChatPanel + TabDock。

## 9. Keyboard Shortcuts Summary

| 快捷键 | 功能 |
|--------|------|
| Ctrl+Tab | 下一个会话 |
| Ctrl+Shift+Tab | 上一个会话 |
| Ctrl+N | 新建会话 |
| Ctrl+W | 关闭当前会话 |
| Ctrl+1~9 | 跳到第 N 个会话 |
| Ctrl+R | 重命名当前会话 |
| Ctrl+F | 折叠/展开 thinking |
| Ctrl+C | 退出 |
| Ctrl+T | 切换主题 |
| Enter | 发送消息 |
| Backspace | 空输入时关闭会话（可选） |

## 10. Color Palette

| 用途 | 颜色 | Hex |
|------|------|-----|
| 激活标签背景 | 蓝色 | `#569cd6` |
| 用户头像 | 绿色 | `#4ec9b0` |
| AI badge 背景 | 绿色 | `#28a745` |
| System badge 背景 | 紫色 | `#9370db` |
| Thinking 箭头/文字 | 黄色 | `#dcdcaa` |
| Thinking 内容背景 | 深黄 | `#2a2a1a` |
| 非激活标签文字 | 灰色 | `#858585` |
| 分隔符 | 深灰 | `#444` |
| 输入提示符 | 绿色 | `#4ec9b0` |
| 主背景 | 深色 | `#1e1e1e` |
| 输入/标签栏背景 | 深色 | `#1a1a1a` |
| 正文文字 | 浅灰 | `#d4d4d4` |
| 时间戳 | 灰色 | `#555` |

## 11. Data Model Changes

### 11.1 Session
```go
type Session struct {
    ID        string
    Label     string    // 动态标签名
    Messages  []Message
    CreatedAt time.Time
}
```

### 11.2 Message
```go
type Message struct {
    Role       Role      // User / Assistant / System
    Content    string
    Thinking   *Thinking // nil if no thinking
    Timestamp  time.Time
}

type Thinking struct {
    Content  string
    Duration time.Duration
    Expanded bool
}
```

### 11.3 Role
```go
type Role int

const (
    RoleUser      Role = iota
    RoleAssistant
    RoleSystem
)
```

## 12. Out of Scope

- 主题编辑器 UI（仅 Ctrl+T 切换预设主题）
- 消息搜索
- 消息编辑/删除
- 文件拖拽上传
- 多行输入模式
- 会话持久化到磁盘
