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
	gs := im.engine.GetGameState()

	for _, incidentConfig := range incidents {
		if _, triggered := gs.TriggeredIncidents[incidentConfig.GetName()]; triggered {
			continue
		}

		conditionsMet := true
		for _, condition := range incidentConfig.GetTriggerConditions() {
			met, err := im.engine.CheckCondition(condition)
			if err != nil {
				logger.Error("Error checking incident condition", zap.String("incident", incidentConfig.GetName()), zap.Error(err))
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

		// logger.Info("Incident triggered", zap.String("incident", incidentConfig.GetName()))
		gs.TriggeredIncidents[incidentConfig.GetName()] = true

		// Publish the trigger event. The engine's main loop will handle applying the effect.
		im.engine.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_INCIDENT_TRIGGERED, &model.EventPayload{
			Payload: &model.EventPayload_IncidentTriggered{IncidentTriggered: &model.IncidentTriggeredEvent{Incident: &model.Incident{Config: incidentConfig}}},
		})
	}
}
