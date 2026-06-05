package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type AnthropicClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func NewAnthropicClient(apiKey, baseURL, model string) (Client, error) {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}
	if model == "" {
		model = "claude-3-opus"
	}
	return &AnthropicClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}, nil
}

type anthropicContent struct {
	Type  string `json:"type"`
	Text  string `json:"text,omitempty"`
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Input any    `json:"input,omitempty"`
}

type anthropicMessage struct {
	Role    string             `json:"role"`
	Content []anthropicContent `json:"content"`
}

type anthropicToolDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"input_schema"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
	Tools     []anthropicToolDef `json:"tools,omitempty"`
	Stream    bool               `json:"stream,omitempty"`
}

type anthropicResponse struct {
	ID         string             `json:"id"`
	Type       string             `json:"type"`
	Role       string             `json:"role"`
	Content    []anthropicContent `json:"content"`
	Model      string             `json:"model"`
	StopReason string             `json:"stop_reason"`
}

func (c *AnthropicClient) ChatCompletion(req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if req.Model == "" {
		req.Model = c.model
	}

	// Extract system prompt and convert messages
	var systemPrompt string
	anthropicMessages := make([]anthropicMessage, 0, len(req.Messages))
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
			continue
		}
		anthMsg := anthropicMessage{
			Role:    msg.Role,
			Content: []anthropicContent{{Type: "text", Text: msg.Content}},
		}
		anthropicMessages = append(anthropicMessages, anthMsg)
	}

	// Convert tools
	anthropicTools := make([]anthropicToolDef, len(req.Tools))
	for i, t := range req.Tools {
		anthropicTools[i] = anthropicToolDef{
			Name:        t.Function.Name,
			Description: t.Function.Description,
			InputSchema: t.Function.Parameters,
		}
	}

	anthroReq := anthropicRequest{
		Model:     req.Model,
		MaxTokens: 4096,
		System:    systemPrompt,
		Messages:  anthropicMessages,
		Tools:     anthropicTools,
	}

	data, err := json.Marshal(anthroReq)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/messages", c.baseURL)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-api-key", c.apiKey)
	request.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed: %s", string(body))
	}

	var anthroResp anthropicResponse
	err = json.NewDecoder(resp.Body).Decode(&anthroResp)
	if err != nil {
		return nil, err
	}

	// Convert back to ChatCompletionResponse
	result := ChatCompletionResponse{
		ID:     anthroResp.ID,
		Object: "chat.completion",
		Model:  anthroResp.Model,
	}

	var toolCalls []ToolCall
	var textContent string
	for _, block := range anthroResp.Content {
		switch block.Type {
		case "text":
			textContent += block.Text
		case "tool_use":
			inputJSON, _ := json.Marshal(block.Input)
			toolCalls = append(toolCalls, ToolCall{
				ID:   block.ID,
				Type: "function",
				Function: ToolCallFunction{
					Name:      block.Name,
					Arguments: string(inputJSON),
				},
			})
		}
	}

	finishReason := anthroResp.StopReason
	if finishReason == "end_turn" {
		finishReason = "stop"
	} else if finishReason == "tool_use" {
		finishReason = "tool_calls"
	}

	result.Choices = []ResponseChoice{
		{
			Message: ResponseMessage{
				Role:      anthroResp.Role,
				Content:   textContent,
				ToolCalls: toolCalls,
			},
			FinishReason: finishReason,
		},
	}

	return &result, nil
}

func (c *AnthropicClient) ChatCompletionStream(req ChatCompletionRequest, callback func(string)) error {
	if req.Model == "" {
		req.Model = c.model
	}

	// Extract system prompt and convert messages
	var systemPrompt string
	anthropicMessages := make([]anthropicMessage, 0, len(req.Messages))
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
			continue
		}
		anthMsg := anthropicMessage{
			Role:    msg.Role,
			Content: []anthropicContent{{Type: "text", Text: msg.Content}},
		}
		anthropicMessages = append(anthropicMessages, anthMsg)
	}

	anthroReq := anthropicRequest{
		Model:     req.Model,
		MaxTokens: 4096,
		System:    systemPrompt,
		Messages:  anthropicMessages,
		Stream:    true,
	}

	data, err := json.Marshal(anthroReq)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/messages", c.baseURL)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-api-key", c.apiKey)
	request.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed: %s", string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	var currentText string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "event: ") {
			continue
		}
		if strings.HasPrefix(line, "data: ") {
			data := line[6:]
			if data == "[DONE]" {
				break
			}
			var event struct {
				Type  string `json:"type"`
				Delta *struct {
					Text string `json:"text"`
				} `json:"delta,omitempty"`
				ContentBlock *struct {
					Text string `json:"text"`
					Type string `json:"type"`
				} `json:"content_block,omitempty"`
			}
			if err := json.Unmarshal([]byte(data), &event); err == nil {
				switch event.Type {
				case "content_block_delta":
					if event.Delta != nil && event.Delta.Text != "" {
						currentText += event.Delta.Text
						callback(event.Delta.Text)
					}
				case "content_block_start":
					if event.ContentBlock != nil && event.ContentBlock.Text != "" {
						currentText += event.ContentBlock.Text
						callback(event.ContentBlock.Text)
					}
				}
			}
		}
	}

	return scanner.Err()
}

func (c *AnthropicClient) SetAPIKey(key string) {
	c.apiKey = key
}

func (c *AnthropicClient) SetBaseURL(url string) {
	c.baseURL = url
}

func (c *AnthropicClient) GetModel() string {
	return c.model
}

func (c *AnthropicClient) ListModels() ([]string, error) {
	return []string{
		"claude-3-opus",
		"claude-3-sonnet",
		"claude-3-haiku",
		"claude-2.1",
	}, nil
}
