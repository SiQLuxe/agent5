package tool

import (
	"context"
	"fmt"
	"time"
)

type TaskTool struct {
	Manager *SubagentManager
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
		},
		Required: []string{"description", "prompt", "subagent_type"},
	}
}

func (t *TaskTool) Execute(ctx ToolContext, params map[string]interface{}) ToolResult {
	description, _ := params["description"].(string)
	prompt, _ := params["prompt"].(string)
	subagentType, _ := params["subagent_type"].(string)

	if description == "" || prompt == "" || subagentType == "" {
		return ToolResult{Error: "description, prompt, and subagent_type are required"}
	}

	cfg := SubAgentConfig{
		Name:         subagentType,
		SystemPrompt: buildSubagentPrompt(subagentType, description),
		Model:        subagentType,
	}

	sa := t.Manager.Spawn(context.Background(), cfg)
	if sa.Status == StatusFailed {
		return ToolResult{
			Error: fmt.Sprintf("failed to spawn subagent: %s", sa.Error),
		}
	}

	result := fmt.Sprintf("[subagent:%s] task %q started at %s (agent: %s)",
		sa.ID, description, sa.CreatedAt.Format(time.RFC3339), subagentType)

	return ToolResult{
		Success: true,
		Data:    result,
	}
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
