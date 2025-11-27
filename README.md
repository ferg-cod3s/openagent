# OpenAgent

OpenAgent is an evolution layer for coding agents like OpenCode and Claude Code. It provides provider-agnostic, self-improving agents via continuous evolution.

## Features

- **Provider-agnostic LLM abstraction** - Support for OpenAI, Anthropic, and Ollama
- **Agent runtime** - Autonomous agents with policies and sandboxing
- **Memory systems** - Episodic, vector, and structured storage
- **YAML workflow engine** - Define and execute complex agent workflows
- **Evolution engine** - Mutation, fitness evaluation, and selection for self-improving agents

## Requirements

- Go 1.21+

## Installation

```bash
go install github.com/ferg-cod3s/openagent/cmd/openagent@latest
```

Or build from source:

```bash
git clone https://github.com/ferg-cod3s/openagent.git
cd openagent
go build -o openagent ./cmd/openagent
```

## Quick Start

### CLI Usage

```bash
# Show help
openagent --help

# Show version
openagent version

# List available providers
openagent provider list

# Test provider connection
export OPENAI_API_KEY=your-key
openagent provider test openai

# Run a workflow
openagent run workflow.yaml

# Manage agents
openagent agent list
openagent agent create my-agent
```

### Programmatic Usage

#### Provider

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/ferg-cod3s/openagent/pkg/provider"
)

func main() {
    // Create a provider
    p := provider.NewOpenAI(provider.Config{
        APIKey: "your-api-key",
        Model:  "gpt-4",
    })
    
    // Make a completion request
    resp, err := p.Complete(context.Background(), &provider.CompletionRequest{
        Messages: []provider.Message{
            {Role: "user", Content: "Hello!"},
        },
    })
    if err != nil {
        panic(err)
    }
    
    fmt.Println(resp.Content)
}
```

#### Agent

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/ferg-cod3s/openagent/pkg/agent"
    "github.com/ferg-cod3s/openagent/pkg/provider"
)

func main() {
    // Create a provider
    p := provider.NewOpenAI(provider.Config{
        APIKey: "your-api-key",
    })
    
    // Create an agent
    a := agent.New(agent.Config{
        ID:           "coding-agent",
        Name:         "Coding Assistant",
        SystemPrompt: "You are a helpful coding assistant.",
    }, p)
    
    // Run the agent
    result, err := a.Run(context.Background(), "Write a hello world in Go")
    if err != nil {
        panic(err)
    }
    
    fmt.Println(result.Output)
}
```

#### Workflow

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/ferg-cod3s/openagent/pkg/workflow"
)

func main() {
    // Parse a workflow
    parser := workflow.NewParser()
    w, err := parser.ParseFile("workflow.yaml")
    if err != nil {
        panic(err)
    }
    
    // Create an engine
    engine := workflow.NewEngine()
    
    // Register actions
    engine.RegisterAction("echo", func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
        return inputs, nil
    })
    
    // Execute the workflow
    result, err := engine.Execute(context.Background(), w)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Workflow completed in %v\n", result.Duration)
}
```

## Project Structure

```
openagent/
├── cmd/
│   └── openagent/       # CLI application
│       ├── main.go
│       └── cmd/         # Cobra commands
├── pkg/
│   ├── provider/        # LLM provider abstraction
│   ├── agent/           # Agent runtime and policies
│   ├── memory/          # Memory storage systems
│   ├── evolution/       # Evolution engine
│   └── workflow/        # YAML workflow engine
├── go.mod
├── LICENSE
└── README.md
```

## Core Packages

### pkg/provider

LLM provider abstraction supporting:
- OpenAI (GPT-4, GPT-3.5)
- Anthropic (Claude 3)
- Ollama (local models)

### pkg/agent

Agent runtime with:
- Configurable policies
- Sandbox configuration
- Hooks for extensibility
- Conversation history management

### pkg/memory

Memory storage with:
- Episodic memory
- Semantic memory
- Working memory
- Vector similarity search (interface)

### pkg/evolution

Evolution engine with:
- Random mutation
- Tournament selection
- Single-point crossover
- Configurable fitness evaluation

### pkg/workflow

YAML workflow engine with:
- Sequential and parallel execution
- Conditional steps
- Error handling
- Timeout configuration

## Design Principles

- **SOLID** - Single responsibility, open-closed, Liskov substitution, interface segregation, dependency inversion
- **YAGNI** - You Aren't Gonna Need It - minimal implementation that works
- **Provider-agnostic** - Easy to switch between LLM providers
- **Self-improving** - Agents can evolve and improve via continuous evolution

## License

Apache 2.0 - See [LICENSE](LICENSE) for details