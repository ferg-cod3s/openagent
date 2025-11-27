package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultOpenAIURL = "https://api.openai.com/v1"

// OpenAIProvider implements the Provider interface for OpenAI.
type OpenAIProvider struct {
	config Config
	client *http.Client
}

// NewOpenAI creates a new OpenAI provider.
func NewOpenAI(cfg Config) *OpenAIProvider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultOpenAIURL
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	timeout := 30
	if cfg.Timeout > 0 {
		timeout = cfg.Timeout
	}
	return &OpenAIProvider{
		config: cfg,
		client: &http.Client{Timeout: time.Duration(timeout) * time.Second},
	}
}

// Name returns the provider name.
func (p *OpenAIProvider) Name() string {
	return string(OpenAI)
}

type openAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	Stop        []string  `json:"stop,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

type openAIResponse struct {
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

// Complete sends a completion request to OpenAI.
func (p *OpenAIProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	model := req.Model
	if model == "" {
		model = p.config.Model
	}
	if model == "" {
		model = "gpt-4"
	}

	oaiReq := openAIRequest{
		Model:       model,
		Messages:    req.Messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stop:        req.Stop,
	}

	body, err := json.Marshal(oaiReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	for k, v := range p.config.HTTPHeaders {
		httpReq.Header.Set(k, v)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var oaiResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&oaiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	result := &CompletionResponse{
		ID:    oaiResp.ID,
		Model: oaiResp.Model,
		Usage: Usage{
			PromptTokens:     oaiResp.Usage.PromptTokens,
			CompletionTokens: oaiResp.Usage.CompletionTokens,
			TotalTokens:      oaiResp.Usage.TotalTokens,
		},
	}

	for _, choice := range oaiResp.Choices {
		result.Choices = append(result.Choices, Choice{
			Index:        choice.Index,
			Message:      choice.Message,
			FinishReason: choice.FinishReason,
		})
	}

	if len(result.Choices) > 0 {
		result.Content = result.Choices[0].Message.Content
	}

	return result, nil
}

// Stream sends a streaming completion request to OpenAI.
func (p *OpenAIProvider) Stream(ctx context.Context, req *CompletionRequest, handler StreamHandler) error {
	model := req.Model
	if model == "" {
		model = p.config.Model
	}
	if model == "" {
		model = "gpt-4"
	}

	oaiReq := openAIRequest{
		Model:       model,
		Messages:    req.Messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stop:        req.Stop,
		Stream:      true,
	}

	body, err := json.Marshal(oaiReq)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	for k, v := range p.config.HTTPHeaders {
		httpReq.Header.Set(k, v)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("openai error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return p.handleStreamResponse(resp.Body, handler)
}

func (p *OpenAIProvider) handleStreamResponse(body io.Reader, handler StreamHandler) error {
	decoder := json.NewDecoder(body)
	for {
		var chunk struct {
			ID      string `json:"id"`
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
		}

		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("decode stream chunk: %w", err)
		}

		done := false
		content := ""
		if len(chunk.Choices) > 0 {
			content = chunk.Choices[0].Delta.Content
			done = chunk.Choices[0].FinishReason == "stop"
		}

		if err := handler(&StreamChunk{
			ID:      chunk.ID,
			Content: content,
			Done:    done,
		}); err != nil {
			return err
		}

		if done {
			return nil
		}
	}
}

// Models returns available OpenAI models.
func (p *OpenAIProvider) Models(ctx context.Context) ([]string, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, p.config.BaseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var modelsResp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	models := make([]string, len(modelsResp.Data))
	for i, m := range modelsResp.Data {
		models[i] = m.ID
	}

	return models, nil
}
