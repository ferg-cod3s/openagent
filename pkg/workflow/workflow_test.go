package workflow

import (
	"context"
	"testing"
)

func TestYAMLParser(t *testing.T) {
	p := NewParser()

	yaml := `
name: test-workflow
version: "1.0"
description: A test workflow
steps:
  - id: step1
    name: First Step
    type: task
    action: test
`

	w, err := p.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Name != "test-workflow" {
		t.Errorf("expected name 'test-workflow', got %q", w.Name)
	}
	if w.Version != "1.0" {
		t.Errorf("expected version '1.0', got %q", w.Version)
	}
	if len(w.Steps) != 1 {
		t.Errorf("expected 1 step, got %d", len(w.Steps))
	}
	if w.Steps[0].ID != "step1" {
		t.Errorf("expected step ID 'step1', got %q", w.Steps[0].ID)
	}
}

func TestDefaultValidator(t *testing.T) {
	v := NewValidator()

	// Valid workflow
	w := &Workflow{
		Name: "test",
		Steps: []Step{
			{ID: "s1", Name: "Step 1"},
		},
	}
	if err := v.Validate(w); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Missing name
	w = &Workflow{Steps: []Step{{ID: "s1", Name: "Step 1"}}}
	if err := v.Validate(w); err == nil {
		t.Error("expected error for missing name")
	}

	// No steps
	w = &Workflow{Name: "test", Steps: []Step{}}
	if err := v.Validate(w); err == nil {
		t.Error("expected error for no steps")
	}

	// Missing step ID
	w = &Workflow{Name: "test", Steps: []Step{{Name: "Step 1"}}}
	if err := v.Validate(w); err == nil {
		t.Error("expected error for missing step ID")
	}

	// Missing step name
	w = &Workflow{Name: "test", Steps: []Step{{ID: "s1"}}}
	if err := v.Validate(w); err == nil {
		t.Error("expected error for missing step name")
	}
}

func TestDefaultEngine(t *testing.T) {
	e := NewEngine()

	// Register action
	e.RegisterAction("echo", func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"echo": inputs["message"]}, nil
	})

	// Execute step
	step := &Step{
		ID:     "s1",
		Name:   "Echo",
		Action: "echo",
		With:   map[string]any{"message": "hello"},
	}

	result, err := e.ExecuteStep(context.Background(), step, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != StatusCompleted {
		t.Errorf("expected status Completed, got %s", result.Status)
	}

	if result.Output["echo"] != "hello" {
		t.Errorf("expected echo 'hello', got %v", result.Output["echo"])
	}
}

func TestEngineExecuteWorkflow(t *testing.T) {
	e := NewEngine()

	e.RegisterAction("noop", func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"done": true}, nil
	})

	w := &Workflow{
		Name: "test",
		Steps: []Step{
			{ID: "s1", Name: "Step 1", Action: "noop"},
			{ID: "s2", Name: "Step 2", Action: "noop"},
		},
	}

	result, err := e.Execute(context.Background(), w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != StatusCompleted {
		t.Errorf("expected status Completed, got %s", result.Status)
	}

	if len(result.Steps) != 2 {
		t.Errorf("expected 2 step results, got %d", len(result.Steps))
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{Field: "name", Message: "required"}
	if err.Error() != "name: required" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}
