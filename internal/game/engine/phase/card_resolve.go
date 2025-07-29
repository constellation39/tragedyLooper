package phase

import (
	model "tragedylooper/pkg/proto/v1"
)

// --- CardResolvePhase ---
type CardResolvePhase struct{ basePhase }

func (p *CardResolvePhase) Type() model.GamePhase { return model.GamePhase_CARD_RESOLVE }
func (p *CardResolvePhase) Enter(ge GameEngine) Phase {
	ge.ResolveMovement()
	ge.ResolveOtherCards()
	return &AbilitiesPhase{}
}
