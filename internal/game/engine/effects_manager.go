package engine

import (
	"tragedylooper/internal/game/engine/effecthandler"
	model "tragedylooper/pkg/proto/v1"
)

// GetEffectDescription finds the appropriate handler and returns the description for an effect.
// effect: The effect to get the description for.
// Returns: The description string for the effect.
func (ge *GameEngine) GetEffectDescription(effect *model.Effect) string {
	return effecthandler.GetEffectDescription(ge, effect)
}