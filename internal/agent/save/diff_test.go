package save

import (
	"strings"
	"testing"
)

func TestUnifiedDiffNewFile(t *testing.T) {
	oldContent := ""
	newContent := "line1\nline2\n"
	diff := UnifiedDiff("test.txt", oldContent, newContent)

	if !strings.Contains(diff, "+line1") {
		t.Fatalf("expected '+line1' in diff, got:\n%s", diff)
	}
	if !strings.Contains(diff, "+line2") {
		t.Fatalf("expected '+line2' in diff, got:\n%s", diff)
	}
}

func TestUnifiedDiffModifiedFile(t *testing.T) {
	oldContent := "aaa\nbbb\nccc\n"
	newContent := "aaa\nmodified\nccc\n"
	diff := UnifiedDiff("test.txt", oldContent, newContent)

	if !strings.Contains(diff, "-bbb") {
		t.Fatalf("expected '-bbb' in diff, got:\n%s", diff)
	}
	if !strings.Contains(diff, "+modified") {
		t.Fatalf("expected '+modified' in diff, got:\n%s", diff)
	}
}

func TestUnifiedDiffNoChange(t *testing.T) {
	content := "same\ncontent\n"
	diff := UnifiedDiff("test.txt", content, content)
	if diff != "" {
		t.Fatalf("expected empty diff for unchanged content, got:\n%s", diff)
	}
}
