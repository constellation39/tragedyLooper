package phase

import (
	model "tragedylooper/pkg/proto/v1"
)

// --- MastermindSetupPhase ---
type MastermindSetupPhase struct{ basePhase }

func (p *MastermindSetupPhase) Type() model.GamePhase { return model.GamePhase_MASTERMIND_SETUP }
func (p *MastermindSetupPhase) Enter(ge GameEngine) Phase {
	// TODO: Mastermind places characters and sets up their board.
	// For now, we transition directly.
	return &LoopStartPhase{}
}
