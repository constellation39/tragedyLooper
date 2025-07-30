package phase

import (
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

// --- LoopEndPhase ---
type LoopEndPhase struct{ basePhase }

func (p *LoopEndPhase) Type() model.GamePhase { return model.GamePhase_GAME_PHASE_LOOP_END }
func (p *LoopEndPhase) Enter(ge GameEngine) Phase {
	if ge.GetGameState().CurrentLoop >= ge.GetGameRepo().GetScript().LoopCount {
		// Protagonists get a final chance to guess
		return &ProtagonistGuessPhase{}
	} else {
		return &LoopStartPhase{}
	}
}
