package phase

import (
	model "tragedylooper/pkg/proto/v1"
)

// --- SetupPhase ---
type SetupPhase struct{ basePhase }

func (p *SetupPhase) Type() model.GamePhase { return model.GamePhase_SETUP }
func (p *SetupPhase) Enter(ge GameEngine) Phase {
	// TODO: Implement logic for Mastermind to choose sub-scenario and place characters.
	// For now, we transition directly.
	return &MastermindSetupPhase{}
}
