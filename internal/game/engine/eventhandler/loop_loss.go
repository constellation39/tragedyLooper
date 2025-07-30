package eventhandler

import (
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

func init() {
	Register(model.GameEventType_GAME_EVENT_TYPE_LOOP_LOSS, &LoopLossHandler{})
}

// LoopLossHandler handles the LoopLossEvent.
type LoopLossHandler struct{}

// Handle currently does nothing, as this event is informational.
func (h *LoopLossHandler) Handle(ge GameEngine, event *model.GameEvent) error {
	// No state change, this is for logging/notification purposes.
	return nil
}
