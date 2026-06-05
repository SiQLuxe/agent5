# Mouse Wheel Scrolls ChatPanel Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make mouse wheel scroll the ChatPanel content instead of navigating input history.

**Architecture:** Enable terminal mouse tracking (`EnableMouse(true)`) and intercept `MouseWheelUp`/`MouseWheelDown` at application level via `SetMouseCapture`, routing to ChatPanel scroll methods.

**Tech Stack:** Go, tview, tcell

---

### Task 1: Add mouse capture to App

**Files:**
- Modify: `internal/ui/app.go:166-168`

- [ ] **Step 1: Add EnableMouse and SetMouseCapture**

Insert after line 167 (`a.SetInputCapture(a.handleInput)`):

```go
	a.EnableMouse(true)
	a.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		switch action {
		case tview.MouseWheelUp:
			a.chatPanel.SetAutoScroll(false)
			a.chatPanel.ScrollUp(3)
			return 0, nil
		case tview.MouseWheelDown:
			a.chatPanel.ScrollDown(3)
			return 0, nil
		}
		return action, event
	})
```

Make sure `"github.com/gdamore/tcell/v2"` is imported (it already is at line 9).

- [ ] **Step 2: Build**

Run: `go build ./cmd/agent`
Expected: binary compiles clean, no errors.

- [ ] **Step 3: Run tests**

Run: `go test ./internal/ui/... -count=1 -short`
Expected: all tests pass.

- [ ] **Step 4: Commit**

```bash
git add internal/ui/app.go
git commit -m "feat: mouse wheel scrolls ChatPanel instead of navigating input history"
```
