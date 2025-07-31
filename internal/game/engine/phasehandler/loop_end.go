package phasehandler

import (
	"time"
	model "tragedylooper/pkg/proto/tragedylooper/v1"
)

// --- LoopEndPhase ---
type LoopEndPhase struct{}

// HandleAction is the default implementation for Phase interface, does nothing and returns nil.
func (p *LoopEndPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	return nil
}

// HandleEvent is the default implementation for Phase interface, does nothing and returns nil.
func (p *LoopEndPhase) HandleEvent(ge GameEngine, event *model.GameEvent) Phase { return nil }

// HandleTimeout is the default implementation for Phase interface, does nothing and returns nil.
func (p *LoopEndPhase) HandleTimeout(ge GameEngine) Phase { return nil }

// Exit is the default implementation for Phase interface, does nothing.
func (p *LoopEndPhase) Exit(ge GameEngine) {}

// TimeoutDuration is the default implementation for Phase interface, returns 0, indicating no timeout.
func (p *LoopEndPhase) TimeoutDuration() time.Duration { return 0 }

func (p *LoopEndPhase) Type() model.GamePhase { return model.GamePhase_GAME_PHASE_LOOP_END }
func (p *LoopEndPhase) Enter(ge GameEngine) Phase {
	if ge.GetGameState().CurrentLoop >= ge.GetGameRepo().GetScript().LoopCount {
		// Protagonists get a final chance to guess
		return GetPhase(model.GamePhase_GAME_PHASE_PROTAGONIST_GUESS)
	} else {
		return GetPhase(model.GamePhase_GAME_PHASE_LOOP_START)
	}
}

func init() {
	RegisterPhase(&LoopEndPhase{})
}
