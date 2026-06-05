package save

import (
	"fmt"
	"strings"
)

func UnifiedDiff(path, oldContent, newContent string) string {
	if oldContent == newContent {
		return ""
	}

	oldLines := strings.SplitAfter(oldContent, "\n")
	newLines := strings.SplitAfter(newContent, "\n")

	if len(oldContent) > 0 && !strings.HasSuffix(oldContent, "\n") {
		oldLines = append(oldLines, "\n\\ No newline at end of file\n")
	}
	if len(newContent) > 0 && !strings.HasSuffix(newContent, "\n") {
		newLines = append(newLines, "\n\\ No newline at end of file\n")
	}

	var b strings.Builder
	fmt.Fprintf(&b, "--- %s\n", path)
	fmt.Fprintf(&b, "+++ %s\n", path)

	maxLen := len(oldLines)
	if len(newLines) > maxLen {
		maxLen = len(newLines)
	}

	for i := 0; i < maxLen; i++ {
		oldLine := ""
		newLine := ""
		if i < len(oldLines) {
			oldLine = oldLines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}
		if oldLine != newLine || i >= len(oldLines) || i >= len(newLines) {
			if i == 0 {
				fmt.Fprintf(&b, "@@ -1,%d +1,%d @@\n", len(oldLines), len(newLines))
			}
			if i < len(oldLines) {
				fmt.Fprintf(&b, "-%s", oldLine)
			}
			if i < len(newLines) {
				fmt.Fprintf(&b, "+%s", newLine)
			}
		}
	}

	result := b.String()
	return result
}
