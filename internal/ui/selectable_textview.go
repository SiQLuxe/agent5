package ui

import (
	"strings"

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

	selecting        bool
	selectAnchorRow  int
	selectAnchorCol  int
	selectEndRow     int
	selectEndCol     int
	selectionVisible bool

	highlights   []string
	scrollOffset int
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
						breakAt = j
						break
					}
				}
			}
			if breakAt == start {
				breakAt = end
			}
			result = append(result, runes[start:breakAt])
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
