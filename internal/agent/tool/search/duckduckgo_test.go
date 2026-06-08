package search

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDuckDuckGoProvider_Name(t *testing.T) {
	p := NewDuckDuckGoProvider()
	if p.Name() != "duckduckgo" {
		t.Fatalf("expected duckduckgo, got %q", p.Name())
	}
}

func TestDuckDuckGoProvider_Search_Success(t *testing.T) {
	html := `<html><body><table><tr>
<td><a class="result-link" href="http://example.com/1">Title 1</a></td>
</tr><tr>
<td class="result-snippet">Snippet 1</td>
</tr><tr>
<td><a class="result-link" href="http://example.com/2">Title 2</a></td>
</tr><tr>
<td class="result-snippet">Snippet 2</td>
</tr></table></body></html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if r.FormValue("q") != "test query" {
			t.Fatalf("expected q=test query, got %q", r.FormValue("q"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	}))
	defer server.Close()

	p := &DuckDuckGoProvider{
		client:  http.DefaultClient,
		baseURL: server.URL,
	}

	results, err := p.Search(context.Background(), "test query", 5)
	if err != nil {
		t.Fatalf("search error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Title != "Title 1" {
		t.Fatalf("expected 'Title 1', got %q", results[0].Title)
	}
	if results[0].URL != "http://example.com/1" {
		t.Fatalf("expected http://example.com/1, got %q", results[0].URL)
	}
	if results[0].Snippet != "Snippet 1" {
		t.Fatalf("expected 'Snippet 1', got %q", results[0].Snippet)
	}
}

func TestDuckDuckGoProvider_RateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	p := &DuckDuckGoProvider{
		client:  http.DefaultClient,
		baseURL: server.URL,
	}

	_, err := p.Search(context.Background(), "test", 5)
	if err == nil {
		t.Fatal("expected error for rate limited response")
	}
	if !strings.Contains(err.Error(), "rate limited") {
		t.Fatalf("expected rate limit error, got: %v", err)
	}
}

func TestDuckDuckGoProvider_UnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	p := &DuckDuckGoProvider{
		client:  http.DefaultClient,
		baseURL: server.URL,
	}

	_, err := p.Search(context.Background(), "test", 5)
	if err == nil {
		t.Fatal("expected error for unexpected status")
	}
}

func TestDuckDuckGoProvider_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := &DuckDuckGoProvider{
		client:  http.DefaultClient,
		baseURL: server.URL,
	}

	_, err := p.Search(ctx, "test", 5)
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
}

func TestParseDuckDuckGoResults(t *testing.T) {
	html := `<html><body><table><tr>
<td><a class="result-link" href="http://example.com/1">Result One</a></td>
</tr><tr>
<td class="result-snippet">First snippet text</td>
</tr><tr>
<td><a class="result-link" href="http://example.com/2">Result Two</a></td>
</tr><tr>
<td class="result-snippet">Second snippet text</td>
</tr></table></body></html>`

	results, err := parseDuckDuckGoResults(strings.NewReader(html), 5)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Title != "Result One" {
		t.Fatalf("expected 'Result One', got %q", results[0].Title)
	}
	if results[0].URL != "http://example.com/1" {
		t.Fatalf("expected http://example.com/1, got %q", results[0].URL)
	}
	if results[0].Snippet != "First snippet text" {
		t.Fatalf("expected 'First snippet text', got %q", results[0].Snippet)
	}
	if results[1].Title != "Result Two" {
		t.Fatalf("expected 'Result Two', got %q", results[1].Title)
	}
}

func TestParseDuckDuckGoResults_MaxResults(t *testing.T) {
	html := `<html><body><table><tr>
<td><a class="result-link" href="http://example.com/1">One</a></td>
</tr><tr>
<td class="result-snippet">Snippet 1</td>
</tr><tr>
<td><a class="result-link" href="http://example.com/2">Two</a></td>
</tr><tr>
<td class="result-snippet">Snippet 2</td>
</tr><tr>
<td><a class="result-link" href="http://example.com/3">Three</a></td>
</tr><tr>
<td class="result-snippet">Snippet 3</td>
</tr></table></body></html>`

	results, err := parseDuckDuckGoResults(strings.NewReader(html), 2)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestParseDuckDuckGoResults_Empty(t *testing.T) {
	html := `<html><body></body></html>`
	results, err := parseDuckDuckGoResults(strings.NewReader(html), 5)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestParseDuckDuckGoResults_NoMaxResults(t *testing.T) {
	html := `<html><body><table><tr>
<td><a class="result-link" href="http://example.com/1">One</a></td>
</tr><tr>
<td class="result-snippet">Snippet 1</td>
</tr><tr>
<td><a class="result-link" href="http://example.com/2">Two</a></td>
</tr><tr>
<td class="result-snippet">Snippet 2</td>
</tr></table></body></html>`

	results, err := parseDuckDuckGoResults(strings.NewReader(html), 0)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}
