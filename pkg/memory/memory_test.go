package memory

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryStore(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	// Test Save
	m := &Memory{
		Type:    TypeEpisodic,
		Content: "Test memory",
	}
	err := store.Save(ctx, m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.ID == "" {
		t.Error("expected ID to be set")
	}

	// Test Get
	got, err := store.Get(ctx, m.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Content != "Test memory" {
		t.Errorf("expected content 'Test memory', got %q", got.Content)
	}

	// Test Get non-existent
	_, err = store.Get(ctx, "non-existent")
	if err == nil {
		t.Error("expected error for non-existent memory")
	}

	// Test List
	memories, err := store.List(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(memories) != 1 {
		t.Errorf("expected 1 memory, got %d", len(memories))
	}

	// Test List with filter
	memories, err = store.List(ctx, &Filter{Type: TypeSemantic})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(memories) != 0 {
		t.Errorf("expected 0 memories with filter, got %d", len(memories))
	}

	// Test Delete
	err = store.Delete(ctx, m.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = store.Get(ctx, m.ID)
	if err == nil {
		t.Error("expected error for deleted memory")
	}

	// Test Clear
	store.Save(ctx, &Memory{Content: "m1"})
	store.Save(ctx, &Memory{Content: "m2"})
	err = store.Clear(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	memories, _ = store.List(ctx, nil)
	if len(memories) != 0 {
		t.Errorf("expected 0 memories after clear, got %d", len(memories))
	}
}

func TestInMemoryStoreFiltering(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	now := time.Now()
	store.Save(ctx, &Memory{Content: "m1", Type: TypeEpisodic, CreatedAt: now.Add(-2 * time.Hour)})
	store.Save(ctx, &Memory{Content: "m2", Type: TypeSemantic, CreatedAt: now.Add(-1 * time.Hour)})
	store.Save(ctx, &Memory{Content: "m3", Type: TypeEpisodic, CreatedAt: now})

	// Filter by type
	memories, _ := store.List(ctx, &Filter{Type: TypeEpisodic})
	if len(memories) != 2 {
		t.Errorf("expected 2 episodic memories, got %d", len(memories))
	}

	// Filter by time range
	since := now.Add(-90 * time.Minute)
	memories, _ = store.List(ctx, &Filter{Since: &since})
	if len(memories) != 2 {
		t.Errorf("expected 2 memories since 90 min ago, got %d", len(memories))
	}

	// Filter with limit
	memories, _ = store.List(ctx, &Filter{Limit: 1})
	if len(memories) != 1 {
		t.Errorf("expected 1 memory with limit, got %d", len(memories))
	}

	// Filter with offset
	memories, _ = store.List(ctx, &Filter{Offset: 1})
	if len(memories) != 2 {
		t.Errorf("expected 2 memories with offset, got %d", len(memories))
	}
}
