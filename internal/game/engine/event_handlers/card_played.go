package handlers

import (
	model "tragedylooper/pkg/proto/v1"
)

// CardPlayedHandler handles the CardPlayedEvent.
type CardPlayedHandler struct{}

// Handle currently does nothing, as this event is informational.
func (h *CardPlayedHandler) Handle(state *model.GameState, event *model.GameEvent) error {
	// No state change, this is for logging/notification purposes.
	// The actual card effect is resolved in a later phase.
	return nil
}
