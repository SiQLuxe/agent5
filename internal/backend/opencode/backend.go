package opencode

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/example/agent-tui/internal/backend"
)

var errNotImplemented = errors.New("not implemented")

type Config struct {
	Binary    string `json:"binary"`
	AutoStart bool   `json:"auto_start"`
	APIURL    string `json:"api_url"`
	APIKey    string `json:"api_key"`
}

type OpencodeBackend struct {
	config  Config
	client  *http.Client
	baseURL string
	procMgr *backend.ProcessManager
	sseConn *sseConnection
	mu      sync.Mutex
}

func NewBackend(cfg Config) *OpencodeBackend {
	return &OpencodeBackend{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (b *OpencodeBackend) Name() string {
	return string(backend.TypeOpencode)
}

func (b *OpencodeBackend) Start(ctx context.Context) error {
	if b.config.AutoStart {
		args := []string{"serve"}
		binary := b.config.Binary
		if binary == "" {
			binary = "opencode"
		}
		b.procMgr = backend.NewProcessManager(binary, args, nil)
		if err := b.procMgr.Start(ctx); err != nil {
			return fmt.Errorf("failed to start opencode: %w", err)
		}
		b.baseURL = "http://127.0.0.1:4096"
	} else {
		b.baseURL = b.config.APIURL
	}

	h, err := b.Health(ctx)
	if err != nil {
		return fmt.Errorf("opencode health check failed: %w", err)
	}
	if !h.Healthy {
		return fmt.Errorf("opencode is not healthy")
	}
	return nil
}

func (b *OpencodeBackend) Stop(ctx context.Context) error {
	if b.sseConn != nil {
		b.sseConn.Close()
	}
	if b.procMgr != nil {
		return b.procMgr.Stop(ctx)
	}
	return nil
}

func (b *OpencodeBackend) Health(ctx context.Context) (*backend.HealthInfo, error) {
	if b.baseURL == "" {
		return &backend.HealthInfo{Healthy: false}, nil
	}
	url := b.baseURL + "/global/health"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return &backend.HealthInfo{Healthy: false}, nil
	}
	defer resp.Body.Close()
	return &backend.HealthInfo{Healthy: resp.StatusCode == http.StatusOK}, nil
}

func (b *OpencodeBackend) CreateSession(ctx context.Context, title string) (*backend.Session, error) {
	return nil, errNotImplemented
}

func (b *OpencodeBackend) ListSessions(ctx context.Context) ([]*backend.Session, error) {
	return nil, errNotImplemented
}

func (b *OpencodeBackend) GetSession(ctx context.Context, id string) (*backend.Session, error) {
	return nil, errNotImplemented
}

func (b *OpencodeBackend) DeleteSession(ctx context.Context, id string) error {
	return errNotImplemented
}

func (b *OpencodeBackend) SendMessage(ctx context.Context, sessionID string, msg *backend.Message) (*backend.MessageResult, error) {
	return nil, errNotImplemented
}

func (b *OpencodeBackend) SendMessageStream(ctx context.Context, sessionID string, msg *backend.Message, onChunk func(*backend.Chunk)) error {
	return errNotImplemented
}

func (b *OpencodeBackend) GetMessages(ctx context.Context, sessionID string) ([]*backend.Message, error) {
	return nil, errNotImplemented
}

func (b *OpencodeBackend) ExecuteCommand(ctx context.Context, sessionID string, command string) (*backend.CommandResult, error) {
	return nil, errNotImplemented
}

func (b *OpencodeBackend) ExecuteShell(ctx context.Context, command string) (*backend.CommandResult, error) {
	return nil, errNotImplemented
}

func (b *OpencodeBackend) ReadFile(ctx context.Context, path string) (string, error) {
	return "", errNotImplemented
}

func (b *OpencodeBackend) SearchText(ctx context.Context, pattern string) ([]backend.SearchResult, error) {
	return nil, errNotImplemented
}

func (b *OpencodeBackend) Events(ctx context.Context) (<-chan *backend.Event, error) {
	return nil, errNotImplemented
}

func init() {
	backend.RegisterBackendBuilder(backend.TypeOpencode, func(cfg backend.BackendConfig) (backend.AgentBackend, error) {
		return NewBackend(Config{
			Binary:    cfg.Binary,
			AutoStart: cfg.AutoStart,
			APIURL:    cfg.APIURL,
			APIKey:    cfg.APIKey,
		}), nil
	})
}

type sseConnection struct {
	closed bool
}

func (s *sseConnection) Close() {}
