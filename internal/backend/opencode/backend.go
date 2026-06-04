package opencode

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/example/agent-tui/internal/backend"
)

type Config struct {
	Binary    string `json:"binary"`
	AutoStart bool   `json:"auto_start"`
	APIURL    string `json:"api_url"`
	APIKey    string `json:"api_key"`
	WorkDir   string `json:"work_dir"`
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
		config:  cfg,
		baseURL: cfg.APIURL,
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
		port, err := findFreePort()
		if err != nil {
			return fmt.Errorf("find free port: %w", err)
		}
		args := []string{"serve", "--port", fmt.Sprintf("%d", port)}
		binary := b.config.Binary
		if binary == "" {
			binary = "opencode"
		}
		b.baseURL = fmt.Sprintf("http://127.0.0.1:%d", port)
		b.procMgr = backend.NewProcessManager(binary, args, nil)
		if b.config.WorkDir != "" {
			b.procMgr.SetDir(b.config.WorkDir)
		}
		if err := b.procMgr.Start(ctx); err != nil {
			return fmt.Errorf("failed to start opencode: %w", err)
		}
	} else {
		b.baseURL = b.config.APIURL
	}

	if err := b.waitForHealth(ctx); err != nil {
		b.cleanupProcess()
		return err
	}
	return nil
}

func (b *OpencodeBackend) cleanupProcess() {
	if b.procMgr != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		b.procMgr.Stop(ctx)
		b.procMgr = nil
	}
}

func (b *OpencodeBackend) waitForHealth(ctx context.Context) error {
	const (
		interval     = 500 * time.Millisecond
		maxWait      = 15 * time.Second
		requestLimit = 2 * time.Second
	)
	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		checkCtx, checkCancel := context.WithTimeout(ctx, requestLimit)
		h, err := b.Health(checkCtx)
		checkCancel()
		if err == nil && h.Healthy {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("health check cancelled: %w", ctx.Err())
		case <-time.After(interval):
		}
	}
	// final attempt
	finalCtx, finalCancel := context.WithTimeout(ctx, requestLimit)
	defer finalCancel()
	h, err := b.Health(finalCtx)
	if err != nil {
		return fmt.Errorf("opencode health check failed: %w", err)
	}
	if !h.Healthy {
		return fmt.Errorf("opencode is not healthy")
	}
	return nil
}

func findFreePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	addr := listener.Addr().(*net.TCPAddr)
	listener.Close()
	return addr.Port, nil
}

func (b *OpencodeBackend) Stop(ctx context.Context) error {
	b.mu.Lock()
	if b.sseConn != nil {
		b.sseConn.Close()
	}
	b.mu.Unlock()
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

func (b *OpencodeBackend) doRequest(ctx context.Context, method, path string, body, result interface{}) error {
	url := b.baseURL + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
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
