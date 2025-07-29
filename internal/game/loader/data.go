package loader

import (
	"fmt"
	"os"
	"path/filepath"
	v1 "tragedylooper/pkg/proto/v1"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

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

func RegisterLoaders(dataDir, scriptName string) {
	Register(func(r *Repository) error {
		filePath := filepath.Join(dataDir, "AbilityConfigLib.json")
		var abilities v1.AbilityConfigLib
		if err := loadDataFromJSON(filePath, &abilities); err != nil {
			return err
		}
		r.abilities = abilities.Abilities
		return nil
	})

	Register(func(r *Repository) error {
		filePath := filepath.Join(dataDir, "CardConfigLib.json")
		var cards v1.CardConfigLib
		if err := loadDataFromJSON(filePath, &cards); err != nil {
			return err
		}
		r.cards = cards.Cards
		return nil
	})

	Register(func(r *Repository) error {
		filePath := filepath.Join(dataDir, "CharacterConfigLib.json")
		var characters v1.CharacterConfigLib
		if err := loadDataFromJSON(filePath, &characters); err != nil {
			return err
		}
		r.characters = characters.Characters
		return nil
	})

	Register(func(r *Repository) error {
		filePath := filepath.Join(dataDir, "ScriptConfig", scriptName+".json")
		var script v1.ScriptConfig
		if err := loadDataFromJSON(filePath, &script); err != nil {
			return err
		}
		r.script = &script
		return nil
	})

	Register(func(r *Repository) error {
		filePath := filepath.Join(dataDir, "IncidentConfigLib.json")
		var incidents v1.IncidentConfigLib
		if err := loadDataFromJSON(filePath, &incidents); err != nil {
			return err
		}
		r.incidents = incidents.Incidents
		return nil
	})
}