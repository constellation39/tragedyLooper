package eventhandler

import (
	model "tragedylooper/pkg/proto/v1"
)

func init() {
	Register(model.GameEventType_CARD_REVEALED, &CardRevealedHandler{})
}

// CardRevealedHandler handles the CardRevealedEvent.
type CardRevealedHandler struct{}

// Handle currently does nothing, as this event is informational.
func (h *CardRevealedHandler) Handle(state *model.GameState, event *model.EventPayload) error {
	// No state change, this is for logging/notification purposes.
	return nil
}
