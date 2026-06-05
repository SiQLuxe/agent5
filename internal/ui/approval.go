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
