package handlers

import (
	model "tragedylooper/pkg/proto/v1"
)

// LoopWinHandler handles the LoopWinEvent.
type LoopWinHandler struct{}

// Handle currently does nothing, as this event is informational.
func (h *LoopWinHandler) Handle(state *model.GameState, event *model.GameEvent) error {
	// No state change, this is for logging/notification purposes.
	return nil
}
