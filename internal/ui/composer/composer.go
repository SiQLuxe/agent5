package composer

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Composer struct {
	*tview.TextArea
}

func New() *Composer {
	c := &Composer{
		TextArea: tview.NewTextArea(),
	}
	c.SetWordWrap(true)
	c.SetSize(8, 0)
	return c
}

func (c *Composer) SetInput(s string) {
	c.SetText(s, true)
}

func (c *Composer) GetInput() string {
	return c.GetText()
}

func (c *Composer) ClearInput() {
	c.SetText("", true)
}

func (c *Composer) SetBackgroundColor(color tcell.Color) {
	c.TextArea.SetBackgroundColor(color)
}
