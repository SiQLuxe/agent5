package search

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type DuckDuckGoProvider struct {
	client  *http.Client
	baseURL string
}

func NewDuckDuckGoProvider() *DuckDuckGoProvider {
	return &DuckDuckGoProvider{
		client:  &http.Client{},
		baseURL: "https://lite.duckduckgo.com/lite/",
	}
}

func (d *DuckDuckGoProvider) Name() string {
	return "duckduckgo"
}

func (d *DuckDuckGoProvider) Search(ctx context.Context, query string, count int) ([]WebSearchResult, error) {
	form := url.Values{}
	form.Set("q", query)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.baseURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("duckduckgo: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; AgentTUI/1.0)")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("duckduckgo: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("duckduckgo: rate limited (status %d)", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("duckduckgo: unexpected status %d", resp.StatusCode)
	}

	results, err := parseDuckDuckGoResults(resp.Body, count)
	if err != nil {
		return nil, fmt.Errorf("duckduckgo: parse results: %w", err)
	}
	return results, nil
}

func parseDuckDuckGoResults(r io.Reader, maxResults int) ([]WebSearchResult, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var results []WebSearchResult
	var current *WebSearchResult

	var f func(*html.Node)
	f = func(n *html.Node) {
		if maxResults > 0 && len(results) >= maxResults {
			return
		}

		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "result-link") {
					current = &WebSearchResult{}
					for _, a := range n.Attr {
						if a.Key == "href" {
							current.URL = a.Val
							break
						}
					}
					if n.FirstChild != nil {
						current.Title = strings.TrimSpace(n.FirstChild.Data)
					}
				}
			}
		}

		if n.Type == html.ElementNode && n.Data == "td" && current != nil {
			for _, attr := range n.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "result-snippet") {
					if n.FirstChild != nil {
						current.Snippet = strings.TrimSpace(n.FirstChild.Data)
					}
					results = append(results, *current)
					current = nil
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	if current != nil && maxResults > 0 && len(results) < maxResults {
		results = append(results, *current)
	}

	return results, nil
}
