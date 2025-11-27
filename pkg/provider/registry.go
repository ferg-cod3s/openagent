package provider

import "fmt"

// Registry manages provider instances.
type Registry struct {
	providers map[ProviderType]Provider
}

// NewRegistry creates a new provider registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[ProviderType]Provider),
	}
}

// Register adds a provider to the registry.
func (r *Registry) Register(ptype ProviderType, p Provider) {
	r.providers[ptype] = p
}

// Get retrieves a provider by type.
func (r *Registry) Get(ptype ProviderType) (Provider, error) {
	p, ok := r.providers[ptype]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", ptype)
	}
	return p, nil
}

// New creates a new provider based on type and config.
func New(ptype ProviderType, cfg Config) (Provider, error) {
	switch ptype {
	case OpenAI:
		return NewOpenAI(cfg), nil
	case Anthropic:
		return NewAnthropic(cfg), nil
	case Ollama:
		return NewOllama(cfg), nil
	default:
		return nil, fmt.Errorf("unknown provider type: %s", ptype)
	}
}
