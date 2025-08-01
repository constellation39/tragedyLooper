package phasehandler

import (
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

// --- MastermindSetupPhase ---
type MastermindSetupPhase struct {
	BasePhase
}

func (p *MastermindSetupPhase) Type() model.GamePhase {
	return model.GamePhase_GAME_PHASE_MASTERMIND_SETUP
}
func (p *MastermindSetupPhase) Enter(ge GameEngine) {
	// TODO: Mastermind places characters and sets up their board.
	// For now, we transition directly.
}

func init() {
	RegisterPhase(&MastermindSetupPhase{})
}
