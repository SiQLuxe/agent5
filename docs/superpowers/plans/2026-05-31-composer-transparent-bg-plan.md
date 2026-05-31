# Composer Transparent Background Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove the dark background block from the composer input area by setting terminal default background on all composer widgets.

**Architecture:** Single-file change in `internal/ui/composer/composer.go`. Add three `SetBackgroundColor(tcell.ColorDefault)` calls at the end of `New()` — one each for `textArea`, `prompt`, and the root `Flex`. If insufficient (tview still paints a background), `applyTheme()` in `app.go` can also call `composer.SetBackgroundColor(tcell.ColorDefault)` as a second line of defense.

**Tech Stack:** Go, tview, tcell

---

### Task 1: Set terminal default background on composer sub-widgets

**Files:**
- Modify: `internal/ui/composer/composer.go:28-36`

- [ ] **Step 1: Add SetBackgroundColor calls**

Edit `internal/ui/composer/composer.go` — add the three calls after the Flex assembly, before the return:

```go
// At end of New(), before return
textArea.SetBackgroundColor(tcell.ColorDefault)
prompt.SetBackgroundColor(tcell.ColorDefault)
flex.SetBackgroundColor(tcell.ColorDefault)
```

Resulting `New()` function:

```go
func New() *Composer {
	textArea := tview.NewTextArea()
	textArea.SetWordWrap(true)
	textArea.SetSize(8, 0)
	textArea.SetTextStyle(tcell.StyleDefault.Background(tcell.ColorDefault))
	textArea.SetPlaceholderStyle(tcell.StyleDefault.Background(tcell.ColorDefault))
	textArea.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorDefault))

	prompt := tview.NewTextView()
	prompt.SetText("> ")
	prompt.SetDynamicColors(true)

	flex := tview.NewFlex().SetDirection(tview.FlexColumn)
	flex.AddItem(prompt, 2, 0, false)
	flex.AddItem(textArea, 0, 1, true)

	textArea.SetBackgroundColor(tcell.ColorDefault)
	prompt.SetBackgroundColor(tcell.ColorDefault)
	flex.SetBackgroundColor(tcell.ColorDefault)

	return &Composer{
		Flex:     flex,
		textArea: textArea,
		prompt:   prompt,
	}
}
```

- [ ] **Step 2: Build to verify compilation**

Run: `go build ./...`
Expected: exits 0 with no errors

- [ ] **Step 3: Run existing tests**

Run: `go test ./internal/ui/composer/ -v`
Expected: all tests pass

- [ ] **Step 4: Commit**

```bash
git add internal/ui/composer/composer.go
git commit -m "fix: set terminal default background on composer widgets"
```
