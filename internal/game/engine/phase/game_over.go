package phase

import (
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

// --- GameOverPhase ---
type GameOverPhase struct{ basePhase }

func (p *GameOverPhase) Type() model.GamePhase { return model.GamePhase_GAME_PHASE_GAME_OVER }
func (p *GameOverPhase) Enter(ge GameEngine) Phase {
	// Clean up, announce winner, etc.
	ge.StopGameLoop()
	return nil
}
