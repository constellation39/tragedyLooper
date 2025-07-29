package eventhandler

import (
	model "tragedylooper/pkg/proto/v1"
)

func init() {
	Register(model.GameEventType_GOODWILL_ADJUSTED, &GoodwillAdjustedHandler{})
}

// GoodwillAdjustedHandler handles the GoodwillAdjustedEvent.
type GoodwillAdjustedHandler struct{}

// Handle updates the character's goodwill in the game state.
func (h *GoodwillAdjustedHandler) Handle(state *model.GameState, event *model.EventPayload) error {
	e, ok := event.Payload.(*model.EventPayload_GoodwillAdjusted)
	if !ok {
		return nil // Or handle error appropriately
	}

	if char, ok := state.Characters[e.GoodwillAdjusted.CharacterId]; ok {
		char.Goodwill += e.GoodwillAdjusted.Amount
		e.GoodwillAdjusted.NewGoodwill = char.Goodwill
	}
	return nil
}
