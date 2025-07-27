package loader

import (
	"encoding/json"
	"os"
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
	if err := loadData(filepath.Join(dataDir, "abilities"), &gameData.Abilities); err != nil {
		return nil, err
	}
	if err := loadData(filepath.Join(dataDir, "characters"), &gameData.Characters, gameData); err != nil {
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

type tragedyConditionJSON struct {
	TragedyType interface{}          `json:"tragedy_type"`
	Day         int32              `json:"day"`
	CulpritID   string             `json:"culprit_id"`
	Conditions  []*model.Condition `json:"conditions"`
	TargetRule  interface{}        `json:"target_rule"`
	Abilities   []*model.Ability   `json:"abilities"`
}

func convertTragedy(tcj *tragedyConditionJSON) *model.TragedyCondition {
	var tragedyType model.TragedyType
	switch v := tcj.TragedyType.(type) {
	case string:
		tragedyType = model.TragedyType(model.TragedyType_value[v])
	case float64:
		tragedyType = model.TragedyType(v)
	}

	var targetRule model.TargetRuleType
	switch v := tcj.TargetRule.(type) {
	case string:
		targetRule = model.TargetRuleType(model.TargetRuleType_value[v])
	case float64:
		targetRule = model.TargetRuleType(v)
	}

	return &model.TragedyCondition{
		TragedyType: tragedyType,
		Day:         tcj.Day,
		CulpritId:   tcj.CulpritID,
		Conditions:  tcj.Conditions,
		TargetRule:  targetRule,
		Abilities:   tcj.Abilities,
	}
}

type gameEndConditionJSON struct {
	Type string `json:"type"`
}

type scriptJSON struct {
	Name           string                  `json:"name"`
	Loops          int32                   `json:"loops"`
	DaysPerLoop    int32                   `json:"days_per_loop"`
	WinConditions  []*gameEndConditionJSON `json:"win_conditions"`
	LoseConditions []*gameEndConditionJSON `json:"lose_conditions"`
	Tragedies      []*tragedyConditionJSON `json:"tragedies"`
	Characters     []*model.CharacterConfig  `json:"characters"`
}

func convertScript(sj *scriptJSON) *model.Script {
	winConditions := make([]*model.GameEndCondition, len(sj.WinConditions))
	for i, wc := range sj.WinConditions {
		winConditions[i] = &model.GameEndCondition{Type: model.GameEndConditionType(model.GameEndConditionType_value[wc.Type])}
	}

	loseConditions := make([]*model.GameEndCondition, len(sj.LoseConditions))
	for i, lc := range sj.LoseConditions {
		loseConditions[i] = &model.GameEndCondition{Type: model.GameEndConditionType(model.GameEndConditionType_value[lc.Type])}
	}

	tragedies := make([]*model.TragedyCondition, len(sj.Tragedies))
	for i, t := range sj.Tragedies {
		tragedies[i] = convertTragedy(t)
	}

	return &model.Script{
		Name:           sj.Name,
		LoopCount:      sj.Loops,
		DaysPerLoop:    sj.DaysPerLoop,
		WinConditions:  winConditions,
		LoseConditions: loseConditions,
		Tragedies:      tragedies,
		Characters:     sj.Characters,
	}
}

type characterJSON struct {
	Name      string   `json:"name"`
	Abilities []string `json:"abilities"`
}

type abilityJSON struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	TriggerType string        `json:"trigger_type"`
	Effect      *model.Effect `json:"effect"`
	OncePerLoop bool          `json:"once_per_loop"`
	RefusalRole string        `json:"refusal_role"`
	Target      *model.Target `json:"target"`
}

func loadData[T any](dir string, target *map[string]*T, gameData ...*GameData) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() {
			data, err := os.ReadFile(filepath.Join(dir, file.Name()))
			if err != nil {
				return err
			}

			baseDir := filepath.Base(dir)
			if baseDir == "tragedies" {
				var item tragedyConditionJSON
				if err := json.Unmarshal(data, &item); err != nil {
					return err
				}
				converted := convertTragedy(&item)
				(*target)[file.Name()] = any(converted).(*T)
			} else if baseDir == "scripts" {
				var item scriptJSON
				if err := json.Unmarshal(data, &item); err != nil {
					return err
				}
				converted := convertScript(&item)
				(*target)[file.Name()] = any(converted).(*T)
			} else if baseDir == "characters" {
				var item characterJSON
				if err := json.Unmarshal(data, &item); err != nil {
					return err
				}
				character := &model.Character{Name: item.Name}
				for _, abilityName := range item.Abilities {
					if ability, ok := gameData[0].Abilities[abilityName]; ok {
						character.Abilities = append(character.Abilities, ability)
					}
				}
				(*target)[file.Name()] = any(character).(*T)
			} else if baseDir == "abilities" {
				var item abilityJSON
				if err := json.Unmarshal(data, &item); err != nil {
					return err
				}
				ability := &model.Ability{
					Name:        item.Name,
					Description: item.Description,
					TriggerType: model.AbilityTriggerType(model.AbilityTriggerType_value[item.TriggerType]),
					Effect:      item.Effect,
					OncePerLoop: item.OncePerLoop,
					RefusalRole: model.PlayerRole(model.PlayerRole_value[item.RefusalRole]),
					Target:      item.Target,
				}
				(*target)[file.Name()] = any(ability).(*T)
			} else {
				var item T
				if err := json.Unmarshal(data, &item); err != nil {
					return err
				}
				(*target)[file.Name()] = &item
			}
		}
	}

	return nil
}
