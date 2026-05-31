# macOS Keyboard Compatibility and Tab Bar Visual Fix

## Problem

1. **Alt shortcuts don't work on macOS**: macOS Option key sends Unicode characters (˜, ∑, ®, etc.) instead of ModAlt key events. All `Alt+N/W/R/.//,/T/Y/S` shortcuts were non-functional on macOS Terminal.app.
2. **`q` quit shortcut conflicts with typing**: Pressing `q` in composer exits the app. User cannot type the letter `q` in messages.
3. **Tab bar no active visual indicator**: Active tab has no background color, making it hard to identify which session is selected.

## Solution

### 1. Replace Alt shortcuts with Ctrl shortcuts (completed)

All `Alt+` shortcuts replaced with macOS-compatible `Ctrl+` or `Tab` equivalents:

| Original | Replacement | Action |
|----------|-------------|--------|
| Alt+N    | Ctrl+N      | New session |
| Alt+W    | Ctrl+W      | Close session |
| Alt+R    | Ctrl+R      | Rename session |
| Alt+.    | Tab         | Next session |
| Alt+,    | Shift+Tab   | Previous session |
| Alt+T    | Ctrl+T      | Toggle thinking |
| Alt+Y    | Ctrl+Y      | Toggle collapse |
| Alt+Shift+T | Ctrl+K   | Toggle theme |
| Alt+{1-9} | (removed)  | Direct session switch (unreliable on macOS) |

**Files changed:**
- `internal/ui/app.go` — `handleInput()`: replaced `ModAlt` branch with `KeyCtrl*`/`KeyTab`/`KeyBacktab` cases
- `internal/ui/keymap.go` — `ShortHelp()`/`FullHelp()`: updated display text
- `internal/ui/app_test.go` — added 8 shortcut tests

### 2. Remove `q` quit shortcut

Delete the `q` → quit handler. Only `Ctrl+C` remains as exit mechanism.

```go
// DELETE this block:
case event.Rune() == 'q' && event.Modifiers() == tcell.ModNone:
    a.Stop()
    return nil
```

**Files changed:**
- `internal/ui/app.go` — remove case block
- `internal/ui/keymap.go` — remove `q` from help text

### 3. Tab bar active background color

Active tab uses theme accent color as background (`activeBg: colors.Accent`, `activeFg: white`). Applied through existing `SetColors()` call in `applyTheme()`.

Inactive tabs: `ColorGray` text, `ColorDefault` background. Bar background: `ColorDefault`.

**Files changed:**
- `internal/ui/app.go` — `applyTheme()`: add `tabDock.SetColors()` call

### 4. Diagnostic tool cleanup

Remove `tools/keydiag/` directory after investigation is complete.

## Test Plan

- Build: `go build -o agent ./cmd/agent` succeeds
- Test: `go test ./...` passes all tests
- Manual: Run `./agent` and verify Ctrl+N/W/Tab etc. create/switch/close sessions
- Manual: Verify typing 'q' in composer inputs 'q' instead of quitting
