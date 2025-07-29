package handlers

import (
	model "tragedylooper/internal/game/proto/v1"
)

// IntrigueAdjustedHandler handles the IntrigueAdjustedEvent.
type IntrigueAdjustedHandler struct{}

// Handle updates the character's intrigue in the game state.
func (h *IntrigueAdjustedHandler) Handle(state *model.GameState, event *model.GameEvent) error {
	var e model.IntrigueAdjustedEvent
	if err := event.Payload.UnmarshalTo(&e); err != nil {
		return err
	}

	if char, ok := state.Characters[e.CharacterId]; ok {
		char.Intrigue += e.Amount
		e.NewIntrigue = char.Intrigue
	}
	return nil
}
