package handlers

import (
	model "tragedylooper/internal/game/proto/v1"
)

// LoopLossHandler handles the LoopLossEvent.
type LoopLossHandler struct{}

// Handle currently does nothing, as this event is informational.
func (h *LoopLossHandler) Handle(state *model.GameState, event *model.GameEvent) error {
	// No state change, this is for logging/notification purposes.
	return nil
}
