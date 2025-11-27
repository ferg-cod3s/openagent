package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// InMemoryStore implements Store using an in-memory map.
type InMemoryStore struct {
	mu       sync.RWMutex
	memories map[string]*Memory
}

// NewInMemoryStore creates a new in-memory store.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		memories: make(map[string]*Memory),
	}
}

// Save stores a memory.
func (s *InMemoryStore) Save(ctx context.Context, m *Memory) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	now := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now

	s.memories[m.ID] = m
	return nil
}

// Get retrieves a memory by ID.
func (s *InMemoryStore) Get(ctx context.Context, id string) (*Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	m, ok := s.memories[id]
	if !ok {
		return nil, fmt.Errorf("memory not found: %s", id)
	}
	return m, nil
}

// Delete removes a memory.
func (s *InMemoryStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.memories[id]; !ok {
		return fmt.Errorf("memory not found: %s", id)
	}
	delete(s.memories, id)
	return nil
}

// List returns all memories matching the filter.
func (s *InMemoryStore) List(ctx context.Context, filter *Filter) ([]*Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Memory
	for _, m := range s.memories {
		if filter != nil {
			if filter.Type != "" && m.Type != filter.Type {
				continue
			}
			if filter.Since != nil && m.CreatedAt.Before(*filter.Since) {
				continue
			}
			if filter.Until != nil && m.CreatedAt.After(*filter.Until) {
				continue
			}
		}
		result = append(result, m)
	}

	// Sort by created time (newest first by default)
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	// Apply offset and limit
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(result) {
			result = result[filter.Offset:]
		} else if filter.Offset >= len(result) {
			return []*Memory{}, nil
		}
		if filter.Limit > 0 && filter.Limit < len(result) {
			result = result[:filter.Limit]
		}
	}

	return result, nil
}

// Clear removes all memories.
func (s *InMemoryStore) Clear(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.memories = make(map[string]*Memory)
	return nil
}
