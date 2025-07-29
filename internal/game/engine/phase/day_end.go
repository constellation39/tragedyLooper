package phases

import (
	model "tragedylooper/pkg/proto/v1"
)

// --- DayEndPhase ---
type DayEndPhase struct{ basePhase }

func (p *DayEndPhase) Type() model.GamePhase { return model.GamePhase_DAY_END }
func (p *DayEndPhase) Enter(ge GameEngine) Phase {
	if ge.GetGameState().CurrentDay >= ge.GetGameState().DaysPerLoop {
		return &LoopEndPhase{}
	} else {
		return &DayStartPhase{}
	}
}
