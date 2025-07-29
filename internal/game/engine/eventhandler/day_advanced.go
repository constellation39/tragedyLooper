package eventhandler

import (
	model "tragedylooper/pkg/proto/v1"
)

// DayAdvancedHandler handles the DayAdvancedEvent.
type DayAdvancedHandler struct{}

// Handle clears the day's events from the game state.
func (h *DayAdvancedHandler) Handle(state *model.GameState, event *model.GameEvent) error {
	state.DayEvents = []*model.GameEvent{}
	return nil
}
