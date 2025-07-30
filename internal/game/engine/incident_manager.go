package engine

import (
	model "tragedylooper/pkg/proto/v1"

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

		allConditionsMet := true
		for _, condition := range incident.GetConfig().GetTriggerConditions() {
			met, err := im.engine.CheckCondition(condition)
			if err != nil {
				logger.Error("Error checking incident condition", zap.String("incident", incident.GetConfig().GetName()), zap.Error(err))
				allConditionsMet = false
				break
			}
			if !met {
				allConditionsMet = false
				break
			}
		}

		if allConditionsMet {
			logger.Info("Incident triggered", zap.String("incident", incident.GetConfig().GetName()))
			incident.HasTriggeredThisLoop = true

			// 发布触发事件
			im.engine.ApplyAndPublishEvent(model.GameEventType_INCIDENT_TRIGGERED, &model.EventPayload{
				Payload: &model.EventPayload_IncidentTriggered{IncidentTriggered: &model.IncidentTriggeredEvent{Incident: incident}},
			})

			// 应用事件效果
			if incident.GetConfig().GetEffect() != nil {
				if err := im.engine.ApplyEffect(incident.GetConfig().GetEffect(), nil, nil, nil); err != nil {
					logger.Error("Error applying incident effect", zap.String("incident", incident.GetConfig().GetName()), zap.Error(err))
				}
			}
		}
	}
}
