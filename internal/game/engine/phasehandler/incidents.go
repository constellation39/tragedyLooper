package phasehandler

import (
	"time"
	model "tragedylooper/pkg/proto/tragedylooper/v1"

	"go.uber.org/zap"
)

// IncidentsPhase is the phase where incident conditions are checked and triggered.
type IncidentsPhase struct{}

// HandleEvent is the default implementation for Phase interface, does nothing and returns nil.
func (p *IncidentsPhase) HandleEvent(ge GameEngine, event *model.GameEvent) Phase {
	return nil
}

// HandleTimeout is the default implementation for Phase interface, returns the next phase.
func (p *IncidentsPhase) HandleTimeout(ge GameEngine) Phase {
	ge.Logger().Info("IncidentsPhase timed out, moving to DayEnd")
	return GetPhase(model.GamePhase_GAME_PHASE_DAY_END)
}

// Exit is the default implementation for Phase interface, does nothing.
func (p *IncidentsPhase) Exit(ge GameEngine) {}

// TimeoutDuration is the default implementation for Phase interface, returns a short duration.
func (p *IncidentsPhase) TimeoutDuration() time.Duration { return 5 * time.Second }

// Type returns the phase type.
func (p *IncidentsPhase) Type() model.GamePhase { return model.GamePhase_GAME_PHASE_INCIDENTS }

// Enter is called when the phase begins.
func (p *IncidentsPhase) Enter(ge GameEngine) Phase {
	ge.Logger().Info("Entering IncidentsPhase, checking for incidents.")
	p.triggerIncidents(ge)

	// The phase will transition via timeout, allowing time for any triggered events
	// (and subsequent player choices) to be processed.
	return nil
}

// HandleAction handles actions during the incidents phase, primarily for choices.
func (p *IncidentsPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	if payload, ok := action.Payload.(*model.PlayerActionPayload_ChooseOption); ok {
		ge.Logger().Info("Player made a choice during IncidentsPhase", zap.Any("choice", payload))
		// The engine's effect handler is responsible for continuing the effect chain.
	}
	return nil
}

func (p *IncidentsPhase) triggerIncidents(ge GameEngine) {
	logger := ge.Logger().Named("TriggerIncidents")
	incidents := ge.GetGameRepo().GetIncidents()
	gs := ge.GetGameState()

	for _, incidentConfig := range incidents {
		if _, triggered := gs.TriggeredIncidents[incidentConfig.GetName()]; triggered {
			continue
		}

		conditionsMet := true
		for _, condition := range incidentConfig.GetTriggerConditions() {
			met, err := ge.CheckCondition(condition)
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

		gs.TriggeredIncidents[incidentConfig.GetName()] = true

		ge.TriggerEvent(model.GameEventType_GAME_EVENT_TYPE_INCIDENT_TRIGGERED, &model.EventPayload{
			Payload: &model.EventPayload_IncidentTriggered{IncidentTriggered: &model.IncidentTriggeredEvent{Incident: &model.Incident{Config: incidentConfig}}},
		})
	}
}

func init() {
	RegisterPhase(&IncidentsPhase{})
}
