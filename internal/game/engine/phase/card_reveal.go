package phase

import (
	model "tragedylooper/pkg/proto/v1"
)

// --- CardRevealPhase ---
type CardRevealPhase struct{ basePhase }

func (p *CardRevealPhase) Type() model.GamePhase { return model.GamePhase_CARD_REVEAL }
func (p *CardRevealPhase) Enter(ge GameEngine) Phase {
	// Reveal all cards played this turn.
	ge.ApplyAndPublishEvent(model.GameEventType_CARD_REVEALED, &model.CardRevealedEvent{Cards: ge.GetGameState().PlayedCardsThisDay})
	return &CardResolvePhase{}
}
