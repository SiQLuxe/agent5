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

	// Remove trailing empty string from split if content ends with \n
	if oldContent != "" && strings.HasSuffix(oldContent, "\n") {
		oldLines = oldLines[:len(oldLines)-1]
	}
	if newContent != "" && strings.HasSuffix(newContent, "\n") {
		newLines = newLines[:len(newLines)-1]
	}

	var b strings.Builder
	fmt.Fprintf(&b, "--- %s\n", path)
	fmt.Fprintf(&b, "+++ %s\n", path)

	o, n := 0, 0
	for o < len(oldLines) || n < len(newLines) {
		// Skip matching lines
		for o < len(oldLines) && n < len(newLines) && oldLines[o] == newLines[n] {
			o++
			n++
		}

		if o >= len(oldLines) && n >= len(newLines) {
			break
		}

		// Start of a change hunk
		oldStart := o
		newStart := n
		for o < len(oldLines) || n < len(newLines) {
			if o < len(oldLines) && n < len(newLines) && oldLines[o] == newLines[n] {
				break
			}
			if o < len(oldLines) {
				o++
			}
			if n < len(newLines) {
				n++
			}
		}

		oldCount := o - oldStart
		newCount := n - newStart
		if oldCount > 0 && newCount > 0 {
			fmt.Fprintf(&b, "@@ -%d,%d +%d,%d @@\n", oldStart+1, oldCount, newStart+1, newCount)
		} else if oldCount > 0 {
			fmt.Fprintf(&b, "@@ -%d,%d +%d,0 @@\n", oldStart+1, oldCount, newStart+1)
		} else {
			fmt.Fprintf(&b, "@@ -%d,0 +%d,%d @@\n", oldStart+1, newStart+1, newCount)
		}

		for i := oldStart; i < oldStart+oldCount; i++ {
			fmt.Fprintf(&b, "-%s", oldLines[i])
		}
		for i := newStart; i < newStart+newCount; i++ {
			fmt.Fprintf(&b, "+%s", newLines[i])
		}
	}

	return b.String()
}
