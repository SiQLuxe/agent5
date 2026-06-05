package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type AnthropicClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func NewAnthropicClient(apiKey, baseURL, model string) (Client, error) {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}
	if model == "" {
		model = "claude-3-opus"
	}
	return &AnthropicClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}, nil
}

func (c *AnthropicClient) ChatCompletion(req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if req.Model == "" {
		req.Model = c.model
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/messages", c.baseURL)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-api-key", c.apiKey)
	request.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed: %s", string(body))
	}

	var response ChatCompletionResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *AnthropicClient) ChatCompletionStream(req ChatCompletionRequest, callback func(string)) error {
	if req.Model == "" {
		req.Model = c.model
	}
	req.Stream = true

	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/messages", c.baseURL)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-api-key", c.apiKey)
	request.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed: %s", string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 6 && line[:6] == "data: " {
			line = line[6:]
			if line == "[DONE]" {
				break
			}
			var streamResp StreamingResponse
			if err := json.Unmarshal([]byte(line), &streamResp); err == nil {
				for _, choice := range streamResp.Choices {
					if choice.Delta.Content != "" {
						callback(choice.Delta.Content)
					}
				}
			}
		}
	}

	return scanner.Err()
}

func (c *AnthropicClient) SetAPIKey(key string) {
	c.apiKey = key
}

func (c *AnthropicClient) SetBaseURL(url string) {
	c.baseURL = url
}

func (c *AnthropicClient) GetModel() string {
	return c.model
}

func (c *AnthropicClient) ListModels() ([]string, error) {
	return []string{
		"claude-3-opus",
		"claude-3-sonnet",
		"claude-3-haiku",
		"claude-2.1",
	}, nil
}