package opencode

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/example/agent-tui/internal/backend"
)

type sseConnection struct {
	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
	out    chan *backend.Event
	closed bool
}

func (b *OpencodeBackend) Events(ctx context.Context) (<-chan *backend.Event, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.sseConn != nil && !b.sseConn.closed {
		return b.sseConn.out, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	b.sseConn = &sseConnection{
		ctx:    ctx,
		cancel: cancel,
		out:    make(chan *backend.Event, 100),
	}

	go b.sseLoop(ctx, b.sseConn.out)
	return b.sseConn.out, nil
}

func (b *OpencodeBackend) sseLoop(ctx context.Context, out chan<- *backend.Event) {
	defer close(out)

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
	if !s.closed {
		s.cancel()
		s.closed = true
	}
}
