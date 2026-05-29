# AI Stream Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Connect the UI send flow to the existing AI Assistant streaming API, replacing the `Echo:` placeholder with real AI responses streamed character-by-character.

**Architecture:** `sendMessage` adds user + empty assistant message, spawns a goroutine calling `aiAssistant.ChatStream`, and the chunk callback uses `QueueUpdateDraw` (main goroutine) to append content to the last assistant message and refresh the panel. The `isLoading` flag prevents concurrent sends.

**Tech Stack:** Go, tview, `github.com/example/agent-tui/internal/service` (AIAssistant)

---

## File Structure

**Modified:**
- `internal/ui/app.go` — Add `aiAssistant *service.AIAssistant` field, update `SetAIAssistant` signature, rewrite `sendMessage`, add loading guard in `handleInput`
- `internal/ui/app_test.go` — Update tests for new streaming behavior

---

### Task 1: Add `aiAssistant` field and update `SetAIAssistant`

**Files:**
- Modify: `internal/ui/app.go:23-40`

- [ ] **Step 1: Write failing test for no-op when aiAssistant is nil**

```go
func TestSendMessageNoAssistant(t *testing.T) {
    a := NewApp()
    a.newSession()
    a.composer.SetInput("hello")
    a.sendMessage()
    // Without aiAssistant, sendMessage should be a no-op (guard)
    if a.composer.GetInput() != "hello" {
        t.Fatal("expected composer preserved when no ai assistant")
    }
    if s := a.activeSessionPtr(); s != nil {
        if len(s.Messages) != 0 {
            t.Fatalf("expected 0 messages with no ai assistant, got %d", len(s.Messages))
        }
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd internal/ui && go test -run TestSendMessageNoAssistant -v`
Expected: FAIL — `sendMessage` currently adds messages regardless of aiAssistant

- [ ] **Step 3: Add `aiAssistant` field to App struct and update `SetAIAssistant` signature**

Add to the import block in `internal/ui/app.go`:
```go
"github.com/example/agent-tui/internal/service"
```

Add field to `App` struct:
```go
type App struct {
    // ...existing fields...
    aiAssistant   *service.AIAssistant
}
```

Change `SetAIAssistant`:
```go
func (a *App) SetAIAssistant(ai *service.AIAssistant) {
    a.aiAssistant = ai
}
```

- [ ] **Step 4: Add guard in `sendMessage`**

Replace the beginning of `sendMessage`:
```go
func (a *App) sendMessage() {
    text := a.composer.GetInput()
    if strings.TrimSpace(text) == "" {
        return
    }
    if a.aiAssistant == nil {
        return
    }
    s := a.activeSessionPtr()
    if s == nil {
        return
    }
    // ... rest of sendMessage
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `cd internal/ui && go test -run TestSendMessageNoAssistant -v`
Expected: PASS

- [ ] **Step 6: Update existing `TestSendMessage` to set an aiAssistant**

Replace the test body:
```go
func TestSendMessage(t *testing.T) {
    a := NewApp()
    a.newSession()
    // Set up a real AIAssistant with mock client for synchronous response
    h := history.NewHistory("")
    sessionID := a.activeSessionPtr().ID
    mockClient := &MockAIClientForApp{
        mockResponse: "Hello!",
    }
    aiAssistant := service.NewAIAssistant(mockClient, h)
    a.SetAIAssistant(aiAssistant)

    a.composer.SetInput("hello")
    a.sendMessage()

    if a.composer.GetInput() != "" {
        t.Fatal("expected composer cleared after send")
    }
    if s := a.activeSessionPtr(); s != nil {
        msgs := s.Messages
        if len(msgs) != 2 {
            t.Fatalf("expected 2 messages (user + assistant), got %d", len(msgs))
        }
        if msgs[0].Role != RoleUser || msgs[0].Content != "hello" {
            t.Fatalf("unexpected first message: %+v", msgs[0])
        }
        if msgs[1].Role != RoleAssistant || msgs[1].Content != "" {
            t.Fatalf("expected empty assistant placeholder, got: %+v", msgs[1])
        }
    }
    if !a.IsLoading() {
        t.Fatal("expected loading after send")
    }
}
```

- [ ] **Step 7: Add mock client type for app tests in the test file**

Add at the top of `internal/ui/app_test.go`:
```go
import (
    "testing"
    "github.com/example/agent-tui/internal/ai"
    "github.com/example/agent-tui/internal/data/history"
    "github.com/example/agent-tui/internal/service"
)

type MockAIClientForApp struct {
    mockResponse string
}

func (m *MockAIClientForApp) ChatCompletion(req ai.ChatCompletionRequest) (*ai.ChatCompletionResponse, error) {
    return &ai.ChatCompletionResponse{}, nil
}

func (m *MockAIClientForApp) ChatCompletionStream(req ai.ChatCompletionRequest, callback func(string)) error {
    if m.mockResponse != "" {
        callback(m.mockResponse)
    }
    return nil
}

func (m *MockAIClientForApp) SetAPIKey(key string)    {}
func (m *MockAIClientForApp) SetBaseURL(url string)    {}
func (m *MockAIClientForApp) GetModel() string          { return "mock" }
func (m *MockAIClientForApp) ListModels() ([]string, error) { return []string{"mock"}, nil }
```

- [ ] **Step 8: Run all app tests to verify**

Run: `cd internal/ui && go test -v`
Expected: All tests PASS (some tests that call sendMessage without aiAssistant may still fail — we'll fix those inline as needed)

- [ ] **Step 9: Commit**

```bash
git add internal/ui/app.go internal/ui/app_test.go
git commit -m "feat: add aiAssistant field with nil guard in sendMessage"
```

---

### Task 2: Add loading guard in handleInput

**Files:**
- Modify: `internal/ui/app.go:172-178`

- [ ] **Step 1: Write failing test for blocked send during loading**

```go
func TestSendMessageBlockedDuringLoading(t *testing.T) {
    a := NewApp()
    a.newSession()
    a.SetLoading(true)
    a.composer.SetInput("hello")

    // Simulate Enter key press
    ev := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
    result := a.handleInput(ev)

    if result != nil {
        t.Fatal("expected Enter to be consumed (nil) during loading")
    }
    if s := a.activeSessionPtr(); s != nil {
        if len(s.Messages) != 0 {
            t.Fatal("expected no messages added during loading")
        }
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd internal/ui && go test -run TestSendMessageBlockedDuringLoading -v`
Expected: FAIL — Enter currently sends even when loading (test expects nil/blocked)

- [ ] **Step 3: Add loading guard in `handleInput`**

In the Enter handler case (`case event.Key() == tcell.KeyEnter && event.Modifiers() == tcell.ModNone:`), add a loading check:

```go
case event.Key() == tcell.KeyEnter && event.Modifiers() == tcell.ModNone:
    if a.isLoading {
        return nil
    }
    if strings.TrimSpace(a.composer.GetInput()) == "" {
        return nil
    }
    a.sendMessage()
    return nil
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd internal/ui && go test -run TestSendMessageBlockedDuringLoading -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ui/app.go internal/ui/app_test.go
git commit -m "feat: block Enter during AI loading state"
```

---

### Task 3: Rewrite `sendMessage` with streaming

**Files:**
- Modify: `internal/ui/app.go:333-345`

- [ ] **Step 1: Write failing test for streaming callback**

```go
func TestSendMessageStreamsContent(t *testing.T) {
    a := NewApp()
    a.newSession()
    h := history.NewHistory("")
    mockClient := &MockAIClientForApp{
        mockResponse: "Hello world",
    }
    aiAssistant := service.NewAIAssistant(mockClient, h)
    a.SetAIAssistant(aiAssistant)

    a.composer.SetInput("hi")
    a.sendMessage()

    // At this point, user + empty assistant messages exist
    s := a.activeSessionPtr()

    // The goroutine should have completed (mock is synchronous),
    // so the assistant message should now have content
    if len(s.Messages) != 2 {
        t.Fatalf("expected 2 messages, got %d", len(s.Messages))
    }
    if s.Messages[1].Content != "Hello world" {
        t.Fatalf("expected streaming content 'Hello world', got '%s'", s.Messages[1].Content)
    }
    if a.IsLoading() {
        t.Fatal("expected loading cleared after stream completes")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd internal/ui && go test -run TestSendMessageStreamsContent -v`
Expected: FAIL — `sendMessage` still uses `Echo:` placeholder

- [ ] **Step 3: Implement streaming `sendMessage`**gh

Replace the `sendMessage` function body:

```go
func (a *App) sendMessage() {
    text := a.composer.GetInput()
    if strings.TrimSpace(text) == "" {
        return
    }
    if a.aiAssistant == nil {
        return
    }
    s := a.activeSessionPtr()
    if s == nil {
        return
    }

    s.AddMessage(RoleUser, text)
    s.AddMessage(RoleAssistant, "") // placeholder for streaming
    a.composer.ClearInput()
    a.chatPanel.SetSession(s)
    a.isLoading = true

    sessionID := s.ID
    go func() {
        var fullResponse string
        err := a.aiAssistant.ChatStream(sessionID, text, func(chunk string) {
            fullResponse += chunk
            a.QueueUpdateDraw(func() {
                if s := a.activeSessionPtr(); s != nil && len(s.Messages) > 0 {
                    s.Messages[len(s.Messages)-1].Content = fullResponse
                    a.chatPanel.SetSession(s)
                }
            })
        })

        a.QueueUpdateDraw(func() {
            a.isLoading = false
            if err != nil {
                if s := a.activeSessionPtr(); s != nil {
                    s.AddMessage(RoleSystem, "Error: "+err.Error())
                }
            }
            a.chatPanel.SetSession(a.activeSessionPtr())
            a.chatPanel.ScrollToBottom()
        })
    }()
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd internal/ui && go test -run TestSendMessageStreamsContent -v`
Expected: PASS

- [ ] **Step 5: Run all app tests**

Run: `cd internal/ui && go test -v`
Expected: All tests PASS

- [ ] **Step 6: Commit**

```bash
git add internal/ui/app.go internal/ui/app_test.go
git commit -m "feat: replace Echo placeholder with AI streaming via ChatStream"
```

---

### Task 4: Full build and smoke test

- [ ] **Step 1: Build the application**

Run: `go build ./cmd/agent/`
Expected: Binary builds without errors

- [ ] **Step 2: Run all tests in the project**

Run: `go test ./...`
Expected: All tests pass
