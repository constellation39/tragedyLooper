package loader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	v1 "tragedylooper/internal/game/proto/v1"
)

func loadDataFromJSON[T any](filePath string) (T, error) {
	var zero T

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return zero, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	var items T
	if err := json.Unmarshal(bytes, &items); err != nil {
		return zero, fmt.Errorf("failed to decode json from %s: %w", filePath, err)
	}

	return items, nil
}

func LoadAbility(dataDir string) (*v1.AbilityConfigLib, error) {
	filePath := filepath.Join(dataDir, "AbilityConfigLib.json")
	abilities, err := loadDataFromJSON[*v1.AbilityConfigLib](filePath)
	if err != nil {
		return nil, err
	}
	return abilities, nil
}

func LoadCard(dataDir string) (*v1.CardConfigLib, error) {
	filePath := filepath.Join(dataDir, "CardConfigLib.json")
	cards, err := loadDataFromJSON[*v1.CardConfigLib](filePath)
	if err != nil {
		return nil, err
	}

	return cards, nil
}

func LoadCharacter(dataDir string) (*v1.CharacterConfigLib, error) {
	filePath := filepath.Join(dataDir, "CharacterConfigLib.json")
	characters, err := loadDataFromJSON[*v1.CharacterConfigLib](filePath)
	if err != nil {
		return nil, err
	}
	return characters, nil
}

func LoadScript(dataDir, scriptName string) (*v1.ScriptConfig, error) {
	filePath := filepath.Join(dataDir, "ScriptConfig", scriptName+".json")
	script, err := loadDataFromJSON[*v1.ScriptConfig](filePath)
	if err != nil {
		return nil, err
	}
	return script, nil
}

func LoadIncidents(dataDir string) (*v1.IncidentConfigLib, error) {
	filePath := filepath.Join(dataDir, "IncidentConfigLib.json")
	incidents, err := loadDataFromJSON[*v1.IncidentConfigLib](filePath)
	if err != nil {
		return nil, err
	}
	return incidents, nil
}
