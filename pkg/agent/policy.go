package agent

import (
	"context"
	"fmt"
)

// DefaultPolicy implements a basic policy with configurable rules.
type DefaultPolicy struct {
	AllowedActions  map[string]bool
	MaxTokensPerRun int
	MaxRuns         int
	runCount        int
}

// NewDefaultPolicy creates a new default policy.
func NewDefaultPolicy() *DefaultPolicy {
	return &DefaultPolicy{
		AllowedActions:  make(map[string]bool),
		MaxTokensPerRun: 100000,
		MaxRuns:         1000,
	}
}

// AllowAction marks an action type as allowed.
func (p *DefaultPolicy) AllowAction(actionType string) {
	p.AllowedActions[actionType] = true
}

// DenyAction marks an action type as denied.
func (p *DefaultPolicy) DenyAction(actionType string) {
	p.AllowedActions[actionType] = false
}

// Validate checks if an action is allowed.
func (p *DefaultPolicy) Validate(ctx context.Context, action Action) error {
	if allowed, exists := p.AllowedActions[action.Type]; exists && !allowed {
		return fmt.Errorf("action %q is not allowed by policy", action.Type)
	}
	return nil
}

// OnError handles errors according to policy.
func (p *DefaultPolicy) OnError(ctx context.Context, err error) error {
	return err
}

// RestrictivePolicy implements a strict policy that denies by default.
type RestrictivePolicy struct {
	AllowedActions map[string]bool
}

// NewRestrictivePolicy creates a new restrictive policy.
func NewRestrictivePolicy() *RestrictivePolicy {
	return &RestrictivePolicy{
		AllowedActions: make(map[string]bool),
	}
}

// AllowAction marks an action type as allowed.
func (p *RestrictivePolicy) AllowAction(actionType string) {
	p.AllowedActions[actionType] = true
}

// Validate checks if an action is allowed. Denies by default.
func (p *RestrictivePolicy) Validate(ctx context.Context, action Action) error {
	if allowed, exists := p.AllowedActions[action.Type]; !exists || !allowed {
		return fmt.Errorf("action %q is not allowed by restrictive policy", action.Type)
	}
	return nil
}

// OnError handles errors according to policy.
func (p *RestrictivePolicy) OnError(ctx context.Context, err error) error {
	return fmt.Errorf("policy violation: %w", err)
}
