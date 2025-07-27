package loader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"tragedylooper/internal/game/proto/v1"
)

func loadData[T any](filePath string) (T, error) {
	var zero T

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return zero, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	var items T
	if err := json.Unmarshal(bytes, &items); err != nil {
		return zero, fmt.Errorf("failed to unmarshal json from %s: %w", filePath, err)
	}

	return items, nil
}

func LoadAbility(dataDir string) (*v1.AbilityLib, error) {
	filePath := filepath.Join(dataDir, "ability.json")
	abilities, err := loadData[*v1.AbilityLib](filePath)
	if err != nil {
		return nil, err
	}
	return abilities, nil
}

func LoadCard(dataDir string) (*v1.CardLib, error) {
	filePath := filepath.Join(dataDir, "card.json")
	cards, err := loadData[*v1.CardLib](filePath)
	if err != nil {
		return nil, err
	}

	return cards, nil
}

func LoadCharacter(dataDir string) (*v1.CharacterLib, error) {
	filePath := filepath.Join(dataDir, "character.json")
	characters, err := loadData[*v1.CharacterLib](filePath)
	if err != nil {
		return nil, err
	}
	return characters, nil
}

func LoadScript(dataDir, scriptName string) (*v1.Script, error) {
	filePath := filepath.Join(dataDir, "scripts", scriptName+".json")
	script, err := loadData[*v1.Script](filePath)
	if err != nil {
		return nil, err
	}
	return script, nil
}

func LoadIncidents(dataDir string) (*v1.IncidentConfigLib, error) {
	filePath := filepath.Join(dataDir, "Incident.json")
	incidents, err := loadData[*v1.IncidentConfigLib](filePath)
	if err != nil {
		return nil, err
	}
	return incidents, nil
}
