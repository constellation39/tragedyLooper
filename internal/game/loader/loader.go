package loader

import (
	"fmt"
	"sync"
	"tragedylooper/internal/game/proto/v1"
)

type Loader interface {
	LoadGameDataAccessor(name string) (GameDataAccessor, error)
}

type GameDataAccessor interface {
	GetAbility(id int32) (*v1.Ability, error)
	GetCard(id int32) (*v1.Card, error)
	GetCharacter(id int32) (*v1.Character, error)
	GetIncident(id int32) (*v1.Incident, error)
	GetScript() (*v1.Script, error)

	GetAbilities() map[int32]*v1.Ability
	GetCards() map[int32]*v1.Card
	GetCharacters() map[int32]*v1.Character
	GetIncidents() map[int32]*v1.Incident
}

// jsonLoader implements the Loader interface for loading data from JSON files.
type jsonLoader struct {
	dataDir  string
	sync.Map // map[string]GameDataAccessor
}

// NewJSONLoader creates a new loader that reads from the given data directory.
func NewJSONLoader(dataDir string) Loader {
	return &jsonLoader{dataDir: dataDir}
}

// gameDataAccessor implements the GameDataAccessor interface.
// It pre-loads all data using a Loader and stores it in memory for fast access.
type gameDataAccessor struct {
	loader      Loader
	abilities   *v1.AbilityLib
	cards       *v1.CardLib
	characters  *v1.CharacterLib
	incidents   *v1.IncidentConfigLib
	scriptCache *v1.Script
}

func (l *jsonLoader) LoadGameDataAccessor(name string) (GameDataAccessor, error) {
	if gda, ok := l.Load(name); ok {
		return gda.(GameDataAccessor), nil
	}

	abilities, err := LoadAbility(l.dataDir)
	if err != nil {
		panic(fmt.Sprintf("failed to load abilities: %v", err))
	}

	cards, err := LoadCard(l.dataDir)
	if err != nil {
		panic(fmt.Sprintf("failed to load cards: %v", err))
	}

	characters, err := LoadCharacter(l.dataDir)
	if err != nil {
		panic(fmt.Sprintf("failed to load characters: %v", err))
	}

	script, err := LoadScript(l.dataDir, name)
	if err != nil {
		panic(fmt.Sprintf("failed to load script %s: %v", name, err))
	}

	incidents, err := LoadIncidents(l.dataDir)
	if err != nil {
		panic(fmt.Sprintf("failed to load incidents for script %s: %v", name, err))
	}

	gda := &gameDataAccessor{
		loader:      l,
		abilities:   abilities,
		cards:       cards,
		characters:  characters,
		incidents:   incidents,
		scriptCache: script,
	}

	l.Store(name, gda)
	return gda, nil
}

func (g *gameDataAccessor) GetAbility(id int32) (*v1.Ability, error) {
	a, ok := g.abilities.Abilities[id]
	if !ok {
		return nil, fmt.Errorf("ability with id %d not found", id)
	}
	return a, nil
}

func (g *gameDataAccessor) GetCard(id int32) (*v1.Card, error) {
	c, ok := g.cards.Cards[id]
	if !ok {
		return nil, fmt.Errorf("card with id %d not found", id)
	}
	return c, nil
}

func (g *gameDataAccessor) GetCharacter(id int32) (*v1.CharacterConfig, error) {
	c, ok := g.characters.Characters[id]
	if !ok {
		return nil, fmt.Errorf("character with id %d not found", id)
	}
	return c, nil
}

func (g *gameDataAccessor) GetIncident(id int32) (*v1.Incident, error) {
	i, ok := g.incidents.Incidents[id]
	if !ok {
		return nil, fmt.Errorf("incident with id %s not found", id)
	}
	return i, nil
}

func (g *gameDataAccessor) GetScript() (*v1.Script, error) {
	if g.scriptCache == nil {
		return nil, fmt.Errorf("script not loaded")
	}
	return g.scriptCache, nil
}

func (g *gameDataAccessor) GetAbilities() map[int32]*v1.Ability {
	return g.abilities
}

func (g *gameDataAccessor) GetCards() map[int32]*v1.Card {
	return g.cards
}

func (g *gameDataAccessor) GetCharacters() map[int32]*v1.CharacterConfig {
	return g.characters
}

func (g *gameDataAccessor) GetIncidents() map[int32]*v1.IncidentConfig {
	return g.incidents
}
