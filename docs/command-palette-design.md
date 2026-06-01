# Command Palette: Filtering, Commands, Skill Loading

## Feature 1: Command Palette Real-Time Filtering

### Layout

```
┌─ CommandPalette ─────────────┐
│ > filter text...             │  ← InputField (new)
├──────────────────────────────┤
│ 📂 New Session               │
│ 📂 Search                    │
│ ⚡ code-review               │
│ ...                          │
└──────────────────────────────┘
```

### Data Flow

```
InputField.SetChangedFunc(text)
  → CommandPalette.SetFilter(text)
    → applyFilter() — prefix match on allCommands
      → tview.List.Clear() + repopulate
```

### Keyboard Interaction

| Key | Behavior |
|-----|----------|
| Typing | Real-time prefix filtering |
| Enter | Execute selected command |
| Esc | Clear filter text → if already empty, close palette |
| Tab | Insert `/cmdName ` into composer and execute |
| ↓/↑ | Navigate command list |
| Ctrl+Backspace | Delete word before cursor |

### Focus Management

- `/` or `Ctrl+P` → focus to InputField
- Enter/Tab execute → focus back to composer
- Esc close → focus back to composer

### Implementation

Modify `internal/ui/command_palette.go`:
- Add `*tview.InputField` as first item in Flex
- Wire `SetChangedFunc` to `SetFilter`
- Need to handle `SetInputCapture` on palette to intercept Enter/Esc/Tab when InputField has focus
- Add `KeyboardInterceptHandler` method for the App to call

---

## Feature 2: New Builtin Commands

### Commands to Add (6)

| Name | Description | Implementation |
|------|-------------|----------------|
| Previous Session | Switch to the previous session | `a.prevSession()` |
| Rename Session | Rename the current session | Overlay with InputField |
| Scroll to Top | Scroll chat to the top | `a.chatPanel.ScrollToTop()` |
| Scroll to Bottom | Scroll chat to the bottom | `a.chatPanel.ScrollToBottom()` |
| Clear Input | Clear the text input area | `a.composer.ClearInput()` |
| Reload Skills | Reload skills from disk | Rerun LoadSkillsDir + sync |

### Rename Session Overlay

New mode `ModeRename`:

```
┌─ Rename Session ────────────┐
│ > my-session-name           │  InputField
│                             │
│ Enter: confirm  Esc: cancel │
└─────────────────────────────┘
```

- Uses a pages overlay similar to search/help
- Pre-fills InputField with current session label
- Enter updates session label + tab label
- Esc cancels and returns to chat

---

## Feature 3: Skill Loading Recursion + Hot Reload

### Directory Recursion

Replace flat `ioutil.ReadDir` with `filepath.Walk`:

```
skills/
├── code-review/SKILL.md
├── debug/
│   ├── SKILL.md
│   └── test/
│       └── SKILL.md           ← also found
└── templates/SKILL.md
```

- Skill name = directory name (leaf dir containing SKILL.md)
- Path prefix preserved for display only
- Backward compatible with existing flat layout

### Hot Reload (Polling)

- `App` starts background goroutine with `time.NewTicker(5s)`
- Maintains `map[string]time.Time` of SKILL.md paths → mtimes
- On tick: stat all known files + discover new ones
- Changed file → reload → update SkillRegistry → sync CommandRegistry
- Removed file → unregister from SkillRegistry → sync CommandRegistry
- Uses `QueueUpdateDraw` for thread-safe UI updates

### New Methods Needed

```go
// skill_loader.go
func LoadSkillsDirRecursive(registry *SkillRegistry, rootDir string) error

// app.go
func (a *App) startSkillWatcher()
func (a *App) reloadSkills()
```

---

## Files to Modify

| File | Changes |
|------|---------|
| `internal/ui/command_palette.go` | Add InputField, integrate filtering |
| `internal/ui/app.go` | 6 new commands, Rename overlay, skill watcher |
| `internal/ui/app_test.go` | Tests for new commands and rename |
| `internal/service/skill_loader.go` | Recursive loading |
| `cmd/agent/main.go` | Pass skills path for reload |
