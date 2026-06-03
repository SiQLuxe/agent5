package opencode

import (
	"context"

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

	return b.sendMessageStreamViaSSE(ctx, sessionID, onChunk)
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
