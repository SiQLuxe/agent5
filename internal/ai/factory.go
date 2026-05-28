package ai

import (
	"fmt"

	cfg "github.com/example/agent-tui/internal/data/config"
)

func NewClientFromConfig(c *cfg.Config) (Client, error) {
	if c == nil {
		return nil, fmt.Errorf("nil config")
	}

	clientType := c.DefaultClient
	model := ""

	switch clientType {
	case "openai":
		apiKey := c.Models.OpenAI.APIKey
		if apiKey == "" {
			return nil, fmt.Errorf("openai api key not set")
		}
		baseURL := c.Models.OpenAI.BaseURL
		model = c.Models.OpenAI.DefaultModel
		return NewOpenAIClient(apiKey, baseURL, model)
	case "deepseek":
		apiKey := c.Models.DeepSeek.APIKey
		if apiKey == "" {
			return nil, fmt.Errorf("deepseek api key not set")
		}
		baseURL := c.Models.DeepSeek.BaseURL
		model = c.Models.DeepSeek.DefaultModel
		return NewDeepSeekClient(apiKey, baseURL, model)
	case "local":
		baseURL := c.Models.Local.BaseURL
		model = c.Models.Local.DefaultModel
		return NewLocalClient("", baseURL, model)
	default:
		return nil, fmt.Errorf("unsupported client: %s", clientType)
	}
}
