package composer

import (
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

type ComposerColors struct {
	Background string
	Prompt     string
	Separator  string
	Text       string
}

func DefaultColors() ComposerColors {
	return ComposerColors{
		Background: "#1a1a1a",
		Prompt:     "#4ec9b0",
		Separator:  "#3c3c3c",
		Text:       "#d4d4d4",
	}
}

type Composer struct {
	width  int
	input  string
	colors ComposerColors
}

func NewComposer() *Composer {
	return &Composer{
		width:  80,
		input:  "",
		colors: DefaultColors(),
	}
}

func (c *Composer) SetWidth(width int) {
	c.width = width
}

func (c *Composer) SetColors(colors ComposerColors) {
	c.colors = colors
}

func (c *Composer) SetInput(input string) {
	c.input = input
}

func (c *Composer) GetInput() string {
	return c.input
}

func (c *Composer) ClearInput() {
	c.input = ""
}

func (c *Composer) AppendInput(char string) {
	c.input += char
}

func (c *Composer) Backspace() {
	if len(c.input) > 0 {
		_, size := utf8.DecodeLastRuneInString(c.input)
		c.input = c.input[:len(c.input)-size]
	}
}

func (c *Composer) View() string {
	prompt := lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.colors.Prompt)).
		Bold(true).
		Render("❯ ")

	inputText := lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.colors.Text)).
		Render(c.input)

	cursor := lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.colors.Text)).
		Background(lipgloss.Color(c.colors.Text)).
		Render(" ")

	content := prompt + inputText + cursor

	return lipgloss.NewStyle().
		Width(c.width).
		Background(lipgloss.Color(c.colors.Background)).
		Padding(0, 2).
		BorderTop(true).
		BorderForeground(lipgloss.Color(c.colors.Separator)).
		Render(content)
}
