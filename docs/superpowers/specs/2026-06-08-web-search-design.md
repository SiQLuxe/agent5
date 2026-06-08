# Web Search & Fetch Design

**Date:** 2026-06-08
**Status:** Draft

**Goal:** Add `web_search` and `web_fetch` tools to the agent system, allowing the ReAct loop to query the internet for real-time information and read specific web pages.

## Background

The agent system currently has file and command execution tools (`read_file`, `write_file`, `search_text`, `exec_command`, `chat_llm`, `task`) but cannot access the internet. Adding web search enables use cases like researching documentation, checking current information, and fetching reference material during task execution.

## Architecture

```
Agent ReAct Loop
  └─ WebSearchTool.Execute(ctx, params)
       query = params["query"]
       max_results = params.get("max_results", 5)
       └─ WebSearchProvider.Search(ctx, query, maxResults)
            ├─ [1] TavilyProvider       (tavily_api_key set)
            ├─ [2] BingProvider          (bing_api_key set)
            ├─ [3] SearXNGProvider       (searxng_url set)
            └─ [4] DuckDuckGoProvider    (always available, last resort)
  └─ WebFetchTool.Execute(ctx, params)
       url = params["url"]
       max_length = params.get("max_length", 8000)
       └─ http.Get → HTML → extract text → return
```

## Components

### New files

| File | Responsibility |
|------|----------------|
| `internal/agent/tool/web_search.go` | `WebSearchTool` implementing `tool.Tool` |
| `internal/agent/tool/web_fetch.go` | `WebFetchTool` implementing `tool.Tool` |
| `internal/agent/tool/search/provider.go` | `WebSearchProvider` interface, `SearchResponse`, `SearchResult` types |
| `internal/agent/tool/search/chain.go` | `ChainProvider` — tries providers in priority order |
| `internal/agent/tool/search/duckduckgo.go` | DuckDuckGo HTML search (always available, no key) |
| `internal/agent/tool/search/tavily.go` | Tavily API search (optional, needs key) |
| `internal/agent/tool/search/bing.go` | Bing Web Search API (optional, needs key) |
| `internal/agent/tool/search/searxng.go` | SearXNG self-hosted search (optional, needs URL) |

### Types

```go
// search/provider.go
type SearchResult struct {
    Title   string
    URL     string
    Content string   // summary/snippet
    Source  string   // provider name
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

### ChainProvider

```go
// search/chain.go
type ChainProvider struct {
    providers []WebSearchProvider
}

func NewChainProvider(providers ...WebSearchProvider) *ChainProvider
func (c *ChainProvider) Search(ctx context.Context, query string, maxResults int) (*SearchResponse, error)
```

`Search` iterates providers in order. For each that `IsAvailable()`, attempts search. Returns first success. If all fail, returns combined error.

## Provider Implementations

### DuckDuckGo (`search/duckduckgo.go`)
- Always available, no configuration needed
- HTTP GET `https://html.duckduckgo.com/html/?q=<query>`
- Parse HTML with goquery: `result__title` anchors + `result__snippet` divs
- Set `User-Agent` header (e.g., `Mozilla/5.0`)
- Rate limiting: 1 request per second minimum
- No official API, relies on scraping — stability is acceptable but not guaranteed

### Tavily (`search/tavily.go`)
- Requires `tavily_api_key` in config
- POST `https://api.tavily.com/search`
- JSON request: `{api_key, query, search_depth: "basic", max_results, include_answer: false}`
- JSON response: parse `results[].{title, url, content}`
- Free tier: 1000 requests/month

### Bing (`search/bing.go`)
- Requires `bing_api_key` in config
- GET `https://api.bing.microsoft.com/v7.0/search?q=<query>&count=<max>`
- Header: `Ocp-Apim-Subscription-Key`
- JSON response: parse `webPages.value[].{name, url, snippet}`
- Free tier: 1000 requests/month (S1)

### SearXNG (`search/searxng.go`)
- Requires `searxng_url` in config
- GET `{searxng_url}/search?q=<query>&format=json`
- JSON response: parse `results[].{title, url, content}`
- No API key needed, runs on user's own server

## WebFetchTool

```go
func (t *WebFetchTool) Execute(ctx ToolContext, params map[string]interface{}) ToolResult
```

Parameters:
- `url` (required) — target URL, must be http/https
- `max_length` (optional, default 8000) — max characters to return

Behavior:
1. Validate URL scheme (http/https only)
2. HTTP GET with 10s timeout and User-Agent
3. Parse HTML with goquery
4. Extract `<body>` text content
5. Remove `<script>`, `<style>`, `<nav>`, `<footer>`, `<header>` tags
6. Collapse whitespace, truncate to `max_length`
7. Return as tool result

## Config

Add to `configs/config.toml`:

```toml
[web_search]
max_results = 5
request_timeout = 10
tavily_api_key = ""
bing_api_key = ""
searxng_url = ""
```

New config types in `internal/data/config/config.go`:

```go
type WebSearchConfig struct {
    MaxResults     int    `toml:"max_results"`
    RequestTimeout int    `toml:"request_timeout"`
    TavilyAPIKey   string `toml:"tavily_api_key"`
    BingAPIKey     string `toml:"bing_api_key"`
    SearXNGUrl     string `toml:"searxng_url"`
}
```

## Wiring (cmd/agent/main.go)

```go
webCfg := cfg.WebSearch
providers := []search.WebSearchProvider{
    search.NewTavilyProvider(webCfg.TavilyAPIKey),
    search.NewBingProvider(webCfg.BingAPIKey),
    search.NewSearXNGProvider(webCfg.SearXNGUrl),
    search.NewDuckDuckGoProvider(),
}
chain := search.NewChainProvider(providers...)
toolReg.Register(search.NewWebSearchTool(chain))
toolReg.Register(&tool.WebFetchTool{})
```

Only providers with keys/URL set will be available. DuckDuckGo is always the final fallback.

## Error Handling

| Scenario | Behavior |
|----------|----------|
| No provider available | Tool returns error: "no web search provider available" |
| Provider timeout | ChainProvider tries next provider |
| All providers fail | Tool returns combined error from all providers |
| Invalid URL in web_fetch | Tool error: "invalid URL" |
| HTTP error in web_fetch | Tool error with status code |
| Response too large | Truncated to max_length, result includes truncation notice |

## Testing

| Component | Approach |
|-----------|----------|
| `ChainProvider` | Mock providers returning success/failure, verify chain logic |
| `DuckDuckGoProvider` | Mock HTTP server returning sample HTML |
| `TavilyProvider` | Mock HTTP server returning sample JSON |
| `BingProvider` | Mock HTTP server returning sample JSON |
| `SearXNGProvider` | Mock HTTP server returning sample JSON |
| `WebSearchTool` | Mock provider, verify schema + execute flow |
| `WebFetchTool` | Mock HTTP server with sample HTML page |
| Config parsing | Test TOML with/without keys |

## Skill UI Overlay

See `docs/superpowers/specs/2026-06-01-skill-system-design.md` for the complete skill UI overlay design. The backend (registry, loader, parser, executor) is implemented. Remaining work:

- `internal/ui/skill_overlay.go` — tview overlay with filter, navigation, Tab completion
- `internal/ui/app.go` — ModeSkill state, `/` detection, Ctrl+P shortcut
- Wiring: `/command` → overlay → parser → executor → RoleSkill message
