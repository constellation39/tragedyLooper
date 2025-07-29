package eventhandler

import (
	model "tragedylooper/pkg/proto/v1"
)

func init() {
	Register(model.GameEventType_PARANOIA_ADJUSTED, &ParanoiaAdjustedHandler{})
}

// ParanoiaAdjustedHandler handles the ParanoiaAdjustedEvent.
type ParanoiaAdjustedHandler struct{}

// Handle updates the character's paranoia in the game state.
func (h *ParanoiaAdjustedHandler) Handle(state *model.GameState, event *model.EventPayload) error {
	e, ok := event.Payload.(*model.EventPayload_ParanoiaAdjusted)
	if !ok {
		return nil // Or handle error appropriately
	}

	if char, ok := state.Characters[e.ParanoiaAdjusted.CharacterId]; ok {
		char.Paranoia += e.ParanoiaAdjusted.Amount
		// The event payload is updated to reflect the new value, though this is a side effect.
		// Consider if this is the desired behavior.
		e.ParanoiaAdjusted.NewParanoia = char.Paranoia
	}
	return nil
}
