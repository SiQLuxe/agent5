# SelectableTextView Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a custom tview widget `SelectableTextView` supporting mouse text selection, copy, and select-all, then wire it into `ChatPanel`.

**Architecture:** Embed `*tview.Box`, implement `tview.Primitive` interface with custom `Draw()`, `MouseHandler()`, `InputHandler()`. Replace `*tview.TextView` in `ChatPanel` - no external API changes.

**Tech Stack:** Go 1.26.3, `github.com/rivo/tview`, `github.com/gdamore/tcell/v2`

---

### Task 1: Create SelectableTextView with text rendering

**Files:**
- Create: `internal/ui/selectable_textview.go`
- Test: `internal/ui/selectable_textview_test.go`

- [ ] **Step 1: Write failing test for basic rendering**

```go
package ui

import (
    "testing"
    "github.com/gdamore/tcell/v2"
)

func TestSelectableTextView_RendersText(t *testing.T) {
    tv := NewSelectableTextView()
    tv.SetText("hello")
    // Nothing visible to assert without a screen,
    // but ensure no panic and GetText works
    if tv.GetText(false) != "hello" {
        t.Fatalf("expected 'hello', got %q", tv.GetText(false))
    }
}
```

Run: `go test ./internal/ui/ -run TestSelectableTextView_RendersText -v`
Expected: FAIL — `NewSelectableTextView` not defined

- [ ] **Step 2: Implement basic struct and text storage**

```go
package ui

import (
    "github.com/gdamore/tcell/v2"
    "github.com/rivo/tview"
)

type SelectableTextView struct {
    *tview.Box
    text          string
    dynamicColors bool
    regions       bool
    scrollable    bool
    wordWrap      bool
    maxLines      int

    selecting         bool
    selectAnchorRow   int
    selectAnchorCol   int
    selectEndRow      int
    selectEndCol      int
    selectionVisible  bool

    highlights []string
}

func NewSelectableTextView() *SelectableTextView {
    t := &SelectableTextView{
        Box:        tview.NewBox(),
        scrollable: true,
        wordWrap:   true,
    }
    t.SetBackgroundColor(tcell.ColorDefault)
    return t
}

func (t *SelectableTextView) SetText(text string) *SelectableTextView {
    t.text = text
    return t
}

func (t *SelectableTextView) GetText(stripAllTags bool) string {
    if stripAllTags {
        return tview.StripTags(t.text)
    }
    return t.text
}

func (t *SelectableTextView) SetDynamicColors(dynamic bool) *SelectableTextView {
    t.dynamicColors = dynamic
    return t
}

func (t *SelectableTextView) SetScrollable(scrollable bool) *SelectableTextView {
    t.scrollable = scrollable
    return t
}

func (t *SelectableTextView) SetWordWrap(wrap bool) *SelectableTextView {
    t.wordWrap = wrap
    return t
}

func (t *SelectableTextView) SetRegions(regions bool) *SelectableTextView {
    t.regions = regions
    return t
}

func (t *SelectableTextView) SetTextStyle(style tcell.Style) *SelectableTextView {
    return t
}
```

- [ ] **Step 3: Run test to see it pass**

Run: `go test ./internal/ui/ -run TestSelectableTextView_RendersText -v`
Expected: PASS

- [ ] **Step 4: Write failing test for Draw (dummy screen)**

```go
func TestSelectableTextView_Draw(t *testing.T) {
    tv := NewSelectableTextView()
    tv.SetRect(0, 0, 20, 5)
    tv.SetText("hello world")
    screen := tcell.NewSimulationScreen("")
    if err := screen.Init(); err != nil {
        t.Fatal(err)
    }
    defer screen.Fini()
    screen.SetSize(20, 5)

    tv.Draw(screen)
    screen.Show()

    contents := screen.GetContents()
    if len(contents) == 0 {
        t.Fatal("expected screen contents")
    }
}
```

Run: `go test ./internal/ui/ -run TestSelectableTextView_Draw -v`
Expected: FAIL — `Draw` method doesn't render yet

- [ ] **Step 5: Implement basic Draw method**

```go
func (t *SelectableTextView) Draw(screen tcell.Screen) {
    t.Box.Draw(screen)
    if t.text == "" {
        return
    }
    x, y, width, height := t.Box.GetInnerRect()
    if width <= 0 || height <= 0 {
        return
    }

    lines := t.splitLines(width)
    style := tcell.StyleDefault.Background(t.GetBackgroundColor())

    for lineIdx := 0; lineIdx < len(lines) && lineIdx < height; lineIdx++ {
        line := lines[lineIdx]
        drawX := x
        for col := 0; col < len(line) && drawX < x+width; col++ {
            ch := line[col]
            cellStyle := style
            if t.selectionVisible && t.isInSelection(lineIdx, col) {
                cellStyle = cellStyle.Reverse(true)
            }
            screen.SetContent(drawX, y+lineIdx, ch, nil, cellStyle)
            drawX++
        }
    }
}

func (t *SelectableTextView) splitLines(width int) [][]rune {
    if t.text == "" {
        return nil
    }
    rawLines := strings.Split(t.text, "\n")
    var result [][]rune
    for _, raw := range rawLines {
        runes := []rune(raw)
        if width <= 0 || !t.wordWrap {
            result = append(result, runes)
            continue
        }
        if len(runes) <= width {
            result = append(result, runes)
            continue
        }
        // Word-wrap: break at space boundaries when possible
        start := 0
        for start < len(runes) {
            end := start + width
            if end > len(runes) {
                end = len(runes)
            }
            // Look backwards for a space to break at
            breakAt := end
            if end < len(runes) {
                for j := end; j > start; j-- {
                    if runes[j-1] == ' ' {
                        breakAt = j
                        break
                    }
                }
            }
            result = append(result, runes[start:breakAt])
            // Skip trailing space on next line
            start = breakAt
            if start < len(runes) && runes[start] == ' ' {
                start++
            }
        }
    }
    return result
}

func (t *SelectableTextView) isInSelection(lineIdx, colIdx int) bool {
    anchorRow, anchorCol := t.selectAnchorRow, t.selectAnchorCol
    endRow, endCol := t.selectEndRow, t.selectEndCol
    if !t.selectionVisible {
        return false
    }
    startRow, startCol := anchorRow, anchorCol
    endRow2, endCol2 := endRow, endCol
    if startRow > endRow2 || (startRow == endRow2 && startCol > endCol2) {
        startRow, startCol = endRow, endCol
        endRow2, endCol2 = anchorRow, anchorCol
    }
    if lineIdx < startRow || lineIdx > endRow2 {
        return false
    }
    if lineIdx == startRow && lineIdx == endRow2 {
        return colIdx >= startCol && colIdx <= endCol2
    }
    if lineIdx == startRow {
        return colIdx >= startCol
    }
    if lineIdx == endRow2 {
        return colIdx <= endCol2
    }
    return true
}
```

Add import: `"strings"` to the import block.

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test ./internal/ui/ -run "TestSelectableTextView" -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/ui/selectable_textview.go internal/ui/selectable_textview_test.go
git commit -m "feat: add SelectableTextView widget with text rendering"
```

---

### Task 2: Add dynamic colors and region/highlight support

**Files:**
- Modify: `internal/ui/selectable_textview.go`
- Test: `internal/ui/selectable_textview_test.go`

- [ ] **Step 1: Write failing test for color rendering**

```go
func TestSelectableTextView_ColorTags(t *testing.T) {
    tv := NewSelectableTextView()
    tv.SetDynamicColors(true)
    tv.SetText(`[red]hello[-] world`)
    // GetText with stripAllTags should strip color tags
    result := tv.GetText(true)
    if result != "hello world" {
        t.Fatalf("expected 'hello world', got %q", result)
    }
}
```

Run: `go test ./internal/ui/ -run TestSelectableTextView_ColorTags -v`
Expected: FAIL — `GetText(true)` doesn't strip tags yet

- [ ] **Step 2: Implement color tag support and region parsing**

Create a small helper type for styled cells:

```go
// At top of file, add these types and helpers

type cellInfo struct {
    ch    rune
    style tcell.Style
}

func (t *SelectableTextView) parseCells(width int) [][]cellInfo {
    text := t.text
    // Handle ANSI
    if t.dynamicColors {
        text = tview.TranslateANSI(text)
    }
    return parseStyledText(text, width, t.wordWrap, t.dynamicColors, t.regions)
}
```

Add a `parseStyledText` function that parses `[color]`, `["id"]` tags and produces `[][]cellInfo`. This replicates the core of tview's text rendering:

```go
func parseStyledText(text string, width int, wordWrap, dynamicColors, regions bool) [][]cellInfo {
    type styleState struct {
        fg, bg      tcell.Color
        bold, underline, reverse, blink, dim bool
    }
    var states []styleState
    cur := styleState{fg: tcell.ColorDefault, bg: tcell.ColorDefault}

    var lines [][]cellInfo
    var curLine []cellInfo

    flushLine := func() {
        if len(curLine) > 0 {
            lines = append(lines, curLine)
            curLine = nil
        }
    }

    pushState := func() { states = append(states, cur) }
    popState := func() { if len(states) > 0 { cur = states[len(states)-1]; states = states[:len(states)-1] } }

    resetStyle := func() {
        cur = styleState{fg: tcell.ColorDefault, bg: tcell.ColorDefault}
    }

    // Parse color tag format: [fg:bg:attr]
    // fg/bg can be: color name, #hex, or "-" to keep default
    // attr can contain: b (bold), u (underline), r (reverse), d (dim), l (blink), s (strike), "" (reset all)
    applyTag := func(tag string) {
        parts := strings.Split(tag, ":")
        // Reset all attributes when tag has no attributes and is a plain color
        if len(parts) == 1 && parts[0] != "" && parts[0][0] != '#' {
            // Could be color name or "::-"
        }
        for i := range parts {
            parts[i] = strings.TrimSpace(parts[i])
        }

        if len(parts) >= 1 && parts[0] != "" && parts[0] != "-" {
            cur.fg = tcell.GetColor(parts[0])
        }
        if len(parts) >= 2 {
            if parts[1] == "-" {
                cur.bg = tcell.ColorDefault
            } else if parts[1] != "" {
                cur.bg = tcell.GetColor(parts[1])
            }
        }
        if len(parts) >= 3 {
            for _, attr := range parts[2] {
                switch attr {
                case 'b': cur.bold = true
                case 'u': cur.underline = true
                case 'r': cur.reverse = true
                case 'd': cur.dim = true
                case 'l': cur.blink = true
                case 's': cur.strike = true
                }
            }
        }
    }

    regionActive := false
    runes := []rune(text)
    for i := 0; i < len(runes); i++ {
        if runes[i] != '[' {
            if runes[i] == '\n' {
                flushLine()
                continue
            }
            cell := cellInfo{ch: runes[i], style: buildStyle(cur)}
            curLine = append(curLine, cell)
            continue
        }
        // Potential tag
        end := strings.Index(string(runes[i:]), "]")
        if end < 0 {
            // Unclosed bracket, treat as literal
            cell := cellInfo{ch: '[', style: buildStyle(cur)}
            curLine = append(curLine, cell)
            continue
        }
        tag := string(runes[i+1 : i+end])
        i += end

        if tag == "" || !dynamicColors {
            // [] — empty tag, just skip
            continue
        }

        if tag[0] == '"' {
            // Region tag ["id"] or [""] to close
            if regions {
                regionActive = tag != `""`
            }
            continue
        }

        if tag == "-" || tag == ":-" || tag == "::-" {
            resetStyle()
            continue
        }

        if strings.HasPrefix(tag, "strikethrough") {
            // ignore strikethrough closing
            continue
        }

        if strings.HasPrefix(tag, "<") {
            // ignore HTML-style
            continue
        }

        // Regular color/attribute tag
        applyTag(tag)
    }
    flushLine()

    if wordWrap && width > 0 {
        lines = applyWordWrap(lines, width)
    }
    return lines
}

func buildStyle(s styleState) tcell.Style {
    st := tcell.StyleDefault.Foreground(s.fg).Background(s.bg)
    if s.bold { st = st.Bold(true) }
    if s.underline { st = st.Underline(true) }
    if s.reverse { st = st.Reverse(true) }
    if s.blink { st = st.Blink(true) }
    return st
}

// applyWordWrap wraps lines at word boundaries to fit width.
func applyWordWrap(lines [][]cellInfo, width int) [][]cellInfo {
    var result [][]cellInfo
    for _, line := range lines {
        if len(line) <= width {
            result = append(result, line)
            continue
        }
        start := 0
        for start < len(line) {
            end := start + width
            if end > len(line) {
                end = len(line)
            }
            // Look backwards for a space to break at
            breakAt := end
            if end < len(line) {
                for j := end; j > start; j-- {
                    if line[j-1].ch == ' ' {
                        breakAt = j
                        break
                    }
                }
            }
            result = append(result, line[start:breakAt])
            start = breakAt
            if start < len(line) && line[start].ch == ' ' {
                start++
            }
        }
    }
    return result
}
```

Update `GetText`:
```go
func (t *SelectableTextView) GetText(stripAllTags bool) string {
    if stripAllTags {
        return tview.StripTags(t.text)
    }
    return t.text
}
```

- [ ] **Step 3: Run test to verify it passes**

Run: `go test ./internal/ui/ -run "TestSelectableTextView" -v`
Expected: PASS

- [ ] **Step 4: Write failing test for regions/highlights**

```go
func TestSelectableTextView_Regions(t *testing.T) {
    tv := NewSelectableTextView()
    tv.SetRegions(true)
    tv.SetText(`before ["id"]highlight[""] after`)
    if tv.GetText(true) != "before highlight after" {
        t.Fatalf("unexpected stripped text: %q", tv.GetText(true))
    }
}
```

- [ ] **Step 5: Implement region handling in parseStyledText**

Modify the parse loop to handle `["id"]` region tags:
- When parsing, skip `["..."]` tags
- Track region boundaries with `tv.highlights`
- For highlighted regions, apply `style.Reverse(true)`

Add `Highlight` methods:
```go
func (t *SelectableTextView) Highlight(regionIDs ...string) *SelectableTextView {
    t.highlights = regionIDs
    return t
}
```

- [ ] **Step 6: Run tests**

Run: `go test ./internal/ui/ -run "TestSelectableTextView" -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/ui/selectable_textview.go internal/ui/selectable_textview_test.go
git commit -m "feat: add dynamic colors and region support to SelectableTextView"
```

---

### Task 3: Add mouse selection and keyboard shortcuts

**Files:**
- Modify: `internal/ui/selectable_textview.go`
- Test: `internal/ui/selectable_textview_test.go`

- [ ] **Step 1: Write failing test for mouse selection**

```go
func TestSelectableTextView_MouseSelection(t *testing.T) {
    tv := NewSelectableTextView()
    tv.SetRect(0, 0, 20, 5)
    tv.SetText("hello world select this")
    // Simulate mouse at column 6 (start of "world")
    tv.handleMousePress(0, 6)
    tv.handleMouseDrag(0, 19) // drag to end
    tv.handleMouseRelease()

    if !tv.HasSelection() {
        t.Fatal("expected selection")
    }
    sel := tv.GetSelection()
    if sel != "world select this" {
        t.Fatalf("expected 'world select this', got %q", sel)
    }
}
```

Run: `go test ./internal/ui/ -run TestSelectableTextView_MouseSelection -v`
Expected: FAIL — methods not defined

- [ ] **Step 2: Implement MouseHandler and selection logic**

```go
// Selection methods
func (t *SelectableTextView) HasSelection() bool {
    return t.selectionVisible
}

func (t *SelectableTextView) GetSelection() string {
    if !t.selectionVisible {
        return ""
    }
    lines := t.splitLines(100) // use stored width from last draw
    // Extract text from anchor to end coordinates
    // Normalize coords so anchor is always before end
    startRow, startCol := t.selectAnchorRow, t.selectAnchorCol
    endRow, endCol := t.selectEndRow, t.selectEndCol
    if startRow > endRow || (startRow == endRow && startCol > endCol) {
        startRow, startCol = endRow, endCol
        endRow, endCol = t.selectAnchorRow, t.selectAnchorCol
    }

    var sel strings.Builder
    for r := startRow; r <= endRow && r < len(lines); r++ {
        line := string(lines[r])
        if r == startRow && r == endRow {
            sel.WriteString(line[startCol:endCol])
        } else if r == startRow {
            sel.WriteString(line[startCol:])
        } else if r == endRow {
            sel.WriteString(line[:endCol])
        } else {
            sel.WriteString(line)
        }
        if r < endRow {
            sel.WriteString("\n")
        }
    }
    return sel.String()
}

func (t *SelectableTextView) ClearSelection() *SelectableTextView {
    t.selecting = false
    t.selectionVisible = false
    return t
}

func (t *SelectableTextView) SelectAll() *SelectableTextView {
    lines := t.splitLines(100)
    if len(lines) == 0 {
        return t
    }
    t.selectAnchorRow, t.selectAnchorCol = 0, 0
    lastLine := len(lines) - 1
    t.selectEndRow, t.selectEndCol = lastLine, len(lines[lastLine])
    t.selectionVisible = true
    return t
}

// Mouse handling
func (t *SelectableTextView) handleMousePress(line, col int) {
    t.selectAnchorRow = line
    t.selectAnchorCol = col
    t.selectEndRow = line
    t.selectEndCol = col
    t.selecting = true
    t.selectionVisible = false
}

func (t *SelectableTextView) handleMouseDrag(line, col int) {
    if !t.selecting {
        return
    }
    t.selectEndRow = line
    t.selectEndCol = col
    t.selectionVisible = true
}

func (t *SelectableTextView) handleMouseRelease() {
    t.selecting = false
}

func (t *SelectableTextView) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
    return t.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
        x, y := event.Position()
        switch action {
        case tview.MouseLeftDown:
            setFocus(t)
            innerX, innerY, _, _ := t.Box.GetInnerRect()
            line := y - innerY
            col := x - innerX
            if line >= 0 && col >= 0 {
                t.handleMousePress(line, col)
                consumed = true
            }
        case tview.MouseMove:
            if t.selecting {
                innerX, innerY, _, _ := t.Box.GetInnerRect()
                line := y - innerY
                col := x - innerX
                if line >= 0 && col >= 0 {
                    t.handleMouseDrag(line, col)
                }
                consumed = true
            }
        case tview.MouseLeftUp:
            t.handleMouseRelease()
            consumed = true
        case tview.MouseScrollUp:
            // Handled by ChatPanel.SetMouseCapture
        case tview.MouseScrollDown:
            // Handled by ChatPanel.SetMouseCapture
        }
        return
    })
}
```

- [ ] **Step 3: Run test**

Run: `go test ./internal/ui/ -run TestSelectableTextView_MouseSelection -v`
Expected: PASS

- [ ] **Step 4: Write failing test for keyboard shortcuts**

```go
func TestSelectableTextView_Keyboard(t *testing.T) {
    tv := NewSelectableTextView()
    tv.SetRect(0, 0, 20, 5)
    tv.SetText("hello world")
    tv.SelectAll()

    if !tv.HasSelection() {
        t.Fatal("SelectAll should create selection")
    }
    if tv.GetSelection() != "hello world" {
        t.Fatalf("expected 'hello world', got %q", tv.GetSelection())
    }

    tv.ClearSelection()
    if tv.HasSelection() {
        t.Fatal("ClearSelection should clear selection")
    }
}
```

- [ ] **Step 5: Implement InputHandler**

```go
func (t *SelectableTextView) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
    return t.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
        switch event.Key() {
        case tcell.KeyCtrlC, tcell.KeyCtrlQ:
            if t.HasSelection() {
                t.copyToClipboard(t.GetSelection())
                t.ClearSelection()
            }
        case tcell.KeyCtrlA:
            t.SelectAll()
        case tcell.KeyEscape:
            t.ClearSelection()
        case tcell.KeyUp:
            row, _ := t.GetScrollOffset()
            t.ScrollTo(row-1, 0)
        case tcell.KeyDown:
            row, _ := t.GetScrollOffset()
            t.ScrollTo(row+1, 0)
        case tcell.KeyPgUp:
            _, _, _, height := t.GetInnerRect()
            row, _ := t.GetScrollOffset()
            t.ScrollTo(row-height, 0)
        case tcell.KeyPgDn:
            _, _, _, height := t.GetInnerRect()
            row, _ := t.GetScrollOffset()
            t.ScrollTo(row+height, 0)
        }
    })
}
```

- [ ] **Step 6: Run tests**

Run: `go test ./internal/ui/ -run "TestSelectableTextView" -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/ui/selectable_textview.go internal/ui/selectable_textview_test.go
git commit -m "feat: add mouse selection and keyboard shortcuts to SelectableTextView"
```

---

### Task 4: Add clipboard support

**Files:**
- Modify: `internal/ui/selectable_textview.go`

- [ ] **Step 1: Implement clipboard copy via tcell**

Add imports: `"os/exec"`, `"runtime"` for clipboard fallback:

```go
func (t *SelectableTextView) copyToClipboard(text string) {
    if text == "" {
        return
    }
    // Use platform-specific clipboard command
    var cmd *exec.Cmd
    switch runtime.GOOS {
    case "windows":
        cmd = exec.Command("clip")
    case "darwin":
        cmd = exec.Command("pbcopy")
    default: // linux
        if _, err := exec.LookPath("wl-copy"); err == nil {
            cmd = exec.Command("wl-copy")
        } else {
            cmd = exec.Command("xclip", "-selection", "clipboard")
        }
    }
    if cmd != nil {
        cmd.Stdin = strings.NewReader(text)
        cmd.Run()
    }
}
```

- [ ] **Step 2: Implement ScrollTo, ScrollToEnd, GetScrollOffset methods**

```go
func (t *SelectableTextView) scrollOffset int

func (t *SelectableTextView) ScrollTo(row, column int) *SelectableTextView {
    t.scrollOffset = row
    if t.scrollOffset < 0 {
        t.scrollOffset = 0
    }
    return t
}

func (t *SelectableTextView) ScrollToEnd() *SelectableTextView {
    t.scrollOffset = -1 // signal for "end" (computed during Draw)
    return t
}

func (t *SelectableTextView) GetScrollOffset() (row, column int) {
    return t.scrollOffset, 0
}
```

- [ ] **Step 3: Build to verify compilation**

Run: `go build ./internal/ui/`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add internal/ui/selectable_textview.go
git commit -m "feat: add clipboard and scroll support to SelectableTextView"
```

---

### Task 5: Wire SelectableTextView into ChatPanel

**Files:**
- Modify: `internal/ui/chatpanel.go`

- [ ] **Step 1: Replace field type**

```go
// Before:
type ChatPanel struct {
    *tview.TextView
    // ...
}

// After:
type ChatPanel struct {
    *SelectableTextView
    // ...
}
```

- [ ] **Step 2: Update constructor**

```go
func NewChatPanel() *ChatPanel {
    c := &ChatPanel{
        SelectableTextView: NewSelectableTextView(),
    }
    c.SetDynamicColors(true)
    c.SetScrollable(true)
    c.SetWordWrap(true)
    c.SetRegions(true)
    c.SetTextStyle(tcell.StyleDefault.Background(tcell.ColorDefault))
    return c
}
```

- [ ] **Step 3: Update SetText call in refresh()**

The current code:
```go
c.SetText(tview.TranslateANSI(c.session.RenderMessages(80, DefaultThemes[0].Colors)))
```

Should become:
```go
c.SetText(tview.TranslateANSI(c.session.RenderMessages(80, DefaultThemes[0].Colors)))
```

No change needed — `SetText` signature is the same.

- [ ] **Step 4: Update app.go mouse capture reference**

In `app.go`, the current code:
```go
a.chatPanel.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
    switch action {
    case tview.MouseScrollUp:
        ...
```

This should still work since `SelectableTextView` embeds `*tview.Box` which provides `SetMouseCapture`. The mouse capture wraps our `MouseHandler`. Verify scroll behavior is preserved.

- [ ] **Step 5: Build and run tests**

Run: `go build ./...`
Expected: no errors

Run: `go test ./internal/ui/ -v -count=1`
Expected: all existing tests pass

- [ ] **Step 6: Commit**

```bash
git add internal/ui/chatpanel.go
git commit -m "feat: wire SelectableTextView into ChatPanel"
```

---

### Task 6: Build and manual verification

**Files:** (none)

- [ ] **Step 1: Build the full application**

Run: `go build -o agent-tui.exe ./cmd/agent`
Expected: binary created successfully

- [ ] **Step 2: Manual verification checklist**

- Launch the app (`./agent-tui.exe`)
- Send a message to generate conversation content
- Verify text is displayed correctly with colors
- Click and drag to select text — verify selection highlight appears
- Verify scrolling works with mouse wheel
- Press Ctrl+A — verify all text is selected
- Press Escape — verify selection clears
- Press Ctrl+Q — verify selection copies (paste externally to verify)
- Resize terminal — verify word wrap still works
- Press Ctrl+F and search — verify search highlights still work

- [ ] **Step 3: Run test suite if available**

Run: `go test ./... -count=1`
Expected: all tests pass
