package ui

import (
	"testing"
)

func TestApprovalModalCreation(t *testing.T) {
	modal := NewApprovalModal()
	if modal == nil {
		t.Fatal("expected non-nil ApprovalModal")
	}
}

func TestApprovalModalSetContent(t *testing.T) {
	modal := NewApprovalModal()
	modal.SetContent("test.txt", "--- test.txt\n+++ test.txt\n@@ -1 +1 @@\n-old\n+new\n")
}

func TestApprovalModalResult(t *testing.T) {
	modal := NewApprovalModal()
	modal.SetContent("test.txt", "diff content")
	if modal.Result() {
		t.Fatal("expected false before approval")
	}
}
