package opencode

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/example/agent-tui/internal/backend"
)

type sseConnection struct {
	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
	out    chan *backend.Event
	closed atomic.Bool
}

func (b *OpencodeBackend) Events(ctx context.Context) (<-chan *backend.Event, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.sseConn != nil && !b.sseConn.closed.Load() {
		return b.sseConn.out, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	b.sseConn = &sseConnection{
		ctx:    ctx,
		cancel: cancel,
		out:    make(chan *backend.Event, 100),
	}

	go b.sseLoop(ctx, b.sseConn)
	return b.sseConn.out, nil
}

func (b *OpencodeBackend) sseLoop(ctx context.Context, conn *sseConnection) {
	out := conn.out
	defer close(out)
	defer conn.closed.Store(true)

	req, err := http.NewRequestWithContext(ctx, "GET", b.baseURL+"/event", nil)
	if err != nil {
		return
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	var eventType string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "event: ") {
			eventType = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			var payload interface{}
			if err := json.Unmarshal([]byte(data), &payload); err != nil {
				payload = data
			}
			select {
			case out <- &backend.Event{Type: eventType, Payload: payload}:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (s *sseConnection) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.closed.Load() {
		s.cancel()
		s.closed.Store(true)
	}
}

// streamListener wraps an event channel with sessionID filtering for SendMessageStream.
type streamListener struct {
	sessionID string
	out       chan *backend.Chunk
	done      chan struct{}
}

func newStreamListener(sessionID string) *streamListener {
	return &streamListener{
		sessionID: sessionID,
		out:       make(chan *backend.Chunk, 10),
		done:      make(chan struct{}),
	}
}

func (l *streamListener) match(event *backend.Event) bool {
	if event.Type != "message.chunk" && event.Type != "message.completed" {
		return false
	}
	payload, ok := event.Payload.(map[string]interface{})
	if !ok {
		return false
	}
	sid, ok := payload["session_id"].(string)
	return ok && sid == l.sessionID
}

func (b *OpencodeBackend) sendMessageStreamViaSSE(ctx context.Context, sessionID string, onChunk func(*backend.Chunk)) error {
	events, err := b.Events(ctx)
	if err != nil {
		return err
	}

	for event := range events {
		if event.Type == "message.completed" {
			payload, ok := event.Payload.(map[string]interface{})
			if ok {
				sid, _ := payload["session_id"].(string)
				if sid == sessionID {
					if content, exists := payload["content"]; exists {
						onChunk(&backend.Chunk{Content: fmt.Sprintf("%v", content), Done: false})
					}
					onChunk(&backend.Chunk{Done: true})
					return nil
				}
			}
		}
		if event.Type == "message.chunk" {
			payload, ok := event.Payload.(map[string]interface{})
			if ok {
				sid, _ := payload["session_id"].(string)
				if sid == sessionID {
					if content, exists := payload["content"]; exists {
						onChunk(&backend.Chunk{Content: fmt.Sprintf("%v", content), Done: false})
					}
				}
			}
		}
	}
	return nil
}
