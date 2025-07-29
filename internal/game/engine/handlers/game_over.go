package handlers

import (
	model "tragedylooper/internal/game/proto/v1"
)

// GameOverHandler handles the GameOverEvent.
type GameOverHandler struct{}

// Handle currently does nothing, as this event is informational.
func (h *GameOverHandler) Handle(state *model.GameState, event *model.GameEvent) error {
	// No state change, this is for logging/notification purposes.
	return nil
}
