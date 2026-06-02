package composer

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Composer struct {
	*tview.Flex
	textArea    *tview.TextArea
	prompt      *tview.TextView
	leftBorder  *tview.Box
	accentColor tcell.Color
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

	leftBorder := tview.NewBox()
	leftBorder.SetBackgroundColor(tcell.ColorDefault)

	c := &Composer{
		textArea:   textArea,
		prompt:     prompt,
		leftBorder: leftBorder,
	}

	leftBorder.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		style := tcell.StyleDefault.Foreground(c.accentColor)
		for row := 0; row < height; row++ {
			screen.SetContent(x, y+row, '│', nil, style)
		}
		return x, y, width, height
	})

	flex := tview.NewFlex().SetDirection(tview.FlexColumn)
	flex.AddItem(leftBorder, 1, 0, false)
	flex.AddItem(prompt, 2, 0, false)
	flex.AddItem(textArea, 0, 1, true)

	textArea.SetBackgroundColor(tcell.ColorDefault)
	prompt.SetBackgroundColor(tcell.ColorDefault)
	leftBorder.SetBackgroundColor(tcell.ColorDefault)
	flex.SetBackgroundColor(tcell.ColorDefault)

	c.Flex = flex
	return c
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

func (c *Composer) SetAccentColor(color tcell.Color) {
	c.accentColor = color
}

func (c *Composer) SetOnTextChanged(fn func(text string)) {
	c.textArea.SetChangedFunc(func() {
		text := c.textArea.GetText()
		if fn != nil {
			fn(text)
		}
	})
}

func (c *Composer) SetBackgroundColor(color tcell.Color) {
	c.textArea.SetBackgroundColor(color)
	c.prompt.SetBackgroundColor(color)
	c.leftBorder.SetBackgroundColor(color)
}
