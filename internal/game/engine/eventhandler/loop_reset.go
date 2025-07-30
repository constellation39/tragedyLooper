package eventhandler

import (
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

func init() {
	Register(model.GameEventType_GAME_EVENT_TYPE_LOOP_RESET, &LoopResetHandler{})
}

// LoopResetHandler handles the LoopResetEvent.
type LoopResetHandler struct{}

// Handle clears the loop's events from the game state.
func (h *LoopResetHandler) Handle(ge GameEngine, event *model.GameEvent) error {
	state := ge.GetGameState()
	state.LoopEvents = []*model.GameEvent{}
	return nil
}
