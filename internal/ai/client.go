package ai

import "errors"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ToolFunction struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

type ToolDefinition struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

type ChatCompletionRequest struct {
	Model    string           `json:"model"`
	Messages []Message        `json:"messages"`
	Tools    []ToolDefinition `json:"tools,omitempty"`
	Stream   bool             `json:"stream"`
}

type ResponseMessage struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type ResponseChoice struct {
	Index        int             `json:"index"`
	Message      ResponseMessage `json:"message"`
	FinishReason string          `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type DeltaMessage struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type StreamingChoice struct {
	Index        int          `json:"index"`
	Delta        DeltaMessage `json:"delta"`
	FinishReason string       `json:"finish_reason"`
}

type ChatCompletionResponse struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Choices []ResponseChoice `json:"choices"`
	Usage   Usage            `json:"usage"`
}

type StreamingResponse struct {
	ID      string            `json:"id"`
	Object  string            `json:"object"`
	Created int64             `json:"created"`
	Model   string            `json:"model"`
	Choices []StreamingChoice `json:"choices"`
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