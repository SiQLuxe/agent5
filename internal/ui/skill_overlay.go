package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/example/agent-tui/internal/service"
)

type SkillOverlay struct {
	*tview.Flex
	app        *App
	list       *tview.List
	detailView *tview.TextView
	skills     []*service.Skill
	infoText   *tview.TextView
}

func NewSkillOverlay(app *App) *SkillOverlay {
	list := tview.NewList()
	list.SetMainTextColor(tcell.ColorWhite)
	list.SetSecondaryTextColor(tcell.ColorGray)
	list.SetSelectedBackgroundColor(tcell.ColorDarkCyan)
	list.ShowSecondaryText(true)

	detailView := tview.NewTextView()
	detailView.SetDynamicColors(true)
	detailView.SetWordWrap(true)
	detailView.SetBorder(true)
	detailView.SetTitle(" Details ")

	infoText := tview.NewTextView()
	infoText.SetDynamicColors(true)
	infoText.SetTextAlign(tview.AlignCenter)
	infoText.SetText("[gray]Enter: activate  ↑↓: navigate  Esc: close[-]")

	so := &SkillOverlay{
		Flex:       tview.NewFlex().SetDirection(tview.FlexRow),
		app:        app,
		list:       list,
		detailView: detailView,
		infoText:   infoText,
	}

	so.SetBorder(true)
	so.SetTitle(" Skills ")
	so.SetBackgroundColor(tcell.ColorDefault)

	content := tview.NewFlex().SetDirection(tview.FlexColumn)
	content.AddItem(list, 0, 1, true)
	content.AddItem(detailView, 0, 2, false)

	so.AddItem(content, 0, 1, true)
	so.AddItem(infoText, 1, 0, false)

	list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		so.showDetail(index)
	})

	so.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			so.app.exitSkillOverlay()
			return nil
		case tcell.KeyEnter:
			so.activateSelected()
			return nil
		}
		return event
	})

	return so
}

func (so *SkillOverlay) LoadSkills() {
	so.list.Clear()
	so.skills = nil
	so.detailView.Clear()

	if so.app.skillRegistry == nil {
		so.list.AddItem("No skills available", "Press Esc to close", 0, nil)
		return
	}

	so.skills = so.app.skillRegistry.List()
	for _, s := range so.skills {
		secondary := s.Description
		prefix := "\u26a1 "
		so.list.AddItem(prefix+s.Name, secondary, 0, nil)
	}

	if len(so.skills) > 0 {
		so.showDetail(0)
	} else {
		so.list.AddItem("No skills available", "Press Esc to close", 0, nil)
	}
}

func (so *SkillOverlay) showDetail(index int) {
	if index < 0 || index >= len(so.skills) {
		return
	}
	s := so.skills[index]
	typeStr := "prompt"
	if s.Type == service.SkillHandler {
		typeStr = "handler"
	}
	var detail strings.Builder
	detail.WriteString("[yellow]Name:[-] ")
	detail.WriteString(s.Name)
	detail.WriteString("\n")
	detail.WriteString("[yellow]Description:[-] ")
	detail.WriteString(s.Description)
	detail.WriteString("\n")
	detail.WriteString("[yellow]Type:[-] ")
	detail.WriteString(typeStr)
	detail.WriteString("\n")
	if s.Prompt != "" {
		preview := s.Prompt
		runes := []rune(preview)
		if len(runes) > 200 {
			preview = string(runes[:200]) + "..."
		}
		detail.WriteString("[yellow]Prompt:[-]\n")
		detail.WriteString(preview)
	}
	so.detailView.SetText(detail.String())
}

func (so *SkillOverlay) activateSelected() {
	idx := so.list.GetCurrentItem()
	if idx < 0 || idx >= len(so.skills) {
		return
	}
	skill := so.skills[idx]
	so.app.exitSkillOverlay()
	so.app.executeSkill(skill.Name)
}
