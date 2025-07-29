package eventhandler

import (
	model "tragedylooper/pkg/proto/v1"
)

func init() {
	Register(model.GameEventType_LOOP_LOSS, &LoopLossHandler{})
}

// LoopLossHandler handles the LoopLossEvent.
type LoopLossHandler struct{}

// Handle currently does nothing, as this event is informational.
func (h *LoopLossHandler) Handle(state *model.GameState, event *model.EventPayload) error {
	// No state change, this is for logging/notification purposes.
	return nil
}
