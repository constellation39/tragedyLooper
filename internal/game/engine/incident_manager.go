package engine

import (
	model "tragedylooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

type incidentManager struct {
	engine *GameEngine
}

func newIncidentManager(engine *GameEngine) *incidentManager {
	return &incidentManager{
		engine: engine,
	}
}

func (im *incidentManager) TriggerIncidents() {
	logger := im.engine.logger.Named("TriggerIncidents")
	incidents := im.engine.gameConfig.GetIncidents()

	for _, incidentConfig := range incidents {
		incident := &model.Incident{Config: incidentConfig}
		if incident.GetHasTriggeredThisLoop() {
			continue
		}

		conditionsMet := true
		for _, condition := range incident.GetConfig().GetTriggerConditions() {
			met, err := im.engine.CheckCondition(condition)
			if err != nil {
				logger.Error("Error checking incident condition", zap.String("incident", incident.GetConfig().GetName()), zap.Error(err))
				conditionsMet = false
				break
			}
			if !met {
				conditionsMet = false
				break
			}
		}

		if !conditionsMet {
			continue
		}

		logger.Info("Incident triggered", zap.String("incident", incident.GetConfig().GetName()))
		incident.HasTriggeredThisLoop = true

		// Publish the trigger event. The engine's main loop will handle applying the effect.
		im.engine.ApplyAndPublishEvent(model.GameEventType_GAME_EVENT_TYPE_INCIDENT_TRIGGERED, &model.EventPayload{
			Payload: &model.EventPayload_IncidentTriggered{IncidentTriggered: &model.IncidentTriggeredEvent{Incident: incident}},
		})
	}
}
