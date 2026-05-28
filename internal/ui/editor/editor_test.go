package editor

import (
	"os"
	"testing"
)

func TestSetContent(t *testing.T) {
	ce := NewCodeEditor()
	ce.SetContent("func main() {}")
	
	if ce.GetContent() != "func main() {}" {
		t.Errorf("expected 'func main() {}', got '%s'", ce.GetContent())
	}
}

func TestCursorPosition(t *testing.T) {
	ce := NewCodeEditor()
	ce.SetCursorPosition(5, 10)
	
	line, col := ce.GetCursorPosition()
	if line != 5 {
		t.Errorf("expected line 5, got %d", line)
	}
	if col != 10 {
		t.Errorf("expected column 10, got %d", col)
	}
}

func TestInsertLine(t *testing.T) {
	ce := NewCodeEditor()
	ce.SetContent("line1\nline3")
	ce.InsertLine(1, "line2")
	
	lines := ce.GetLines()
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
	if lines[1] != "line2" {
		t.Errorf("expected 'line2' at index 1, got '%s'", lines[1])
	}
}

func TestDeleteLine(t *testing.T) {
	ce := NewCodeEditor()
	ce.SetContent("line1\nline2\nline3")
	ce.DeleteLine(1)
	
	lines := ce.GetLines()
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}
	if lines[1] != "line3" {
		t.Errorf("expected 'line3' at index 1, got '%s'", lines[1])
	}
}

func TestHighlight(t *testing.T) {
	ce := NewCodeEditor()
	ce.SetContent("package main")
	ce.SetLanguage("go")
	
	result := ce.Highlight()
	if result == "" {
		t.Error("highlight result should not be empty")
	}
}

func TestLoadSaveFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test*.go")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	testContent := "package test"
	os.WriteFile(tmpFile.Name(), []byte(testContent), 0644)
	
	ce := NewCodeEditor()
	err = ce.LoadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to load file: %v", err)
	}
	
	if ce.GetContent() != testContent {
		t.Errorf("expected '%s', got '%s'", testContent, ce.GetContent())
	}
	
	newContent := "package updated"
	ce.SetContent(newContent)
	err = ce.SaveFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to save file: %v", err)
	}
	
	data, _ := os.ReadFile(tmpFile.Name())
	if string(data) != newContent {
		t.Errorf("expected '%s', got '%s'", newContent, string(data))
	}
}