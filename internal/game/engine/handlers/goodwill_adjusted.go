package handlers

import (
	model "tragedylooper/internal/game/proto/v1"
)

// GoodwillAdjustedHandler handles the GoodwillAdjustedEvent.
type GoodwillAdjustedHandler struct{}

// Handle updates the character's goodwill in the game state.
func (h *GoodwillAdjustedHandler) Handle(state *model.GameState, event *model.GameEvent) error {
	var e model.GoodwillAdjustedEvent
	if err := event.Payload.UnmarshalTo(&e); err != nil {
		return err
	}

	if char, ok := state.Characters[e.CharacterId]; ok {
		char.Goodwill += e.Amount
		e.NewGoodwill = char.Goodwill
	}
	return nil
}
