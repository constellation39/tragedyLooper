package eventhandler

import (
	model "tragedylooper/pkg/proto/v1"
)

func init() {
	Register(model.GameEventType_GAME_ENDED, &GameOverHandler{})
}

// GameOverHandler handles the GameOverEvent.
type GameOverHandler struct{}

// Handle currently does nothing, as this event is informational.
func (h *GameOverHandler) Handle(state *model.GameState, event *model.EventPayload) error {
	// No state change, this is for logging/notification purposes.
	return nil
}
