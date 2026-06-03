package server

import (
	"encoding/json"
	"net/http"
)

type appendPromptRequest struct {
	Text string `json:"text"`
}

type showToastRequest struct {
	Title   string `json:"title,omitempty"`
	Message string `json:"message"`
	Variant string `json:"variant,omitempty"`
}

type executeCommandRequest struct {
	Command string `json:"command"`
}

func (s *Server) handleAppendPrompt(w http.ResponseWriter, r *http.Request) {
	var req appendPromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if s.handler != nil {
		s.handler.AppendPrompt(req.Text)
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (s *Server) handleSubmitPrompt(w http.ResponseWriter, r *http.Request) {
	if s.handler != nil {
		s.handler.SubmitPrompt()
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (s *Server) handleShowToast(w http.ResponseWriter, r *http.Request) {
	var req showToastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if s.handler != nil {
		s.handler.ShowToast(req.Title, req.Message, req.Variant)
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (s *Server) handleExecuteCommand(w http.ResponseWriter, r *http.Request) {
	var req executeCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if s.handler != nil {
		s.handler.ExecuteCommand(req.Command)
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
