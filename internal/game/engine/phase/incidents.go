package phase

import (
	model "tragedylooper/pkg/proto/v1"
)

// --- IncidentsPhase ---
type IncidentsPhase struct{ basePhase }

func (p *IncidentsPhase) Type() model.GamePhase { return model.GamePhase_INCIDENTS }
func (p *IncidentsPhase) Enter(ge GameEngine) Phase {
	ge.TriggerIncidents()
	return &DayEndPhase{}
}
