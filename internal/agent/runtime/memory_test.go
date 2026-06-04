package runtime

import (
	"testing"
)

func TestAgentMemoryAppend(t *testing.T) {
	m := NewMemory()
	m.Append("user", "hello")
	m.Append("assistant", "hi")

	if len(m.ShortTerm) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(m.ShortTerm))
	}
	if m.ShortTerm[0].Role != "user" || m.ShortTerm[0].Content != "hello" {
		t.Fatalf("unexpected first message: %+v", m.ShortTerm[0])
	}
}

func TestAgentMemorySetGet(t *testing.T) {
	m := NewMemory()
	m.Set("key1", "value1")
	m.Set("key2", "value2")

	if m.Get("key1") != "value1" {
		t.Fatalf("expected 'value1', got %s", m.Get("key1"))
	}
	if m.Get("key3") != "" {
		t.Fatal("expected empty string for missing key")
	}
}

func TestAgentMemoryClear(t *testing.T) {
	m := NewMemory()
	m.Append("user", "hello")
	m.Set("key", "val")
	m.Clear()

	if len(m.ShortTerm) != 0 {
		t.Fatal("expected empty ShortTerm after clear")
	}
}
