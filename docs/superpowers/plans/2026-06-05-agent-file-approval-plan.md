# Agent File Write Approval — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Before the agent's `write_file` tool writes to disk, show the user a diff and require y/n approval.

**Architecture:** The existing `ToolContext.Approval` callback (currently unused) is wired through Agent → ReAct loop → ToolContext. `WriteFileTool` calls it before `os.WriteFile()`, reading existing content for diff comparison. A TUI overlay displays the diff and captures user input.

**Tech Stack:** Go, tview, tcell, standard library `os`, `strings`, `fmt`

---

### Task 1: Create diff utility (`internal/agent/save/diff.go`)

**Files:**
- Create: `internal/agent/save/diff.go`
- Test: `internal/agent/save/diff_test.go`

- [ ] **Step 1: Write the failing test**

```go
package save

import (
	"strings"
	"testing"
)

func TestUnifiedDiffNewFile(t *testing.T) {
	oldContent := ""
	newContent := "line1\nline2\n"
	diff := UnifiedDiff("test.txt", oldContent, newContent)

	if !strings.Contains(diff, "+line1") {
		t.Fatalf("expected '+line1' in diff, got:\n%s", diff)
	}
	if !strings.Contains(diff, "+line2") {
		t.Fatalf("expected '+line2' in diff, got:\n%s", diff)
	}
}

func TestUnifiedDiffModifiedFile(t *testing.T) {
	oldContent := "aaa\nbbb\nccc\n"
	newContent := "aaa\nmodified\nccc\n"
	diff := UnifiedDiff("test.txt", oldContent, newContent)

	if !strings.Contains(diff, "-bbb") {
		t.Fatalf("expected '-bbb' in diff, got:\n%s", diff)
	}
	if !strings.Contains(diff, "+modified") {
		t.Fatalf("expected '+modified' in diff, got:\n%s", diff)
	}
}

func TestUnifiedDiffNoChange(t *testing.T) {
	content := "same\ncontent\n"
	diff := UnifiedDiff("test.txt", content, content)
	if diff != "" {
		t.Fatalf("expected empty diff for unchanged content, got:\n%s", diff)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```
cd G:\mllm\agent5
go test ./internal/agent/save/ -v
```

Expected: FAIL — package doesn't exist yet

- [ ] **Step 3: Create the package and implementation**

Create directory `internal/agent/save/` and file `diff.go`:

```go
package save

import (
	"fmt"
	"strings"
)

func UnifiedDiff(path, oldContent, newContent string) string {
	if oldContent == newContent {
		return ""
	}

	oldLines := strings.SplitAfter(oldContent, "\n")
	newLines := strings.SplitAfter(newContent, "\n")

	if len(oldContent) > 0 && !strings.HasSuffix(oldContent, "\n") {
		oldLines = append(oldLines, "\n\\ No newline at end of file\n")
	}
	if len(newContent) > 0 && !strings.HasSuffix(newContent, "\n") {
		newLines = append(newLines, "\n\\ No newline at end of file\n")
	}

	var b strings.Builder
	fmt.Fprintf(&b, "--- %s\n", path)
	fmt.Fprintf(&b, "+++ %s\n", path)

	maxLen := len(oldLines)
	if len(newLines) > maxLen {
		maxLen = len(newLines)
	}

	for i := 0; i < maxLen; i++ {
		oldLine := ""
		newLine := ""
		if i < len(oldLines) {
			oldLine = oldLines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}
		if oldLine != newLine || i >= len(oldLines) || i >= len(newLines) {
			if i == 0 {
				fmt.Fprintf(&b, "@@ -1,%d +1,%d @@\n", len(oldLines), len(newLines))
			}
			if i < len(oldLines) {
				fmt.Fprintf(&b, "-%s", oldLine)
			}
			if i < len(newLines) {
				fmt.Fprintf(&b, "+%s", newLine)
			}
		}
	}

	result := b.String()
	return result
}
```

- [ ] **Step 4: Run test to verify it passes**

```
cd G:\mllm\agent5
go test ./internal/agent/save/ -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```
git add internal/agent/save/
git commit -m "feat: add unified diff utility for file change preview"
```

---

### Task 2: Change ApprovalFn signature and wire into WriteFileTool

**Files:**
- Modify: `internal/agent/tool/tool.go` (ApprovalFn type signature)
- Modify: `internal/agent/tool/write_file.go` (call Approval before write)
- Modify: `internal/agent/tool/write_file_test.go` (test rejection)
- Test: `internal/agent/tool/`

- [ ] **Step 1: Update ApprovalFn type in tool.go**

Change the `Approval` field from:
```go
Approval func(toolName string, params map[string]interface{}) bool
```
to:
```go
Approval func(toolName string, params map[string]interface{}, oldContent, newContent string) bool
```

- [ ] **Step 2: Update WriteFileTool.Execute() to call Approval**

```go
func (t *WriteFileTool) Execute(ctx ToolContext, params map[string]interface{}) ToolResult {
	path, _ := params["path"].(string)
	content, _ := params["content"].(string)
	if path == "" {
		return ToolResult{Error: "path parameter is required"}
	}

	if ctx.SandboxDir != "" {
		if filepath.IsAbs(path) {
			if !strings.HasPrefix(filepath.Clean(path), filepath.Clean(ctx.SandboxDir)) {
				return ToolResult{Error: fmt.Sprintf("path %q is outside sandbox %q", path, ctx.SandboxDir)}
			}
		} else {
			path = filepath.Join(ctx.SandboxDir, path)
		}
	}

	if ctx.Approval != nil {
		var oldContent string
		if data, err := os.ReadFile(path); err == nil {
			oldContent = string(data)
		}
		if !ctx.Approval(t.Name(), params, oldContent, content) {
			return ToolResult{Error: "file write rejected by user"}
		}
	}

	dir := filepath.Dir(path)
	if dir != "." {
		os.MkdirAll(dir, 0755)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return ToolResult{Error: fmt.Sprintf("write file: %s", err)}
	}
	return ToolResult{Success: true, Data: fmt.Sprintf("wrote %d bytes to %s", len(content), path)}
}
```

- [ ] **Step 3: Write the failing test for approval rejection**

Add to `write_file_test.go`:

```go
func TestWriteFileToolRejected(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "rejected.txt")
	os.WriteFile(file, []byte("original"), 0644)

	rejected := false
	tool := &WriteFileTool{}
	result := tool.Execute(ToolContext{
		Context: context.Background(),
		Approval: func(toolName string, params map[string]interface{}, oldContent, newContent string) bool {
			if toolName != "write_file" {
				t.Fatalf("expected tool name 'write_file', got %s", toolName)
			}
			if oldContent != "original" {
				t.Fatalf("expected old content 'original', got %s", oldContent)
			}
			if newContent != "new content" {
				t.Fatalf("expected new content 'new content', got %s", newContent)
			}
			return false // reject
		},
	}, map[string]interface{}{
		"path":    file,
		"content": "new content",
	})
	if result.Success {
		t.Fatal("expected failure when approval rejects")
	}
	if result.Error != "file write rejected by user" {
		t.Fatalf("expected rejection error, got: %s", result.Error)
	}
	data, _ := os.ReadFile(file)
	if string(data) != "original" {
		t.Fatalf("file should not have been modified, got: %s", string(data))
	}
}

func TestWriteFileToolApproved(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "approved.txt")

	approved := false
	tool := &WriteFileTool{}
	result := tool.Execute(ToolContext{
		Context: context.Background(),
		Approval: func(toolName string, params map[string]interface{}, oldContent, newContent string) bool {
			approved = true
			return true
		},
	}, map[string]interface{}{
		"path":    file,
		"content": "new content",
	})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	if !approved {
		t.Fatal("expected approval callback to be called")
	}
	data, _ := os.ReadFile(file)
	if string(data) != "new content" {
		t.Fatalf("expected 'new content', got %s", string(data))
	}
}
```

- [ ] **Step 4: Run test to verify it fails**

```
cd G:\mllm\agent5
go test ./internal/agent/tool/ -run TestWriteFileToolRejected -v
```

Expected: FAIL — WriteFileTool doesn't call Approval yet

- [ ] **Step 5: Run all tool tests to verify they pass**

```
cd G:\mllm\agent5
go test ./internal/agent/tool/ -v
```

Expected: all PASS

- [ ] **Step 6: Fix existing tests that use ToolContext (they now need Approval field)**

Check if any other files in `internal/agent/tool/` reference `ToolContext{` and update them to compile with the new field (setting `Approval: nil` is sufficient since it's optional).

- [ ] **Step 7: Commit**

```
git add internal/agent/tool/tool.go internal/agent/tool/write_file.go internal/agent/tool/write_file_test.go
git commit -m "feat: add approval callback to write_file tool"
```

---

### Task 3: Wire ApprovalFn through Agent → ReAct loop

**Files:**
- Modify: `internal/agent/runtime/agent.go` (add `ApprovalFn` to Config)
- Modify: `internal/agent/runtime/react.go` (pass ApprovalFn into ToolContext)
- Modify: `internal/agent/runtime/react.go` (stream variant)
- Modify: `internal/agent/runtime/agent_test.go` (update tests)

- [ ] **Step 1: Add ApprovalFn to Config**

```go
type Config struct {
	Name         string
	Model        string
	SystemPrompt string
	MaxReActLoop int
	Temperature  float64
	ContextLimit int
	SandboxDir   string
	ApprovalFn   func(toolName string, params map[string]interface{}, oldContent, newContent string) bool
}
```

- [ ] **Step 2: Pass ApprovalFn into ToolContext in reactLoop**

Change the ToolContext creation in `react.go:38-41`:

```go
tc := tool.ToolContext{
	Context:    nil,
	SandboxDir: a.Config.SandboxDir,
	Approval:   a.Config.ApprovalFn,
}
```

Do the same change in `reactLoopStream` (line 89-92).

- [ ] **Step 3: Update agent_test.go**

Change `NewAgent(Config{...})` calls — no code change needed since ApprovalFn is optional (zero value is nil), but verify the test still compiles.

- [ ] **Step 4: Run agent tests**

```
cd G:\mllm\agent5
go test ./internal/agent/runtime/ -v
```

Expected: PASS

- [ ] **Step 5: Run all agent tests**

```
cd G:\mllm\agent5
go test ./internal/agent/... -v
```

Expected: PASS

- [ ] **Step 6: Commit**

```
git add internal/agent/runtime/
git commit -m "feat: wire approval callback through agent config to tool context"
```

---

### Task 4: Create TUI diff overlay component

**Files:**
- Create: `internal/ui/approval.go`
- Create: `internal/ui/approval_test.go`

- [ ] **Step 1: Write the failing test**

`internal/ui/approval_test.go`:

```go
package ui

import (
	"testing"
)

func TestApprovalModalCreation(t *testing.T) {
	modal := NewApprovalModal()
	if modal == nil {
		t.Fatal("expected non-nil ApprovalModal")
	}
}

func TestApprovalModalSetContent(t *testing.T) {
	modal := NewApprovalModal()
	if modal == nil {
		t.Fatal("expected non-nil ApprovalModal")
	}
	modal.SetContent("test.txt", "--- test.txt\n+++ test.txt\n@@ -1 +1 @@\n-old\n+new\n")
}

func TestApprovalModalResult(t *testing.T) {
	modal := NewApprovalModal()
	modal.SetContent("test.txt", "diff content")
	// Default result should be false (not yet decided)
	if modal.Result() {
		t.Fatal("expected false before approval")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```
cd G:\mllm\agent5
go test ./internal/ui/ -run TestApproval -v
```

Expected: FAIL

- [ ] **Step 3: Create approval.go implementation**

```go
package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ApprovalModal struct {
	*tview.Flex
	path    *tview.TextView
	diff    *tview.TextView
	prompt  *tview.TextView
	result  bool
	decided chan struct{}
}

func NewApprovalModal() *ApprovalModal {
	path := tview.NewTextView()
	path.SetDynamicColors(true)
	path.SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorYellow))

	diff := tview.NewTextView()
	diff.SetDynamicColors(true)
	diff.SetWordWrap(false)
	diff.SetScrollable(true)

	prompt := tview.NewTextView()
	prompt.SetDynamicColors(true)
	prompt.SetTextAlign(tview.AlignCenter)
	prompt.SetText("[::b]Approve? [green]y[white]/[red]n[white]  |  [gray]d[white]: show full diff[-]")

	m := &ApprovalModal{
		Flex:    tview.NewFlex().SetDirection(tview.FlexRow),
		path:    path,
		diff:    diff,
		prompt:  prompt,
		result:  false,
		decided: make(chan struct{}),
	}
	m.SetBorder(true)
	m.SetTitle(" File Write Approval ")
	m.SetBackgroundColor(tcell.ColorDefault)
	m.AddItem(path, 1, 0, false)
	m.AddItem(diff, 0, 1, false)
	m.AddItem(prompt, 1, 0, false)

	m.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Rune() == 'y' || event.Rune() == 'Y':
			m.setResult(true)
		case event.Rune() == 'n' || event.Rune() == 'N' || event.Key() == tcell.KeyEsc:
			m.setResult(false)
		}
		return nil
	})

	return m
}

func (m *ApprovalModal) setResult(v bool) {
	m.result = v
	select {
	case <-m.decided:
	default:
		close(m.decided)
	}
}

func (m *ApprovalModal) SetContent(filePath, diffContent string) {
	m.path.SetText("[yellow]" + filePath + "[-]")
	m.diff.SetText(diffContent)
}

func (m *ApprovalModal) Result() bool {
	return m.result
}

func (m *ApprovalModal) Wait() bool {
	<-m.decided
	return m.result
}
```

- [ ] **Step 4: Run test to verify it passes**

```
cd G:\mllm\agent5
go test ./internal/ui/ -run TestApproval -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```
git add internal/ui/approval.go internal/ui/approval_test.go
git commit -m "feat: add approval modal UI component"
```

---

### Task 5: Wire everything in main.go and app.go

**Files:**
- Modify: `cmd/agent/main.go` (create ApprovalFn closure that shows modal)
- Modify: `internal/ui/app.go` (store reference to approval modal, handle keyboard)

- [ ] **Step 1: Create approval wiring in main.go**

Add to `cmd/agent/main.go`, around where agent configs are created:

```go
app := ui.NewApp()

// Approval overlay is part of the app's pages system
approvalModal := ui.NewApprovalModal()
app.AddApprovalModal(approvalModal)

approvalFn := func(toolName string, params map[string]interface{}, oldContent, newContent string) bool {
	path, _ := params["path"].(string)
	diffContent := save.UnifiedDiff(path, oldContent, newContent)

	// Schedule the modal display on the UI thread
	app.QueueUpdateDraw(func() {
		approvalModal.SetContent(path, diffContent)
		// Show the overlay page
		app.ShowApprovalModal()
	})

	// Wait for user decision (blocks the goroutine, not the UI)
	return approvalModal.Wait()
}

// Pass approvalFn to each agent config
for _, ac := range cfg.AgentRoles {
	if !ac.Enabled {
		continue
	}
	agentCfg := runtime.Config{
		Name:         ac.Name,
		Model:        ac.Model,
		SystemPrompt: ac.SystemPrompt,
		MaxReActLoop: ac.MaxReActLoop,
		SandboxDir:   ac.SandboxDir,
		ApprovalFn:   approvalFn,
	}
	// ... rest unchanged
}
```

- [ ] **Step 2: Add approvalModal field to App struct and wire into pages**

Add to `App` struct in `internal/ui/app.go`:
```go
approvalModal *ApprovalModal
```

In `NewApp()`, after existing page setup:
```go
a.approvalModal = NewApprovalModal()
a.pages.AddPage("approval", a.approvalModal, true, false)
```

Add show/hide methods:
```go
func (a *App) ShowApproval() {
	a.pages.ShowPage("approval")
	a.SetFocus(a.approvalModal)
}

func (a *App) HideApproval() {
	a.pages.HidePage("approval")
}
```

- [ ] **Step 3: Wire approval in main.go**

Add import for `save` package:
```go
"github.com/example/agent-tui/internal/agent/save"
```

After agent configs are loaded but before `orch := orchestrator.NewOrchestrator(...)`:

```go
approvalFn := func(toolName string, params map[string]interface{}, oldContent, newContent string) bool {
	path, _ := params["path"].(string)
	diffContent := save.UnifiedDiff(path, oldContent, newContent)

	app.QueueUpdateDraw(func() {
		app.approvalModal.SetContent(path, diffContent)
		app.ShowApproval()
	})

	return app.approvalModal.Wait()
}
```

Then pass `ApprovalFn: approvalFn` into each `runtime.Config{}` in the agent creation loop.
cd G:\mllm\agent5
go build ./cmd/agent/
```

Expected: no errors

- [ ] **Step 5: Commit**

```
git add cmd/agent/main.go internal/ui/app.go
git commit -m "feat: wire approval modal into application"
```
