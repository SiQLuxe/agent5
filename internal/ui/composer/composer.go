package composer

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Composer struct {
	*tview.Flex
	textArea *tview.TextArea
	prompt   *tview.TextView
}

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

func (c *Composer) SetInput(s string) {
	c.textArea.SetText(s, true)
}

func (c *Composer) GetInput() string {
	return c.textArea.GetText()
}

func (c *Composer) ClearInput() {
	c.textArea.SetText("", true)
}

func (c *Composer) SetPromptColor(color string) {
	c.prompt.SetText(fmt.Sprintf("[%s::b]> [-]", color))
}

func (c *Composer) SetBackgroundColor(color tcell.Color) {
	c.textArea.SetBackgroundColor(color)
	c.prompt.SetBackgroundColor(color)
}
