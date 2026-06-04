package orchestrator

import (
	"testing"
)

func TestDecomposerPassthrough(t *testing.T) {
	d := NewDecomposer()
	task := &Task{ID: "t1", Type: TaskCustom, Content: "do something"}
	steps, err := d.Decompose(task)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}
	if len(steps) != 1 {
		t.Fatalf("expected 1 step (passthrough), got %d", len(steps))
	}
}

func TestDecomposerSequential(t *testing.T) {
	d := NewDecomposer()
	d.Register(TaskAnalyze, DefaultSequentialStrategy)
	task := &Task{ID: "t1", Type: TaskAnalyze, Content: "add login"}
	steps, err := d.Decompose(task)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}
	if len(steps) != 4 {
		t.Fatalf("expected 4 steps, got %d", len(steps))
	}
	if steps[0].Type != TaskAnalyze {
		t.Fatalf("first step should be analyze")
	}
	if steps[3].Type != TaskReview {
		t.Fatalf("last step should be review")
	}
}
