package runtime

import (
	"fmt"
	"time"

	"github.com/example/agent-tui/internal/agent/tool"
)

func (a *Agent) buildMessages(sessionID, task string) []Message {
	msgs := []Message{
		{Role: "system", Content: a.Config.SystemPrompt},
	}
	ctx := a.Session.GetContext(sessionID, a.Config.ContextLimit)
	for _, m := range ctx {
		msgs = append(msgs, Message{Role: m.Role, Content: m.Content})
	}
	msgs = append(msgs, Message{Role: "user", Content: task})
	return msgs
}

func (a *Agent) reactLoop(sessionID, task string) (string, error) {
	messages := a.buildMessages(sessionID, task)

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

			toolCallID := resp.ToolCall.Name + "-" + fmt.Sprintf("%d", i)
			messages = append(messages, Message{Role: "assistant", Content: ""})

			if result.Success {
				content := fmt.Sprintf("%v", result.Data)
				messages = append(messages, Message{Role: "tool", Content: content, ToolCallID: toolCallID})
			} else {
				messages = append(messages, Message{Role: "tool", Content: fmt.Sprintf("error: %s", result.Error), ToolCallID: toolCallID})
			}

		case "final":
			a.Session.AddMessage(sessionID, "user", task)
			a.Session.AddMessage(sessionID, "assistant", resp.Content)
			return resp.Content, nil
		}
	}

	return "", fmt.Errorf("max react loop iterations (%d) reached", a.Config.MaxReActLoop)
}

func (a *Agent) reactLoopStream(sessionID, task string, onChunk func(string)) (string, error) {
	messages := a.buildMessages(sessionID, task)

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

			toolCallID := resp.ToolCall.Name + "-" + fmt.Sprintf("%d", i)
			messages = append(messages, Message{Role: "assistant", Content: ""})

			if result.Success {
				content := fmt.Sprintf("%v", result.Data)
				messages = append(messages, Message{Role: "tool", Content: content, ToolCallID: toolCallID})
			} else {
				messages = append(messages, Message{Role: "tool", Content: fmt.Sprintf("error: %s", result.Error), ToolCallID: toolCallID})
			}

		case "final":
			a.Session.AddMessage(sessionID, "user", task)
			a.Session.AddMessage(sessionID, "assistant", resp.Content)
			return resp.Content, nil
		}
	}

	return "", fmt.Errorf("max react loop iterations (%d) reached", a.Config.MaxReActLoop)
}
