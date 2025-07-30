package eventhandler

import (
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

func init() {
	Register(model.GameEventType_INTRIGUE_ADJUSTED, &IntrigueAdjustedHandler{})
}

// IntrigueAdjustedHandler handles the IntrigueAdjustedEvent.
type IntrigueAdjustedHandler struct{}

// Handle updates the character's intrigue in the game state.
func (h *IntrigueAdjustedHandler) Handle(ge GameEngine, event *model.GameEvent) error {
	e, ok := event.Payload.Payload.(*model.EventPayload_IntrigueAdjusted)
	if !ok {
		return nil // Or handle error appropriately
	}

	state := ge.GetGameState()
	if char, ok := state.Characters[e.IntrigueAdjusted.CharacterId]; ok {
		char.Intrigue += e.IntrigueAdjusted.Amount
		e.IntrigueAdjusted.NewIntrigue = char.Intrigue
	}
	return nil
}
