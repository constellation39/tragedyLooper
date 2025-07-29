package eventhandler

import (
	model "tragedylooper/pkg/proto/v1"
)

func init() {
	Register(model.GameEventType_INTRIGUE_ADJUSTED, &IntrigueAdjustedHandler{})
}

// IntrigueAdjustedHandler handles the IntrigueAdjustedEvent.
type IntrigueAdjustedHandler struct{}

// Handle updates the character's intrigue in the game state.
func (h *IntrigueAdjustedHandler) Handle(state *model.GameState, event *model.GameEvent) error {
	e := event.Payload.GetIntrigueAdjusted()
	if e == nil {
		return nil // Or handle error appropriately
	}

	if char, ok := state.Characters[e.CharacterId]; ok {
		char.Intrigue += e.Amount
		e.NewIntrigue = char.Intrigue
	}
	return nil
}
