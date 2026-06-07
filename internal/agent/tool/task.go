package tool

import (
	"context"
	"fmt"
)

type SubagentRunner interface {
	Run(sessionID, task, systemPrompt string, tools *Registry) (string, error)
}

type TaskTool struct {
	Manager *SubagentManager
	Runner  SubagentRunner
	Session interface {
		CreateChildSession(parentID, name string) string
		AddMessage(sessionID, role, content string) error
	}
	Tools *Registry
	Depth int
}

func (t *TaskTool) Name() string { return "task" }

func (t *TaskTool) Description() string {
	return "Delegate a task to a specialized subagent. The subagent runs independently with its own context."
}

func (t *TaskTool) Schema() ToolSchema {
	return ToolSchema{
		Parameters: map[string]ParamSchema{
			"description": {
				Type:        "string",
				Description: "A short (3-5 word) description of the task",
			},
			"prompt": {
				Type:        "string",
				Description: "The detailed task for the subagent to perform",
			},
			"subagent_type": {
				Type:        "string",
				Description: "Type of subagent: general, explore, review",
			},
			"background": {
				Type:        "bool",
				Description: "Run in background (return agent_id immediately)",
			},
		},
		Required: []string{"description", "prompt", "subagent_type"},
	}
}

func (t *TaskTool) Execute(ctx ToolContext, params map[string]interface{}) ToolResult {
	description, _ := params["description"].(string)
	prompt, _ := params["prompt"].(string)
	subagentType, _ := params["subagent_type"].(string)
	background, _ := params["background"].(bool)

	if description == "" || prompt == "" || subagentType == "" {
		return ToolResult{Error: "description, prompt, and subagent_type are required"}
	}

	if t.Depth <= 0 {
		return ToolResult{Error: "max subagent nesting depth reached"}
	}
	if t.Runner == nil {
		return ToolResult{Error: "SubagentRunner not configured"}
	}

	subTools := filterTools(t.Tools, subagentType)

	cfg := SubAgentConfig{
		Name:         subagentType,
		SystemPrompt: buildSubagentPrompt(subagentType, description),
		Model:        subagentType,
		MaxDepth:     t.Depth - 1,
	}

	sa := t.Manager.Spawn(context.Background(), cfg)
	if sa.Status == StatusFailed {
		return ToolResult{
			Error: fmt.Sprintf("failed to spawn subagent: %s", sa.Error),
		}
	}

	sessionID := sa.ID
	if t.Session != nil {
		parentID := ""
		if p, ok := params["session_id"].(string); ok {
			parentID = p
		}
		sessionID = t.Session.CreateChildSession(parentID, "subagent:"+subagentType)
	}

	if background {
		go func() {
			result, err := t.Runner.Run(sessionID, prompt, cfg.SystemPrompt, subTools)
			if err != nil {
				t.Manager.Fail(sa.ID, err.Error())
				if t.Session != nil {
					t.Session.AddMessage(sessionID, "system",
						fmt.Sprintf("[subagent:%s] failed: %s", sa.ID, err.Error()))
				}
				return
			}
			t.Manager.Complete(sa.ID, result)
		}()

		return ToolResult{
			Success: true,
			Data:    fmt.Sprintf("[subagent:%s] task %q started in background (session: %s)", sa.ID, description, sessionID),
		}
	}

	result, err := t.Runner.Run(sessionID, prompt, cfg.SystemPrompt, subTools)
	if err != nil {
		t.Manager.Fail(sa.ID, err.Error())
		return ToolResult{Error: err.Error()}
	}
	t.Manager.Complete(sa.ID, result)

	return ToolResult{
		Success: true,
		Data:    result,
	}
}

func filterTools(reg *Registry, agentType string) *Registry {
	out := NewRegistry()
	if reg == nil {
		return out
	}
	switch agentType {
	case "explore":
		for _, name := range []string{"read_file", "search_text"} {
			if t, ok := reg.Get(name); ok {
				out.Register(t)
			}
		}
	case "review":
		for _, name := range []string{"read_file", "search_text", "exec_command"} {
			if t, ok := reg.Get(name); ok {
				out.Register(t)
			}
		}
	default:
		for _, t := range reg.List() {
			out.Register(t)
		}
	}
	return out
}

func buildSubagentPrompt(agentType, description string) string {
	switch agentType {
	case "explore":
		return fmt.Sprintf("You are a read-only code exploration agent. Task: %s. Search and analyze code without making changes.", description)
	case "review":
		return fmt.Sprintf("You are a code review agent. Task: %s. Review code for bugs, security issues, and best practices.", description)
	default:
		return fmt.Sprintf("You are a general-purpose agent. Task: %s.", description)
	}
}
