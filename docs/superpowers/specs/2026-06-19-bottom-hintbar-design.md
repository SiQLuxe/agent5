# Bottom HintBar — Design

Date: 2026-06-19
Status: Approved

## 目标

把临时提示（如「再次按 Ctrl+C 退出」）从顶部 `statusBar` 迁到屏幕**最底部**的独立行控件 `HintBar`，对齐 vim/lazygit/k9s 等主流 TUI。同时为后续"操作反馈/双击确认/瞬时错误"提供通用入口。

## 背景

当前 `fireCtrlCHint` 把提示写到 `statusBar`（顶部），覆盖 mode/tasks/connection 持久信息，且远离用户视线焦点（composer 在底部）。业界主流 TUI 把临时反馈放在屏幕最下方独立一行。

## 范围

**改动**
- 新增 `internal/ui/hint/hint.go` + 对应单测
- `internal/ui/app.go`：加字段 `hintBar`、布局插入最底行、`fireCtrlCHint` 改用 `hintBar`
- `internal/ui/app_test.go`：调整 Ctrl+C 测试断言对象

**不改**
- `statusBar` 行为（持续信息）。
- `ShowToast`（HTTP 路径），仍走 statusBar，本次不动。
- selectable_textview、tabDock、composer、suggestionMenu。
- SIGINT 兜底、信号层逻辑。

## 布局变化

```
statusBar       (1 行, 顶)
chatPanel       (flex, 主体)
suggestionMenu  (1 行)
composer        (3 行)
tabDock         (1 行)
hintBar         (1 行, 最底)   ← 新增
```

固定占 1 行；平时空白；提示文本写入即显示，调用方用 timer 清除。

## 行为

| 场景 | hintBar 内容 |
|---|---|
| 启动后无操作 | 空白 |
| 首次 Ctrl+C（无选中） | `[gray]再次按 Ctrl+C 退出[-]` |
| 3 秒后无操作 | 自动清空 |
| 2 秒内再次 Ctrl+C | 退出（hintBar 内容无关） |
| 后续接入复制成功等 | 同 API，自定义 level/dur |

## API

```go
// internal/ui/hint/hint.go
type Level int
const (
    LevelInfo Level = iota
    LevelWarn
    LevelError
)

type HintBar struct { *tview.TextView }
func New() *HintBar
func (h *HintBar) Show(text string, level Level)  // 同步设置文本
func (h *HintBar) Clear()
```

颜色：
- `LevelInfo`  → `[gray]`
- `LevelWarn`  → `[yellow]`
- `LevelError` → `[red]`

App 侧通用入口（备用，本次仅 Ctrl+C 用）：
```go
func (a *App) ShowHint(text string, level hint.Level, dur time.Duration)
```

## Ctrl+C 路径

```go
func (a *App) fireCtrlCHint() {
    if a.ctrlCHint != nil {
        a.ctrlCHint()
        return
    }
    a.hintBar.Show("再次按 Ctrl+C 退出", hint.LevelInfo)
    go func() {
        time.Sleep(3 * time.Second)
        a.QueueUpdateDraw(func() { a.hintBar.Clear() })
    }()
}
```

⚠️ `Show` 同步调用——`fireCtrlCHint` 已在 tview 主事件循环 goroutine（inputCapture 内），不能用 QueueUpdateDraw（会自死锁，详见上一份 spec）。3s 清除 timer 在独立 goroutine，QueueUpdateDraw 安全。

## 并发与重入

连续触发 Ctrl+C：
- 第二次按落入"2s 退出"分支，不会再 `Show`，问题不存在。
- 后续若复用该 API 让其它键短时间内多次触发：每次 `Show` 直接覆盖 `SetText`，最早 timer 触发时 `Clear` 也只是 `SetText("")`——可能把仍在显示的新提示提前清除。本次不解决（YAGNI），后续接入更多调用点时再考虑"令牌/版本号"机制。

## 测试

### `internal/ui/hint/hint_test.go`（新增）
- `New()` 后 `GetText` 为空。
- `Show("hello", LevelInfo)` 后 `GetText(true)`（stripped）含 `"hello"`，原文含 `[gray]` 标签。
- `LevelWarn` 含 `[yellow]`、`LevelError` 含 `[red]`。
- `Clear` 后 `GetText` 为空。

### `internal/ui/app_test.go`（调整）
- `TestCtrlC_*` 现有 4 个测试无需改：仍用 `ctrlCHint` 注入跳过真实显示路径。
- 加 `TestCtrlC_DefaultHintBarPathSetsText`：
  - 不注入 `ctrlCHint`。
  - 直接调 `a.fireCtrlCHint()`（绕过 inputCapture，避免触发 timer 风险）。
  - 断言 `a.hintBar.GetText(true)` 含「再次按 Ctrl+C 退出」。
  - 不等清除 timer（goroutine 泄漏可接受，单测进程退出收尾）。

## 风险

- **多 1 行屏幕高度**：~4% 可见区，可接受。
- **timer goroutine 泄漏**（清除时 QueueUpdateDraw 阻塞于未启动的 Application）：仅出现在测试中。生产 Application 一直 Run，无问题。

## 不在范围

- 把 `ShowToast` 迁到 hintBar（后续单独评估）。
- 接入复制成功/双击删除/错误反馈（后续按需）。
- 基于令牌的提示并发治理（YAGNI）。
- HintBar 高度自适应（A 方案：永远占 1 行）。
