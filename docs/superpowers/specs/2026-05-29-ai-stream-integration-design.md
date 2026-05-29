# AI Stream Integration Design

**Date:** 2026-05-29
**Status:** Draft

**Goal:** Connect the UI send flow to the existing AI Assistant streaming API, replacing the `Echo:` placeholder with real AI responses streamed character-by-character.

## Architecture

```
sendMessage()
  → UI: add user message + empty assistant placeholder
  → go aiAssistant.ChatStream(sessionID, text, chunkCallback)
      (runs in background goroutine)
      → chunkCallback on each chunk (from bg goroutine):
          QueueUpdateDraw { append chunk to last message; refresh panel }
      → ChatStream returns (bg goroutine):
          QueueUpdateDraw { clear loading; show error if any; scroll to end }
```

**Concurrency:** `Session.Messages` is only mutated inside `QueueUpdateDraw` callbacks (main goroutine), so no mutex is needed. The chunk callback runs in a background goroutine but only queues work — it never touches UI state directly.

## Components

### App changes (`internal/ui/app.go`)

**New field:**
```go
aiAssistant *service.AIAssistant
```

**SetAIAssistant:** Store the `*service.AIAssistant` reference instead of no-op.

**sendMessage rewrite:**
1. Guard: if no active session or no aiAssistant, return
2. Add `RoleUser` message with input text
3. Add `RoleAssistant` with empty content (placeholder for streaming)
4. Clear composer, refresh panel, set `isLoading = true`
5. Spawn goroutine calling `aiAssistant.ChatStream(s.ID, text, callback)`
6. Each callback call: append chunk to fullResponse, `QueueUpdateDraw` to update the last assistant message content and refresh panel
7. On completion/error: `QueueUpdateDraw` to clear loading, show error message in chat if any, scroll to end

### Key bindings (`internal/ui/app.go` handleInput)

- During `isLoading`, block Enter (prevent sending while AI is responding):
  ```go
  case event.Key() == tcell.KeyEnter && event.Modifiers() == tcell.ModNone:
      if a.isLoading { return nil }
      // existing send logic...
  ```

### No changes needed

- `internal/ui/session.go` — UI rendering is already driven by `Session.Messages`
- `internal/ui/chatpanel.go` — `SetSession` already refreshes the view
- `internal/ui/composer/` — no changes needed
- `internal/ui/status/` — no changes needed
- `internal/ui/tabbar/` — no changes needed
- `internal/ui/chatpanel.go` — no changes needed
