package phase

import (
	model "tragedylooper/pkg/proto/v1"
)

// --- CardEffectsPhase ---
type CardEffectsPhase struct{ basePhase }

func (p *CardEffectsPhase) Type() model.GamePhase { return model.GamePhase_CARD_EFFECTS }

func (p *CardEffectsPhase) Enter(ge GameEngine) Phase {
	// Here you would implement the logic to resolve the effects of all other cards
	// that are not related to movement.
	//
	// For now, this is a placeholder.

	// After card effects are resolved, we might move to the abilities phase.
	return &AbilitiesPhase{}
}
