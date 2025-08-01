package eventhandler

import (
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

func init() {
	Register(model.GameEventType_GAME_EVENT_TYPE_GOODWILL_ADJUSTED, &GoodwillAdjustedHandler{})
}

// GoodwillAdjustedHandler handles the GoodwillAdjustedEvent.
type GoodwillAdjustedHandler struct{}

// Handle updates the character's goodwill in the game state.
func (h *GoodwillAdjustedHandler) Handle(ge GameEngine, event *model.GameEvent) error {
	e, ok := event.Payload.Payload.(*model.EventPayload_GoodwillAdjusted)
	if !ok {
		return nil // Or handle error appropriately
	}

	state := ge.GetGameState()
	if char, ok := state.Characters[e.GoodwillAdjusted.CharacterId]; ok {
		char.Goodwill += e.GoodwillAdjusted.Amount
		e.GoodwillAdjusted.NewGoodwill = char.Goodwill
	}
	return nil
}
