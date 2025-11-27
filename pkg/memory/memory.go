// Package memory implements episodic, vector, and structured storage.
package memory

import (
	"context"
	"time"
)

// Memory represents a single memory item.
type Memory struct {
	ID        string                 `json:"id"`
	Type      MemoryType             `json:"type"`
	Content   string                 `json:"content"`
	Embedding []float64              `json:"embedding,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Score     float64                `json:"score,omitempty"`
}

// MemoryType represents the type of memory.
type MemoryType string

const (
	TypeEpisodic   MemoryType = "episodic"
	TypeSemantic   MemoryType = "semantic"
	TypeProcedural MemoryType = "procedural"
	TypeWorking    MemoryType = "working"
)

// Store defines the interface for memory storage.
type Store interface {
	// Save stores a memory.
	Save(ctx context.Context, m *Memory) error
	// Get retrieves a memory by ID.
	Get(ctx context.Context, id string) (*Memory, error)
	// Delete removes a memory.
	Delete(ctx context.Context, id string) error
	// List returns all memories matching the filter.
	List(ctx context.Context, filter *Filter) ([]*Memory, error)
	// Clear removes all memories.
	Clear(ctx context.Context) error
}

// VectorStore defines the interface for vector-based memory storage.
type VectorStore interface {
	Store
	// Search finds similar memories using vector similarity.
	Search(ctx context.Context, embedding []float64, limit int) ([]*Memory, error)
	// SearchByText finds similar memories using text (generates embedding first).
	SearchByText(ctx context.Context, text string, limit int) ([]*Memory, error)
}

// Filter contains options for filtering memories.
type Filter struct {
	Type      MemoryType             `json:"type,omitempty"`
	Since     *time.Time             `json:"since,omitempty"`
	Until     *time.Time             `json:"until,omitempty"`
	Limit     int                    `json:"limit,omitempty"`
	Offset    int                    `json:"offset,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	OrderBy   string                 `json:"order_by,omitempty"`
	OrderDesc bool                   `json:"order_desc,omitempty"`
}

// Embedder generates embeddings for text.
type Embedder interface {
	// Embed generates an embedding for the given text.
	Embed(ctx context.Context, text string) ([]float64, error)
	// EmbedBatch generates embeddings for multiple texts.
	EmbedBatch(ctx context.Context, texts []string) ([][]float64, error)
	// Dimension returns the embedding dimension.
	Dimension() int
}
