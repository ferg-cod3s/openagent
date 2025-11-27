package evolution

import (
	"context"
	"math/rand"

	"github.com/google/uuid"
)

// RandomMutator implements random mutations on genes.
type RandomMutator struct {
	rng *rand.Rand
}

// NewRandomMutator creates a new random mutator.
func NewRandomMutator(seed int64) *RandomMutator {
	return &RandomMutator{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// Mutate applies random mutations to an individual's genome.
func (m *RandomMutator) Mutate(ctx context.Context, ind *Individual, rate float64) (*Individual, error) {
	newInd := &Individual{
		ID:       uuid.New().String(),
		Age:      0,
		Metadata: make(map[string]interface{}),
		Genome: Genome{
			Genes:   make(map[string]Gene),
			Version: ind.Genome.Version + 1,
		},
	}

	for name, gene := range ind.Genome.Genes {
		newGene := Gene{
			Name:     gene.Name,
			Value:    gene.Value,
			Mutable:  gene.Mutable,
			MinValue: gene.MinValue,
			MaxValue: gene.MaxValue,
		}

		if gene.Mutable && m.rng.Float64() < rate {
			newGene.Value = m.mutateValue(gene)
		}

		newInd.Genome.Genes[name] = newGene
	}

	return newInd, nil
}

func (m *RandomMutator) mutateValue(gene Gene) interface{} {
	switch v := gene.Value.(type) {
	case float64:
		min, max := 0.0, 1.0
		if gene.MinValue != nil {
			min = gene.MinValue.(float64)
		}
		if gene.MaxValue != nil {
			max = gene.MaxValue.(float64)
		}
		delta := (max - min) * 0.1 * (m.rng.Float64()*2 - 1)
		newVal := v + delta
		if newVal < min {
			newVal = min
		}
		if newVal > max {
			newVal = max
		}
		return newVal
	case int:
		min, max := 0, 100
		if gene.MinValue != nil {
			min = gene.MinValue.(int)
		}
		if gene.MaxValue != nil {
			max = gene.MaxValue.(int)
		}
		delta := m.rng.Intn(3) - 1
		newVal := v + delta
		if newVal < min {
			newVal = min
		}
		if newVal > max {
			newVal = max
		}
		return newVal
	case bool:
		return !v
	case string:
		return v
	default:
		return v
	}
}

// TournamentSelector implements tournament selection.
type TournamentSelector struct {
	rng            *rand.Rand
	tournamentSize int
}

// NewTournamentSelector creates a new tournament selector.
func NewTournamentSelector(seed int64, size int) *TournamentSelector {
	if size < 2 {
		size = 2
	}
	return &TournamentSelector{
		rng:            rand.New(rand.NewSource(seed)),
		tournamentSize: size,
	}
}

// Select chooses individuals using tournament selection.
func (s *TournamentSelector) Select(ctx context.Context, pop *Population, count int) ([]*Individual, error) {
	selected := make([]*Individual, 0, count)

	for i := 0; i < count; i++ {
		winner := s.tournament(pop.Individuals)
		selected = append(selected, winner)
	}

	return selected, nil
}

func (s *TournamentSelector) tournament(individuals []*Individual) *Individual {
	var best *Individual
	for i := 0; i < s.tournamentSize; i++ {
		idx := s.rng.Intn(len(individuals))
		candidate := individuals[idx]
		if best == nil || candidate.Fitness > best.Fitness {
			best = candidate
		}
	}
	return best
}

// SinglePointCrossover implements single-point crossover.
type SinglePointCrossover struct {
	rng *rand.Rand
}

// NewSinglePointCrossover creates a new single-point crossover.
func NewSinglePointCrossover(seed int64) *SinglePointCrossover {
	return &SinglePointCrossover{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// Cross combines two individuals using single-point crossover.
func (c *SinglePointCrossover) Cross(ctx context.Context, parent1, parent2 *Individual) (*Individual, error) {
	child := &Individual{
		ID:       uuid.New().String(),
		Age:      0,
		Metadata: make(map[string]interface{}),
		Genome: Genome{
			Genes:   make(map[string]Gene),
			Version: max(parent1.Genome.Version, parent2.Genome.Version) + 1,
		},
	}

	geneNames := make([]string, 0, len(parent1.Genome.Genes))
	for name := range parent1.Genome.Genes {
		geneNames = append(geneNames, name)
	}

	crossoverPoint := c.rng.Intn(len(geneNames))

	for i, name := range geneNames {
		var gene Gene
		if i < crossoverPoint {
			gene = parent1.Genome.Genes[name]
		} else {
			if g, ok := parent2.Genome.Genes[name]; ok {
				gene = g
			} else {
				gene = parent1.Genome.Genes[name]
			}
		}
		child.Genome.Genes[name] = gene
	}

	return child, nil
}
