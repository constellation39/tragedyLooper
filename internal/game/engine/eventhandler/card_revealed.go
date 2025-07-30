package eventhandler

import (
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

func init() {
	Register(model.GameEventType_CARD_REVEALED, &CardRevealedHandler{})
}

// CardRevealedHandler handles the CardRevealedEvent.
type CardRevealedHandler struct{}

// Handle currently does nothing, as this event is informational.
func (h *CardRevealedHandler) Handle(ge GameEngine, event *model.GameEvent) error {
	// No state change, this is for logging/notification purposes.
	return nil
}
