package phasehandler

import (
	"time"
	model "github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1"
)

// --- SetupPhase ---
type SetupPhase struct{}

// HandleAction is the default implementation for Phase interface, does nothing and returns nil.
func (p *SetupPhase) HandleAction(ge GameEngine, player *model.Player, action *model.PlayerActionPayload) Phase {
	return nil
}

// HandleEvent is the default implementation for Phase interface, does nothing and returns nil.
func (p *SetupPhase) HandleEvent(ge GameEngine, event *model.GameEvent) Phase { return nil }

// HandleTimeout is the default implementation for Phase interface, does nothing and returns nil.
func (p *SetupPhase) HandleTimeout(ge GameEngine) Phase { return nil }

// Exit is the default implementation for Phase interface, does nothing.
func (p *SetupPhase) Exit(ge GameEngine) {}

// TimeoutDuration is the default implementation for Phase interface, returns 0, indicating no timeout.
func (p *SetupPhase) TimeoutDuration() time.Duration { return 0 }

func (p *SetupPhase) Type() model.GamePhase { return model.GamePhase_GAME_PHASE_SETUP }
func (p *SetupPhase) Enter(ge GameEngine) {
	// TODO: Implement logic for Mastermind to choose sub-scenario and place characters.
	// For now, we transition directly.
}

func init() {
	RegisterPhase(&SetupPhase{})
}
