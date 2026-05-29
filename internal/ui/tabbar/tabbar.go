package tabbar

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Tab struct {
	ID    string
	Label string
}

type TabDock struct {
	*tview.Box
	tabs       []Tab
	active     int
	bgColor    tcell.Color
	activeFg   tcell.Color
	activeBg   tcell.Color
	inactiveFg tcell.Color
	inactiveBg tcell.Color
	onClick    func(idx int)
}

func New() *TabDock {
	t := &TabDock{
		Box:        tview.NewBox(),
		tabs:       []Tab{},
		active:     0,
		bgColor:    tcell.ColorDefault,
		activeFg:   tcell.ColorWhite,
		activeBg:   tcell.ColorBlue,
		inactiveFg: tcell.ColorGray,
		inactiveBg: tcell.ColorDefault,
	}
	t.SetDrawFunc(t.draw)
	return t
}

func (t *TabDock) AddTab(tab Tab) {
	t.tabs = append(t.tabs, tab)
}

func (t *TabDock) RemoveTab(index int) {
	if index < 0 || index >= len(t.tabs) {
		return
	}
	t.tabs = append(t.tabs[:index], t.tabs[index+1:]...)
	if index < t.active {
		t.active--
	}
	if t.active >= len(t.tabs) && len(t.tabs) > 0 {
		t.active = len(t.tabs) - 1
	}
}

func (t *TabDock) UpdateTab(index int, label string) {
	if index < 0 || index >= len(t.tabs) {
		return
	}
	t.tabs[index].Label = label
}

func (t *TabDock) SetActive(index int) {
	if index >= 0 && index < len(t.tabs) {
		t.active = index
	}
}

func (t *TabDock) ActiveIndex() int {
	return t.active
}

func (t *TabDock) TabCount() int {
	return len(t.tabs)
}

func (t *TabDock) SetOnClick(fn func(idx int)) {
	t.onClick = fn
	t.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseLeftClick && t.onClick != nil {
			x, _ := event.Position()
			_, _, w, _ := t.GetRect()
			idx := t.tabAtX(x, w)
			if idx >= 0 {
				t.onClick(idx)
				return action, nil
			}
			return action, nil
		}
		return action, event
	})
}

func (t *TabDock) tabAtX(x, width int) int {
	if len(t.tabs) == 0 || width <= 0 {
		return -1
	}
	labelTotal := 0
	for _, tab := range t.tabs {
		labelTotal += len(tab.Label) + 4
	}
	remaining := width - labelTotal
	extraPerTab := 0
	if len(t.tabs) > 0 && remaining > 0 {
		extraPerTab = remaining / len(t.tabs)
	}
	cx := 0
	for i, tab := range t.tabs {
		tabW := len(tab.Label) + 4 + extraPerTab
		if x >= cx && x < cx+tabW {
			return i
		}
		cx += tabW
	}
	return -1
}

func (t *TabDock) SetColors(activeFg, activeBg, inactiveFg, inactiveBg tcell.Color) {
	t.activeFg = activeFg
	t.activeBg = activeBg
	t.inactiveFg = inactiveFg
	t.inactiveBg = inactiveBg
}

func (t *TabDock) draw(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
	if len(t.tabs) == 0 {
		return x, y, width, height
	}

	for cy := y; cy < y+height; cy++ {
		for cx := x; cx < x+width; cx++ {
			screen.SetContent(cx, cy, ' ', nil, tcell.StyleDefault.Background(t.bgColor))
		}
	}

	labelTotal := 0
	for _, tab := range t.tabs {
		labelTotal += len(tab.Label) + 4
	}
	remaining := width - labelTotal
	extraPerTab := 0
	if len(t.tabs) > 0 && remaining > 0 {
		extraPerTab = remaining / len(t.tabs)
	}

	cx := x
	for i, tab := range t.tabs {
		tabW := len(tab.Label) + 4 + extraPerTab
		fg, bg := t.inactiveFg, t.inactiveBg
		if i == t.active {
			fg, bg = t.activeFg, t.activeBg
		}
		style := tcell.StyleDefault.Foreground(fg).Background(bg)
		label := " " + tab.Label + " "
		for j, ch := range label {
			screen.SetContent(cx+j, y, ch, nil, style)
		}
		cx += tabW
	}

	return x, y, width, height
}
