# Composer Transparent Background Design

**Date:** 2026-05-31
**Status:** Draft

**Goal:** Remove the dark background block from the composer input area (bottom-left input box), making it use the terminal default background so terminal-level transparency can pass through.

## Background

The composer area (prompt `"> "` + `tview.TextArea`) currently renders with a visible dark background block. The root app already uses `tcell.ColorDefault`, but the composer's child widgets (TextArea, prompt TextView, Flex container) independently draw their backgrounds, creating the dark block.

## Design

Set `SetBackgroundColor(tcell.ColorDefault)` on all three composer sub-widgets in `New()`:

| Widget | Current | After |
|---|---|---|
| `TextArea` | `tcell.ColorDefault` (via style, but widget bg may differ per tview version) | explicit `SetBackgroundColor(tcell.ColorDefault)` |
| `prompt TextView` | default (inherits from Flex) | `SetBackgroundColor(tcell.ColorDefault)` |
| `Flex` (composer root) | default | `SetBackgroundColor(tcell.ColorDefault)` |

### File changed

`internal/ui/composer/composer.go` — add three `.SetBackgroundColor(tcell.ColorDefault)` calls at end of `New()`.

### Fallback

If this alone is insufficient (e.g., tview still paints a background), `SetBackgroundColor` exposes a clean overridable point; the caller (`applyTheme`) can always call `composer.SetBackgroundColor(tcell.ColorDefault)` as a second line of defense.

## What's NOT changing

- `applyTheme()` — `InputBg` theme color stays unused but available.
- `SetBackgroundColor` method — remains for future theme control.
- Tests — none test background color.

## Verification

1. `go build ./...` compiles
2. Run the TUI; the input box area shows no distinct background block
3. If user's terminal has transparency configured, the background behind the terminal window should show through the input area
