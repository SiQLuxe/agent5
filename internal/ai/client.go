package ai

import "errors"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int     `json:"index"`
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type StreamingResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int `json:"index"`
		Delta        struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

type Client interface {
	ChatCompletion(req ChatCompletionRequest) (*ChatCompletionResponse, error)
	ChatCompletionStream(req ChatCompletionRequest, callback func(string)) error
	SetAPIKey(key string)
	SetBaseURL(url string)
	GetModel() string
	ListModels() ([]string, error)
}

type ClientType string

const (
	ClientOpenAI    ClientType = "openai"
	ClientDeepSeek  ClientType = "deepseek"
	ClientAnthropic ClientType = "anthropic"
	ClientLocal     ClientType = "local"
)

func NewClient(clientType ClientType, apiKey, baseURL, model string) (Client, error) {
	switch clientType {
	case ClientOpenAI:
		return NewOpenAIClient(apiKey, baseURL, model)
	case ClientDeepSeek:
		return NewDeepSeekClient(apiKey, baseURL, model)
	case ClientAnthropic:
		return NewAnthropicClient(apiKey, baseURL, model)
	case ClientLocal:
		return NewLocalClient(apiKey, baseURL, model)
	default:
		return nil, ErrInvalidClientType
	}
}

var ErrInvalidClientType = errors.New("invalid client type")