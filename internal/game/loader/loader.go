package loader

import (
	"fmt"
	"os"
	"path/filepath"
	v1 "tragedylooper/pkg/proto/v1"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// GameConfig defines the interface for accessing game configuration data.
// This allows the game engine to be decoupled from the concrete data loader implementation,
// facilitating easier testing with mock data.
type GameConfig interface {
	GetScript() *v1.ScriptConfig

	GetAbilities() map[int32]*v1.AbilityConfig
	GetCards() map[int32]*v1.CardConfig
	GetCharacters() map[int32]*v1.CharacterConfig
	GetIncidents() map[int32]*v1.IncidentConfig
}

// LoadConfig loads all game data from the specified directory.
func LoadConfig(dataDir, scriptID string) (GameConfig, error) {
	repo := newRepository()

	if err := loadAbilities(repo, dataDir); err != nil {
		return nil, err
	}
	if err := loadCards(repo, dataDir); err != nil {
		return nil, err
	}
	if err := loadCharacters(repo, dataDir); err != nil {
		return nil, err
	}
	if err := loadIncidents(repo, dataDir); err != nil {
		return nil, err
	}
	if err := loadScript(repo, dataDir, scriptID); err != nil {
		return nil, err
	}

	return repo, nil
}

func loadDataFromJSON[T proto.Message](filePath string, data T) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %w", filePath, err)
	}
	bytes, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", absPath, err)
	}

	if err := protojson.Unmarshal(bytes, data); err != nil {
		return fmt.Errorf("failed to decode json from %s: %w", absPath, err)
	}

	return nil
}

func loadAbilities(repo *repository, dataDir string) error {
	filePath := filepath.Join(dataDir, "AbilityConfigLib.json")
	var abilities v1.AbilityConfigLib
	if err := loadDataFromJSON(filePath, &abilities); err != nil {
		return err
	}
	repo.abilities = abilities.Abilities
	return nil
}

func loadCards(repo *repository, dataDir string) error {
	filePath := filepath.Join(dataDir, "CardConfigLib.json")
	var cards v1.CardConfigLib
	if err := loadDataFromJSON(filePath, &cards); err != nil {
		return err
	}
	repo.cards = cards.Cards
	return nil
}

func loadCharacters(repo *repository, dataDir string) error {
	filePath := filepath.Join(dataDir, "CharacterConfigLib.json")
	var characters v1.CharacterConfigLib
	if err := loadDataFromJSON(filePath, &characters); err != nil {
		return err
	}
	repo.characters = characters.Characters
	return nil
}

func loadIncidents(repo *repository, dataDir string) error {
	filePath := filepath.Join(dataDir, "IncidentConfigLib.json")
	var incidents v1.IncidentConfigLib
	if err := loadDataFromJSON(filePath, &incidents); err != nil {
		return err
	}
	repo.incidents = incidents.Incidents
	return nil
}

func loadScript(repo *repository, dataDir, scriptID string) error {
	filePath := filepath.Join(dataDir, "ScriptConfig", scriptID+".json")
	var script v1.ScriptConfig
	if err := loadDataFromJSON(filePath, &script); err != nil {
		return err
	}
	repo.script = &script
	return nil
}
