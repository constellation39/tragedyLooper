package loader

import (
	"fmt"
	"path/filepath"

	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	v1 "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
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

func loadDataFromCUE[T proto.Message](filePath string, data T) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %w", filePath, err)
	}

	c := cuecontext.New()
	bis := load.Instances([]string{absPath}, nil)
	if len(bis) == 0 {
		return fmt.Errorf("no CUE instances found for %s", absPath)
	}
	b := bis[0]
	if err := b.Err; err != nil {
		return fmt.Errorf("failed to load CUE instance for %s: %w", absPath, err)
	}
	v := c.BuildInstance(b)
	if err := v.Err(); err != nil {
		return fmt.Errorf("failed to build CUE instance for %s: %w", absPath, err)
	}

	if err := v.Decode(data); err != nil {
		return fmt.Errorf("failed to decode CUE from %s: %w", absPath, err)
	}

	return nil
}

func loadAbilities(repo *repository, dataDir string) error {
	filePath := filepath.Join(dataDir, "AbilityConfigLib.cue")
	var abilities v1.AbilityConfigLib
	if err := loadDataFromCUE(filePath, &abilities); err != nil {
		return err
	}
	repo.abilities = abilities.Abilities
	return nil
}

func loadCards(repo *repository, dataDir string) error {
	filePath := filepath.Join(dataDir, "CardConfigLib.cue")
	var cards v1.CardConfigLib
	if err := loadDataFromCUE(filePath, &cards); err != nil {
		return err
	}
	repo.cards = cards.Cards
	return nil
}

func loadCharacters(repo *repository, dataDir string) error {
	filePath := filepath.Join(dataDir, "CharacterConfigLib.cue")
	var characters v1.CharacterConfigLib
	if err := loadDataFromCUE(filePath, &characters); err != nil {
		return err
	}
	repo.characters = characters.Characters
	return nil
}

func loadIncidents(repo *repository, dataDir string) error {
	filePath := filepath.Join(dataDir, "IncidentConfigLib.cue")
	var incidents v1.IncidentConfigLib
	if err := loadDataFromCUE(filePath, &incidents); err != nil {
		return err
	}
	repo.incidents = incidents.Incidents
	return nil
}

func loadScript(repo *repository, dataDir, scriptID string) error {
	filePath := filepath.Join(dataDir, "ScriptConfig", scriptID+".cue")
	var script v1.ScriptConfig
	if err := loadDataFromCUE(filePath, &script); err != nil {
		return err
	}
	repo.script = &script
	return nil
}
