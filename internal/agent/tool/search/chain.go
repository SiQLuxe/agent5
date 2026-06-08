package search

import (
	"context"
	"errors"
	"fmt"
	"log"
)

// ChainProvider tries providers in order, falling back on error.
type ChainProvider struct {
	providers []WebSearchProvider
}

// NewChainProvider creates a ChainProvider with the given providers in priority order.
func NewChainProvider(providers ...WebSearchProvider) *ChainProvider {
	return &ChainProvider{providers: providers}
}

// Search tries each provider in order. If a provider fails, it logs the error
// and tries the next one. Returns results from the first successful provider.
// If all providers fail, returns an error listing all failures.
func (c *ChainProvider) Search(ctx context.Context, query string, count int) ([]WebSearchResult, error) {
	if len(c.providers) == 0 {
		return nil, errors.New("no search providers configured")
	}

	var errs []error
	for _, p := range c.providers {
		results, err := p.Search(ctx, query, count)
		if err == nil {
			return results, nil
		}
		log.Printf("[search] provider %s failed: %v", p.Name(), err)
		errs = append(errs, fmt.Errorf("%s: %w", p.Name(), err))
	}

	return nil, fmt.Errorf("all search providers failed: %v", errors.Join(errs...))
}

// Name returns the chain provider name.
func (c *ChainProvider) Name() string {
	if len(c.providers) == 0 {
		return "chain(empty)"
	}
	names := ""
	for i, p := range c.providers {
		if i > 0 {
			names += ","
		}
		names += p.Name()
	}
	return "chain(" + names + ")"
}
