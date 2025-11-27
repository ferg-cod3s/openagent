package agent

import (
	"context"
	"testing"
	"time"

	"github.com/ferg-cod3s/openagent/pkg/provider"
)

// mockProvider is a test mock for the Provider interface.
type mockProvider struct {
	name     string
	response *provider.CompletionResponse
	err      error
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Complete(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func (m *mockProvider) Stream(ctx context.Context, req *provider.CompletionRequest, handler provider.StreamHandler) error {
	return nil
}

func (m *mockProvider) Models(ctx context.Context) ([]string, error) {
	return []string{"test-model"}, nil
}

func TestNewAgent(t *testing.T) {
	p := &mockProvider{name: "test"}
	cfg := Config{
		ID:   "test-agent",
		Name: "Test Agent",
	}

	a := New(cfg, p)
	if a == nil {
		t.Fatal("expected agent, got nil")
	}
	if a.ID() != "test-agent" {
		t.Errorf("expected ID 'test-agent', got %q", a.ID())
	}
	if a.Name() != "Test Agent" {
		t.Errorf("expected Name 'Test Agent', got %q", a.Name())
	}
	if a.State() != StateIdle {
		t.Errorf("expected state Idle, got %s", a.State())
	}
}

func TestAgentRun(t *testing.T) {
	p := &mockProvider{
		name: "test",
		response: &provider.CompletionResponse{
			Content: "Hello, world!",
			Usage: provider.Usage{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		},
	}
	cfg := Config{
		ID:      "test-agent",
		Name:    "Test Agent",
		Timeout: 5 * time.Second,
	}

	a := New(cfg, p)

	result, err := a.Run(context.Background(), "Hi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("expected success")
	}
	if result.Output != "Hello, world!" {
		t.Errorf("expected output 'Hello, world!', got %q", result.Output)
	}
	if result.Usage.TotalTokens != 15 {
		t.Errorf("expected 15 total tokens, got %d", result.Usage.TotalTokens)
	}

	history := a.History()
	if len(history) != 2 {
		t.Errorf("expected 2 messages in history, got %d", len(history))
	}
}

func TestAgentState(t *testing.T) {
	p := &mockProvider{name: "test", response: &provider.CompletionResponse{Content: "ok"}}
	a := New(Config{ID: "test"}, p)

	if a.State() != StateIdle {
		t.Errorf("expected Idle, got %s", a.State())
	}

	a.Pause()
	if a.State() != StateIdle {
		t.Errorf("expected Idle after pause (not running), got %s", a.State())
	}

	a.Stop()
	if a.State() != StateStopped {
		t.Errorf("expected Stopped, got %s", a.State())
	}
}

func TestAgentHistory(t *testing.T) {
	p := &mockProvider{name: "test", response: &provider.CompletionResponse{Content: "ok"}}
	a := New(Config{ID: "test"}, p)

	_, _ = a.Run(context.Background(), "test")
	if len(a.History()) != 2 {
		t.Errorf("expected 2 messages, got %d", len(a.History()))
	}

	a.ClearHistory()
	if len(a.History()) != 0 {
		t.Errorf("expected 0 messages after clear, got %d", len(a.History()))
	}
}

func TestDefaultPolicy(t *testing.T) {
	p := NewDefaultPolicy()

	p.AllowAction("read")
	p.DenyAction("write")

	err := p.Validate(context.Background(), Action{Type: "read"})
	if err != nil {
		t.Errorf("expected nil error for allowed action, got %v", err)
	}

	err = p.Validate(context.Background(), Action{Type: "write"})
	if err == nil {
		t.Error("expected error for denied action")
	}

	err = p.Validate(context.Background(), Action{Type: "other"})
	if err != nil {
		t.Errorf("expected nil error for unknown action in default policy, got %v", err)
	}
}

func TestRestrictivePolicy(t *testing.T) {
	p := NewRestrictivePolicy()

	p.AllowAction("read")

	err := p.Validate(context.Background(), Action{Type: "read"})
	if err != nil {
		t.Errorf("expected nil error for allowed action, got %v", err)
	}

	err = p.Validate(context.Background(), Action{Type: "write"})
	if err == nil {
		t.Error("expected error for non-allowed action")
	}
}
