package phasehandler

import (
	"time"
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

// --- MastermindSetupPhase ---
type MastermindSetupPhase struct{}

// HandleAction is the default implementation for Phase interface, does nothing and returns nil.
func (p *MastermindSetupPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	return nil
}

// HandleEvent is the default implementation for Phase interface, does nothing and returns nil.
func (p *MastermindSetupPhase) HandleEvent(ge GameEngine, event *model.GameEvent) Phase { return nil }

// HandleTimeout is the default implementation for Phase interface, does nothing and returns nil.
func (p *MastermindSetupPhase) HandleTimeout(ge GameEngine) Phase { return nil }

// Exit is the default implementation for Phase interface, does nothing.
func (p *MastermindSetupPhase) Exit(ge GameEngine) {}

// TimeoutDuration is the default implementation for Phase interface, returns 0, indicating no timeout.
func (p *MastermindSetupPhase) TimeoutDuration() time.Duration { return 0 }

func (p *MastermindSetupPhase) Type() model.GamePhase {
	return model.GamePhase_GAME_PHASE_MASTERMIND_SETUP
}
func (p *MastermindSetupPhase) Enter(ge GameEngine) Phase {
	// TODO: Mastermind places characters and sets up their board.
	// For now, we transition directly.
	return GetPhase(model.GamePhase_GAME_PHASE_LOOP_START)
}

func init() {
	RegisterPhase(&MastermindSetupPhase{})
}
