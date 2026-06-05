# Mouse Wheel Scrolls ChatPanel

## Problem

Mouse wheel scrolling triggers input history navigation (up/down through past
messages) instead of scrolling the chat output area. This happens because tcell
translates mouse wheel events into `KeyUp`/`KeyDown` sequences, and
`handleInput` intercepts them as history navigation.

## Goal

- Mouse wheel scrolls the ChatPanel content area
- Only keyboard ↑/↓ arrows navigate input history
- No other mouse or keyboard behavior changes

## Solution

Enable terminal mouse tracking and intercept `MouseWheelUp`/`MouseWheelDown`
events at the application level, routing them to ChatPanel scrolling methods.

### Changes

**File:** `internal/ui/app.go`, in `NewApp()`

1. **`EnableMouse(true)`** — tells the terminal to send mouse scroll events as
   proper `MouseWheelUp`/`MouseWheelDown` events instead of translating them
   to key sequences. Existing click events (tab switching, focus changes) are
   unaffected.

2. **`SetMouseCapture`** — intercepts mouse wheel events after `EnableMouse`:
   ```go
   a.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
       switch action {
       case tview.MouseWheelUp:
           a.chatPanel.SetAutoScroll(false)
           a.chatPanel.ScrollUp(3)
           return 0, nil
       case tview.MouseWheelDown:
           a.chatPanel.ScrollDown(3)
           return 0, nil
       }
       return action, event
   })
   ```

### Behavior

| Input | Effect |
|-------|--------|
| Mouse wheel up | ChatPanel scrolls up 3 lines, auto-scroll disabled |
| Mouse wheel down | ChatPanel scrolls down 3 lines, auto-scroll unchanged |
| Keyboard ↑/↓ | History navigation (unchanged) |
| Other mouse events (click, drag) | Pass through normally |

### Edge Cases

- **No session / empty chat:** `ScrollUp`/`ScrollDown` are no-ops when the
  panel has no content (tview handles this internally).
- **Terminal without mouse tracking:** Falls back to current behavior (scroll
  wheel triggers history nav). Affects only very old terminals; all modern
  terminals support mouse tracking including macOS Terminal.app, iTerm2,
  Kitty, Alacritty, xterm.
- **Tab bar clicks:** Unchanged — `tabbar.SetMouseCapture` continues to
  receive `MouseLeftClick` events as before.
- **Text selection in TextArea:** `EnableMouse` may affect text selection in
  some terminals; this is acceptable as the primary interaction model is
  keyboard-driven editing.
