# Chat Viewport & Message Navigation Design

**Date**: 2026-05-29
**Status**: Draft

## Problem

When chat messages accumulate, three issues emerge:

1. **Long messages dominate the viewport** — a single AI response (code, analysis) can fill the entire screen, hiding other messages
2. **Scrolling is crude** — only pgup/pgdn and ctrl+up/down, no scrollbar indicator, no way to locate specific content
3. **Performance degrades** — `View()` renders ALL messages every frame, then slices by line offset. O(n) with message count.

## Solution: Viewport Rendering + Message Folding + Scrollbar + Search

### 1. Viewport On-Demand Rendering

**Current**: `ChatPanel.View()` renders all messages → joins into string → splits by `\n` → slices by scroll offset.

**New**: Pre-compute each message's rendered line count, then only render messages within the visible window.

**Data structure change** (`session.go`):

```go
type Message struct {
    Role      Role
    Content   string
    Thinking  *Thinking
    Timestamp time.Time
    Collapsed bool     // new: whether message is folded
    lineCount int      // new: cached rendered line count (0 = uncached)
}
```

**Algorithm**:

1. On `View()`, iterate messages from top, accumulating `lineCount` (or collapsed line count = 2)
2. Skip messages entirely above the viewport
3. Render only messages that intersect the viewport
4. For partially visible messages, render full then slice the visible portion

**Cache invalidation**: `lineCount` is recalculated only when:
- Message content changes (streaming)
- Window width changes (reflow)
- Collapse state toggles

### 2. Long Message Folding

**Rules**:
- Messages exceeding `maxVisibleLines` (default: 8, configurable via `config.toml` `[ui] max_visible_lines`) are auto-collapsed on arrival
- Collapsed display: badge + timestamp + first 3 lines + `▼ [展开 N 行]` indicator
- User messages default to expanded; assistant messages with >8 lines default to collapsed
- New messages (just arrived) are always expanded initially, then collapse after next user message

**Interaction**:
- `ctrl+l` toggles collapse of the focused message
- `ctrl+y` remains as toggle thinking (existing behavior unchanged)
- When a collapsed message is focused (cursor on it), it temporarily previews more content
- Collapsed messages contribute only their collapsed line count to viewport calculations

**Collapsed rendering**:
```
👾 玄 14:30
   func quickSort(arr []int) {
     if len(arr) <= 1 {
   ▼ [展开 45 行代码]
```

### 3. Scrollbar Indicator

**Design**:
- 1-column-wide scrollbar on the right edge of the chat panel
- Characters: `█` for thumb position, `░` for track
- Thumb height = max(1, viewportHeight² / totalLines)
- Thumb position = (scrollOffset / maxScroll) * trackHeight

**Example** (20-line viewport, 100 total lines):
```
┌────────────────────────────┬─┐
│ message content            │░│
│ message content            │█│
│ message content            │█│
│ message content            │░│
│ message content            │░│
└────────────────────────────┴─┘
```

**Status line**: Bottom of chat panel shows `行 45/100` when scrolled away from bottom.

### 4. Message Search

**Interaction**:
- `ctrl+f` enters search mode
- Composer input becomes search input (placeholder: "搜索消息...")
- Real-time highlight of matches as you type
- `enter` jumps to next match, `shift+enter` to previous
- `esc` exits search mode, restores composer

**Implementation**:
- Simple substring match across all message content
- Search results cached as `[]int` (message indices)
- Current match index tracked, viewport auto-scrolls to match
- Matched text highlighted with inverted colors

### 5. Performance Optimizations

| Optimization | Impact |
|---|---|
| `lineCount` cache per message | Avoid re-rendering unchanged messages |
| Viewport-only rendering | O(visible) instead of O(total) |
| Incremental scroll | Only re-render on scroll delta, not full re-render |
| Lazy line count | Calculate only when message enters viewport vicinity |

**Scroll behavior**:
- Auto-scroll to bottom on new message (if already at bottom)
- Preserve scroll position if user has scrolled up
- `ctrl+end` / `G` to jump to bottom
- `ctrl+home` / `g` to jump to top

### Key Bindings Summary

| Key | Action |
|---|---|
| `pgup` / `pgdown` | Scroll half page |
| `ctrl+up` / `ctrl+down` | Scroll 1 line |
| `ctrl+y` | Toggle thinking (existing) |
| `ctrl+l` | Toggle collapse on focused message |
| `ctrl+f` | Enter search mode |
| `enter` (search) | Next match |
| `shift+enter` (search) | Previous match |
| `esc` (search) | Exit search |
| `G` | Scroll to bottom |
| `g` | Scroll to top |

### Files to Modify

| File | Changes |
|---|---|
| `internal/ui/session.go` | Add `Collapsed`, `lineCount` to `Message` |
| `internal/ui/chatpanel.go` | Viewport rendering, folding, scrollbar, search |
| `internal/ui/ui.go` | New key bindings, search mode state |
| `internal/ui/composer/composer.go` | Search mode input handling |
| `configs/config.toml` | Add `[ui] max_visible_lines` setting |
| `internal/data/config/config.go` | Parse `max_visible_lines` config |

### Out of Scope

- Message persistence/search index (future: SQLite FTS)
- Regex search (simple substring first)
- Message bookmarks/pins
- Multi-session search
