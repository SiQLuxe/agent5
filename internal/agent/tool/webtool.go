package tool

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/example/agent-tui/internal/agent/tool/search"
)

type WebSearchTool struct {
	provider search.WebSearchProvider
}

func NewWebSearchTool(provider search.WebSearchProvider) *WebSearchTool {
	return &WebSearchTool{provider: provider}
}

func (t *WebSearchTool) Name() string { return "web_search" }

func (t *WebSearchTool) Description() string {
	return "Search the web for information. Returns a list of results with title, URL, and snippet."
}

func (t *WebSearchTool) Schema() ToolSchema {
	return ToolSchema{
		Parameters: map[string]ParamSchema{
			"query": {
				Type:        "string",
				Description: "The search query",
			},
			"count": {
				Type:        "integer",
				Description: "Number of results to return (default 5)",
			},
		},
		Required: []string{"query"},
	}
}

func (t *WebSearchTool) Execute(ctx ToolContext, params map[string]interface{}) ToolResult {
	query, ok := params["query"].(string)
	if !ok || query == "" {
		return ToolResult{Error: "web_search: query is required"}
	}

	count := 5
	if c, ok := params["count"].(float64); ok {
		count = int(c)
	}

	results, err := t.provider.Search(ctx.Context, query, count)
	if err != nil {
		return ToolResult{Error: fmt.Sprintf("web_search: %s", err)}
	}

	return ToolResult{Success: true, Data: results}
}

// WebFetchTool fetches and extracts content from a URL.
type WebFetchTool struct{}

func NewWebFetchTool() *WebFetchTool {
	return &WebFetchTool{}
}

func (t *WebFetchTool) Name() string { return "web_fetch" }

func (t *WebFetchTool) Description() string {
	return "Fetch and extract the main content from a URL. Returns the page title and extracted text content."
}

func (t *WebFetchTool) Schema() ToolSchema {
	return ToolSchema{
		Parameters: map[string]ParamSchema{
			"url": {
				Type:        "string",
				Description: "The URL to fetch",
			},
			"selector": {
				Type:        "string",
				Description: "CSS selector to target specific content (optional, defaults to body)",
			},
		},
		Required: []string{"url"},
	}
}

// WebFetchResult represents the result of a web fetch operation.
type WebFetchResult struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	URL     string `json:"url"`
}

func (t *WebFetchTool) Execute(ctx ToolContext, params map[string]interface{}) ToolResult {
	urlStr, ok := params["url"].(string)
	if !ok || urlStr == "" {
		return ToolResult{Error: "web_fetch: url is required"}
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return ToolResult{Error: fmt.Sprintf("web_fetch: invalid URL: %s", urlStr)}
	}

	selector := "body"
	if s, ok := params["selector"].(string); ok && s != "" {
		selector = s
	}

	req, err := http.NewRequestWithContext(ctx.Context, http.MethodGet, urlStr, nil)
	if err != nil {
		return ToolResult{Error: fmt.Sprintf("web_fetch: create request: %s", err)}
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; AgentTUI/1.0)")
	req.Header.Set("Accept", "text/html")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ToolResult{Error: fmt.Sprintf("web_fetch: request failed: %s", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ToolResult{Error: fmt.Sprintf("web_fetch: HTTP %d", resp.StatusCode)}
	}

	limitedBody := io.LimitReader(resp.Body, 1<<20)

	doc, err := goquery.NewDocumentFromReader(limitedBody)
	if err != nil {
		return ToolResult{Error: fmt.Sprintf("web_fetch: parse HTML: %s", err)}
	}

	title := ""
	doc.Find("title").Each(func(i int, s *goquery.Selection) {
		title = strings.TrimSpace(s.Text())
	})

	var contentBuilder strings.Builder
	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		contentBuilder.WriteString(strings.TrimSpace(s.Text()))
		contentBuilder.WriteString("\n")
	})

	content := strings.TrimSpace(contentBuilder.String())

	return ToolResult{
		Success: true,
		Data: WebFetchResult{
			Title:   title,
			Content: content,
			URL:     urlStr,
		},
	}
}
