package opencode

import (
	"context"
	"fmt"

	"github.com/example/agent-tui/internal/backend"
)

type opencodePart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type opencodePromptBody struct {
	Parts []opencodePart `json:"parts"`
}

type opencodeMessageResponse struct {
	Info struct {
		ID string `json:"id"`
	} `json:"info"`
	Parts []opencodePart `json:"parts"`
}

func (b *OpencodeBackend) SendMessage(ctx context.Context, sessionID string, msg *backend.Message) (*backend.MessageResult, error) {
	body := opencodePromptBody{
		Parts: []opencodePart{
			{Type: "text", Text: msg.Content},
		},
	}
	var resp opencodeMessageResponse
	if err := b.doRequest(ctx, "POST", "/session/"+sessionID+"/message", body, &resp); err != nil {
		return nil, err
	}

	result := &backend.MessageResult{
		SessionID: sessionID,
		MessageID: resp.Info.ID,
	}
	for _, p := range resp.Parts {
		if p.Type == "text" {
			result.Content += p.Text
		}
	}
	return result, nil
}

func (b *OpencodeBackend) SendMessageStream(ctx context.Context, sessionID string, msg *backend.Message, onChunk func(*backend.Chunk)) error {
	body := opencodePromptBody{
		Parts: []opencodePart{
			{Type: "text", Text: msg.Content},
		},
	}
	if err := b.doRequest(ctx, "POST", "/session/"+sessionID+"/prompt_async", body, nil); err != nil {
		return err
	}

	events, err := b.Events(ctx)
	if err != nil {
		return err
	}

	for event := range events {
		if event.Type == "message.completed" {
			payload, ok := event.Payload.(map[string]interface{})
			if ok {
				if content, exists := payload["content"]; exists {
					onChunk(&backend.Chunk{Content: fmt.Sprintf("%v", content), Done: false})
				}
			}
			onChunk(&backend.Chunk{Done: true})
			return nil
		}
		if event.Type == "message.chunk" {
			payload, ok := event.Payload.(map[string]interface{})
			if ok {
				if content, exists := payload["content"]; exists {
					onChunk(&backend.Chunk{Content: fmt.Sprintf("%v", content), Done: false})
				}
			}
		}
	}
	return nil
}

func (b *OpencodeBackend) GetMessages(ctx context.Context, sessionID string) ([]*backend.Message, error) {
	var messages []struct {
		Info struct {
			ID   string `json:"id"`
			Role string `json:"role"`
		} `json:"info"`
		Parts []opencodePart `json:"parts"`
	}

	if err := b.doRequest(ctx, "GET", "/session/"+sessionID+"/message", nil, &messages); err != nil {
		return nil, err
	}

	result := make([]*backend.Message, len(messages))
	for i, m := range messages {
		content := ""
		for _, p := range m.Parts {
			if p.Type == "text" {
				content += p.Text
			}
		}
		result[i] = &backend.Message{
			Role:    backend.MessageRole(m.Info.Role),
			Content: content,
		}
	}
	return result, nil
}
