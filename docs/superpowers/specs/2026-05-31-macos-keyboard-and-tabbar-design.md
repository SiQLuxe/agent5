# macOS Keyboard Compatibility and Tab Bar Visual Fix

## Problem

1. **Alt shortcuts don't work on macOS**: macOS Option key sends Unicode characters (˜, ∑, ®, etc.) instead of ModAlt key events. All `Alt+N/W/R/.//,/T/Y/S` shortcuts non-functional on macOS.
2. **`q` quit shortcut conflicts with typing**: Pressing `q` in composer exits app. User cannot type `q` in messages.
3. **Tab bar no active visual indicator**: Active tab has no background color, making it hard to identify selected session.

## Design

### 1. Dual shortcut support (Ctrl + Alt)

Both Ctrl and Alt variants work simultaneously:

| Ctrl (macOS/Linux/Win) | Alt (Linux/Win) | Action |
|------------------------|-----------------|--------|
| Ctrl+N                 | Alt+N           | New session |
| Ctrl+W                 | Alt+W           | Close session |
| Ctrl+R                 | Alt+R           | Rename session |
| Tab                    | Alt+.           | Next session |
| Shift+Tab              | Alt+,           | Previous session |
| Ctrl+T                 | Alt+T           | Toggle thinking |
| Ctrl+Y                 | Alt+Y           | Toggle collapse |
| Ctrl+K                 | Alt+Shift+S     | Toggle theme |
| —                      | Alt+{1-9}       | Direct session switch |
| Ctrl+F                 | —               | Search chat |
| Ctrl+O                 | —               | Help |
| Ctrl+C                 | —               | Quit |

Ctrl variants work on all platforms. Alt variants work on Windows/Linux only (macOS Option key doesn't generate ModAlt by default).

**Files changed:**
- `internal/ui/app.go` — `handleInput()`: Ctrl `KeyCtrl*`/`KeyTab`/`KeyBacktab` cases added, `ModAlt` block restored for cross-platform compatibility
- `internal/ui/keymap.go` — `ShortHelp()`/`FullHelp()`: shows both variants
- `internal/ui/app_test.go` — 8 Ctrl shortcut tests + Alt compatibility

### 2. Remove `q` quit shortcut

Delete the `q` → quit handler. Only `Ctrl+C` remains as exit mechanism.

```go
// REMOVED:
case event.Rune() == 'q' && event.Modifiers() == tcell.ModNone:
    a.Stop()
```

Also removed `Quit` field from `KeyMap` struct.

**Files changed:**
- `internal/ui/app.go` — remove case block
- `internal/ui/keymap.go` — remove `Quit` field, update help text
- `internal/ui/app_test.go`, `internal/ui/keymap_test.go` — update test assertions

### 3. Tab bar active background color

Active tab uses theme accent color (`colors.Accent`, default `#569cd6`) as background with white text. Applied through `tabDock.SetColors()` in `applyTheme()`:

```go
a.tabDock.SetColors(white, hexToTCell(colors.Accent), gray, default)
```

Inactive tabs: gray text, default background. Bar background: default.

**Files changed:**
- `internal/ui/app.go` — `applyTheme()`: added `tabDock.SetColors()` call

### 4. Diagnostic tool cleanup

Removed `tools/keydiag/` directory and `/tmp/keydiag`/`/tmp/tdiag` binaries.

## Test Plan

- `go build -o agent ./cmd/agent` — succeeds
- `go test ./...` — all tests pass
- Manual: run `./agent`, verify Ctrl+N / Alt+N create session, Tab / Alt+. switch session
- Manual: verify typing 'q' in composer inputs 'q' instead of quitting
- Manual: verify active tab shows accent background color
- Verification: `cat /tmp/tcell-diag.log` confirms tcell produces correct Key events
