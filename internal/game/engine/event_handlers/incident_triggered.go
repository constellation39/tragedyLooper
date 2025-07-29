package handlers

import (
	model "tragedylooper/internal/game/proto/v1"
)

// IncidentTriggeredHandler handles the IncidentTriggeredEvent.
type IncidentTriggeredHandler struct{}

// Handle currently does nothing, as this event is informational.
func (h *IncidentTriggeredHandler) Handle(state *model.GameState, event *model.GameEvent) error {
	// No state change, this is for logging/notification purposes.
	return nil
}
