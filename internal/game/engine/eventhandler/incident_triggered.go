package eventhandler

import (
	model "tragedylooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

func init() {
	Register(model.GameEventType_GAME_EVENT_TYPE_INCIDENT_TRIGGERED, &IncidentTriggeredHandler{})
}

// IncidentTriggeredHandler handles the IncidentTriggeredEvent.
type IncidentTriggeredHandler struct{}

// Handle applies the effects of the triggered incident to the game state.
func (h *IncidentTriggeredHandler) Handle(ge GameEngine, event *model.GameEvent) error {
	payload, ok := event.Payload.Payload.(*model.EventPayload_IncidentTriggered)
	if !ok {
		return nil // Or return an error if this is unexpected
	}

	incident := payload.IncidentTriggered.Incident
	if incident == nil || incident.Config == nil {
		return nil // Or return an error
	}

	ge.Logger().Info("Applying effects for triggered incident", zap.String("incident", incident.Config.Name))

	if incident.Config.Effect != nil {
		if err := ge.ApplyEffect(incident.Config.Effect, nil, nil, nil); err != nil {
			ge.Logger().Error("Error applying incident effect",
				zap.String("incident", incident.Config.Name),
				zap.Error(err),
			)
			// Decide if we should continue or stop on error
		}
	}

	return nil
}
