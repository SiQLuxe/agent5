# Slash Command Autocomplete Menu

## Goal

Replace the existing inline `tview.List`-based command suggestion (which only shows skill commands) with a dedicated `SuggestionMenu` component that shows all commands (builtin + skill) in a two-column aligned layout with a status bar.

## Background

The project already has:
- An inline `suggestionList` (`*tview.List`) in `app.go` shown when the user types `/` in the composer
- A full-screen `CommandPalette` (Ctrl+P) with InputField filtering
- Both use the shared `CommandRegistry` for command data

The existing inline suggestion only filters `CmdSkill` commands, uses `tview.List` with unaligned main+secondary text, and has no status bar.

## Component: `SuggestionMenu`

New file `internal/ui/suggestion/suggestion.go`, new package `suggestion`.

### Layout

```
┌─ SuggestionMenu (tview.Flex, FlexRow) ────────┐
│  /se                                          │  ← headerBar (TextView)
│  Search          搜索当前会话中的消息          │  ← list (Box, custom draw)
│  Scroll to Top   滚动到消息顶部                │
│  Scroll to Bottom 滚动到消息底部               │
│  ...                                          │
├────────────────────────────────────────────────┤
│  ↑↓ 选择  Enter 回填  Esc 关闭                 │  ← statusBar (TextView)
└────────────────────────────────────────────────┘
```

### Struct

```go
type SuggestionMenu struct {
    *tview.Flex
    items      []*service.Command   // 所有命令
    filtered   []*service.Command   // 当前过滤结果
    filterText string
    selected   int                  // 选中索引，0-based
    list       *tview.Box           // 命令列表区域（自定义绘制）
    headerBar  *tview.TextView      // 显示当前过滤前缀
    statusBar  *tview.TextView      // 底部提示
    visible    bool
}
```

### Custom List Rendering

`tview.Box.SetDrawFunc` draws items line by line:

1. Compute column widths: command name = min(40% of available width, 20 chars), description = remaining
2. Use `runewidth` to account for CJK characters in column alignment
3. Selected row: `tcell.Style.Background(tcell.ColorOrange)` foreground white
4. Unselected: default background, white foreground
5. Description text: gray color

### Methods

- `New() *SuggestionMenu` — create component, initialize headerBar/list/statusBar in Flex(FlexRow), all hidden by default (Height=0)
- `SetCommands(cmds []*service.Command)` — set candidate pool
- `SetFilter(text string)` — real-time prefix filter (case-insensitive), reset selected=0 if filtered list changes, hide if 0 results
- `SelectNext()` — clamp to list bounds
- `SelectPrev()` — clamp to list bounds
- `Selected() *service.Command` — return selected command or nil
- `Show()` — set visible=true, set Flex heights
- `Hide()` — set visible=false, set all heights to 0
- `Visible() bool`
- `Height() int` — return current desired height (min(filtered count, 8) + 2 for header+status)

## Integration: `App` Changes

### Layout Change

In `NewApp()`, replace `a.suggestionList` with `a.suggestionMenu` in `chatFlex`:

```
chatFlex layout:
  StatusBar       (row 1, fixed 1)
  ChatPanel       (row 2, flex)
  SuggestionMenu  (row 3, fixed height or 0)
  Composer        (row 4, fixed 3)
  TabDock         (row 5, fixed 1)
```

### Data Flow

`onComposerChange(text string)`:
1. If text starts with `/`:
   - prefix = text[1:] (text after `/`)
   - iterate `CommandRegistry.List()`, prefix match on `cmd.Name` (case-insensitive)
   - set matches as filtered items
   - if matches > 0: `suggestionMenu.Show()`, else `Hide()`
2. If text does not start with `/`:
   - `suggestionMenu.Hide()`

### Keyboard Handling

| Key | Menu Visible | Menu Hidden |
|-----|-------------|-------------|
| `↓` | `suggestionMenu.SelectNext()` | Input history navigation (existing) |
| `↑` | `suggestionMenu.SelectPrev()` | Input history navigation (existing) |
| `Enter` | Fill `"/"+cmd.Name+" "` into composer, hide menu | Send message (existing) |
| `Tab` | Same as Enter fill + hide | Pass through (existing) |
| `Esc` | Hide menu | Pass through |

### State Removed

Remove from `App` struct:
- `suggestionList *tview.List`
- `suggestionCmds []*service.Command`

## Edge Cases

- **No match:** menu hides (Height=0), composer shows typed text normally
- **Single match:** auto-selected, Enter/Tab fills it
- **Composer empty:** no `/` prefix, menu stays hidden
- **Backspace to just `/`:** empty prefix → show all commands
- **Backspace past `/`:** composer empty → hide menu
- **Menu height overflow:** cap at max 8 items + 2 rows = 10 rows total
- **CJK in command names:** use `go-runewidth` for column alignment

## Testing

### Unit tests (`internal/ui/suggestion/suggestion_test.go`)

- `TestNew` — non-nil, has header/list/status
- `TestSetCommands` — items stored correctly
- `TestSetFilter` — prefix filtering works (case-insensitive)
- `TestSetFilterNoMatch` — empty filtered list
- `TestNavigation` — SelectNext/SelectPrev bounds
- `TestSelected` — returns correct command
- `TestShowHide` — visibility state + Height()
- `TestHeight` — max 10 rows (8 items + header + status)

### Integration tests (update `internal/ui/app_test.go`)

- `TestSlashShowsAllCommands` — type `/` and verify menu shows builtin + skill commands
- `TestSlashFiltering` — type `/Se` and verify menu shows only matching
- `TestSlashEnterFills` — enter fills command name with trailing space
- `TestSlashNoSlashHides` — menu hidden when no `/` prefix
- `TestSlashEscHides` — Esc closes menu, composer unchanged

## Files

| File | Action |
|------|--------|
| `internal/ui/suggestion/suggestion.go` | Create — new SuggestionMenu component |
| `internal/ui/suggestion/suggestion_test.go` | Create — unit tests |
| `internal/ui/app.go` | Modify — replace suggestionList with suggestionMenu |
| `internal/ui/app_test.go` | Modify — update/add tests |
