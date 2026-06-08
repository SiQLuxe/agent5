package search

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTavilyProvider_Name(t *testing.T) {
	p := NewTavilyProvider("test-key")
	if p.Name() != "tavily" {
		t.Fatalf("expected tavily, got %q", p.Name())
	}
}

func TestTavilyProvider_Search_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/search" {
			t.Fatalf("expected /search, got %s", r.URL.Path)
		}

		var req tavilyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.APIKey != "test-key" {
			t.Fatalf("expected test-key, got %q", req.APIKey)
		}
		if req.Query != "golang" {
			t.Fatalf("expected golang, got %q", req.Query)
		}
		if req.MaxResults != 5 {
			t.Fatalf("expected 5, got %d", req.MaxResults)
		}
		if req.SearchDepth != "basic" {
			t.Fatalf("expected basic, got %q", req.SearchDepth)
		}

		resp := tavilyResponse{
			Results: []tavilyResult{
				{Title: "Go Language", URL: "https://golang.org", Content: "Go is a programming language", Score: 0.9},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &TavilyProvider{
		client:  server.Client(),
		apiKey:  "test-key",
		baseURL: server.URL,
	}

	results, err := p.Search(context.Background(), "golang", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Title != "Go Language" {
		t.Fatalf("expected 'Go Language', got %q", results[0].Title)
	}
	if results[0].URL != "https://golang.org" {
		t.Fatalf("expected https://golang.org, got %q", results[0].URL)
	}
	if results[0].Snippet != "Go is a programming language" {
		t.Fatalf("expected 'Go is a programming language', got %q", results[0].Snippet)
	}
}

func TestTavilyProvider_EmptyAPIKey(t *testing.T) {
	p := NewTavilyProvider("")
	_, err := p.Search(context.Background(), "test", 5)
	if err == nil {
		t.Fatal("expected error for empty API key")
	}
}

func TestTavilyProvider_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	p := &TavilyProvider{
		client:  server.Client(),
		apiKey:  "bad-key",
		baseURL: server.URL,
	}

	_, err := p.Search(context.Background(), "test", 5)
	if err == nil {
		t.Fatal("expected error for unauthorized")
	}
}

func TestTavilyProvider_RateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	p := &TavilyProvider{
		client:  server.Client(),
		apiKey:  "test-key",
		baseURL: server.URL,
	}

	_, err := p.Search(context.Background(), "test", 5)
	if err == nil {
		t.Fatal("expected error for rate limit")
	}
}

func TestTavilyContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := NewTavilyProvider("test-key")
	_, err := p.Search(ctx, "test", 5)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
