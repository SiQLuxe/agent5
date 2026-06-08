package search

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearXNGProvider_Name(t *testing.T) {
	p := NewSearXNGProvider("http://localhost:8888")
	if p.Name() != "searxng" {
		t.Fatalf("expected searxng, got %q", p.Name())
	}
}

func TestSearXNGProvider_Search_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Query().Get("q") != "golang" {
			t.Fatalf("expected q=golang, got %q", r.URL.Query().Get("q"))
		}
		if r.URL.Query().Get("format") != "json" {
			t.Fatalf("expected format=json, got %q", r.URL.Query().Get("format"))
		}

		resp := searxngResponse{
			Results: []searxngResult{
				{Title: "Go Language", URL: "https://golang.org", Content: "Go is a programming language", Engine: "google"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &SearXNGProvider{
		client:  server.Client(),
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
	if results[0].Content != "Go is a programming language" {
		t.Fatalf("expected 'Go is a programming language', got %q", results[0].Content)
	}
}

func TestSearXNGProvider_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	p := &SearXNGProvider{
		client:  server.Client(),
		baseURL: server.URL,
	}

	_, err := p.Search(context.Background(), "test", 5)
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestSearXNGProvider_EmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := searxngResponse{Results: []searxngResult{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &SearXNGProvider{
		client:  server.Client(),
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
