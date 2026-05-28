package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type DeepSeekClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func NewDeepSeekClient(apiKey, baseURL, model string) (Client, error) {
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}
	if model == "" {
		model = "deepseek-chat"
	}
	return &DeepSeekClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}, nil
}

func (c *DeepSeekClient) ChatCompletion(req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if req.Model == "" {
		req.Model = c.model
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

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

func (c *DeepSeekClient) ChatCompletionStream(req ChatCompletionRequest, callback func(string)) error {
	if req.Model == "" {
		req.Model = c.model
	}
	req.Stream = true

	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

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

func (c *DeepSeekClient) SetAPIKey(key string) {
	c.apiKey = key
}

func (c *DeepSeekClient) SetBaseURL(url string) {
	c.baseURL = url
}

func (c *DeepSeekClient) GetModel() string {
	return c.model
}

func (c *DeepSeekClient) ListModels() ([]string, error) {
	return []string{
		"deepseek-chat",
		"deepseek-code",
		"deepseek-r1",
	}, nil
}