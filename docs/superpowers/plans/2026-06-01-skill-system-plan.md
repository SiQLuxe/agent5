# Skill System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a skill system to the agent5 TUI: `/` commands, overlay menu, Tab completion, and agentskills.io-compatible skill loading from `skills/` directory.

**Architecture:** New `service/skill_*.go` files for types, loading, parsing, and execution. New `ui/skill_overlay.go` for the overlay list. `app.go` gains `ModeSkill` state and keyboard handling. `cmd/agent/main.go` loads skills at startup.

**Tech Stack:** Go 1.26, tview, tcell, gopkg.in/yaml.v3 (new dep)

---

### Task 0: Add yaml.v3 dependency

- [ ] **Add yaml.v3 to go.mod**

```bash
go get gopkg.in/yaml.v3 && go mod tidy
```

Verify `gopkg.in/yaml.v3` appears in go.mod.

---

### Task 1: Skill types and SkillRegistry

**Files:**
- Create: `internal/service/skill_registry.go`
- Test: `internal/service/skill_registry_test.go`

- [ ] **Step 1: Write failing test**

```go
package service

import (
	"os"
	"testing"
)

func TestSkillRegistryRegisterAndGet(t *testing.T) {
	r := NewSkillRegistry()
	skill := &Skill{Name: "test-skill", Description: "a test"}
	err := r.Register(skill)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	got, ok := r.Get("test-skill")
	if !ok {
		t.Fatal("Get returned not found")
	}
	if got.Name != "test-skill" {
		t.Fatalf("expected name test-skill, got %s", got.Name)
	}
}

func TestSkillRegistryList(t *testing.T) {
	r := NewSkillRegistry()
	r.Register(&Skill{Name: "a", Description: "skill a"})
	r.Register(&Skill{Name: "b", Description: "skill b"})
	skills := r.List()
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}
}

func TestSkillRegistryDuplicate(t *testing.T) {
	r := NewSkillRegistry()
	r.Register(&Skill{Name: "dup", Description: "dup"})
	err := r.Register(&Skill{Name: "dup", Description: "dup"})
	if err == nil {
		t.Fatal("expected error on duplicate register")
	}
}

func TestSkillRegistryGetNotFound(t *testing.T) {
	r := NewSkillRegistry()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Fatal("expected false for nonexistent skill")
	}
}

func TestSkillRegistryDuplicateErrorIsErrExist(t *testing.T) {
	r := NewSkillRegistry()
	r.Register(&Skill{Name: "dup2", Description: "dup2"})
	err := r.Register(&Skill{Name: "dup2", Description: "dup2"})
	if !os.IsExist(err) {
		t.Fatal("expected os.ErrExist for duplicate")
	}
}

func TestHandlerRegistration(t *testing.T) {
	r := NewSkillRegistry()
	called := false
	r.RegisterHandler("ping", func(ctx SkillContext) string {
		called = true
		return "pong"
	})
	r.Register(&Skill{Name: "ping", Description: "ping test"})
	got, ok := r.Get("ping")
	if !ok {
		t.Fatal("skill ping not found")
	}
	if got.Type != SkillHandler {
		t.Fatalf("expected SkillHandler type, got %v", got.Type)
	}
	if !called {
		// Handler should NOT be called during registration
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/service/ -run TestSkillRegistry\|TestHandlerRegistration -v
```

Expected: FAIL (NewSkillRegistry etc not defined).

- [ ] **Step 3: Write implementation**

New file `internal/service/skill_registry.go`:

```go
package service

import (
	"os"
	"sync"
)

type SkillType int

const (
	SkillPrompt  SkillType = iota
	SkillHandler
)

type Skill struct {
	Name        string
	Description string
	Type        SkillType
	Prompt      string
}

type SkillContext struct {
	Input   string
	SessionID string
	AI      *AIAssistant
}

type SkillRegistry struct {
	mu       sync.RWMutex
	skills   map[string]*Skill
	handlers map[string]func(ctx SkillContext) string
}

func NewSkillRegistry() *SkillRegistry {
	return &SkillRegistry{
		skills:   make(map[string]*Skill),
		handlers: make(map[string]func(ctx SkillContext) string),
	}
}

func (r *SkillRegistry) Register(skill *Skill) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.skills[skill.Name]; exists {
		return os.ErrExist
	}
	// Auto-detect type: if handler is registered, it's SkillHandler
	r.mu.RUnlock()
	_, hasHandler := r.handlers[skill.Name]
	r.mu.RLock()
	if hasHandler {
		skill.Type = SkillHandler
	}
	r.skills[skill.Name] = skill
	return nil
}

func (r *SkillRegistry) Get(name string) (*Skill, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.skills[name]
	return s, ok
}

func (r *SkillRegistry) List() []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Skill, 0, len(r.skills))
	for _, s := range r.skills {
		out = append(out, s)
	}
	return out
}

func (r *SkillRegistry) RegisterHandler(name string, fn func(ctx SkillContext) string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[name] = fn
	// Update existing skill type
	if s, ok := r.skills[name]; ok {
		s.Type = SkillHandler
	}
}

func (r *SkillRegistry) GetHandler(name string) (func(ctx SkillContext) string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fn, ok := r.handlers[name]
	return fn, ok
}

func (r *SkillRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.skills, name)
	delete(r.handlers, name)
}

func (r *SkillRegistry) IsHandler(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.handlers[name]
	return ok
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/service/ -run TestSkillRegistry\|TestHandlerRegistration -v
```

Expected: All PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/service/skill_registry.go internal/service/skill_registry_test.go
git commit -m "feat: add SkillRegistry with prompt/handler types"
```

---

### Task 2: SKILL.md loader (frontmatter + body)

**Files:**
- Create: `internal/service/skill_loader.go`
- Test: `internal/service/skill_loader_test.go`

- [ ] **Step 1: Write failing test**

```go
package service

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFixture(t *testing.T, path, content string) {
	t.Helper()
	os.MkdirAll(filepath.Dir(path), 0755)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadSkillsFromDir(t *testing.T) {
	baseDir := t.TempDir()
	writeFixture(t, filepath.Join(baseDir, "code-review", "SKILL.md"),
		"---\nname: code-review\ndescription: Review code for bugs\n---\nYou are a code reviewer.")

	r := NewSkillRegistry()
	if err := LoadSkillsDir(r, baseDir); err != nil {
		t.Fatalf("LoadSkillsDir: %v", err)
	}

	s, ok := r.Get("code-review")
	if !ok {
		t.Fatal("code-review not registered")
	}
	if s.Name != "code-review" {
		t.Fatalf("name: got %q", s.Name)
	}
	if s.Description != "Review code for bugs" {
		t.Fatalf("desc: got %q", s.Description)
	}
	if s.Prompt != "You are a code reviewer." {
		t.Fatalf("prompt: got %q", s.Prompt)
	}
	if s.Type != SkillPrompt {
		t.Fatalf("expected SkillPrompt")
	}
}

func TestLoadSkillsInvalidFrontmatter(t *testing.T) {
	baseDir := t.TempDir()
	writeFixture(t, filepath.Join(baseDir, "bad", "SKILL.md"), "no frontmatter here")

	r := NewSkillRegistry()
	if err := LoadSkillsDir(r, baseDir); err != nil {
		t.Fatalf("LoadSkillsDir: %v", err)
	}
	if _, ok := r.Get("bad"); ok {
		t.Fatal("bad skill should not register")
	}
}

func TestLoadSkillsMissingDir(t *testing.T) {
	r := NewSkillRegistry()
	if err := LoadSkillsDir(r, "/nonexistent/path/xyz"); err != nil {
		t.Fatalf("missing dir: %v", err)
	}
}

func TestLoadSkillsNameMismatch(t *testing.T) {
	baseDir := t.TempDir()
	writeFixture(t, filepath.Join(baseDir, "code-review", "SKILL.md"),
		"---\nname: review\ndescription: review\n---\nbody")

	r := NewSkillRegistry()
	if err := LoadSkillsDir(r, baseDir); err != nil {
		t.Fatalf("LoadSkillsDir: %v", err)
	}
	if _, ok := r.Get("review"); ok {
		t.Fatal("should not register when name != dir")
	}
}

func TestParseFrontmatter(t *testing.T) {
	input := "---\nname: my-skill\ndescription: My skill\n---\nBody text"
	name, desc, body, err := parseFrontmatter(input)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if name != "my-skill" {
		t.Fatalf("name: %q", name)
	}
	if desc != "My skill" {
		t.Fatalf("desc: %q", desc)
	}
	if body != "Body text" {
		t.Fatalf("body: %q", body)
	}
}

func TestParseFrontmatterNoDelimiter(t *testing.T) {
	_, _, _, err := parseFrontmatter("no delimiters")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadSkillsAutoDetectHandler(t *testing.T) {
	baseDir := t.TempDir()
	writeFixture(t, filepath.Join(baseDir, "ping", "SKILL.md"),
		"---\nname: ping\ndescription: Ping test\n---\nping body")

	r := NewSkillRegistry()
	r.RegisterHandler("ping", func(ctx SkillContext) string { return "pong" })
	if err := LoadSkillsDir(r, baseDir); err != nil {
		t.Fatalf("LoadSkillsDir: %v", err)
	}
	s, ok := r.Get("ping")
	if !ok {
		t.Fatal("ping not found")
	}
	if s.Type != SkillHandler {
		t.Fatalf("expected SkillHandler, got %v", s.Type)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/service/ -run TestLoadSkills\|TestParseFrontmatter -v
```

Expected: FAIL.

- [ ] **Step 3: Write implementation to `internal/service/skill_loader.go`**

```go
package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type skillFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

func LoadSkillsDir(registry *SkillRegistry, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillPath := filepath.Join(dir, entry.Name(), "SKILL.md")
		data, err := os.ReadFile(skillPath)
		if err != nil {
			continue
		}
		name, desc, body, err := parseFrontmatter(string(data))
		if err != nil {
			continue
		}
		if name != entry.Name() {
			continue
		}
		skillType := SkillPrompt
		if registry.IsHandler(name) {
			skillType = SkillHandler
		}
		registry.Register(&Skill{
			Name:        name,
			Description: desc,
			Type:        skillType,
			Prompt:      body,
		})
	}
	return nil
}

func parseFrontmatter(input string) (name, description, body string, err error) {
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "---") {
		return "", "", "", fmt.Errorf("missing opening ---")
	}
	rest := input[3:]
	idx := strings.Index(rest, "---")
	if idx < 0 {
		return "", "", "", fmt.Errorf("missing closing ---")
	}
	yamlPart := strings.TrimSpace(rest[:idx])
	body = strings.TrimSpace(rest[idx+3:])

	var fm skillFrontmatter
	if err := yaml.Unmarshal([]byte(yamlPart), &fm); err != nil {
		return "", "", "", err
	}
	if fm.Name == "" || fm.Description == "" {
		return "", "", "", fmt.Errorf("name and description required")
	}
	return fm.Name, fm.Description, body, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/service/ -run TestLoadSkills\|TestParseFrontmatter -v
```

Expected: All PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/service/skill_loader.go internal/service/skill_loader_test.go
git commit -m "feat: add SKILL.md loader with frontmatter parsing"
```

---

### Task 3: Command parser

**Files:**
- Create: `internal/service/skill_parser.go`
- Test: `internal/service/skill_parser_test.go`

- [ ] **Step 1: Write failing test**

```go
package service

import "testing"

func TestParseCommandBasic(t *testing.T) {
	cmd := ParseCommand("/review")
	if cmd == nil {
		t.Fatal("nil result")
	}
	if cmd.Name != "review" {
		t.Fatalf("name: got %q", cmd.Name)
	}
	if cmd.Args != "" {
		t.Fatalf("args: got %q", cmd.Args)
	}
}

func TestParseCommandWithArgs(t *testing.T) {
	cmd := ParseCommand("/review main.go")
	if cmd == nil {
		t.Fatal("nil")
	}
	if cmd.Name != "review" {
		t.Fatalf("name: %q", cmd.Name)
	}
	if cmd.Args != "main.go" {
		t.Fatalf("args: %q", cmd.Args)
	}
}

func TestParseCommandMultiWord(t *testing.T) {
	cmd := ParseCommand("/exec go run main.go")
	if cmd.Name != "exec" {
		t.Fatalf("name: %q", cmd.Name)
	}
	if cmd.Args != "go run main.go" {
		t.Fatalf("args: %q", cmd.Args)
	}
}

func TestParseCommandNoSlash(t *testing.T) {
	if cmd := ParseCommand("hello"); cmd != nil {
		t.Fatal("expected nil")
	}
}

func TestParseCommandEmpty(t *testing.T) {
	if cmd := ParseCommand(""); cmd != nil {
		t.Fatal("expected nil")
	}
}

func TestParseCommandJustSlash(t *testing.T) {
	if cmd := ParseCommand("/"); cmd != nil {
		t.Fatal("expected nil")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/service/ -run TestParseCommand -v
```

Expected: FAIL.

- [ ] **Step 3: Write implementation to `internal/service/skill_parser.go`**

```go
package service

import "strings"

type ParsedCommand struct {
	Name string
	Args string
}

func ParseCommand(input string) *ParsedCommand {
	input = strings.TrimSpace(input)
	if len(input) < 2 || input[0] != '/' {
		return nil
	}
	rest := input[1:]
	parts := strings.Fields(rest)
	if len(parts) == 0 {
		return nil
	}
	result := &ParsedCommand{Name: parts[0]}
	if spaceIdx := strings.Index(rest, " "); spaceIdx >= 0 {
		result.Args = strings.TrimSpace(rest[spaceIdx+1:])
	}
	return result
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/service/ -run TestParseCommand -v
```

Expected: All PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/service/skill_parser.go internal/service/skill_parser_test.go
git commit -m "feat: add /command parser"
```

---

### Task 4: SkillExecutor

**Files:**
- Create: `internal/service/skill_executor.go`
- Test: `internal/service/skill_executor_test.go`

- [ ] **Step 1: Write failing test**

```go
package service

import "testing"

func TestExecutePromptSkill(t *testing.T) {
	r := NewSkillRegistry()
	r.Register(&Skill{Name: "review", Description: "review", Type: SkillPrompt, Prompt: "Review this: %s"})
	ex := NewSkillExecutor(r, nil)
	result, err := ex.Execute(&ParsedCommand{Name: "review", Args: "main.go"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if result != "Review this: main.go" {
		t.Fatalf("got %q", result)
	}
}

func TestExecuteHandlerSkill(t *testing.T) {
	r := NewSkillRegistry()
	r.RegisterHandler("ping", func(ctx SkillContext) string {
		return "pong:" + ctx.Input
	})
	r.Register(&Skill{Name: "ping", Description: "ping test"})
	ex := NewSkillExecutor(r, nil)
	result, err := ex.Execute(&ParsedCommand{Name: "ping", Args: "hello"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if result != "pong:hello" {
		t.Fatalf("got %q", result)
	}
}

func TestExecuteNotFound(t *testing.T) {
	r := NewSkillRegistry{}
	ex := NewSkillExecutor(r, nil)
	_, err := ex.Execute(&ParsedCommand{Name: "nope"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExecutePromptWithoutArgs(t *testing.T) {
	r := NewSkillRegistry{}
	r.Register(&Skill{Name: "hello", Description: "hello", Type: SkillPrompt, Prompt: "Say hello"})
	ex := NewSkillExecutor(r, nil)
	result, err := ex.Execute(&ParsedCommand{Name: "hello"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if result != "Say hello" {
		t.Fatalf("got %q", result)
	}
}

func TestExecutePromptNoPercentS(t *testing.T) {
	r := NewSkillRegistry{}
	r.Register(&Skill{Name: "greet", Description: "greet", Type: SkillPrompt, Prompt: "Hello there"})
	ex := NewSkillExecutor(r, nil)
	result, err := ex.Execute(&ParsedCommand{Name: "greet", Args: "world"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if result != "Hello there" {
		t.Fatalf("got %q", result)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/service/ -run TestExecute -v
```

Expected: FAIL.

- [ ] **Step 3: Write implementation to `internal/service/skill_executor.go`**

```go
package service

import (
	"fmt"
	"strings"
)

type SkillExecutor struct {
	registry    *SkillRegistry
	aiAssistant *AIAssistant
}

func NewSkillExecutor(registry *SkillRegistry, ai *AIAssistant) *SkillExecutor {
	return &SkillExecutor{registry: registry, aiAssistant: ai}
}

func (e *SkillExecutor) Execute(cmd *ParsedCommand) (string, error) {
	skill, ok := e.registry.Get(cmd.Name)
	if !ok {
		return "", fmt.Errorf("skill not found: %s", cmd.Name)
	}

	switch skill.Type {
	case SkillHandler:
		fn, ok := e.registry.GetHandler(skill.Name)
		if !ok {
			return "", fmt.Errorf("no handler for: %s", skill.Name)
		}
		return fn(SkillContext{Input: cmd.Args}), nil

	case SkillPrompt:
		prompt := skill.Prompt
		if cmd.Args != "" && strings.Contains(prompt, "%s") {
			prompt = strings.ReplaceAll(prompt, "%s", cmd.Args)
		}
		if e.aiAssistant != nil {
			return e.aiAssistant.Chat("skill-"+skill.Name, prompt)
		}
		return prompt, nil
	}

	return "", fmt.Errorf("unknown skill type")
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/service/ -run TestExecute -v
```

Expected: All PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/service/skill_executor.go internal/service/skill_executor_test.go
git commit -m "feat: add SkillExecutor for prompt/handler dispatch"
```

---

### Task 5: Add RoleSkill to session.go

**Files:**
- Modify: `internal/ui/session.go`

- [ ] **Step 1: Add `RoleSkill` to Role constants and rendering**

In `internal/ui/session.go`:

```go
const (
	RoleUser Role = iota
	RoleAssistant
	RoleSystem
	RoleSkill
)
```

Update `Role.String()`:

```go
case RoleSkill:
	return "skill"
```

Add `Label` field to `Message` struct:

```go
type Message struct {
	Role      Role
	Content   string
	Label     string
	Thinking  *Thinking
	Timestamp time.Time
	Collapsed bool
	lineCount int
}
```

Add `RoleSkill` render case in `renderMessageToBuilder`:

```go
case RoleSkill:
	badge := fmt.Sprintf("[%s:%s:b] \u2699 %s [-:-:-]", theme.Accent, theme.PanelBg, msg.Label)
	sb.WriteString(badge)
	sb.WriteString(" ")
	sb.WriteString(ts)
	sb.WriteString("\n")
	renderContent(sb, msg.Content, false, width, theme, 1)
```

- [ ] **Step 2: Build and run existing tests**

```bash
go build ./...
go test ./internal/ui/ -run TestSession\|TestRender -v
```

Expected: All PASS (existing tests use Role constants, not iota values directly).

- [ ] **Step 3: Commit**

```bash
git add internal/ui/session.go
git commit -m "feat: add RoleSkill for skill execution messages"
```

---

### Task 6: SkillOverlay UI component

**Files:**
- Create: `internal/ui/skill_overlay.go`

- [ ] **Step 1: Write implementation**

```go
package ui

import (
	"github.com/example/agent-tui/internal/service"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type SkillOverlay struct {
	*tview.Flex
	list       *tview.List
	filterText string
	skills     []*service.Skill
}

func NewSkillOverlay() *SkillOverlay {
	list := tview.NewList()
	list.SetMainTextColor(tcell.ColorWhite)
	list.SetSecondaryTextColor(tcell.ColorGray)
	list.SetSelectedBackgroundColor(tcell.ColorDarkCyan)
	list.ShowSecondaryText(true)

	o := &SkillOverlay{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
		list: list,
	}
	o.AddItem(list, 0, 1, true)
	o.SetBackgroundColor(tcell.ColorDefault)
	return o
}

func (o *SkillOverlay) SetSkills(skills []*service.Skill) {
	o.skills = skills
	o.applyFilter()
}

func (o *SkillOverlay) SetFilter(text string) {
	o.filterText = text
	o.applyFilter()
}

func (o *SkillOverlay) applyFilter() {
	o.list.Clear()
	for _, s := range o.skills {
		if o.filterText != "" && !hasPrefixFold(s.Name, o.filterText) {
			continue
		}
		desc := s.Description
		if len([]rune(desc)) > 50 {
			desc = string([]rune(desc)[:50]) + "..."
		}
		o.list.AddItem(s.Name, desc, 0, nil)
	}
}

func (o *SkillOverlay) SelectedSkillName() string {
	if o.list.GetItemCount() == 0 {
		return ""
	}
	main, _ := o.list.GetItemText(o.list.GetCurrentItem())
	return main
}

func (o *SkillOverlay) GetItemCount() int {
	return o.list.GetItemCount()
}

func hasPrefixFold(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return s[:len(prefix)] == prefix
}
```

- [ ] **Step 2: Build to verify compilation**

```bash
go build ./...
```

Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add internal/ui/skill_overlay.go
git commit -m "feat: add SkillOverlay UI component"
```

---

### Task 7: Wire ModeSkill into App

**Files:**
- Modify: `internal/ui/app.go`
- Test: `internal/ui/app_test.go`

This is the core integration task. Changes to `app.go`:

#### 7a: Add ModeSkill and new fields

```go
const (
	ModeChat AppMode = iota
	ModeSearch
	ModeHelp
	ModeSkill
)

// In App struct, add:
skillOverlay   *SkillOverlay
skillRegistry  *service.SkillRegistry
skillExecutor  *service.SkillExecutor
pendingSlash   bool
```

#### 7b: Create overlay in NewApp()

In `NewApp()`, after creating chatFlex:

```go
a.skillOverlay = NewSkillOverlay()
a.chatFlex.AddItem(a.skillOverlay, 0, 0, false) // hidden initially

// Ctrl+P shortcut: already handled in handleInput below
```

#### 7c: Add Skill setter

```go
func (a *App) SetSkillRegistry(registry *service.SkillRegistry) {
	a.skillRegistry = registry
	a.skillExecutor = service.NewSkillExecutor(registry, a.aiAssistant)
}
```

Also add AIAssistant setter integration — update `SetAIAssistant`:

```go
func (a *App) SetAIAssistant(ai *service.AIAssistant) {
	a.aiAssistant = ai
	if a.skillExecutor != nil {
		a.skillExecutor = service.NewSkillExecutor(a.skillRegistry, ai)
	}
}
```

#### 7d: Add ModeSkill input handling

In `handleInput()`, before the chat mode switch:

```go
// Pending slash check (from previous event)
if a.pendingSlash {
	a.pendingSlash = false
	text := a.composer.GetInput()
	if strings.HasPrefix(text, "/") {
		a.enterSkill()
		return nil
	}
}
```

In the chat mode shortcuts (before `switch` block or inside it):

```go
case event.Key() == tcell.KeyCtrlP:
	a.enterSkill()
	return nil
```

#### 7e: Add ModeSkill section in handleInput

After the `ModeSearch` / `ModeHelp` sections:

```go
case ModeSkill:
	// If the overlay has items, Enter executes the selected skill
	if event.Key() == tcell.KeyEnter {
		name := a.skillOverlay.SelectedSkillName()
		if name != "" {
			a.executeSkill(name)
		}
		return nil
	}
	if event.Key() == tcell.KeyEsc {
		a.exitSkill()
		return nil
	}
	if event.Key() == tcell.KeyTab {
		name := a.skillOverlay.SelectedSkillName()
		if name != "" {
			a.composer.SetInput("/" + name)
			a.executeSkill(name)
		}
		return nil
	}
	// Let events fall through to the overlay list
	return event
```

#### 7f: ModeSkill enter/exit/execute methods

```go
func (a *App) enterSkill() {
	a.mode = ModeSkill
	a.pendingSlash = false
	if a.skillRegistry != nil {
		a.skillOverlay.SetSkills(a.skillRegistry.List())
	}
	// Show overlay (set height to 5 rows)
	a.chatFlex.RemoveItem(a.skillOverlay)
	a.chatFlex.AddItem(a.skillOverlay, 5, 0, false)
	a.SetFocus(a.skillOverlay)
}

func (a *App) exitSkill() {
	a.mode = ModeChat
	// Hide overlay
	a.chatFlex.RemoveItem(a.skillOverlay)
	a.chatFlex.AddItem(a.skillOverlay, 0, 0, false)
	a.SetFocus(a.composer)
}

func (a *App) executeSkill(name string) {
	a.mode = ModeChat
	a.chatFlex.RemoveItem(a.skillOverlay)
	a.chatFlex.AddItem(a.skillOverlay, 0, 0, false)
	a.SetFocus(a.composer)

	if a.skillExecutor == nil || a.activeSession < 0 {
		return
	}

	s := a.sessions[a.activeSession]
	text := a.composer.GetInput()
	cmd := service.ParseCommand(text)
	if cmd == nil {
		cmd = &service.ParsedCommand{Name: name}
	}

	result, err := a.skillExecutor.Execute(cmd)
	if err != nil {
		s.AddMessage(ui.RoleSkill, "Error: "+err.Error())
	} else {
		s.AddMessage(ui.RoleSkill, result)
	}
	s.Messages[len(s.Messages)-1].Label = name
	a.composer.ClearInput()
	a.chatPanel.SetSession(s)
}
```

#### 7g: Wire / key detection in handleInput

In the chat mode event processing, add before the final `return event`:

```go
case event.Rune() == '/' && a.mode == ModeChat:
	a.pendingSlash = true
	return event // let it reach composer
```

#### 7h: Write/update test

```go
func TestModeSkillCtrlP(t *testing.T) {
	a := NewApp()
	ev := tcell.NewEventKey(tcell.KeyCtrlP, 0, tcell.ModNone)
	result := a.handleInput(ev)
	if result != nil {
		t.Fatal("expected Ctrl+P consumed (nil)")
	}
	if a.mode != ModeSkill {
		t.Fatalf("expected ModeSkill, got %d", a.mode)
	}
}

func TestModeSkillEscExits(t *testing.T) {
	a := NewApp()
	a.enterSkill()
	ev := tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone)
	result := a.handleInput(ev)
	if result != nil {
		t.Fatal("expected Esc consumed (nil)")
	}
	if a.mode != ModeChat {
		t.Fatalf("expected ModeChat, got %d", a.mode)
	}
}
```

- [ ] **Step 1: Write the test** (above), run it:

```bash
go test ./internal/ui/ -run TestModeSkill -v
```

Expected: FAIL (functions not yet implemented).

- [ ] **Step 2: Implement all changes to `internal/ui/app.go`** as described above.

- [ ] **Step 3: Run tests to verify**

```bash
go test ./internal/ui/ -run TestModeSkill\|TestNewApp\|TestShortcut -v
```

Expected: All PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/ui/app.go internal/ui/app_test.go
git commit -m "feat: wire ModeSkill, / detection, Ctrl+P into App"
```

---

### Task 8: Wire skill loading into main.go

**Files:**
- Modify: `cmd/agent/main.go`
- Create: `skills/code-review/SKILL.md` (sample skill)

- [ ] **Step 1: Update `cmd/agent/main.go`**

```go
func main() {
	cfg, err := config.LoadConfig("configs/config.toml")
	if err != nil {
		cfg = config.GetDefaultConfig()
	}
	if cfg.DefaultClient == "" {
		cfg.DefaultClient = "local"
	}

	aiClient, err := ai.NewClientFromConfig(cfg)
	if err != nil {
		log.Fatalf("failed to create AI client: %v", err)
	}

	h := history.NewHistory("")
	aiAssistant := service.NewAIAssistant(aiClient, h)

	app := ui.NewApp()
	app.SetAIAssistant(aiAssistant)

	// Load skills
	skillRegistry := service.NewSkillRegistry()
	if err := service.LoadSkillsDir(skillRegistry, "skills"); err != nil {
		log.Printf("warning: loading skills: %v", err)
	}
	app.SetSkillRegistry(skillRegistry)

	app.AddWelcomeMessage()
	if err := app.Run(); err != nil {
		log.Fatalf("application error: %v", err)
	}
}
```

- [ ] **Step 2: Create a sample prompt skill**

`skills/code-review/SKILL.md`:

```markdown
---
name: code-review
description: Review Go source code for bugs, performance issues, and style problems.
---
You are a senior Go code reviewer. Review the following code and identify:
1. Bugs and concurrency issues
2. Performance problems
3. Code style and readability concerns
4. Missing error handling

For each issue, provide the file/line, the problem, and a suggested fix.
```

- [ ] **Step 3: Build and run**

```bash
go build -o agent5 ./cmd/agent/
./agent5
```

Expected: App launches. Press Ctrl+P to see the skill overlay with `code-review` listed. Select it and execute.

- [ ] **Step 4: Commit**

```bash
git add cmd/agent/main.go skills/code-review/SKILL.md
git commit -m "feat: wire skill loading into main.go with sample skill"
```

---

### Task 9: Clean up old PluginRegistry

**Files:**
- Remove: `internal/service/plugin_registry.go`
- Remove: `internal/service/collaboration_manager.go` and its test (if unused)
- Remove: `internal/service/collaboration_manager_test.go`
- Remove: `plugins/` directory
- Remove: Any imports of `plugin_registry` in other files

- [ ] **Step 1: Check for importers**

```bash
rg "plugin_registry\|PluginRegistry" --type go
```

Expected: Only internal references within the service package.

- [ ] **Step 2: Remove plugin_registry.go**

```bash
rm internal/service/plugin_registry.go
```

- [ ] **Step 3: Remove plugins directory**

```bash
rm -rf plugins/
```

- [ ] **Step 4: Build to verify**

```bash
go build ./...
```

Expected: No errors.

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "cleanup: remove old PluginRegistry and plugins dir"
```
