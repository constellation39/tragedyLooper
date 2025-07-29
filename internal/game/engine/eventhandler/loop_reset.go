package eventhandler

import (
	model "tragedylooper/pkg/proto/v1"
)

// LoopResetHandler handles the LoopResetEvent.
type LoopResetHandler struct{}

// Handle clears the loop's events from the game state.
func (h *LoopResetHandler) Handle(state *model.GameState, event *model.GameEvent) error {
	state.LoopEvents = []*model.GameEvent{}
	return nil
}
