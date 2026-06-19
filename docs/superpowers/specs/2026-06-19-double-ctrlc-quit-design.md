# Double Ctrl+C to Quit — Design

Date: 2026-06-19
Status: Approved

## 目标

把 TUI 中"单次 Ctrl+C 直接退出"改为"两次 Ctrl+C 退出"。首次按下仅提示，2 秒内再次按下才退出。

## 背景

当前 `internal/ui/app.go` 中 `KeyCtrlC` 直接调用 `a.Stop()`，误触会立刻丢失会话。常见 CLI（fish、ipython、jupyter）采用双击退出避免误触。

## 范围

**改动**：`internal/ui/app.go`（KeyCtrlC 分支 + 新增字段）。
**不改**：
- `cmd/agent/main.go` SIGINT handler（保持单次退出，作为外部 `kill -INT` 的兜底）。
- `internal/ui/selectable_textview.go` 的 KeyCtrlC（独立组件，文本复制语义）。
- README 文档（行为本质不变，无需更新）。

## 行为

| 场景 | 行为 |
|---|---|
| chatPanel 有选中文本 | 复制选中内容（保持现状） |
| 无选中，首次 Ctrl+C | 状态栏提示「再次按 Ctrl+C 退出」 |
| 无选中，2 秒内再次 Ctrl+C | 退出程序 |
| 无选中，超过 2 秒后再按 | 视为新一次"首次"，再次提示 |

时间窗口：**2 秒**，硬编码常量。

## 实现要点

### App 字段
```go
type App struct {
    // ...existing code...
    lastCtrlCAt time.Time
    // ctrlCHint shows the "press Ctrl+C again to quit" hint.
    // Overridable for tests; nil 时走 ShowToast。
    ctrlCHint func()
}
```

### KeyCtrlC 分支（替换现有逻辑）
```go
case event.Key() == tcell.KeyCtrlC:
    if a.chatPanel.HasSelection() {
        a.chatPanel.CopySelection()
        return nil
    }
    if time.Since(a.lastCtrlCAt) < 2*time.Second {
        a.Stop()
        return nil
    }
    a.lastCtrlCAt = time.Now()
    a.fireCtrlCHint()
    return nil
```

`fireCtrlCHint` 优先调 `ctrlCHint` 字段（测试注入用），否则走 `ShowToast`。这样测试不会因 `QueueUpdateDraw` 在未启动的 `Application` 上阻塞。

### 提示渠道
生产复用现有 `App.ShowToast` → `StatusBar.ShowMessage`，3 秒后自动清除。

提示展示 3 秒 vs 退出窗口 2 秒：用户在 2-3 秒之间看到提示但需要重新触发——可接受，提示是软引导。

### 不引入
- 不加 mutex：tview 的 InputCapture 在主 goroutine 串行执行，`lastCtrlCAt` 仅在该回调读写。
- 不加 timer：用时间戳比较即可，零后台 goroutine。
- 不抽组件：YAGNI，仅此一处。

## 测试

`internal/ui/app_test.go` 新增用例：
1. 首次 KeyCtrlC（无选中）→ `Stop` 未被调用，`lastCtrlCAt` 已设置。
2. 2 秒窗口内第二次 KeyCtrlC → 走退出路径。
3. 模拟超过 2 秒后第二次 → 仍走"首次"路径，不退出。
4. 有选中文本时 KeyCtrlC → 走复制路径，不影响 `lastCtrlCAt`。

实现策略：直接构造 `tcell.NewEventKey(tcell.KeyCtrlC, 0, tcell.ModNone)` 喂给 `inputCapture`；通过覆写或注入"退出回调"来检测 Stop 调用（避免真正退出 tview app）。

## 风险

- **窗口过短/过长**：2 秒为业界常见值，覆盖大多数误触场景。可接受。
- **状态栏被其他消息覆盖**：若用户首次 Ctrl+C 后 toast 立刻被新消息盖掉，看不到提示也不影响功能（仍需 2 秒内再按）。可接受。

## 不在范围

- 配置化窗口时长（YAGNI）。
- 信号层 SIGINT 同步双击（B 选项）。
- 全局 `quitGuard` 抽象（YAGNI）。
