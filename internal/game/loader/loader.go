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
	GetIncidents() map[string]*v1.Incident
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
	if err !=.GetAbilities() map[int32]*v1.Ability {
	return g.abilities.Abilities
}

func (g *gameDataAccessor) GetCards() map[int32]*v1.Card {
	return g.cards.Cards
}

func (g *gameDataAccessor) GetCharacters() map[int32]*v1.Character {
	return g.characters.Characters
}

func (g *gameDataAccessor) GetIncidents() map[string]*v1.Incident {
	return g.incidents.Incidents
}
}
