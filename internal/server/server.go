package server

import (
	"context"
	"net/http"
	"sync"
)

type TUIActionHandler interface {
	AppendPrompt(text string)
	SubmitPrompt()
	ShowToast(title, message, variant string)
	ExecuteCommand(command string)
}

type Server struct {
	handler TUIActionHandler
	srv     *http.Server
	mu      sync.Mutex
}

func New(handler TUIActionHandler) *Server {
	return &Server{handler: handler}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/tui/append-prompt", s.handleAppendPrompt)
	mux.HandleFunc("POST /api/tui/submit-prompt", s.handleSubmitPrompt)
	mux.HandleFunc("POST /api/tui/show-toast", s.handleShowToast)
	mux.HandleFunc("POST /api/tui/execute-command", s.handleExecuteCommand)
	return mux
}

func (s *Server) Start(ctx context.Context, addr string) error {
	s.mu.Lock()
	s.srv = &http.Server{
		Addr:    addr,
		Handler: s.Handler(),
	}
	s.mu.Unlock()
	return s.srv.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.srv != nil {
		return s.srv.Shutdown(ctx)
	}
	return nil
}
