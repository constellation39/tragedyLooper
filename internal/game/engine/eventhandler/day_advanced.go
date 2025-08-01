package eventhandler

import (
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

func init() {
	Register(model.GameEventType_GAME_EVENT_TYPE_DAY_ADVANCED, &DayAdvancedHandler{})
}

// DayAdvancedHandler handles the DayAdvancedEvent.
type DayAdvancedHandler struct{}

// Handle clears the day's events from the game state.
func (h *DayAdvancedHandler) Handle(ge GameEngine, event *model.GameEvent) error {
	state := ge.GetGameState()
	state.DayEvents = []*model.GameEvent{}
	return nil
}
