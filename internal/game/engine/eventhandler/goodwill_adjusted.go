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
func (h *GoodwillAdjustedHandler) Handle(state *model.GameState, event *model.GameEvent) error {
	e := event.Payload.GetGoodwillAdjusted()
	if e == nil {
		return nil // Or handle error appropriately
	}

	if char, ok := state.Characters[e.CharacterId]; ok {
		char.Goodwill += e.Amount
		e.NewGoodwill = char.Goodwill
	}
	return nil
}
