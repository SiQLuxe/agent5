package service

import (
	"github.com/example/agent-tui/internal/ai"
	"github.com/example/agent-tui/internal/data/history"
)

type AIAssistant struct {
	client  ai.Client
	history *history.History
}

func NewAIAssistant(client ai.Client, history *history.History) *AIAssistant {
	return &AIAssistant{
		client:  client,
		history: history,
	}
}

func (a *AIAssistant) Chat(sessionID, message string) (string, error) {
	a.history.AddMessage(sessionID, "user", message)

	msgs := a.history.GetMessages(sessionID)
	aiMessages := make([]ai.Message, len(msgs))
	for i, msg := range msgs {
		aiMessages[i] = ai.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	req := ai.ChatCompletionRequest{
		Model:    a.client.GetModel(),
		Messages: aiMessages,
		Stream:   false,
	}

	resp, err := a.client.ChatCompletion(req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) > 0 {
		content := resp.Choices[0].Message.Content
		a.history.AddMessage(sessionID, "assistant", content)
		return content, nil
	}

	return "", nil
}

func (a *AIAssistant) ChatStream(sessionID, message string, callback func(string)) error {
	a.history.AddMessage(sessionID, "user", message)

	msgs := a.history.GetMessages(sessionID)
	aiMessages := make([]ai.Message, len(msgs))
	for i, msg := range msgs {
		aiMessages[i] = ai.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	req := ai.ChatCompletionRequest{
		Model:    a.client.GetModel(),
		Messages: aiMessages,
		Stream:   true,
	}

	var fullResponse string
	err := a.client.ChatCompletionStream(req, func(content string) {
		fullResponse += content
		callback(content)
	})

	if err == nil && fullResponse != "" {
		a.history.AddMessage(sessionID, "assistant", fullResponse)
	}

	return err
}

func (a *AIAssistant) SwitchModel(model string) {
	a.client.SetAPIKey("")
}

func (a *AIAssistant) ListModels() ([]string, error) {
	return a.client.ListModels()
}

func (a *AIAssistant) CreateSession(name string) string {
	return a.history.CreateSession(name)
}

func (a *AIAssistant) ListSessions() []history.SessionInfo {
	return a.history.GetSessions()
}