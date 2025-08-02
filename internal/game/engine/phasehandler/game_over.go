package phasehandler

import (
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

// --- GameOverPhase ---
type GameOverPhase struct {
	BasePhase
}

func (p *GameOverPhase) Type() model.GamePhase { return model.GamePhase_GAME_PHASE_GAME_OVER }
func (p *GameOverPhase) Enter(ge GameEngine) bool {
	// Clean up, announce winner, etc.
	return false
}

func init() {
	RegisterPhase(&GameOverPhase{})
}
