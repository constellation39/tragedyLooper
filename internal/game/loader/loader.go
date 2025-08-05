package loader

import (
	"fmt"
	"os"
	"path/filepath"

	v1 "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ScriptConfig defines the interface for accessing all loaded game configuration data.
type ScriptConfig interface {
	GetScript() *v1.ScriptConfig

	PrivateInfo() *v1.PrivateInfo
	PublicInfo() *v1.PublicInfo

	GetPlot(id int32) *v1.PlotConfig
	GetPlotMap() map[int32]*v1.PlotConfig
	GetRole(id int32) *v1.RoleConfig
	GetRoleMap() map[int32]*v1.RoleConfig
	GetCard(id int32) *v1.CardConfig
	GetCardMap() map[int32]*v1.CardConfig
	GetAbility(id int32) *v1.AbilityConfig
	GetAbilityMap() map[int32]*v1.AbilityConfig

	GetMainPlot() *v1.PlotConfig
	GetSubPlot(id int32) *v1.PlotConfig
	GetSubPlotMap() map[int32]*v1.PlotConfig
	GetCharacter(id int32) *v1.CharacterConfig
	GetCharacterMap() map[int32]*v1.CharacterConfig
	GetIncident(id int32) *v1.IncidentConfig
	GetIncidentMap() map[int32]*v1.IncidentConfig
	GetLoopCount() int32
	GetDaysPerLoop() int32
	GetCanDiscuss() bool
}

// scriptConfig is the concrete implementation of the ScriptConfig interface.
// It holds all the loaded script and library data.
type scriptConfig struct {
	script  *v1.ScriptConfig
	modelId int32
}

// newRepository creates a new, empty scriptConfig repository.
func newRepository(modelId int32) *scriptConfig {
	return &scriptConfig{
		modelId: modelId,
		script:  &v1.ScriptConfig{},
	}
}

// LoadConfig loads all game data from the specified directory and script.
func LoadConfig(dataDir, scriptID string, modelId int32) (ScriptConfig, error) {
	repo := newRepository(modelId)

	// Load the main script file.
	scriptPath := filepath.Join(dataDir, "scripts", scriptID+".json")
	if err := loadDataFromJSON(scriptPath, repo.script); err != nil {
		return nil, fmt.Errorf("failed to load script '%s': %w", scriptID, err)
	}

	// After loading, validation would be performed here.
	// if err := repo.script.Validate(); err != nil {
	// 	return nil, fmt.Errorf("script validation failed: %w", err)
	// }

	return repo, nil
}

// loadDataFromJSON is a generic function that loads and decodes data from a JSON file
// into a given protocol buffer message.
func loadDataFromJSON(filePath string, data proto.Message) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %w", filePath, err)
	}

	jsonBytes, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to read JSON file %s: %w", absPath, err)
	}

	// Using protojson to unmarshal is safer for protobuf messages.
	if err := protojson.Unmarshal(jsonBytes, data); err != nil {
		return fmt.Errorf("failed to unmarshal JSON from %s: %w", absPath, err)
	}

	return nil
}

func (s *scriptConfig) GetScript() *v1.ScriptConfig {
	return s.script
}

func (s *scriptConfig) getModel() *v1.ScriptModel {
	return s.script.GetScriptModels()[s.modelId]
}

func (s *scriptConfig) PrivateInfo() *v1.PrivateInfo {
	if model := s.getModel(); model != nil {
		return model.GetPrivateInfo()
	}
	return nil
}

func (s *scriptConfig) PublicInfo() *v1.PublicInfo {
	if model := s.getModel(); model != nil {
		return model.GetPublicInfo()
	}
	return nil
}

func (s *scriptConfig) GetPlot(id int32) *v1.PlotConfig {
	if plot, ok := s.script.GetMainPlots()[id]; ok {
		return plot
	}
	if plot, ok := s.script.GetSubPlots()[id]; ok {
		return plot
	}
	return nil
}

func (s *scriptConfig) GetPlotMap() map[int32]*v1.PlotConfig {
	plots := make(map[int32]*v1.PlotConfig)
	for id, plot := range s.script.GetMainPlots() {
		plots[id] = plot
	}
	for id, plot := range s.script.GetSubPlots() {
		plots[id] = plot
	}
	return plots
}

func (s *scriptConfig) GetRole(id int32) *v1.RoleConfig {
	return s.script.GetRoles()[id]
}

func (s *scriptConfig) GetRoleMap() map[int32]*v1.RoleConfig {
	return s.script.GetRoles()
}

func (s *scriptConfig) GetCard(id int32) *v1.CardConfig {
	if card, ok := s.script.GetMastermindCards()[id]; ok {
		return card
	}
	if card, ok := s.script.GetProtagonistCards()[id]; ok {
		return card
	}
	return nil
}

func (s *scriptConfig) GetCardMap() map[int32]*v1.CardConfig {
	cards := make(map[int32]*v1.CardConfig)
	for id, card := range s.script.GetMastermindCards() {
		cards[id] = card
	}
	for id, card := range s.script.GetProtagonistCards() {
		cards[id] = card
	}
	return cards
}

func (s *scriptConfig) GetAbility(id int32) *v1.AbilityConfig {
	for _, role := range s.script.GetRoles() {
		if ability, ok := role.GetAbilities()[id]; ok {
			return ability
		}
	}
	return nil
}

func (s *scriptConfig) GetAbilityMap() map[int32]*v1.AbilityConfig {
	abilities := make(map[int32]*v1.AbilityConfig)
	for _, role := range s.script.GetRoles() {
		for id, ability := range role.GetAbilities() {
			abilities[id] = ability
		}
	}
	return abilities
}

func (s *scriptConfig) GetMainPlot() *v1.PlotConfig {
	if privateInfo := s.PrivateInfo(); privateInfo != nil {
		return s.GetPlot(privateInfo.GetMainPlotId())
	}
	return nil
}

func (s *scriptConfig) GetSubPlot(id int32) *v1.PlotConfig {
	return s.script.GetSubPlots()[id]
}

func (s *scriptConfig) GetSubPlotMap() map[int32]*v1.PlotConfig {
	privateInfo := s.PrivateInfo()
	if privateInfo == nil {
		return nil
	}
	subplots := make(map[int32]*v1.PlotConfig)
	for _, id := range privateInfo.GetSubPlotsIds() {
		if plot, ok := s.script.GetSubPlots()[id]; ok {
			subplots[id] = plot
		}
	}
	return subplots
}

func (s *scriptConfig) GetCharacter(id int32) *v1.CharacterConfig {
	return s.script.GetCharacters()[id]
}

func (s *scriptConfig) GetCharacterMap() map[int32]*v1.CharacterConfig {
	return s.script.GetCharacters()
}

func (s *scriptConfig) GetIncident(id int32) *v1.IncidentConfig {
	return s.script.GetIncidents()[id]
}

func (s *scriptConfig) GetIncidentMap() map[int32]*v1.IncidentConfig {
	return s.script.GetIncidents()
}

func (s *scriptConfig) GetLoopCount() int32 {
	if publicInfo := s.PublicInfo(); publicInfo != nil {
		return publicInfo.GetLoopCount()
	}
	return 0
}

func (s *scriptConfig) GetDaysPerLoop() int32 {
	if publicInfo := s.PublicInfo(); publicInfo != nil {
		return publicInfo.GetDaysPerLoop()
	}
	return 0
}

func (s *scriptConfig) GetCanDiscuss() bool {
	if publicInfo := s.PublicInfo(); publicInfo != nil {
		return publicInfo.GetCanDiscuss()
	}
	return false
}

// Global access functions

type cfgPtr interface {
	*v1.AbilityConfig |
		*v1.CardConfig |
		*v1.CharacterConfig |
		*v1.IncidentConfig |
		*v1.PlotConfig |
		*v1.RoleConfig
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
		return any(r.GetAbilityMap()).(map[int32]T), nil
	case *v1.CardConfig:
		return any(r.GetCardMap()).(map[int32]T), nil
	case *v1.CharacterConfig:
		return any(r.GetCharacterMap()).(map[int32]T), nil
	case *v1.IncidentConfig:
		return any(r.GetIncidentMap()).(map[int32]T), nil
	case *v1.PlotConfig:
		return any(r.GetPlotMap()).(map[int32]T), nil
	case *v1.RoleConfig:
		return any(r.GetRoleMap()).(map[int32]T), nil
	default:
		var t T
		return nil, fmt.Errorf("unsupported config type: %T", t)
	}
}
