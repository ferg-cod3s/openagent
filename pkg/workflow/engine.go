package workflow

import (
	"context"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// DefaultEngine implements the Engine interface.
type DefaultEngine struct {
	actions map[string]ActionHandler
}

// NewEngine creates a new workflow engine.
func NewEngine() *DefaultEngine {
	return &DefaultEngine{
		actions: make(map[string]ActionHandler),
	}
}

// RegisterAction registers an action handler.
func (e *DefaultEngine) RegisterAction(name string, handler ActionHandler) {
	e.actions[name] = handler
}

// Execute runs a workflow.
func (e *DefaultEngine) Execute(ctx context.Context, w *Workflow) (*WorkflowResult, error) {
	start := time.Now()
	result := &WorkflowResult{
		WorkflowName: w.Name,
		Status:       StatusRunning,
		Steps:        make([]*StepResult, 0, len(w.Steps)),
		StartTime:    start,
	}

	// Apply workflow timeout
	if w.Timeout != "" {
		d, err := time.ParseDuration(w.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout: %w", err)
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, d)
		defer cancel()
	}

	// Execute steps sequentially for now
	for _, step := range w.Steps {
		stepResult, err := e.ExecuteStep(ctx, &step, nil)
		result.Steps = append(result.Steps, stepResult)

		if err != nil {
			result.Status = StatusFailed
			result.Error = err
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(result.StartTime)
			return result, err
		}
	}

	result.Status = StatusCompleted
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	return result, nil
}

// ExecuteStep runs a single step.
func (e *DefaultEngine) ExecuteStep(ctx context.Context, step *Step, inputs map[string]interface{}) (*StepResult, error) {
	start := time.Now()
	result := &StepResult{
		StepID:    step.ID,
		Status:    StatusRunning,
		StartTime: start,
	}

	// Apply step timeout
	if step.Timeout != "" {
		d, err := time.ParseDuration(step.Timeout)
		if err != nil {
			result.Status = StatusFailed
			result.Error = fmt.Errorf("invalid timeout: %w", err)
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(result.StartTime)
			return result, result.Error
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, d)
		defer cancel()
	}

	// Merge inputs with step.With
	allInputs := make(map[string]interface{})
	for k, v := range inputs {
		allInputs[k] = v
	}
	for k, v := range step.With {
		allInputs[k] = v
	}

	// Execute action
	handler, ok := e.actions[step.Action]
	if !ok {
		result.Status = StatusFailed
		result.Error = fmt.Errorf("unknown action: %s", step.Action)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, result.Error
	}

	output, err := handler(ctx, allInputs)
	if err != nil {
		result.Status = StatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	result.Status = StatusCompleted
	result.Output = output
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	return result, nil
}

// YAMLParser implements the Parser interface.
type YAMLParser struct{}

// NewParser creates a new YAML parser.
func NewParser() *YAMLParser {
	return &YAMLParser{}
}

// Parse parses a workflow from YAML bytes.
func (p *YAMLParser) Parse(data []byte) (*Workflow, error) {
	var w Workflow
	if err := yaml.Unmarshal(data, &w); err != nil {
		return nil, fmt.Errorf("parse workflow: %w", err)
	}
	return &w, nil
}

// ParseFile parses a workflow from a file.
func (p *YAMLParser) ParseFile(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return p.Parse(data)
}

// DefaultValidator implements the Validator interface.
type DefaultValidator struct{}

// NewValidator creates a new validator.
func NewValidator() *DefaultValidator {
	return &DefaultValidator{}
}

// Validate checks if a workflow is valid.
func (v *DefaultValidator) Validate(w *Workflow) error {
	if w.Name == "" {
		return &ValidationError{Field: "name", Message: "required"}
	}
	if len(w.Steps) == 0 {
		return &ValidationError{Field: "steps", Message: "at least one step required"}
	}
	for i, step := range w.Steps {
		if step.ID == "" {
			return &ValidationError{Field: fmt.Sprintf("steps[%d].id", i), Message: "required"}
		}
		if step.Name == "" {
			return &ValidationError{Field: fmt.Sprintf("steps[%d].name", i), Message: "required"}
		}
	}
	return nil
}
