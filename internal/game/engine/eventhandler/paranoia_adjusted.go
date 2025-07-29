package handlers

import (
	model "tragedylooper/pkg/proto/v1"
)

// ParanoiaAdjustedHandler handles the ParanoiaAdjustedEvent.
type ParanoiaAdjustedHandler struct{}

// Handle updates the character's paranoia in the game state.
func (h *ParanoiaAdjustedHandler) Handle(state *model.GameState, event *model.GameEvent) error {
	var e model.ParanoiaAdjustedEvent
	if err := event.Payload.UnmarshalTo(&e); err != nil {
		return err
	}

	if char, ok := state.Characters[e.CharacterId]; ok {
		char.Paranoia += e.Amount
		// The event payload is updated to reflect the new value, though this is a side effect.
		// Consider if this is the desired behavior.
		e.NewParanoia = char.Paranoia
	}
	return nil
}
