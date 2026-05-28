package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type LocalClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func NewLocalClient(apiKey, baseURL, model string) (Client, error) {
	if baseURL == "" {
		baseURL = "http://192.168.101.127:4004/"
	}
	if model == "" {
		model = "kimi-k2.6"
	}
	return &LocalClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{Timeout: 60 * time.Second}, // 60秒超时
	}, nil
}

func (c *LocalClient) ChatCompletion(req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if req.Model == "" {
		req.Model = c.model
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// 尝试多个可能的端点路径
	endpoints := []string{
		"%s/v1/chat/completions",
		"%s/chat/completions",
		"%sv1/chat/completions",
		"%s/v1/chat/completions/",
		"%sapi/chat/completions",
	}

	var lastErr error
	for _, endpoint := range endpoints {
		url := fmt.Sprintf(endpoint, c.baseURL)
		request, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
		if err != nil {
			lastErr = err
			continue
		}

		request.Header.Set("Content-Type", "application/json")
		if c.apiKey != "" {
			request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
		}

		resp, err := c.client.Do(request)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			lastErr = fmt.Errorf("endpoint not found: %s", url)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("API request failed (status %d): %s", resp.StatusCode, string(body))
			continue
		}

		var response ChatCompletionResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			lastErr = err
			continue
		}

		return &response, nil
	}

	return nil, fmt.Errorf("all endpoints failed, last error: %w", lastErr)
}

func (c *LocalClient) ChatCompletionStream(req ChatCompletionRequest, callback func(string)) error {
	if req.Model == "" {
		req.Model = c.model
	}
	req.Stream = true

	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	// 尝试多个可能的端点路径
	endpoints := []string{
		"%s/v1/chat/completions",
		"%s/chat/completions",
		"%sv1/chat/completions",
		"%s/v1/chat/completions/",
		"%sapi/chat/completions",
	}

	var lastErr error
	for _, endpoint := range endpoints {
		url := fmt.Sprintf(endpoint, c.baseURL)
		request, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
		if err != nil {
			lastErr = err
			continue
		}

		request.Header.Set("Content-Type", "application/json")
		if c.apiKey != "" {
			request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
		}

		resp, err := c.client.Do(request)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			lastErr = fmt.Errorf("endpoint not found: %s", url)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("API request failed (status %d): %s", resp.StatusCode, string(body))
			continue
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
		return nil
	}

	return fmt.Errorf("all endpoints failed, last error: %w", lastErr)
}

func (c *LocalClient) SetAPIKey(key string) {
	c.apiKey = key
}

func (c *LocalClient) SetBaseURL(url string) {
	c.baseURL = url
}

func (c *LocalClient) GetModel() string {
	return c.model
}

func (c *LocalClient) ListModels() ([]string, error) {
	return []string{
		"kimi-k2.6",
		"local-model",
		"llama-3",
		"mistral",
		"phi-3",
	}, nil
}