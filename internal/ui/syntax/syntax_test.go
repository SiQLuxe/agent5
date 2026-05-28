package syntax

import "testing"

func TestHighlight(t *testing.T) {
	code := `func main() { fmt.Println("Hello") }`
	result := Highlight(code, "go")

	if len(result) == 0 {
		t.Error("Expected non-empty result")
	}
}

func TestHighlightWithStyle(t *testing.T) {
	code := `print("hello")`
	result := HighlightWithStyle(code, "python", "monokai")

	if len(result) == 0 {
		t.Error("Expected non-empty result")
	}
}

func TestDetectLanguage(t *testing.T) {
	code := `func main() {}`
	result := DetectLanguage(code)

	if result == "" {
		t.Error("Expected language detection to work")
	}
}