package runtime

import (
	"fmt"
	"time"

	"github.com/example/agent-tui/internal/agent/tool"
)

func (a *Agent) reactLoop(task string) (string, error) {
	messages := []Message{
		{Role: "system", Content: a.Config.SystemPrompt},
		{Role: "user", Content: task},
	}

	for i := 0; i < a.Config.MaxReActLoop; i++ {
		start := time.Now()
		resp, err := a.llm.ChatWithTools(messages, a.Tools.AsToolDefinitions(), a.Config.Model)
		if err != nil {
			return "", fmt.Errorf("llm call: %w", err)
		}
		a.Logger.Log(string(resp.Type), resp.Content, "", nil, nil, time.Since(start))

		switch resp.Type {
		case "tool_call":
			if resp.ToolCall == nil {
				continue
			}
			t, ok := a.Tools.Get(resp.ToolCall.Name)
			if !ok {
				errMsg := fmt.Sprintf("tool %q not found", resp.ToolCall.Name)
				messages = append(messages, Message{Role: "tool", Content: errMsg})
				a.Logger.Log("tool_error", errMsg, resp.ToolCall.Name, resp.ToolCall.Arguments, nil, 0)
				continue
			}

			toolStart := time.Now()
			tc := tool.ToolContext{
				Context:    nil,
				SandboxDir: a.Config.SandboxDir,
				Approval:   a.Config.ApprovalFn,
			}
			result := t.Execute(tc, resp.ToolCall.Arguments)
			a.Logger.Log("tool_result", fmt.Sprintf("%+v", result.Data), resp.ToolCall.Name, resp.ToolCall.Arguments, &result, time.Since(toolStart))

			if result.Success {
				content := fmt.Sprintf("%v", result.Data)
				messages = append(messages, Message{Role: "tool", Content: content})
			} else {
				messages = append(messages, Message{Role: "tool", Content: fmt.Sprintf("error: %s", result.Error)})
			}

		case "final":
			a.Memory.Append("assistant", resp.Content)
			return resp.Content, nil
		}
	}

	return "", fmt.Errorf("max react loop iterations (%d) reached", a.Config.MaxReActLoop)
}

func (a *Agent) reactLoopStream(task string, onChunk func(string)) (string, error) {
	messages := []Message{
		{Role: "system", Content: a.Config.SystemPrompt},
		{Role: "user", Content: task},
	}

	for i := 0; i < a.Config.MaxReActLoop; i++ {
		start := time.Now()
		resp, err := a.llm.ChatWithToolsStream(messages, a.Tools.AsToolDefinitions(), a.Config.Model, onChunk)
		if err != nil {
			return "", fmt.Errorf("llm call: %w", err)
		}
		a.Logger.Log(string(resp.Type), resp.Content, "", nil, nil, time.Since(start))

		switch resp.Type {
		case "tool_call":
			if resp.ToolCall == nil {
				continue
			}
			t, ok := a.Tools.Get(resp.ToolCall.Name)
			if !ok {
				errMsg := fmt.Sprintf("tool %q not found", resp.ToolCall.Name)
				messages = append(messages, Message{Role: "tool", Content: errMsg})
				a.Logger.Log("tool_error", errMsg, resp.ToolCall.Name, resp.ToolCall.Arguments, nil, 0)
				continue
			}

			toolStart := time.Now()
			tc := tool.ToolContext{
				Context:    nil,
				SandboxDir: a.Config.SandboxDir,
				Approval:   a.Config.ApprovalFn,
			}
			result := t.Execute(tc, resp.ToolCall.Arguments)
			a.Logger.Log("tool_result", fmt.Sprintf("%+v", result.Data), resp.ToolCall.Name, resp.ToolCall.Arguments, &result, time.Since(toolStart))

			if result.Success {
				content := fmt.Sprintf("%v", result.Data)
				messages = append(messages, Message{Role: "tool", Content: content})
			} else {
				messages = append(messages, Message{Role: "tool", Content: fmt.Sprintf("error: %s", result.Error)})
			}

		case "final":
			a.Memory.Append("assistant", resp.Content)
			return resp.Content, nil
		}
	}

	return "", fmt.Errorf("max react loop iterations (%d) reached", a.Config.MaxReActLoop)
}
