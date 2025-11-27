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

const defaultAnthropicURL = "https://api.anthropic.com/v1"

// AnthropicProvider implements the Provider interface for Anthropic.
type AnthropicProvider struct {
	config Config
	client *http.Client
}

// NewAnthropic creates a new Anthropic provider.
func NewAnthropic(cfg Config) *AnthropicProvider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultAnthropicURL
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	timeout := 30
	if cfg.Timeout > 0 {
		timeout = cfg.Timeout
	}
	return &AnthropicProvider{
		config: cfg,
		client: &http.Client{Timeout: time.Duration(timeout) * time.Second},
	}
}

// Name returns the provider name.
func (p *AnthropicProvider) Name() string {
	return string(Anthropic)
}

type anthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []anthropicMessage `json:"messages"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float64            `json:"temperature,omitempty"`
	TopP        float64            `json:"top_p,omitempty"`
	StopSeqs    []string           `json:"stop_sequences,omitempty"`
	Stream      bool               `json:"stream,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// Complete sends a completion request to Anthropic.
func (p *AnthropicProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	model := req.Model
	if model == "" {
		model = p.config.Model
	}
	if model == "" {
		model = "claude-3-opus-20240229"
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	messages := make([]anthropicMessage, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = anthropicMessage{Role: m.Role, Content: m.Content}
	}

	antReq := anthropicRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		StopSeqs:    req.Stop,
	}

	body, err := json.Marshal(antReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.config.BaseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
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
		return nil, fmt.Errorf("anthropic error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var antResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&antResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	content := ""
	if len(antResp.Content) > 0 {
		content = antResp.Content[0].Text
	}

	result := &CompletionResponse{
		ID:      antResp.ID,
		Content: content,
		Model:   antResp.Model,
		Usage: Usage{
			PromptTokens:     antResp.Usage.InputTokens,
			CompletionTokens: antResp.Usage.OutputTokens,
			TotalTokens:      antResp.Usage.InputTokens + antResp.Usage.OutputTokens,
		},
		Choices: []Choice{
			{
				Index:        0,
				Message:      Message{Role: antResp.Role, Content: content},
				FinishReason: antResp.StopReason,
			},
		},
	}

	return result, nil
}

// Stream sends a streaming completion request to Anthropic.
func (p *AnthropicProvider) Stream(ctx context.Context, req *CompletionRequest, handler StreamHandler) error {
	model := req.Model
	if model == "" {
		model = p.config.Model
	}
	if model == "" {
		model = "claude-3-opus-20240229"
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	messages := make([]anthropicMessage, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = anthropicMessage{Role: m.Role, Content: m.Content}
	}

	antReq := anthropicRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		StopSeqs:    req.Stop,
		Stream:      true,
	}

	body, err := json.Marshal(antReq)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.config.BaseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
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
		return fmt.Errorf("anthropic error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return p.handleStreamResponse(resp.Body, handler)
}

func (p *AnthropicProvider) handleStreamResponse(body io.Reader, handler StreamHandler) error {
	decoder := json.NewDecoder(body)
	for {
		var event struct {
			Type  string `json:"type"`
			Delta struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"delta"`
		}

		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("decode stream event: %w", err)
		}

		done := event.Type == "message_stop"
		content := ""
		if event.Type == "content_block_delta" {
			content = event.Delta.Text
		}

		if err := handler(&StreamChunk{
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

// Models returns available Anthropic models.
func (p *AnthropicProvider) Models(ctx context.Context) ([]string, error) {
	return []string{
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
		"claude-2.1",
		"claude-2.0",
		"claude-instant-1.2",
	}, nil
}
