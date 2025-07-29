package effecthandler

import (
	model "tragedylooper/pkg/proto/v1"
)

// GameEngine provides the necessary methods for handlers to interact with the game state and engine logic.
// This interface helps to decouple the handlers from the main engine package.
type GameEngine interface {
	GetGameState() *model.GameState
	ApplyAndPublishEvent(eventType model.GameEventType, payload *model.EventPayload)
	ResolveSelectorToCharacters(gs *model.GameState, sel *model.TargetSelector, ctx *EffectContext) ([]int32, error)
	GetCharacterByID(id int32) *model.Character
	MoveCharacter(char *model.Character, dx, dy int)
}

// EffectContext provides contextual information for effect resolution and application.
type EffectContext struct {
	Ability *model.Ability
	Payload *model.UseAbilityPayload
	Choice  *model.ChooseOptionPayload
}

// EffectHandler defines the interface for processing a specific type of game effect.
type EffectHandler interface {
	// ResolveChoices checks if the effect requires a player choice and returns the available options.
	ResolveChoices(ge GameEngine, effect *model.Effect, ctx *EffectContext) ([]*model.Choice, error)

	// Apply executes the effect's logic, applying state changes and publishing events.
	Apply(ge GameEngine, effect *model.Effect, ctx *EffectContext) error

	// GetDescription returns a human-readable description of the effect.
	GetDescription(effect *model.Effect) string
}
