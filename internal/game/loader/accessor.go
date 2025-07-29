package loader

import v1 "tragedylooper/pkg/proto/v1"

// GameDataAccessor defines the interface for accessing game configuration data.
// This allows the game engine to be decoupled from the concrete data loader implementation,
// facilitating easier testing with mock data.
type GameDataAccessor interface {
	GetScript() *v1.ScriptConfig

	GetAbilities() map[int32]*v1.AbilityConfig
	GetCards() map[int32]*v1.CardConfig
	GetCharacters() map[int32]*v1.CharacterConfig
	GetIncidents() map[int32]*v1.IncidentConfig
}
