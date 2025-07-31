package phasehandler

import (
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

// --- SetupPhase ---
type SetupPhase struct{ basePhase }

func (p *SetupPhase) Type() model.GamePhase { return model.GamePhase_GAME_PHASE_SETUP }
func (p *SetupPhase) Enter(ge GameEngine) Phase {
	// TODO: Implement logic for Mastermind to choose sub-scenario and place characters.
	// For now, we transition directly.
	return GetPhase(model.GamePhase_GAME_PHASE_MASTERMIND_SETUP)
}

func init() {
	RegisterPhase(&SetupPhase{})
}
