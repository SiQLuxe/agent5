package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type SearXNGProvider struct {
	client  *http.Client
	baseURL string
}

func NewSearXNGProvider(baseURL string) *SearXNGProvider {
	return &SearXNGProvider{
		client:  &http.Client{},
		baseURL: baseURL,
	}
}

func (s *SearXNGProvider) Name() string {
	return "searxng"
}

type searxngResponse struct {
	Results []searxngResult `json:"results"`
}

type searxngResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
	Engine  string `json:"engine"`
}

func (s *SearXNGProvider) Search(ctx context.Context, query string, count int) ([]WebSearchResult, error) {
	u, err := url.Parse(s.baseURL + "/search")
	if err != nil {
		return nil, fmt.Errorf("searxng: parse URL: %w", err)
	}

	q := u.Query()
	q.Set("q", query)
	q.Set("format", "json")
	q.Set("language", "en-US")
	q.Set("categories", "general")
	if count > 0 {
		q.Set("limit", fmt.Sprintf("%d", count))
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("searxng: create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("searxng: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("searxng: unexpected status %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("searxng: read response: %w", err)
	}

	var searxngResp searxngResponse
	if err := json.Unmarshal(respBody, &searxngResp); err != nil {
		return nil, fmt.Errorf("searxng: parse response: %w", err)
	}

	results := make([]WebSearchResult, 0, len(searxngResp.Results))
	for _, r := range searxngResp.Results {
		results = append(results, WebSearchResult{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: r.Content,
			Content: r.Content,
		})
	}

	return results, nil
}
