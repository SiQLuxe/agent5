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
	message   string
}

func New() *StatusBar {
	s := &StatusBar{
		TextView: tview.NewTextView(),
	}
	s.SetDynamicColors(true)
	s.SetTextAlign(tview.AlignLeft)
	s.SetTextStyle(tcell.StyleDefault.Background(tcell.ColorDefault))
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

func (s *StatusBar) ShowMessage(msg string) {
	s.message = msg
	s.SetText(msg)
}

func (s *StatusBar) ClearMessage() {
	s.message = ""
	s.refresh()
}

func (s *StatusBar) refresh() {
	if s.message != "" {
		s.SetText(s.message)
		return
	}
	connStr := "●"
	if !s.connected {
		connStr = "○"
	}
	s.SetText(fmt.Sprintf("  %s  Mode: %s  Tasks: %d  Ctrl+O 帮助", connStr, s.mode, s.tasks))
}
