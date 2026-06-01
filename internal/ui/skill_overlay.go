package ui

import (
	"github.com/example/agent-tui/internal/service"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type SkillOverlay struct {
	*tview.Flex
	list       *tview.List
	filterText string
	skills     []*service.Skill
}

func NewSkillOverlay() *SkillOverlay {
	list := tview.NewList()
	list.SetMainTextColor(tcell.ColorWhite)
	list.SetSecondaryTextColor(tcell.ColorGray)
	list.SetSelectedBackgroundColor(tcell.ColorDarkCyan)
	list.ShowSecondaryText(true)
	o := &SkillOverlay{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
		list: list,
	}
	o.AddItem(list, 0, 1, true)
	o.SetBackgroundColor(tcell.ColorDefault)
	return o
}

func (o *SkillOverlay) SetSkills(skills []*service.Skill) {
	o.skills = skills
	o.applyFilter()
}

func (o *SkillOverlay) SetFilter(text string) {
	o.filterText = text
	o.applyFilter()
}

func (o *SkillOverlay) applyFilter() {
	o.list.Clear()
	for _, s := range o.skills {
		if o.filterText != "" && !hasPrefixFold(s.Name, o.filterText) {
			continue
		}
		desc := s.Description
		if len([]rune(desc)) > 60 {
			desc = string([]rune(desc)[:60]) + "..."
		}
		o.list.AddItem(s.Name, desc, 0, nil)
	}
}

func (o *SkillOverlay) SelectedSkillName() string {
	if o.list.GetItemCount() == 0 {
		return ""
	}
	main, _ := o.list.GetItemText(o.list.GetCurrentItem())
	return main
}

func (o *SkillOverlay) GetItemCount() int {
	return o.list.GetItemCount()
}

func hasPrefixFold(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return s[:len(prefix)] == prefix
}
