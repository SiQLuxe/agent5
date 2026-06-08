package search

import (
	"context"
	"errors"
	"testing"
)

// mockProvider implements WebSearchProvider for testing.
type mockProvider struct {
	name    string
	results []WebSearchResult
	err     error
}

func (m *mockProvider) Search(_ context.Context, query string, count int) ([]WebSearchResult, error) {
	return m.results, m.err
}

func (m *mockProvider) Name() string { return m.name }

func TestChainProvider_EmptyProviders(t *testing.T) {
	cp := NewChainProvider()
	results, err := cp.Search(context.Background(), "test", 5)
	if err == nil {
		t.Fatal("expected error for empty providers")
	}
	if results != nil {
		t.Fatal("expected nil results for empty providers")
	}
}

func TestChainProvider_FirstProviderSucceeds(t *testing.T) {
	p1 := &mockProvider{
		name: "provider1",
		results: []WebSearchResult{
			{Title: "Result 1", URL: "http://example.com/1", Snippet: "Snippet 1"},
		},
	}
	p2 := &mockProvider{
		name: "provider2",
		results: []WebSearchResult{
			{Title: "Result 2", URL: "http://example.com/2", Snippet: "Snippet 2"},
		},
	}
	cp := NewChainProvider(p1, p2)
	results, err := cp.Search(context.Background(), "test", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Title != "Result 1" {
		t.Fatalf("expected 'Result 1', got %q", results[0].Title)
	}
}

func TestChainProvider_FallbackOnFailure(t *testing.T) {
	p1 := &mockProvider{
		name: "failing",
		err:  errors.New("provider error"),
	}
	p2 := &mockProvider{
		name: "succeeding",
		results: []WebSearchResult{
			{Title: "Fallback Result", URL: "http://example.com/fallback", Snippet: "Fallback"},
		},
	}
	cp := NewChainProvider(p1, p2)
	results, err := cp.Search(context.Background(), "test", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Title != "Fallback Result" {
		t.Fatalf("expected 'Fallback Result', got %q", results[0].Title)
	}
}

func TestChainProvider_AllFail(t *testing.T) {
	p1 := &mockProvider{
		name: "fail1",
		err:  errors.New("error 1"),
	}
	p2 := &mockProvider{
		name: "fail2",
		err:  errors.New("error 2"),
	}
	cp := NewChainProvider(p1, p2)
	results, err := cp.Search(context.Background(), "test", 5)
	if err == nil {
		t.Fatal("expected error when all providers fail")
	}
	if results != nil {
		t.Fatal("expected nil results when all providers fail")
	}
}

func TestChainProvider_Name(t *testing.T) {
	p1 := &mockProvider{name: "a"}
	p2 := &mockProvider{name: "b"}
	cp := NewChainProvider(p1, p2)
	expected := "chain(a,b)"
	if cp.Name() != expected {
		t.Fatalf("expected %q, got %q", expected, cp.Name())
	}
}

func TestChainProvider_EmptyName(t *testing.T) {
	cp := NewChainProvider()
	expected := "chain(empty)"
	if cp.Name() != expected {
		t.Fatalf("expected %q, got %q", expected, cp.Name())
	}
}
