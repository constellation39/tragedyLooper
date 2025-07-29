package eventhandler

import (
	model "tragedylooper/pkg/proto/v1"
)

func init() {
	Register(model.GameEventType_LOOP_RESET, &LoopResetHandler{})
}

// LoopResetHandler handles the LoopResetEvent.
type LoopResetHandler struct{}

// Handle clears the loop's events from the game state.
func (h *LoopResetHandler) Handle(state *model.GameState, event *model.EventPayload) error {
	state.LoopEvents = []*model.GameEvent{}
	return nil
}
