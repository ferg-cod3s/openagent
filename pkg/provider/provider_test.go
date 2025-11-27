package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewOpenAI(t *testing.T) {
	cfg := Config{
		APIKey: "test-key",
	}
	p := NewOpenAI(cfg)
	if p == nil {
		t.Fatal("expected provider, got nil")
	}
	if p.Name() != "openai" {
		t.Errorf("expected name 'openai', got %q", p.Name())
	}
}

func TestOpenAIComplete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Errorf("expected /chat/completions, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Authorization header")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "chatcmpl-123",
			"object": "chat.completion",
			"created": 1677652288,
			"model": "gpt-4",
			"choices": [{
				"index": 0,
				"message": {"role": "assistant", "content": "Hello!"},
				"finish_reason": "stop"
			}],
			"usage": {"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15}
		}`))
	}))
	defer server.Close()

	p := NewOpenAI(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	resp, err := p.Complete(context.Background(), &CompletionRequest{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "Hello!" {
		t.Errorf("expected content 'Hello!', got %q", resp.Content)
	}
	if resp.Usage.TotalTokens != 15 {
		t.Errorf("expected 15 total tokens, got %d", resp.Usage.TotalTokens)
	}
}

func TestNewAnthropic(t *testing.T) {
	cfg := Config{
		APIKey: "test-key",
	}
	p := NewAnthropic(cfg)
	if p == nil {
		t.Fatal("expected provider, got nil")
	}
	if p.Name() != "anthropic" {
		t.Errorf("expected name 'anthropic', got %q", p.Name())
	}
}

func TestNewOllama(t *testing.T) {
	p := NewOllama(Config{})
	if p == nil {
		t.Fatal("expected provider, got nil")
	}
	if p.Name() != "ollama" {
		t.Errorf("expected name 'ollama', got %q", p.Name())
	}
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()
	p := NewOpenAI(Config{APIKey: "test"})
	r.Register(OpenAI, p)

	got, err := r.Get(OpenAI)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name() != "openai" {
		t.Errorf("expected 'openai', got %q", got.Name())
	}

	_, err = r.Get(Anthropic)
	if err == nil {
		t.Error("expected error for unregistered provider")
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		ptype    ProviderType
		expected string
	}{
		{OpenAI, "openai"},
		{Anthropic, "anthropic"},
		{Ollama, "ollama"},
	}

	for _, tt := range tests {
		p, err := New(tt.ptype, Config{APIKey: "test"})
		if err != nil {
			t.Errorf("unexpected error for %s: %v", tt.ptype, err)
			continue
		}
		if p.Name() != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, p.Name())
		}
	}

	_, err := New("unknown", Config{})
	if err == nil {
		t.Error("expected error for unknown provider")
	}
}
