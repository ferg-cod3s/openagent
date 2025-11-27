// Package provider implements LLM provider abstractions for multiple backends.
package provider

import (
	"context"
)

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionRequest contains parameters for a completion request.
type CompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	Stop        []string  `json:"stop,omitempty"`
}

// CompletionResponse contains the response from a completion request.
type CompletionResponse struct {
	ID      string   `json:"id"`
	Content string   `json:"content"`
	Model   string   `json:"model"`
	Usage   Usage    `json:"usage"`
	Choices []Choice `json:"choices,omitempty"`
}

// Usage contains token usage information.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Choice represents a completion choice.
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Provider defines the interface for LLM providers.
type Provider interface {
	// Name returns the provider name.
	Name() string

	// Complete sends a completion request and returns the response.
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

	// Stream sends a completion request and streams the response.
	Stream(ctx context.Context, req *CompletionRequest, handler StreamHandler) error

	// Models returns the list of available models.
	Models(ctx context.Context) ([]string, error)
}

// StreamHandler handles streaming completion responses.
type StreamHandler func(chunk *StreamChunk) error

// StreamChunk represents a chunk of streamed response.
type StreamChunk struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Done    bool   `json:"done"`
}

// Config contains common provider configuration.
type Config struct {
	APIKey      string `json:"api_key"`
	BaseURL     string `json:"base_url,omitempty"`
	Model       string `json:"model,omitempty"`
	MaxRetries  int    `json:"max_retries,omitempty"`
	Timeout     int    `json:"timeout,omitempty"`
	HTTPHeaders map[string]string `json:"http_headers,omitempty"`
}

// ProviderType represents the type of LLM provider.
type ProviderType string

const (
	OpenAI    ProviderType = "openai"
	Anthropic ProviderType = "anthropic"
	Ollama    ProviderType = "ollama"
)
