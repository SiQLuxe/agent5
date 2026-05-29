package status

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type StatusBar struct {
	*tview.TextView
	mode      string
	tasks     int
	connected bool
}

func New() *StatusBar {
	s := &StatusBar{
		TextView: tview.NewTextView(),
	}
	s.SetDynamicColors(true)
	s.SetTextAlign(tview.AlignLeft)
	return s
}

func (s *StatusBar) SetMode(mode string) {
	s.mode = mode
	s.refresh()
}

func (s *StatusBar) SetTasks(n int) {
	s.tasks = n
	s.refresh()
}

func (s *StatusBar) SetConnected(v bool) {
	s.connected = v
	s.refresh()
}

func (s *StatusBar) SetBackgroundColor(color tcell.Color) {
	s.TextView.SetBackgroundColor(color)
}

func (s *StatusBar) refresh() {
	connStr := "●"
	if !s.connected {
		connStr = "○"
	}
	s.SetText(fmt.Sprintf("  %s  Mode: %s  Tasks: %d", connStr, s.mode, s.tasks))
}
