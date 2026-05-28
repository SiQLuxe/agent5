package status

import "testing"

func TestStatusBar_SetMode(t *testing.T) {
	sb := NewStatusBar()
	sb.SetMode("Chat")

	if sb.mode != "Chat" {
		t.Errorf("Expected mode 'Chat', got '%s'", sb.mode)
	}
}

func TestStatusBar_SetConnected(t *testing.T) {
	sb := NewStatusBar()
	sb.SetConnected(false)

	if sb.connection != false {
		t.Error("Expected connection to be false")
	}
}

func TestStatusBar_SetTaskCount(t *testing.T) {
	sb := NewStatusBar()
	sb.SetTaskCount(5)

	if sb.taskCount != 5 {
		t.Errorf("Expected task count 5, got %d", sb.taskCount)
	}
}

func TestStatusBar_View(t *testing.T) {
	sb := NewStatusBar()
	result := sb.View(80)

	if len(result) == 0 {
		t.Error("Expected non-empty view")
	}
}