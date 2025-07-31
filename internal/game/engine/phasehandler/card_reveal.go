package phasehandler

import (
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

// CardRevealPhase 卡牌揭示阶段
type CardRevealPhase struct{}

// HandleAction is the default implementation for Phase interface, does nothing and returns nil.
func (p *CardRevealPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	return nil
}

// HandleEvent is the default implementation for Phase interface, does nothing and returns nil.
func (p *CardRevealPhase) HandleEvent(ge GameEngine, event *model.GameEvent) Phase { return nil }

// HandleTimeout is the default implementation for Phase interface, does nothing and returns nil.
func (p *CardRevealPhase) HandleTimeout(ge GameEngine) Phase { return nil }

// Exit is the default implementation for Phase interface, does nothing.
func (p *CardRevealPhase) Exit(ge GameEngine) {}

// TimeoutDuration is the default implementation for Phase interface, returns 0, indicating no timeout.
func (p *CardRevealPhase) TimeoutDuration() time.Duration { return 0 }


func (p *CardRevealPhase) Type() model.GamePhase { return model.GamePhase_GAME_PHASE_CARD_REVEAL }
func (p *CardRevealPhase) Enter(ge GameEngine) Phase {
	// 揭示本回合打出的所有牌。
	allPlayedCards := make(map[int32]*model.CardList)
	for playerID, cardList := range ge.GetGameState().PlayedCardsThisDay {
		allPlayedCards[playerID] = cardList
	}
	ge.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_CARD_REVEALED, &model.EventPayload{
		Payload: &model.EventPayload_CardRevealed{CardRevealed: &model.CardRevealedEvent{Cards: allPlayedCards}},
	})
	return GetPhase(model.GamePhase_GAME_PHASE_CARD_EFFECTS)
}

func init() {
	RegisterPhase(&CardRevealPhase{})
}
