package bridge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/example/agent-tui/internal/agent/tool"
)

type OpencodeBridge struct {
	baseURL   string
	apiKey    string
	client    *http.Client
	sessionID string
}

func NewOpencodeBridge(baseURL, apiKey string) *OpencodeBridge {
	return &OpencodeBridge{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (b *OpencodeBridge) Name() string { return "delegate_opencode" }

func (b *OpencodeBridge) Description() string {
	return "Delegate a task to opencode (external AI coding agent)"
}

func (b *OpencodeBridge) Schema() tool.ToolSchema {
	return tool.ToolSchema{
		Parameters: map[string]tool.ParamSchema{
			"task": {Type: "string", Description: "Task description for opencode"},
		},
		Required: []string{"task"},
	}
}

func (b *OpencodeBridge) Execute(ctx tool.ToolContext, params map[string]interface{}) tool.ToolResult {
	task, _ := params["task"].(string)
	if task == "" {
		return tool.ToolResult{Error: "task parameter is required"}
	}

	// Create session if needed
	if b.sessionID == "" {
		sessionID, err := b.createSession(ctx.Context)
		if err != nil {
			return tool.ToolResult{Error: fmt.Sprintf("create session: %s", err)}
		}
		b.sessionID = sessionID
	}

	// Send task as message
	result, err := b.sendMessage(ctx.Context, task)
	if err != nil {
		return tool.ToolResult{Error: fmt.Sprintf("send message: %s", err)}
	}

	return tool.ToolResult{Success: true, Data: result}
}

func (b *OpencodeBridge) createSession(ctx context.Context) (string, error) {
	body := map[string]string{"title": "agent-delegated-task"}
	data, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", b.baseURL+"/session", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.ID, nil
}

func (b *OpencodeBridge) sendMessage(ctx context.Context, task string) (string, error) {
	body := map[string]interface{}{
		"parts": []map[string]string{
			{"type": "text", "text": task},
		},
	}
	data, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", b.baseURL+"/session/"+b.sessionID+"/message", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var msg struct {
		Parts []map[string]string `json:"parts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		return "", fmt.Errorf("decode response: %s", err)
	}

	for _, part := range msg.Parts {
		if part["type"] == "text" {
			return part["text"], nil
		}
	}
	return "task delegated to opencode", nil
}
