package opencode

import (
	"context"
	"net/url"

	"github.com/example/agent-tui/internal/backend"
)

type opencodeCommandBody struct {
	Command   string `json:"command"`
	Arguments string `json:"arguments"`
}

type opencodeShellBody struct {
	Command string `json:"command"`
}

func (b *OpencodeBackend) ExecuteCommand(ctx context.Context, sessionID string, command string) (*backend.CommandResult, error) {
	body := opencodeCommandBody{Command: command}
	var resp struct {
		Parts []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"parts"`
	}
	if err := b.doRequest(ctx, "POST", "/session/"+sessionID+"/command", body, &resp); err != nil {
		return nil, err
	}

	result := &backend.CommandResult{}
	for _, p := range resp.Parts {
		result.Stdout += p.Text
	}
	return result, nil
}

func (b *OpencodeBackend) ExecuteShell(ctx context.Context, command string) (*backend.CommandResult, error) {
	body := opencodeShellBody{Command: command}
	var resp struct {
		Parts []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"parts"`
	}
	if err := b.doRequest(ctx, "POST", "/session/shell", body, &resp); err != nil {
		return nil, err
	}

	result := &backend.CommandResult{}
	for _, p := range resp.Parts {
		result.Stdout += p.Text
	}
	return result, nil
}

func (b *OpencodeBackend) ReadFile(ctx context.Context, path string) (string, error) {
	params := url.Values{}
	params.Set("path", path)

	var resp struct {
		Type    string `json:"type"`
		Content string `json:"content"`
	}
	if err := b.doRequest(ctx, "GET", "/file/content?"+params.Encode(), nil, &resp); err != nil {
		return "", err
	}
	return resp.Content, nil
}

func (b *OpencodeBackend) SearchText(ctx context.Context, pattern string) ([]backend.SearchResult, error) {
	params := url.Values{}
	params.Set("pattern", pattern)

	var matches []struct {
		Path       struct{ Text string } `json:"path"`
		Lines      struct{ Text string } `json:"lines"`
		LineNumber int                  `json:"line_number"`
	}
	if err := b.doRequest(ctx, "GET", "/find?"+params.Encode(), nil, &matches); err != nil {
		return nil, err
	}

	results := make([]backend.SearchResult, len(matches))
	for i, m := range matches {
		results[i] = backend.SearchResult{
			Path: m.Path.Text, LineNumber: m.LineNumber, Content: m.Lines.Text,
		}
	}
	return results, nil
}
