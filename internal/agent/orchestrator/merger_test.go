package orchestrator

import (
	"strings"
	"testing"
)

func TestConcatMerge(t *testing.T) {
	results := []*Task{
		{ID: "1", Type: TaskAnalyze, Status: StatusCompleted, Result: "Analysis done"},
		{ID: "2", Type: TaskCode, Status: StatusCompleted, Result: "Code done"},
		{ID: "3", Type: TaskReview, Status: StatusFailed, Error: "bugs found"},
	}
	output := ConcatMerge(results)
	if !strings.Contains(output, "Analysis done") {
		t.Fatal("expected Analysis done in output")
	}
	if !strings.Contains(output, "FAILED") {
		t.Fatal("expected FAILED marker")
	}
}
