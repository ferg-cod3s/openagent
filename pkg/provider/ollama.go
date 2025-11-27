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

const defaultOllamaURL = "http://localhost:11434"

// OllamaProvider implements the Provider interface for Ollama.
type OllamaProvider struct {
	config Config
	client *http.Client
}

// NewOllama creates a new Ollama provider.
func NewOllama(cfg Config) *OllamaProvider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultOllamaURL
	}
	timeout := 120
	if cfg.Timeout > 0 {
		timeout = cfg.Timeout
	}
	return &OllamaProvider{
		config: cfg,
		client: &http.Client{Timeout: time.Duration(timeout) * time.Second},
	}
}

// Name returns the provider name.
func (p *OllamaProvider) Name() string {
	return string(Ollama)
}

type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Options  *ollamaOptions  `json:"options,omitempty"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaOptions struct {
	Temperature float64  `json:"temperature,omitempty"`
	TopP        float64  `json:"top_p,omitempty"`
	NumPredict  int      `json:"num_predict,omitempty"`
	Stop        []string `json:"stop,omitempty"`
}

type ollamaResponse struct {
	Model              string        `json:"model"`
	CreatedAt          string        `json:"created_at"`
	Message            ollamaMessage `json:"message"`
	Done               bool          `json:"done"`
	TotalDuration      int64         `json:"total_duration"`
	LoadDuration       int64         `json:"load_duration"`
	PromptEvalCount    int           `json:"prompt_eval_count"`
	PromptEvalDuration int64         `json:"prompt_eval_duration"`
	EvalCount          int           `json:"eval_count"`
	EvalDuration       int64         `json:"eval_duration"`
}

// Complete sends a completion request to Ollama.
func (p *OllamaProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	model := req.Model
	if model == "" {
		model = p.config.Model
	}
	if model == "" {
		model = "llama2"
	}

	messages := make([]ollamaMessage, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = ollamaMessage{Role: m.Role, Content: m.Content}
	}

	ollamaReq := ollamaRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
	}

	if req.Temperature > 0 || req.TopP > 0 || req.MaxTokens > 0 || len(req.Stop) > 0 {
		ollamaReq.Options = &ollamaOptions{
			Temperature: req.Temperature,
			TopP:        req.TopP,
			NumPredict:  req.MaxTokens,
			Stop:        req.Stop,
		}
	}

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.config.BaseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
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
		return nil, fmt.Errorf("ollama error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var ollamaResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	result := &CompletionResponse{
		Content: ollamaResp.Message.Content,
		Model:   ollamaResp.Model,
		Usage: Usage{
			PromptTokens:     ollamaResp.PromptEvalCount,
			CompletionTokens: ollamaResp.EvalCount,
			TotalTokens:      ollamaResp.PromptEvalCount + ollamaResp.EvalCount,
		},
		Choices: []Choice{
			{
				Index:        0,
				Message:      Message{Role: ollamaResp.Message.Role, Content: ollamaResp.Message.Content},
				FinishReason: "stop",
			},
		},
	}

	return result, nil
}

// Stream sends a streaming completion request to Ollama.
func (p *OllamaProvider) Stream(ctx context.Context, req *CompletionRequest, handler StreamHandler) error {
	model := req.Model
	if model == "" {
		model = p.config.Model
	}
	if model == "" {
		model = "llama2"
	}

	messages := make([]ollamaMessage, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = ollamaMessage{Role: m.Role, Content: m.Content}
	}

	ollamaReq := ollamaRequest{
		Model:    model,
		Messages: messages,
		Stream:   true,
	}

	if req.Temperature > 0 || req.TopP > 0 || req.MaxTokens > 0 || len(req.Stop) > 0 {
		ollamaReq.Options = &ollamaOptions{
			Temperature: req.Temperature,
			TopP:        req.TopP,
			NumPredict:  req.MaxTokens,
			Stop:        req.Stop,
		}
	}

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.config.BaseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
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
		return fmt.Errorf("ollama error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return p.handleStreamResponse(resp.Body, handler)
}

func (p *OllamaProvider) handleStreamResponse(body io.Reader, handler StreamHandler) error {
	decoder := json.NewDecoder(body)
	for {
		var chunk ollamaResponse
		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("decode stream chunk: %w", err)
		}

		if err := handler(&StreamChunk{
			Content: chunk.Message.Content,
			Done:    chunk.Done,
		}); err != nil {
			return err
		}

		if chunk.Done {
			return nil
		}
	}
}

// Models returns available Ollama models.
func (p *OllamaProvider) Models(ctx context.Context) ([]string, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, p.config.BaseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var modelsResp struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	models := make([]string, len(modelsResp.Models))
	for i, m := range modelsResp.Models {
		models[i] = m.Name
	}

	return models, nil
}
