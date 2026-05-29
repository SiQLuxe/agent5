# TView Migration Design

**Date:** 2026-05-29
**Status:** Draft
**Goal:** Replace Bubble Tea v2 / Bubbles v2 / Lipgloss v2 with rivo/tview (tcell-backed) to fix Chinese IME input and simplify the UI stack.

## Motivation

Chinese IME input is broken under Bubble Tea v2 because `termenv` (the terminal layer used by Bubble Tea) does not properly handle IME composition events on Windows ConPTY. tview uses tcell/v2 under the hood, which has native `EventIME` support and handles IME correctly on all platforms.

## Architecture

```
tview.Application
└── tview.Pages (root)
    ├── "chat" → tview.Flex (vertical)
    │   ├── StatusBar (tview.TextView, 1 row, fixed)
    │   ├── ChatPanel (tview.TextView, flex, scrollable)
    │   ├── Composer (tview.TextArea, 1-8 rows, dynamic)
    │   └── TabDock (tview.Box, 1 row, fixed, custom Draw)
    ├── "search" → SearchModal (tview.Flex centered overlay)
    │   ├── tview.InputField (search query)
    │   └── tview.TextView (match count + nav hints)
    └── "help" → HelpModal (tview.TextView overlay)
```

## Component Mapping

| Current (Bubble Tea) | tview Equivalent | Notes |
|---|---|---|
| `tea.Program` + `tea.Model` | `tview.Application` + `Primitive` tree | Event loop replacement |
| StatusBar (lipgloss custom) | `tview.TextView` (fixed 1 row) | SetTextAlign, SetBackgroundColor |
| ChatPanel (bubbles viewport wrapper) | `tview.TextView` (scrollable + dynamic colors) | SetScrollable(true), SetDynamicColors(true) |
| Composer (bubbles textarea wrapper) | `tview.TextArea` | Native IME, WordWrap, MaxHeight |
| TabDock (lipgloss custom Draw) | `tview.Box` (custom Draw callback) | Reuse tab logic, map to tcell.Style |
| Help (bubbles help wrapper) | `tview.TextView` (Pages overlay) | Key list as plain text |
| Search (hijacked composer) | `tview.InputField` (Pages overlay) | Enter/Shift+Enter navigate, Esc exit |
| Keymap (bubbles key) | `tview.InputCapture` | App-level capture, stop propagation on match |
| Theme (lipgloss Colors) | `tview.Styles` + `tcell.Style` | Map ThemeService colors to tview globals |
| Mouse events | tview native (tcell) | No special handling needed |

## Layout (Vertical Flex)

```
+---------------------------+
| StatusBar (1 row fixed)   |
+---------------------------+
| ChatPanel (flex)          |
|                           |
|                           |
+---------------------------+
| Composer (1-8 rows)       |
+---------------------------+
| TabDock (1 row fixed)     |
+---------------------------+
```

## Event Handling

```
Application.SetInputCapture()
  ↓
key.Matches (quit / send / switch / search / help / etc.)
  ├─ match → execute action + return nil (stop propagation)
  └─ no match → return event (pass to focused Primitive)
                     ↓
               Composer (tview.TextArea handles natively, IME works)
```

- Ctrl+Enter → Send (captured at App level, `tcell.KeyCtrlEnter`)
- Enter → Newline (TextArea default behavior)
- Ctrl+F / `/` → Enter search mode (switch to "search" page)
- Esc → Exit search/help (switch back to "chat" page, refocus Composer)
- Alt+{1-9} → Switch session
- Tab → Next focus or TextArea default

## Theming

- `tview.Styles.PrimitiveBackgroundColor` ← mapped from `ThemeService`
- `tview.Styles.ContrastBackgroundColor` ← tab bar background
- `tview.Styles.PrimaryTextColor` ← message text
- `tview.Styles.BorderColor` ← borders
- ChatPanel: `SetDynamicColors(true)` — chroma ANSI output works directly
- TabDock: custom Draw with `tcell.Style`, colors from ThemeService
- Composer: `SetBackgroundColor()` / `SetTextStyle()` from theme
- Color conversion helper: hex string → `tcell.Color`

## Files Changed

### Rewrite (10 files)

| File | Current Lines | Estimated New | Description |
|---|---|---|---|
| `ui.go` | 514 | 300-400 | App struct + init + InputCapture + page wiring |
| `chatpanel.go` | 173 | 80-100 | tview.TextView wrapper |
| `keymap.go` | 112 | 60-80 | tcell/tview key constants |
| `composer/composer.go` | 102 | 60-80 | tview.TextArea wrapper |
| `composer/composer_test.go` | 58 | ~50 | Adapt tests |
| `status/status.go` | 66 | 40-50 | tview.TextView wrapper |
| `status/status_test.go` | 39 | ~30 | Adapt tests |
| `tabbar/tabbar.go` | 188 | 150-200 | tview.Box custom Draw |
| `tabbar/tabbar_test.go` | 86 | ~60 | Adapt tests |
| `ui_test.go` | 104 | ~100 | Rewrite for tview |
| `cmd/agent/main.go` | 46 | ~20 | tview.Application setup |

### Delete (2 files)
- `internal/ui/editor/editor.go`
- `internal/ui/editor/editor_test.go`

### Unchanged (6 files)
- `session.go` (274 lines) — business logic, pure functions
- `session_render_test.go` (113 lines)
- `session_test.go` (111 lines)
- `theme.go` (81 lines)
- `theme_service.go` (108 lines)
- `syntax/syntax.go` (59 lines)

## Dependency Changes

**Remove:**
- `charm.land/bubbletea/v2`
- `charm.land/bubbles/v2`
- `charm.land/lipgloss/v2`
- All indirect charmbracelet dependencies (colorprofile, x/termios, etc.)

**Add:**
- `github.com/rivo/tview` (pulls in `github.com/gdamore/tcell/v2`)

## Testing Strategy

- All existing test files updated to use tview types
- Session render tests: unchanged (pure string output)
- Composer tests: test SetValue/GetValue, focus, paste
- TabDock tests: test tab add/remove/click
- Status tests: test SetText/View content
- UI tests: test Application setup, page switching, key dispatch
- Manual: Chinese IME input verification on Windows

## Success Criteria

- All existing tests pass after migration
- Chinese IME input (拼音, 五笔) works correctly in the Composer
- All existing features work identically: multi-session, search, help, theme switching, mouse
