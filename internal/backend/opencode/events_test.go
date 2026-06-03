package opencode

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestEvents(t *testing.T) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)

	mux.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		flusher, ok := w.(http.Flusher)
		if !ok {
			return
		}
		fmt.Fprintf(w, "event: message.chunk\ndata: {\"content\":\"hello\"}\n\n")
		flusher.Flush()
		fmt.Fprintf(w, "event: message.completed\ndata: {\"content\":\"world\"}\n\n")
		flusher.Flush()
	})

	b := NewBackend(Config{AutoStart: false, APIURL: srv.URL})
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	events, err := b.Events(ctx)
	if err != nil {
		t.Fatalf("Events failed: %v", err)
	}

	var count int
	for range events {
		count++
		if count >= 2 {
			break
		}
	}
	if count < 2 {
		t.Errorf("expected at least 2 events, got %d", count)
	}
}
