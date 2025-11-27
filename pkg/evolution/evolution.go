// Package evolution implements mutation, fitness evaluation, and selection for agents.
package evolution

import (
	"context"
	"time"
)

// Individual represents an evolvable entity with a genome.
type Individual struct {
	ID        string                 `json:"id"`
	Genome    Genome                 `json:"genome"`
	Fitness   float64                `json:"fitness"`
	Age       int                    `json:"age"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// Genome represents the genetic information of an individual.
type Genome struct {
	Genes   map[string]Gene `json:"genes"`
	Version int             `json:"version"`
}

// Gene represents a single genetic trait.
type Gene struct {
	Name     string      `json:"name"`
	Value    interface{} `json:"value"`
	Mutable  bool        `json:"mutable"`
	MinValue interface{} `json:"min_value,omitempty"`
	MaxValue interface{} `json:"max_value,omitempty"`
}

// Population represents a collection of individuals.
type Population struct {
	Individuals []*Individual `json:"individuals"`
	Generation  int           `json:"generation"`
	BestFitness float64       `json:"best_fitness"`
	AvgFitness  float64       `json:"avg_fitness"`
}

// Mutator defines the interface for mutation operations.
type Mutator interface {
	// Mutate applies mutations to an individual's genome.
	Mutate(ctx context.Context, ind *Individual, rate float64) (*Individual, error)
}

// FitnessEvaluator defines the interface for fitness evaluation.
type FitnessEvaluator interface {
	// Evaluate computes the fitness of an individual.
	Evaluate(ctx context.Context, ind *Individual) (float64, error)
	// EvaluatePopulation computes fitness for all individuals.
	EvaluatePopulation(ctx context.Context, pop *Population) error
}

// Selector defines the interface for selection operations.
type Selector interface {
	// Select chooses individuals for reproduction.
	Select(ctx context.Context, pop *Population, count int) ([]*Individual, error)
}

// Crossover defines the interface for crossover operations.
type Crossover interface {
	// Cross combines two individuals to produce offspring.
	Cross(ctx context.Context, parent1, parent2 *Individual) (*Individual, error)
}

// Engine orchestrates the evolutionary process.
type Engine interface {
	// Initialize creates the initial population.
	Initialize(ctx context.Context, size int) (*Population, error)
	// Evolve runs one generation of evolution.
	Evolve(ctx context.Context, pop *Population) (*Population, error)
	// Run executes the evolutionary process for n generations.
	Run(ctx context.Context, generations int) (*Population, error)
}

// Config contains evolution engine configuration.
type Config struct {
	PopulationSize  int     `json:"population_size"`
	MutationRate    float64 `json:"mutation_rate"`
	CrossoverRate   float64 `json:"crossover_rate"`
	ElitismCount    int     `json:"elitism_count"`
	TournamentSize  int     `json:"tournament_size"`
	MaxGenerations  int     `json:"max_generations"`
	TargetFitness   float64 `json:"target_fitness,omitempty"`
	StagnationLimit int     `json:"stagnation_limit,omitempty"`
}

// DefaultConfig returns a default evolution configuration.
func DefaultConfig() *Config {
	return &Config{
		PopulationSize:  100,
		MutationRate:    0.1,
		CrossoverRate:   0.7,
		ElitismCount:    2,
		TournamentSize:  3,
		MaxGenerations:  1000,
		StagnationLimit: 50,
	}
}
