package loader

import (
	"fmt"
	"sync"
)

type ( 
    Loader func(r *Repository) error 
)

var (
	loaders []Loader
	mu      sync.Mutex
)

func Register(loader Loader) {
	mu.Lock()
	defer mu.Unlock()
	loaders = append(loaders, loader)
}

func Load(r *Repository) error {
	mu.Lock()
	defer mu.Unlock()

	for _, loader := range loaders {
		if err := loader(r); err != nil {
			return fmt.Errorf("failed to load data: %w", err)
		}
	}
	return nil
}