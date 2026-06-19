// Package hint provides a one-line bottom-of-screen hint bar for transient
// feedback, double-press confirmations, and short-lived warnings.
//
// HintBar is a thin wrapper around tview.TextView. It is callable directly
// from the tview main event-loop goroutine (e.g. inside an InputCapture
// handler) — Show / Clear are synchronous SetText calls and never enqueue
// to the application's update channel. Callers that run in a different
// goroutine (e.g. a timer firing Clear) must wrap the call in
// Application.QueueUpdateDraw themselves.
package hint

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Level controls the foreground color used to render a hint.
type Level int

const (
	// LevelInfo renders in low-emphasis gray. Use for neutral feedback
	// ("已复制") and double-press confirmations ("再次按 Ctrl+C 退出").
	LevelInfo Level = iota
	// LevelWarn renders in yellow. Use for soft warnings.
	LevelWarn
	// LevelError renders in red. Use for transient error reports.
	LevelError
)

// HintBar is a single-line transient hint surface, intended to live at the
// very bottom of the application layout.
type HintBar struct {
	*tview.TextView
}

// New constructs an empty HintBar.
func New() *HintBar {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetTextAlign(tview.AlignLeft)
	tv.SetTextStyle(tcell.StyleDefault.Background(tcell.ColorDefault))
	return &HintBar{TextView: tv}
}

// Show writes text to the bar with a color appropriate for level. Calling
// Show again replaces any prior content immediately.
func (h *HintBar) Show(text string, level Level) {
	h.SetText(fmt.Sprintf("  %s%s[-]", colorTag(level), text))
}

// Clear empties the bar.
func (h *HintBar) Clear() {
	h.SetText("")
}

func colorTag(l Level) string {
	switch l {
	case LevelWarn:
		return "[yellow]"
	case LevelError:
		return "[red]"
	default:
		return "[gray]"
	}
}
