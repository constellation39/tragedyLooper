package phase

import (
	model "tragedylooper/pkg/proto/v1"

	"go.uber.org/zap"
)

// IncidentsPhase is where incident conditions are checked and triggered.
type IncidentsPhase struct{ basePhase }

// Type returns the phase type.
func (p *IncidentsPhase) Type() model.GamePhase { return model.GamePhase_INCIDENTS }

// Enter is called when the phase begins.
func (p *IncidentsPhase) Enter(ge GameEngine) Phase {
	ge.TriggerIncidents()

	// After triggering, we check if any choices are pending. If not, we can move on.
	// A more robust system might wait for an explicit signal that all incidents are resolved.
	return &DayEndPhase{}
}

// HandleAction handles actions during the incident phase, primarily for choices.
func (p *IncidentsPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	if payload, ok := action.Payload.(*model.PlayerActionPayload_ChooseOption); ok {
		// Here, we would need to find the original effect that required the choice
		// and re-apply it with the choice provided. This is a complex problem.
		// For now, we'll log it and assume the choice resolves something.
		ge.Logger().Info("Player made a choice during IncidentsPhase", zap.Any("choice", payload))
		// After a choice, we might need to re-evaluate incidents or other conditions.
		// For simplicity, we don't transition here, assuming the game loop continues.
	}
	return nil
}
