package loader

import (
	"fmt"
	"path/filepath"

	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	v1 "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
	"google.golang.org/protobuf/proto"
)

// ScriptConfig defines the interface for accessing all loaded game configuration data.
type ScriptConfig interface {
	GetScript() *v1.ScriptConfig
	GetPlotConfig() *v1.PlotConfig
	GetModel(id int32) ScriptModel
}

type ScriptModel interface {
}

// scriptConfig is the concrete implementation of the ScriptConfig interface.
// It holds all the loaded script and library data.
type scriptConfig struct {
	script *v1.ScriptConfig
}

// newRepository creates a new, empty scriptConfig repository.
func newRepository() *scriptConfig {
	return &scriptConfig{
		script: &v1.ScriptConfig{},
	}
}

// LoadConfig loads all game data from the specified directory and script.
// It uses a data-driven approach by iterating over a local map that points to the repository's fields.
func LoadConfig(dataDir, scriptID string) (ScriptConfig, error) {
	repo := newRepository()

	// Load the main script file.
	scriptPath := filepath.Join(dataDir, "scripts", scriptID+".cue")
	if err := loadDataFromCUE(scriptPath, repo.script); err != nil {
		return nil, fmt.Errorf("failed to load script '%s': %w", scriptID, err)
	}

	return repo, nil
}

// loadDataFromCUE is a generic function that loads and decodes data from a CUE file
// into a given protocol buffer message.
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

// Getters for the loaded data from the default loader

func (s *scriptConfig) GetScript() *v1.ScriptConfig { return s.script }

func (s *scriptConfig) GetModel(id int32) ScriptModel {
	if s.script == nil || s.script.ScriptModels == nil {
		return nil
	}
	model, ok := s.script.ScriptModels[id]
	if !ok {
		return nil
	}
	return &model
}

func (s *scriptConfig) GetCharacter(id int32) *v1.CharacterConfig {
	if s.characters == nil || s.characters.Characters == nil {
		return nil
	}
	char, ok := s.characters.Characters[id]
	if !ok {
		return nil
	}
	return &char
}

func (s *scriptConfig) GetIncident(id int32) *v1.IncidentConfig {
	if s.incidents == nil || s.incidents.Incidents == nil {
		return nil
	}
	incident, ok := s.incidents.Incidents[id]
	if !ok {
		return nil
	}
	return &incident
}

func (s *scriptConfig) GetCard(id int32) *v1.CardConfig {
	if s.cards == nil || s.cards.Cards == nil {
		return nil
	}
	card, ok := s.cards.Cards[id]
	if !ok {
		return nil
	}
	return &card
}

func (s *scriptConfig) GetAbility(id int32) *v1.AbilityConfig {
	if s.abilities == nil || s.abilities.Abilities == nil {
		return nil
	}
	ability, ok := s.abilities.Abilities[id]
	if !ok {
		return nil
	}
	return &ability
}

// Global access functions

type cfgPtr interface {
	*v1.AbilityConfig |
	*v1.CardConfig |
	*v1.CharacterConfig |
	*v1.IncidentConfig
}

func Get[T cfgPtr](r ScriptConfig, id int32) (T, error) {
	m, err := pickMap[T](r)
	if err != nil {
		var zero T
		return zero, err
	}
	v, ok := m[id]
	if !ok {
		var zero T
		return zero, fmt.Errorf("id=%d not found", id)
	}
	return v, nil
}

func All[T cfgPtr](r ScriptConfig) (map[int32]T, error) {
	return pickMap[T](r)
}

func pickMap[T cfgPtr](r ScriptConfig) (map[int32]T, error) {
	var zero T
	switch any(zero).(type) {
	case *v1.AbilityConfig:
		return any(r.GetAbilities()).(map[int32]T), nil
	case *v1.CardConfig:
		return any(r.GetCards()).(map[int32]T), nil
	case *v1.CharacterConfig:
		return any(r.GetCharacters()).(map[int32]T), nil
	case *v1.IncidentConfig:
		return any(r.GetIncidents()).(map[int32]T), nil
	default:
		return nil, fmt.Errorf("unsupported config type")
	}
}
