# SelectableTextView — 可选中文字的 tview 自定义控件

## 概述

为 Agent-TUI 的聊天面板添加鼠标文本选中功能。方案是基于 `tview.Box` 创建自定义控件 `SelectableTextView`，在不更换 TUI 框架、不改变现有渲染流水线的前提下，实现完整的文本选中、复制、全选能力。

## 结构

### 文件

新增 `internal/ui/selectable_textview.go`，对外部 API 完全透明。`ChatPanel` 替换内部 `*tview.TextView` 为 `*SelectableTextView`，其余代码不变。

### 数据结构

```go
type SelectableTextView struct {
    *tview.Box
    text          string
    dynamicColors bool
    regions       bool
    scrollable    bool
    wordWrap      bool
    maxLines      int

    // Selection state
    selecting          bool
    selectAnchorRow    int
    selectAnchorCol    int
    selectEndRow       int
    selectEndCol       int
    selectionVisible   bool

    highlights []string
}
```

选中坐标使用逻辑行/列（软换行后），与屏幕坐标解耦。选中范围始终归一化为 anchor <= end。

## 功能

### Draw 管线

1. `Box.Draw(screen)` 绘制边框/背景
2. 获取内部可见区域 `(x, y, w, h)`
3. 解析 `text` 为换行后的 cell grid，保留颜色样式
4. 逐行逐 cell 绘制：
   - 在选中范围内 → `style.Reverse(true)`
   - 否则 → 正常样式
5. 叠加搜索高亮区域标签

文本变更时全量解析，选中变更仅重绘不重新解析。

### 鼠标事件

| 操作 | 行为 |
|---|---|
| `MouseLeft` 按下 | 记录锚点位置，标记选中开始，清除旧选中 |
| `MouseLeft` 拖拽 | 更新终点位置，显示选中高亮，请求重绘 |
| `MouseLeft` 释放 | 结束拖拽状态，选中保留 |
| `MouseScrollUp/Down` | 与现有行为一致 |
| `Shift + MouseLeft` | 从已有锚点扩展到点击位置 |

坐标转换：`tcell.EventMouse.Position()` → 减去 `GetInnerRect()` 偏移 → 加上滚动偏移 → 逻辑行号。

### 键盘操作

| 按键 | 行为 |
|---|---|
| `Ctrl+C` / `Ctrl+Q` | 复制选中文本到剪贴板 |
| `Ctrl+A` | 全选 |
| `Escape` | 清除选中 |
| `Shift+↑/↓/←/→` | 扩展/收缩选中范围 |
| `PgUp/PgDn/↑/↓` | 滚动视口，保持选中 |

剪贴板：优先 `tcell.Screen.SetClipboard()`，否则通过平台命令回退。

选中文本提取：根据锚点/终点坐标从原始文本裁剪对应子串。

## 不涉及的范围

- `ChatPanel` 对外 API 不变（`SetText`、`ScrollTo`、`Highlight` 等）
- `session.go` 的 `renderMessageToBuilder` 渲染格式不变
- `app.go` 的布局和鼠标滚轮处理不变
- 搜索高亮功能不变
