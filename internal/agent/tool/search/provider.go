package search

import "context"

// WebSearchResult represents a single search result.
type WebSearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
	Content string `json:"content,omitempty"` // optional full content
}

// WebSearchProvider defines the interface for web search providers.
// Implementations must handle their own API keys, rate limiting, and errors.
type WebSearchProvider interface {
	// Search performs a web search and returns results.
	// query: the search query string
	// count: maximum number of results to return (0 = provider default)
	Search(ctx context.Context, query string, count int) ([]WebSearchResult, error)

	// Name returns the provider name (for logging/debugging).
	Name() string
}
