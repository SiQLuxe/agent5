# Command Palette: Filtering, Commands, Skill Loading — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use subagent-driven-development or executing-plans to implement task-by-task.

**Goal:** Add real-time filtering to command palette, 6 new builtin commands with rename overlay, and recursive+hot-reload skill loading.

**Architecture:** 5 sequential tasks building on each other: (1) palette InputField + filtering, (2) new builtin commands, (3) rename overlay, (4) recursive skill loading, (5) hot-reload polling.

**Tech Stack:** Go, tview (TUI framework)

---

### Task 1: Command Palette with InputField Filtering

**Files:**
- Modify: `internal/ui/command_palette.go`
- Modify: `internal/ui/app.go` (handleInput ModeCommandPalette)
- Create: `internal/ui/command_palette_test.go`

- [ ] **Step 1: Add InputField + navigation methods to CommandPalette**

Replace `command_palette.go` with version that adds `filterInput *tview.InputField` at top of Flex, `GetFilterInput()`, `SelectNext()`, `SelectPrev()` methods, and wires `SetChangedFunc` → `SetFilter`.

Key changes:
- Flex layout: `InputField (row 1, fixed height 1) + List (row 2, fills remaining)`
- `InputField` label: `[::b]/ []`, placeholder: `"Type to filter commands..."`
- `SelectNext()` increments `SetCurrentItem` if not at end
- `SelectPrev()` decrements `SetCurrentItem` if not at start
- `SetFilter("")` on `SetCommands` and `SetMode` to reset on open

- [ ] **Step 2: Update app.go handleInput for up/down navigation + Enter/Esc behavior**

In `case ModeCommandPalette:`:
- ↑ intercept → `palette.SelectPrev()`
- ↓ intercept → `palette.SelectNext()`
- PgUp/PgDn → SelectPrev/Next 5 times
- Esc → if filter text non-empty, clear filter + input text; else exit palette
- Enter/Tab → unchanged (execute selected command)
- Backspace → `return event` (let InputField handle)

In `enterCommandPalette`:
- Add `a.commandPalette.SetFilter("")` and `a.commandPalette.GetFilterInput().SetText("")`
- Change `a.SetFocus(a.commandPalette)` → `a.SetFocus(a.commandPalette.GetFilterInput())`

- [ ] **Step 3: Create test file `internal/ui/command_palette_test.go`**

Tests:
- `TestNewCommandPalette` — verify non-nil + has filter input
- `TestCommandPaletteFilterPrefix` — 3 commands, filter "Se" → 1 result (Search)
- `TestCommandPaletteFilterSkillsOnly` — ShowSkills mode hides builtins
- `TestCommandPaletteNavigation` — SelectNext/SelectPrev with bounds
- `TestCommandPaletteSelectedEmpty` — nil when no commands

- [ ] **Step 4: Run tests: `go test ./... -count=1` — ALL pass**

- [ ] **Step 5: Commit**

---

### Task 2: Add 6 New Builtin Commands

**Files:**
- Modify: `internal/ui/app.go` (registerBuiltinCommands)
- Modify: `internal/ui/app_test.go` (tests)

Commands to add:
1. `Previous Session` → `a.prevSession()`
2. `Rename Session` → `a.enterRename()` (wired in Task 3)
3. `Scroll to Top` → `a.chatPanel.ScrollToTop()`
4. `Scroll to Bottom` → `a.chatPanel.ScrollToBottom()`
5. `Clear Input` → `a.composer.ClearInput()`
6. `Reload Skills` → `a.reloadSkills()` (wired in Task 5)

Tests in `app_test.go`:
- `TestCommandPaletteAllNewCommands` — verify all 6 registered as CmdBuiltin
- `TestCommandScrollToTop` — smoke test (no panic)
- `TestCommandScrollToBottom` — smoke test (no panic)
- `TestCommandClearInput` — verify composer cleared
- `TestCommandPreviousSession` — verify switches to prev session

- [ ] **Run tests: `go test ./... -count=1` — ALL pass**
- [ ] **Commit**

---

### Task 3: Rename Session Overlay

**Files:**
- Modify: `internal/ui/app.go`
- Modify: `internal/ui/app_test.go`

Add `ModeRename` constant. Add `renameInput *tview.InputField` and `renamePage *tview.Flex` to App struct. Build rename overlay page (centered InputField + instruction text). Add to `a.pages`.

In `handleInput`:
- `case ModeRename`: Enter → `applyRename()`, Esc → `exitRename()`, others → `return event`

Methods:
- `enterRename()` — sets mode, pre-fills input with current session label, switches page, focus input
- `applyRename()` — updates session.Label + tabDock label, calls exitRename
- `exitRename()` — mode = ModeChat, switch to chat, focus composer

Tests:
- `TestRenameSession` — enterRename pre-fills correctly
- `TestRenameSessionApply` — applyRename updates label + exits
- `TestRenameSessionCancel` — exitRename preserves original label
- `TestRenameSessionEmptyName` — empty name defaults to "New Session"
- Update `TestShortcutRenameSession_CtrlR_NoPanic` → verify ModeRename

- [ ] **Run tests: `go test ./... -count=1` — ALL pass**
- [ ] **Commit**

---

### Task 4: Recursive Skill Loading

**Files:**
- Modify: `internal/service/skill_loader.go`
- Modify: `internal/service/skill_loader_test.go`
- Modify: `internal/service/skill_registry.go` (add ClearPrompts)

- [ ] **Step 1: Add ClearPrompts to SkillRegistry**

```go
func (r *SkillRegistry) ClearPrompts() {
    r.mu.Lock()
    defer r.mu.Unlock()
    for name, skill := range r.skills {
        if skill.Type == SkillPrompt {
            delete(r.skills, name)
        }
    }
}
```

- [ ] **Step 2: Add ClearCategory to CommandRegistry**

```go
func (r *CommandRegistry) ClearCategory(cat CommandCategory) {
    r.mu.Lock()
    defer r.mu.Unlock()
    for name, cmd := range r.commands {
        if cmd.Category == cat {
            delete(r.commands, name)
        }
    }
}
```

- [ ] **Step 3: Write failing test for recursive loading**

```go
func TestLoadSkillsDirRecursive(t *testing.T) {
    baseDir := t.TempDir()
    writeFixture(t, filepath.Join(baseDir, "code-review", "SKILL.md"), "---\nname: code-review\ndescription: Review\n---\nbody")
    writeFixture(t, filepath.Join(baseDir, "debug", "test", "SKILL.md"), "---\nname: test\ndescription: Test\n---\nbody")
    writeFixture(t, filepath.Join(baseDir, "debug", "SKILL.md"), "---\nname: debug\ndescription: Debug\n---\nbody")
    r := NewSkillRegistry()
    if err := LoadSkillsDir(r, baseDir); err != nil {
        t.Fatalf("LoadSkillsDir: %v", err)
    }
    if _, ok := r.Get("code-review"); !ok { t.Fatal("code-review not found") }
    if _, ok := r.Get("debug"); !ok { t.Fatal("debug not found") }
    if _, ok := r.Get("test"); !ok { t.Fatal("nested test not found") }
}
```

- [ ] **Step 4: Run test to confirm it fails** (current LoadSkillsDir is flat)
- [ ] **Step 5: Replace LoadSkillsDir with filepath.Walk version**

```go
func LoadSkillsDir(registry *SkillRegistry, dir string) error {
    info, err := os.Stat(dir)
    if err != nil {
        if os.IsNotExist(err) { return nil }
        return err
    }
    if !info.IsDir() { return nil }
    return filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
        if err != nil { return nil }
        if fi.IsDir() || fi.Name() != "SKILL.md" { return nil }
        parentDir := filepath.Base(filepath.Dir(path))
        data, err := os.ReadFile(path)
        if err != nil { return nil }
        name, desc, body, err := parseFrontmatter(string(data))
        if err != nil { return nil }
        if name != parentDir { return nil }
        skillType := SkillPrompt
        if registry.IsHandler(name) { skillType = SkillHandler }
        registry.Register(&Skill{Name: name, Description: desc, Type: skillType, Prompt: body})
        return nil
    })
}
```

- [ ] **Step 6: Run tests: `go test ./internal/service/... -count=1` — ALL pass**
- [ ] **Step 7: Commit**

---

### Task 5: Hot-Reload Skill Polling + ReloadSkills

**Files:**
- Modify: `internal/ui/app.go`
- Modify: `internal/service/skill_loader.go` (add ReloadSkillsDir)
- Modify: `internal/service/skill_loader_test.go` (add test)

- [ ] **Step 1: Add ReloadSkillsDir to skill_loader.go**

```go
func ReloadSkillsDir(registry *SkillRegistry, dir string) error {
    registry.ClearPrompts()
    return LoadSkillsDir(registry, dir)
}
```

- [ ] **Step 2: Add skillsDir to App struct + startSkillWatcher**

Add fields:
```go
skillRegistry *service.SkillRegistry
skillsDir     string
```

Add methods:
```go
func (a *App) SetSkillRegistry(r *service.SkillRegistry) { a.skillRegistry = r }
func (a *App) SetSkillsDir(dir string) { a.skillsDir = dir }

func (a *App) reloadSkills() {
    if a.skillRegistry == nil || a.skillsDir == "" { return }
    if err := service.ReloadSkillsDir(a.skillRegistry, a.skillsDir); err != nil {
        return
    }
    if a.commandRegistry != nil {
        a.commandRegistry.ClearCategory(service.CmdSkill)
        a.commandRegistry.SyncSkills(a.skillRegistry)
    }
}

func (a *App) startSkillWatcher() {
    go func() {
        mtimes := make(map[string]time.Time)
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()
        for range ticker.C {
            changed := false
            filepath.Walk(a.skillsDir, func(path string, fi os.FileInfo, err error) error {
                if err != nil { return nil }
                if fi.IsDir() || fi.Name() != "SKILL.md" { return nil }
                if old, ok := mtimes[path]; !ok || fi.ModTime() != old {
                    changed = true
                    mtimes[path] = fi.ModTime()
                }
                return nil
            })
            if changed {
                a.QueueUpdateDraw(func() { a.reloadSkills() })
            }
        }
    }()
}
```

- [ ] **Step 3: Update main.go to wire everything**

```go
app.SetSkillRegistry(skillRegistry)
app.SetSkillsDir("skills")
app.startSkillWatcher()
```

Place before `app.AddWelcomeMessage()`.

- [ ] **Step 4: Write test for ReloadSkillsDir**

```go
func TestReloadSkillsDir(t *testing.T) {
    baseDir := t.TempDir()
    writeFixture(t, filepath.Join(baseDir, "alpha", "SKILL.md"),
        "---\nname: alpha\ndescription: Alpha\n---\nOriginal")
    r := NewSkillRegistry()
    LoadSkillsDir(r, baseDir)
    if _, ok := r.Get("alpha"); !ok { t.Fatal("alpha not loaded") }

    // Modify existing skill
    writeFixture(t, filepath.Join(baseDir, "alpha", "SKILL.md"),
        "---\nname: alpha\ndescription: Alpha updated\n---\nUpdated body")
    // Add new skill
    writeFixture(t, filepath.Join(baseDir, "beta", "SKILL.md"),
        "---\nname: beta\ndescription: Beta\n---\nNew skill")

    if err := ReloadSkillsDir(r, baseDir); err != nil {
        t.Fatalf("ReloadSkillsDir: %v", err)
    }
    s, ok := r.Get("alpha")
    if !ok { t.Fatal("alpha missing after reload") }
    if s.Prompt != "Updated body" { t.Fatalf("expected updated body, got %q", s.Prompt) }
    if _, ok := r.Get("beta"); !ok { t.Fatal("beta not found after reload") }
}
```

- [ ] **Step 5: Run tests: `go test ./... -count=1` — ALL pass**
- [ ] **Step 6: Build: `go build ./cmd/agent/` — no errors**
- [ ] **Step 7: Commit**

---

### Verification

```bash
cd /Users/luxe/Downloads/agentx/agent5
go build ./cmd/agent/
go test ./... -count=1
```
