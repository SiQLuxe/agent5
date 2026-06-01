package ui

import (
	"github.com/example/agent-tui/internal/service"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type PaletteMode int

const (
	ShowAll    PaletteMode = iota
	ShowSkills
)

type CommandPalette struct {
	*tview.Flex
	list         *tview.List
	allCommands  []*service.Command
	filtered     []*service.Command
	filterText   string
	mode         PaletteMode
}

func NewCommandPalette() *CommandPalette {
	list := tview.NewList()
	list.SetMainTextColor(tcell.ColorWhite)
	list.SetSecondaryTextColor(tcell.ColorGray)
	list.SetSelectedBackgroundColor(tcell.ColorDarkCyan)
	list.ShowSecondaryText(true)

	p := &CommandPalette{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
		list: list,
	}
	p.AddItem(list, 0, 1, true)
	p.SetBackgroundColor(tcell.ColorDefault)
	return p
}

func (p *CommandPalette) SetCommands(cmds []*service.Command) {
	p.allCommands = cmds
	p.applyFilter()
}

func (p *CommandPalette) SetMode(m PaletteMode) {
	p.mode = m
	p.applyFilter()
}

func (p *CommandPalette) SetFilter(text string) {
	p.filterText = text
	p.applyFilter()
}

func (p *CommandPalette) FilterText() string {
	return p.filterText
}

func (p *CommandPalette) applyFilter() {
	p.filtered = nil
	p.list.Clear()

	for _, cmd := range p.allCommands {
		if p.mode == ShowSkills && cmd.Category != service.CmdSkill {
			continue
		}
		if p.filterText != "" && !hasPrefixFold(cmd.Name, p.filterText) {
			continue
		}
		p.filtered = append(p.filtered, cmd)
		displayName := cmd.Name
		prefix := "  "
		if cmd.Category == service.CmdBuiltin {
			prefix = "\U0001f4c2 " // folder
		} else {
			prefix = "\u26a1 " // zap
		}
		p.list.AddItem(prefix+displayName, cmd.Description, 0, nil)
	}
}

func (p *CommandPalette) SelectedCommand() *service.Command {
	idx := p.list.GetCurrentItem()
	if idx < 0 || idx >= len(p.filtered) {
		return nil
	}
	return p.filtered[idx]
}

func (p *CommandPalette) GetItemCount() int {
	return p.list.GetItemCount()
}

func hasPrefixFold(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return s[:len(prefix)] == prefix
}
