package eventhandler

import (
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

func init() {
	Register(model.GameEventType_GAME_EVENT_TYPE_CARD_PLAYED, &CardPlayedHandler{})
}

// CardPlayedHandler handles the CardPlayedEvent.
type CardPlayedHandler struct{}

// Handle currently does nothing, as this event is informational.
func (h *CardPlayedHandler) Handle(ge GameEngine, event *model.GameEvent) error {
	// No state change, this is for logging/notification purposes.
	// The actual card effect is resolved in a later phasehandler.
	return nil
}
