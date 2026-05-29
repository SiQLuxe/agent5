package ui

import "testing"

func TestModel_Creation(t *testing.T) {
	m := NewModel()

	if m == nil {
		t.Error("Expected model to be created")
	}

	if m.tabDock == nil {
		t.Error("Expected tabDock to be initialized")
	}

	if m.chatPanel == nil {
		t.Error("Expected chatPanel to be initialized")
	}

	if len(m.sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(m.sessions))
	}

	if m.keyMap.Quit.Help().Key == "" {
		t.Error("Expected keyMap to be initialized")
	}
}

func TestModel_NewSession(t *testing.T) {
	m := NewModel()
	m.newSession()

	if len(m.sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(m.sessions))
	}
	if m.activeSession != 1 {
		t.Errorf("Expected active session 1, got %d", m.activeSession)
	}
}

func TestModel_CloseSession(t *testing.T) {
	m := NewModel()
	m.newSession()
	m.closeSession()

	if len(m.sessions) != 1 {
		t.Errorf("Expected 1 session after close, got %d", len(m.sessions))
	}
}

func TestModel_NextSession(t *testing.T) {
	m := NewModel()
	m.newSession()
	m.newSession()
	m.switchToSession(0)
	m.nextSession()

	if m.activeSession != 1 {
		t.Errorf("Expected active session 1, got %d", m.activeSession)
	}
}

func TestModel_PrevSession(t *testing.T) {
	m := NewModel()
	m.newSession()
	m.switchToSession(1)
	m.prevSession()

	if m.activeSession != 0 {
		t.Errorf("Expected active session 0, got %d", m.activeSession)
	}
}

func TestModel_IsLoading(t *testing.T) {
	m := NewModel()

	if m.IsLoading() != false {
		t.Error("Expected IsLoading to be false initially")
	}
}

func TestModel_AddChatMessage(t *testing.T) {
	m := NewModel()

	m.AddChatMessage("user", "Hello")
	msgs := m.GetChatMessages()

	if len(msgs) != 1 {
		t.Errorf("Expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Role != RoleUser {
		t.Errorf("Expected RoleUser, got %v", msgs[0].Role)
	}
}

func TestModel_SwitchToSession(t *testing.T) {
	m := NewModel()
	m.newSession()
	m.switchToSession(0)

	if m.activeSession != 0 {
		t.Errorf("Expected active session 0, got %d", m.activeSession)
	}
}
