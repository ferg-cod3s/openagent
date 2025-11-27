package evolution

import (
	"context"
	"testing"
)

func TestRandomMutator(t *testing.T) {
	m := NewRandomMutator(42)

	ind := &Individual{
		ID: "test",
		Genome: Genome{
			Genes: map[string]Gene{
				"rate": {Name: "rate", Value: 0.5, Mutable: true, MinValue: 0.0, MaxValue: 1.0},
				"size": {Name: "size", Value: 10, Mutable: true, MinValue: 0, MaxValue: 100},
				"flag": {Name: "flag", Value: true, Mutable: true},
				"name": {Name: "name", Value: "test", Mutable: false},
			},
		},
	}

	mutated, err := m.Mutate(context.Background(), ind, 1.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mutated.ID == ind.ID {
		t.Error("expected new ID")
	}

	// Non-mutable gene should not change
	if mutated.Genome.Genes["name"].Value != "test" {
		t.Error("non-mutable gene should not change")
	}
}

func TestTournamentSelector(t *testing.T) {
	s := NewTournamentSelector(42, 3)

	pop := &Population{
		Individuals: []*Individual{
			{ID: "a", Fitness: 0.1},
			{ID: "b", Fitness: 0.5},
			{ID: "c", Fitness: 0.9},
			{ID: "d", Fitness: 0.3},
			{ID: "e", Fitness: 0.7},
		},
	}

	selected, err := s.Select(context.Background(), pop, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(selected) != 3 {
		t.Errorf("expected 3 selected, got %d", len(selected))
	}
}

func TestSinglePointCrossover(t *testing.T) {
	c := NewSinglePointCrossover(42)

	parent1 := &Individual{
		ID: "p1",
		Genome: Genome{
			Genes: map[string]Gene{
				"a": {Name: "a", Value: 1},
				"b": {Name: "b", Value: 2},
				"c": {Name: "c", Value: 3},
			},
			Version: 1,
		},
	}

	parent2 := &Individual{
		ID: "p2",
		Genome: Genome{
			Genes: map[string]Gene{
				"a": {Name: "a", Value: 10},
				"b": {Name: "b", Value: 20},
				"c": {Name: "c", Value: 30},
			},
			Version: 2,
		},
	}

	child, err := c.Cross(context.Background(), parent1, parent2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if child.ID == parent1.ID || child.ID == parent2.ID {
		t.Error("expected new ID for child")
	}

	if child.Genome.Version <= parent2.Genome.Version {
		t.Error("expected child version to be higher than parents")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.PopulationSize != 100 {
		t.Errorf("expected PopulationSize 100, got %d", cfg.PopulationSize)
	}
	if cfg.MutationRate != 0.1 {
		t.Errorf("expected MutationRate 0.1, got %f", cfg.MutationRate)
	}
	if cfg.CrossoverRate != 0.7 {
		t.Errorf("expected CrossoverRate 0.7, got %f", cfg.CrossoverRate)
	}
}
