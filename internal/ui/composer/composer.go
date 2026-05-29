package composer

import (
	"charm.land/bubbles/v2/textarea"
	"charm.land/lipgloss/v2"
)

type ComposerColors struct {
	Background string
	Prompt     string
	Text       string
}

func DefaultColors() ComposerColors {
	return ComposerColors{
		Background: "#1a1a1a",
		Prompt:     "#4ec9b0",
		Text:       "#d4d4d4",
	}
}

type Composer struct {
	textarea textarea.Model
	width    int
	colors   ComposerColors
}

func NewComposer() *Composer {
	ta := textarea.New()
	ta.DynamicHeight = true
	ta.MinHeight = 1
	ta.MaxHeight = 8
	ta.ShowLineNumbers = false
	ta.Prompt = "\u276f "  // ❯

	c := &Composer{
		textarea: ta,
		width:    80,
		colors:   DefaultColors(),
	}
	c.applyColors()
	return c
}

func (c *Composer) SetWidth(width int) {
	c.width = width
	c.textarea.SetWidth(width)
}

func (c *Composer) SetColors(colors ComposerColors) {
	c.colors = colors
	c.applyColors()
}

func (c *Composer) SetInput(input string) {
	c.textarea.SetValue(input)
}

func (c *Composer) GetInput() string {
	return c.textarea.Value()
}

func (c *Composer) ClearInput() {
	c.textarea.SetValue("")
}

func (c *Composer) AppendInput(char string) {
	c.textarea.InsertString(char)
}

func (c *Composer) Backspace() {
	val := c.textarea.Value()
	if len(val) > 0 {
		runes := []rune(val)
		c.textarea.SetValue(string(runes[:len(runes)-1]))
	}
}

func (c *Composer) View() string {
	return c.textarea.View()
}

func (c *Composer) applyColors() {
	s := textarea.DefaultStyles(false)
	s.Focused.Base = lipgloss.NewStyle().
		Background(lipgloss.Color(c.colors.Background)).
		BorderTop(true).
		BorderForeground(lipgloss.Color("#3c3c3c"))
	s.Focused.Text = lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.colors.Text))
	s.Focused.Prompt = lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.colors.Prompt))
	s.Focused.CursorLine = lipgloss.NewStyle()
	s.Focused.LineNumber = lipgloss.NewStyle()
	c.textarea.SetStyles(s)
}
