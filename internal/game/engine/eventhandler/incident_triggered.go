package eventhandler

import (
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

func init() {
	Register(model.GameEventType_INCIDENT_TRIGGERED, &IncidentTriggeredHandler{})
}

// IncidentTriggeredHandler handles the IncidentTriggeredEvent.
type IncidentTriggeredHandler struct{}

// Handle currently does nothing, as this event is informational.
func (h *IncidentTriggeredHandler) Handle(ge GameEngine, event *model.GameEvent) error {
	// No state change, this is for logging/notification purposes.
	return nil
}
