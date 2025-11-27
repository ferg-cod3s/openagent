# OpenAgent Development TODO

This document outlines the implementation plan for completing OpenAgent's core features. Each task is designed to be atomic and independently committable.

## Phase 1: CLI Integration (7 tasks)

**Goal:** Make the CLI functional with existing packages

### 1. Wire up provider test command
- Connect `provider test` command to actual provider implementations
- Test OpenAI, Anthropic, and Ollama connections
- Validate API keys and make test API calls
- Return meaningful error messages

**Files:** `cmd/openagent/cmd/root.go`

### 2. Add config file loading
- Implement ~/.openagent.yaml configuration loading
- Support provider API keys and default settings
- Add validation for config structure
- Handle missing config gracefully

**Files:** `cmd/openagent/cmd/root.go`, new `pkg/config/config.go`

### 3. Implement agent create command
- Save agent configurations to disk (~/.openagent/agents/)
- Support configurable system prompts and provider selection
- Validate agent configuration before saving
- Generate unique agent IDs

**Files:** `cmd/openagent/cmd/root.go`, new `pkg/store/agent_store.go`

### 4. Implement agent list command
- Read saved agents from disk
- Display agent details (ID, name, provider, created date)
- Support filtering and sorting options
- Handle empty agent directory

**Files:** `cmd/openagent/cmd/root.go`

### 5. Implement agent run command
- Load agent configuration from disk
- Initialize provider and agent runtime
- Support interactive chat mode
- Save conversation history

**Files:** `cmd/openagent/cmd/root.go`

### 6. Wire up workflow validate command
- Use workflow.Parser to validate YAML files
- Check for syntax errors and structural issues
- Validate step types and required fields
- Provide detailed error messages

**Files:** `cmd/openagent/cmd/root.go`

### 7. Wire up run command
- Parse and execute workflow YAML files
- Register standard actions (http, llm, file, etc.)
- Display execution progress and results
- Handle workflow errors gracefully

**Files:** `cmd/openagent/cmd/root.go`, new `pkg/workflow/actions.go`

---

## Phase 2: Evolution Engine (2 tasks)

**Goal:** Complete genetic algorithm implementation

### 8. Implement evolution engine
- Create DefaultEngine implementing Engine interface
- Build main evolution loop (Initialize, Evolve, Run)
- Integrate existing operators (mutation, selection, crossover)
- Track generations and best individuals
- Support configurable stopping criteria

**Files:** `pkg/evolution/engine.go`, `pkg/evolution/engine_test.go`

### 9. Add fitness evaluator
- Implement DefaultFitnessEvaluator
- Support custom fitness functions
- Add parallel fitness evaluation for populations
- Include sample fitness functions for coding agents

**Files:** `pkg/evolution/fitness.go`, `pkg/evolution/fitness_test.go`

---

## Phase 3: Advanced Workflows (3 tasks)

**Goal:** Add parallel execution and control flow

### 10. Implement parallel workflow execution
- Execute StepTypeParallel steps concurrently using goroutines
- Implement proper synchronization and error collection
- Support timeout for parallel groups
- Merge results from parallel steps

**Files:** `pkg/workflow/engine.go`, `pkg/workflow/workflow_test.go`

### 11. Implement conditional step evaluation
- Evaluate Step.If conditions before execution
- Support common comparison operators (==, !=, <, >, etc.)
- Access workflow context and previous step outputs
- Skip steps when conditions are false

**Files:** `pkg/workflow/engine.go`, new `pkg/workflow/condition.go`

### 12. Implement workflow step dependencies
- Build DAG from Step.DependsOn fields
- Topologically sort steps for execution order
- Detect circular dependencies
- Execute steps when all dependencies complete

**Files:** `pkg/workflow/engine.go`, new `pkg/workflow/dag.go`

---

## Phase 4: Vector/Semantic Search (3 tasks)

**Goal:** Enable semantic memory capabilities

### 13. Add OpenAI embeddings provider
- Implement Embedder interface using OpenAI API
- Support text-embedding-3-small and text-embedding-3-large models
- Add batch embedding support
- Handle rate limiting and retries

**Files:** `pkg/provider/embedder.go`, `pkg/provider/embedder_test.go`

### 14. Implement in-memory vector store
- Create InMemoryVectorStore implementing VectorStore interface
- Implement cosine similarity search
- Support k-nearest neighbor queries
- Add filtering by metadata
- Optimize with indexing for larger datasets

**Files:** `pkg/memory/vector_store.go`, `pkg/memory/vector_store_test.go`

### 15. Integrate vector search into agent memory
- Auto-generate embeddings when saving semantic memories
- Implement SearchByText for natural language queries
- Add similarity threshold configuration
- Update agent to use semantic search for context retrieval

**Files:** `pkg/memory/memory.go`, `pkg/agent/agent.go`

---

## Summary

**Total Tasks:** 15
**Estimated Lines of Code:** ~2,500 (roughly doubling current codebase)

### Task Dependencies
- Phase 1 tasks are independent and can be done in any order
- Phase 2 requires no dependencies (uses existing operators)
- Phase 3 builds on task 7 (sequential workflow execution)
- Phase 4 tasks should be done in order: 13 → 14 → 15

### Testing Strategy
Each task should include:
- Unit tests for new functionality
- Integration tests where applicable
- Manual CLI testing for user-facing features
- Update existing tests if behavior changes

### Documentation Updates
After completing each phase:
- Update README.md examples
- Add godoc comments for new public APIs
- Create example workflows for new features
- Update CLI help text

---

## Future Enhancements (Not in Current Plan)

- Persistent database backends (PostgreSQL, SQLite)
- External vector databases (Pinecone, Weaviate, Milvus)
- Sandbox implementation with actual isolation
- Policy enforcement in agent runtime
- Workflow retry logic and error handlers
- Multi-agent collaboration
- Observability and metrics
- Web UI for agent management
