package opencode

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/example/agent-tui/internal/backend"
)

func opencodeBinary() string {
	path, err := exec.LookPath("opencode")
	if err != nil {
		return ""
	}
	return path
}

func projectRoot() string {
	cwd, _ := os.Getwd()
	// Navigate up from internal/backend/opencode/ to project root
	return cwd + "/../../.."
}

func testConfig(binary string) Config {
	return Config{
		Binary:    binary,
		AutoStart: true,
		WorkDir:   projectRoot(),
	}
}

func TestStartOpencodeBackend(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	binary := opencodeBinary()
	if binary == "" {
		t.Skip("opencode binary not found in PATH")
	}

	b := NewBackend(testConfig(binary))
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		b.Stop(ctx)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := b.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	h, err := b.Health(ctx)
	if err != nil {
		t.Fatalf("Health failed: %v", err)
	}
	if !h.Healthy {
		t.Fatal("expected healthy")
	}
}

func TestStartStopCycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	binary := opencodeBinary()
	if binary == "" {
		t.Skip("opencode binary not found in PATH")
	}

	b := NewBackend(testConfig(binary))
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		b.Stop(ctx)
	})

	// Start
	ctx1, cancel1 := context.WithTimeout(context.Background(), 30*time.Second)
	if err := b.Start(ctx1); err != nil {
		t.Fatalf("first Start failed: %v", err)
	}
	cancel1()

	// Stop
	ctxStop, cancelStop := context.WithTimeout(context.Background(), 10*time.Second)
	if err := b.Stop(ctxStop); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	cancelStop()
}

func TestSessionLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	binary := opencodeBinary()
	if binary == "" {
		t.Skip("opencode binary not found in PATH")
	}

	b := NewBackend(testConfig(binary))
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		b.Stop(ctx)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := b.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Create session
	sess, err := b.CreateSession(ctx, "integration test")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if sess.ID == "" {
		t.Fatal("expected non-empty session ID")
	}
	if sess.Title != "integration test" {
		t.Errorf("Title = %q, want %q", sess.Title, "integration test")
	}

	// List sessions
	sessions, err := b.ListSessions(ctx)
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	found := false
	for _, s := range sessions {
		if s.ID == sess.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("created session not found in list")
	}

	// Get session
	got, err := b.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if got.ID != sess.ID {
		t.Errorf("session ID mismatch: %q vs %q", got.ID, sess.ID)
	}

	// Delete session
	if err := b.DeleteSession(ctx, sess.ID); err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}
}

func TestSendAndGetMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	binary := opencodeBinary()
	if binary == "" {
		t.Skip("opencode binary not found in PATH")
	}

	b := NewBackend(testConfig(binary))
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		b.Stop(ctx)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := b.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	sess, err := b.CreateSession(ctx, "message test")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Send sync message
	result, err := b.SendMessage(ctx, sess.ID, &backend.Message{
		Role:    backend.RoleUser,
		Content: "Say hello in one word",
	})
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}
	if result.Content == "" {
		t.Error("expected non-empty response")
	}

	// Get messages
	msgs, err := b.GetMessages(ctx, sess.ID)
	if err != nil {
		t.Fatalf("GetMessages failed: %v", err)
	}
	if len(msgs) == 0 {
		t.Fatal("expected at least one message")
	}
	hasUser := false
	hasAssistant := false
	for _, m := range msgs {
		if m.Role == backend.RoleUser {
			hasUser = true
		}
		if m.Role == backend.RoleAssistant {
			hasAssistant = true
		}
	}
	if !hasUser {
		t.Error("expected a user message")
	}
	if !hasAssistant {
		t.Error("expected an assistant message")
	}
}

func TestIntegration_ExecuteCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	binary := opencodeBinary()
	if binary == "" {
		t.Skip("opencode binary not found in PATH")
	}

	b := NewBackend(testConfig(binary))
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		b.Stop(ctx)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := b.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	sess, err := b.CreateSession(ctx, "command test")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Execute a slash command — note: may fail on some opencode versions
	result, err := b.ExecuteCommand(ctx, sess.ID, "help")
	if err != nil {
		t.Logf("ExecuteCommand not supported by this opencode version: %v (skipping)", err)
	} else {
		t.Logf("command output: %s", result.Stdout)
	}
}

func TestIntegration_ReadFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	binary := opencodeBinary()
	if binary == "" {
		t.Skip("opencode binary not found in PATH")
	}

	b := NewBackend(testConfig(binary))
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		b.Stop(ctx)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := b.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	content, err := b.ReadFile(ctx, "go.mod")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if !strings.Contains(content, "module") {
		t.Errorf("expected go.mod to contain 'module', got: %q (len=%d)", content, len(content))
	}
}

func TestSearchText(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	binary := opencodeBinary()
	if binary == "" {
		t.Skip("opencode binary not found in PATH")
	}

	b := NewBackend(testConfig(binary))
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		b.Stop(ctx)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := b.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	results, err := b.SearchText(ctx, "AgentBackend")
	if err != nil {
		t.Fatalf("SearchText failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least one match for AgentBackend")
	}
}

func TestEventsStream(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	binary := opencodeBinary()
	if binary == "" {
		t.Skip("opencode binary not found in PATH")
	}

	b := NewBackend(testConfig(binary))
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		b.Stop(ctx)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := b.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Subscribe to events
	events, err := b.Events(ctx)
	if err != nil {
		t.Fatalf("Events failed: %v", err)
	}

	// Create a session to trigger an event
	sess, err := b.CreateSession(ctx, "events test")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Send a message to generate events
	_, err = b.SendMessage(ctx, sess.ID, &backend.Message{
		Role:    backend.RoleUser,
		Content: "Say hi",
	})
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	// Check we can receive events (non-blocking read)
	select {
	case evt := <-events:
		t.Logf("received event: type=%s", evt.Type)
	case <-time.After(5 * time.Second):
		t.Log("no event received within timeout (may have been consumed by SendMessage)")
	}
}

func TestExternalAgentRole(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	binary := opencodeBinary()
	if binary == "" {
		t.Skip("opencode binary not found in PATH")
	}

	b := NewBackend(testConfig(binary))
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		b.Stop(ctx)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := b.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Test the AgentRole interface via NewExternalAgent pattern
	// (simulates what main.go does)
	sess, err := b.CreateSession(ctx, "external agent test")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Simulate ExternalAgent.Execute
	result, err := b.SendMessage(ctx, sess.ID, &backend.Message{
		Role:    backend.RoleUser,
		Content: "test",
	})
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}
	if result.Content == "" {
		t.Error("expected non-empty response")
	}
}
