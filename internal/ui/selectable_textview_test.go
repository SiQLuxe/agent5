package ui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestSelectableTextView_RendersText(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetText("hello")
	if tv.GetText(false) != "hello" {
		t.Fatalf("expected 'hello', got %q", tv.GetText(false))
	}
}

func TestSelectableTextView_Draw(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetRect(0, 0, 20, 5)
	tv.SetText("hello world")
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(20, 5)
	tv.Draw(screen)
	screen.Show()
}

func TestSelectableTextView_GetTextStripTags(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetDynamicColors(true)
	tv.SetText(`[red]hello[-] world`)
	result := tv.GetText(true)
	if result != "hello world" {
		t.Fatalf("expected 'hello world', got %q", result)
	}
}

func TestSelectableTextView_Clear(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetText("hello")
	tv.Clear()
	if tv.GetText(false) != "" {
		t.Fatalf("expected empty, got %q", tv.GetText(false))
	}
}

func TestSelectableTextView_ColorTags(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetDynamicColors(true)
	tv.SetText(`[red]hello[-] [blue]world[-]`)
	// GetText(stripAllTags=true) should strip color tags
	result := tv.GetText(true)
	if result != "hello world" {
		t.Fatalf("expected 'hello world', got %q", result)
	}
}

func TestSelectableTextView_Highlight(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetRegions(true)
	tv.SetText(`before ["id"]highlight[""] after`)
	tv.Highlight("id")
	if tv.GetText(true) != "before highlight after" {
		t.Fatalf("unexpected stripped text: %q", tv.GetText(true))
	}
	// Draw with highlight should not panic
	tv.SetRect(0, 0, 30, 3)
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(30, 3)
	tv.Draw(screen)
	screen.Show()
}

func TestSelectableTextView_WordWrap(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetWordWrap(true)
	tv.SetText("hello world foo bar")
	lines := tv.splitLines(10)
	if len(lines) < 2 {
		t.Fatalf("expected multiple lines from word wrap, got %d", len(lines))
	}
	// Each line should not exceed wrap width
	for i, line := range lines {
		if len(line) > 10 {
			t.Fatalf("line %d exceeds wrap width: %d > 10", i, len(line))
		}
	}
}

func TestSelectableTextView_SelectAll(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetRect(0, 0, 20, 5)
	tv.SetText("hello world")
	tv.SelectAll()
	if !tv.HasSelection() {
		t.Fatal("SelectAll should create selection")
	}
	if tv.GetSelection() != "hello world" {
		t.Fatalf("expected 'hello world', got %q", tv.GetSelection())
	}
}

func TestSelectableTextView_ClearSelection(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetText("hello world")
	tv.SelectAll()
	tv.ClearSelection()
	if tv.HasSelection() {
		t.Fatal("ClearSelection should clear selection")
	}
}

func TestSelectableTextView_GetSelectionMultiline(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetText("line one\nline two\nline three")
	tv.SelectAll()
	sel := tv.GetSelection()
	if sel != "line one\nline two\nline three" {
		t.Fatalf("unexpected multiline selection: %q", sel)
	}
}

func TestSelectableTextView_ScrollTo(t *testing.T) {
	tv := NewSelectableTextView()
	tv.ScrollTo(5, 0)
	row, col := tv.GetScrollOffset()
	if row != 5 || col != 0 {
		t.Fatalf("expected (5,0), got (%d,%d)", row, col)
	}
	tv.ScrollTo(-1, 0)
	row, _ = tv.GetScrollOffset()
	if row != 0 {
		t.Fatalf("expected 0 after negative scroll, got %d", row)
	}
}

func TestSelectableTextView_ScrollToEnd(t *testing.T) {
	tv := NewSelectableTextView()
	tv.ScrollToEnd()
	row, col := tv.GetScrollOffset()
	if row != 0 || col != 0 {
		t.Fatalf("expected (0,0), got (%d,%d)", row, col)
	}
}

func TestSelectableTextView_CopySelection(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetText("hello world")
	tv.SelectAll()
	if !tv.HasSelection() {
		t.Fatal("SelectAll should create selection")
	}
	tv.CopySelection()
	if tv.HasSelection() {
		t.Fatal("CopySelection should clear selection")
	}
}

func TestSelectableTextView_CopySelectionNoSelection(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetText("hello world")
	tv.CopySelection() // should not panic when no selection
}

func TestSelectableTextView_DrawScrollToEnd(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetWordWrap(false)
	tv.SetRect(0, 0, 20, 3)

	lines := ""
	for i := 0; i < 10; i++ {
		lines += "line" + string(rune('0'+i)) + "\n"
	}
	tv.SetText(lines)

	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(20, 10)

	tv.ScrollToEnd()
	tv.Draw(screen)
	screen.Show()

	if tv.lastRenderedOffset < 7 {
		t.Fatalf("ScrollToEnd should show last lines, got offset %d (expected >= 7 for 10 lines in height 3)", tv.lastRenderedOffset)
	}
}

func TestSelectableTextView_GetScrollOffsetAfterScrollToEnd(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetRect(0, 0, 20, 5)

	lines := ""
	for i := 0; i < 20; i++ {
		lines += "line" + string(rune('0'+(i%10))) + "\n"
	}
	tv.SetText(lines)

	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(20, 10)

	tv.ScrollToEnd()
	tv.Draw(screen)
	screen.Show()

	row, col := tv.GetScrollOffset()
	if row == 0 && col == 0 && tv.lastRenderedOffset > 0 {
		t.Fatal("GetScrollOffset after ScrollToEnd should return actual rendered offset, not (0,0)")
	}
}

func TestSelectableTextView_DrawScrollOffsetApplied(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetWordWrap(false)
	tv.SetRect(0, 0, 20, 5)

	lines := ""
	for i := 0; i < 20; i++ {
		lines += "line" + string(rune('0'+(i%10))) + "\n"
	}
	tv.SetText(lines)

	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(20, 10)

	// Draw with scroll offset 0
	tv.ScrollTo(0, 0)
	tv.Draw(screen)
	screen.Show()
	if tv.lastRenderedOffset != 0 {
		t.Fatalf("expected lastRenderedOffset 0, got %d", tv.lastRenderedOffset)
	}

	// Draw with scroll offset 10
	tv.ScrollTo(10, 0)
	tv.Draw(screen)
	screen.Show()
	if tv.lastRenderedOffset != 10 {
		t.Fatalf("expected lastRenderedOffset 10, got %d", tv.lastRenderedOffset)
	}
}

func TestSelectableTextView_MouseSelectionWithScroll(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetWordWrap(false)
	tv.SetRect(0, 0, 20, 5)

	lines := ""
	for i := 0; i < 10; i++ {
		lines += "line" + string(rune('0'+i)) + "\n"
	}
	tv.SetText(lines)

	// Scroll down by 3 lines, then draw to update lastRenderedOffset
	tv.ScrollTo(3, 0)
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(20, 10)
	tv.Draw(screen)
	screen.Show()

	if tv.lastRenderedOffset != 3 {
		t.Fatalf("expected lastRenderedOffset 3 after ScrollTo(3,0), got %d", tv.lastRenderedOffset)
	}

	// MouseHandler adds lastRenderedOffset to visual line, so visual line 0 → absolute line 3
	// We verify lastRenderedOffset is correct; the MouseHandler wiring is tested separately
}

func TestSelectableTextView_MouseHandler(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetRect(0, 0, 20, 5)
	tv.SetText("hello world select this")
	tv.handleMousePress(0, 6)
	tv.handleMouseDrag(0, 20)
	tv.handleMouseRelease()

	if !tv.HasSelection() {
		t.Fatal("expected selection after mouse drag")
	}
	sel := tv.GetSelection()
	if len(sel) == 0 {
		t.Fatal("expected non-empty selection")
	}
}

// Regression: mouse click on chat panel must NOT steal focus from composer.
// Otherwise after clicking the chat area (e.g. after creating a new session)
// keyboard input no longer reaches the composer.
func TestSelectableTextView_MouseLeftDownDoesNotStealFocus(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetRect(0, 0, 20, 5)
	tv.SetText("hello world")

	var focused tview.Primitive
	setFocus := func(p tview.Primitive) { focused = p }

	handler := tv.MouseHandler()
	if handler == nil {
		t.Fatal("expected non-nil mouse handler")
	}
	ev := tcell.NewEventMouse(2, 1, tcell.Button1, tcell.ModNone)
	handler(tview.MouseLeftDown, ev, setFocus)

	if focused != nil {
		t.Fatalf("expected setFocus NOT to be called on MouseLeftDown, but was called with %T", focused)
	}
}

func TestSelectableTextView_Write(t *testing.T) {
	tv := NewSelectableTextView()
	n, err := tv.Write([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Fatalf("expected 5, got %d", n)
	}
	if tv.GetText(false) != "hello" {
		t.Fatalf("expected 'hello', got %q", tv.GetText(false))
	}
}
