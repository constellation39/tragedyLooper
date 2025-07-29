package phases

import (
	model "tragedylooper/internal/game/proto/v1"
)

// --- IncidentsPhase ---
type IncidentsPhase struct{ basePhase }

func (p *IncidentsPhase) Type() model.GamePhase { return model.GamePhase_INCIDENTS }
func (p *IncidentsPhase) Enter(ge GameEngine) Phase {
	ge.TriggerIncidents()
	return &DayEndPhase{}
}
