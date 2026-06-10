package suggestion

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/rivo/tview"

	"github.com/example/agent-tui/internal/service"
)

const maxVisibleItems = 8

type SuggestionMenu struct {
	*tview.Flex
	items        []*service.Command
	filtered     []*service.Command
	filterText   string
	selected     int
	scrollOffset int
	list         *tview.Box
	visible      bool
}

func New() *SuggestionMenu {
	list := tview.NewBox()
	list.SetBackgroundColor(tcell.ColorDefault)

	m := &SuggestionMenu{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
		list: list,
	}

	m.AddItem(list, 0, 1, false)
	m.SetBackgroundColor(tcell.ColorDefault)

	list.SetDrawFunc(m.drawList)

	return m
}

func (m *SuggestionMenu) drawList(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
	if !m.visible || len(m.filtered) == 0 {
		clearStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
		for row := y; row < y+height; row++ {
			for col := x; col < x+width; col++ {
				screen.SetContent(col, row, ' ', nil, clearStyle)
			}
		}
		return x, y, width, height
	}

	nameColWidth := width * 40 / 100
	if nameColWidth > 20 {
		nameColWidth = 20
	}
	descColWidth := width - nameColWidth - 1
	if descColWidth < 0 {
		descColWidth = 0
	}

	endIdx := m.scrollOffset + height
	if endIdx > len(m.filtered) {
		endIdx = len(m.filtered)
	}
	for idx := m.scrollOffset; idx < endIdx; idx++ {
		cmd := m.filtered[idx]
		rowY := y + idx - m.scrollOffset

		style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
		if idx == m.selected {
			style = tcell.StyleDefault.Reverse(true)
		}

		name := cmd.Name
		if runewidth.StringWidth(name) > nameColWidth {
			name = runewidth.Truncate(name, nameColWidth-1, "…")
		}
		name = runewidth.FillRight(name, nameColWidth)
		drawString(screen, x, rowY, style, name)

		screen.SetContent(x+nameColWidth, rowY, ' ', nil, style)

		descStyle := tcell.StyleDefault.Foreground(tcell.ColorGray)
		if idx == m.selected {
			descStyle = style
		}
		desc := cmd.Description
		if runewidth.StringWidth(desc) > descColWidth {
			desc = runewidth.Truncate(desc, descColWidth-1, "…")
		}
		drawString(screen, x+nameColWidth+1, rowY, descStyle, desc)
	}

	return x, y, width, height
}

func drawString(screen tcell.Screen, x, y int, style tcell.Style, s string) {
	for _, r := range s {
		w := runewidth.RuneWidth(r)
		if w == 0 {
			continue
		}
		screen.SetContent(x, y, r, nil, style)
		x += w
	}
}

func (m *SuggestionMenu) Filtered() []*service.Command {
	return m.filtered
}

func (m *SuggestionMenu) SetCommands(cmds []*service.Command) {
	m.items = cmds
	m.selected = 0
	m.scrollOffset = 0
	m.applyFilter()
}

func (m *SuggestionMenu) SetFilter(text string) {
	m.filterText = text
	m.selected = 0
	m.scrollOffset = 0
	m.applyFilter()
}

func (m *SuggestionMenu) applyFilter() {
	m.filtered = nil
	filter := strings.ToLower(m.filterText)
	for _, cmd := range m.items {
		if strings.HasPrefix(strings.ToLower(cmd.Name), filter) {
			m.filtered = append(m.filtered, cmd)
		}
	}
}

func (m *SuggestionMenu) SelectNext() {
	if m.selected >= len(m.filtered)-1 {
		return
	}
	m.selected++
	if m.selected-m.scrollOffset >= maxVisibleItems {
		m.scrollOffset++
	}
}

func (m *SuggestionMenu) SelectPrev() {
	if m.selected <= 0 {
		return
	}
	m.selected--
	if m.selected < m.scrollOffset {
		m.scrollOffset--
	}
}

func (m *SuggestionMenu) Selected() *service.Command {
	if m.selected < 0 || m.selected >= len(m.filtered) {
		return nil
	}
	return m.filtered[m.selected]
}

func (m *SuggestionMenu) Show() {
	m.visible = true
}

func (m *SuggestionMenu) Hide() {
	m.visible = false
}

func (m *SuggestionMenu) Visible() bool {
	return m.visible
}

func (m *SuggestionMenu) Height() int {
	if !m.visible || len(m.filtered) == 0 {
		return 0
	}
	n := len(m.filtered)
	if n > maxVisibleItems {
		n = maxVisibleItems
	}
	return n
}
