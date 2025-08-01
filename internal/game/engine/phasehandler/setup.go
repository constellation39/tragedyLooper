package phasehandler

import (
	
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

// --- SetupPhase ---
type SetupPhase struct{
	BasePhase
}



func (p *SetupPhase) Type() model.GamePhase { return model.GamePhase_GAME_PHASE_SETUP }
func (p *SetupPhase) Enter(ge GameEngine) {
	// TODO: Implement logic for Mastermind to choose sub-scenario and place characters.
	// For now, we transition directly.
}

func init() {
	RegisterPhase(&SetupPhase{})
}
