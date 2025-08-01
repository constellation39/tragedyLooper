package eventhandler

import (
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

func init() {
	Register(model.GameEventType_GAME_EVENT_TYPE_LOOP_WIN, &LoopWinHandler{})
}

// LoopWinHandler handles the LoopWinEvent.
type LoopWinHandler struct{}

// Handle currently does nothing, as this event is informational.
func (h *LoopWinHandler) Handle(ge GameEngine, event *model.GameEvent) error {
	// No state change, this is for logging/notification purposes.
	return nil
}
