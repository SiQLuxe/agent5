# Mouse Wheel Scrolls Only ChatPanel

## Problem

Mouse wheel scrolling currently captures ALL scroll events at the Application
level (`SetMouseCapture`) and unconditionally routes them to ChatPanel,
regardless of where the mouse cursor is on screen. This means scrolling in the
input area (composer), tab bar, or status bar still scrolls the chat panel.

## Goal

- Mouse wheel scrolls ChatPanel **only when the mouse cursor is over the chat
  area**
- Mouse wheel over other areas (composer, tab bar, status bar) does nothing
  (follows tview's default behavior)
- Only keyboard ↑/↓ arrows navigate input history
- No other mouse or keyboard behavior changes

## Solution

Use **per-primitive** `SetMouseCapture` on ChatPanel instead of intercepting
events at the Application level. tview's built-in event routing automatically
dispatches mouse events to the correct primitive based on cursor position
(checked via `InRect` in each primitive's `WrapMouseHandler`).

### Changes

**File:** `internal/ui/app.go`, in `NewApp()`

1. **Remove** `MouseScrollUp`/`MouseScrollDown` handling from Application-level
   `SetMouseCapture` — the capture function becomes a no-op pass-through.

2. **Add** `a.chatPanel.SetMouseCapture(...)` to handle scroll events only when
   they reach the ChatPanel (which only happens when the mouse is over its area):

   ```go
   a.chatPanel.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
       switch action {
       case tview.MouseScrollUp:
           a.chatPanel.SetAutoScroll(false)
           a.chatPanel.ScrollUp(3)
           return tview.MouseConsumed, nil
       case tview.MouseScrollDown:
           a.chatPanel.ScrollDown(3)
           return tview.MouseConsumed, nil
       }
       return action, event
   })
   ```

### How Event Routing Works

```
Mouse scroll event
  → tcell receives MouseWheelUp/Down
  → Application.fireMouseActions()
    → App-level SetMouseCapture (no-op for scroll, unchanged for other events)
    → Route to root → pages → chatFlex
      → chatFlex.Flex.MouseHandler() iterates children in Z-order:
        → tabDock: InRect? No (or doesn't handle scroll) → skip
        → composer: InRect? No → skip
        → suggestionMenu: hidden → skip
        → chatPanel: InRect? Yes → calls chatPanel.SetMouseCapture
          → returns MouseConsumed → scroll handled
        → statusBar: skip
```

No coordinate arithmetic needed — tview's `WrapMouseHandler` / `Flex` routing
handles the hit-testing natively.

### Behavior

| Input | Cursor position | Effect |
|-------|----------------|--------|
| Mouse wheel up | Over chat area | ChatPanel scrolls up 3 lines, auto-scroll disabled |
| Mouse wheel up | Over composer/tabBar/statusBar | No-op |
| Mouse wheel down | Over chat area | ChatPanel scrolls down 3 lines |
| Mouse wheel down | Over composer/tabBar/statusBar | No-op |
| Keyboard ↑/↓ | Anywhere | History navigation (unchanged) |
| Other mouse events | Anywhere | Pass through normally |

### Edge Cases

- **No session / empty chat:** `ScrollUp`/`ScrollDown` are no-ops (tview
  handles internally).
- **Terminal without mouse tracking:** Falls back to default behavior.
- **Tab bar clicks:** Unchanged — `tabbar.SetMouseCapture` handles
  `MouseLeftClick`.
- **Text selection in TextArea:** Unchanged.
- **Search/help/rename overlay active:** Overlay page handles its own mouse
  routing; scroll events don't reach the underlying chatFlex.
- **Window resize:** `GetRect`/`InRect` adapt automatically.
