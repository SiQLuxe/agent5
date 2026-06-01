# Skill System Design

**Date:** 2026-06-01
**Status:** Draft

**Goal:** Add a skill system to the agent5 TUI, allowing users to invoke reusable capabilities via `/` commands, an overlay menu, and Tab completion. Skills follow the [agentskills.io](https://agentskills.io/specification) directory/frontmatter standard.

## Background

The agent5 TUI currently runs as a simple AI chat client. Supporting infrastructure exists (PluginRegistry, AgentRoles, TaskOrchestrator, CollaborationManager) but is not wired into the running app. This spec replaces PluginRegistry with a proper Skill system and connects it to the UI.

## Directory Structure

Skills live under `skills/` at the project root, one subdirectory per skill:

```
skills/
  code-review/
    SKILL.md              # Required: agentskills.io frontmatter + body
    scripts/              # Optional: executable code
    references/           # Optional: reference docs
    assets/              # Optional: templates, resources
  analyze/
    SKILL.md
  ...
```

## SKILL.md Format

Strictly follows [agentskills.io specification](https://agentskills.io/specification):

```markdown
---
name: code-review
description: Review Go code for bugs, performance issues, and style problems.
---
You are a senior code reviewer. Review the following code...
```

### Frontmatter fields

| Field | Required | Notes |
|-------|----------|-------|
| `name` | Yes | 1-64 chars, lowercase+hyphens, must match directory name |
| `description` | Yes | Max 1024 chars, describes what and when to use |
| `license` | No | License name or bundled file reference |
| `compatibility` | No | Environment requirements |
| `metadata` | No | Arbitrary key-value pairs |
| `allowed-tools` | No | Experimental, space-separated tool list |

### prompt vs handler distinction

No custom frontmatter fields. Distinction is made in Go code:

- **Default** (no registered handler): SKILL.md body is used as system prompt sent to AIAssistant.Chat()
- **Handler registered**: A Go function registered by name executes instead

```go
registry.RegisterHandler("execute-go", func(ctx SkillContext) string {
    // Run Go code via debugger
})
```

## Components

| Package | File | Responsibility |
|---------|------|----------------|
| `service` | `skill_registry.go` | CRUD for Skill objects, replaces PluginRegistry |
| `service` | `skill_loader.go` | Scans `skills/*/SKILL.md`, parses frontmatter, registers skills |
| `service` | `skill_parser.go` | Parses `/command args` from input text |
| `service` | `skill_executor.go` | Dispatches execution: prompt → LLM, handler → Go func |
| `ui` | `skill_overlay.go` | tview-based overlay list with filter, navigation, Tab completion |
| `ui` | `app.go` | ModeSkill state, `/` detection, Ctrl+P shortcut |

## Data Flow

```
User types "/review main.go"
  → app.handleInput detects leading '/'
  → enters ModeSkill, shows skill_overlay
  → real-time filter as user types
  → Enter selects / Tab completes
  → skill_parser.ParseCommand("/review main.go")
      → {Name: "review", Args: "main.go"}
  → skill_executor.Execute("review", "main.go")
      → type=prompt → AIAssistant.Chat(skill.Prompt + args)
      → type=handler → skill.Handler(SkillContext{Args: "main.go"})
  → result written to Session.Messages (new RoleSkill)
  → ChatPanel refreshes display
```

## UI: Skill Overlay

A `tview.Flex` component dynamically inserted into the chatFlex layout above the composer. Not a full-screen Pages overlay — only occupies the space needed for the skill list.

```
┌─────────────────────────────────┐
│ 技能列表 (3/8)                  │
│ ┌───────────────────────────┐   │
│ │ ✓ code-review   代码审查   │   │
│ │   execute-go    执行 Go   │   │
│ │   analyze       项目分析   │   │
│ └───────────────────────────┘   │
│ Tab:补全  ↑↓:导航  Enter:执行   │
│ Esc:关闭                        │
└─────────────────────────────────┘
     ↑ overlay row inserted here
┌─────────────────────────────────┐
│ >  /review                       │
└─────────────────────────────────┘
```

Show: `chatFlex.AddItem(overlay, overlayHeight, 0, false)`  
Hide: `chatFlex.RemoveItem(overlay)`

### Keyboard

| Key | Action |
|-----|--------|
| `/` at input start | Enter ModeSkill, show overlay |
| Typing | Filter skill list (prefix match) |
| `Tab` | Complete selected skill name |
| `↑` `↓` | Navigate overlay list |
| `Enter` | Execute selected skill |
| `Esc` | Exit ModeSkill, return to normal input |
| `Ctrl+P` | Open skill overlay without `/` |

## Skill Message Rendering

New message role `RoleSkill` in `session.go`. Messages render with distinct styling (e.g., accent/purple color) to differentiate from user and assistant messages.

```
[技能] code-review
──────────────────────────────
Review completed. Found 3 issues:
1. main.go:42 - unhandled error
```

## Error Handling

| Scenario | Behavior |
|----------|----------|
| `skills/` missing | Silent skip, no error |
| Invalid SKILL.md frontmatter | Skip skill, log warning |
| name != directory name | Skip skill, log warning |
| LLM error during execution | Write error as chat message |
| No matching skill | Keep `/xxx` text, no overlay |

## Cleanup

- `internal/service/plugin_registry.go` → replaced by `skill_registry.go`

## Testing

| Component | Approach |
|-----------|----------|
| `skill_loader` | Test fixtures with valid/invalid SKILL.md files |
| `skill_parser` | Unit tests for `/cmd`, `/cmd args`, empty input |
| `skill_registry` | Register/query/list/duplicate |
| `skill_executor` | Mock AIAssistant and handler func |
| Integration | app_test.go: `/` trigger → overlay → execution → message |
