// Package agent implements the agent runtime, policies, and sandbox.
package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ferg-cod3s/openagent/pkg/provider"
)

// State represents the current state of an agent.
type State string

const (
	StateIdle    State = "idle"
	StateRunning State = "running"
	StatePaused  State = "paused"
	StateStopped State = "stopped"
	StateError   State = "error"
)

// Config contains agent configuration.
type Config struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Description  string         `json:"description,omitempty"`
	Model        string         `json:"model,omitempty"`
	Provider     string         `json:"provider,omitempty"`
	MaxTokens    int            `json:"max_tokens,omitempty"`
	Temperature  float64        `json:"temperature,omitempty"`
	SystemPrompt string         `json:"system_prompt,omitempty"`
	Timeout      time.Duration  `json:"timeout,omitempty"`
	Sandbox      *SandboxConfig `json:"sandbox,omitempty"`
}

// SandboxConfig contains sandbox configuration.
type SandboxConfig struct {
	Enabled     bool     `json:"enabled"`
	AllowNet    bool     `json:"allow_net"`
	AllowFS     bool     `json:"allow_fs"`
	AllowedDirs []string `json:"allowed_dirs,omitempty"`
	MaxMemoryMB int      `json:"max_memory_mb,omitempty"`
	MaxCPUPct   int      `json:"max_cpu_pct,omitempty"`
}

// Agent represents an autonomous agent.
type Agent struct {
	mu       sync.RWMutex
	config   Config
	state    State
	provider provider.Provider
	history  []provider.Message
	policy   Policy
	hooks    []Hook
}

// Policy defines constraints and behaviors for an agent.
type Policy interface {
	// Validate checks if an action is allowed.
	Validate(ctx context.Context, action Action) error
	// OnError handles errors according to policy.
	OnError(ctx context.Context, err error) error
}

// Action represents an agent action.
type Action struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

// Hook allows extending agent behavior.
type Hook interface {
	// BeforeRun is called before the agent runs.
	BeforeRun(ctx context.Context, a *Agent) error
	// AfterRun is called after the agent runs.
	AfterRun(ctx context.Context, a *Agent, result *Result) error
	// OnMessage is called when a message is received.
	OnMessage(ctx context.Context, a *Agent, msg *provider.Message) error
}

// Result contains the result of an agent run.
type Result struct {
	Success   bool               `json:"success"`
	Output    string             `json:"output"`
	Messages  []provider.Message `json:"messages,omitempty"`
	Usage     *provider.Usage    `json:"usage,omitempty"`
	Error     error              `json:"-"`
	Duration  time.Duration      `json:"duration"`
	Timestamp time.Time          `json:"timestamp"`
}

// New creates a new agent with the given configuration.
func New(cfg Config, p provider.Provider) *Agent {
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Minute
	}
	return &Agent{
		config:   cfg,
		state:    StateIdle,
		provider: p,
		history:  make([]provider.Message, 0),
	}
}

// ID returns the agent ID.
func (a *Agent) ID() string {
	return a.config.ID
}

// Name returns the agent name.
func (a *Agent) Name() string {
	return a.config.Name
}

// State returns the current agent state.
func (a *Agent) State() State {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.state
}

// SetPolicy sets the agent's policy.
func (a *Agent) SetPolicy(p Policy) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.policy = p
}

// AddHook adds a hook to the agent.
func (a *Agent) AddHook(h Hook) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.hooks = append(a.hooks, h)
}

// History returns the conversation history.
func (a *Agent) History() []provider.Message {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return append([]provider.Message{}, a.history...)
}

// ClearHistory clears the conversation history.
func (a *Agent) ClearHistory() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.history = make([]provider.Message, 0)
}

// Run executes the agent with the given input.
func (a *Agent) Run(ctx context.Context, input string) (*Result, error) {
	a.mu.Lock()
	if a.state == StateRunning {
		a.mu.Unlock()
		return nil, fmt.Errorf("agent is already running")
	}
	a.state = StateRunning
	a.mu.Unlock()

	start := time.Now()
	result := &Result{Timestamp: start}

	defer func() {
		a.mu.Lock()
		if a.state == StateRunning {
			a.state = StateIdle
		}
		a.mu.Unlock()
	}()

	// Execute hooks
	for _, h := range a.hooks {
		if err := h.BeforeRun(ctx, a); err != nil {
			result.Error = err
			return result, err
		}
	}

	// Build messages
	messages := make([]provider.Message, 0, len(a.history)+2)
	if a.config.SystemPrompt != "" {
		messages = append(messages, provider.Message{
			Role:    "system",
			Content: a.config.SystemPrompt,
		})
	}
	messages = append(messages, a.history...)
	messages = append(messages, provider.Message{
		Role:    "user",
		Content: input,
	})

	// Create request
	req := &provider.CompletionRequest{
		Model:       a.config.Model,
		Messages:    messages,
		MaxTokens:   a.config.MaxTokens,
		Temperature: a.config.Temperature,
	}

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, a.config.Timeout)
	defer cancel()

	// Execute completion
	resp, err := a.provider.Complete(ctx, req)
	if err != nil {
		result.Error = err
		a.mu.Lock()
		a.state = StateError
		a.mu.Unlock()
		return result, err
	}

	// Update history
	a.mu.Lock()
	a.history = append(a.history, provider.Message{Role: "user", Content: input})
	if resp.Content != "" {
		a.history = append(a.history, provider.Message{Role: "assistant", Content: resp.Content})
	}
	a.mu.Unlock()

	// Build result
	result.Success = true
	result.Output = resp.Content
	result.Messages = append([]provider.Message{}, a.history...)
	result.Usage = &resp.Usage
	result.Duration = time.Since(start)

	// Execute after hooks
	for _, h := range a.hooks {
		if err := h.AfterRun(ctx, a, result); err != nil {
			return result, err
		}
	}

	return result, nil
}

// Stop stops the agent.
func (a *Agent) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.state = StateStopped
}

// Pause pauses the agent.
func (a *Agent) Pause() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.state == StateRunning {
		a.state = StatePaused
	}
}

// Resume resumes a paused agent.
func (a *Agent) Resume() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.state == StatePaused {
		a.state = StateIdle
	}
}
