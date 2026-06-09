package ui

import (
	"os/exec"
	"runtime"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type cellInfo struct {
	ch    rune
	style tcell.Style
}

type SelectableTextView struct {
	*tview.Box
	text          string
	dynamicColors bool
	regions       bool
	scrollable    bool
	wordWrap      bool
	maxLines      int

	selecting        bool
	selectAnchorRow  int
	selectAnchorCol  int
	selectEndRow     int
	selectEndCol     int
	selectionVisible bool

	highlights         []string
	scrollOffset       int
	lastRenderedOffset int
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
		return stripTags(t.text)
	}
	return t.text
}

func stripTags(text string) string {
	var b strings.Builder
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '[' {
			end := strings.Index(string(runes[i:]), "]")
			if end >= 0 {
				i += end
				continue
			}
		}
		b.WriteRune(runes[i])
	}
	return b.String()
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

func (t *SelectableTextView) Clear() *SelectableTextView {
	t.text = ""
	return t
}

func (t *SelectableTextView) Write(p []byte) (n int, err error) {
	t.text += string(p)
	return len(p), nil
}

func (t *SelectableTextView) Highlight(regionIDs ...string) *SelectableTextView {
	t.highlights = regionIDs
	return t
}

func (t *SelectableTextView) HasSelection() bool {
	return t.selectionVisible
}

func (t *SelectableTextView) GetSelection() string {
	if !t.selectionVisible {
		return ""
	}
	lines := t.splitLines(100)
	if len(lines) == 0 {
		return ""
	}
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
			if startCol < len(line) && endCol <= len(line) {
				sel.WriteString(line[startCol:endCol])
			}
		} else if r == startRow {
			if startCol < len(line) {
				sel.WriteString(line[startCol:])
			}
		} else if r == endRow {
			if endCol <= len(line) {
				sel.WriteString(line[:endCol])
			}
		} else {
			sel.WriteString(line)
		}
		if r < endRow {
			sel.WriteString("\n")
		}
	}
	return sel.String()
}

func (t *SelectableTextView) CopySelection() {
	if t.HasSelection() {
		t.copyToClipboard(t.GetSelection())
		t.ClearSelection()
	}
}

func (t *SelectableTextView) ClearSelection() *SelectableTextView {
	t.selecting = false
	t.selectionVisible = false
	return t
}

func (t *SelectableTextView) copyToClipboard(text string) {
	if text == "" {
		return
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("clip")
	case "darwin":
		cmd = exec.Command("pbcopy")
	default:
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

func (t *SelectableTextView) SelectAll() *SelectableTextView {
	lines := t.splitLines(100)
	if len(lines) == 0 {
		return t
	}
	t.selectAnchorRow = 0
	t.selectAnchorCol = 0
	lastLine := len(lines) - 1
	t.selectEndRow = lastLine
	t.selectEndCol = len(lines[lastLine])
	t.selectionVisible = true
	return t
}

func (t *SelectableTextView) ScrollTo(row, column int) *SelectableTextView {
	t.scrollOffset = row
	if t.scrollOffset < 0 {
		t.scrollOffset = 0
	}
	return t
}

func (t *SelectableTextView) ScrollToEnd() *SelectableTextView {
	t.scrollOffset = -1
	return t
}

func (t *SelectableTextView) GetScrollOffset() (row, column int) {
	if t.scrollOffset < 0 {
		return t.lastRenderedOffset, 0
	}
	return t.scrollOffset, 0
}

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
			line := y - innerY + t.lastRenderedOffset
			col := x - innerX
			if line >= 0 && col >= 0 {
				t.handleMousePress(line, col)
				consumed = true
			}
		case tview.MouseMove:
			if t.selecting {
				innerX, innerY, _, _ := t.Box.GetInnerRect()
				line := y - innerY + t.lastRenderedOffset
				col := x - innerX
				if line >= 0 && col >= 0 {
					t.handleMouseDrag(line, col)
				}
				consumed = true
			}
		case tview.MouseLeftUp:
			t.handleMouseRelease()
			consumed = true
		}
		return
	})
}

func (t *SelectableTextView) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return t.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyCtrlC, tcell.KeyCtrlQ:
			t.CopySelection()
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

func (t *SelectableTextView) Draw(screen tcell.Screen) {
	t.Box.Draw(screen)
	if t.text == "" {
		return
	}
	x, y, width, height := t.Box.GetInnerRect()
	if width <= 0 || height <= 0 {
		return
	}

	cells := t.parseCells(width)
	defaultStyle := tcell.StyleDefault.Background(t.GetBackgroundColor())

	startLine := t.scrollOffset
	if startLine < 0 || startLine > len(cells)-height {
		startLine = len(cells) - height
	}
	if startLine < 0 {
		startLine = 0
	}
	t.lastRenderedOffset = startLine

	for lineIdx := 0; lineIdx < height && startLine+lineIdx < len(cells); lineIdx++ {
		absLine := startLine + lineIdx
		line := cells[absLine]
		drawX := x
		for col := 0; col < len(line) && drawX < x+width; col++ {
			cell := line[col]
			cellStyle := cell.style
			if cellStyle == (tcell.Style{}) {
				cellStyle = defaultStyle
			}
			if t.selectionVisible && t.isInSelection(absLine, col) {
				cellStyle = cellStyle.Reverse(true)
			}
			screen.SetContent(drawX, y+lineIdx, cell.ch, nil, cellStyle)
			drawX++
		}
	}
}

func (t *SelectableTextView) parseCells(width int) [][]cellInfo {
	text := t.text
	if t.dynamicColors {
		text = tview.TranslateANSI(text)
	}
	return parseStyledText(text, width, t.wordWrap, t.dynamicColors, t.regions, t.highlights)
}

type styleState struct {
	fg, bg                           tcell.Color
	bold, underline, reverse, blink bool
}

func buildStyle(s styleState) tcell.Style {
	st := tcell.StyleDefault.Foreground(s.fg).Background(s.bg)
	if s.bold {
		st = st.Bold(true)
	}
	if s.underline {
		st = st.Underline(true)
	}
	if s.reverse {
		st = st.Reverse(true)
	}
	if s.blink {
		st = st.Blink(true)
	}
	return st
}

func parseStyledText(text string, width int, wordWrap, dynamicColors, regions bool, highlights []string) [][]cellInfo {
	cur := styleState{fg: tcell.ColorDefault, bg: tcell.ColorDefault}

	var lines [][]cellInfo
	var curLine []cellInfo

	flushLine := func() {
		if len(curLine) > 0 {
			lines = append(lines, curLine)
			curLine = nil
		}
	}

	highlightSet := make(map[string]bool, len(highlights))
	for _, h := range highlights {
		highlightSet[h] = true
	}

	resetStyle := func() {
		cur = styleState{fg: tcell.ColorDefault, bg: tcell.ColorDefault}
	}

	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '[' && dynamicColors {
			end := strings.Index(string(runes[i:]), "]")
			if end < 0 {
				cell := cellInfo{ch: '[', style: buildStyle(cur)}
				curLine = append(curLine, cell)
				continue
			}
			tag := string(runes[i+1 : i+end])
			i += end

			if tag == "" {
				continue
			}

			if tag[0] == '"' {
				if regions {
					regionID := strings.Trim(tag, "\"")
					if regionID != "" && highlightSet[regionID] {
						cur.reverse = !cur.reverse
					}
				}
				continue
			}

			if tag == "-" || tag == ":-" || tag == "::-" {
				resetStyle()
				continue
			}

			parts := strings.Split(tag, ":")
			for j := range parts {
				parts[j] = strings.TrimSpace(parts[j])
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
					case 'b':
						cur.bold = true
					case 'u':
						cur.underline = true
					case 'r':
						cur.reverse = true
					}
				}
			}
			continue
		}

		if runes[i] == '\n' {
			flushLine()
			continue
		}

		cell := cellInfo{ch: runes[i], style: buildStyle(cur)}
		curLine = append(curLine, cell)
	}
	flushLine()

	if wordWrap && width > 0 {
		lines = applyWordWrap(lines, width)
	}
	return lines
}

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
			breakAt := end
			if end < len(line) {
				for j := end; j > start; j-- {
					if line[j-1].ch == ' ' {
						breakAt = j - 1
						break
					}
				}
			}
			if breakAt <= start {
				breakAt = end
			}
			result = append(result, line[start:breakAt])
			start = breakAt + 1
		}
	}
	return result
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
		start := 0
		for start < len(runes) {
			end := start + width
			if end > len(runes) {
				end = len(runes)
			}
			breakAt := end
			if end < len(runes) {
				for j := end; j > start; j-- {
					if runes[j-1] == ' ' {
						breakAt = j - 1
						break
					}
				}
			}
			if breakAt <= start {
				breakAt = end
			}
			result = append(result, runes[start:breakAt])
			start = breakAt + 1
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
