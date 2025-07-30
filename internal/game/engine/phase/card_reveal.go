package phase

import (
	model "tragedylooper/pkg/proto/v1"
)

// --- CardRevealPhase ---
type CardRevealPhase struct{ basePhase }

func (p *CardRevealPhase) Type() model.GamePhase { return model.GamePhase_CARD_REVEAL }
func (p *CardRevealPhase) Enter(ge GameEngine) Phase {
	// Reveal all cards played this turn.
	allPlayedCards := make(map[int32]*model.CardList)
	for playerID, cardList := range ge.GetGameState().PlayedCardsThisDay {
		allPlayedCards[playerID] = cardList
	}
	ge.ApplyAndPublishEvent(model.GameEventType_CARD_REVEALED, &model.EventPayload{
		Payload: &model.EventPayload_CardRevealed{CardRevealed: &model.CardRevealedEvent{Cards: allPlayedCards}},
	})
	return &CardResolvePhase{}
}
