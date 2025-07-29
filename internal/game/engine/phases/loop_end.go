package phases

import (
	model "tragedylooper/pkg/proto/v1"
)

// --- LoopEndPhase ---
type LoopEndPhase struct{ basePhase }

func (p *LoopEndPhase) Type() model.GamePhase { return model.GamePhase_LOOP_END }
func (p *LoopEndPhase) Enter(ge GameEngine) Phase {
	if ge.GetGameState().CurrentLoop >= ge.GetGameConfig().GetScript().LoopCount {
		// Protagonists get a final chance to guess
		return &ProtagonistGuessPhase{}
	} else {
		return &LoopStartPhase{}
	}
}
