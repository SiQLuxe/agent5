# Web Search & Fetch + Skill UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `web_search` and `web_fetch` tools to the agent ReAct loop, and build the Skill UI overlay for browsing/executing skills via `/command` or `Ctrl+P`.

**Architecture:** Web search uses a `WebSearchProvider` interface with `ChainProvider` trying DuckDuckGo (always free) then optional Tavily/Bing/SearXNG. Web fetch is a standalone HTML scraper. Skill overlay is a tview `Flex` inserted above the composer, with keyboard navigation and real-time filtering.

**Tech Stack:** Go 1.26, tview (TUI), goquery (HTML parsing), net/http, existing agent tool framework.

---

### Task 1: Add goquery dependency + config types

**Files:**
- Modify: `go.mod`, `go.sum`
- Modify: `internal/data/config/config.go`
- Modify: `internal/data/config/config_test.go`
- Test: `internal/data/config/config_test.go`

- [ ] **Step 1: Add goquery dependency**

Run: `go get github.com/PuerkitoBio/goquery@latest`

- [ ] **Step 2: Add WebSearchConfig to Config struct**

Add to `internal/data/config/config.go`:

```go
type WebSearchConfig struct {
    MaxResults     int    `toml:"max_results"`
    RequestTimeout int    `toml:"request_timeout"`
    TavilyAPIKey   string `toml:"tavily_api_key"`
    BingAPIKey     string `toml:"bing_api_key"`
    SearXNGUrl     string `toml:"searxng_url"`
}
```

Add field to `Config` struct:

```go
type Config struct {
    // ... existing fields ...
    WebSearch WebSearchConfig `toml:"web_search"`
}
```

Add defaults in `GetDefaultConfig()`:

```go
WebSearch: WebSearchConfig{
    MaxResults:     5,
    RequestTimeout: 10,
},
```

- [ ] **Step 3: Write config test**

In `internal/data/config/config_test.go`:

```go
func TestWebSearchConfig(t *testing.T) {
    tomlData := `
[web_search]
max_results = 3
request_timeout = 15
tavily_api_key = "test-key"
bing_api_key = "bing-key"
searxng_url = "http://searx:8888"
`
    var cfg Config
    if _, err := toml.Decode(tomlData, &cfg); err != nil {
        t.Fatalf("decode: %v", err)
    }
    if cfg.WebSearch.MaxResults != 3 {
        t.Fatalf("expected 3, got %d", cfg.WebSearch.MaxResults)
    }
    if cfg.WebSearch.TavilyAPIKey != "test-key" {
        t.Fatalf("expected test-key, got %q", cfg.WebSearch.TavilyAPIKey)
    }
    if cfg.WebSearch.SearXNGUrl != "http://searx:8888" {
        t.Fatalf("expected http://searx:8888, got %q", cfg.WebSearch.SearXNGUrl)
    }
}
```

- [ ] **Step 4: Verify tests pass**

Run: `go test ./internal/data/config/... -v -run TestWebSearchConfig`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum internal/data/config/config.go internal/data/config/config_test.go
git commit -m "feat: add goquery dep and WebSearchConfig"
```

---

### Task 2: Provider interface + types + ChainProvider

**Files:**
- Create: `internal/agent/tool/search/provider.go`
- Create: `internal/agent/tool/search/chain.go`
- Create: `internal/agent/tool/search/chain_test.go`

- [ ] **Step 1: Write failing test for ChainProvider**

```go
// chain_test.go
package search

import (
    "context"
    "errors"
    "testing"
)

type mockProvider struct {
    name     string
    avail    bool
    results  []SearchResult
    failWith error
}

func (m *mockProvider) Name() string { return m.name }
func (m *mockProvider) IsAvailable() bool { return m.avail }
func (m *mockProvider) Search(ctx context.Context, q string, max int) (*SearchResponse, error) {
    if m.failWith != nil {
        return nil, m.failWith
    }
    return &SearchResponse{Results: m.results, Source: m.name}, nil
}

func TestChainProvider_SkipsUnavailable(t *testing.T) {
    chain := NewChainProvider(
        &mockProvider{name: "a", avail: false},
        &mockProvider{name: "b", avail: true, results: []SearchResult{{Title: "ok"}}},
    )
    resp, err := chain.Search(context.Background(), "test", 5)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if resp.Source != "b" {
        t.Fatalf("expected source 'b', got %q", resp.Source)
    }
}

func TestChainProvider_FallbackOnError(t *testing.T) {
    chain := NewChainProvider(
        &mockProvider{name: "a", avail: true, failWith: errors.New("timeout")},
        &mockProvider{name: "b", avail: true, results: []SearchResult{{Title: "ok"}}},
    )
    resp, err := chain.Search(context.Background(), "test", 5)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if resp.Source != "b" {
        t.Fatalf("expected source 'b', got %q", resp.Source)
    }
}

func TestChainProvider_AllFail(t *testing.T) {
    chain := NewChainProvider(
        &mockProvider{name: "a", avail: true, failWith: errors.New("err1")},
        &mockProvider{name: "b", avail: true, failWith: errors.New("err2")},
    )
    _, err := chain.Search(context.Background(), "test", 5)
    if err == nil {
        t.Fatal("expected error, got nil")
    }
}

func TestChainProvider_NoAvailable(t *testing.T) {
    chain := NewChainProvider(
        &mockProvider{name: "a", avail: false},
    )
    _, err := chain.Search(context.Background(), "test", 5)
    if err == nil {
        t.Fatal("expected error when no provider available")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/agent/tool/search/... -v`
Expected: FAIL — package doesn't exist yet

- [ ] **Step 3: Write minimal implementation**

`internal/agent/tool/search/provider.go`:

```go
package search

import "context"

type SearchResult struct {
    Title   string
    URL     string
    Content string
    Source  string
}

type SearchResponse struct {
    Results []SearchResult
    Source  string
}

type WebSearchProvider interface {
    Name() string
    IsAvailable() bool
    Search(ctx context.Context, query string, maxResults int) (*SearchResponse, error)
}
```

`internal/agent/tool/search/chain.go`:

```go
package search

import (
    "context"
    "fmt"
    "strings"
)

type ChainProvider struct {
    providers []WebSearchProvider
}

func NewChainProvider(providers ...WebSearchProvider) *ChainProvider {
    return &ChainProvider{providers: providers}
}

func (c *ChainProvider) Search(ctx context.Context, query string, maxResults int) (*SearchResponse, error) {
    var errs []string
    for _, p := range c.providers {
        if !p.IsAvailable() {
            continue
        }
        resp, err := p.Search(ctx, query, maxResults)
        if err == nil {
            return resp, nil
        }
        errs = append(errs, fmt.Sprintf("%s: %v", p.Name(), err))
    }
    if len(errs) == 0 {
        return nil, fmt.Errorf("no web search provider available")
    }
    return nil, fmt.Errorf("all providers failed: %s", strings.Join(errs, "; "))
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/agent/tool/search/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/agent/tool/search/
git commit -m "feat: add WebSearchProvider interface and ChainProvider"
```

---

### Task 3: DuckDuckGo provider

**Files:**
- Create: `internal/agent/tool/search/duckduckgo.go`
- Create: `internal/agent/tool/search/duckduckgo_test.go`

- [ ] **Step 1: Write failing test**

```go
// duckduckgo_test.go
package search

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestDuckDuckGoProvider_Search(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Query().Get("q") != "hello" {
            t.Fatalf("unexpected query: %q", r.URL.Query().Get("q"))
        }
        w.Write([]byte(`
<html><body>
<div class="result">
    <a class="result__title" href="https://example.com">Example</a>
    <div class="result__snippet">Hello world snippet</div>
</div>
</body></html>
`))
    }))
    defer server.Close()

    p := &DuckDuckGoProvider{baseURL: server.URL + "/"}
    resp, err := p.Search(context.Background(), "hello", 5)
    if err != nil {
        t.Fatalf("search failed: %v", err)
    }
    if len(resp.Results) == 0 {
        t.Fatal("expected at least 1 result")
    }
    if resp.Results[0].Title != "Example" {
        t.Fatalf("expected title 'Example', got %q", resp.Results[0].Title)
    }
    if resp.Results[0].URL != "https://example.com" {
        t.Fatalf("expected URL 'https://example.com', got %q", resp.Results[0].URL)
    }
    if resp.Results[0].Content != "Hello world snippet" {
        t.Fatalf("expected content 'Hello world snippet', got %q", resp.Results[0].Content)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/agent/tool/search/... -v -run TestDuckDuckGoProvider`
Expected: FAIL with undefined `DuckDuckGoProvider`

- [ ] **Step 3: Write DuckDuckGo provider**

```go
// duckduckgo.go
package search

import (
    "context"
    "fmt"
    "net/http"
    "net/url"
    "strings"
    "time"

    "github.com/PuerkitoBio/goquery"
)

type DuckDuckGoProvider struct {
    baseURL string
    client  *http.Client
}

func NewDuckDuckGoProvider() *DuckDuckGoProvider {
    return &DuckDuckGoProvider{
        baseURL: "https://html.duckduckgo.com/html/",
        client:  &http.Client{Timeout: 10 * time.Second},
    }
}

func (d *DuckDuckGoProvider) Name() string { return "duckduckgo" }

func (d *DuckDuckGoProvider) IsAvailable() bool { return true }

func (d *DuckDuckGoProvider) Search(ctx context.Context, query string, maxResults int) (*SearchResponse, error) {
    req, err := http.NewRequestWithContext(ctx, "POST", d.baseURL, strings.NewReader(url.Values{"q": {query}}.Encode()))
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; AgentTUI/1.0)")

    resp, err := d.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("parse HTML: %w", err)
    }

    var results []SearchResult
    doc.Find(".result").Each(func(i int, s *goquery.Selection) {
        if maxResults > 0 && len(results) >= maxResults {
            return
        }
        titleSel := s.Find(".result__title a")
        title := strings.TrimSpace(titleSel.Text())
        href, _ := titleSel.Attr("href")
        snippet := strings.TrimSpace(s.Find(".result__snippet").Text())
        if title != "" {
            results = append(results, SearchResult{
                Title:   title,
                URL:     href,
                Content: snippet,
                Source:  "duckduckgo",
            })
        }
    })

    if len(results) == 0 {
        return nil, fmt.Errorf("no results found")
    }
    return &SearchResponse{Results: results, Source: "duckduckgo"}, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/agent/tool/search/... -v -run TestDuckDuckGoProvider`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/agent/tool/search/duckduckgo.go internal/agent/tool/search/duckduckgo_test.go
git commit -m "feat: add DuckDuckGo search provider"
```

---

### Task 4: Tavily provider

**Files:**
- Create: `internal/agent/tool/search/tavily.go`
- Create: `internal/agent/tool/search/tavily_test.go`

- [ ] **Step 1: Write failing test**

```go
// tavily_test.go
package search

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestTavilyProvider_Search(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var body map[string]interface{}
        json.NewDecoder(r.Body).Decode(&body)
        if body["api_key"] != "test-key" {
            t.Fatalf("unexpected api_key: %v", body["api_key"])
        }
        json.NewEncoder(w).Encode(map[string]interface{}{
            "results": []map[string]interface{}{
                {"title": "Hello Page", "url": "https://example.com", "content": "Hello content"},
            },
        })
    }))
    defer server.Close()

    p := NewTavilyProvider("test-key")
    p.baseURL = server.URL + "/"
    p.client = server.Client()

    resp, err := p.Search(context.Background(), "hello", 5)
    if err != nil {
        t.Fatalf("search failed: %v", err)
    }
    if len(resp.Results) != 1 {
        t.Fatalf("expected 1 result, got %d", len(resp.Results))
    }
    if resp.Results[0].Title != "Hello Page" {
        t.Fatalf("expected 'Hello Page', got %q", resp.Results[0].Title)
    }
}

func TestTavilyProvider_NotAvailableWithoutKey(t *testing.T) {
    p := NewTavilyProvider("")
    if p.IsAvailable() {
        t.Fatal("expected not available without key")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/agent/tool/search/... -v -run TestTavilyProvider`
Expected: FAIL with undefined `NewTavilyProvider`

- [ ] **Step 3: Write Tavily provider**

```go
// tavily.go
package search

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type TavilyProvider struct {
    apiKey  string
    baseURL string
    client  *http.Client
}

func NewTavilyProvider(apiKey string) *TavilyProvider {
    return &TavilyProvider{
        apiKey:  apiKey,
        baseURL: "https://api.tavily.com/search",
        client:  &http.Client{Timeout: 10 * time.Second},
    }
}

func (t *TavilyProvider) Name() string { return "tavily" }
func (t *TavilyProvider) IsAvailable() bool { return t.apiKey != "" }

func (t *TavilyProvider) Search(ctx context.Context, query string, maxResults int) (*SearchResponse, error) {
    body := map[string]interface{}{
        "api_key":        t.apiKey,
        "query":          query,
        "search_depth":   "basic",
        "max_results":    maxResults,
        "include_answer": false,
    }
    data, _ := json.Marshal(body)

    req, err := http.NewRequestWithContext(ctx, "POST", t.baseURL, bytes.NewReader(data))
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := t.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    var result struct {
        Results []struct {
            Title   string `json:"title"`
            URL     string `json:"url"`
            Content string `json:"content"`
        } `json:"results"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("parse response: %w", err)
    }

    var results []SearchResult
    for _, r := range result.Results {
        results = append(results, SearchResult{
            Title:   r.Title,
            URL:     r.URL,
            Content: r.Content,
            Source:  "tavily",
        })
    }
    return &SearchResponse{Results: results, Source: "tavily"}, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/agent/tool/search/... -v -run TestTavilyProvider`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/agent/tool/search/tavily.go internal/agent/tool/search/tavily_test.go
git commit -m "feat: add Tavily search provider"
```

---

### Task 5: Bing provider

**Files:**
- Create: `internal/agent/tool/search/bing.go`
- Create: `internal/agent/tool/search/bing_test.go`

- [ ] **Step 1: Write failing test**

```go
// bing_test.go
package search

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestBingProvider_Search(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Header.Get("Ocp-Apim-Subscription-Key") != "test-key" {
            t.Fatalf("missing API key header")
        }
        if r.URL.Query().Get("q") != "hello" {
            t.Fatalf("unexpected query: %q", r.URL.Query().Get("q"))
        }
        json.NewEncoder(w).Encode(map[string]interface{}{
            "webPages": map[string]interface{}{
                "value": []map[string]interface{}{
                    {"name": "Hello", "url": "https://example.com", "snippet": "Hello snippet"},
                },
            },
        })
    }))
    defer server.Close()

    p := NewBingProvider("test-key")
    p.baseURL = server.URL + "/"
    p.client = server.Client()

    resp, err := p.Search(context.Background(), "hello", 5)
    if err != nil {
        t.Fatalf("search failed: %v", err)
    }
    if len(resp.Results) != 1 {
        t.Fatalf("expected 1 result, got %d", len(resp.Results))
    }
}

func TestBingProvider_NotAvailableWithoutKey(t *testing.T) {
    p := NewBingProvider("")
    if p.IsAvailable() {
        t.Fatal("expected not available without key")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/agent/tool/search/... -v -run TestBingProvider`
Expected: FAIL with undefined `NewBingProvider`

- [ ] **Step 3: Write Bing provider**

```go
// bing.go
package search

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type BingProvider struct {
    apiKey  string
    baseURL string
    client  *http.Client
}

func NewBingProvider(apiKey string) *BingProvider {
    return &BingProvider{
        apiKey:  apiKey,
        baseURL: "https://api.bing.microsoft.com/v7.0/search",
        client:  &http.Client{Timeout: 10 * time.Second},
    }
}

func (b *BingProvider) Name() string { return "bing" }
func (b *BingProvider) IsAvailable() bool { return b.apiKey != "" }

func (b *BingProvider) Search(ctx context.Context, query string, maxResults int) (*SearchResponse, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", b.baseURL, nil)
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }
    q := req.URL.Query()
    q.Add("q", query)
    q.Add("count", fmt.Sprintf("%d", maxResults))
    req.URL.RawQuery = q.Encode()
    req.Header.Set("Ocp-Apim-Subscription-Key", b.apiKey)

    resp, err := b.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    var result struct {
        WebPages struct {
            Value []struct {
                Name    string `json:"name"`
                URL     string `json:"url"`
                Snippet string `json:"snippet"`
            } `json:"value"`
        } `json:"webPages"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("parse response: %w", err)
    }

    var results []SearchResult
    for _, r := range result.WebPages.Value {
        results = append(results, SearchResult{
            Title:   r.Name,
            URL:     r.URL,
            Content: r.Snippet,
            Source:  "bing",
        })
    }
    return &SearchResponse{Results: results, Source: "bing"}, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/agent/tool/search/... -v -run TestBingProvider`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/agent/tool/search/bing.go internal/agent/tool/search/bing_test.go
git commit -m "feat: add Bing search provider"
```

---

### Task 6: SearXNG provider

**Files:**
- Create: `internal/agent/tool/search/searxng.go`
- Create: `internal/agent/tool/search/searxng_test.go`

- [ ] **Step 1: Write failing test**

```go
// searxng_test.go
package search

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestSearXNGProvider_Search(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Query().Get("q") != "hello" {
            t.Fatalf("unexpected query: %q", r.URL.Query().Get("q"))
        }
        json.NewEncoder(w).Encode(map[string]interface{}{
            "results": []map[string]interface{}{
                {"title": "Hello", "url": "https://example.com", "content": "Hello content"},
            },
        })
    }))
    defer server.Close()

    p := NewSearXNGProvider(server.URL + "/search")
    p.client = server.Client()

    resp, err := p.Search(context.Background(), "hello", 5)
    if err != nil {
        t.Fatalf("search failed: %v", err)
    }
    if len(resp.Results) != 1 {
        t.Fatalf("expected 1 result, got %d", len(resp.Results))
    }
}

func TestSearXNGProvider_NotAvailableWithoutURL(t *testing.T) {
    p := NewSearXNGProvider("")
    if p.IsAvailable() {
        t.Fatal("expected not available without URL")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/agent/tool/search/... -v -run TestSearXNGProvider`
Expected: FAIL with undefined `NewSearXNGProvider`

- [ ] **Step 3: Write SearXNG provider**

```go
// searxng.go
package search

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type SearXNGProvider struct {
    baseURL string
    client  *http.Client
}

func NewSearXNGProvider(searxngURL string) *SearXNGProvider {
    return &SearXNGProvider{
        baseURL: searxngURL,
        client:  &http.Client{Timeout: 10 * time.Second},
    }
}

func (s *SearXNGProvider) Name() string { return "searxng" }
func (s *SearXNGProvider) IsAvailable() bool { return s.baseURL != "" }

func (s *SearXNGProvider) Search(ctx context.Context, query string, maxResults int) (*SearchResponse, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL, nil)
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }
    q := req.URL.Query()
    q.Add("q", query)
    q.Add("format", "json")
    req.URL.RawQuery = q.Encode()

    resp, err := s.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    var result struct {
        Results []struct {
            Title   string `json:"title"`
            URL     string `json:"url"`
            Content string `json:"content"`
        } `json:"results"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("parse response: %w", err)
    }

    var results []SearchResult
    for _, r := range result.Results {
        results = append(results, SearchResult{
            Title:   r.Title,
            URL:     r.URL,
            Content: r.Content,
            Source:  "searxng",
        })
    }
    return &SearchResponse{Results: results, Source: "searxng"}, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/agent/tool/search/... -v -run TestSearXNGProvider`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/agent/tool/search/searxng.go internal/agent/tool/search/searxng_test.go
git commit -m "feat: add SearXNG search provider"
```

---

### Task 7: WebSearchTool

**Files:**
- Create: `internal/agent/tool/web_search.go`
- Create: `internal/agent/tool/web_search_test.go`

- [ ] **Step 1: Write failing test**

```go
// web_search_test.go
package tool

import (
    "context"
    "testing"

    "github.com/example/agent-tui/internal/agent/tool/search"
)

type mockWebSearch struct {
    resp *search.SearchResponse
    err  error
}

func (m *mockWebSearch) Name() string { return "mock" }
func (m *mockWebSearch) IsAvailable() bool { return true }
func (m *mockWebSearch) Search(ctx context.Context, q string, max int) (*search.SearchResponse, error) {
    return m.resp, m.err
}

func TestWebSearchTool_Schema(t *testing.T) {
    p := &mockWebSearch{resp: &search.SearchResponse{}}
    tool := NewWebSearchTool(p)
    if tool.Name() != "web_search" {
        t.Fatalf("expected 'web_search', got %q", tool.Name())
    }
    schema := tool.Schema()
    if _, ok := schema.Parameters["query"]; !ok {
        t.Fatal("expected 'query' parameter")
    }
}

func TestWebSearchTool_Execute(t *testing.T) {
    p := &mockWebSearch{
        resp: &search.SearchResponse{
            Results: []search.SearchResult{
                {Title: "Result 1", URL: "https://example.com", Content: "Snippet"},
            },
            Source: "mock",
        },
    }
    tool := NewWebSearchTool(p)
    result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
        "query": "hello world",
    })
    if !result.Success {
        t.Fatalf("expected success, got error: %s", result.Error)
    }
    data := result.Data.(string)
    if !contains(data, "Result 1") {
        t.Fatalf("expected result in output, got: %s", data)
    }
}

func TestWebSearchTool_MissingQuery(t *testing.T) {
    p := &mockWebSearch{}
    tool := NewWebSearchTool(p)
    result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{})
    if result.Success {
        t.Fatal("expected failure for missing query")
    }
}

func containsStr(s, substr string) bool {
    for i := 0; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return true
        }
    }
    return false
}

func contains(s, substr string) bool {
    return len(s) >= len(substr) && containsStr(s, substr)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/agent/tool/... -v -run TestWebSearchTool`
Expected: FAIL with undefined `NewWebSearchTool`

- [ ] **Step 3: Write WebSearchTool**

```go
// web_search.go
package tool

import (
    "context"
    "fmt"
    "strings"

    "github.com/example/agent-tui/internal/agent/tool/search"
)

type WebSearchTool struct {
    provider search.WebSearchProvider
}

func NewWebSearchTool(provider search.WebSearchProvider) *WebSearchTool {
    return &WebSearchTool{provider: provider}
}

func (t *WebSearchTool) Name() string { return "web_search" }

func (t *WebSearchTool) Description() string { return "Search the internet for real-time information" }

func (t *WebSearchTool) Schema() ToolSchema {
    return ToolSchema{
        Parameters: map[string]ParamSchema{
            "query":       {Type: "string", Description: "Search query"},
            "max_results": {Type: "string", Description: "Maximum number of results (optional, default 5)"},
        },
        Required: []string{"query"},
    }
}

func (t *WebSearchTool) Execute(ctx ToolContext, params map[string]interface{}) ToolResult {
    query, _ := params["query"].(string)
    if query == "" {
        return ToolResult{Error: "query parameter is required"}
    }

    maxResults := 5
    if mr, ok := params["max_results"].(string); ok && mr != "" {
        fmt.Sscanf(mr, "%d", &maxResults)
    }

    c := ctx.Context
    if c == nil {
        c = context.Background()
    }

    resp, err := t.provider.Search(c, query, maxResults)
    if err != nil {
        return ToolResult{Error: fmt.Sprintf("web search failed: %s", err)}
    }

    var b strings.Builder
    b.WriteString(fmt.Sprintf("Search results from %s:\n\n", resp.Source))
    for i, r := range resp.Results {
        b.WriteString(fmt.Sprintf("%d. %s\n", i+1, r.Title))
        b.WriteString(fmt.Sprintf("   URL: %s\n", r.URL))
        if r.Content != "" {
            b.WriteString(fmt.Sprintf("   %s\n", r.Content))
        }
        b.WriteString("\n")
    }
    return ToolResult{Success: true, Data: b.String()}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/agent/tool/... -v -run TestWebSearchTool`
Expected: PASS

- [ ] **Step 5: Move `contains` helper to shared test file**

Create `internal/agent/tool/tool_test_helpers.go` (test-only, file name `*_test.go`):

```go
package tool

func stringContains(s, substr string) bool {
    if len(s) < len(substr) {
        return false
    }
    for i := 0; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return true
        }
    }
    return false
}
```

Update `web_search_test.go` to use `stringContains` instead of `contains`/`containsStr`.

- [ ] **Step 6: Commit**

```bash
git add internal/agent/tool/web_search.go internal/agent/tool/web_search_test.go internal/agent/tool/tool_test_helpers.go
git commit -m "feat: add WebSearchTool"
```

---

### Task 8: WebFetchTool

**Files:**
- Create: `internal/agent/tool/web_fetch.go`
- Create: `internal/agent/tool/web_fetch_test.go`

- [ ] **Step 1: Write failing test**

```go
// web_fetch_test.go
package tool

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestWebFetchTool(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`
<html><body>
<h1>Hello Page</h1>
<p>This is the <b>content</b> of the page.</p>
<script>alert('x')</script>
</body></html>
`))
    }))
    defer server.Close()

    tool := &WebFetchTool{}
    result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
        "url": server.URL,
    })
    if !result.Success {
        t.Fatalf("expected success, got error: %s", result.Error)
    }
    data := result.Data.(string)
    if !stringContains(data, "Hello Page") {
        t.Fatalf("expected 'Hello Page' in output, got: %s", data)
    }
    if stringContains(data, "<script>") {
        t.Fatal("script tag should be removed")
    }
}

func TestWebFetchTool_InvalidURL(t *testing.T) {
    tool := &WebFetchTool{}
    result := tool.Execute(ToolContext{}, map[string]interface{}{
        "url": "ftp://invalid",
    })
    if result.Success {
        t.Fatal("expected failure for invalid URL")
    }
}

func TestWebFetchTool_MissingURL(t *testing.T) {
    tool := &WebFetchTool{}
    result := tool.Execute(ToolContext{}, map[string]interface{}{})
    if result.Success {
        t.Fatal("expected failure for missing URL")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/agent/tool/... -v -run TestWebFetchTool`
Expected: FAIL with undefined `WebFetchTool`

- [ ] **Step 3: Write WebFetchTool**

```go
// web_fetch.go
package tool

import (
    "fmt"
    "io"
    "net/http"
    "strings"
    "time"

    "github.com/PuerkitoBio/goquery"
)

type WebFetchTool struct{}

func (t *WebFetchTool) Name() string { return "web_fetch" }

func (t *WebFetchTool) Description() string { return "Fetch and read content from a URL" }

func (t *WebFetchTool) Schema() ToolSchema {
    return ToolSchema{
        Parameters: map[string]ParamSchema{
            "url":        {Type: "string", Description: "URL to fetch"},
            "max_length": {Type: "string", Description: "Maximum characters to return (optional, default 8000)"},
        },
        Required: []string{"url"},
    }
}

func (t *WebFetchTool) Execute(ctx ToolContext, params map[string]interface{}) ToolResult {
    url, _ := params["url"].(string)
    if url == "" {
        return ToolResult{Error: "url parameter is required"}
    }
    if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
        return ToolResult{Error: "invalid URL: only http/https allowed"}
    }

    maxLength := 8000
    if ml, ok := params["max_length"].(string); ok && ml != "" {
        fmt.Sscanf(ml, "%d", &maxLength)
    }

    client := &http.Client{Timeout: 10 * time.Second}
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return ToolResult{Error: fmt.Sprintf("create request: %s", err)}
    }
    req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; AgentTUI/1.0)")

    resp, err := client.Do(req)
    if err != nil {
        return ToolResult{Error: fmt.Sprintf("fetch failed: %s", err)}
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return ToolResult{Error: fmt.Sprintf("HTTP %d", resp.StatusCode)}
    }

    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return ToolResult{Error: fmt.Sprintf("parse HTML: %s", err)}
    }

    doc.Find("script, style, nav, footer, header, noscript").Remove()

    text := strings.TrimSpace(doc.Find("body").Text())

    collapsed := strings.Join(strings.Fields(text), " ")

    if len(collapsed) > maxLength {
        collapsed = collapsed[:maxLength] + "... [truncated]"
    }

    return ToolResult{Success: true, Data: collapsed}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/agent/tool/... -v -run TestWebFetchTool`
Expected: PASS

- [ ] **Step 5: Run tests**

Run: `go test ./internal/agent/tool/... -v -run TestWebFetchTool`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/agent/tool/web_fetch.go internal/agent/tool/web_fetch_test.go
git commit -m "feat: add WebFetchTool"
```

---

### Task 9: Wire tools into main.go + config.toml

**Files:**
- Modify: `cmd/agent/main.go`
- Modify: `configs/config.toml`
- Modify: `configs/config.example.toml`

- [ ] **Step 1: Add web search + web fetch to tool registration**

In `cmd/agent/main.go`, after existing tool registrations:

```go
import (
    // ... existing imports ...
    "github.com/example/agent-tui/internal/agent/tool/search"
)

// In main(), after tool registration:
webProviders := []search.WebSearchProvider{
    search.NewTavilyProvider(cfg.WebSearch.TavilyAPIKey),
    search.NewBingProvider(cfg.WebSearch.BingAPIKey),
    search.NewSearXNGProvider(cfg.WebSearch.SearXNGUrl),
    search.NewDuckDuckGoProvider(),
}
toolReg.Register(search.NewWebSearchTool(search.NewChainProvider(webProviders...)))
toolReg.Register(&tool.WebFetchTool{})
```

- [ ] **Step 2: Update configs**

`configs/config.toml`:

```toml
[web_search]
max_results = 5
request_timeout = 10
tavily_api_key = ""
bing_api_key = ""
searxng_url = ""
```

Same for `configs/config.example.toml`.

- [ ] **Step 3: Build and verify**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add cmd/agent/main.go configs/config.toml configs/config.example.toml
git commit -m "feat: wire web search and web fetch tools into agent"
```

---

### Task 10: Skill UI Overlay

**Files:**
- Create: `internal/ui/skill_overlay.go`
- Modify: `internal/ui/app.go`
- Create: `internal/ui/skill_overlay_test.go`

- [ ] **Step 1: Write failing test for skill overlay creation**

```go
// skill_overlay_test.go
package ui

import (
    "testing"
)

func TestNewSkillOverlay(t *testing.T) {
    // Verify basic creation (needs tview app context for full test)
    // This is a lightweight construction test
    skills := []SkillItem{
        {Name: "code-review", Description: "Review Go code"},
        {Name: "analyze", Description: "Analyze project"},
    }
    overlay := NewSkillOverlay(skills, func(name, args string) {})
    if overlay == nil {
        t.Fatal("expected non-nil overlay")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/... -v -run TestNewSkillOverlay`
Expected: FAIL with undefined `SkillItem`, `NewSkillOverlay`

- [ ] **Step 3: Write SkillOverlay component**

```go
// skill_overlay.go
package ui

import (
    "strings"

    "github.com/gdamore/tcell/v2"
    "github.com/rivo/tview"
)

type SkillItem struct {
    Name        string
    Description string
}

type SkillOverlay struct {
    *tview.Flex
    list       *tview.List
    skills     []SkillItem
    filtered   []SkillItem
    onExecute  func(name, args string)
    visible    bool
    filterText string
}

func NewSkillOverlay(skills []SkillItem, onExecute func(name, args string)) *SkillOverlay {
    so := &SkillOverlay{
        Flex:      tview.NewFlex().SetDirection(tview.FlexRow),
        list:      tview.NewList(),
        skills:    skills,
        onExecute: onExecute,
    }
    so.list.SetBorder(true).SetTitle(" Skills ").SetTitleAlign(tview.AlignLeft)
    so.list.ShowSecondaryText(true)
    so.list.SetSelectedFocusOnly(true)
    so.AddItem(so.list, 0, 1, true)
    so.applyFilter("")
    return so
}

func (so *SkillOverlay) applyFilter(prefix string) {
    so.filterText = prefix
    so.list.Clear()
    so.filtered = nil
    for _, s := range so.skills {
        if prefix == "" || strings.HasPrefix(s.Name, prefix) {
            so.filtered = append(so.filtered, s)
            so.list.AddItem(s.Name, s.Description, 0, func() {
                if so.onExecute != nil {
                    so.onExecute(s.Name, "")
                }
            })
        }
    }
}

func (so *SkillOverlay) Filter(prefix string) {
    so.applyFilter(prefix)
}

func (so *SkillOverlay) SelectedName() string {
    idx := so.list.GetCurrentItem()
    if idx < 0 || idx >= len(so.filtered) {
        return ""
    }
    return so.filtered[idx].Name
}

func (so *SkillOverlay) Show() {
    so.visible = true
}

func (so *SkillOverlay) Hide() {
    so.visible = false
}

func (so *SkillOverlay) IsVisible() bool {
    return so.visible
}
```

- [ ] **Step 4: Add SkillItem type export**

In `internal/ui/types.go` or wherever types are defined, export `SkillItem` for use by the app. If no types.go exists, define `SkillItem` in `skill_overlay.go`.

- [ ] **Step 5: Wire SkillOverlay into App**

In `internal/ui/app.go`:

Add fields to `App` struct:
```go
skillOverlay   *SkillOverlay
skillMode      bool
```

In `App` initialization (or a new `SetSkillOverlay` method):
```go
func (a *App) SetSkillOverlay(overlay *SkillOverlay) {
    a.skillOverlay = overlay
}
```

Modify input handler:
- When input starts with `/` and `skillOverlay != nil`: enter skill mode, show overlay
- `Esc` in skill mode: hide overlay, exit skill mode
- `Tab` in skill mode: complete selected skill name
- `Enter` in skill mode: execute selected skill via `skillExecutor`
- `Ctrl+P`: toggle skill overlay

- [ ] **Step 6: Wire skill list from registry**

In `cmd/agent/main.go`, after creating `skillRegistry`:
```go
var skillItems []ui.SkillItem
for _, s := range skillRegistry.List() {
    skillItems = append(skillItems, ui.SkillItem{Name: s.Name, Description: s.Description})
}
overlay := ui.NewSkillOverlay(skillItems, func(name, args string) {
    cmd := &service.ParsedCommand{Name: name, Args: args}
    result, _ := skillExecutor.Execute(cmd)
    // write result to session
})
app.SetSkillOverlay(overlay)
```

- [ ] **Step 7: Build and verify**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 8: Run all tool + ui tests**

Run: `go test ./internal/agent/tool/... ./internal/ui/... -v 2>&1 | tail -20`
Expected: all pass

- [ ] **Step 9: Commit**

```bash
git add internal/ui/skill_overlay.go internal/ui/skill_overlay_test.go cmd/agent/main.go
git commit -m "feat: add Skill UI overlay component"
```

---

### Task 11: Run full test suite + final cleanup

- [ ] **Step 1: Run all tests**

Run: `go test ./... 2>&1`
Expected: all pass (except `TestExternalAgentRole` timeout which is pre-existing)

- [ ] **Step 2: Run go vet**

Run: `go vet ./...`
Expected: no output

- [ ] **Step 3: Push**

```bash
git push
```
