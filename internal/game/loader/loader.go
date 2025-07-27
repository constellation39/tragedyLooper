package loader

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"tragedylooper/internal/game/proto/v1"
)

// Loader defines the interface for loading all static game data from a persistent source.
// It is kept separate to allow for different data sources (e.g., files, database).
type Loader interface {
	LoadAbilities(ctx context.Context) (map[string]*v1.Ability, error)
	LoadCards(ctx context.Context) (map[string]*v1.Card, error)
	LoadCharacters(ctx context.Context) (map[string]*v1.Character, error)
	LoadScript(ctx context.Context, name string) (*v1.Script, error)
	LoadIncidents(ctx context.Context) (map[string]*v1.Incident, error)
}

// GameDataAccessor defines an interface for high-performance, cached access to game data.
// It provides methods to retrieve individual data entries by their ID.
type GameDataAccessor interface {
	GetAbility(id string) (*v1.Ability, error)
	GetCard(id string) (*v1.Card, error)
	GetCharacter(id string) (*v1.Character, error)
	GetIncident(id string) (*v1.Incident, error)
	GetScript(ctx context.Context, name string) (*v1.Script, error)
}

// jsonLoader implements the Loader interface for loading data from JSON files.
type jsonLoader struct {
	dataDir string
}

// newJSONLoader creates a new loader that reads from the given data directory.
func newJSONLoader(dataDir string) Loader {
	return &jsonLoader{dataDir: dataDir}
}

func (l *jsonLoader) loadJSONFile(_ context.Context, filename string, v interface{}) error {
	path := filepath.Join(l.dataDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal json from %s: %w", path, err)
	}
	return nil
}

func (l *jsonLoader) LoadAbilities(ctx context.Context) (map[string]*v1.Ability, error) {
	var data struct {
		Abilities map[string]*v1.Ability `json:"abilities"`
	}
	if err := l.loadJSONFile(ctx, "ability.json", &data); err != nil {
		return nil, err
	}
	return data.Abilities, nil
}

func (l *jsonLoader) LoadCards(ctx context.Context) (map[string]*v1.Card, error) {
	var data struct {
		Cards map[string]*v1.Card `json:"cards"`
	}
	if err := l.loadJSONFile(ctx, "card.json", &data); err != nil {
		return nil, err
	}
	return data.Cards, nil
}

func (l *jsonLoader) LoadCharacters(ctx context.Context) (map[string]*v1.Character, error) {
	var data struct {
		Characters map[string]*v1.Character `json:"characters"`
	}
	if err := l.loadJSONFile(ctx, "character.json", &data); err != nil {
		return nil, err
	}
	return data.Characters, nil
}

func (l *jsonLoader) LoadScript(ctx context.Context, name string) (*v1.Script, error) {
	var script v1.Script
	filename := filepath.Join("scripts", name+".json")
	if err := l.loadJSONFile(ctx, filename, &script); err != nil {
		return nil, err
	}
	return &script, nil
}

func (l *jsonLoader) LoadIncidents(ctx context.Context) (map[string]*v1.Incident, error) {
	var data struct {
		Incidents map[string]*v1.Incident `json:"incidents"`
	}
	if err := l.loadJSONFile(ctx, "IncidentConfig.json", &data); err != nil {
		return nil, err
	}
	return data.Incidents, nil
}

// gameDataAccessor implements the GameDataAccessor interface.
// It pre-loads all data using a Loader and stores it in memory for fast access.
type gameDataAccessor struct {
	loader           Loader
	abilities        map[string]*v1.Ability
	cards            map[string]*v1.Card
	characters       map[string]*v1.Character
	incidents        map[string]*v1.Incident
	scriptCache      map[string]*v1.Script
	scriptCacheMutex sync.RWMutex
}

// New creates a new GameDataAccessor by initializing a loader for the given data directory
// and pre-loading all necessary game data into a cache.
func New(ctx context.Context, dataDir string) (GameDataAccessor, error) {
	loader := newJSONLoader(dataDir)

	abilities, err := loader.LoadAbilities(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load abilities: %w", err)
	}

	cards, err := loader.LoadCards(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load cards: %w", err)
	}

	characters, err := loader.LoadCharacters(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load characters: %w", err)
	}

	incidents, err := loader.LoadIncidents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load incidents: %w", err)
	}

	return &gameDataAccessor{
		loader:      loader,
		abilities:   abilities,
		cards:       cards,
		characters:  characters,
		incidents:   incidents,
		scriptCache: make(map[string]*v1.Script),
	}, nil
}

func (g *gameDataAccessor) GetAbility(id string) (*v1.Ability, error) {
	ability, ok := g.abilities[id]
	if !ok {
		return nil, fmt.Errorf("ability with id '%s' not found", id)
	}
	return ability, nil
}

func (g *gameDataAccessor) GetCard(id string) (*v1.Card, error) {
	card, ok := g.cards[id]
	if !ok {
		return nil, fmt.Errorf("card with id '%s' not found", id)
	}
	return card, nil
}

func (g *gameDataAccessor) GetCharacter(id string) (*v1.Character, error) {
	character, ok := g.characters[id]
	if !ok {
		return nil, fmt.Errorf("character with id '%s' not found", id)
	}
	return character, nil
}

func (g *gameDataAccessor) GetIncident(id string) (*v1.Incident, error) {
	incident, ok := g.incidents[id]
	if !ok {
		return nil, fmt.Errorf("incident with id '%s' not found", id)
	}
	return incident, nil
}

func (g *gameDataAccessor) GetScript(ctx context.Context, name string) (*v1.Script, error) {
	g.scriptCacheMutex.RLock()
	script, found := g.scriptCache[name]
	g.scriptCacheMutex.RUnlock()
	if found {
		return script, nil
	}

	g.scriptCacheMutex.Lock()
	defer g.scriptCacheMutex.Unlock()

	// Double-check in case another goroutine loaded it while we were waiting for the lock.
	script, found = g.scriptCache[name]
	if found {
		return script, nil
	}

	loadedScript, err := g.loader.LoadScript(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to load script '%s': %w", name, err)
	}

	g.scriptCache[name] = loadedScript
	return loadedScript, nil
}
