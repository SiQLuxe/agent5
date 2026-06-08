package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type BingProvider struct {
	client  *http.Client
	apiKey  string
	baseURL string
}

func NewBingProvider(apiKey string) *BingProvider {
	return &BingProvider{
		client:  &http.Client{},
		apiKey:  apiKey,
		baseURL: "https://api.bing.microsoft.com",
	}
}

func (b *BingProvider) Name() string {
	return "bing"
}

type bingResponse struct {
	WebPages *bingWebPages `json:"webPages"`
}

type bingWebPages struct {
	Value []bingResult `json:"value"`
}

type bingResult struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

func (b *BingProvider) Search(ctx context.Context, query string, count int) ([]WebSearchResult, error) {
	if b.apiKey == "" {
		return nil, fmt.Errorf("bing: API key not configured")
	}

	u, err := url.Parse(b.baseURL + "/v7.0/search")
	if err != nil {
		return nil, fmt.Errorf("bing: parse URL: %w", err)
	}

	q := u.Query()
	q.Set("q", query)
	q.Set("count", fmt.Sprintf("%d", count))
	q.Set("mkt", "en-US")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("bing: create request: %w", err)
	}
	req.Header.Set("Ocp-Apim-Subscription-Key", b.apiKey)

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bing: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("bing: authentication failed (status %d)", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("bing: rate limited (status %d)", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bing: unexpected status %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("bing: read response: %w", err)
	}

	var bingResp bingResponse
	if err := json.Unmarshal(respBody, &bingResp); err != nil {
		return nil, fmt.Errorf("bing: parse response: %w", err)
	}

	if bingResp.WebPages == nil {
		return []WebSearchResult{}, nil
	}

	results := make([]WebSearchResult, 0, len(bingResp.WebPages.Value))
	for _, r := range bingResp.WebPages.Value {
		results = append(results, WebSearchResult{
			Title:   r.Name,
			URL:     r.URL,
			Snippet: r.Snippet,
		})
	}

	return results, nil
}
