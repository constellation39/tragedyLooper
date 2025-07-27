package loader

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	model "tragedylooper/internal/game/proto/v1"
)

// GameData holds all the static data for the game.
type GameData struct {
	Scripts    map[string]*model.Script
	Characters map[string]*model.Character
	Abilities  map[string]*model.Ability
	Cards      map[string]*model.Card
	Tragedies  map[string]*model.TragedyCondition
	Incidents  map[string]*model.Incident
}

// LoadGameData loads all game data from the specified directory.
func LoadGameData(dataDir string) (*GameData, error) {
	gameData := &GameData{
		Scripts:    make(map[string]*model.Script),
		Characters: make(map[string]*model.Character),
		Abilities:  make(map[string]*model.Ability),
		Cards:      make(map[string]*model.Card),
		Tragedies:  make(map[string]*model.TragedyCondition),
		Incidents:  make(map[string]*model.Incident),
	}

	if err := loadData(filepath.Join(dataDir, "scripts"), &gameData.Scripts); err != nil {
		return nil, err
	}
	if err := loadData(filepath.Join(dataDir, "characters"), &gameData.Characters); err != nil {
		return nil, err
	}
	if err := loadData(filepath.Join(dataDir, "abilities"), &gameData.Abilities); err != nil {
		return nil, err
	}
	if err := loadData(filepath.Join(dataDir, "cards"), &gameData.Cards); err != nil {
		return nil, err
	}
	if err := loadData(filepath.Join(dataDir, "tragedies"), &gameData.Tragedies); err != nil {
		return nil, err
	}
	if err := loadData(filepath.Join(dataDir, "incidents"), &gameData.Incidents); err != nil {
		return nil, err
	}

	return gameData, nil
}

func loadData[T any](dir string, target *map[string]*T) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() {
			data, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
			if err != nil {
				return err
			}

			var item T
			if err := json.Unmarshal(data, &item); err != nil {
				return err
			}

			(*target)[file.Name()] = &item
		}
	}

	return nil
}