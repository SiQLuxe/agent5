package runtime

import (
	"testing"
)

func TestAgentLogger(t *testing.T) {
	log := NewLogger(100)
	log.Log("thought", "I should read the file", "", nil, nil, 0)

	if len(log.Entries()) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(log.Entries()))
	}
	if log.Entries()[0].Phase != "thought" {
		t.Fatalf("expected phase 'thought', got %s", log.Entries()[0].Phase)
	}
}

func TestAgentLoggerMaxEntries(t *testing.T) {
	log := NewLogger(3)
	for i := 0; i < 5; i++ {
		log.Log("thought", "msg", "", nil, nil, 0)
	}
	if len(log.Entries()) != 3 {
		t.Fatalf("expected 3 entries (max), got %d", len(log.Entries()))
	}
}

func TestAgentLoggerClear(t *testing.T) {
	log := NewLogger(100)
	log.Log("thought", "msg", "", nil, nil, 0)
	log.Clear()
	if len(log.Entries()) != 0 {
		t.Fatal("expected 0 entries after clear")
	}
}
