package search

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBingProvider_Name(t *testing.T) {
	p := NewBingProvider("test-key")
	if p.Name() != "bing" {
		t.Fatalf("expected bing, got %q", p.Name())
	}
}

func TestBingProvider_Search_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Ocp-Apim-Subscription-Key") != "test-key" {
			t.Fatalf("expected test-key header, got %q", r.Header.Get("Ocp-Apim-Subscription-Key"))
		}
		if r.URL.Query().Get("q") != "golang" {
			t.Fatalf("expected q=golang, got %q", r.URL.Query().Get("q"))
		}
		if r.URL.Query().Get("count") != "5" {
			t.Fatalf("expected count=5, got %q", r.URL.Query().Get("count"))
		}

		resp := bingResponse{
			WebPages: &bingWebPages{
				Value: []bingResult{
					{Name: "Go Programming", URL: "https://golang.org", Snippet: "Go is fast"},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &BingProvider{
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
	if results[0].Title != "Go Programming" {
		t.Fatalf("expected 'Go Programming', got %q", results[0].Title)
	}
	if results[0].URL != "https://golang.org" {
		t.Fatalf("expected https://golang.org, got %q", results[0].URL)
	}
	if results[0].Snippet != "Go is fast" {
		t.Fatalf("expected 'Go is fast', got %q", results[0].Snippet)
	}
}

func TestBingProvider_EmptyAPIKey(t *testing.T) {
	p := NewBingProvider("")
	_, err := p.Search(context.Background(), "test", 5)
	if err == nil {
		t.Fatal("expected error for empty API key")
	}
}

func TestBingProvider_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	p := &BingProvider{
		client:  server.Client(),
		apiKey:  "bad-key",
		baseURL: server.URL,
	}

	_, err := p.Search(context.Background(), "test", 5)
	if err == nil {
		t.Fatal("expected error for unauthorized")
	}
}

func TestBingProvider_RateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	p := &BingProvider{
		client:  server.Client(),
		apiKey:  "test-key",
		baseURL: server.URL,
	}

	_, err := p.Search(context.Background(), "test", 5)
	if err == nil {
		t.Fatal("expected error for rate limit")
	}
}

func TestBingProvider_NoResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := bingResponse{}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &BingProvider{
		client:  server.Client(),
		apiKey:  "test-key",
		baseURL: server.URL,
	}

	results, err := p.Search(context.Background(), "nonexistent", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}
