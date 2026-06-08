package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type TavilyProvider struct {
	client  *http.Client
	apiKey  string
	baseURL string
}

func NewTavilyProvider(apiKey string) *TavilyProvider {
	return &TavilyProvider{
		client:  &http.Client{},
		apiKey:  apiKey,
		baseURL: "https://api.tavily.com",
	}
}

func (t *TavilyProvider) Name() string {
	return "tavily"
}

type tavilyRequest struct {
	APIKey      string `json:"api_key"`
	Query       string `json:"query"`
	MaxResults  int    `json:"max_results"`
	SearchDepth string `json:"search_depth"`
}

type tavilyResponse struct {
	Results []tavilyResult `json:"results"`
}

type tavilyResult struct {
	Title   string  `json:"title"`
	URL     string  `json:"url"`
	Content string  `json:"content"`
	Score   float64 `json:"score"`
}

func (t *TavilyProvider) Search(ctx context.Context, query string, count int) ([]WebSearchResult, error) {
	if t.apiKey == "" {
		return nil, fmt.Errorf("tavily: API key not configured")
	}

	reqBody := tavilyRequest{
		APIKey:      t.apiKey,
		Query:       query,
		MaxResults:  count,
		SearchDepth: "basic",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("tavily: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		t.baseURL+"/search", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("tavily: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tavily: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("tavily: authentication failed (status %d)", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("tavily: rate limited (status %d)", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tavily: unexpected status %d", resp.StatusCode)
	}

	var tavilyResp tavilyResponse
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("tavily: read response: %w", err)
	}

	if err := json.Unmarshal(respBody, &tavilyResp); err != nil {
		return nil, fmt.Errorf("tavily: parse response: %w", err)
	}

	results := make([]WebSearchResult, 0, len(tavilyResp.Results))
	for _, r := range tavilyResp.Results {
		results = append(results, WebSearchResult{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: r.Content,
			Content: r.Content,
		})
	}

	return results, nil
}
