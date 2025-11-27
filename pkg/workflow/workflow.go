// Package workflow implements a YAML-based workflow engine.
package workflow

import (
	"context"
	"fmt"
	"time"
)

// Workflow represents a complete workflow definition.
type Workflow struct {
	Name        string            `yaml:"name" json:"name"`
	Version     string            `yaml:"version" json:"version"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
	Env         map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	Steps       []Step            `yaml:"steps" json:"steps"`
	OnError     *ErrorHandler     `yaml:"on_error,omitempty" json:"on_error,omitempty"`
	Timeout     string            `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// Step represents a single workflow step.
type Step struct {
	ID        string            `yaml:"id" json:"id"`
	Name      string            `yaml:"name" json:"name"`
	Type      StepType          `yaml:"type" json:"type"`
	Action    string            `yaml:"action,omitempty" json:"action,omitempty"`
	With      map[string]any    `yaml:"with,omitempty" json:"with,omitempty"`
	Env       map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	If        string            `yaml:"if,omitempty" json:"if,omitempty"`
	DependsOn []string          `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
	Timeout   string            `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Retries   int               `yaml:"retries,omitempty" json:"retries,omitempty"`
	OnError   *ErrorHandler     `yaml:"on_error,omitempty" json:"on_error,omitempty"`
}

// StepType represents the type of step.
type StepType string

const (
	StepTypeAgent    StepType = "agent"
	StepTypeTask     StepType = "task"
	StepTypeParallel StepType = "parallel"
	StepTypeSequence StepType = "sequence"
	StepTypeDecision StepType = "decision"
	StepTypeLoop     StepType = "loop"
)

// ErrorHandler defines error handling behavior.
type ErrorHandler struct {
	Action  string `yaml:"action" json:"action"`
	Message string `yaml:"message,omitempty" json:"message,omitempty"`
}

// StepResult contains the result of a step execution.
type StepResult struct {
	StepID    string                 `json:"step_id"`
	Status    StepStatus             `json:"status"`
	Output    map[string]interface{} `json:"output,omitempty"`
	Error     error                  `json:"-"`
	StartTime time.Time              `json:"start_time"`
	EndTime   time.Time              `json:"end_time"`
	Duration  time.Duration          `json:"duration"`
}

// StepStatus represents the status of a step.
type StepStatus string

const (
	StatusPending   StepStatus = "pending"
	StatusRunning   StepStatus = "running"
	StatusCompleted StepStatus = "completed"
	StatusFailed    StepStatus = "failed"
	StatusSkipped   StepStatus = "skipped"
)

// WorkflowResult contains the result of a workflow execution.
type WorkflowResult struct {
	WorkflowName string        `json:"workflow_name"`
	Status       StepStatus    `json:"status"`
	Steps        []*StepResult `json:"steps"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	Duration     time.Duration `json:"duration"`
	Error        error         `json:"-"`
}

// Engine defines the workflow execution engine interface.
type Engine interface {
	// Execute runs a workflow.
	Execute(ctx context.Context, w *Workflow) (*WorkflowResult, error)
	// ExecuteStep runs a single step.
	ExecuteStep(ctx context.Context, step *Step, inputs map[string]interface{}) (*StepResult, error)
	// RegisterAction registers an action handler.
	RegisterAction(name string, handler ActionHandler)
}

// ActionHandler handles action execution.
type ActionHandler func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error)

// Parser defines the workflow parser interface.
type Parser interface {
	// Parse parses a workflow from YAML bytes.
	Parse(data []byte) (*Workflow, error)
	// ParseFile parses a workflow from a file.
	ParseFile(path string) (*Workflow, error)
}

// Validator validates workflow definitions.
type Validator interface {
	// Validate checks if a workflow is valid.
	Validate(w *Workflow) error
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
