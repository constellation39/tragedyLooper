package phase

import (
	model "tragedylooper/pkg/proto/v1"
)

// CardRevealPhase 卡牌揭示阶段
type CardRevealPhase struct{ basePhase }

func (p *CardRevealPhase) Type() model.GamePhase { return model.GamePhase_CARD_REVEAL }
func (p *CardRevealPhase) Enter(ge GameEngine) Phase {
	// 揭示本回合打出的所有牌。
	allPlayedCards := make(map[int32]*model.CardList)
	for playerID, cardList := range ge.GetGameState().PlayedCardsThisDay {
		allPlayedCards[playerID] = cardList
	}
	ge.ApplyAndPublishEvent(model.GameEventType_CARD_REVEALED, &model.EventPayload{
		Payload: &model.EventPayload_CardRevealed{CardRevealed: &model.CardRevealedEvent{Cards: allPlayedCards}},
	})
	return &CardResolvePhase{}
}