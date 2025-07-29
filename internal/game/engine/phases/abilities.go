package phases

import (
	model "tragedylooper/internal/game/proto/v1"
)

// --- AbilitiesPhase ---
type AbilitiesPhase struct{ basePhase }

func (p *AbilitiesPhase) Type() model.GamePhase { return model.GamePhase_ABILITIES }
func (p *AbilitiesPhase) Enter(ge GameEngine) Phase {
	// Players can use abilities.
	// This phase might require player input and have a timeout.
	return &IncidentsPhase{}
}
