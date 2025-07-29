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
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	if err := protojson.Unmarshal(bytes, data); err != nil {
		return fmt.Errorf("failed to decode json from %s: %w", filePath, err)
	}

	return nil
}

func LoadAbility(dataDir string) (*v1.AbilityConfigLib, error) {
	filePath := filepath.Join(dataDir, "AbilityConfigLib.json")
	var abilities v1.AbilityConfigLib
	if err := loadDataFromJSON(filePath, &abilities); err != nil {
		return nil, err
	}
	return &abilities, nil
}

func LoadCard(dataDir string) (*v1.CardConfigLib, error) {
	filePath := filepath.Join(dataDir, "CardConfigLib.json")
	var cards v1.CardConfigLib
	if err := loadDataFromJSON(filePath, &cards); err != nil {
		return nil, err
	}

	return &cards, nil
}

func LoadCharacter(dataDir string) (*v1.CharacterConfigLib, error) {
	filePath := filepath.Join(dataDir, "CharacterConfigLib.json")
	var characters v1.CharacterConfigLib
	if err := loadDataFromJSON(filePath, &characters); err != nil {
		return nil, err
	}
	return &characters, nil
}

func LoadScript(dataDir, scriptName string) (*v1.ScriptConfig, error) {
	filePath := filepath.Join(dataDir, "ScriptConfig", scriptName+".json")
	var script v1.ScriptConfig
	if err := loadDataFromJSON(filePath, &script); err != nil {
		return nil, err
	}
	return &script, nil
}

func LoadIncidents(dataDir string) (*v1.IncidentConfigLib, error) {
	filePath := filepath.Join(dataDir, "IncidentConfigLib.json")
	var incidents v1.IncidentConfigLib
	if err := loadDataFromJSON(filePath, &incidents); err != nil {
		return nil, err
	}
	return &incidents, nil
}
